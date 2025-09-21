// Package logger provides logging functionality for the application.
package logger

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Config holds all the configuration for the main application logger.
type Config struct {
	Level       string
	Format      string // "json" or "console"
	ServiceName string
	Environment string
	IsProd      bool
}

// stackTraceHook is a zerolog hook that adds stack traces to error logs.
// This avoids modifying the global zerolog state.
type stackTraceHook struct{}

//nolint:revive
func (h stackTraceHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level == zerolog.ErrorLevel {
		e.Stack()
	}
}

// New creates a new zerolog.Logger based on the provided configuration and writer.
func New(writer io.Writer, cfg Config) zerolog.Logger {
	logLevel, err := zerolog.ParseLevel(strings.ToLower(cfg.Level))
	if err != nil || cfg.Level == "" {
		logLevel = zerolog.InfoLevel // Default to InfoLevel if parsing fails
	}

	// Create a logger instance with context
	context := zerolog.New(writer).With().
		Timestamp().
		Str("service", cfg.ServiceName).
		Str("environment", cfg.Environment)

	// Add stack trace hook for error-level logs instead of setting global state.
	logger := context.Logger().Hook(stackTraceHook{}).Level(logLevel)

	return logger
}

// NewWriter creates an io.Writer for the logger based on the environment.
// It can optionally be wrapped with a New Relic writer for log forwarding.
func NewWriter(cfg Config) io.Writer {
	var writer io.Writer

	// In development, use a human-friendly console writer.
	if !cfg.IsProd || cfg.Format == "console" {
		writer = zerolog.ConsoleWriter{ //nolint:exhaustruct
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05",
		}
	} else {
		// In production, write structured JSON to stdout.
		writer = os.Stdout
	}

	return writer
}

// NewRelicApp initializes and returns a New Relic application instance.
// Returns (nil, nil) if New Relic is disabled or the license key is empty.
func NewRelicApp(appName, licenseKey string, enabled bool) (*newrelic.Application, error) {
	if !enabled || licenseKey == "" {
		return nil, nil // Not an error, just disabled
	}

	return newrelic.NewApplication(
		newrelic.ConfigAppName(appName),
		newrelic.ConfigLicense(licenseKey),
		newrelic.ConfigAppLogForwardingEnabled(true),
		newrelic.ConfigDistributedTracerEnabled(true),
	)
}

// WithTraceContext adds New Relic transaction context to a logger.
func WithTraceContext(ctx context.Context, logger zerolog.Logger) zerolog.Logger {
	// Check if a transaction is in the context
	if txn := newrelic.FromContext(ctx); txn != nil {
		// Get trace metadata from the transaction
		metadata := txn.GetTraceMetadata()

		return logger.With().
			Str("trace.id", metadata.TraceID).
			Str("span.id", metadata.SpanID).
			Logger()
	}
	return logger
}

// --- GORM Logger Adapter ---

// GormLoggerConfig holds configuration for the GORM logger adapter.
type GormLoggerConfig struct {
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

// GormLoggerAdapter adapts zerolog to be used as the GORM logger.
type GormLoggerAdapter struct {
	zlog zerolog.Logger
	cfg  GormLoggerConfig
}

// NewGormLoggerAdapter creates a new GORM logger adapter instance.
func NewGormLoggerAdapter(zlog zerolog.Logger, cfg GormLoggerConfig) gormlogger.Interface {
	return &GormLoggerAdapter{
		zlog: zlog.With().Str("component", "gorm").Logger(),
		cfg:  cfg,
	}
}

// LogMode sets the log mode. We control the level via zerolog, so this is a no-op.
//
//nolint:revive
func (l *GormLoggerAdapter) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	return &newLogger
}

// Info logs an info message.
//
//nolint:revive
func (l *GormLoggerAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	l.zlog.Info().Msgf(msg, data...)
}

// Warn logs a warning message.
//
//nolint:revive
func (l *GormLoggerAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.zlog.Warn().Msgf(msg, data...)
}

// Error logs an error message.
//
//nolint:revive
func (l *GormLoggerAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	l.zlog.Error().Msgf(msg, data...)
}

// Trace logs a SQL query.
//
//nolint:revive
func (l *GormLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.zlog.GetLevel() > zerolog.DebugLevel {
		return // Do not log if the level is higher than debug
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	logEvent := l.zlog.With().Dur("elapsed", elapsed).Int64("rows", rows).Str("sql", sql).Logger()

	switch {
	case err != nil && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.cfg.IgnoreRecordNotFoundError):
		logEvent.Error().Err(err).Msgf("gorm query error: %s", sql)
	case elapsed > l.cfg.SlowThreshold && l.cfg.SlowThreshold != 0:
		logEvent.Warn().Msgf("gorm slow query: %s", sql)
	default:
		logEvent.Debug().Msg("gorm query")
	}
}
