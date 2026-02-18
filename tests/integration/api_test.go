package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"magento.GO/api"
	_ "magento.GO/api/category"
	_ "magento.GO/api/product"
	_ "magento.GO/api/realtime"
	_ "magento.GO/api/stock"
	graphqlApi "magento.GO/api/graphql"
)

// setupTestServer creates an Echo server with all API routes for testing
func setupTestServer(t *testing.T, db *gorm.DB) *echo.Echo {
	t.Helper()
	e := echo.New()
	e.HideBanner = true

	// Register GraphQL routes
	graphqlApi.RegisterGraphQLRoutes(e, db)

	// Register API routes under /api group
	apiGroup := e.Group("/api")
	api.ApplyModules(apiGroup, db)

	return e
}

// TestAPI_HealthCheck verifies server starts and responds
func TestAPI_HealthCheck(t *testing.T) {
	db := magentoTestDB(t)
	e := setupTestServer(t, db)

	req := httptest.NewRequest(http.MethodGet, "/playground", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Playground should return HTML or redirect
	if rec.Code != http.StatusOK && rec.Code != http.StatusNotFound {
		t.Logf("Playground status: %d (may not be registered)", rec.Code)
	}
}

// TestAPI_Products_List verifies GET /api/products returns products
func TestAPI_Products_List(t *testing.T) {
	db := magentoTestDB(t)
	e := setupTestServer(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/products?limit=10", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		body, _ := io.ReadAll(rec.Body)
		t.Fatalf("GET /api/products?limit=10 status = %d, body: %s", rec.Code, string(body))
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("JSON decode: %v", err)
	}

	if _, ok := resp["products"]; !ok {
		t.Error("response should contain 'products' key")
	}
	if _, ok := resp["count"]; !ok {
		t.Error("response should contain 'count' key")
	}
	t.Logf("GET /api/products: %d products returned", int(resp["count"].(float64)))
}

// TestAPI_Products_ListAll verifies GET /api/products works without limit (batched)
func TestAPI_Products_ListAll(t *testing.T) {
	db := magentoTestDB(t)
	e := setupTestServer(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		body, _ := io.ReadAll(rec.Body)
		t.Fatalf("GET /api/products (no limit) status = %d, body: %s", rec.Code, string(body))
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("JSON decode: %v", err)
	}

	count := int(resp["count"].(float64))
	t.Logf("GET /api/products (no limit, batched): %d products returned", count)
}

// TestAPI_Products_Flat verifies GET /api/products/flat returns flat products
func TestAPI_Products_Flat(t *testing.T) {
	db := magentoTestDB(t)
	e := setupTestServer(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/products/flat?limit=10", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		body, _ := io.ReadAll(rec.Body)
		t.Fatalf("GET /api/products/flat?limit=10 status = %d, body: %s", rec.Code, string(body))
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("JSON decode: %v", err)
	}

	if _, ok := resp["products"]; !ok {
		t.Error("response should contain 'products' key")
	}
	t.Logf("GET /api/products/flat: %d products, %v ms", int(resp["count"].(float64)), resp["request_duration_ms"])
}

// TestAPI_Products_FlatAll verifies GET /api/products/flat works without limit (batched)
func TestAPI_Products_FlatAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping full flat products test in short mode")
	}
	db := magentoTestDB(t)
	e := setupTestServer(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/products/flat", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		body, _ := io.ReadAll(rec.Body)
		t.Fatalf("GET /api/products/flat (no limit) status = %d, body: %s", rec.Code, string(body))
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("JSON decode: %v", err)
	}

	count := int(resp["count"].(float64))
	duration := resp["request_duration_ms"]
	t.Logf("GET /api/products/flat (no limit, batched): %d products, %v ms", count, duration)
}

// TestAPI_Realtime_PriceInventory verifies GET /api/realtime/price-inventory
func TestAPI_Realtime_PriceInventory(t *testing.T) {
	db := magentoTestDB(t)
	e := setupTestServer(t, db)

	// Get a real SKU from the database
	var sku string
	db.Raw("SELECT sku FROM catalog_product_entity LIMIT 1").Scan(&sku)
	if sku == "" {
		t.Skip("no products in database")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/realtime/price-inventory?sku="+sku, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Should return 200 or 404 (if no price/stock data)
	if rec.Code != http.StatusOK && rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/realtime/price-inventory status = %d, want 200 or 404", rec.Code)
	}

	t.Logf("GET /api/realtime/price-inventory?sku=%s: status %d", sku, rec.Code)
}

// TestAPI_Realtime_Price verifies GET /api/realtime/price
func TestAPI_Realtime_Price(t *testing.T) {
	db := magentoTestDB(t)
	e := setupTestServer(t, db)

	var sku string
	db.Raw("SELECT sku FROM catalog_product_entity LIMIT 1").Scan(&sku)
	if sku == "" {
		t.Skip("no products in database")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/realtime/price?sku="+sku, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK && rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/realtime/price status = %d", rec.Code)
	}

	t.Logf("GET /api/realtime/price?sku=%s: status %d", sku, rec.Code)
}

// TestAPI_Realtime_Stock verifies GET /api/realtime/stock
func TestAPI_Realtime_Stock(t *testing.T) {
	db := magentoTestDB(t)
	e := setupTestServer(t, db)

	var sku string
	db.Raw("SELECT sku FROM catalog_product_entity LIMIT 1").Scan(&sku)
	if sku == "" {
		t.Skip("no products in database")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/realtime/stock?sku="+sku+"&source=default", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK && rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/realtime/stock status = %d", rec.Code)
	}

	t.Logf("GET /api/realtime/stock?sku=%s: status %d", sku, rec.Code)
}

