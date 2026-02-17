package category

import (
	"gorm.io/gorm"
	"log"
	entity "magento.GO/model/entity"
	categoryEntity "magento.GO/model/entity/category"
	"sync"
)

var (
	categoryAttrMetaCache map[uint]entity.EavAttribute
	categoryAttrMetaOnce  sync.Once
	treeCache             map[uint16][]*CategoryTreeNode
	treeCacheLock         sync.RWMutex

	// Singleton for CategoryRepository
	categoryRepoInstance *CategoryRepository
	categoryRepoOnce     sync.Once
)

// GetCategoryRepository returns the singleton instance of CategoryRepository
func GetCategoryRepository(db *gorm.DB) *CategoryRepository {
	categoryRepoOnce.Do(func() {
		categoryRepoInstance = NewCategoryRepository(db)
	})
	return categoryRepoInstance
}

// CategoryRepository provides access to category data with in-memory caching for performance.
type CategoryRepository struct {
	db *gorm.DB
	// cache stores categories per store: cache[storeID][categoryID] = CategoryWithAttributes
	cache     map[uint16]map[uint]CategoryWithAttributes
	cacheLock sync.RWMutex
}

// NewCategoryRepository creates a new repository instance
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// FetchAllWithAttributes returns all categories with their EAV attributes (int, varchar, text) for a given store.
// Pass storeID=0 for global attributes.
func (r *CategoryRepository) FetchAllWithAttributes(storeID uint16) ([]categoryEntity.Category, error) {
	cats, err := r.FetchAllWithAttributesMap(storeID)
	if err != nil {
		return nil, err
	}
	result := make([]categoryEntity.Category, 0, len(cats))
	for _, cat := range cats {
		result = append(result, cat.Category)
	}
	return result, nil
}

// FetchAllWithAttributesMap returns a map[category_id]CategoryWithAttributes for a given store.
// The first call for each storeID loads all categories from the database and caches them in memory.
// Subsequent calls return the cached data for fast access.
// Thread-safe for concurrent use.
func (r *CategoryRepository) FetchAllWithAttributesMap(storeID uint16) (map[uint]CategoryWithAttributes, error) {

	if r.cache == nil {
		r.cache = make(map[uint16]map[uint]CategoryWithAttributes)
	}
	// Check if cache for this storeID exists
	if catsIface, ok := r.GetCacheCategory(storeID, 0); ok {
		if cats, ok := catsIface.(map[uint]CategoryWithAttributes); ok {
			return cats, nil
		}
	}

	// Not cached: load from DB
	var categories []categoryEntity.Category
	err := r.db.
		Preload("Products").
		Preload("Ints", "store_id = ?", storeID).
		Preload("Varchars", "store_id = ?", storeID).
		Preload("Texts", "store_id = ?", storeID).
		Find(&categories).Error
	if err != nil {
		return nil, err
	}
	attrMeta, err := LoadCategoryAttributeMeta(r.db)
	if err != nil {
		return nil, err
	}
	cache := make(map[uint]CategoryWithAttributes, len(categories))
	for _, cat := range categories {
		flat := FlattenCategoryAttributesWithLabels(&cat, attrMeta)
		cache[cat.EntityID] = CategoryWithAttributes{
			Category:   cat,
			Attributes: flat,
		}
	}

	r.cacheLock.Lock()
	r.cache[storeID] = cache
	r.cacheLock.Unlock()

	return cache, nil
}

// InvalidateCache clears the in-memory category cache for all stores.
// The next call to FetchAllWithAttributes or FetchAllWithAttributesMap will reload from the database.
func (r *CategoryRepository) InvalidateCache() {
	r.cacheLock.Lock()
	defer r.cacheLock.Unlock()
	r.cache = make(map[uint16]map[uint]CategoryWithAttributes)
	// Invalidate tree cache as well
	treeCacheLock.Lock()
	treeCache = make(map[uint16][]*CategoryTreeNode)
	treeCacheLock.Unlock()
}

func (r *CategoryRepository) GetByIDWithAttributesAndFlat(id uint, storeID uint16) (*categoryEntity.Category, map[string]map[string]interface{}, error) {
	cats, flats, err := r.GetByIDsWithAttributesAndFlat([]uint{id}, storeID)
	if err != nil {
		return nil, nil, err
	}
	return &cats[0], flats[0], nil
}

