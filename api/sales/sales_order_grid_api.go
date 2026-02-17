package sales

import (
	"encoding/json"
	"net/http"
	//"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	//"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"

	"magento.GO/config"
	"magento.GO/model/entity/sales"
)

// RegisterSalesOrderGridRoutes registers the routes for SalesOrderGrid CRUD operations with basic auth
func RegisterSalesOrderGridRoutes(api *echo.Group, db *gorm.DB) {
	g := api.Group("/orders")

	g.GET("", func(c echo.Context) error {
		cacheKey := "orders:all"
		ctx := config.RedisCtx()

		// Only use Redis if configured
		if config.RedisClient != nil {
			if cached, err := config.RedisClient.Get(ctx, cacheKey).Result(); err == nil {
				var orders []sales.SalesOrderGrid
				if err := json.Unmarshal([]byte(cached), &orders); err == nil {
					return c.JSON(http.StatusOK, orders)
				}
			}
		}

		var orders []sales.SalesOrderGrid
		if err := db.Find(&orders).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}

		// Save to Redis if configured
		if config.RedisClient != nil {
			if data, err := json.Marshal(orders); err == nil {
				config.RedisClient.Set(ctx, cacheKey, data, 5*time.Minute)
			}
		}

		return c.JSON(http.StatusOK, orders)
	})

	g.GET("/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id"})
		}
		var order sales.SalesOrderGrid
		if err := db.First(&order, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.JSON(http.StatusNotFound, echo.Map{"error": "not found"})
			}
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, order)
	})

	g.POST("", func(c echo.Context) error {
		var order sales.SalesOrderGrid
		if err := c.Bind(&order); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}
		order.CreatedAt = ptrTime(time.Now())
		if err := db.Create(&order).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusCreated, order)
	})

	g.PUT("/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id"})
		}
		var order sales.SalesOrderGrid
		if err := db.First(&order, id).Error; err != nil {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "not found"})
		}
		if err := c.Bind(&order); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}
		order.UpdatedAt = ptrTime(time.Now())
		if err := db.Save(&order).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, order)
	})

	g.DELETE("/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id"})
		}
		if err := db.Delete(&sales.SalesOrderGrid{}, id).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}
		return c.NoContent(http.StatusNoContent)
	})
}

// ptrTime is a helper to get a pointer to a time.Time
func ptrTime(t time.Time) *time.Time {
	return &t
}

/*
API Endpoints (all require Basic Auth):
GET    /api/orders         - List all orders
GET    /api/orders/:id     - Get order by ID
POST   /api/orders         - Create new order
PUT    /api/orders/:id     - Update order by ID
DELETE /api/orders/:id     - Delete order by ID

See Echo routing docs: https://echo.labstack.com/docs/routing
*/
