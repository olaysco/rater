package rater

import (
	"log/slog"
	"net/http"

	"github.com/olaysco/rater/pkg/limiters"
)

// Configurable strategy for extracting client key
type KeyFunc func(r *http.Request) string

func defaultKeyFunc(r *http.Request) string {
	clientKey := r.Header.Get("X-API-KEY")
	if clientKey == "" {
		clientKey = r.RemoteAddr
	}

	return clientKey
}

// RateLimitMiddleware applies rate limiting using the default key function.
func RateLimitMiddleware(limiter limiters.Limiter, next http.Handler, logger *slog.Logger) http.Handler {
	return RateLmitMiddlewareWithKeyFunc(limiter, logger, defaultKeyFunc, next)
}

func RateLmitMiddlewareWithKeyFunc(limiter limiters.Limiter, logger *slog.Logger, keyFunc KeyFunc, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Identify the user
		clientKey := keyFunc(r)

		if !limiter.Allow(clientKey) {
			w.Header().Set("Retry-After", "1")
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)

			logger.Warn("rate limit exceeded",
				slog.String("client_key", clientKey),
				slog.String("path", r.URL.Path),
				slog.String("method", r.Method),
				slog.String("remote_addr", r.RemoteAddr),
				slog.Int("status", http.StatusTooManyRequests),
			)

			return
		}

		next.ServeHTTP(w, r)
	})
}
