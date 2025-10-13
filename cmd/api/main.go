package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"selfier/internal/modules/job"
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

func parseLogLevel(env string) slog.Level {
	switch strings.ToLower(env) {
	case "development":
		return slog.LevelDebug
	default:
		return slog.LevelInfo // default fallback
	}
}

func main() {

	// 1. Load configuration
	cfg := config.LoadConfig()

	// 2. Initialize logger
	log := logger.NewLogger(logger.Config{
		Service: cfg.Primary.ServiceName,
		Env:     cfg.Primary.Env,
		Version: cfg.Primary.Version,
		Level:   parseLogLevel(cfg.Primary.Env),
	})

	if parseLogLevel(cfg.Primary.Env) == slog.LevelDebug {
		log.Debug("debug mode enabled")
		log.Debug("config", slog.Any("config", cfg))
	}

	// 3. Dependencies

	db := database.NewDatabase(&cfg.Database)
	s3Client := aws.NewS3(&cfg.AWS)
	inngestClient := inngest.NewClient(&cfg.Inngest)

	jobRepo := job.NewJobRepo(db)
	taskRepo := job.NewTaskRepo(db)
	imageRepo := job.NewImageRepo(db)
	taskCreatedProducer := job.NewTaskCreatedProducerInngest(inngestClient)
	storage := job.NewStorageS3(s3Client)
	deselfiePipeline := job.NewDeselfiePipelineAPI()

	jobService := job.NewService(
		jobRepo,
		taskRepo,
		imageRepo,
		taskCreatedProducer,
		storage,
		deselfiePipeline,
	)

	jobHTTPHandler := job.NewHandler(jobService)
	jobEventHandler := job.NewEventHandlerIngest(jobService)

	// 4. Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router.NewRouter(jobHTTPHandler, jobEventHandler, &cfg.Inngest, &cfg.Server, log),
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start combined server", slog.String("error", err.Error()))
		}
	}()

	log.Info("server started", slog.String("docs", fmt.Sprintf("http://localhost:%s/docs", cfg.Server.Port)))

	// 5. Graceful shutdown
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

	log.Info("Server stopped")
}
