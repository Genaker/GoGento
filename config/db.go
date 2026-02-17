package config

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func NewDB() (*gorm.DB, error) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = GetDBConnectionString()
	}

	logMode := logger.Info
	if os.Getenv("GORM_LOG") == "off" {
		logMode = logger.Silent
	}

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // Use log.Logger for Printf support
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logMode,     // Log level
			Colorful:      true,        // Enable color
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:      gormLogger,
		PrepareStmt: true, // Enable prepared statements
	})
	if err != nil {
		return nil, err
	}

	// Get generic database object
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(25)                 // Maximum open connections
	sqlDB.SetMaxIdleConns(25)                 // Maximum idle connections
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // Maximum connection lifetime
	sqlDB.SetConnMaxIdleTime(2 * time.Minute) // Maximum idle time

	return db, nil
}

// GetDBConnectionString returns formatted MySQL connection string
func GetDBConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASS"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DB"),
	)
}

// GetMigrationDSN returns migration-compatible DSN
func GetMigrationDSN() string {
	return fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASS"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DB"),
	)
}
