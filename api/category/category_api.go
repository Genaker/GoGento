package category

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	categoryEntity "magento.GO/model/entity/category"
	repo "magento.GO/model/repository/category"
	"net/http"
	"strconv"
	"strings"
)

type CategoryWithAttributes struct {
	categoryEntity.Category
	Attributes map[string]map[string]interface{} `json:"attributes"`
}

// RegisterCategoryAPI registers the category API routes
func RegisterCategoryAPI(g *echo.Group, db *gorm.DB) {
	r := repo.GetCategoryRepository(db)
	fullHandler := func(c echo.Context) error {
		storeID := uint16(0)
		if sid := c.QueryParam("store_id"); sid != "" {
			if sidParsed, err := strconv.ParseUint(sid, 10, 16); err == nil {
				storeID = uint16(sidParsed)
			}
		}
		categories, err := r.FetchAllWithAttributes(storeID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"categories": categories,
			"total":      len(categories),
		})
	}
	g.GET("/categories", fullHandler)
	g.GET("/categories/full", fullHandler) // Alias route

	// New: Get category by IDssss
	g.GET("/category/:id", func(c echo.Context) error {
		storeID := uint16(0)
		if sid := c.QueryParam("store_id"); sid != "" {
			if sidParsed, err := strconv.ParseUint(sid, 10, 16); err == nil {
				storeID = uint16(sidParsed)
			}
		}
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid category id"})
		}
		cat, err := r.GetByIDsWithAttributes([]uint{uint(id)}, storeID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "category not found"})
		}
		return c.JSON(http.StatusOK, cat)
	})

	g.GET("/category/:ids/flat", func(c echo.Context) error {
		storeID := uint16(0)
		if sid := c.QueryParam("store_id"); sid != "" {
			if sidParsed, err := strconv.ParseUint(sid, 10, 16); err == nil {
				storeID = uint16(sidParsed)
			}
		}
		idsStr := c.Param("ids")
		idStrs := strings.Split(idsStr, ",")
		var ids []uint
		for _, s := range idStrs {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			id64, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				continue // skip invalid IDs
			}
			ids = append(ids, uint(id64))
		}
		if len(ids) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "no valid category ids"})
		}

		var results []CategoryWithAttributes
		for _, id := range ids {
			cat, flat, err := r.GetByIDWithAttributesAndFlat(id, storeID)
			if err != nil {
				continue // skip not found
			}
			results = append(results, CategoryWithAttributes{
				Category:   *cat,
				Attributes: flat,
			})
		}
		return c.JSON(http.StatusOK, results)
	})

	g.GET("/category/tree", func(c echo.Context) error {
		storeID := uint16(0)
		if sid := c.QueryParam("store_id"); sid != "" {
			if sidParsed, err := strconv.ParseUint(sid, 10, 16); err == nil {
				storeID = uint16(sidParsed)
			}
		}
		tree, err := r.BuildCategoryTree(storeID, 0)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, tree)
	})

	g.GET("/category/cache", func(c echo.Context) error {
		storeID := uint16(0)
		if sid := c.QueryParam("store_id"); sid != "" {
			if sidParsed, err := strconv.ParseUint(sid, 10, 16); err == nil {
				storeID = uint16(sidParsed)
			}
		}
		cats, ok := r.GetCacheCategory(storeID, 0)
		if !ok {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "no cache for store"})
		}
		return c.JSON(http.StatusOK, cats)
	})

	g.GET("/category/cache/:id", func(c echo.Context) error {
		storeID := uint16(0)
		if sid := c.QueryParam("store_id"); sid != "" {
			if sidParsed, err := strconv.ParseUint(sid, 10, 16); err == nil {
				storeID = uint16(sidParsed)
			}
		}
		idStr := c.Param("id")
		if idStr == "" {
			cats, ok := r.GetCacheCategory(storeID, 0)
			if !ok {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "no cache for store"})
			}
			return c.JSON(http.StatusOK, cats)
		}
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
		}
		cat, ok := r.GetCacheCategory(storeID, uint(id))
		if !ok {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found in cache"})
		}
		return c.JSON(http.StatusOK, cat)
	})
}

/* Usage Example (in your main or route setup):

import (
	"github.com/labstack/echo/v4"
	categoryapi "magento.magento.GO/api/category"
	"gorm.io/gorm"
)

func main() {
	e := echo.New()
	var db *gorm.DB // initialize your GORM DB
	categoryapi.RegisterCategoryAPI(e, db)
	e.Start(":8080")
}
*/
