package ratelimit

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisStore is a Redis-backed implementation of RateLimiterStore.
type RedisStore struct {
	client *redis.Client
	script *redis.Script
}

// NewRedisStore creates a new RedisStore.
func NewRedisStore(address string) *RedisStore {
	rdb := redis.NewClient(&redis.Options{
		Addr: address,
	})

	// This Lua script atomically performs the sliding window rate limit check.
	// KEYS[1]: The key for the sorted set (e.g., "ratelimit:mygroup")
	// ARGV[1]: The current timestamp (nanoseconds)
	// ARGV[2]: The window size (nanoseconds)
	// ARGV[3]: The maximum number of requests in the window
	// Returns: 1 if the request is allowed, 0 if it's denied.
	luaScript := `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local window_start = now - window

-- Remove old entries from the sorted set
redis.call('ZREMRANGEBYSCORE', key, 0, window_start)

-- Get the current count of requests in the window
local current_count = redis.call('ZCARD', key)

	-- Check if the limit has been reached
	if current_count > limit then
	  return 0
	end

-- Add the new request timestamp and set expiration
redis.call('ZADD', key, now, now)
redis.call('PEXPIRE', key, window / 1000000 + 1000) -- Expire key after window + 1s buffer in ms

return 1
`

	return &RedisStore{
		client: rdb,
		script: redis.NewScript(luaScript),
	}
}

// Allow checks if a request for a given key is allowed using the Redis script.
func (s *RedisStore) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	now := time.Now().UnixNano()
	windowNano := window.Nanoseconds()

	result, err := s.script.Run(ctx, s.client, []string{key}, now, windowNano, limit).Result()
	if err != nil {
		return false, err
	}

	// The script returns 1 for allowed, 0 for denied.
	return result.(int64) == 1, nil
}
