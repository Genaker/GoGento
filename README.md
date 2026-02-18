# GoGento Catalog — Magento GraphQL & REST API in Go

Magento 2 API and Frontend in Go — HTML server side rendering, GraphQL and REST, one binary, no slow PHP.

## The world’s fastest framework for building e-Commerce MAGENTO websites!

![image](https://github.com/user-attachments/assets/eaacc9e0-e497-4d3c-a4d9-faeadc7fd6e5)

**GoGento Catalog** connects your Magento 2 database to modern frontends via Echo and GORM. Single binary, ~300+ req/s on a single CPU, sub-30ms for 100 products with 100 EAV attributes. EAV flattening, stock, prices, categories. Concurrent-safe cache, extensible registry, standalone GraphQL mode. Works with Venia, React, Next.js, Vue.

**Why GoGento?** If you run Magento 2 and want a fast, headless API without PHP — deploy one binary, point it at your MySQL, and serve catalog data to PWA, mobile apps, or third-party integrations. No Magento runtime, no Composer, no PHP-FPM. Lower memory, faster cold starts, simpler ops. Use your existing Magento DB schema; products, categories, EAV attributes, stock, and prices flow through unchanged.

## Architecture

![Architecture Overview](doc/images/architecture-overview.png)

**Architecture.** Repository layer for DB access, service layer for logic, API handlers for HTTP. GraphQL schema matches Magento/Venia conventions so frontends built for Magento GraphQL work with minimal changes. REST flat endpoints return products with attributes as keys. Optional global in-memory cache for hot paths; disable with `PRODUCT_FLAT_CACHE=off` for direct DB. Cron jobs for background tasks. Extensible registry for custom resolvers and fields.

**Deployment.** Run full API (REST + GraphQL) or standalone GraphQL server. Single executable, config via env vars. systemd, Docker, or bare metal. No separate app server. Scale horizontally by running more instances behind a load balancer.

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
MAGENTO_CRYPT_KEY=your_crypt_key  # From app/etc/env.php for Realtime API auth
```

## Features

| Feature | Description | Doc |
|---------|-------------|-----|
| **GraphQL API** | Products, categories, search; Magento/Venia-compatible schema. Store header, pagination, filters. | [graphql.md](doc/graphql.md) |
| **REST API** | Flat products (EAV as keys), orders CRUD. Basic auth. Optional store ID. | [rest-api.md](doc/rest-api.md) |
| **Realtime API** | Sub-30ms price/inventory lookups. HMAC-SHA256 signed. Raw SQL, parallel queries. | [realtime-api.md](doc/realtime-api.md) |
| **Standalone GraphQL** | Run GraphQL only: `go run ./cmd/graphql`. No REST, smaller footprint. | [installation.md](doc/installation.md) |
| **EAV flattening** | Attributes as keys, stock_item, index_prices. FetchWithAllAttributesFlat. | [eav-products.md](doc/eav-products.md) |
| **Global cache** | In-memory, concurrent-safe. ~300 req/s. Set `PRODUCT_FLAT_CACHE=off` to bypass. | [cache.md](doc/cache.md) |
| **Registry & cache** | Global cache, registry, singleton repos. Per-request isolation. | [registry.md](doc/registry.md) |
| **Product Import** | Magmi alternative. Bulk CSV import with parallel EAV writes. ~127k products/min. | [rest-api.md](doc/rest-api.md) |
| **Cron jobs** | Scheduled and on-demand. `go run cli.go cron:start --job Name`. | [cron.md](doc/cron.md) |
| **Extending** | Add entities, custom resolvers, GraphQL extensions. Tailwind. | [extending.md](doc/extending.md) |
| **Testing** | SQLite unit tests (no Magento needed) + MySQL integration tests. | [testing.md](doc/testing.md) |

## Quick Start

```bash
cd gogento-catalog
go mod tidy
go run magento.go
```

- **GraphQL:** `POST http://localhost:8080/graphql`
- **Playground:** `GET http://localhost:8080/playground`

```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "Store: 1" \
  -d '{"query":"query { products { total_count } }"}'
```

## Documentation

[Technical index](doc/technical.md) · [Installation](doc/installation.md) · [Production](doc/production.md)

## Product Import — Magmi Alternative in Go

GoGento includes a high-performance bulk product importer — a Go-native replacement for [Magmi](https://github.com/dweeves/magmi-mirror). It reads CSV files and writes directly to Magento's EAV tables using parallel bulk upserts, bypassing PHP entirely.

### Why not Magmi?

Magmi is a PHP-based direct-SQL importer that was fast for its era but has limitations:

- **PHP memory ceiling** — large catalogs (50k+ products) hit memory limits or require careful tuning
- **Single-threaded** — one EAV type at a time, no parallel DB writes
- **Unmaintained** — last meaningful update years ago, no Magento 2.4+ testing
- **No API mode** — CLI only, no REST endpoint for programmatic imports

GoGento's importer solves all of these:

- **Parallel DB writes** — EAV (varchar, int, decimal, text, datetime), stock, gallery, and price tables are flushed concurrently via goroutines
- **Constant memory** — streaming CSV parse + fixed batch buffers, memory doesn't grow with file size
- **Raw SQL mode** — optional `--raw-sql` flag bypasses ORM overhead for maximum throughput
- **Dual interface** — CLI (`products:import`) for files, REST API (`POST /api/stock/import`) for programmatic use
- **Single binary** — no PHP, no Composer, no JVM; deploy and run

### Benchmark: 100,000 Products (MySQL CE)

Each product has 50 EAV attributes (20 varchar, 10 int, 10 decimal, 5 text, 5 datetime).

| Metric | Value |
|--------|-------|
| Products imported | 100,000 |
| EAV rows upserted | 5,000,000 |
| Total time | 55.8 seconds |
| **Throughput** | **1,792 products/sec** |
| **Throughput** | **107,493 products/min** |
| EAV row rate | 89,578 rows/sec |

> Tested on real Magento MySQL database (Community Edition schema). Performance on production systems may vary based on disk I/O, indexes, and connection pool settings.

### Benchmark: 100,000 Products (SQLite)

| Metric | Value |
|--------|-------|
| Products imported | 100,000 |
| EAV rows upserted | 5,000,000 |
| Total time | ~47 seconds |
| **Throughput** | **~2,100 products/sec** |
| **Throughput** | **~127,000 products/min** |
| EAV row rate | ~106,000 rows/sec |

> SQLite in-memory benchmark demonstrates Go-side processing overhead is minimal.

### Benchmark: Stock Import (MySQL)

Tested on real Magento MySQL database with 1,000 existing products.

| Metric | Value |
|--------|-------|
| Products updated | 1,000 |
| Total time | ~156 ms |
| **Throughput** | **~6,400 products/sec** |
| **Throughput** | **~385,000 products/min** |

Stock import uses GORM's batch upsert (`CreateInBatches` with `ON CONFLICT DO UPDATE`) for efficient bulk updates with configurable batch size (default 500).

### Schema Compatibility

GoGento supports both Magento schema variants with **automatic runtime detection**:

| Edition | EAV Link Column | Detection |
|---------|-----------------|-----------|
| **Community Edition (CE)** | `entity_id` | Default for SQLite and standard MySQL |
| **Enterprise Edition (EE)** | `row_id` | Auto-detected from EAV table structure |

**How Detection Works:**

At startup, `DetectSchema(db)` queries `DESCRIBE catalog_product_entity_varchar`:
- If `row_id` column exists → EE schema, sets `IsEnterprise = true`
- If `entity_id` column exists → CE schema, sets `IsEnterprise = false`

The `Product` struct contains both fields with `json:"...,omitempty"` tags, so JSON responses only include the relevant ID (zero values are omitted).

```go
// EAVLinkID() returns the appropriate ID for EAV foreign keys
product.EAVLinkID()  // Returns RowID for EE, EntityID for CE
```

**No configuration required** — just connect to your database and GoGento adapts automatically.

### Usage

```bash
# Import from CSV (raw SQL, batch 1000)
gogento products:import -f products.csv --raw-sql --batch-size 1000

# Stock import via API
curl -X POST http://localhost:8080/api/stock/import \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{"items": [{"sku": "SKU-001", "qty": 100, "is_in_stock": 1}]}'
```

See [rest-api.md](doc/rest-api.md) for full CSV format, API request/response, and all available flags.

## Realtime Pricing & Inventory API

High-performance, stateless endpoints for real-time price and stock lookups with sub-30ms response times.

### Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /api/realtime/price-inventory?sku=X` | Price + stock (parallel fetch) |
| `GET /api/realtime/price?sku=X` | Lowest price only |
| `GET /api/realtime/stock?sku=X&source=default` | Stock quantity only |
| `GET /api/realtime/tier-prices?sku=X` | All tier prices |

### Features

- **HMAC-SHA256 Authentication** — Stateless, signed requests using Magento's crypt key
- **Parallel Queries** — Price and inventory fetched concurrently via errgroup
- **Raw SQL** — Direct `database/sql` queries, no ORM overhead
- **Lowest Price Logic** — Uses SQL `LEAST()` across base price, special price, and tier price
- **Auto Schema Detection** — Supports both CE (`entity_id`) and EE (`row_id`)

### PHP Integration

```php
// Generate HMAC signature
$signature = hash_hmac('sha256', $customerId, $cryptKey);

// Request with signed headers
$response = file_get_contents(
    'http://gogento:8080/api/realtime/price-inventory?sku=SKU-001',
    false,
    stream_context_create([
        'http' => [
            'header' => "X-Customer-ID: {$customerId}\r\nX-Customer-Sig: {$signature}"
        ]
    ])
);
```

### JavaScript Integration (Vanilla JS)

```javascript
// Signature should be generated server-side and passed to frontend
async function getRealtimePrice(sku, customerId, signature) {
    const response = await fetch(
        `https://gogento.example.com/api/realtime/price-inventory?sku=${encodeURIComponent(sku)}`,
        {
            method: 'GET',
            headers: {
                'X-Customer-ID': customerId,
                'X-Customer-Sig': signature
            }
        }
    );
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json();
}

