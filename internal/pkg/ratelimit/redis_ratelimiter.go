package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sammidev/goca/internal/pkg/logger"
	"github.com/ulule/limiter/v3"
	redisStore "github.com/ulule/limiter/v3/drivers/store/redis"
)

// RedisRateLimiter is a RateLimiter implementation using Redis and ulule/limiter
type RedisRateLimiter struct {
	limiter *limiter.Limiter
	rate    limiter.Rate
	logger  logger.Logger
}

// NewRedisRateLimiter creates a new RedisRateLimiter with the specified Redis client, rate, and logger
func NewRedisRateLimiter(redisClient *redis.Client, prefix string, rate limiter.Rate, log logger.Logger) (*RedisRateLimiter, error) {
	store, err := redisStore.NewStoreWithOptions(redisClient, limiter.StoreOptions{
		Prefix: prefix,
	})
	if err != nil {
		log.Fatal("Failed to initialize rate limiter store", "error", err)
		return nil, err
	}

	limiterInstance := limiter.New(store, rate)
	return &RedisRateLimiter{
		limiter: limiterInstance,
		rate:    rate,
		logger:  log.WithComponent("ratelimit"),
	}, nil
}

// Check checks if a request is allowed for the given key without consuming it
func (r *RedisRateLimiter) Check(ctx context.Context, key string) (LimitResult, error) {
	limitContext, err := r.limiter.Peek(ctx, key)
	if err != nil {
		r.logger.Error("Failed to check rate limit", "key", key, "error", err)
		return LimitResult{}, err
	}

	return LimitResult{
		Limit:      limitContext.Limit,
		Remaining:  limitContext.Remaining,
		Reset:      limitContext.Reset,
		IsExceeded: limitContext.Reached,
	}, nil
}

// Take consumes one request for the given key
func (r *RedisRateLimiter) Take(ctx context.Context, key string) (LimitResult, error) {
	limitContext, err := r.limiter.Get(ctx, key)
	if err != nil {
		r.logger.Error("Failed to apply rate limit", "key", key, "error", err)
		return LimitResult{}, err
	}

	return LimitResult{
		Limit:      limitContext.Limit,
		Remaining:  limitContext.Remaining,
		Reset:      limitContext.Reset,
		IsExceeded: limitContext.Reached,
	}, nil
}

// GetLimit returns the configured rate limit (requests per period)
func (r *RedisRateLimiter) GetLimit() int64 {
	return r.rate.Limit
}

// GetPeriod returns the configured period for the rate limit
func (r *RedisRateLimiter) GetPeriod() time.Duration {
	return r.rate.Period
}
