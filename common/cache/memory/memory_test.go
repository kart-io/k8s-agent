package memory

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/kart-io/k8s-agent/common/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryCache(t *testing.T) {
	ctx := context.Background()

	t.Run("Set and Get", func(t *testing.T) {
		c := NewMemoryCache()
		defer c.Close()

		err := c.Set(ctx, "key1", []byte("value1"), time.Minute)
		require.NoError(t, err)

		value, err := c.Get(ctx, "key1")
		require.NoError(t, err)
		assert.Equal(t, []byte("value1"), value)
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		c := NewMemoryCache()
		defer c.Close()

		_, err := c.Get(ctx, "nonexistent")
		assert.ErrorIs(t, err, cache.ErrKeyNotFound)
	})

	t.Run("Delete", func(t *testing.T) {
		c := NewMemoryCache()
		defer c.Close()

		err := c.Set(ctx, "key1", []byte("value1"), time.Minute)
		require.NoError(t, err)

		err = c.Delete(ctx, "key1")
		require.NoError(t, err)

		_, err = c.Get(ctx, "key1")
		assert.ErrorIs(t, err, cache.ErrKeyNotFound)
	})

	t.Run("Exists", func(t *testing.T) {
		c := NewMemoryCache()
		defer c.Close()

		exists, err := c.Exists(ctx, "key1")
		require.NoError(t, err)
		assert.False(t, exists)

		err = c.Set(ctx, "key1", []byte("value1"), time.Minute)
		require.NoError(t, err)

		exists, err = c.Exists(ctx, "key1")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Expiration", func(t *testing.T) {
		c := NewMemoryCache()
		defer c.Close()

		err := c.Set(ctx, "key1", []byte("value1"), time.Millisecond*100)
		require.NoError(t, err)

		value, err := c.Get(ctx, "key1")
		require.NoError(t, err)
		assert.Equal(t, []byte("value1"), value)

		time.Sleep(time.Millisecond * 150)

		_, err = c.Get(ctx, "key1")
		assert.ErrorIs(t, err, cache.ErrKeyNotFound)
	})

	t.Run("GetWithTTL", func(t *testing.T) {
		c := NewMemoryCache()
		defer c.Close()

		err := c.Set(ctx, "key1", []byte("value1"), time.Second*5)
		require.NoError(t, err)

		value, ttl, err := c.GetWithTTL(ctx, "key1")
		require.NoError(t, err)
		assert.Equal(t, []byte("value1"), value)
		assert.True(t, ttl > 0 && ttl <= time.Second*5)
	})

	t.Run("MGet and MSet", func(t *testing.T) {
		c := NewMemoryCache()
		defer c.Close()

		items := map[string][]byte{
			"key1": []byte("value1"),
			"key2": []byte("value2"),
			"key3": []byte("value3"),
		}

		err := c.MSet(ctx, items, time.Minute)
		require.NoError(t, err)

		results, err := c.MGet(ctx, "key1", "key2", "key3", "key4")
		require.NoError(t, err)
		assert.Equal(t, 3, len(results))
		assert.Equal(t, []byte("value1"), results["key1"])
		assert.Equal(t, []byte("value2"), results["key2"])
		assert.Equal(t, []byte("value3"), results["key3"])
	})

	t.Run("Increment and Decrement", func(t *testing.T) {
		c := NewMemoryCache()
		defer c.Close()

		value, err := c.Increment(ctx, "counter", 5)
		require.NoError(t, err)
		assert.Equal(t, int64(5), value)

		value, err = c.Increment(ctx, "counter", 3)
		require.NoError(t, err)
		assert.Equal(t, int64(8), value)

		value, err = c.Decrement(ctx, "counter", 2)
		require.NoError(t, err)
		assert.Equal(t, int64(6), value)
	})

	t.Run("Key prefix", func(t *testing.T) {
		c := NewMemoryCache(cache.WithKeyPrefix("test"))
		defer c.Close()

		err := c.Set(ctx, "key1", []byte("value1"), time.Minute)
		require.NoError(t, err)

		value, err := c.Get(ctx, "key1")
		require.NoError(t, err)
		assert.Equal(t, []byte("value1"), value)
	})

	t.Run("Clear with prefix", func(t *testing.T) {
		c := NewMemoryCache(cache.WithKeyPrefix("test"))
		defer c.Close()

		err := c.Set(ctx, "key1", []byte("value1"), time.Minute)
		require.NoError(t, err)

		err = c.Clear(ctx)
		require.NoError(t, err)

		_, err = c.Get(ctx, "key1")
		assert.ErrorIs(t, err, cache.ErrKeyNotFound)
	})
}

