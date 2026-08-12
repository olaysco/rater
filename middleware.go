package rater

import (
	"fmt"
	"net/http"
)

func RateLimitMiddleware(limiter Limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//identify the user
		clientKey := r.Header.Get("X-API-Key")
		if clientKey == "" {
			clientKey = r.RemoteAddr
		}

		if !limiter.Allow(clientKey) {
			w.Header().Set("Retry-After", "1") // e.g., wait 1 second
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Content-Type", "application/json")

			// 2. Write HTTP status code
			w.WriteHeader(http.StatusTooManyRequests)

			// 3. Write JSON response body
			fmt.Fprintln(w, `{"error": "Too Many Requests", "message": "Rate limit exceeded. Try again later."}`)

			return
		}

		next.ServeHTTP(w, r)
	})
}
