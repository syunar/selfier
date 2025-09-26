// Package database
package database

import (
	"fmt"
	"log"
	"time"

	"selfier/pkg/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewConnection creates and returns a new GORM DB instance based on the provided configuration.
// It also configures the connection pool and pings the database to ensure connectivity.
func NewConnection(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// 1. Create the Data Source Name (DSN) string from the config

	// 3. Open the database connection
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{ //nolint:exhaustruct
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 4. Get the underlying sql.DB object to configure the connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 5. Set connection pool settings from the configuration
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Minute)

	// 6. Ping the database to verify the connection is alive
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established successfully.")
	return db, nil
}