// Usage
getRealtimePrice('SKU-001', '42', 'a1b2c3...')
    .then(data => {
        document.getElementById('price').textContent = `$${data.price}`;
        document.getElementById('stock').textContent = data.stock > 0 ? 'In Stock' : 'Out of Stock';
    });
```

### React Integration

```jsx
import { useState, useEffect } from 'react';

function useRealtimePrice(sku, customerId, signature) {
    const [data, setData] = useState({ price: null, stock: null, loading: true });

    useEffect(() => {
        if (!sku) return;
        
        fetch(`/api/realtime/price-inventory?sku=${encodeURIComponent(sku)}`, {
            headers: {
                'X-Customer-ID': customerId,
                'X-Customer-Sig': signature
            }
        })
        .then(res => res.json())
        .then(json => setData({ price: json.price, stock: json.stock, loading: false }))
        .catch(() => setData(prev => ({ ...prev, loading: false })));
    }, [sku, customerId, signature]);

    return data;
}

// Usage in component
function ProductPrice({ sku, customerId, signature }) {
    const { price, stock, loading } = useRealtimePrice(sku, customerId, signature);

    if (loading) return <span>Loading...</span>;

    return (
        <div>
            <span className="price">${price?.toFixed(2)}</span>
            <span className={stock > 0 ? 'in-stock' : 'out-of-stock'}>
                {stock > 0 ? `${stock} in stock` : 'Out of stock'}
            </span>
        </div>
    );
}
```

### Hyvä Integration (Alpine.js)

```html
<!-- Hyvä theme component using Alpine.js -->
<div x-data="realtimePrice('<?= $escaper->escapeJs($block->getSku()) ?>')" 
     x-init="fetchPrice()">
    
    <span x-show="loading">Loading...</span>
    
    <template x-if="!loading">
        <div>
            <span class="price" x-text="'$' + price?.toFixed(2)"></span>
            <span :class="stock > 0 ? 'text-green-600' : 'text-red-600'"
                  x-text="stock > 0 ? stock + ' in stock' : 'Out of stock'">
            </span>
        </div>
    </template>
