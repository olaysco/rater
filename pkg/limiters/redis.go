package limiters

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLimiter struct {
	client     *redis.Client
	scriptSHA  string
	capacity   int
	refillRate float64
}

// Lua script for token bucket algorithm in Redis
var tokenBucketScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refillRate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local data = redis.call('HMGET', key, 'tokens', 'last_refill')
local tokens = tonumber(data[1])
local lastRefill = tonumber(data[2])

if not tokens then
    tokens = capacity
    lastRefill = now
end

local elapsed = math.max(0, now - lastRefill)
tokens = math.min(capacity, tokens + (elapsed * refillRate))

if tokens >= 1 then
    tokens = tokens - 1
    redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
    return 1
else
    redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
    return 0
end
`

func NewRedisLimiter(client *redis.Client, capacity int, refillRate float64) (*RedisLimiter, error) {
	// Load the Lua script once and use its SHA for future reference
	sha, err := client.ScriptLoad(context.Background(), tokenBucketScript).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to load Redis Lua script: %w", err)
	}

	return &RedisLimiter{
		client:     client,
		scriptSHA:  sha,
		capacity:   capacity,
		refillRate: refillRate,
	}, nil
}

// Allow checks if a request is allowed for the given userId using Redis as the backend.
func (rl *RedisLimiter) Allow(userId string) bool {
	// Get the current time in seconds with fractional part
	now := float64(time.Now().UnixNano()) / 1e9
	allow, err := rl.client.EvalSha(context.Background(), rl.scriptSHA, []string{userId}, rl.capacity, rl.refillRate, now).Int()
	if allow > 0 || err != nil {
		return true
	}

	return false
}
