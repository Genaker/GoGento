package cmd

import (
	"fmt"
	"magento.GO/config"
	"magento.GO/model/entity/product"
	_ "os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var migrateCmd = &cobra.Command{
	Use:   "db:migrate",
	Short: "Run database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		// Step 1: Run SQL migrations
		runSQLMigrations()

		// Step 2: Run GORM AutoMigrate
		runGORMMigrations()
	},
}

func runSQLMigrations() {
	dbURL := config.GetMigrationDSN()

	migrationsPath := filepath.Join("file://", config.GetBasePath(), "migrations")

	m, err := migrate.New(migrationsPath, dbURL)
	if err != nil {
		fmt.Printf("SQL migration initialization failed: %v\n", err)
		return
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		fmt.Printf("SQL migration failed: %v\n", err)
		return
	}

	fmt.Println("SQL migrations applied successfully")
}

func runGORMMigrations() {
	db, err := config.NewDB()
	if err != nil {
		fmt.Printf("GORM migration connection failed: %v\n", err)
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			&product.ProductJson{},
			// Add other models...
		)
	})

	if err != nil {
		fmt.Printf("GORM AutoMigrate failed: %v\n", err)
		return
	}

	fmt.Println("GORM model migrations completed")
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
