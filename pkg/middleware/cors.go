package middleware

import (
	"log/slog"
	"net/http"
)

// CORSMiddleware is a standard Go HTTP middleware that handles CORS preflight requests.
// It wraps an http.Handler (like your main router) and intercepts requests.
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// This is the actual handler that will be called for each request.
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				// Not a CORS request, pass it to the next handler.
				next.ServeHTTP(w, r)
				return
			}

			isAllowed := false
			for _, o := range allowedOrigins {
				if o == origin || o == "*" {
					isAllowed = true
					break
				}
			}

			slog.Info("CORS MIDDLEWARE CHECK",
				"request_origin", origin,
				"request_method", r.Method,
				"allowed_origins", allowedOrigins,
				"is_allowed", isAllowed,
			)

			if isAllowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Add("Vary", "Origin") // Best practice for caching
			}

			// Handle the preflight OPTIONS request specifically.
			if r.Method == http.MethodOptions {
				if isAllowed {
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")

					// Reflect the requested headers, which is more robust than hardcoding.
					requestedHeaders := r.Header.Get("Access-Control-Request-Headers")
					if requestedHeaders != "" {
						w.Header().Set("Access-Control-Allow-Headers", requestedHeaders)
					}

					w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
					w.WriteHeader(http.StatusNoContent)
					return // IMPORTANT: Stop processing here for preflight.
				} else {
					// Deny preflight from a disallowed origin.
					w.WriteHeader(http.StatusForbidden)
					return
				}
			}

			// For all other requests, pass them on to the router.
			next.ServeHTTP(w, r)
		})
	}
}
