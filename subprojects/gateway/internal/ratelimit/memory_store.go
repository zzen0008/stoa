package ratelimit

import (
	"context"
	"sync"
	"time"
)

// MemoryStore is an in-memory implementation of RateLimiterStore.
type MemoryStore struct {
	mu      sync.Mutex
	windows map[string][]int64
}

// NewMemoryStore creates a new MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		windows: make(map[string][]int64),
	}
}

// Allow checks if a request for a given key is allowed.
func (s *MemoryStore) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixNano()
	windowStart := now - window.Nanoseconds()

	// Remove timestamps older than the window
	timestamps := s.windows[key]
	validTimestamps := make([]int64, 0, len(timestamps))
	for _, ts := range timestamps {
		if ts > windowStart {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// Check if the limit is exceeded
	if int64(len(validTimestamps)) >= limit {
		s.windows[key] = validTimestamps
		return false, nil
	}

	// Add the new timestamp and allow the request
	s.windows[key] = append(validTimestamps, now)
	return true, nil
}
