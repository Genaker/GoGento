# Testing Strategy

GoGento Catalog uses a two-tier testing approach: **SQLite unit tests** for fast, isolated testing without dependencies, and **MySQL integration tests** for real-world validation against Magento databases.

## SQLite Tests (Unit/Service Tests)

### Benefits

| Benefit | Description |
|---------|-------------|
| **No Dependencies** | No Magento installation or MySQL server required |
| **Fast Execution** | In-memory SQLite runs tests in milliseconds |
| **Portable** | Tests run anywhere Go is installed |
| **CI/CD Ready** | Perfect for automated pipelines |
| **Schema Simulation** | Can simulate both CE and EE schemas dynamically |
| **Isolated** | Each test creates its own database, no shared state |

### How It Works

SQLite tests create in-memory databases with Magento-like schemas:

```go
func createCESchema(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatal(err)
    }
    
    // Create CE-style table (entity_id only)
    db.Exec(`CREATE TABLE catalog_product_entity (
        entity_id INTEGER PRIMARY KEY AUTOINCREMENT,
        attribute_set_id INTEGER DEFAULT 0,
        type_id TEXT DEFAULT 'simple',
        sku TEXT NOT NULL,
        has_options INTEGER DEFAULT 0,
        required_options INTEGER DEFAULT 0,
        created_at TEXT,
        updated_at TEXT
    )`)
    
    return db
}

func createEESchema(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatal(err)
    }
    
    // Create EE-style table (entity_id + row_id)
    db.Exec(`CREATE TABLE catalog_product_entity (
        entity_id INTEGER NOT NULL,
        row_id INTEGER PRIMARY KEY AUTOINCREMENT,
        attribute_set_id INTEGER DEFAULT 0,
        type_id TEXT DEFAULT 'simple',
        sku TEXT NOT NULL,
        has_options INTEGER DEFAULT 0,
        required_options INTEGER DEFAULT 0,
        created_at TEXT,
        updated_at TEXT,
        created_in INTEGER DEFAULT 1,
        updated_in INTEGER DEFAULT 2147483647
    )`)
    
    return db
}
```

### Limitations

| Limitation | Impact |
|------------|--------|
| **SQL Dialect Differences** | Some MySQL-specific functions (e.g., `LEAST()`) need SQLite alternatives |
| **No Foreign Keys (default)** | SQLite foreign key enforcement is optional |
| **Type Coercion** | SQLite is more permissive with types |
| **No Stored Procedures** | Cannot test MySQL-specific procedures |
| **Performance Not Representative** | SQLite in-memory is faster than real MySQL |

### Running SQLite Tests

```bash
# Run all service tests (SQLite-based)
go test -v ./tests/service/...

# Run specific test
go test -v ./tests/service/... -run TestProduct_CE_Schema
```

## MySQL Integration Tests (Real Magento DB)

### When to Use

- Validate against real Magento data structures
- Performance benchmarking
- Test complex queries with actual indexes
- Verify foreign key constraints
- Test against production-like data volumes

### Setup

Integration tests connect to a real Magento MySQL database:

```bash
# Required environment variables
export MAGENTO_DB_HOST=lccoins-db-1
export MAGENTO_DB_PORT=3306
export MAGENTO_DB_NAME=magento
export MAGENTO_DB_USER=magento
export MAGENTO_DB_PASSWORD=magento
```

### Running Integration Tests

```bash
# Run all integration tests
go test -v ./tests/integration/...

# Run performance benchmarks
go test -v ./tests/integration/... -run TestPerformance

# Run API endpoint tests
go test -v ./tests/integration/... -run 'TestAPI_|TestGraphQL_'
```

### Test Categories

| Test Type | Location | Database | Purpose |
|-----------|----------|----------|---------|
| Schema Detection | `tests/service/` | SQLite | Verify CE/EE detection logic |
| Model Compatibility | `tests/service/` | SQLite | Test GORM models with both schemas |
| HMAC Verification | `tests/service/` | None | Crypto validation |
| Repository Logic | `tests/service/` | SQLite | Query building, data mapping |
| API Endpoints | `tests/integration/` | MySQL | Full stack request/response |
| Performance | `tests/integration/` | MySQL | Benchmark import throughput |

## Test File Structure

```
tests/
├── service/                    # SQLite-based unit tests
│   ├── import_test.go          # Import service tests
│   └── realtime_test.go        # Realtime API component tests
│                               # - HMAC signature tests
│                               # - Repository tests
│                               # - CE/EE schema compatibility
│
└── integration/                # MySQL-based integration tests
    ├── magento_db_test.go      # Database connection, performance
    └── api_test.go             # API endpoint verification
                                # - REST endpoints
                                # - GraphQL queries
                                # - Error handling
```

## Best Practices

### For SQLite Tests

1. **Create fresh DB per test** - Use `:memory:` for isolation
2. **Simulate both schemas** - Test CE and EE variants
3. **Mock external dependencies** - Don't rely on config files
4. **Test edge cases** - NULL values, empty strings, zero values

### For Integration Tests

1. **Skip gracefully** - Check DB connection before running
2. **Don't modify production data** - Use test-specific records or rollback
3. **Log performance metrics** - Help track regressions
4. **Handle missing data** - Skip tests if required records don't exist

```go
func magentoTestDB(t *testing.T) *gorm.DB {
    t.Helper()
    db, err := connectToMagento()
    if err != nil {
        t.Skip("cannot connect to Magento DB:", err)
    }
    return db
}
```

## CI/CD Pipeline Example

```yaml
# .github/workflows/test.yml
name: Tests

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go test -v ./tests/service/...

  integration-tests:
    runs-on: ubuntu-latest
    needs: unit-tests
    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_DATABASE: magento
          MYSQL_USER: magento
          MYSQL_PASSWORD: magento
          MYSQL_ROOT_PASSWORD: root
        ports:
          - 3306:3306
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: go test -v ./tests/integration/...
        env:
          MAGENTO_DB_HOST: localhost
          MAGENTO_DB_PORT: 3306
          MAGENTO_DB_NAME: magento
          MAGENTO_DB_USER: magento
          MAGENTO_DB_PASSWORD: magento
```

## Summary

| Aspect | SQLite Tests | MySQL Integration Tests |
|--------|--------------|------------------------|
| Speed | ~10ms per test | ~100ms-10s per test |
| Dependencies | None | Running MySQL + Magento schema |
| Use Case | Logic validation | Real-world verification |
| CI Suitability | Excellent | Requires DB service |
| Data | Synthetic | Real Magento data |
| Coverage | Models, logic, crypto | Full API stack |
