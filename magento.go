package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"html/template"
	"log"
	categoryApi "magento.GO/api/category"
	productApi "magento.GO/api/product"
	salesApi "magento.GO/api/sales"
	"magento.GO/config"
	"magento.GO/core/cache"
	corelog "magento.GO/core/log"
	"magento.GO/core/registry"
	html "magento.GO/html"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var GlobalRegistry = registry.NewRegistry()
var GlobalCache = cache.GetInstance()

// Middleware to attach a request-isolated registry to each request
func RegistryMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		reqReg := registry.NewRequestRegistry()
		// Store the request start time in the request registry
		reqReg.Set("request_start", time.Now())
		// Attach to context
		c.Set("RequestRegistry", reqReg)
		return next(c)
	}
}

func getAuthMiddleware() echo.MiddlewareFunc {
	skipPaths := config.GetAuthSkipperPaths()
	skipper := func(c echo.Context) bool {
		path := c.Path()
		for _, skip := range skipPaths {
			if path == skip {
				return true
			}
		}
		return false
	}
	authType := os.Getenv("AUTH_TYPE")
	switch authType {
	case "key":
		apiKey := os.Getenv("API_KEY")
		return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
			Validator: func(key string, c echo.Context) (bool, error) {
				return key == apiKey, nil
			},
			Skipper: skipper,
		})
	default:
		return middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
			Validator: func(username, password string, c echo.Context) (bool, error) {
				return username == os.Getenv("API_USER") && password == os.Getenv("API_PASS"), nil
			},
			Skipper: skipper,
		})
	}
}

func CustomRecoverMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		defer func() {
			if r := recover(); r != nil {
				corelog.Error("Panic recovered: %v", r)
				err = echo.NewHTTPError(500, fmt.Sprintf("Internal Server Error: %v", r))
			}
		}()
		return next(c)
	}
}

