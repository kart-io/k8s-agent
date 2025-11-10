package redis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimiter(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	limiter := NewRateLimiter(client, 100, time.Minute)

	assert.NotNil(t, limiter)
	assert.Equal(t, 100, limiter.rate)
	assert.Equal(t, time.Minute, limiter.window)
}

func TestRateLimiter_Allow(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	limiter := NewRateLimiter(client, 5, time.Minute)
	key := "user:123"

	// First 5 requests should be allowed
	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(ctx, key)
		require.NoError(t, err)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// 6th request should be denied
	allowed, err := limiter.Allow(ctx, key)
	require.NoError(t, err)
	assert.False(t, allowed, "Request 6 should be denied")
}

func TestRateLimiter_Reset(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	limiter := NewRateLimiter(client, 3, time.Minute)
	key := "user:456"

	// Use up the rate limit
	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(ctx, key)
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	// Should be rate limited now
	allowed, err := limiter.Allow(ctx, key)
	require.NoError(t, err)
	assert.False(t, allowed)

	// Reset the rate limit
	err = limiter.Reset(ctx, key)
	require.NoError(t, err)

	// Should be allowed again
	allowed, err = limiter.Allow(ctx, key)
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestRateLimiter_GetCount(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	limiter := NewRateLimiter(client, 10, time.Minute)
	key := "user:789"

	// Initially count should be 0
	count, err := limiter.GetCount(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Make 3 requests
	for i := 0; i < 3; i++ {
		_, err := limiter.Allow(ctx, key)
		require.NoError(t, err)
	}

	// Count should be 3
	count, err = limiter.GetCount(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestRateLimiter_GetRemaining(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	limiter := NewRateLimiter(client, 10, time.Minute)
	key := "user:abc"

	// Initially remaining should be 10
	remaining, err := limiter.GetRemaining(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(10), remaining)

	// Make 3 requests
	for i := 0; i < 3; i++ {
		_, err := limiter.Allow(ctx, key)
		require.NoError(t, err)
	}

	// Remaining should be 7
	remaining, err = limiter.GetRemaining(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(7), remaining)

	// Use up all remaining
	for i := 0; i < 8; i++ {
		_, err := limiter.Allow(ctx, key)
		require.NoError(t, err)
	}

	// Remaining should be 0 (not negative)
	remaining, err = limiter.GetRemaining(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(0), remaining)
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	// Use a very short window for testing
	limiter := NewRateLimiter(client, 3, 2*time.Second)
	key := "user:expiry"

	// Use up the rate limit
	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(ctx, key)
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	// Should be rate limited now
	allowed, err := limiter.Allow(ctx, key)
	require.NoError(t, err)
	assert.False(t, allowed)

	// Fast forward time in miniredis to expire the key
	mr.FastForward(3 * time.Second)

	// Should be allowed again after window expiry
	allowed, err = limiter.Allow(ctx, key)
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestRateLimiter_DifferentKeys(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	limiter := NewRateLimiter(client, 2, time.Minute)

	// Use up rate limit for user1
	for i := 0; i < 2; i++ {
		allowed, err := limiter.Allow(ctx, "user:1")
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	// user1 should be rate limited
	allowed, err := limiter.Allow(ctx, "user:1")
	require.NoError(t, err)
	assert.False(t, allowed)

	// user2 should still be allowed (different key)
	allowed, err = limiter.Allow(ctx, "user:2")
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestRateLimiter_ConcurrentRequests(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	limiter := NewRateLimiter(client, 10, time.Minute)
	key := "user:concurrent"

	// Simulate concurrent requests
	successCount := 0
	done := make(chan bool, 15)

	for i := 0; i < 15; i++ {
		go func() {
			allowed, err := limiter.Allow(ctx, key)
			require.NoError(t, err)
			if allowed {
				successCount++
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 15; i++ {
		<-done
	}

	// Only 10 requests should have been allowed
	assert.Equal(t, 10, successCount)
}
