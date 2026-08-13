package main

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	RedisAddr    string
	ServiceAddr  string
	LimiterType  string
	MaxRequests  int64
	Window       time.Duration
	SyncInterval time.Duration
}

func NewConfig() *Config {
	return &Config{
		RedisAddr:    GetOr("REDIS_ADDR", "localhost:6379"),
		ServiceAddr:  GetOr("SERVICE_ADDR", "google.com"),
		LimiterType:  GetOr("LIMITER_TYPE", "hybrid"),
		MaxRequests:  GetOr[int64]("MAX_REQUESTS", 50),
		Window:       GetOr("WINDOW", 2*time.Minute),
		SyncInterval: GetOr("SYNC_INTERVAL", 100*time.Millisecond),
	}
}

func GetOr[T string | int64 | time.Duration](key string, or T) T {
	val := os.Getenv(key)
	if val == "" {
		return or
	}

	var parsed any
	switch any(or).(type) {
	case string:
		parsed = val
	case int64:
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return or
		}
		parsed = v
	case time.Duration:
		v, err := time.ParseDuration(val)
		if err != nil {
			return or
		}
		parsed = v
	}

	return parsed.(T)
}
