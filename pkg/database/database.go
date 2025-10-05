// Package database
package database

import (
	"log"
	"os"
	"time"

	"selfier/pkg/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDatabase creates and returns a new GORM DB instance based on the provided configuration.
// It also configures the connection pool and pings the database to ensure connectivity.
func NewDatabase(cfg *config.DatabaseConfig) *gorm.DB {

	// 3. Open the database connection
	gormLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             500 * time.Millisecond,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: false,
		Colorful:                  true,
	})
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		panic(err)
	}

	// 4. Get the underlying sql.DB object to configure the connection pool
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	// 5. Set connection pool settings from the configuration
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Minute)

	// 6. Ping the database to verify the connection is alive
	if err := sqlDB.Ping(); err != nil {
		panic(err)
	}

	return db
}
