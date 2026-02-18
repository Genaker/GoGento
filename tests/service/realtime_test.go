package servicetest

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	productEntity "magento.GO/model/entity/product"
	inventoryRepo "magento.GO/model/repository/inventory"
	priceRepo "magento.GO/model/repository/price"
)

func TestHMAC_SignatureGeneration(t *testing.T) {
	cryptKey := "3254cdb1ae5233a336cdec765aeb3bb6"
	customerID := "123"

	mac := hmac.New(sha256.New, []byte(cryptKey))
	mac.Write([]byte(customerID))
	sig := hex.EncodeToString(mac.Sum(nil))

	if sig == "" {
		t.Error("signature should not be empty")
	}
	if len(sig) != 64 {
		t.Errorf("signature length = %d, want 64 hex chars", len(sig))
	}
}

func TestHMAC_SignatureVerification(t *testing.T) {
	cryptKey := "3254cdb1ae5233a336cdec765aeb3bb6"
	customerID := "123"

	// Generate signature
	mac := hmac.New(sha256.New, []byte(cryptKey))
	mac.Write([]byte(customerID))
	expected := mac.Sum(nil)
	sigHex := hex.EncodeToString(expected)

	// Verify with same key
	mac2 := hmac.New(sha256.New, []byte(cryptKey))
	mac2.Write([]byte(customerID))
	computed := mac2.Sum(nil)

	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		t.Fatalf("hex decode: %v", err)
	}

	if !hmac.Equal(computed, sig) {
		t.Error("signature verification failed")
	}
}

func TestHMAC_TamperedID_Fails(t *testing.T) {
	cryptKey := "3254cdb1ae5233a336cdec765aeb3bb6"
	customerID := "123"
	tamperedID := "124"

	// Generate signature for original ID
	mac := hmac.New(sha256.New, []byte(cryptKey))
	mac.Write([]byte(customerID))
	sigHex := hex.EncodeToString(mac.Sum(nil))

	// Verify with tampered ID
	mac2 := hmac.New(sha256.New, []byte(cryptKey))
	mac2.Write([]byte(tamperedID))
	computed := mac2.Sum(nil)

	sig, _ := hex.DecodeString(sigHex)

	if hmac.Equal(computed, sig) {
		t.Error("tampered ID should fail verification")
	}
}

func TestInventoryRepository_SQLite(t *testing.T) {
	db := importDB(t)

	// Create inventory_source_item table for SQLite
	db.Exec(`CREATE TABLE IF NOT EXISTS inventory_source_item (
		source_item_id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_code VARCHAR(255) NOT NULL,
		sku VARCHAR(64) NOT NULL,
		quantity DECIMAL(12,4) NOT NULL DEFAULT 0,
		status INTEGER NOT NULL DEFAULT 0
	)`)

	// Insert test data
	db.Exec(`INSERT INTO inventory_source_item (source_code, sku, quantity, status) VALUES (?, ?, ?, ?)`,
		"default", "TEST-SKU-001", 150.5, 1)

	repo, err := inventoryRepo.NewInventoryRepository(db)
	if err != nil {
		t.Fatalf("NewInventoryRepository: %v", err)
	}

	qty, found := repo.GetQuantityBySKU("TEST-SKU-001", "default")
	if !found {
		t.Error("expected to find stock for TEST-SKU-001")
	}
	if qty != 150.5 {
		t.Errorf("quantity = %f, want 150.5", qty)
	}

	// Test not found
	_, found = repo.GetQuantityBySKU("NONEXISTENT", "default")
	if found {
		t.Error("should not find nonexistent SKU")
	}
}

