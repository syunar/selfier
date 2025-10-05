// Package logger
package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/lmittmann/tint" // Import the tint handler
)

// Config defines baseline metadata for logs
type Config struct {
	Service string
	Env     string
	Version string
	Level   slog.Level
}

// NewLogger returns a base slog.Logger with common options
func NewLogger(cfg Config) *slog.Logger {
	// This common ReplaceAttr function can be shared by both handlers
	replaceAttr := func(groups []string, a slog.Attr) slog.Attr {
		switch a.Key {
		// The slog.TimeKey is not used by the tint handler, but is used by the JSON handler.
		// tint uses its own TimeFormat option.
		case slog.TimeKey:
			a.Value = slog.StringValue(time.Now().UTC().Format(time.RFC3339Nano))
		case slog.SourceKey:
			if src, ok := a.Value.Any().(*slog.Source); ok {
				dir, file := filepath.Split(src.File)
				parent := filepath.Base(filepath.Clean(dir))
				short := fmt.Sprintf("%s/%s:%d", parent, file, src.Line)
				a.Value = slog.StringValue(short)
			}
		}
		return a
	}

	// Use colorful tint logs in dev, JSON logs in prod
	var handler slog.Handler
	if cfg.Env == "development" {
		// Use the colorful tint handler for development
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:       cfg.Level,
			AddSource:   true,
			ReplaceAttr: replaceAttr,
			TimeFormat:  time.Kitchen, // A more human-readable time format
		})
	} else {
		// Use the standard JSON handler for production
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:       cfg.Level,
			AddSource:   true,
			ReplaceAttr: replaceAttr,
		})
	}

	return slog.New(handler).With(
		slog.String("service", cfg.Service),
		slog.String("env", cfg.Env),
		slog.String("version", cfg.Version),
	)
}
