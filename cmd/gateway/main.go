package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/olaysco/rater"
	"github.com/olaysco/rater/pkg/limiters"
	"github.com/redis/go-redis/v9"
)

func main() {
	config := NewConfig()
	ctx := context.Background()
	targeturl, e := url.Parse(config.ServiceAddr)
	if e != nil {
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targeturl)
	// Preserve original director behavior
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = targeturl.Host
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: config.RedisAddr,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Couldnt connect to Redis: %v", err)
	}

	limiter, err := limiters.NewLimiter(
		config.LimiterType,
		limiters.WithRedis(redisClient),
		limiters.WithLimits(config.MaxRequests, config.Window),
		limiters.WithSyncInterval(config.SyncInterval),
	)

	if err != nil {
		log.Fatalf("Failed to initialize Redis Limiter: %v", err)
	}

	handler := rater.RateLimitMiddleware(limiter, proxy)
	fmt.Println("API Gateway running on :8080 -> Proxying to ", config.ServiceAddr)
	http.ListenAndServe(":8080", handler)
}
