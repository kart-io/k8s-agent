package redis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcquireLock_Success(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	key := "test-resource"
	ttl := 10 * time.Second

	// Acquire lock
	lock, err := client.AcquireLock(ctx, key, ttl)
	require.NoError(t, err)
	require.NotNil(t, lock)
	assert.Equal(t, "lock:test-resource", lock.key)
	assert.NotEmpty(t, lock.value)
	assert.Equal(t, ttl, lock.ttl)
}

func TestAcquireLock_AlreadyLocked(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	key := "test-resource"
	ttl := 10 * time.Second

	// Acquire lock first time
	lock1, err := client.AcquireLock(ctx, key, ttl)
	require.NoError(t, err)
	require.NotNil(t, lock1)

	// Try to acquire again (should fail)
	lock2, err := client.AcquireLock(ctx, key, ttl)
	require.Error(t, err)
	assert.Nil(t, lock2)
	assert.Contains(t, err.Error(), "lock already held")
}

func TestLock_Release(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	key := "test-resource"
	ttl := 10 * time.Second

	// Acquire lock
	lock, err := client.AcquireLock(ctx, key, ttl)
	require.NoError(t, err)

	// Release lock
	err = lock.Release(ctx)
	require.NoError(t, err)

	// Should be able to acquire again
	lock2, err := client.AcquireLock(ctx, key, ttl)
	require.NoError(t, err)
	assert.NotNil(t, lock2)

	// Cleanup
	err = lock2.Release(ctx)
	require.NoError(t, err)
}

func TestLock_ReleaseExpiredLock(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	key := "test-resource"
	ttl := 1 * time.Second

	// Acquire lock
	lock, err := client.AcquireLock(ctx, key, ttl)
	require.NoError(t, err)

	// Fast forward time in miniredis to expire the lock
	mr.FastForward(2 * time.Second)

	// Try to release expired lock (should fail)
	err = lock.Release(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lock value mismatch")
}

func TestLock_ReleaseWrongValue(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	key := "test-resource"
	ttl := 10 * time.Second

	// Acquire lock
	lock, err := client.AcquireLock(ctx, key, ttl)
	require.NoError(t, err)

	// Manually change the lock value in Redis
	mr.Set(lock.key, "different-value")

	// Try to release (should fail due to value mismatch)
	err = lock.Release(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lock value mismatch")
}

func TestLock_Extend(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	key := "test-resource"
	ttl := 5 * time.Second

	// Acquire lock
	lock, err := client.AcquireLock(ctx, key, ttl)
	require.NoError(t, err)

	// Extend lock
	additionalTTL := 5 * time.Second
	err = lock.Extend(ctx, additionalTTL)
	require.NoError(t, err)
	assert.Equal(t, ttl+additionalTTL, lock.ttl)
}

func TestLock_ExtendExpiredLock(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	key := "test-resource"
	ttl := 1 * time.Second

	// Acquire lock
	lock, err := client.AcquireLock(ctx, key, ttl)
	require.NoError(t, err)

	// Fast forward time in miniredis to expire the lock
	mr.FastForward(2 * time.Second)

	// Try to extend expired lock (should fail)
	err = lock.Extend(ctx, 5*time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lock expired")
}

func TestLock_ExtendWithValueMismatch(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	key := "test-resource"
	ttl := 10 * time.Second

	// Acquire lock
	lock, err := client.AcquireLock(ctx, key, ttl)
	require.NoError(t, err)

	// Manually change the lock value in Redis
	mr.Set(lock.key, "different-value")

	// Try to extend (should fail due to value mismatch)
	err = lock.Extend(ctx, 5*time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lock value mismatch")
}

func TestLock_ConcurrentAccess(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	key := "test-resource"
	ttl := 1 * time.Second

	// Simulate concurrent access
	successCount := 0
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func() {
			lock, err := client.AcquireLock(ctx, key, ttl)
			if err == nil {
				successCount++
				time.Sleep(50 * time.Millisecond)
				_ = lock.Release(ctx)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Only one goroutine should have acquired the lock
	assert.Equal(t, 1, successCount)
}
