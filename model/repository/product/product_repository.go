// Product Repository for Magento EAV Products
//
// Set PRODUCT_FLAT_CACHE=off in your environment to disable the global flatProductsCache.
// When disabled, all flat product queries will hit the database directly.

package product

import (
	"os"
	"strconv"
	"sync"

	"gorm.io/gorm"

	entity "magento.GO/model/entity"
	productEntity "magento.GO/model/entity/product"
)

var (
	attributeCodeMap map[uint16]string
	attributeCodeMapOnce sync.Once
	flatProductsCache = make(map[uint16]map[uint]map[string]interface{})
	flatProductsCacheOnce  sync.Once
	flatProductsCacheLock  sync.RWMutex
	cacheDisabled func() bool = func() bool { return os.Getenv("PRODUCT_FLAT_CACHE") == "off" }

	// Singleton per DB: one repo per gorm.DB instance (allows test isolation)
	productRepoCache = make(map[*gorm.DB]*ProductRepository)
	productRepoMu    sync.RWMutex
)

// GetProductRepository returns a ProductRepository for the given DB.
// Uses one repo per DB instance for test isolation.
func GetProductRepository(db *gorm.DB) *ProductRepository {
	productRepoMu.RLock()
	if r, ok := productRepoCache[db]; ok {
		productRepoMu.RUnlock()
		return r
	}
	productRepoMu.RUnlock()
	productRepoMu.Lock()
	defer productRepoMu.Unlock()
	if r, ok := productRepoCache[db]; ok {
		return r
	}
	r := NewProductRepository(db)
	productRepoCache[db] = r
	return r
}

func getGlobalAttributeCodeMap(db *gorm.DB) map[uint16]string {
	attributeCodeMapOnce.Do(func() {
		attributeCodeMap, _ = LoadAttributeCodeMap(db)
	})
	return attributeCodeMap
}

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db}
}

func (r *ProductRepository) FindAll() ([]productEntity.Product, error) {
	return r.FindAllWithLimit(0)
}

func (r *ProductRepository) FindAllWithLimit(limit int) ([]productEntity.Product, error) {
	// Get product IDs first
	var productIDs []uint
	db := r.db.Model(&productEntity.Product{}).Select("entity_id").Order("entity_id")
	if limit > 0 {
		db = db.Limit(limit)
	}
	if err := db.Pluck("entity_id", &productIDs).Error; err != nil {
		return nil, err
	}

	if len(productIDs) == 0 {
		return []productEntity.Product{}, nil
	}

	// Fetch in batches to avoid placeholder limit
	var allProducts []productEntity.Product
	for i := 0; i < len(productIDs); i += batchSize {
		end := i + batchSize
		if end > len(productIDs) {
			end = len(productIDs)
		}
		var batch []productEntity.Product
		err := r.db.Preload("Categories").Preload("MediaGallery").
			Where("entity_id IN ?", productIDs[i:end]).
			Find(&batch).Error
		if err != nil {
			return nil, err
		}
		allProducts = append(allProducts, batch...)
	}
	return allProducts, nil
}

func (r *ProductRepository) FindByID(id uint) (*productEntity.Product, error) {
	var product productEntity.Product
	err := r.db.Preload("Categories").Preload("MediaGallery").First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) Create(product *productEntity.Product) error {
	return r.db.Create(product).Error
}

func (r *ProductRepository) Update(product *productEntity.Product) error {
	return r.db.Save(product).Error
}

func (r *ProductRepository) Delete(id uint) error {
	return r.db.Delete(&productEntity.Product{}, id).Error
}

// FetchProductIDsByCategoryWithPosition returns product IDs in a category ordered by position (ASC).
func (r *ProductRepository) FetchProductIDsByCategoryWithPosition(categoryID uint, asc bool) ([]uint, error) {
	order := "position ASC"
	if !asc {
		order = "position DESC"
	}
	var rows []struct {
		ProductID uint `gorm:"column:product_id"`
	}
	err := r.db.Table("catalog_category_product").
		Select("product_id").
		Where("category_id = ?", categoryID).
		Order(order).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	ids := make([]uint, len(rows))
	for i, row := range rows {
		ids[i] = row.ProductID
	}
	return ids, nil
}

func (r *ProductRepository) FetchWithAllAttributes(storeID ...uint16) ([]productEntity.Product, error) {
	var products []productEntity.Product
	sid := uint16(0)
	if len(storeID) > 0 {
		sid = storeID[0]
	}
	err := r.db.
		Preload("Categories").
		Preload("MediaGallery").
		Preload("StockItem").
		Preload("ProductIndexPrices").
		Preload("Varchars", "store_id = ?", sid).
		Preload("Ints", "store_id = ?", sid).
		Preload("Decimals", "store_id = ?", sid).
		Preload("Texts", "store_id = ?", sid).
		Preload("Datetimes", "store_id = ?", sid).
		Find(&products).Error
	return products, err
}

const batchSize = 1000 // MySQL placeholder limit is 65535; batching avoids "too many placeholders" error

