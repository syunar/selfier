// Package middleware
package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

func LoggerMiddleware(base *slog.Logger) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {

		ctx = huma.WithValue(ctx, "logger", base)
		start := time.Now()
		next(ctx)
		duration := time.Since(start)

		log := GetLogger(ctx.Context())
		log.Info("request completed",
			slog.String("request_id", GetRequestID(ctx.Context())),
			slog.String("user_id", GetUserID(ctx.Context())),
			slog.String("method", ctx.Method()),
			slog.String("path", ctx.URL().Path),
			slog.Duration("duration_ms", duration),
			slog.Int("status", ctx.Status()),
		)
	}
}

func GetLogger(ctx context.Context) *slog.Logger {
	if v, ok := ctx.Value("logger").(*slog.Logger); ok {
		return v
	}
	return slog.Default()
}
