package rater

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

type RedisLimiter struct {
	client     *redis.Client
	scriptSHA  string
	capacity   int
	refillRate float64
}

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
	sha, err := client.ScriptLoad(tokenBucketScript).Result()
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

func (rl *RedisLimiter) Allow(userId string) bool {

	allow, err := rl.client.EvalSha(rl.scriptSHA, []string{userId}, rl.capacity, rl.refillRate, time.Now().Unix()).Result()

	if allow.(int) <= 0 || err != nil {
		return false
	}

	return true
}
