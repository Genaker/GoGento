package html

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"html/template"
	parts "magento.GO/html/parts"
	productRepo "magento.GO/model/repository/product"
	"net/http"
	"strconv"
	//"io"
	"bytes"
	"log"
	"magento.GO/config"
	categoryRepo "magento.GO/model/repository/category"
	"strings"
	"sync"
	"time"
)

var (
	categoryTreeHTMLCache string
	categoryTreeCacheTime time.Time
	categoryTreeCacheLock sync.RWMutex
)

func RenderCategoryTreeCached(tmpl *template.Template, tree interface{}) (string, error) {
	categoryTreeCacheLock.RLock()
	if time.Since(categoryTreeCacheTime) < 30*time.Minute && categoryTreeHTMLCache != "" {
		cached := categoryTreeHTMLCache
		categoryTreeCacheLock.RUnlock()
		return cached, nil
	}
	categoryTreeCacheLock.RUnlock()

	// Not cached or expired: render and cache
	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, "category_tree", tree)
	if err != nil {
		return "", err
	}

	categoryTreeCacheLock.Lock()
	categoryTreeHTMLCache = buf.String()
	categoryTreeCacheTime = time.Now()
	categoryTreeCacheLock.Unlock()

	return categoryTreeHTMLCache, nil
}

// Helper to build breadcrumbs from a category path string
func buildCategoryBreadcrumbs(repo *categoryRepo.CategoryRepository, path string, storeID uint16) ([]map[string]interface{}, error) {
	var breadcrumbIDs []uint
	exclude := map[uint]bool{0: true, 1: true, 2: true}
	for _, idStr := range strings.Split(path, "/") {
		if idStr == "" {
			continue
		}
		if id, err := strconv.ParseUint(idStr, 10, 64); err == nil {
			uid := uint(id)
			if exclude[uid] {
				continue
			}
			breadcrumbIDs = append(breadcrumbIDs, uid)
		}
	}
	if len(breadcrumbIDs) == 0 {
		return nil, nil
	}
	cats, flats, err := repo.GetByIDsWithAttributesAndFlat(breadcrumbIDs, storeID)
	if err != nil {
		return nil, err
	}
	// Map by ID for quick lookup
	catMap := make(map[uint]map[string]interface{})
	for i := range cats {
		name := ""
		if n, ok := flats[i]["name"]; ok {
			if s, ok := n["value"].(string); ok {
				name = s
			}
		}
		catMap[cats[i].EntityID] = map[string]interface{}{
			"EntityID": cats[i].EntityID,
			"Name":     name,
		}
	}
	// Order as in breadcrumbIDs
	var breadcrumbCats []map[string]interface{}
	for _, id := range breadcrumbIDs {
		if c, ok := catMap[id]; ok {
			breadcrumbCats = append(breadcrumbCats, c)
		}
	}
	return breadcrumbCats, nil
}

// Helper to get the last category ID from various slice types
func getLastCategoryID(idsVal interface{}) (uint, bool) {
	switch ids := idsVal.(type) {
	case []int:
		if len(ids) > 0 {
			return uint(ids[len(ids)-1]), true
		}
	case []float64:
		if len(ids) > 0 {
			return uint(ids[len(ids)-1]), true
		}
	case []uint:
		if len(ids) > 0 {
			return ids[len(ids)-1], true
		}
	case []interface{}:
		if len(ids) > 0 {
			switch v := ids[len(ids)-1].(type) {
			case int:
				return uint(v), true
			case float64:
				return uint(v), true
			case uint:
				return v, true
			}
		}
	}
	return 0, false
}

// RegisterProductHTMLRoutes registers HTML routes for product rendering
func RegisterProductHTMLRoutes(e *echo.Echo, db *gorm.DB) {
	repo := productRepo.GetProductRepository(db)
	catRepo := categoryRepo.GetCategoryRepository(db)

	e.GET("/product/:ids", func(c echo.Context) error {
		idsParam := c.Param("ids")
		var ids []uint
		if idsParam != "" {
			idsStr := strings.Split(idsParam, ",")
			for _, idStr := range idsStr {
				idStr = strings.TrimSpace(idStr)
				if idStr == "" {
					continue
				}
				idUint, err := strconv.ParseUint(idStr, 10, 64)
				if err != nil {
					continue // skip invalid IDs
				}
				ids = append(ids, uint(idUint))
			}
		}
		flatProducts, err := repo.FetchWithAllAttributesFlatByIDs(ids)
		if err != nil {
			log.Println("Repo error:", err)
			return c.String(http.StatusInternalServerError, "Error fetching products")
		}
		var products []map[string]interface{}
		for _, id := range ids {
			if prod, ok := flatProducts[id]; ok {
				// Add breadcrumbs if category_ids is present
				if idsVal, ok := prod["category_ids"]; ok {

					if lastCatID, ok := getLastCategoryID(idsVal); ok && lastCatID > 0 {
						cat, _, err := catRepo.GetByIDWithAttributesAndFlat(lastCatID, 0)
						if err == nil && cat != nil && cat.Path != "" {
							breadcrumbs, _ := buildCategoryBreadcrumbs(catRepo, cat.Path, 0)
							prod["Breadcrumbs"] = breadcrumbs
							// Debug output
							var bcIDs []uint
							for _, bc := range breadcrumbs {
								if id, ok := bc["EntityID"].(uint); ok {
									bcIDs = append(bcIDs, id)
								}
							}
							//log.Printf("Product %v breadcrumbs: %v", id, bcIDs)
						}
					}
				}
				// Process description to be safe HTML for template rendering
				if desc, ok := prod["description"].(string); ok {
					prod["description"] = template.HTML(desc)
				}
				products = append(products, prod)
				//log.Printf("Product %v: %v", id, prod)
			}
		}
		categoryTree, err := catRepo.BuildCategoryTree(0, 0)
		if err != nil {
			log.Println("Category tree error:", err)
			categoryTree = nil
		}
		tmpl := c.Echo().Renderer.(*Template)
		categoryTreeHTML, err := RenderCategoryTreeCached(tmpl.Templates, categoryTree)
		if err != nil {
			log.Println("Category tree render error:", err)
			categoryTreeHTML = ""
		}

		criticalCSS, err := parts.GetCriticalCSS()
		if err != nil {
			criticalCSS = ""
		}

		return c.Render(http.StatusOK, "products.html", map[string]interface{}{
			"Products":         products,
			"Title":            "Product Page - " + products[0]["name"].(string) + " - " + products[0]["sku"].(string) + " - Magento.GO",
			"CriticalCSS":      template.CSS(criticalCSS),
			"MediaUrl":         config.AppConfig.MediaUrl,
			"CategoryTreeHTML": template.HTML(categoryTreeHTML),
		})
	})

	// Register image routes in a separate file
	RegisterImageRoutes(e)
}
