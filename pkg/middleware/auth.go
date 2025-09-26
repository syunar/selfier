package middleware

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
)

// AuthMiddleware simulates extracting user_id from header
func AuthMiddleware() func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		userID := ctx.Header("X-User-ID")
		if userID == "" {
			userID = "anonymous"
		}

		ctx = huma.WithValue(ctx, "user_id", userID)

		next(ctx)
	}
}

func GetUserID(ctx context.Context) string {
	if v, ok := ctx.Value("user_id").(string); ok {
		return v
	}
	return "anonymous"
}