// fetchFlatProducts loads products with all EAV attributes. Uses batched fetching
// to avoid MySQL's prepared statement placeholder limit (65535).
func (r *ProductRepository) fetchFlatProducts(ids []uint, storeID uint16) (map[uint]map[string]interface{}, error) {
	return r.fetchFlatProductsWithLimit(ids, storeID, 0)
}

func (r *ProductRepository) fetchFlatProductsWithLimit(ids []uint, storeID uint16, limit int) (map[uint]map[string]interface{}, error) {
	// If specific IDs provided and within batch size, fetch directly
	if ids != nil && len(ids) > 0 && len(ids) <= batchSize {
		return r.fetchFlatProductsBatch(ids, storeID)
	}

	// If specific IDs provided but too many, batch them
	if ids != nil && len(ids) > batchSize {
		return r.fetchFlatProductsInBatches(ids, storeID)
	}

	// No IDs specified - fetch all products in batches
	return r.fetchAllFlatProductsInBatches(storeID, limit)
}

// fetchFlatProductsBatch fetches a single batch of products (up to batchSize)
func (r *ProductRepository) fetchFlatProductsBatch(ids []uint, storeID uint16) (map[uint]map[string]interface{}, error) {
	var products []productEntity.Product
	db := r.db.
		Preload("Categories").
		Preload("MediaGallery").
		Preload("StockItem").
		Preload("ProductIndexPrices").
		Preload("Varchars", "store_id = ?", storeID).
		Preload("Ints", "store_id = ?", storeID).
		Preload("Decimals", "store_id = ?", storeID).
		Preload("Texts", "store_id = ?", storeID).
		Preload("Datetimes", "store_id = ?", storeID)

	if len(ids) > 0 {
		db = db.Where("entity_id IN ?", ids)
	}

	if err := db.Find(&products).Error; err != nil {
		return nil, err
	}

	attrMap := getGlobalAttributeCodeMap(r.db)
	flatProducts := make(map[uint]map[string]interface{}, len(products))
	for i := range products {
		id := products[i].EntityID
		flatProducts[id] = FlattenProductAttributesWithCodes(&products[i], attrMap)
	}
	return flatProducts, nil
}

// fetchFlatProductsInBatches fetches specific IDs in batches
func (r *ProductRepository) fetchFlatProductsInBatches(ids []uint, storeID uint16) (map[uint]map[string]interface{}, error) {
	result := make(map[uint]map[string]interface{}, len(ids))

	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		batch, err := r.fetchFlatProductsBatch(ids[i:end], storeID)
		if err != nil {
			return nil, err
		}
		for id, prod := range batch {
			result[id] = prod
		}
	}
	return result, nil
}

// fetchAllFlatProductsInBatches fetches all products using offset-based pagination
func (r *ProductRepository) fetchAllFlatProductsInBatches(storeID uint16, limit int) (map[uint]map[string]interface{}, error) {
	// First, get all product IDs
	var productIDs []uint
	db := r.db.Model(&productEntity.Product{}).Select("entity_id").Order("entity_id")
	if limit > 0 {
		db = db.Limit(limit)
	}
	if err := db.Pluck("entity_id", &productIDs).Error; err != nil {
		return nil, err
	}

	if len(productIDs) == 0 {
		return make(map[uint]map[string]interface{}), nil
	}

	// Fetch in batches
	return r.fetchFlatProductsInBatches(productIDs, storeID)
}

func (r *ProductRepository) FetchWithAllAttributesFlat(storeID ...uint16) (map[uint]map[string]interface{}, error) {
	return r.FetchWithAllAttributesFlatWithLimit(0, storeID...)
}

func (r *ProductRepository) FetchWithAllAttributesFlatWithLimit(limit int, storeID ...uint16) (map[uint]map[string]interface{}, error) {
	sid := uint16(0)
	if len(storeID) > 0 {
		sid = storeID[0]
	}

	// If limit specified, don't use cache - fetch directly
	if limit > 0 {
		return r.fetchFlatProductsWithLimit(nil, sid, limit)
	}

	if cacheDisabled() {
		return r.fetchFlatProducts(nil, sid)
	}

	// Check cache
	flatProductsCacheLock.RLock()
	cached, ok := flatProductsCache[sid]
	flatProductsCacheLock.RUnlock()
	if ok {
		return cached, nil
	}

	// Not cached: fetch, flatten, and cache
	flatProducts, err := r.fetchFlatProducts(nil, sid)
	if err != nil {
		return nil, err
	}

	flatProductsCacheLock.Lock()
	flatProductsCache[sid] = flatProducts
	flatProductsCacheLock.Unlock()

	return flatProducts, nil
}

