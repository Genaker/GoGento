package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"magento.GO/config"
	"magento.GO/model/entity/product"
	productRepo "magento.GO/model/repository/product"
	"time"
)

var migrateProductsCmd = &cobra.Command{
	Use:   "products:json:index",
	Short: "Migrate product data to JSON table with timing metrics",
	Run: func(cmd *cobra.Command, args []string) {
		startTotal := time.Now()

		db, err := config.NewDB()
		if err != nil {
			fmt.Printf("Database connection failed: %v\n", err)
			return
		}

		// Timing for data fetching
		startFetch := time.Now()
		repo := productRepo.NewProductRepository(db)
		flatProducts, err := repo.FetchWithAllAttributesFlat(0)
		if err != nil {
			fmt.Printf("Failed to fetch products: %v\n", err)
			return
		}
		fetchDuration := time.Since(startFetch)

		startFetchJson := time.Now()
		var jsonProducts []product.ProductJson
		// Create map with composite keys
		existingEntries := make(map[string]int)
		err = db.Select("id, entity_id, store_id").Find(&jsonProducts).Error
		if err != nil {
			fmt.Printf("Failed to fetch products: %v\n", err)
			return
		} // Populate the map
		for index, entry := range jsonProducts {
			key := fmt.Sprintf("%d_%d", entry.EntityID, entry.StoreID)
			existingEntries[key] = index
		}
		fmt.Printf("Found %d existing product JSON entries\n", len(jsonProducts))
		fetchJsonDuration := time.Since(startFetchJson)

		successCount := 0
		var totalProcessing time.Duration
		var totalDB time.Duration

		const batchSize = 100 // Adjust based on your database's max_allowed_packet
		var insertBatch []product.ProductJson
		var updateBatch []product.ProductJson

		for productID, attributes := range flatProducts {

			// Data processing timing
			processStart := time.Now()
			fullData := map[string]interface{}{}
			for k, v := range attributes {
				fullData[k] = v
			}

			jsonData, err := json.Marshal(fullData)
			if err != nil {
				fmt.Printf("Failed to marshal JSON for product %d: %v\n", productID, err)
				continue
			}
			totalProcessing += time.Since(processStart)

			// Database operation timing
			dbStart := time.Now()
			key := fmt.Sprintf("%d_%d", productID, 0)

			if index, exists := existingEntries[key]; exists {
				// Collect updates
				existing := jsonProducts[index]
				existing.Attributes = jsonData
				existing.UpdatedAt = time.Now()
				updateBatch = append(updateBatch, existing)

				// Batch update when threshold reached
				if len(updateBatch) >= batchSize {
					fmt.Printf("Processing update batch of %d items\n", len(updateBatch))
					if err := bulkUpdate(db, updateBatch, batchSize); err != nil {
						fmt.Printf("Batch update failed: %v\n", err)
					}
					updateBatch = updateBatch[:0] // Reset batch
				}
			} else {
				// Collect inserts
				insertBatch = append(insertBatch, product.ProductJson{
					EntityID:   productID,
					StoreID:    0,
					Attributes: jsonData,
				})

				// Batch insert when threshold reached
				if len(insertBatch) >= batchSize {
					fmt.Printf("Processing insert batch of %d items\n", len(insertBatch))
					if err := bulkInsert(db, insertBatch, batchSize); err != nil {
						fmt.Printf("Batch insert failed: %v\n", err)
					}
					insertBatch = insertBatch[:0] // Reset batch
				}
			}
			totalDB += time.Since(dbStart)

			successCount++
		}

		// Process remaining items in batches
		if len(updateBatch) > 0 {
			fmt.Printf("Processing final update batch of %d items\n", len(updateBatch))
			if err := bulkUpdate(db, updateBatch, batchSize); err != nil {
				fmt.Printf("Final batch update failed: %v\n", err)
			}
		}
		if len(insertBatch) > 0 {
			fmt.Printf("Processing final insert batch of %d items\n", len(insertBatch))
			if err := bulkInsert(db, insertBatch, batchSize); err != nil {
				fmt.Printf("Final batch insert failed: %v\n", err)
			}
		}

		totalDuration := time.Since(startTotal)

		fmt.Printf(`
=== Indexing Report ===
Total products:     %d
Successfully processed: %d
Total time:         %s
  - Data fetch:     %s
  - Data fetch JSON:     %s
  - Avg processing: %s/product
  - DB operations:  %s
=======================
`, len(flatProducts), successCount,
			totalDuration.Round(time.Millisecond),
			fetchDuration.Round(time.Millisecond),
			fetchJsonDuration.Round(time.Millisecond),
			(totalProcessing / time.Duration(len(flatProducts))).Round(time.Microsecond),
			totalDB.Round(time.Millisecond))
	},
}

func init() {
	rootCmd.AddCommand(migrateProductsCmd)
}

func bulkInsert(db *gorm.DB, batch []product.ProductJson, batchSize int) error {
	start := time.Now()
	defer func() {
		fmt.Printf("Inserted batch of %d items in %s\n", len(batch), time.Since(start))
	}()

	result := db.CreateInBatches(batch, batchSize)
	if result.Error != nil {
		return result.Error
	}
	fmt.Printf("Inserted %d records (batch size %d)\n", result.RowsAffected, batchSize)
	return nil
}

func bulkUpdate(db *gorm.DB, batch []product.ProductJson, batchSize int) error {
	start := time.Now()
	totalUpdated := int64(0)

	err := db.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < len(batch); i += batchSize {
			end := i + batchSize
			if end > len(batch) {
				end = len(batch)
			}
			chunk := batch[i:end]

			// Create a slice of update parameters
			updates := make([]map[string]interface{}, len(chunk))
			for i, item := range chunk {
				updates[i] = map[string]interface{}{
					"id":             item.ID,
					"attribute_json": item.Attributes, // Actual JSON value
					"updated_at":     time.Now(),
				}
			}

			// Batch update using actual values
			result := tx.Model(&product.ProductJson{}).
				Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "id"}},
					DoUpdates: clause.AssignmentColumns([]string{"attribute_json", "updated_at"}),
				}).
				Create(updates)

			if result.Error != nil {
				return result.Error
			}
			totalUpdated += result.RowsAffected
		}
		return nil
	})

	fmt.Printf("Updated %d records in %s\n", totalUpdated, time.Since(start))
	return err
}

// Helper to extract IDs from batch
func getIDs(batch []product.ProductJson) []uint {
	ids := make([]uint, len(batch))
	for i, item := range batch {
		ids[i] = item.ID
	}
	return ids
}