func TestInventoryRepository_GetAllBySKU(t *testing.T) {
	db := importDB(t)

	db.Exec(`CREATE TABLE IF NOT EXISTS inventory_source_item (
		source_item_id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_code VARCHAR(255) NOT NULL,
		sku VARCHAR(64) NOT NULL,
		quantity DECIMAL(12,4) NOT NULL DEFAULT 0,
		status INTEGER NOT NULL DEFAULT 0
	)`)

	// Multiple sources
	db.Exec(`INSERT INTO inventory_source_item (source_code, sku, quantity, status) VALUES (?, ?, ?, ?)`,
		"warehouse_a", "MULTI-SKU", 100, 1)
	db.Exec(`INSERT INTO inventory_source_item (source_code, sku, quantity, status) VALUES (?, ?, ?, ?)`,
		"warehouse_b", "MULTI-SKU", 50, 1)

	repo, _ := inventoryRepo.NewInventoryRepository(db)

	items, err := repo.GetAllBySKU("MULTI-SKU")
	if err != nil {
		t.Fatalf("GetAllBySKU: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestInventoryRepository_GetTotalQuantity(t *testing.T) {
	db := importDB(t)

	db.Exec(`CREATE TABLE IF NOT EXISTS inventory_source_item (
		source_item_id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_code VARCHAR(255) NOT NULL,
		sku VARCHAR(64) NOT NULL,
		quantity DECIMAL(12,4) NOT NULL DEFAULT 0,
		status INTEGER NOT NULL DEFAULT 0
	)`)

	db.Exec(`INSERT INTO inventory_source_item (source_code, sku, quantity, status) VALUES (?, ?, ?, ?)`,
		"src1", "TOTAL-SKU", 100, 1)
	db.Exec(`INSERT INTO inventory_source_item (source_code, sku, quantity, status) VALUES (?, ?, ?, ?)`,
		"src2", "TOTAL-SKU", 75, 1)

	repo, _ := inventoryRepo.NewInventoryRepository(db)

	total, err := repo.GetTotalQuantityBySKU("TOTAL-SKU")
	if err != nil {
		t.Fatalf("GetTotalQuantityBySKU: %v", err)
	}
	if total != 175 {
		t.Errorf("total = %f, want 175", total)
	}
}

func TestInventoryRepository_BatchGetQuantities(t *testing.T) {
	db := importDB(t)

	db.Exec(`CREATE TABLE IF NOT EXISTS inventory_source_item (
		source_item_id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_code VARCHAR(255) NOT NULL,
		sku VARCHAR(64) NOT NULL,
		quantity DECIMAL(12,4) NOT NULL DEFAULT 0,
		status INTEGER NOT NULL DEFAULT 0
	)`)

	db.Exec(`INSERT INTO inventory_source_item (source_code, sku, quantity, status) VALUES (?, ?, ?, ?)`,
		"default", "BATCH-1", 10, 1)
	db.Exec(`INSERT INTO inventory_source_item (source_code, sku, quantity, status) VALUES (?, ?, ?, ?)`,
		"default", "BATCH-2", 20, 1)
	db.Exec(`INSERT INTO inventory_source_item (source_code, sku, quantity, status) VALUES (?, ?, ?, ?)`,
		"default", "BATCH-3", 30, 1)

	repo, _ := inventoryRepo.NewInventoryRepository(db)

	result, err := repo.BatchGetQuantities([]string{"BATCH-1", "BATCH-2", "BATCH-3", "BATCH-MISSING"}, "default")
	if err != nil {
		t.Fatalf("BatchGetQuantities: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 results, got %d", len(result))
	}
	if result["BATCH-1"] != 10 {
		t.Errorf("BATCH-1 = %f, want 10", result["BATCH-1"])
	}
	if result["BATCH-2"] != 20 {
		t.Errorf("BATCH-2 = %f, want 20", result["BATCH-2"])
	}
}

func TestPriceRepository_SQLite_BasePriceFallback(t *testing.T) {
	db := importDB(t)

	// SQLite doesn't have information_schema, so schema detection returns CE mode
	repo, err := priceRepo.NewPriceRepository(db)
	if err != nil {
		t.Fatalf("NewPriceRepository: %v", err)
	}

	// Schema detection should not crash on SQLite
	isEE := repo.IsEnterprise()
	if isEE {
		t.Error("SQLite should be detected as CE schema")
	}
}

// createCESchema creates CE-style catalog_product_entity table (entity_id as PK, no row_id)
func createCESchema(t *testing.T) *gorm.DB {
	t.Helper()
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("ce_schema_%s_%d.db", t.Name(), time.Now().UnixNano()))
	t.Cleanup(func() { os.Remove(tmpFile) })

	db, err := gorm.Open(sqlite.Open(tmpFile), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	// CE schema: entity_id is primary key, no row_id column
	db.Exec(`CREATE TABLE catalog_product_entity (
		entity_id INTEGER PRIMARY KEY AUTOINCREMENT,
		attribute_set_id INTEGER NOT NULL DEFAULT 0,
		type_id VARCHAR(32) NOT NULL DEFAULT 'simple',
		sku VARCHAR(64) NOT NULL,
		has_options INTEGER NOT NULL DEFAULT 0,
		required_options INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)

	// CE EAV tables use entity_id as foreign key
	db.Exec(`CREATE TABLE catalog_product_entity_varchar (
		value_id INTEGER PRIMARY KEY AUTOINCREMENT,
		attribute_id INTEGER NOT NULL,
		store_id INTEGER NOT NULL DEFAULT 0,
		entity_id INTEGER NOT NULL,
		value VARCHAR(255),
		UNIQUE(entity_id, attribute_id, store_id)
	)`)

	db.Exec(`CREATE TABLE catalog_product_entity_int (
		value_id INTEGER PRIMARY KEY AUTOINCREMENT,
		attribute_id INTEGER NOT NULL,
		store_id INTEGER NOT NULL DEFAULT 0,
		entity_id INTEGER NOT NULL,
		value INTEGER,
		UNIQUE(entity_id, attribute_id, store_id)
	)`)

	return db
}

// createEESchema creates EE-style catalog_product_entity table (row_id as PK, entity_id for sequence)
func createEESchema(t *testing.T) *gorm.DB {
	t.Helper()
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("ee_schema_%s_%d.db", t.Name(), time.Now().UnixNano()))
	t.Cleanup(func() { os.Remove(tmpFile) })

	db, err := gorm.Open(sqlite.Open(tmpFile), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	// EE schema: row_id is primary key, entity_id references sequence_product
	db.Exec(`CREATE TABLE sequence_product (
		sequence_value INTEGER PRIMARY KEY AUTOINCREMENT
	)`)

	db.Exec(`CREATE TABLE catalog_product_entity (
		row_id INTEGER PRIMARY KEY AUTOINCREMENT,
		entity_id INTEGER NOT NULL,
		attribute_set_id INTEGER NOT NULL DEFAULT 0,
		type_id VARCHAR(32) NOT NULL DEFAULT 'simple',
		sku VARCHAR(64) NOT NULL,
		has_options INTEGER NOT NULL DEFAULT 0,
		required_options INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		created_in INTEGER DEFAULT 1,
		updated_in INTEGER DEFAULT 2147483647
	)`)

	// EE EAV tables use row_id as foreign key
	db.Exec(`CREATE TABLE catalog_product_entity_varchar (
		value_id INTEGER PRIMARY KEY AUTOINCREMENT,
		attribute_id INTEGER NOT NULL,
		store_id INTEGER NOT NULL DEFAULT 0,
		row_id INTEGER NOT NULL,
		value VARCHAR(255),
		UNIQUE(row_id, attribute_id, store_id)
	)`)

	db.Exec(`CREATE TABLE catalog_product_entity_int (
		value_id INTEGER PRIMARY KEY AUTOINCREMENT,
		attribute_id INTEGER NOT NULL,
		store_id INTEGER NOT NULL DEFAULT 0,
		row_id INTEGER NOT NULL,
		value INTEGER,
		UNIQUE(row_id, attribute_id, store_id)
	)`)

	return db
}

// TestProduct_CE_Schema verifies Product model works with CE database (no row_id column)
func TestProduct_CE_Schema(t *testing.T) {
	db := createCESchema(t)
	productEntity.IsEnterprise = false
	defer func() { productEntity.IsEnterprise = false }()

	// Create a product
	p := productEntity.Product{
		SKU:            "TEST-CE-001",
		TypeID:         "simple",
		AttributeSetID: 4,
	}

	err := db.Create(&p).Error
	if err != nil {
		t.Fatalf("Create product on CE schema failed: %v", err)
	}

	if p.EntityID == 0 {
		t.Error("EntityID should be auto-generated")
	}

	// Verify we can read it back
	var loaded productEntity.Product
	err = db.First(&loaded, p.EntityID).Error
	if err != nil {
		t.Fatalf("Load product failed: %v", err)
	}

	if loaded.SKU != "TEST-CE-001" {
		t.Errorf("SKU = %q, want TEST-CE-001", loaded.SKU)
	}

	// EAVLinkID should return EntityID for CE
	if loaded.EAVLinkID() != loaded.EntityID {
		t.Errorf("EAVLinkID() = %d, want EntityID %d", loaded.EAVLinkID(), loaded.EntityID)
	}

	// Insert EAV varchar row using entity_id
	db.Exec(`INSERT INTO catalog_product_entity_varchar (attribute_id, store_id, entity_id, value) VALUES (?, ?, ?, ?)`,
		73, 0, loaded.EntityID, "Test Product Name")

	// Verify EAV row
	var eavCount int64
	db.Raw(`SELECT COUNT(*) FROM catalog_product_entity_varchar WHERE entity_id = ?`, loaded.EntityID).Scan(&eavCount)
	if eavCount != 1 {
		t.Errorf("EAV varchar count = %d, want 1", eavCount)
	}

	t.Logf("CE Schema: Created product EntityID=%d, SKU=%s", loaded.EntityID, loaded.SKU)
}

// TestProduct_EE_Schema verifies Product model works with EE database (row_id as PK)
func TestProduct_EE_Schema(t *testing.T) {
	db := createEESchema(t)
	productEntity.IsEnterprise = true
	defer func() { productEntity.IsEnterprise = false }()

	// For EE, we need to insert into sequence_product first to get entity_id
	db.Exec(`INSERT INTO sequence_product DEFAULT VALUES`)
	var entityID uint
	db.Raw(`SELECT MAX(sequence_value) FROM sequence_product`).Scan(&entityID)

	// Insert product with entity_id (row_id auto-generated)
	db.Exec(`INSERT INTO catalog_product_entity (entity_id, sku, type_id, attribute_set_id) VALUES (?, ?, ?, ?)`,
		entityID, "TEST-EE-001", "simple", 4)

	// Get the row_id
	var rowID uint
	db.Raw(`SELECT row_id FROM catalog_product_entity WHERE entity_id = ?`, entityID).Scan(&rowID)

	if rowID == 0 {
		t.Fatal("row_id should be auto-generated")
	}

	// Load using GORM - note: our struct has row_id as gorm:"-" so we need raw query
	var sku string
	db.Raw(`SELECT sku FROM catalog_product_entity WHERE row_id = ?`, rowID).Scan(&sku)
	if sku != "TEST-EE-001" {
		t.Errorf("SKU = %q, want TEST-EE-001", sku)
	}

	// Test EAVLinkID with manual product
	p := productEntity.Product{EntityID: entityID, RowID: rowID, SKU: sku}

	// EAVLinkID should return RowID for EE
	if p.EAVLinkID() != rowID {
		t.Errorf("EAVLinkID() = %d, want RowID %d", p.EAVLinkID(), rowID)
	}

	// Insert EAV varchar row using row_id
	db.Exec(`INSERT INTO catalog_product_entity_varchar (attribute_id, store_id, row_id, value) VALUES (?, ?, ?, ?)`,
		73, 0, rowID, "EE Test Product Name")

	// Verify EAV row uses row_id
	var eavCount int64
	db.Raw(`SELECT COUNT(*) FROM catalog_product_entity_varchar WHERE row_id = ?`, rowID).Scan(&eavCount)
	if eavCount != 1 {
		t.Errorf("EAV varchar count = %d, want 1", eavCount)
	}

	t.Logf("EE Schema: Created product EntityID=%d, RowID=%d, SKU=%s", entityID, rowID, sku)
}

// TestProduct_EE_Schema_Simulation is kept for backward compatibility
func TestProduct_EE_Schema_Simulation(t *testing.T) {
	db := importDB(t)

	// Simulate EE by adding row_id column to SQLite
	db.Exec(`ALTER TABLE catalog_product_entity ADD COLUMN row_id INTEGER`)

	// Create a product
	p := productEntity.Product{
		SKU:            "TEST-EE-SCHEMA-001",
		TypeID:         "simple",
		AttributeSetID: 4,
	}

	err := db.Create(&p).Error
	if err != nil {
		t.Fatalf("Create product failed: %v", err)
	}

	if p.EntityID == 0 {
		t.Error("EntityID should be auto-generated")
	}

	// Manually set RowID to simulate EE behavior
	p.RowID = p.EntityID + 1000

	// EAVLinkID should return RowID when IsEnterprise=true
	productEntity.IsEnterprise = true
	defer func() { productEntity.IsEnterprise = false }()

	if p.EAVLinkID() != p.RowID {
		t.Errorf("EAVLinkID() = %d, want RowID %d", p.EAVLinkID(), p.RowID)
	}

	// Cleanup
	db.Delete(&p)
}

// TestProduct_JSON_Omitempty_BothSchemas verifies JSON serialization works for both CE and EE
func TestProduct_JSON_Omitempty_BothSchemas(t *testing.T) {
	// CE: EntityID set, RowID zero
	productEntity.IsEnterprise = false
	pCE := productEntity.Product{EntityID: 42, SKU: "CE-SKU"}

	dataCE, _ := json.Marshal(pCE)
	jsonCE := string(dataCE)

	if !strings.Contains(jsonCE, `"entity_id":42`) {
		t.Errorf("CE JSON should contain entity_id, got: %s", jsonCE)
	}
	if strings.Contains(jsonCE, `"row_id"`) {
		t.Errorf("CE JSON should omit row_id when zero, got: %s", jsonCE)
	}

	// EE: RowID set, EntityID might also be set
	productEntity.IsEnterprise = true
	defer func() { productEntity.IsEnterprise = false }()

	pEE := productEntity.Product{EntityID: 100, RowID: 200, SKU: "EE-SKU"}

	dataEE, _ := json.Marshal(pEE)
	jsonEE := string(dataEE)

	if !strings.Contains(jsonEE, `"row_id":200`) {
		t.Errorf("EE JSON should contain row_id, got: %s", jsonEE)
	}
	if !strings.Contains(jsonEE, `"entity_id":100`) {
		t.Errorf("EE JSON should contain entity_id, got: %s", jsonEE)
	}

	// Verify EAVLinkID returns correct value for each
	productEntity.IsEnterprise = false
	if pCE.EAVLinkID() != 42 {
		t.Errorf("CE EAVLinkID() = %d, want 42", pCE.EAVLinkID())
	}

	productEntity.IsEnterprise = true
	if pEE.EAVLinkID() != 200 {
		t.Errorf("EE EAVLinkID() = %d, want 200", pEE.EAVLinkID())
	}
}

// TestProduct_GORM_IgnoresEEColumns verifies GORM ignores row_id/created_in/updated_in on CE databases
func TestProduct_GORM_IgnoresEEColumns(t *testing.T) {
	db := importDB(t)

	// Create product with EE fields set (should be ignored by GORM)
	p := productEntity.Product{
		SKU:            "TEST-GORM-IGNORE-001",
		TypeID:         "simple",
		AttributeSetID: 4,
		RowID:          9999,    // Should be ignored (gorm:"-")
		CreatedIn:      1000000, // Should be ignored (gorm:"-")
		UpdatedIn:      2000000, // Should be ignored (gorm:"-")
	}

	err := db.Create(&p).Error
	if err != nil {
		t.Fatalf("Create should succeed even with EE fields set: %v", err)
	}

	if p.EntityID == 0 {
		t.Error("EntityID should be auto-generated")
	}

	// The EE fields should retain their Go values (not persisted)
	if p.RowID != 9999 {
		t.Errorf("RowID in memory = %d, want 9999 (unchanged)", p.RowID)
	}

	// Cleanup
	db.Delete(&p)
}