</div>

<script>
function realtimePrice(sku) {
    return {
        sku: sku,
        price: null,
        stock: null,
        loading: true,
        customerId: '<?= $escaper->escapeJs($block->getCustomerId()) ?>',
        signature: '<?= $escaper->escapeJs($block->getCustomerSignature()) ?>',
        
        async fetchPrice() {
            try {
                const response = await fetch(
                    `<?= $escaper->escapeUrl($block->getGoGentoUrl()) ?>/api/realtime/price-inventory?sku=${encodeURIComponent(this.sku)}`,
                    {
                        headers: {
                            'X-Customer-ID': this.customerId,
                            'X-Customer-Sig': this.signature
                        }
                    }
                );
                const data = await response.json();
                this.price = data.price;
                this.stock = data.stock;
            } catch (e) {
                console.error('Realtime price fetch failed:', e);
            } finally {
                this.loading = false;
            }
        }
    };
}
</script>
```

```php
<?php
// Block class for Hyvä template
namespace Vendor\Module\Block;

class RealtimePrice extends \Magento\Framework\View\Element\Template
{
    public function getCustomerSignature(): string
    {
        $customerId = $this->getCustomerId();
        $cryptKey = $this->_scopeConfig->getValue('system/crypt/key');
        return hash_hmac('sha256', (string)$customerId, $cryptKey);
    }
    
    public function getGoGentoUrl(): string
    {
        return $this->_scopeConfig->getValue('gogento/general/api_url') ?? 'http://gogento:8080';
    }
}
```

See [realtime-api.md](doc/realtime-api.md) for full PHP client class, batch requests, and Magento Observer integration.

## Performance

- **With cache:** ~300 req/s, ~1 ms single product (ApacheBench)
- **100 products, 100 attrs:** REST ~25 ms, GraphQL ~30 ms (`make test-perf`)
- **Realtime API:** < 30ms price + inventory (parallel raw SQL)
- **Product import:** ~127,000 products/min (100k products, 50 attrs each, raw SQL)
- No N+1: batch Preload with IN clauses

## Environment

```bash
MYSQL_USER=magento MYSQL_PASS=magento MYSQL_HOST=localhost MYSQL_DB=magento
API_USER=admin API_PASS=secret PORT=8080
```

## Tests

```bash
make test      # All tests
make test-perf # GraphQL vs REST benchmark
```
