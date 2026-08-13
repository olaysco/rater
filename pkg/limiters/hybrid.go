package limiters

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type HybridLimiter struct {
	mu           sync.Mutex // Mutex to protect access to only configuration
	maxRequests  int64
	window       time.Duration
	syncInterval time.Duration

	localCache sync.Map // local L1 cache
	stopChan   chan struct{}

	client *redis.Client // Redis client interface for L2 storage
}

type LocalState struct {
	mu                sync.Mutex
	cachedGlobalCount int64
	localDelta        int64
	lastUpdated       time.Time
}

func NewHybridLimiter(client *redis.Client, maxRequests int64, window time.Duration, syncInterval time.Duration) *HybridLimiter {
	hl := &HybridLimiter{
		maxRequests:  maxRequests,
		window:       window,
		syncInterval: syncInterval,
		stopChan:     make(chan struct{}),

		client: client,
	}

	go hl.initSync()
	return hl
}

// Allow checks if a request is allowed for the given userKey in local cache.
func (hl *HybridLimiter) Allow(userKey string) bool {
	now := time.Now()

	value, ok := hl.localCache.Load(userKey)

	// Double check pattern to ensure that only one goroutine initializes the LocalState for a userKey
	if !ok {
		newState := &LocalState{
			cachedGlobalCount: 0,
			localDelta:        1,
			lastUpdated:       now,
		}

		currentState := newState
		current, alreadyStored := hl.localCache.LoadOrStore(userKey, currentState)
		if alreadyStored {
			// Another guy stored it first, use that one
			value = current
		} else {
			return true
		}
	}

	currentState := value.(*LocalState)
	currentState.mu.Lock()
	defer currentState.mu.Unlock()

	currentState.lastUpdated = now
	estimatedCount := currentState.cachedGlobalCount + currentState.localDelta
	if estimatedCount < hl.maxRequests {
		currentState.localDelta++
		return true
	}

	return false
}

func (hl *HybridLimiter) syncToStore() {
	hl.localCache.Range(func(key, value interface{}) bool {
		userKey := key.(string)
		state := value.(*LocalState)

		// Reset localDelta and release lock immediately
		state.mu.Lock()
		deltaToSend := state.localDelta
		state.localDelta = 0
		lastUpdated := state.lastUpdated
		state.mu.Unlock()

		// Cleanup idle keys
		if deltaToSend == 0 && time.Since(lastUpdated) > hl.window {
			hl.localCache.Delete(userKey)
			return true
		}

		redisKey := "rater:hybrid:" + userKey
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

		if deltaToSend == 0 {
			count, err := hl.client.Get(ctx, redisKey).Int64()
			if err == redis.Nil {
				count = 0
			} else if err != nil {
				cancel()
				return true
			}

			state.mu.Lock()
			state.cachedGlobalCount = count
			state.mu.Unlock()

			cancel()
			return true
		}

		newGlobalTotal, err := hl.client.IncrBy(ctx, redisKey, deltaToSend).Result()
		if err != nil {
			state.mu.Lock()
			state.localDelta += deltaToSend
			state.mu.Unlock()
			cancel()
			return true
		}

		// Set key TTL on first write
		hl.client.ExpireNX(ctx, redisKey, hl.window)
		cancel()

		state.mu.Lock()
		state.cachedGlobalCount = newGlobalTotal
		state.mu.Unlock()

		return true
	})
}

func (hl *HybridLimiter) initSync() {
	ticker := time.NewTicker(hl.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hl.syncToStore()
		case <-hl.stopChan:
			return
		}
	}
}

func (hl *HybridLimiter) Close() {
	close(hl.stopChan)
}
