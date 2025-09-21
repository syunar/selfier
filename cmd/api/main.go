package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"

	jobAdapter "selfier/internal/job/adapter"
	jobCore "selfier/internal/job/core"
	"selfier/pkg/config"
	"selfier/pkg/database"
	"selfier/pkg/logger"
	"selfier/pkg/router"
)

// application holds the core dependencies for the application.
// This makes dependency injection cleaner and more manageable.
type application struct {
	config *config.Config
	logger zerolog.Logger
	db     *gorm.DB
}

func main() {
	// === 1. Bootstrap Phase ===
	// Load configuration and initialize logger.
	// We use the standard log package only for failures during this critical phase.
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	appLogger := bootstrapLogger(cfg)
	appLogger.Info().Msg("Configuration and logger initialized")

	// Create an application container to hold our dependencies.
	//nolint:exhaustruct
	app := &application{
		config: cfg,
		logger: appLogger,
	}

	// === 2. Dependency Injection Phase ===
	// Set up database, repositories, services, and handlers.

	// Database connection
	gormLogger := bootstrapGormLogger(cfg, app.logger)
	app.db, err = database.NewConnection(&cfg.Database, &gormLogger)
	if err != nil {
		app.logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	app.logger.Info().Msg("Database connection established")

	// Initialize application layers (repository -> service -> handler)
	jobRep, err := jobAdapter.NewGormJobRepository(app.db)
	if err != nil {
		app.logger.Fatal().Err(err).Msg("Failed to create job repository")
	}
	jobService := jobCore.NewJobService(jobRep)
	jobHandler := jobAdapter.NewJobHTTPHandler(jobService)

	// === 3. Server Initialization and Graceful Shutdown ===
	// Set up and run the HTTP server with graceful shutdown logic.

	// Explicitly create the http.Server for better control
	//nolint:exhaustruct
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", app.config.Server.Port),
		Handler:      router.NewRouter(jobHandler), // Assuming NewRouter returns an http.Handler
		IdleTimeout:  time.Duration(app.config.Server.IdleTimeout) * time.Second,
		ReadTimeout:  time.Duration(app.config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(app.config.Server.WriteTimeout) * time.Second,
	}

	// Run the server in a goroutine so it doesn't block.
	go func() {
		app.logger.Info().Str("addr", srv.Addr).Msg("Starting server")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.logger.Fatal().Err(err).Msg("Server startup failed")
		}
	}()

	// Wait for an interrupt signal to gracefully shut down the server.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block until a signal is received.

	app.logger.Warn().Msg("Shutdown signal received, starting graceful shutdown")

	// Create a context with a timeout to allow existing requests to finish.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Perform cleanup
	if err := srv.Shutdown(ctx); err != nil {
		app.logger.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	// Close database connection
	sqlDB, err := app.db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			app.logger.Error().Err(err).Msg("Error closing database connection")
		} else {
			app.logger.Info().Msg("Database connection closed successfully")
		}
	}

	app.logger.Info().Msg("Server exited gracefully")
}

// bootstrapLogger encapsulates the logic for setting up the main application logger.
func bootstrapLogger(cfg *config.Config) zerolog.Logger {

	logCfg := logger.Config{
		Level:       cfg.Logger.Level,
		Format:      cfg.Logger.Format,
		ServiceName: cfg.Primary.ServiceName,
		Environment: cfg.Primary.Env,
		IsProd:      cfg.Primary.IsProd(),
	}

	writer := logger.NewWriter(logCfg)
	return logger.New(writer, logCfg)
}

// bootstrapGormLogger encapsulates the logic for setting up the GORM logger.
func bootstrapGormLogger(cfg *config.Config, appLogger zerolog.Logger) logger.GormLoggerAdapter {
	gormLogCfg := logger.GormLoggerConfig{
		SlowThreshold:             cfg.Database.GormLogger.SlowQueryThreshold,
		IgnoreRecordNotFoundError: cfg.Database.GormLogger.IgnoreRecordNotFound,
	}
	// Note: We are returning the adapter struct itself, not the interface,
	// because the database.NewConnection function likely expects the interface.
	// The type will be implicitly converted. This is just for type clarity in the function signature.
	return *logger.NewGormLoggerAdapter(appLogger, gormLogCfg).(*logger.GormLoggerAdapter)
}