// TestMemoryCache_NoGoroutineLeak verifies that cleanup goroutine stops after Close().
func TestMemoryCache_NoGoroutineLeak(t *testing.T) {
	// Get initial goroutine count
	initialCount := countGoroutines()

	// Create and close multiple caches
	const numCaches = 10
	for i := 0; i < numCaches; i++ {
		c := NewMemoryCache()
		// Set some data to ensure cache is working
		ctx := context.Background()
		_ = c.Set(ctx, "key", []byte("value"), time.Minute)
		// Close should stop cleanup goroutine
		c.Close()
	}

	// Give time for goroutines to exit
	time.Sleep(100 * time.Millisecond)

	// Check final goroutine count
	finalCount := countGoroutines()

	// Allow small tolerance (test framework goroutines)
	if finalCount > initialCount+2 {
		t.Errorf("Goroutine leak detected: initial=%d, final=%d, diff=%d",
			initialCount, finalCount, finalCount-initialCount)
	}
}

// TestMemoryCache_CloseIdempotent verifies that Close can be called multiple times safely.
func TestMemoryCache_CloseIdempotent(t *testing.T) {
	c := NewMemoryCache()

	// First close should succeed
	err := c.Close()
	assert.NoError(t, err)

	// Second close should also succeed (idempotent)
	err = c.Close()
	assert.NoError(t, err)

	// Third close should also succeed
	err = c.Close()
	assert.NoError(t, err)
}

// countGoroutines returns the current number of goroutines.
func countGoroutines() int {
	// Force GC to clean up any pending goroutines
	time.Sleep(10 * time.Millisecond)
	return numGoroutine()
}

// numGoroutine returns number of goroutines using runtime.
func numGoroutine() int {
	// Use a simple method to count goroutines
	buf := make([]byte, 1<<16)
	stackLen := runtime.Stack(buf, true)
	// Count number of "goroutine" keywords in stack trace
	count := 0
	for i := 0; i < stackLen; {
		if i+9 < stackLen && string(buf[i:i+9]) == "goroutine" {
			count++
		}
		i++
	}
	return count
}

// TestMemoryCache_IncrementBoundaryValues tests edge cases for Increment/Decrement.
func TestMemoryCache_IncrementBoundaryValues(t *testing.T) {
	ctx := context.Background()

	t.Run("Large positive values", func(t *testing.T) {
		c := NewMemoryCache()
		defer c.Close()

		// Test with large positive number
		value, err := c.Increment(ctx, "counter", 1<<50)
		require.NoError(t, err)
		assert.Equal(t, int64(1<<50), value)

		// Increment again
		value, err = c.Increment(ctx, "counter", 1<<40)
		require.NoError(t, err)
		assert.Equal(t, int64(1<<50+1<<40), value)
	})

	t.Run("Large negative values", func(t *testing.T) {
		c := NewMemoryCache()
		defer c.Close()

		// Test with large negative number
		value, err := c.Increment(ctx, "counter", -(1 << 50))
		require.NoError(t, err)
		assert.Equal(t, int64(-(1 << 50)), value)

		// Decrement further
		value, err = c.Decrement(ctx, "counter", 1<<40)
		require.NoError(t, err)
		assert.Equal(t, int64(-(1<<50)-(1<<40)), value)
	})

	t.Run("Overflow behavior", func(t *testing.T) {
		c := NewMemoryCache()
		defer c.Close()

		// Set to max int64
		maxInt64 := int64(9223372036854775807) // 2^63 - 1
		value, err := c.Increment(ctx, "counter", maxInt64)
		require.NoError(t, err)
		assert.Equal(t, maxInt64, value)

		// Increment by 1 should overflow (wraps around in int64 arithmetic)
		value, err = c.Increment(ctx, "counter", 1)
		require.NoError(t, err)
		// This will overflow to min int64
		assert.Equal(t, int64(-9223372036854775808), value)
	})

	t.Run("Zero increment", func(t *testing.T) {
		c := NewMemoryCache()
		defer c.Close()

		// Start at 100
		value, err := c.Increment(ctx, "counter", 100)
		require.NoError(t, err)
		assert.Equal(t, int64(100), value)

		// Increment by 0 should keep same value
		value, err = c.Increment(ctx, "counter", 0)
		require.NoError(t, err)
		assert.Equal(t, int64(100), value)
	})

	t.Run("Alternating increment and decrement", func(t *testing.T) {
		c := NewMemoryCache()
		defer c.Close()

		var value int64
		var err error

		// Pattern: +10, -5, +10, -5
		for i := 0; i < 100; i++ {
			value, err = c.Increment(ctx, "counter", 10)
			require.NoError(t, err)

			value, err = c.Decrement(ctx, "counter", 5)
			require.NoError(t, err)
		}

		// Final value should be (10-5) * 100 = 500
		assert.Equal(t, int64(500), value)
	})

	t.Run("Multiple counters", func(t *testing.T) {
		c := NewMemoryCache()
		defer c.Close()

		// Create multiple independent counters
		v1, _ := c.Increment(ctx, "counter1", 10)
		v2, _ := c.Increment(ctx, "counter2", 20)
		v3, _ := c.Increment(ctx, "counter3", 30)

		assert.Equal(t, int64(10), v1)
		assert.Equal(t, int64(20), v2)
		assert.Equal(t, int64(30), v3)

		// Modify one counter
		v1, _ = c.Increment(ctx, "counter1", 5)
		assert.Equal(t, int64(15), v1)

		// Other counters should be unchanged
		v2, _ = c.Increment(ctx, "counter2", 0)
		v3, _ = c.Increment(ctx, "counter3", 0)
		assert.Equal(t, int64(20), v2)
		assert.Equal(t, int64(30), v3)
	})
}

