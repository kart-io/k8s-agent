package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter provides rate limiting functionality using Redis.
type RateLimiter struct {
	redisClient *redis.Client
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
	}
}

// RateLimitConfig holds rate limit configuration.
type RateLimitConfig struct {
	RequestsPerMinute int           // Number of requests allowed per minute
	BurstSize         int           // Burst allowance
	KeyPrefix         string        // Redis key prefix
	BypassRoles       []string      // Roles that bypass rate limiting
	RetryAfter        time.Duration // How long to wait before retry
}

// DefaultForcedLogoutRateLimit returns default rate limit config for forced logout APIs.
func DefaultForcedLogoutRateLimit() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 100,
		BurstSize:         10,
		KeyPrefix:         "rate_limit:forced_logout:",
		BypassRoles:       []string{"superadmin"},
		RetryAfter:        60 * time.Second,
	}
}

// RateLimit middleware enforces rate limiting per user.
func (rl *RateLimiter) RateLimit(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract user ID from context (set by JWT middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			// If no user ID, skip rate limiting (should not happen if JWT middleware is before this)
			c.Next()
			return
		}

		// Check if user has bypass role
		if userRoles, exists := c.Get("actor_roles"); exists {
			if rl.hasBypassRole(userRoles.([]string), config.BypassRoles) {
				c.Header("X-RateLimit-Bypass", "true")
				c.Next()
				return
			}
		}

		// Build rate limit key
		key := fmt.Sprintf("%s%s", config.KeyPrefix, userID)

		// Check rate limit
		allowed, remaining, resetTime, err := rl.checkRateLimit(c.Request.Context(), key, config)
		if err != nil {
			// On error, allow request but log the error
			// Don't block user due to Redis issues
			c.Header("X-RateLimit-Error", "true")
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			// Rate limit exceeded
			retryAfter := int(time.Until(resetTime).Seconds())
			if retryAfter < 0 {
				retryAfter = int(config.RetryAfter.Seconds())
			}

			c.Header("Retry-After", strconv.Itoa(retryAfter))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": fmt.Sprintf("Rate limit exceeded. Maximum %d requests per minute allowed.", config.RequestsPerMinute),
				"details": map[string]interface{}{
					"limit":               config.RequestsPerMinute,
					"remaining":           0,
					"reset_at":            resetTime.Unix(),
					"retry_after_seconds": retryAfter,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit checks if request is within rate limit using sliding window algorithm.
func (rl *RateLimiter) checkRateLimit(ctx context.Context, key string, config RateLimitConfig) (allowed bool, remaining int, resetTime time.Time, err error) {
	now := time.Now()
	windowStart := now.Add(-1 * time.Minute)
	windowStartMs := windowStart.UnixMilli()
	nowMs := now.UnixMilli()

	// Use Redis sorted set for sliding window
	// Score = timestamp in milliseconds
	// Member = unique request ID (timestamp + random)

	pipe := rl.redisClient.Pipeline()

	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStartMs, 10))

	// Count requests in current window
	countCmd := pipe.ZCount(ctx, key, strconv.FormatInt(windowStartMs, 10), "+inf")

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, fmt.Errorf("redis pipeline error: %w", err)
	}

	count, err := countCmd.Result()
	if err != nil {
		return false, 0, time.Time{}, fmt.Errorf("get count error: %w", err)
	}

	// Check if within limit
	if count >= int64(config.RequestsPerMinute) {
		// Rate limit exceeded
		resetTime = now.Add(1 * time.Minute)
		return false, 0, resetTime, nil
	}

	// Add current request to window
	member := fmt.Sprintf("%d", nowMs)
	if err := rl.redisClient.ZAdd(ctx, key, redis.Z{
		Score:  float64(nowMs),
		Member: member,
	}).Err(); err != nil {
		return false, 0, time.Time{}, fmt.Errorf("zadd error: %w", err)
	}

	// Set expiration on key (cleanup)
	rl.redisClient.Expire(ctx, key, 2*time.Minute)

	// Calculate remaining requests
	remaining = config.RequestsPerMinute - int(count) - 1
	if remaining < 0 {
		remaining = 0
	}

	// Reset time is start of next minute window
	resetTime = now.Add(1 * time.Minute).Truncate(time.Minute)

	return true, remaining, resetTime, nil
}

// hasBypassRole checks if user has any bypass role.
func (rl *RateLimiter) hasBypassRole(userRoles, bypassRoles []string) bool {
	bypassMap := make(map[string]bool)
	for _, role := range bypassRoles {
		bypassMap[role] = true
	}

	for _, role := range userRoles {
		if bypassMap[role] {
			return true
		}
	}
	return false
}

// ResetRateLimit manually resets rate limit for a user (admin function).
func (rl *RateLimiter) ResetRateLimit(ctx context.Context, userID, keyPrefix string) error {
	key := fmt.Sprintf("%s%s", keyPrefix, userID)
	return rl.redisClient.Del(ctx, key).Err()
}

// GetRateLimitStatus returns current rate limit status for a user.
func (rl *RateLimiter) GetRateLimitStatus(ctx context.Context, userID, keyPrefix string, limit int) (used int, remaining int, resetTime time.Time, err error) {
	key := fmt.Sprintf("%s%s", keyPrefix, userID)
	now := time.Now()
	windowStart := now.Add(-1 * time.Minute)
	windowStartMs := windowStart.UnixMilli()

	count, err := rl.redisClient.ZCount(ctx, key, strconv.FormatInt(windowStartMs, 10), "+inf").Result()
	if err != nil {
		return 0, 0, time.Time{}, err
	}

	used = int(count)
	remaining = limit - used
	if remaining < 0 {
		remaining = 0
	}

	resetTime = now.Add(1 * time.Minute).Truncate(time.Minute)

	return used, remaining, resetTime, nil
}
