// Product Repository for Magento EAV Products
//
// Set PRODUCT_FLAT_CACHE=off in your environment to disable the global flatProductsCache.
// When disabled, all flat product queries will hit the database directly.

package product

import (
	"fmt"
	"gorm.io/gorm"
	entity "magento.GO/model/entity"
	productEntity "magento.GO/model/entity/product"
	"os"
	"sync"
)

var (
	attributeCodeMap      map[uint16]string
	attributeCodeMapOnce  sync.Once
	flatProductsCache     = make(map[uint16]map[uint]map[string]interface{})
	flatProductsCacheOnce sync.Once
	flatProductsCacheLock sync.RWMutex
	cacheDisabled         = os.Getenv("PRODUCT_FLAT_CACHE") == "off"

	// Singleton for ProductRepository
	productRepoInstance *ProductRepository
	productRepoOnce     sync.Once
)

// GetProductRepository returns the singleton instance of ProductRepository
func GetProductRepository(db *gorm.DB) *ProductRepository {
	productRepoOnce.Do(func() {
		productRepoInstance = NewProductRepository(db)
	})
	return productRepoInstance
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
	var products []productEntity.Product
	err := r.db.Preload("Categories").Preload("MediaGallery").Find(&products).Error
	return products, err
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

func (r *ProductRepository) fetchFlatProducts(ids []uint, storeID uint16) (map[uint]map[string]interface{}, error) {
	var products []productEntity.Product
	fmt.Println("DEBUG: ids:", ids)
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

	if ids != nil && len(ids) > 0 {
		db = db.Where("entity_id IN ?", ids)
	}

	err := db.Find(&products).Error
	if err != nil {
		return nil, err
	}

	attrMap := getGlobalAttributeCodeMap(r.db)
	flatProducts := make(map[uint]map[string]interface{})
	for i := range products {
		id := products[i].EntityID
		flatProducts[id] = FlattenProductAttributesWithCodes(&products[i], attrMap)
	}

	return flatProducts, nil
}

func (r *ProductRepository) FetchWithAllAttributesFlat(storeID ...uint16) (map[uint]map[string]interface{}, error) {
	sid := uint16(0)
	if len(storeID) > 0 {
		sid = storeID[0]
	}

	if cacheDisabled {
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

	if cacheDisabled {
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

func FlattenProductAttributesWithCodes(product *productEntity.Product, attrMap map[uint16]string) map[string]interface{} {
	attrs := map[string]interface{}{}
	attrs["entity_id"] = product.EntityID
	attrs["sku"] = product.SKU
	attrs["type_id"] = product.TypeID
	attrs["created_at"] = product.CreatedAt
	attrs["updated_at"] = product.UpdatedAt

	for _, v := range product.Varchars {
		key := attrMap[v.AttributeID]
		if key == "" {
			key = fmt.Sprintf("%d", v.AttributeID)
		}
		attrs[key] = v.Value
	}
	for _, v := range product.Ints {
		key := attrMap[v.AttributeID]
		if key == "" {
			key = fmt.Sprintf("%d", v.AttributeID)
		}
		attrs[key] = v.Value
	}
	for _, v := range product.Decimals {
		key := attrMap[v.AttributeID]
		if key == "" {
			key = fmt.Sprintf("%d", v.AttributeID)
		}
		attrs[key] = v.Value
	}
	for _, v := range product.Texts {
		key := attrMap[v.AttributeID]
		if key == "" {
			key = fmt.Sprintf("%d", v.AttributeID)
		}
		attrs[key] = v.Value
	}
	for _, v := range product.Datetimes {
		key := attrMap[v.AttributeID]
		if key == "" {
			key = fmt.Sprintf("%d", v.AttributeID)
		}
		attrs[key] = v.Value
	}

	var categoryIDs []uint
	for _, cat := range product.Categories {
		categoryIDs = append(categoryIDs, cat.EntityID)
	}
	attrs["category_ids"] = categoryIDs

	// Flatten media gallery
	var mediaGallery []map[string]interface{}
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
			"item_id":      product.StockItem.ItemID,
			"qty":          product.StockItem.Qty,
			"is_in_stock":  product.StockItem.IsInStock,
			"min_qty":      product.StockItem.MinQty,
			"max_sale_qty": product.StockItem.MaxSaleQty,
			"manage_stock": product.StockItem.ManageStock,
			"website_id":   product.StockItem.WebsiteID,
		}
		attrs["stock_item"] = stock
	}

	// Flatten product index prices
	var indexPrices []map[string]interface{}
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
