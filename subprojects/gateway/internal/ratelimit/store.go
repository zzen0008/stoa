package ratelimit

import (
	"context"
	"time"
)

// RateLimiterStore defines the interface for rate limiting storage.
type RateLimiterStore interface {
	// Allow checks if a request for a given key is allowed.
	// It returns true if the request is within the limit, and false otherwise.
	Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error)
}