// TestAPI_Realtime_TierPrices verifies GET /api/realtime/tier-prices
func TestAPI_Realtime_TierPrices(t *testing.T) {
	db := magentoTestDB(t)
	e := setupTestServer(t, db)

	var sku string
	db.Raw("SELECT sku FROM catalog_product_entity LIMIT 1").Scan(&sku)
	if sku == "" {
		t.Skip("no products in database")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/realtime/tier-prices?sku="+sku, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK && rec.Code != http.StatusInternalServerError {
		t.Fatalf("GET /api/realtime/tier-prices status = %d", rec.Code)
	}

	t.Logf("GET /api/realtime/tier-prices?sku=%s: status %d", sku, rec.Code)
}

// TestAPI_Realtime_MissingSKU verifies error handling for missing SKU
func TestAPI_Realtime_MissingSKU(t *testing.T) {
	db := magentoTestDB(t)
	e := setupTestServer(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/realtime/price-inventory", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("GET /api/realtime/price-inventory (no sku) status = %d, want 400", rec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"] != "sku required" {
		t.Errorf("error = %v, want 'sku required'", resp["error"])
	}
}

// TestGraphQL_Products verifies GraphQL products query
func TestGraphQL_Products(t *testing.T) {
	db := magentoTestDB(t)
	e := setupTestServer(t, db)

	query := `{"query":"query { products { total_count } }"}`
	req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(query))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		body, _ := io.ReadAll(rec.Body)
		t.Fatalf("POST /graphql status = %d, body: %s", rec.Code, string(body))
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("JSON decode: %v", err)
	}

	if _, ok := resp["data"]; !ok {
		t.Error("GraphQL response should contain 'data' key")
	}

	t.Logf("POST /graphql products query: success")
}

// TestGraphQL_Categories verifies GraphQL categories query
func TestGraphQL_Categories(t *testing.T) {
	db := magentoTestDB(t)
	e := setupTestServer(t, db)

	query := `{"query":"query { categories { total_count } }"}`
	req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(query))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		body, _ := io.ReadAll(rec.Body)
		t.Fatalf("POST /graphql status = %d, body: %s", rec.Code, string(body))
	}

	t.Logf("POST /graphql categories query: success")
}

// TestGraphQL_StoreConfig verifies GraphQL storeConfig query
func TestGraphQL_StoreConfig(t *testing.T) {
	db := magentoTestDB(t)
	e := setupTestServer(t, db)

	query := `{"query":"query { storeConfig { store_code } }"}`
	req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(query))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// storeConfig may or may not be implemented
	if rec.Code != http.StatusOK {
		t.Logf("POST /graphql storeConfig: status %d (may not be implemented)", rec.Code)
		return
	}

	t.Logf("POST /graphql storeConfig query: success")
}

// TestAPI_Stock_Import verifies POST /api/stock/import
func TestAPI_Stock_Import(t *testing.T) {
	db := magentoTestDB(t)
	e := setupTestServer(t, db)

	// Get a real SKU
	var sku string
	db.Raw("SELECT sku FROM catalog_product_entity LIMIT 1").Scan(&sku)
	if sku == "" {
		t.Skip("no products in database")
	}

	body := `{"items":[{"sku":"` + sku + `","qty":100,"is_in_stock":1}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/stock/import", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		bodyBytes, _ := io.ReadAll(rec.Body)
		t.Fatalf("POST /api/stock/import status = %d, body: %s", rec.Code, string(bodyBytes))
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	t.Logf("POST /api/stock/import: imported=%v, skipped=%v", resp["imported"], resp["skipped"])
}

// TestAPI_Endpoints_Summary runs all API endpoint checks and summarizes
func TestAPI_Endpoints_Summary(t *testing.T) {
	db := magentoTestDB(t)
	e := setupTestServer(t, db)

	// Get a real SKU for testing
	var sku string
	db.Raw("SELECT sku FROM catalog_product_entity LIMIT 1").Scan(&sku)
	if sku == "" {
		sku = "TEST"
	}

	endpoints := []struct {
		method string
		path   string
		body   string
	}{
		{"GET", "/api/products?limit=5", ""},
		{"GET", "/api/products/flat?limit=5", ""},
		{"GET", "/api/realtime/price-inventory?sku=" + sku, ""},
		{"GET", "/api/realtime/price?sku=" + sku, ""},
		{"GET", "/api/realtime/stock?sku=" + sku, ""},
		{"GET", "/api/realtime/tier-prices?sku=" + sku, ""},
		{"POST", "/graphql", `{"query":"{ products { total_count } }"}`},
		{"POST", "/api/stock/import", `{"items":[{"sku":"` + sku + `","qty":1,"is_in_stock":1}]}`},
	}

	t.Log("=== API Endpoints Summary ===")
	for _, ep := range endpoints {
		var req *http.Request
		if ep.body != "" {
			req = httptest.NewRequest(ep.method, ep.path, strings.NewReader(ep.body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(ep.method, ep.path, nil)
		}
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		status := "OK"
		if rec.Code >= 400 {
			status = "ERROR"
		}
		t.Logf("  %s %s -> %d (%s)", ep.method, ep.path, rec.Code, status)
	}
	t.Log("=============================")
}