// TestMemoryCache_ContextCancellation tests context cancellation support.
func TestMemoryCache_ContextCancellation(t *testing.T) {
	t.Run("MGet with cancelled context", func(t *testing.T) {
		c := NewMemoryCache()
		defer c.Close()

		// Set some data first
		ctx := context.Background()
		items := map[string][]byte{
			"key1": []byte("value1"),
			"key2": []byte("value2"),
			"key3": []byte("value3"),
		}
		_ = c.MSet(ctx, items, time.Minute)

		// Create cancelled context
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// MGet should return context.Canceled error
		_, err := c.MGet(cancelledCtx, "key1", "key2", "key3")
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("MSet with cancelled context", func(t *testing.T) {
		c := NewMemoryCache()
		defer c.Close()

		// Create cancelled context
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		items := map[string][]byte{
			"key1": []byte("value1"),
			"key2": []byte("value2"),
		}

		// MSet should return context.Canceled error
		err := c.MSet(cancelledCtx, items, time.Minute)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("Clear with cancelled context", func(t *testing.T) {
		c := NewMemoryCache()
		defer c.Close()

		// Set some data first
		ctx := context.Background()
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("key%d", i)
			_ = c.Set(ctx, key, []byte("value"), time.Minute)
		}

		// Create cancelled context
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Clear should return context.Canceled error
		err := c.Clear(cancelledCtx)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("MGet with timeout context", func(t *testing.T) {
		c := NewMemoryCache()
		defer c.Close()

		// Set some data
		ctx := context.Background()
		items := map[string][]byte{
			"key1": []byte("value1"),
			"key2": []byte("value2"),
		}
		_ = c.MSet(ctx, items, time.Minute)

		// Create context with very short timeout
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// Wait for timeout
		time.Sleep(1 * time.Millisecond)

		// MGet should return context.DeadlineExceeded error
		_, err := c.MGet(timeoutCtx, "key1", "key2")
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("Operations succeed with valid context", func(t *testing.T) {
		c := NewMemoryCache()
		defer c.Close()

		ctx := context.Background()

		// MSet should succeed
		items := map[string][]byte{
			"key1": []byte("value1"),
			"key2": []byte("value2"),
		}
		err := c.MSet(ctx, items, time.Minute)
		assert.NoError(t, err)

		// MGet should succeed
		results, err := c.MGet(ctx, "key1", "key2")
		assert.NoError(t, err)
		assert.Equal(t, 2, len(results))

		// Clear should succeed
		err = c.Clear(ctx)
		assert.NoError(t, err)
	})
}
