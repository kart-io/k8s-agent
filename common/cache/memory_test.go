package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/kart-io/k8s-agent/common/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryCache(t *testing.T) {
	ctx := context.Background()

	t.Run("Set and Get", func(t *testing.T) {
		c := cache.NewMemoryCache()
		defer c.Close()

		err := c.Set(ctx, "key1", []byte("value1"), time.Minute)
		require.NoError(t, err)

		value, err := c.Get(ctx, "key1")
		require.NoError(t, err)
		assert.Equal(t, []byte("value1"), value)
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		c := cache.NewMemoryCache()
		defer c.Close()

		_, err := c.Get(ctx, "nonexistent")
		assert.ErrorIs(t, err, cache.ErrKeyNotFound)
	})

	t.Run("Delete", func(t *testing.T) {
		c := cache.NewMemoryCache()
		defer c.Close()

		err := c.Set(ctx, "key1", []byte("value1"), time.Minute)
		require.NoError(t, err)

		err = c.Delete(ctx, "key1")
		require.NoError(t, err)

		_, err = c.Get(ctx, "key1")
		assert.ErrorIs(t, err, cache.ErrKeyNotFound)
	})

	t.Run("Exists", func(t *testing.T) {
		c := cache.NewMemoryCache()
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
		c := cache.NewMemoryCache()
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
		c := cache.NewMemoryCache()
		defer c.Close()

		err := c.Set(ctx, "key1", []byte("value1"), time.Second*5)
		require.NoError(t, err)

		value, ttl, err := c.GetWithTTL(ctx, "key1")
		require.NoError(t, err)
		assert.Equal(t, []byte("value1"), value)
		assert.True(t, ttl > 0 && ttl <= time.Second*5)
	})

	t.Run("MGet and MSet", func(t *testing.T) {
		c := cache.NewMemoryCache()
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
		c := cache.NewMemoryCache()
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
		c := cache.NewMemoryCache(cache.WithKeyPrefix("test"))
		defer c.Close()

		err := c.Set(ctx, "key1", []byte("value1"), time.Minute)
		require.NoError(t, err)

		value, err := c.Get(ctx, "key1")
		require.NoError(t, err)
		assert.Equal(t, []byte("value1"), value)
	})

	t.Run("Clear with prefix", func(t *testing.T) {
		c := cache.NewMemoryCache(cache.WithKeyPrefix("test"))
		defer c.Close()

		err := c.Set(ctx, "key1", []byte("value1"), time.Minute)
		require.NoError(t, err)

		err = c.Clear(ctx)
		require.NoError(t, err)

		_, err = c.Get(ctx, "key1")
		assert.ErrorIs(t, err, cache.ErrKeyNotFound)
	})
}
