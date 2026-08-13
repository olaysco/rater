package main

import "os"

type Config struct {
	RedisAddr   string
	ServiceAddr string
}

func NewConfig() *Config {
	return &Config{
		RedisAddr:   GetOr("REDIS_ADDR", "localhost:6379"),
		ServiceAddr: GetOr("SERVICE_ADDR", "google.com"),
	}
}

func GetOr[T string](key string, or T) T {
	val := os.Getenv(key)
	if val == "" {
		return or
	}

	return T(val)
}
