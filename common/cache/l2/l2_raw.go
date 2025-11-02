package l2

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/kart-io/k8s-agent/common/cache"
)

// L2CacheRaw is a non-generic wrapper for L2Cache that works with []byte.
// This allows L2Cache to be used through the cache.Cache interface in the factory pattern.
type L2CacheRaw struct {
	opts   *cache.L2Options
	local  *ristretto.Cache
	remote cache.Cache
}

// NewL2CacheRaw creates a new L2 cache that implements the cache.Cache interface.
// It stores and retrieves []byte values, handling serialization internally.
func NewL2CacheRaw(remote cache.Cache, opts ...cache.L2Option) (cache.Cache, error) {
	options := cache.DefaultL2Options()
	for _, opt := range opts {
		opt(options)
	}

	// Initialize Ristretto local cache
	localCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: options.LocalCounters,
		MaxCost:     options.LocalSize,
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create local cache: %w", err)
	}

	cache := &L2CacheRaw{
		opts:   options,
		local:  localCache,
		remote: remote,
	}

	return cache, nil
}

// Get retrieves a value from the cache.
func (c *L2CacheRaw) Get(ctx context.Context, key string) ([]byte, error) {
	// Try local cache first (fast path)
	if value, found := c.local.Get(key); found {
		return value.([]byte), nil
	}

	// Fall back to remote cache
	data, err := c.remote.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// Populate local cache for future fast access
	c.local.SetWithTTL(key, data, c.opts.LocalCost, c.opts.LocalTTL)

	return data, nil
}

// Set stores a value in the cache.
func (c *L2CacheRaw) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	// Always write to remote first for consistency
	if err := c.remote.Set(ctx, key, value, expiration); err != nil {
		return fmt.Errorf("failed to set in remote cache: %w", err)
	}

	// Update local cache based on strategy
	if !c.opts.InvalidateOnWrite {
		localTTL := c.opts.LocalTTL
		if expiration > 0 && expiration < localTTL {
			localTTL = expiration
		}
		c.local.SetWithTTL(key, value, c.opts.LocalCost, localTTL)
	} else {
		// Invalidate local cache entry
		c.local.Del(key)
	}

	return nil
}

// Delete removes a value from the cache.
func (c *L2CacheRaw) Delete(ctx context.Context, key string) error {
	// Delete from local cache first
	c.local.Del(key)

	// Delete from remote cache
	return c.remote.Delete(ctx, key)
}

// Exists checks if a key exists in the cache.
func (c *L2CacheRaw) Exists(ctx context.Context, key string) (bool, error) {
	// Check local cache first
	if _, found := c.local.Get(key); found {
		return true, nil
	}

	// Fall back to remote cache
	return c.remote.Exists(ctx, key)
}

// Expire sets expiration time for a key.
func (c *L2CacheRaw) Expire(ctx context.Context, key string, expiration time.Duration) error {
	// Expire in remote cache only
	// Local cache entries have their own TTL
	return c.remote.Expire(ctx, key, expiration)
}

// GetWithTTL retrieves value and remaining TTL.
func (c *L2CacheRaw) GetWithTTL(ctx context.Context, key string) ([]byte, time.Duration, error) {
	// For local cache, we can't get TTL easily, so delegate to remote
	if value, found := c.local.Get(key); found {
		// Return the value with local TTL
		return value.([]byte), c.opts.LocalTTL, nil
	}

	// Get from remote with TTL
	return c.remote.GetWithTTL(ctx, key)
}

// MGet retrieves multiple values by keys.
func (c *L2CacheRaw) MGet(ctx context.Context, keys ...string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	missingKeys := make([]string, 0, len(keys))

	// Try to get from local cache first
	for _, key := range keys {
		if value, found := c.local.Get(key); found {
			result[key] = value.([]byte)
		} else {
			missingKeys = append(missingKeys, key)
		}
	}

	// Fetch missing keys from remote
	if len(missingKeys) > 0 {
		remoteValues, err := c.remote.MGet(ctx, missingKeys...)
		if err != nil {
			return nil, err
		}

		// Populate local cache and result
		for key, value := range remoteValues {
			c.local.SetWithTTL(key, value, c.opts.LocalCost, c.opts.LocalTTL)
			result[key] = value
		}
	}

	return result, nil
}

// MSet stores multiple key-value pairs.
func (c *L2CacheRaw) MSet(ctx context.Context, items map[string][]byte, expiration time.Duration) error {
	// Write to remote first
	if err := c.remote.MSet(ctx, items, expiration); err != nil {
		return fmt.Errorf("failed to mset in remote cache: %w", err)
	}

	// Update local cache based on strategy
	if !c.opts.InvalidateOnWrite {
		localTTL := c.opts.LocalTTL
		if expiration > 0 && expiration < localTTL {
			localTTL = expiration
		}
		for key, value := range items {
			c.local.SetWithTTL(key, value, c.opts.LocalCost, localTTL)
		}
	} else {
		// Invalidate all keys in local cache
		for key := range items {
			c.local.Del(key)
		}
	}

	return nil
}

// Increment atomically increments a counter.
func (c *L2CacheRaw) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	// Invalidate local cache since counter is being modified
	c.local.Del(key)

	// Delegate to remote cache for atomic operation
	return c.remote.Increment(ctx, key, delta)
}

// Decrement atomically decrements a counter.
func (c *L2CacheRaw) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	// Invalidate local cache since counter is being modified
	c.local.Del(key)

	// Delegate to remote cache for atomic operation
	return c.remote.Decrement(ctx, key, delta)
}

// Clear removes all keys from cache.
func (c *L2CacheRaw) Clear(ctx context.Context) error {
	// Clear local cache
	c.local.Clear()

	// Clear remote cache
	return c.remote.Clear(ctx)
}

// Close closes the cache connection.
func (c *L2CacheRaw) Close() error {
	// Close local cache
	c.local.Close()

	// Close remote cache
	return c.remote.Close()
}

// Stats returns cache statistics.
func (c *L2CacheRaw) Stats() map[string]interface{} {
	if !c.opts.EnableMetrics {
		return nil
	}

	metrics := c.local.Metrics

	return map[string]interface{}{
		"hits":          metrics.Hits(),
		"misses":        metrics.Misses(),
		"ratio":         metrics.Ratio(),
		"keys_added":    metrics.KeysAdded(),
		"keys_updated":  metrics.KeysUpdated(),
		"keys_evicted":  metrics.KeysEvicted(),
		"cost_added":    metrics.CostAdded(),
		"cost_evicted":  metrics.CostEvicted(),
		"sets_dropped":  metrics.SetsDropped(),
		"sets_rejected": metrics.SetsRejected(),
		"gets_kept":     metrics.GetsKept(),
		"gets_dropped":  metrics.GetsDropped(),
	}
}

// Helper function to marshal/unmarshal if needed in the future
func marshal(v interface{}) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}
	return json.Marshal(v)
}

func unmarshal(data []byte, v interface{}) error {
	if ptr, ok := v.(*[]byte); ok {
		*ptr = data
		return nil
	}
	return json.Unmarshal(data, v)
}
