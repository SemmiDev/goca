package ratelimit

import (
	"context"
	"time"
)

// LimitResult represents the result of a rate limit check
type LimitResult struct {
	Limit      int64 // The total allowed requests in the period
	Remaining  int64 // The remaining requests allowed in the current period
	Reset      int64 // The time when the rate limit resets
	IsExceeded bool  // Whether the rate limit has been exceeded
}

// RateLimiter defines the interface for rate limiting operations
type RateLimiter interface {
	// Check checks if a request is allowed for the given key.
	// Returns a LimitResult indicating the rate limit status.
	Check(ctx context.Context, key string) (LimitResult, error)

	// Take consumes one request for the given key.
	// Returns a LimitResult indicating the rate limit status.
	Take(ctx context.Context, key string) (LimitResult, error)

	// GetLimit returns the configured rate limit (requests per period)
	GetLimit() int64

	// GetPeriod returns the configured period for the rate limit
	GetPeriod() time.Duration
}
