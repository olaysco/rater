package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/olaysco/rater"
)

func main() {
	targeturl, e := url.Parse("http://olays.co")
	if e != nil {
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targeturl)
	limiter := rater.NewSlidingWindowLog(time.Duration(5*time.Second), 2)

	handler := rater.RateLimitMiddleware(limiter, proxy)
	fmt.Println("🚀 API Gateway running on :8080 -> Proxying to :8081")
	http.ListenAndServe(":8081", handler)
}