// GetByIDWithAttributesAndFlat returns a single category by ID with EAV attributes and products for a given store, plus a flat attributes map.
// Loads from the database with store filter and loads attribute meta for flattening.
func (r *CategoryRepository) GetByIDsWithAttributesAndFlat(ids []uint, storeID uint16) ([]categoryEntity.Category, []map[string]map[string]interface{}, error) {
	if len(ids) == 0 {
		return nil, nil, nil
	}

	// Prepare result slices
	resultCats := make([]categoryEntity.Category, 0, len(ids))
	resultFlats := make([]map[string]map[string]interface{}, 0, len(ids))
	missingIDs := make([]uint, 0)

	// 1. Try cache first

	for _, id := range ids {
		catIface, ok := r.GetCacheCategory(storeID, id)
		if ok {
			if catWithAttrs, ok := catIface.(CategoryWithAttributes); ok {
				resultCats = append(resultCats, catWithAttrs.Category)
				resultFlats = append(resultFlats, catWithAttrs.Attributes)
				continue
			}
		}
		missingIDs = append(missingIDs, id)
	}
	//log.Printf("missingIDs: %v", missingIDs)

	// 2. Fetch missing from DB if needed (do NOT update cache)
	if len(missingIDs) > 0 {
		log.Printf("Get missingIDs from DB: %v", missingIDs)
		cats, err := r.GetByIDsWithAttributes(missingIDs, storeID)
		if err != nil {
			return nil, nil, err
		}
		attrMeta, err := LoadCategoryAttributeMeta(r.db)
		if err != nil {
			return nil, nil, err
		}
		idToCat := make(map[uint]categoryEntity.Category)
		idToFlat := make(map[uint]map[string]map[string]interface{})
		for i := range cats {
			idToCat[cats[i].EntityID] = cats[i]
			flat := FlattenCategoryAttributesWithLabels(&cats[i], attrMeta)
			idToFlat[cats[i].EntityID] = flat
		}
		for _, id := range missingIDs {
			if cat, ok := idToCat[id]; ok {
				resultCats = append(resultCats, cat)
				resultFlats = append(resultFlats, idToFlat[id])
			}
		}
	}

	return resultCats, resultFlats, nil
}

// FlattenCategoryAttributesWithLabels flattens a category's EAV attributes into a map[attribute_code]map[string]interface{}.
// Each entry contains the value, label, and store_id for the attribute.
// attrMeta should be a map[uint]EavAttribute, keyed by attribute_id.
func FlattenCategoryAttributesWithLabels(
	category *categoryEntity.Category,
	attrMeta map[uint]entity.EavAttribute,
) map[string]map[string]interface{} {
	flat := map[string]map[string]interface{}{}
	// Ints
	for _, v := range category.Ints {
		if attr, ok := attrMeta[uint(v.AttributeID)]; ok {
			label := ""
			if attr.FrontendLabel != nil {
				label = *attr.FrontendLabel
			}
			flat[attr.AttributeCode] = map[string]interface{}{
				"value":    v.Value,
				"label":    label,
				"store_id": v.StoreID,
			}
		}
	}
	// Varchars
	for _, v := range category.Varchars {
		if attr, ok := attrMeta[uint(v.AttributeID)]; ok {
			label := ""
			if attr.FrontendLabel != nil {
				label = *attr.FrontendLabel
			}
			flat[attr.AttributeCode] = map[string]interface{}{
				"value":    v.Value,
				"label":    label,
				"store_id": v.StoreID,
			}
		}
	}
	// Texts
	for _, v := range category.Texts {
		if attr, ok := attrMeta[uint(v.AttributeID)]; ok {
			label := ""
			if attr.FrontendLabel != nil {
				label = *attr.FrontendLabel
			}
			flat[attr.AttributeCode] = map[string]interface{}{
				"value":    v.Value,
				"label":    label,
				"store_id": v.StoreID,
			}
		}
	}
	// Add core fields if needed
	flat["entity_id"] = map[string]interface{}{"value": category.EntityID, "label": "Entity ID", "store_id": 0}
	// ...add more core fields as needed
	return flat
}

/* Usage Example:

// attrMeta := map[uint]entity.EavAttribute{ ... } // load from DB
// cat := ... // load category
flat := FlattenCategoryAttributesWithLabels(cat, attrMeta)
for code, info := range flat {
	fmt.Printf("%s: value=%v, label=%s, store_id=%v\n", code, info["value"], info["label"], info["store_id"])
}
*/

/* Usage Example:

// Create repository
repo := NewCategoryRepository(db)

// Get all categories as a slice (uses cache if available)
categories, err := repo.FetchAllWithAttributes(0)
if err != nil {
	// handle error
}
for _, cat := range categories {
	// cat.EntityInts, cat.EntityVarchars, cat.EntityTexts
}

// Get all categories as a map[category_id]Category for a specific store
catMap, err := repo.FetchAllWithAttributesMap(2) // store_id = 2
if err == nil {
	cat := catMap[123] // get category by ID
}

// Invalidate the cache (e.g., after a write operation)
repo.InvalidateCache()
*/

// LoadCategoryAttributeMeta loads all EAV attributes for categories and returns a map[uint]entity.EavAttribute keyed by attribute_id.
// Uses in-memory cache for performance.
func LoadCategoryAttributeMeta(db *gorm.DB) (map[uint]entity.EavAttribute, error) {
	var loadErr error
	categoryAttrMetaOnce.Do(func() {
		var attrs []entity.EavAttribute
		loadErr = db.Find(&attrs).Error
		if loadErr == nil {
			m := make(map[uint]entity.EavAttribute)
			for _, a := range attrs {
				m[uint(a.AttributeID)] = a
			}
			categoryAttrMetaCache = m
		}
	})
	if categoryAttrMetaCache == nil {
		return nil, loadErr
	}
	return categoryAttrMetaCache, nil
}

