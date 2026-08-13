package rater

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type AlwaysBlockLimiter struct{}

func (l *AlwaysBlockLimiter) Allow(key string) bool {
	return false
}

type AlwaysAllowLimiter struct{}

func (l *AlwaysAllowLimiter) Allow(key string) bool {
	return true
}

func TestRateLimitMiddleware_Blocked(t *testing.T) {
	limiter := &AlwaysBlockLimiter{}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimitMiddleware(limiter, nextHandler, logger)

	req := httptest.NewRequest("GET", "/api/v1/resource", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Result().StatusCode != http.StatusTooManyRequests {
		t.Errorf("expect http code %v got %v", http.StatusTooManyRequests, rec.Code)
	}
}

func TestRateLimitMiddleware_Allow(t *testing.T) {
	limiter := &AlwaysAllowLimiter{}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimitMiddleware(limiter, nextHandler, logger)

	req := httptest.NewRequest("GET", "/api/v1/resource", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Result().StatusCode != http.StatusOK {
		t.Errorf("expect http code %v got %v", http.StatusOK, rec.Code)
	}
}
