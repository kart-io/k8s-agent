package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter implements token bucket rate limiting using Redis.
// It uses Redis INCR and EXPIRE for atomic counter operations.
type RateLimiter struct {
	client *Client
	rate   int           // Maximum requests allowed
	window time.Duration // Time window for rate limiting
}

// Lua script for atomic rate limit check.
// This ensures INCR and EXPIRE are executed atomically.
var rateLimitScript = redis.NewScript(`
	local current = redis.call("INCR", KEYS[1])
	if current == 1 then
		redis.call("EXPIRE", KEYS[1], ARGV[1])
	end
	return current
`)

// NewRateLimiter creates a new rate limiter.
//
// Parameters:
//   - client: Redis client instance
//   - rate: Maximum number of requests allowed
//   - window: Time window for the rate limit
//
// Example usage:
//
//	// Allow 100 requests per minute
//	limiter := redis.NewRateLimiter(client, 100, time.Minute)
//	allowed, err := limiter.Allow(ctx, "user:123")
//	if err != nil {
//	    return err
//	}
//	if !allowed {
//	    return errors.New("rate limit exceeded")
//	}
func NewRateLimiter(client *Client, rate int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		client: client,
		rate:   rate,
		window: window,
	}
}

// Allow checks if a request is allowed under the rate limit.
// It increments the counter for the given key and checks if it exceeds the rate.
//
// Parameters:
//   - ctx: Context for timeout control
//   - key: Unique identifier for rate limiting (e.g., user ID, IP address)
//
// Returns:
//   - bool: true if request is allowed, false if rate limit exceeded
//   - error: Error if operation fails
func (r *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	rateLimitKey := fmt.Sprintf("ratelimit:%s", key)

	// Use Lua script to atomically increment and set expiration
	count, err := rateLimitScript.Run(
		ctx,
		r.client.client,
		[]string{rateLimitKey},
		int(r.window.Seconds()),
	).Int64()
	if err != nil {
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}

	allowed := count <= int64(r.rate)

	if !allowed {
		r.client.logger.Debugw("Rate limit exceeded",
			"key", key,
			"count", count,
			"limit", r.rate,
		)
	}

	return allowed, nil
}

// Reset resets the rate limit counter for a given key.
// This can be used for administrative purposes or testing.
func (r *RateLimiter) Reset(ctx context.Context, key string) error {
	rateLimitKey := fmt.Sprintf("ratelimit:%s", key)
	if err := r.client.client.Del(ctx, rateLimitKey).Err(); err != nil {
		return fmt.Errorf("failed to reset rate limit: %w", err)
	}

	r.client.logger.Debugw("Rate limit reset", "key", key)
	return nil
}

// GetCount returns the current count for a given key.
// Returns 0 if the key doesn't exist.
func (r *RateLimiter) GetCount(ctx context.Context, key string) (int64, error) {
	rateLimitKey := fmt.Sprintf("ratelimit:%s", key)
	count, err := r.client.client.Get(ctx, rateLimitKey).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // Key doesn't exist, count is 0
		}
		return 0, fmt.Errorf("failed to get count: %w", err)
	}
	return count, nil
}

// GetRemaining returns the number of requests remaining in the current window.
// Returns 0 if rate limit is exceeded.
func (r *RateLimiter) GetRemaining(ctx context.Context, key string) (int64, error) {
	count, err := r.GetCount(ctx, key)
	if err != nil {
		return 0, err
	}

	remaining := int64(r.rate) - count
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}
