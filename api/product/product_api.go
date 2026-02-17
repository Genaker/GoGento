package product

import (
	"net/http"
	//"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	//"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"

	productRepository "magento.GO/model/repository/product"
	productService "magento.GO/service/product"
)

// Handler for /flat and /full endpoints
func flatProductsHandler(repo *productRepository.ProductRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		flatProducts, err := repo.FetchWithAllAttributesFlat()
		duration := time.Since(start).Milliseconds()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error(), "request_duration_ms": duration})
		}
		c.Response().Header().Set("X-Request-Duration-ms", strconv.FormatInt(duration, 10))
		return c.JSON(http.StatusOK, echo.Map{
			"products":            flatProducts,
			"count":               len(flatProducts),
			"request_duration_ms": duration,
		})
	}
}

func RegisterProductRoutes(api *echo.Group, db *gorm.DB) {
	repo := productRepository.GetProductRepository(db)
	service := productService.NewProductService(repo)
	g := api.Group("/products")

	g.GET("", func(c echo.Context) error {
		start := time.Now()
		products, err := service.ListProducts()
		duration := time.Since(start).Milliseconds()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error(), "request_duration_ms": duration})
		}
		c.Response().Header().Set("X-Request-Duration-ms", strconv.FormatInt(duration, 10))
		return c.JSON(http.StatusOK, echo.Map{
			"products":            products,
			"count":               len(products),
			"request_duration_ms": duration,
		})
	})

	g.GET("/:id", func(c echo.Context) error {
		start := time.Now()
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id"})
		}
		product, err := service.GetProduct(uint(id))
		duration := time.Since(start).Milliseconds()
		if err != nil {
			return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error(), "request_duration_ms": duration})
		}
		c.Response().Header().Set("X-Request-Duration-ms", strconv.FormatInt(duration, 10))
		return c.JSON(http.StatusOK, echo.Map{"product": product, "request_duration_ms": duration})
	})

	g.POST("", func(c echo.Context) error {
		start := time.Now()
		var product productService.ProductInput
		if err := c.Bind(&product); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}
		err := service.CreateProduct(&product)
		duration := time.Since(start).Milliseconds()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error(), "request_duration_ms": duration})
		}
		c.Response().Header().Set("X-Request-Duration-ms", strconv.FormatInt(duration, 10))
		return c.JSON(http.StatusCreated, echo.Map{"product": product, "request_duration_ms": duration})
	})

	g.PUT("/:id", func(c echo.Context) error {
		start := time.Now()
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id"})
		}
		var product productService.ProductInput
		if err := c.Bind(&product); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}
		err = service.UpdateProduct(uint(id), &product)
		duration := time.Since(start).Milliseconds()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error(), "request_duration_ms": duration})
		}
		c.Response().Header().Set("X-Request-Duration-ms", strconv.FormatInt(duration, 10))
		return c.JSON(http.StatusOK, echo.Map{"product": product, "request_duration_ms": duration})
	})

	g.DELETE("/:id", func(c echo.Context) error {
		start := time.Now()
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id"})
		}
		err = service.DeleteProduct(uint(id))
		duration := time.Since(start).Milliseconds()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error(), "request_duration_ms": duration})
		}
		c.Response().Header().Set("X-Request-Duration-ms", strconv.FormatInt(duration, 10))
		return c.NoContent(http.StatusNoContent)
	})

	g.GET("/flat", flatProductsHandler(repo))
	g.GET("/full", flatProductsHandler(repo))

	g.GET("/flat/:ids", func(c echo.Context) error {
		start := time.Now()
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
		duration := time.Since(start).Milliseconds()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error(), "request_duration_ms": duration})
		}

		var result []map[string]interface{}
		for _, id := range ids {
			if prod, ok := flatProducts[id]; ok {
				result = append(result, prod)
			}
		}

		c.Response().Header().Set("X-Request-Duration-ms", strconv.FormatInt(duration, 10))
		return c.JSON(http.StatusOK, echo.Map{
			"products":            result,
			"count":               len(result),
			"request_duration_ms": duration,
		})
	})

}
