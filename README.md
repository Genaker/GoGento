# Magento Go API and Frontend

[![Go Version](https://img.shields.io/github/go-mod/go-version/Genaker/GoGento)](https://golang.org/dl/)
[![Go Report Card](https://goreportcard.com/badge/github.com/Genaker/GoGento)](https://goreportcard.com/report/github.com/Genaker/GoGento)
[![CI](https://github.com/Genaker/GoGento/actions/workflows/ci.yml/badge.svg)](https://github.com/Genaker/GoGento/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub issues](https://img.shields.io/github/issues/Genaker/GoGento)](https://github.com/Genaker/GoGento/issues)
[![GitHub stars](https://img.shields.io/github/stars/Genaker/GoGento)](https://github.com/Genaker/GoGento/stargazers)

A fully functional REST API and HTTP server for Magento using Go, Echo, and GORM.

## The world’s fastest framework for building e-Commerce MAGENTO websites!

![image](https://github.com/user-attachments/assets/eaacc9e0-e497-4d3c-a4d9-faeadc7fd6e5)


## Features
- Echo web server with RESTful routing
- Basic authentication for all endpoints
- GORM ORM for MySQL
- Modular structure for easy extension
- **Concurrent-safe global product cache for fast flat product queries**
- **Flexible product API: fetch all or specific products, with EAV attributes flattened**

## Quick Start

### Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/Genaker/GoGento.git
cd GoGento

# Start all services (MySQL, Redis, and the application)
make docker-up
# or
docker-compose up -d

# View logs
docker-compose logs -f app
```

The API will be available at `http://localhost:8080`

### Using Make (Local Development)

```bash
# Install dependencies
make deps

# Copy environment file and configure
cp .env.example .env
# Edit .env with your database credentials

# Run the server
make run
```

### Manual Setup

See the detailed installation instructions below for manual setup without Docker or Make.

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Security

For reporting security vulnerabilities, please see [SECURITY.md](SECURITY.md).

## Directory Structure
```
magento.GO/
  main.go
  api/
    product/
      product_api.go
    sales/
      order_api.go
  model/
    entity/
      product/
        product.go
        product_attribute.go
        product_link.go
      category/
        category.go
        category_product.go
      sales/
        order.go
        order_grid.go
    repository/
      product/
        product_repository.go
      sales/
        order_repository.go
  service/
    product/
      product_service.go
    sales/
      order_service.go
  config/
    db.go
    env.go
  go.mod
  go.sum
  README.md
```

## Install Go (if not already installed)
On Ubuntu/Debian, you can install Go with:
```
sudo apt update
sudo apt install golang-go
```
Or using snap:
```
sudo snap install go
```
After installation, check your Go version:
```
go version
```

## Environment Variables
Set these in a `.env` file or your environment:
```
MYSQL_USER=magento
MYSQL_PASS=magento
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_DB=magento
API_USER=admin
API_PASS=secret
REDIS_ADDR=""
REDIS_PASS=""
PORT=8080
```

## Install dependencies
```
cd GO
export GO111MODULE=on
go mod tidy
```

## Run the API
```
go run magento.go
```

## Endpoints (all require Basic Auth)
- `GET    /api/orders`         - List all orders
- `GET    /api/orders/:id`     - Get order by ID
- `POST   /api/orders`         - Create new order
- `PUT    /api/orders/:id`     - Update order by ID
- `DELETE /api/orders/:id`     - Delete order by ID
- `GET    /api/products/flat`  - List all flat products (optionally for a store)
- `GET    /api/products/flat/:ids` - List flat products for given comma-separated IDs (optionally for a store)

### Product Flat API Usage
- `GET /api/products/flat` returns all products, flattened (with EAV attributes as keys)
- `GET /api/products/flat/1,2,3` returns only products with IDs 1, 2, 3
- Both endpoints accept an optional store ID (see code for details)

**Example:**
```
GET /api/products/flat/1,2,3
Response:
{
  "products": [
    {
      "entity_id": 1,
      "sku": "foo",
      ...,
      "stock_item": {
        "item_id": 123,
        "qty": 10.0,
        "is_in_stock": 1,
        "min_qty": 0.0,
        "max_sale_qty": 100.0,
        "manage_stock": 1,
        "website_id": 1
      },
      "index_prices": [
        {
          "entity_id": 1,
          "customer_group_id": 0,
          "website_id": 1,
          "tax_class_id": 2,
          "price": 99.99,
          "final_price": 89.99,
          "min_price": 89.99,
          "max_price": 99.99,
          "tier_price": 0.0
        }
      ]
    },
    ...
  ],
  "count": 3,
  "request_duration_ms": 12
}
```

## Authentication
Use HTTP Basic Auth with `API_USER` and `API_PASS`.

## Extending the App (.cursor documentation)

### Add New Magento Table as an API Endpoint

1. **Generate GORM Model**
   - Use the database.mdc rules to generate a GORM model for your Magento table (see `.cursor` documentation or ask the AI: `@cursor please generate GORM model for Magento 2 table [table_name] with relationships and examples`).
   - Place the model in `magento.GO/model/entity/`.

2. **Create API Handler**
   - Create a new file in `magento.GO/api/` (e.g., `product_api.go`).
   - Follow the structure in `sales_order_grid_api.go` for CRUD endpoints.
   - Use Echo's routing and middleware as described in the [Echo Routing Docs](https://echo.labstack.com/docs/routing).

3. **Register Routes**
   - In `magento.GO/main.go`, import your new API handler and register its routes with the Echo instance.

4. **Test Your Endpoint**
   - Use tools like curl or Postman to test your new API endpoint.
   - All endpoints are protected by HTTP Basic Auth.

### Echo Best Practices
- Use Echo's `Group` feature to organize endpoints and apply middleware (see [Echo Routing](https://echo.labstack.com/docs/routing)).
- Use context (`c echo.Context`) for request/response handling.
- Add middleware for logging, recovery, CORS, etc., as needed.
- See the [Echo Cookbook](https://echo.labstack.com/docs/cookbook) for advanced patterns (e.g., grouping, middleware, deployment).

### GORM Model Generation (database.mdc)
- Always include relationships and the `TableName()` method.
- Place each model in its own file under `model/entity/`.
- Include CRUD usage examples as comments in the model file.
- Use the provided SQL query to get table structures for model generation.

### References
- [Echo Documentation](https://echo.labstack.com/docs)
- [Echo Routing](https://echo.labstack.com/docs/routing)
- [GORM Documentation](https://gorm.io/)
- [database.mdc rules](see your .cursor documentation or ask the AI)

## Repository and Service Layers

This project follows Go best practices by separating data access and business logic into repository and service layers:

### Repository Layer (`model/repository/`)
- Handles all database access (CRUD, queries) for each entity.
- Example: `SalesOrderGridRepository` provides methods like `FindAll`, `FindByID`, `Create`, `Update`, `Delete`.
- Keeps SQL/GORM logic out of your API and business logic.

### Service Layer (`service/`)
- Handles business logic and orchestration.
- Calls repository methods to access data.
- Example: `SalesOrderGridService` provides methods like `ListOrders`, `GetOrder`, `CreateOrder`, `UpdateOrder`, `DeleteOrder`.
- Keeps business rules out of your API handlers.

### Example Usage
```go
import (
    "magento.GO/model/repository"
    "magento.GO/service"
)

repo := repository.NewSalesOrderGridRepository(db)
service := service.NewSalesOrderGridService(repo)
orders, err := service.ListOrders()
```

### Summary Table
| Layer       | Directory              | Responsibility                        |
|-------------|-----------------------|----------------------------------------|
| Entity      | model/entity/         | Structs, GORM tags                     |
| Repository  | model/repository/     | DB access, queries, raw SQL            |
| Service     | service/              | Business logic, orchestration          |
| API/Handler | api/                  | HTTP, request/response, call services  |

This structure makes your codebase easier to maintain, test, and extend.

## References
- [Echo Routing](https://echo.labstack.com/docs/routing)
- [GORM](https://gorm.io/)

## Domain-Based Organization

This project organizes models, repositories, and services by domain (e.g., product, sales/order, category) for clarity and scalability.

### Example Structure
```
magento.GO/
  model/
    entity/
      product/
        product.go
        product_attribute.go
        product_link.go
      category/
        category.go
        category_product.go
      sales/
        order.go
        order_grid.go
  model/
    repository/
      product/
        product_repository.go
      sales/
        order_repository.go
  service/
    product/
      product_service.go
    sales/
      order_service.go
```

### Importing with Aliases
When two packages have the same name (e.g., `product` for both entity and repository), use import aliases to avoid conflicts:

```go
import (
    prodentity "magento.GO/model/entity/product"
    productrepo "magento.GO/model/repository/product"
)

// Usage:
var p prodentity.Product
repo := productrepo.NewProductRepository(db)
```

### Benefits
- **Scalability:** Add new domains without clutter.
- **Clarity:** Quickly find all code related to a domain.
- **Consistency:** Mirrors your API and business logic structure.

## Handling Relationships: Products with Categories

This project demonstrates how to efficiently handle relationships (e.g., products with their categories) using GORM, following a clean repository-service pattern.

### Model Structure

**Product model references Category, StockItem, and ProductIndexPrices:**
```go
import cat "magento.GO/model/entity/category"

type Product struct {
    // ... other fields ...
    Categories []cat.Category `gorm:"many2many:catalog_category_product;joinForeignKey:ProductID;joinReferences:CategoryID"`
    StockItem StockItem `gorm:"foreignKey:EntityID;references:ProductID"`
    ProductIndexPrices []ProductIndexPrice `gorm:"foreignKey:EntityID;references:EntityID"`
}
```

### Repository Layer: Preloading Relationships

The repository is responsible for all data access, including loading related entities using GORM's `Preload`:

```go
func (r *ProductRepository) FindAll() ([]productEntity.Product, error) {
    var products []productEntity.Product
    err := r.db.Preload("Categories").Find(&products).Error
    return products, err
}

func (r *ProductRepository) FindByID(id uint) (*productEntity.Product, error) {
    var product productEntity.Product
    err := r.db.Preload("Categories").First(&product, id).Error
    if err != nil {
        return nil, err
    }
    return &product, nil
}
```

### Service Layer: Orchestration Only

The service layer simply calls the repository and does not handle relationship logic:

```go
func (s *ProductService) ListProducts() ([]product.Product, error) {
    return s.repo.FindAll()
}
```

### API Handler: Returning Nested Data

The API handler returns products with their related categories included:

```go
g.GET("", func(c echo.Context) error {
    start := time.Now()
    products, err := service.ListProducts()
    duration := time.Since(start).Milliseconds()
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error(), "request_duration_ms": duration})
    }
    c.Response().Header().Set("X-Request-Duration-ms", strconv.FormatInt(duration, 10))
    return c.JSON(http.StatusOK, echo.Map{
        "products": products,
        "count": len(products),
        "request_duration_ms": duration,
    })
})
```

### Best Practices
- **Repository Layer:** Handles all GORM queries and relationship loading (e.g., `Preload`).
- **Service Layer:** Handles business logic and orchestration, not DB details.
- **API Handler:** Returns the nested data as needed for the client.
- **Model Layer:** Defines relationships using GORM tags.

### References
- [GORM Preload Documentation](https://gorm.io/docs/preload.html)
- [Echo Grouping and Middleware](https://echo.labstack.com/docs/guide#grouping-routes)

## Local Go Cache Database for Main Entities

All main entities (such as categories, products, attributes, etc.) must be stored in a local Go cache database. This cache acts as a persistent layer, ensuring that the application can continue to operate efficiently and reliably, even if the primary data source (such as a remote database or API) is temporarily unavailable.

### Why Use a Local Cache?
- **Performance:** Reduces latency by serving frequently accessed data from memory or local storage.
- **Reliability:** Allows the application to function even during outages or slowdowns of the main data source.
- **Persistence:** Ensures that critical data is not lost and can be quickly restored on restart.

### Implementation Notes
- The cache should be updated whenever entities are created, updated, or deleted.
- On application startup, the cache should be loaded from persistent storage if available.
- The cache can be implemented using Go's built-in data structures, with optional serialization to disk for persistence.


## GORM SQL Query Logging

You can control GORM SQL query logging using the `GORM_LOG` environment variable:

- To **enable** SQL logging (default):
  - Omit `GORM_LOG` or set it to any value other than `off`.
- To **disable** SQL logging:
  ```
  GORM_LOG=off
  ```

All GORM logs are output to the console (stdout) using Go's standard logger. This is configured in `config/db.go`:

```go
import (
    "log"
    "os"
    "gorm.io/gorm/logger"
    "time"
)

logMode := logger.Info
if os.Getenv("GORM_LOG") == "off" {
    logMode = logger.Silent
}

gormLogger := logger.New(
    log.New(os.Stdout, "\r\n", log.LstdFlags),
    logger.Config{
        SlowThreshold: time.Second,
        LogLevel:      logMode,
        Colorful:      true,
    },
)
```

## Fetching and Flattening EAV Attributes

- The repository supports fetching all product EAV attributes for a specific `store_id` (default is `0` for global).
- Use `FetchWithAllAttributes(storeID ...uint16)` and `FetchWithAllAttributesFlat(storeID ...uint16)` to get products with all EAV attributes preloaded and flattened.
- The flattening function can use attribute codes (from the `eav_attribute` table) as keys for a more readable API output.
- **The flat product result also includes:**
  - `stock_item`: Inventory/stock data for the product (qty, is_in_stock, min_qty, max_sale_qty, manage_stock, website_id, etc.)
  - `index_prices`: Array of price index records for the product (entity_id, customer_group_id, website_id, tax_class_id, price, final_price, min_price, max_price, tier_price)



## Performance Metrics (With Global Cache)

The following results were obtained using ApacheBench (ab) to benchmark the `/product/1` endpoint **with global cache enabled**:

```
ab -c 100 -n 1000 http://magento.go:8080/product/1

Concurrency Level:      100
Time taken for tests:   3.363 seconds
Complete requests:      1000
Failed requests:        0
Requests per second:    297.37 [#/sec] (mean)
Time per request:       336.280 [ms] (mean)
Time per request:       3.363 [ms] (mean, across all concurrent requests)
Transfer rate:          1649.19 [Kbytes/sec] received

Percentage of the requests served within a certain time (ms)
  50%    284
  66%    402
  75%    425
  80%    437
  90%    575
  95%    703
  98%    843
  99%    973
 100%   1429 (longest request)
```

**Interpretation:**
- With global cache enabled, the API handled 1000 requests at a concurrency level of 100 with no failed requests.
- Average requests per second: **297.4** (vs. 79.6 without cache)
- Median response time: **284 ms** (vs. 1231 ms without cache)
- 99% of requests completed within **973 ms** (vs. 1693 ms without cache).

> _Enabling global cache resulted in a **~4x increase in throughput** and a **~4x reduction in median response time** for this endpoint._

> _Test environment: All benchmarks were run on a 2 vCPU AWS T4 instance._

### Single Product Generation Time

The following results were obtained using ApacheBench (ab) to benchmark the `/product/1` endpoint with a concurrency level of 1 (single request at a time):

```
ab -c 1 -n 10 http://magento.go:8080/product/1

Concurrency Level:      1
Time taken for tests:   0.011 seconds
Complete requests:      10
Failed requests:        0
Requests per second:    912.83 [#/sec] (mean)
Time per request:       1.095 [ms] (mean)
Time per request:       1.095 [ms] (mean, across all concurrent requests)
Transfer rate:          5062.44 [Kbytes/sec] received

Percentage of the requests served within a certain time (ms)
  50%      1
  66%      1
  75%      1
  80%      1
  90%      2
  95%      2
  98%      2
  99%      2
 100%      2 (longest request)
```

**Interpretation:**
- With global cache enabled and a single request at a time, the API can serve a product page in about **1 ms** on average.
- This demonstrates the extremely low latency possible for individual product requests when using the cache.

## Technical Approach: Go Global Cache

The global cache in this API is designed to dramatically improve performance for frequently accessed product data, especially for flat product queries. Here's how it works:

- **Concurrent-Safe In-Memory Map:**
  - The cache is implemented as a Go `map[uint]Product` (or similar), where the key is the product ID and the value is the flattened product struct or map.
  - This map is stored in a package-level variable, making it accessible throughout the application.

- **Thread Safety with `sync.RWMutex`:**
  - To ensure safe concurrent access (reads and writes) from multiple goroutines, the cache is protected by a `sync.RWMutex`.
  - Read operations (`RLock`) can happen in parallel, while write operations (`Lock`) are exclusive.

- **Cache Population:**
  - On the first request (or on demand), the cache is populated by loading all relevant product data from the database and flattening EAV attributes.
  - The cache can be refreshed or invalidated as needed (e.g., after product updates).

- **Cache Usage:**
  - When a flat product query is received, the handler first checks the cache.
  - If the requested product(s) are present, they are returned directly from memory, bypassing the database and EAV flattening logic.
  - This results in much faster response times and higher throughput, as shown in the performance metrics above.

- **Benefits:**
  - **Significant speedup** for repeated queries.
  - **Reduced database load** and lower latency.
  - **Safe for concurrent use** in a high-traffic API.

- **Implementation Example:**
  ```go
  var (
      flatProductCache = make(map[uint]map[string]interface{})
      flatProductCacheMutex sync.RWMutex
  )

  func GetFlatProductFromCache(id uint) (map[string]interface{}, bool) {
      flatProductCacheMutex.RLock()
      defer flatProductCacheMutex.RUnlock()
      prod, ok := flatProductCache[id]
      return prod, ok
  }
  ```

> _This approach leverages Go's strengths in concurrency and memory management to provide a robust, high-performance caching layer for the API._

## Tailwind CSS: Install & Compile Minimal Build

To use a minimal, production-ready Tailwind CSS build:

1. **Install Tailwind CSS (v3):**
   ```sh
   npm install -D tailwindcss@3
   ```

2. **Create an input CSS file:**
   Create a file named `input.css` in your project root with:
   ```css
   @tailwind base;
   @tailwind components;
   @tailwind utilities;
   ```

3. **Build your CSS for production:**
   ```sh
  npx tailwindcss -i ./input.css -o ./assets/tailwind.min.css --minify --content './html/**/*.html'
   ```
   - This will generate a minimal CSS file containing only the classes used in your HTML templates.

4. **Reference the output in your HTML:**
   ```html
   <link href="/static/tailwind.min.css" rel="stylesheet">
   ```

For more details, see [Tailwind CSS documentation](https://tailwindcss.com/docs/installation) and [optimizing for production](https://tailwindcss.com/docs/optimizing-for-production).

## Global Registry

GoGento provides a thread-safe, application-wide global registry for sharing data across your application.

**Usage:**
```go
import "magento.GO/core/registry"

var GlobalRegistry = registry.NewRegistry()

// Set a global value
GlobalRegistry.SetGlobal("site_name", "MySite")

// Get a global value
site, ok := GlobalRegistry.GetGlobal("site_name")

// Delete a global value
GlobalRegistry.DeleteGlobal("site_name")
```

## Global Cache

GoGento includes a thread-safe, application-wide global cache for storing frequently accessed data.

**Usage:**
```go
import "magento.GO/core/cache"

var GlobalCache = cache.GetInstance()

// Set a value
GlobalCache.Set("user_123_profile", userProfile)

// Get a value
val, ok := GlobalCache.Get("user_123_profile")

// Delete a value
GlobalCache.Delete("user_123_profile")
```

## Request-Isolated Registry

For per-request data, use the request-isolated registry:

```go
reqReg := registry.NewRequestRegistry()
reqReg.Set("user_id", 123)
userID, ok := reqReg.Get("user_id")
reqReg.Delete("user_id")
```

## Hello World Handler and Template Profiling

GoGento includes a Hello World handler that demonstrates:
- Using the request registry to track request start time
- Passing execution time and template compilation time to the template


## Singleton Pattern: Repository, Cache, and Registry

GoGento uses the singleton pattern for its repositories, cache, and registry to ensure that only one instance exists and is shared across the application. This improves performance, ensures thread safety, and avoids redundant data loading.

### Singleton Repository

Repositories are created as singletons so that their in-memory caches (if any) are shared across all requests and handlers.

**Example:**
```go
import "magento.GO/model/repository/product"

// Get the singleton instance
repo := product.GetProductRepository(db)
```
- The first call creates the repository; subsequent calls return the same instance.

### Singleton Cache

The global cache is a singleton, ensuring all parts of the application use the same cache instance.

**Example:**
```go
import "magento.GO/core/cache"

cache := cache.GetInstance()
cache.Set("foo", 123)
val, ok := cache.Get("foo")
```

### Singleton Registry

The global registry is also a singleton, so global data is always shared.

**Example:**
```go
import "magento.GO/core/registry"

reg := registry.GetInstance() // if implemented, or use a global variable
reg.SetGlobal("key", "value")
```

## Shared (Global) vs. Isolated (Per-Request) Registry and Cache

- **Shared (Global) Registry/Cache:**
  - Accessible from anywhere in the application.
  - Data persists for the application's lifetime.
  - Use for configuration, global flags, or data that should be available to all requests.
  - **Example:**
    ```go
    GlobalRegistry.SetGlobal("site_mode", "production")
    GlobalCache.Set("product_1", productData)
    ```

- **Isolated (Per-Request) Registry/Cache:**
  - Created fresh for each request (e.g., via middleware).
  - Data is only visible within the current request and is discarded after the request ends.
  - Use for request-specific data, such as timing, user context, or temporary values.
  - **Example:**
    ```go
    reqReg := registry.NewRequestRegistry()
    reqReg.Set("user_id", 42)
    userID, ok := reqReg.Get("user_id")
    ```

### Summary Table
| Type                | Lifetime         | Scope         | Thread Safe | Usage Example                  |
|---------------------|-----------------|--------------|-------------|-------------------------------|
| Global Registry     | Application     | All requests | Yes         | GlobalRegistry.SetGlobal(...)  |
| Request Registry    | Per-request     | One request  | N/A         | reqReg.Set(...)                |
| Global Cache        | Application     | All requests | Yes         | GlobalCache.Set(...)           |
| Request Cache       | Per-request     | One request  | N/A         | reqCache.Set(...) (if needed)  |
| Singleton Repo      | Application     | All requests | Yes         | product.GetProductRepository() |

---

This pattern ensures high performance, data consistency, and safe concurrent access throughout your GoGento application.

## Running as a Daemon / Background Process

### 1. Run in Background with nohup

You can run the Go application in the background using `nohup` so it continues running after you log out:

```sh
nohup go run magento.go > output.log 2>&1 &
```
Or, if you have built a binary:
```sh
nohup ./magento > output.log 2>&1 &
```
- The process will keep running after you log out.
- Output is written to `output.log`.

### 2. Run as a systemd Service (Recommended for Production)

1. **Build your binary:**
   ```sh
   go build -o magento
   ```
2. **Create a systemd service file** `/etc/systemd/system/magento.service`:
   ```ini
   [Unit]
   Description=Magento Go API

   [Service]
   ExecStart=/var/www/html/react-luma/magento.go/magento
   WorkingDirectory=/var/www/html/react-luma/magento.go
   Restart=always
   User=youruser
   Environment=PORT=8080

   [Install]
   WantedBy=multi-user.target
   ```
   Replace `youruser` with the user you want to run the service as.

3. **Enable and start the service:**
   ```sh
   sudo systemctl daemon-reload
   sudo systemctl start magento
   sudo systemctl enable magento
   ```

4. **Check status:**
   ```sh
   sudo systemctl status magento
   ```

---

## Cron Jobs and CLI Usage

This project supports scheduled and on-demand background jobs using the [robfig/cron](https://github.com/robfig/cron) library and a modular CLI.

### Directory Structure
- `cron/jobs/` — Individual job implementations (e.g., `product_json.go`)
- `cron/sheduler.go` — Cron scheduler setup and job registration
- `cmd/cron.go` — CLI command for starting the scheduler or running jobs on demand

### Running the Cron Scheduler
To start all scheduled cron jobs (as a foreground process):
```bash
go run cli.go cron:start
```
Or as a background daemon:
```bash
nohup go run cli.go cron:start > cron.log 2>&1 &
```

### Running a Single Job by Name
You can run a single job immediately by name:
```bash
go run cli.go cron:start --job ProductJsonJob
# or with parameters:
go run cli.go cron:start --job ProductJsonJob param1 param2
```
- The job name is case-insensitive.
- All extra arguments are passed to the job function as parameters.

### Adding New Jobs
1. Create a new file in `cron/jobs/` and implement your job as `func(params ...string)`.
2. Register the job in `cron/sheduler.go` for scheduled execution (wrap in a closure: `func() { jobs.YourJob() }`).
3. Add a case for your job in `cmd/cron.go` to allow CLI execution by name.

### Example Job Implementation
```go
// cron/jobs/product_json.go
func ProductJsonJob(params ...string) {
    fmt.Println("Running ProductJsonJob", params)
    // Your logic here
}
```

### Example CLI Command
```bash
go run cli.go cron:start --job ProductJsonJob 123 store2
```

### Production Daemon (systemd)
For production, create a systemd service to run the scheduler as a daemon. See the project documentation for a sample unit file.

### Example: Adding and Using a CronJobs

1. **Register the job in `config/cron.go`:**
   ```go
   var CronJobs = map[string]CronJob{
       "productjsonjob": {Schedule: "0 * * * *", Job: jobs.ProductJsonJob},
       "testjob":        {Schedule: "@every 10s", Job: jobs.TestJob},
   }
   ```
   You can add more Jobs here 

2. **Run the test job manually:**
   ```bash
   go run cli.go cron:start --job TestJob
   # or with parameters:
   go run cli.go cron:start --job TestJob foo bar
   ```

3. **Scheduled run:**
   - When the scheduler is started, `TestJob` will run every 10 seconds automatically.

---
