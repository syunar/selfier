package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"selfier/internal/module/aideselfie"
	"selfier/internal/module/event"
	"selfier/internal/module/job"
	"selfier/pkg/aws"
	"selfier/pkg/config"
	"selfier/pkg/database"
	"selfier/pkg/inngest"
	"selfier/pkg/logger"
	"selfier/pkg/router"
	"strings"
	"syscall"
	"time"
)

func parseLogLevel(levelStr string) slog.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo // default fallback
	}
}

func main() {

	// 1. Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	log.Printf("config: %+v", cfg)

	// 2. Initialize logger
	log := logger.NewLogger(logger.Config{
		Service: cfg.Primary.ServiceName,
		Env:     cfg.Primary.Env,
		Version: cfg.Primary.Version,
		Level:   parseLogLevel(cfg.Logger.Level),
	})

	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
	}
	log.Info("connected to database", slog.String("host", cfg.Database.Host))

	s3Client, err := aws.NewS3(&cfg.AWS)
	if err != nil {
		log.Error("failed to connect to s3", slog.String("error", err.Error()))
	}
	log.Info("connected to s3", slog.String("endpoint", cfg.AWS.EndpointURL))

	// inngest producer client
	producer := inngest.NewClient(&cfg.Inngest)

	// 4. Initialize dependencies
	jobRepository, _ := job.NewJobRepositoryGorm(db)
	JobImageRepository, _ := job.NewJobImageRepositoryGorm(db)
	jobObjectStorage := job.NewJobObjectStorageS3(s3Client, cfg.AWS.UploadBucket)
	jobEventPublisher := job.NewJobEventPublisherInngest(producer)

	aideselfieProvider := aideselfie.NewAideselfieProviderModal()
	aideselfieService := aideselfie.NewAideselfieService(aideselfieProvider)

	jobService := job.NewJobService(
		jobRepository,
		JobImageRepository,
		jobObjectStorage,
		jobEventPublisher,
	)
	jobHTTPHandler := job.NewJobHTTPHandler(jobService)

	eventHandler := event.NewEventHandler(jobService, aideselfieService)

	// 5. Start the HTTP server
	srv := &http.Server{ //nolint:exhaustruct
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router.NewRouter(jobHTTPHandler, log),
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	go func() {
		log.Info("Starting server", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Server startup failed", slog.String("error", err.Error()))
		}
	}()

	// Inngest Serve
	consumer := inngest.NewConsumerClient(&cfg.Inngest, eventHandler)
	go func() {
		log.Info(
			"Starting inngest consumer server",
			slog.String("addr", fmt.Sprintf(":%s", cfg.Inngest.Port)),
		)
		err = http.ListenAndServe(fmt.Sprintf(":%s", cfg.Inngest.Port), consumer.Serve())
	}()

	// 6. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Warn("Shutdown signal received, starting graceful shutdown")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", slog.String("error", err.Error()))
	}
	sqlDB, err := db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			log.Error("Error closing database connection", slog.String("error", err.Error()))
		} else {
			log.Info("Database connection closed successfully")
		}
	}
}
