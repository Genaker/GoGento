package html

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"html/template"
	"log"
	"magento.GO/config"
	parts "magento.GO/html/parts"
	categoryRepo "magento.GO/model/repository/category"
	productRepo "magento.GO/model/repository/product"
	"net/http"
	"strconv"
	"time"
)

// PaginationData holds all pagination-related information
type PaginationData struct {
	Page        int
	Limit       int
	TotalItems  int
	TotalPages  int
	PageNumbers []int
	PrevPage    int
	NextPage    int
}

// calculatePagination creates pagination data based on total items and request parameters
func calculatePagination(c echo.Context, totalItems int) PaginationData {
	// Get pagination parameters
	limit := 20
	if lStr := c.QueryParam("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	page := 1
	if pStr := c.QueryParam("p"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			page = p
		}
	}

	// Calculate total pages
	totalPages := (totalItems + limit - 1) / limit
	if page > totalPages {
		page = totalPages
	}
	if page < 1 {
		page = 1
	}

	// Calculate page numbers to show
	maxPagesToShow := 5 // Show at most 5 pages
	startPage := page - maxPagesToShow/2
	if startPage < 1 {
		startPage = 1
	}
	endPage := startPage + maxPagesToShow - 1
	if endPage > totalPages {
		endPage = totalPages
		// Adjust start page if we're near the end
		startPage = endPage - maxPagesToShow + 1
		if startPage < 1 {
			startPage = 1
		}
	}

	var pageNumbers []int
	for i := startPage; i <= endPage; i++ {
		pageNumbers = append(pageNumbers, i)
	}

	// Calculate prev/next pages
	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 1
	}
	nextPage := page + 1
	if nextPage > totalPages {
		nextPage = totalPages
	}

	return PaginationData{
		Page:        page,
		Limit:       limit,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		PageNumbers: pageNumbers,
		PrevPage:    prevPage,
		NextPage:    nextPage,
	}
}

// RegisterCategoryHTMLRoutes registers HTML routes for category rendering
func RegisterCategoryHTMLRoutes(e *echo.Echo, db *gorm.DB) {
	repo := categoryRepo.GetCategoryRepository(db)
	prodRepo := productRepo.GetProductRepository(db)

	e.GET("/category/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid category ID")
		}

		start := time.Now()
		cat, flat, err := repo.GetByIDWithAttributesAndFlat(uint(id), 0)
		log.Printf("GetByIDWithAttributesAndFlat took %s", time.Since(start))
		if err != nil || cat == nil {
			return c.String(http.StatusNotFound, "Category not found")
		}

		// Extract product IDs from cat.Products
		var productIDs []uint
		for _, cp := range cat.Products {
			productIDs = append(productIDs, cp.ProductID)
		}

		// Get pagination data
		pagination := calculatePagination(c, len(productIDs))

		// Apply pagination to product IDs
		startIdx := (pagination.Page - 1) * pagination.Limit
		endIdx := startIdx + pagination.Limit
		if startIdx > len(productIDs) {
			startIdx = len(productIDs)
		}
		if endIdx > len(productIDs) {
			endIdx = len(productIDs)
		}
		pagedProductIDs := productIDs[startIdx:endIdx]

		// Fetch product data with attributes
		var products []map[string]interface{}
		if len(pagedProductIDs) > 0 {
			start := time.Now()
			flatProducts, err := prodRepo.FetchWithAllAttributesFlatByIDs(pagedProductIDs)
			log.Printf("FetchWithAllAttributesFlatByIDs took %s", time.Since(start))
			if err == nil {
				for _, id := range pagedProductIDs {
					if prod, ok := flatProducts[id]; ok {
						products = append(products, prod)
					}
				}
			}
		}

		// Get category tree
		tmpl := c.Echo().Renderer.(*Template)
		start = time.Now()
		categoryTree, err := repo.BuildCategoryTree(0, 0)
		log.Printf("BuildCategoryTree took %s", time.Since(start))
		var categoryTreeHTML string
		if err == nil {
			start = time.Now()
			categoryTreeHTML, err = RenderCategoryTreeCached(tmpl.Templates, categoryTree)
			log.Printf("RenderCategoryTreeCached took %s", time.Since(start))
			if err != nil {
				log.Println("Category tree render error:", err)
				categoryTreeHTML = ""
			}
		}

		// Get critical CSS
		criticalCSS, err := parts.GetCriticalCSSCached()
		if err != nil {
			criticalCSS = ""
		}

		// Get title
		var title string
		if nameMap, ok := flat["name"]; ok {
			if val, ok := nameMap["Value"].(string); ok {
				title = val
			}
		}
		if title == "" {
			title = fmt.Sprintf("%v", cat.EntityID) // fallback to ID if name is missing
		}
		title = "Category Page - " + title + " - Magento.GO"

		// Render template with all data
		return c.Render(http.StatusOK, "parts/category_layout.html", map[string]interface{}{
			"Category":         cat,
			"Attributes":       flat,
			"Title":            title,
			"Products":         products,
			"CriticalCSS":      template.CSS(criticalCSS),
			"CategoryTreeHTML": template.HTML(categoryTreeHTML),
			"MediaUrl":         config.AppConfig.MediaUrl,
			"Page":             pagination.Page,
			"TotalPages":       pagination.TotalPages,
			"Limit":            pagination.Limit,
			"PageNumbers":      pagination.PageNumbers,
			"PrevPage":         pagination.PrevPage,
			"NextPage":         pagination.NextPage,
		})
	})
}
