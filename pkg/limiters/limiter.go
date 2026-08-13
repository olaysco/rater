package limiters

import (
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter interface {
	Allow(key string) bool
}

type LimiterOptions struct {
	Type         string // "redis", "hybrid", "token_bucket", "sliding_window"
	Client       *redis.Client
	MaxRequests  int64
	Window       time.Duration
	SyncInterval time.Duration
	RefillRate   float64
}

const (
	REDIS                  = "redis"                  // Requires client, maxRequets and RefillRate
	HYBRID                 = "hybrid"                 // Requires client, maxRequets, window and SyncInterval
	TOKEN_BUCKET           = "token_bucket"           // Requires maxRequets and RefillRate
	SLIDING_WINDOW_COUNTER = "sliding_window_counter" // Requires window and maxRequets
)

var ErrInvalidArgs = errors.New("invalid arguments for limiter creation")

type Option func(*LimiterOptions)

func WithRedis(client *redis.Client) Option {
	return func(o *LimiterOptions) { o.Client = client }
}

func WithLimits(maxRequests int64, window time.Duration) Option {
	return func(o *LimiterOptions) {
		o.MaxRequests = maxRequests
		o.Window = window
	}
}

func WithSyncInterval(interval time.Duration) Option {
	return func(o *LimiterOptions) { o.SyncInterval = interval }
}

func WithRefillRate(refillRate float64) Option {
	return func(o *LimiterOptions) { o.RefillRate = refillRate }
}

// NewLiimiter creates a new Limiter.
func NewLimiter(limiterType string, opts ...Option) (Limiter, error) {
	options := &LimiterOptions{
		Type: limiterType,
	}

	for _, opt := range opts {
		opt(options)
	}

	switch limiterType {
	case REDIS:
		if options.Client == nil || options.MaxRequests <= 0 || options.RefillRate <= 0 {
			return nil, ErrInvalidArgs
		}
		return NewRedisLimiter(options.Client, int(options.MaxRequests), options.RefillRate)
	case HYBRID:
		if options.Client == nil || options.MaxRequests <= 0 || options.Window <= 0 || options.SyncInterval <= 0 {
			return nil, ErrInvalidArgs
		}
		return NewHybridLimiter(options.Client, options.MaxRequests, options.Window, options.SyncInterval), nil
	case TOKEN_BUCKET:
		if options.MaxRequests <= 0 || options.RefillRate <= 0 {
			return nil, ErrInvalidArgs
		}
		return NewTokenBucket(float64(options.MaxRequests), options.RefillRate), nil
	case SLIDING_WINDOW_COUNTER:
		if options.Window <= 0 || options.MaxRequests <= 0 {
			return nil, ErrInvalidArgs
		}
		return NewSlidingWindowCounter(options.Window, int(options.MaxRequests)), nil
	}

	return nil, ErrInvalidArgs
}
