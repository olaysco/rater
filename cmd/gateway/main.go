package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-redis/redis"
	"github.com/olaysco/rater"
)

func main() {
	targeturl, e := url.Parse("http://olays.co")
	if e != nil {
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targeturl)

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	if err := redisClient.Ping().Err(); err != nil {
		log.Fatalf("Couldnt connect to Redis: %v", err)
	}

	limiter, err := rater.NewRedisLimiter(redisClient, 50, 10)
	if err != nil {
		log.Fatalf("Failed to initialize Redis Limiter: %v", err)
	}

	handler := rater.RateLimitMiddleware(limiter, proxy)
	fmt.Println("API Gateway running on :8081 -> Proxying to :8081")
	http.ListenAndServe(":8081", handler)
}