// InvalidateCategoryAttributeMetaCache clears the attribute meta cache (for use after attribute changes)
func InvalidateCategoryAttributeMetaCache() {
	categoryAttrMetaCache = nil
	categoryAttrMetaOnce = sync.Once{}
}

type CategoryWithAttributes struct {
	categoryEntity.Category
	Attributes map[string]map[string]interface{} `json:"attributes"`
}

func (r *CategoryRepository) GetByIDsWithAttributes(ids []uint, storeID uint16) ([]categoryEntity.Category, error) {
	var cats []categoryEntity.Category
	err := r.db.
		Preload("Products").
		Preload("Ints", "store_id = ?", storeID).
		Preload("Varchars", "store_id = ?", storeID).
		Preload("Texts", "store_id = ?", storeID).
		Where("entity_id IN ?", ids).
		Find(&cats).Error
	if err != nil {
		return nil, err
	}
	return cats, nil
}

type CategoryTreeNode struct {
	Category   categoryEntity.Category
	Attributes map[string]map[string]interface{}
	Children   []*CategoryTreeNode
}

// BuildCategoryTree builds a tree of categories (with flat attributes) starting from the given parentID (usually 0 for root).
func (r *CategoryRepository) BuildCategoryTree(storeID uint16, parentID uint) ([]*CategoryTreeNode, error) {
	// Use the cache for performance
	treeCacheLock.RLock()
	if treeCache != nil {
		if tree, ok := treeCache[storeID]; ok {
			treeCacheLock.RUnlock()
			// Return only the subtree starting from parentID
			if parentID == 0 {
				return tree, nil
			}
			// Find subtree for parentID
			var findSubtree func(nodes []*CategoryTreeNode, pid uint) []*CategoryTreeNode
			findSubtree = func(nodes []*CategoryTreeNode, pid uint) []*CategoryTreeNode {
				for _, node := range nodes {
					if node.Category.EntityID == pid {
						return node.Children
					}
					if len(node.Children) > 0 {
						if sub := findSubtree(node.Children, pid); sub != nil {
							return sub
						}
					}
				}
				return nil
			}
			return findSubtree(tree, parentID), nil
		}
	}
	treeCacheLock.RUnlock()

	catMap, err := r.FetchAllWithAttributesMap(storeID)
	if err != nil {
		return nil, err
	}

	// Map of parentID to []*CategoryTreeNode
	childrenMap := make(map[uint][]*CategoryTreeNode)
	for _, catWithAttrs := range catMap {
		node := &CategoryTreeNode{
			Category:   catWithAttrs.Category,
			Attributes: catWithAttrs.Attributes,
		}
		childrenMap[catWithAttrs.Category.ParentID] = append(childrenMap[catWithAttrs.Category.ParentID], node)
	}

	// Recursive function to build the tree
	var build func(parentID uint) []*CategoryTreeNode
	build = func(parentID uint) []*CategoryTreeNode {
		nodes := childrenMap[parentID]
		for _, node := range nodes {
			node.Children = build(node.Category.EntityID)
		}
		return nodes
	}

	tree := build(0)

	// Cache the full tree for this storeID
	treeCacheLock.Lock()
	if treeCache == nil {
		treeCache = make(map[uint16][]*CategoryTreeNode)
	}
	treeCache[storeID] = tree
	treeCacheLock.Unlock()

	// Return the requested subtree
	if parentID == 0 {
		return tree, nil
	}
	// Find subtree for parentID
	var findSubtree func(nodes []*CategoryTreeNode, pid uint) []*CategoryTreeNode
	findSubtree = func(nodes []*CategoryTreeNode, pid uint) []*CategoryTreeNode {
		for _, node := range nodes {
			if node.Category.EntityID == pid {
				return node.Children
			}
			if len(node.Children) > 0 {
				if sub := findSubtree(node.Children, pid); sub != nil {
					return sub
				}
			}
		}
		return nil
	}
	return findSubtree(tree, parentID), nil
}

// GetCacheCategory returns the cached category (with attributes) by id if provided, or all cached categories for a given storeID if id is zero.
// If id == 0, returns all cached categories for the storeID. If id > 0, returns the specific category and a bool indicating if found.
func (r *CategoryRepository) GetCacheCategory(storeID uint16, id uint) (interface{}, bool) {
	r.cacheLock.RLock()
	log.Printf("CategoryId/StoreId: %v/%v", id, storeID)
	defer r.cacheLock.RUnlock()
	if r.cache == nil {
		return nil, false
	}
	cats, ok := r.cache[storeID]
	if !ok {
		return nil, false
	}
	if id == 0 {
		return cats, true
	}
	cat, found := cats[id]
	return cat, found
}