func main() {
	corelog.Init()
	defer corelog.Close()
	config.LoadEnv()
	config.LoadAppConfig()
	// Initialize Redis
	config.InitRedis()
	redisStatus := "Redis not configured or not reachable, Redis caching disabled."
	if config.RedisClient != nil {
		err := config.RedisClient.Ping(config.RedisCtx()).Err()
		if err == nil {
			redisStatus = "Redis connection successful."
		} else {
			config.RedisClient = nil // Disable Redis if not reachable
			redisStatus = "Redis configured but not reachable, caching disabled."
		}
	}
	corelog.Info(redisStatus)

	db, err := config.NewDB()
	if err != nil {
		corelog.Fatal("failed to connect to DB: %v", err)
	}

	// Check DB connection
	sqldb, err := db.DB()
	if err != nil {
		corelog.Fatal("failed to get DB instance: %v", err)
	}
	if err := sqldb.Ping(); err != nil {
		corelog.Fatal("database connection failed: %v", err)
	}
	corelog.Info("Database connection successful.")

	e := echo.New()

	// Middleware to add cache control headers
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			log.Printf("Request path: %s", path) // Log the request path
			if strings.HasPrefix(path, "/static/") {
				log.Println("Setting cache headers") // Log when setting headers
				c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			return next(c)
		}
	})

	// Serve static files from the 'assets' directory at '/static/*'
	e.Static("/static", "assets")

	// Add timing middleware first
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Wrap the response writer
			w := &responseWriterWithTiming{
				ResponseWriter: c.Response().Writer,
				start:          start,
			}
			c.Response().Writer = w

			err := next(c)

			// If headers haven't been written yet, write them now
			if !w.headerWritten {
				duration := time.Since(start)
				msWithPrecision := float64(duration.Microseconds()) / 1000.0 // Convert to ms with decimals

				w.Header().Set("X-Page-Generation-Time-ms", fmt.Sprintf("%.3f", msWithPrecision))
				w.Header().Set("X-Page-Generation-Time-μs", strconv.FormatInt(duration.Microseconds(), 10))
				w.Header().Set("X-Page-Generation-Time", duration.String())
				w.Header().Set("Server-Timing", fmt.Sprintf("app;dur=%.3f;desc=\"Magento.GO Response Time\"", msWithPrecision))
				w.headerWritten = true
			}

			return err
		}
	})

	// Other middleware after timing
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.Decompress())

	// Attach the registry middleware
	e.Use(RegistryMiddleware)

	// Register the template renderer
	t := &html.Template{
		Templates: template.Must(template.New("").Funcs(html.GetTemplateFuncs()).ParseGlob("html/**/*.html")),
	}
	e.Renderer = t

	for _, tmpl := range t.Templates.Templates() {
		log.Println("Loaded template: %s", tmpl.Name())
	}

	apiGroup := e.Group("/api")
	apiGroup.Use(getAuthMiddleware())

	salesApi.RegisterSalesOrderGridRoutes(apiGroup, db)
	productApi.RegisterProductRoutes(apiGroup, db)
	categoryApi.RegisterCategoryAPI(apiGroup, db)

	// Health check endpoint (no auth required)
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{
			"status":  "healthy",
			"service": "GoGento",
			"version": "1.0.1",
		})
	})

	// Not Autorised HTML Routes
	html.RegisterProductHTMLRoutes(e, db)
	html.RegisterCategoryHTMLRoutes(e, db)
	html.RegisterHelloWorldRoute(e)

	fmt.Println(`
	╔═══════════════════════════════════════════════════════════════════════════════════════╗
	║                                                                                       ║
	║ ███╗   ███╗ █████╗  ██████╗ ███████╗███╗   ██╗████████╗ ██████╗    ██████╗  ██████╗   ║
	║ ████╗ ████║██╔══██╗██╔════╝ ██╔════╝████╗  ██║╚══██╔══╝██╔═══██╗   ██╔════╝ ██╔═══██╗ ║
	║ ██╔████╔██║███████║██║  ███╗█████╗  ██╔██╗ ██║   ██║   ██║   ██║   ██║  ███╗██║   ██║ ║
	║ ██║╚██╔╝██║██╔══██║██║   ██║██╔══╝  ██║╚██╗██║   ██║   ██║   ██║   ██║   ██║██║   ██║ ║
	║ ██║ ╚═╝ ██║██║  ██║╚██████╔╝███████╗██║ ╚████║   ██║   ╚██████╔╝   ╚██████╔╝╚██████╔╝ ║
	║ ╚═╝     ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═══╝   ╚═╝    ╚═════╝  ░░ ╚═════╝  ╚═════╝  ║
	║                                                                                       ║
	╚═══════════════════════════════════════════════════════════════════════════════════════╝
	Magento GO(GoGento) server V1.0.1
	`)

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := 500
		msg := err.Error()
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			msg = fmt.Sprintf("%v", he.Message)
		}
		corelog.Error("HTTP error: %d %s %s - %s", code, c.Request().Method, c.Request().URL.Path, msg)
		c.String(code, msg)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on :%s", port)
	e.Logger.Fatal(e.Start(":" + port))

}

type responseWriterWithTiming struct {
	http.ResponseWriter
	start         time.Time
	headerWritten bool
}

func (r *responseWriterWithTiming) WriteHeader(code int) {
	if !r.headerWritten {
		duration := time.Since(r.start)
		msWithPrecision := float64(duration.Microseconds()) / 1000.0 // Convert to ms with decimals

		r.Header().Set("X-Page-Generation-Time-ms", fmt.Sprintf("%.3f", msWithPrecision))
		r.Header().Set("Server-Timing", fmt.Sprintf("app;dur=%.3f;desc=\"Magento.GO Response Time\"", msWithPrecision))
		r.headerWritten = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseWriterWithTiming) Write(b []byte) (int, error) {
	if !r.headerWritten {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(b)
}
