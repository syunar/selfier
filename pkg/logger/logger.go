// Package logger
package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
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
	opts := &slog.HandlerOptions{ //nolint:exhaustruct
		Level:     cfg.Level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
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
		},
	}

	// Use JSON logs in prod, text logs in dev
	var handler slog.Handler
	if cfg.Env == "development" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler).With(
		slog.String("service", cfg.Service),
		slog.String("env", cfg.Env),
		slog.String("version", cfg.Version),
	)
}