func (r *ProductRepository) FetchWithAllAttributesFlatByIDs(ids []uint, storeID ...uint16) (map[uint]map[string]interface{}, error) {
	sid := uint16(0)
	if len(storeID) > 0 {
		sid = storeID[0]
	}

	if cacheDisabled() {
		return r.fetchFlatProducts(ids, sid)
	}

	// If no ids provided, fallback to all (cached)
	if ids == nil || len(ids) == 0 {
		return r.FetchWithAllAttributesFlat(sid)
	}

	result := make(map[uint]map[string]interface{})
	missingIDs := make([]uint, 0, len(ids))

	// Check cache for each id
	flatProductsCacheLock.RLock()
	cached, ok := flatProductsCache[sid]
	flatProductsCacheLock.RUnlock()
	if ok {
		for _, id := range ids {
			if prod, found := cached[id]; found {
				result[id] = prod
			} else {
				missingIDs = append(missingIDs, id)
			}
		}
	} else {
		missingIDs = ids
	}

	// If there are missing IDs, fetch them from DB
	if len(missingIDs) > 0 {
		fetched, err := r.fetchFlatProducts(missingIDs, sid)
		if err != nil {
			return nil, err
		}
		// Add to result and update cache
		/* there is an issue when fetching all after fetch by ID
		flatProductsCacheLock.Lock()
		if flatProductsCache[sid] == nil {
			flatProductsCache[sid] = make(map[uint]map[string]interface{})
		}
		for id, prod := range fetched {
			result[id] = prod
			flatProductsCache[sid][id] = prod
		}
		flatProductsCacheLock.Unlock()
		*/
		result = fetched
	}

	return result, nil
}

func attrKey(attrMap map[uint16]string, attrID uint16) string {
	if k := attrMap[attrID]; k != "" {
		return k
	}
	return strconv.FormatUint(uint64(attrID), 10)
}

func FlattenProductAttributesWithCodes(product *productEntity.Product, attrMap map[uint16]string) map[string]interface{} {
	n := 5 + len(product.Varchars) + len(product.Ints) + len(product.Decimals) + len(product.Texts) + len(product.Datetimes)
	if len(product.Categories) > 0 {
		n++
	}
	if len(product.MediaGallery) > 0 {
		n++
	}
	if product.StockItem.ProductID != 0 {
		n++
	}
	if len(product.ProductIndexPrices) > 0 {
		n++
	}
	attrs := make(map[string]interface{}, n)
	attrs["entity_id"] = product.EntityID
	attrs["sku"] = product.SKU
	attrs["type_id"] = product.TypeID
	attrs["created_at"] = product.CreatedAt
	attrs["updated_at"] = product.UpdatedAt

	for _, v := range product.Varchars {
		attrs[attrKey(attrMap, v.AttributeID)] = v.Value
	}
	for _, v := range product.Ints {
		attrs[attrKey(attrMap, v.AttributeID)] = v.Value
	}
	for _, v := range product.Decimals {
		attrs[attrKey(attrMap, v.AttributeID)] = v.Value
	}
	for _, v := range product.Texts {
		attrs[attrKey(attrMap, v.AttributeID)] = v.Value
	}
	for _, v := range product.Datetimes {
		attrs[attrKey(attrMap, v.AttributeID)] = v.Value
	}

	var categoryIDs []uint
	if len(product.Categories) > 0 {
		categoryIDs = make([]uint, 0, len(product.Categories))
		for _, cat := range product.Categories {
			categoryIDs = append(categoryIDs, cat.EntityID)
		}
	}
	attrs["category_ids"] = categoryIDs

	mediaGallery := make([]map[string]interface{}, 0, len(product.MediaGallery))
	for _, mg := range product.MediaGallery {
		mediaGallery = append(mediaGallery, map[string]interface{}{
			"value_id":   mg.ValueID,
			"value":      mg.Value,
			"media_type": mg.MediaType,
			"disabled":   mg.Disabled,
		})
	}
	attrs["media_gallery"] = mediaGallery

	// Flatten stock item
	if product.StockItem.ProductID != 0 {
		stock := map[string]interface{}{
			"item_id": product.StockItem.ItemID,
			"qty": product.StockItem.Qty,
			"is_in_stock": product.StockItem.IsInStock,
			"min_qty": product.StockItem.MinQty,
			"max_sale_qty": product.StockItem.MaxSaleQty,
			"manage_stock": product.StockItem.ManageStock,
			"website_id": product.StockItem.WebsiteID,
		}
		attrs["stock_item"] = stock
	}

	indexPrices := make([]map[string]interface{}, 0, len(product.ProductIndexPrices))
	for _, ip := range product.ProductIndexPrices {
		indexPrices = append(indexPrices, map[string]interface{}{
			"entity_id":         ip.EntityID,
			"customer_group_id": ip.CustomerGroupID,
			"website_id":        ip.WebsiteID,
			"tax_class_id":      ip.TaxClassID,
			"price":             ip.Price,
			"final_price":       ip.FinalPrice,
			"min_price":         ip.MinPrice,
			"max_price":         ip.MaxPrice,
			"tier_price":        ip.TierPrice,
		})
	}
	attrs["index_prices"] = indexPrices

	return attrs
}

func LoadAttributeCodeMap(db *gorm.DB) (map[uint16]string, error) {
	var attrs []entity.EavAttribute
	err := db.Find(&attrs).Error
	if err != nil {
		return nil, err
	}
	m := make(map[uint16]string)
	for _, a := range attrs {
		m[a.AttributeID] = a.AttributeCode
	}
	return m, nil
} 