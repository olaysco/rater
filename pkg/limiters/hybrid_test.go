package limiters

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestHybridLimiter(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	window := 1 * time.Minute

	syncInterval := 50 * time.Millisecond
	limiter := NewHybridLimiter(client, 5, window, syncInterval)
	defer limiter.Close()

	userKey := "user_123"

	for i := 0; i < 5; i++ {
		if !limiter.Allow(userKey) {
			t.Errorf("expected request %d to be allowed", i+1)
		}
	}

	if limiter.Allow(userKey) {
		t.Error("expected 6th request to be denied")
	}

	time.Sleep(syncInterval + 20*time.Millisecond)

	redisKey := "rater:hybrid:" + userKey
	val, err := mr.Get(redisKey)
	if err != nil || val != "5" {
		t.Fatalf("expected Redis key %s to be 5, got %s (err: %v)", redisKey, val, err)
	}

	// We simulate the passage of time
	mr.FastForward(window + 2*time.Second)

	time.Sleep(syncInterval + 20*time.Millisecond)

	if !limiter.Allow(userKey) {
		t.Error("expected request after sync to be allowed")
	}
}
