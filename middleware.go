package rater

import (
	"fmt"
	"net/http"
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

func RateLimitMiddleware(limiter Limiter, next http.Handler) http.Handler {
	return RateLmitMiddlewareWithKeyFunc(limiter, defaultKeyFunc, next)
}

func RateLmitMiddlewareWithKeyFunc(limiter Limiter, keyFunc KeyFunc, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Identify the user
		clientKey := keyFunc(r)

		if !limiter.Allow(clientKey) {
			w.Header().Set("Retry-After", "1")
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintln(w, `{"error": "Too Many Requests", "message": "Rate limit exceeded. Try again later."}`)

			return
		}

		next.ServeHTTP(w, r)
	})
}
