package limiters

import (
	"sync"
	"time"
)

type Token struct {
	count          float64
	lastRefillTime time.Time
}

type TokenBucket struct {
	mu         sync.Mutex
	capacity   float64
	refillRate float64
	tokens     map[string]*Token
}

func NewTokenBucket(capacity float64, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		refillRate: refillRate,
		tokens:     make(map[string]*Token),
	}
}

func (tb *TokenBucket) Allow(userKey string) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()

	userToken, ok := tb.tokens[userKey]
	if !ok {
		tb.tokens[userKey] = &Token{
			count:          tb.capacity,
			lastRefillTime: now,
		}

		return true
	}

	elapsed := now.Sub(userToken.lastRefillTime)
	token := elapsed.Seconds() * tb.refillRate
	userToken.count = min(tb.capacity, userToken.count+token)
	userToken.lastRefillTime = now

	if userToken.count >= 1 {
		userToken.count -= 1

		return true
	}

	return false
}
