// Package middleware
package middleware

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

func RequestIDMiddleware() func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		reqID := ctx.Header("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}

		ctx.SetHeader("X-Request-ID", reqID)
		ctx = huma.WithValue(ctx, "request_id", reqID)

		next(ctx)
	}

}

func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value("request_id").(string); ok {
		return v
	}
	return ""
}
