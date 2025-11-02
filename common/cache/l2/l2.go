// Package l2 provides L2 cache implementation with local and remote levels.
package l2

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/ristretto"

	"github.com/kart-io/k8s-agent/common/cache"
	"github.com/kart-io/k8s-agent/common/serializers"
)

// L2Cache implements a two-level cache with local (Ristretto) and remote (Redis) backends.
// It provides:
// - Fast local cache for frequently accessed data (~5ms latency)
// - Distributed remote cache for shared data (~50ms latency)
// - Write-through semantics for consistency
// - Automatic synchronization between levels
// - Generic type support with JSON serialization
type L2Cache[T any] struct {
	opts    *cache.L2Options
	local   *ristretto.Cache
	remote  cache.Cache
	metrics *l2Metrics // Custom metrics tracking
}

// l2Metrics tracks L2 cache-specific metrics.
type l2Metrics struct {
	localHits    atomic.Uint64
	localMisses  atomic.Uint64
	remoteHits   atomic.Uint64
	remoteMisses atomic.Uint64
}

// NewL2Cache creates a new two-level cache with the given remote backend and options.
//
// Example usage:
//
//	redisCache := NewRedisCache(redisClient, commoncache.Options{KeyPrefix: "agent:"})
//	l2Cache, err := cache.NewL2Cache[Agent](redisCache,
//	    cache.WithLocalSize(10000),
//	    cache.WithLocalTTL(5*time.Minute),
//	    cache.WithWriteThrough(true),
//	)
func NewL2Cache[T any](remote cache.Cache, opts ...cache.L2Option) (*L2Cache[T], error) {
	options := cache.DefaultL2Options()
	for _, opt := range opts {
		opt(options)
	}

	// Use default JSON serializer if not specified
	if options.Serializer == nil {
		options.Serializer = serializers.NewJSONSerializer()
	}

	// Initialize Ristretto local cache
	localCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: options.LocalCounters,
		MaxCost:     options.LocalSize,
		BufferItems: 64, // Recommended by Ristretto docs
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create local cache: %w", err)
	}

	cache := &L2Cache[T]{
		opts:   options,
		local:  localCache,
		remote: remote,
	}

	// Initialize metrics if enabled
	if options.EnableMetrics {
		cache.metrics = &l2Metrics{}
	}

	return cache, nil
}

// Get retrieves a value from the cache.
// It first checks the local cache for fast access, then falls back to remote cache.
// If found in remote, the value is automatically populated into local cache.
func (c *L2Cache[T]) Get(ctx context.Context, key string) (T, error) {
	var zero T

	// Try local cache first (fast path)
	if value, found := c.local.Get(key); found {
		if c.opts.EnableMetrics && c.metrics != nil {
			c.metrics.localHits.Add(1)
		}
		return value.(T), nil
	}

	// Fall back to remote cache
	data, err := c.remote.Get(ctx, key)
	if err != nil {
		if c.opts.EnableMetrics && c.metrics != nil {
			c.metrics.remoteMisses.Add(1)
		}
		return zero, err
	}

	// Deserialize using configured serializer
	var value T
	if err := c.opts.Serializer.Unmarshal(data, &value); err != nil {
		return zero, fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	// Populate local cache for future fast access
	c.local.SetWithTTL(key, value, c.opts.LocalCost, c.opts.LocalTTL)

	if c.opts.EnableMetrics && c.metrics != nil {
		c.metrics.remoteHits.Add(1)
	}

	return value, nil
}

// Set stores a value in the cache.
// With write-through enabled (default), writes go to both local and remote immediately.
// With write-through disabled, only local cache is updated (useful for read-heavy workloads).
func (c *L2Cache[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) error {
	// Serialize using configured serializer
	data, err := c.opts.Serializer.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Always write to remote first for consistency
	if err := c.remote.Set(ctx, key, data, ttl); err != nil {
		return fmt.Errorf("failed to set in remote cache: %w", err)
	}

	// Update local cache
	if !c.opts.InvalidateOnWrite {
		// Populate local cache with the new value
		localTTL := c.opts.LocalTTL
		if ttl > 0 && ttl < localTTL {
			localTTL = ttl // Use shorter TTL if remote TTL is less
		}
		c.local.SetWithTTL(key, value, c.opts.LocalCost, localTTL)
	} else {
		// Invalidate local cache to force reload on next Get
		c.local.Del(key)
	}

	return nil
}

// Delete removes a value from both cache levels.
func (c *L2Cache[T]) Delete(ctx context.Context, key string) error {
	// Delete from local cache
	c.local.Del(key)

	// Delete from remote cache
	if err := c.remote.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete from remote cache: %w", err)
	}

	return nil
}

// Exists checks if a key exists in either cache level.
// It checks local first, then remote.
func (c *L2Cache[T]) Exists(ctx context.Context, key string) (bool, error) {
	// Check local cache first
	if _, found := c.local.Get(key); found {
		return true, nil
	}

	// Check remote cache
	return c.remote.Exists(ctx, key)
}

// Clear removes all entries from both cache levels.
func (c *L2Cache[T]) Clear(ctx context.Context) error {
	// Clear local cache
	c.local.Clear()

	// Clear remote cache
	if err := c.remote.Clear(ctx); err != nil {
		return fmt.Errorf("failed to clear remote cache: %w", err)
	}

	return nil
}

// GetMulti retrieves multiple values from the cache.
// It first checks local cache for all keys, then fetches missing keys from remote.
// Missing keys found in remote are automatically populated into local cache.
func (c *L2Cache[T]) GetMulti(ctx context.Context, keys []string) (map[string]T, error) {
	result := make(map[string]T)
	missingKeys := make([]string, 0, len(keys))

	// Check local cache for all keys
	for _, key := range keys {
		if value, found := c.local.Get(key); found {
			result[key] = value.(T)
		} else {
			missingKeys = append(missingKeys, key)
		}
	}

	// If all keys found in local cache, return early
	if len(missingKeys) == 0 {
		return result, nil
	}

	// Fetch missing keys from remote cache
	remoteData, err := c.remote.MGet(ctx, missingKeys...)
	if err != nil {
		return result, fmt.Errorf("failed to get from remote cache: %w", err)
	}

	// Deserialize and populate local cache
	for key, data := range remoteData {
		var value T
		if err := c.opts.Serializer.Unmarshal(data, &value); err != nil {
			// Skip malformed entries but continue processing others
			continue
		}
		c.local.SetWithTTL(key, value, c.opts.LocalCost, c.opts.LocalTTL)
		result[key] = value
	}

	return result, nil
}

// SetMulti stores multiple values in the cache.
// With write-through enabled, writes go to both local and remote.
func (c *L2Cache[T]) SetMulti(ctx context.Context, items map[string]T, ttl time.Duration) error {
	// Serialize all values for remote storage
	remoteItems := make(map[string][]byte, len(items))
	for key, value := range items {
		data, err := c.opts.Serializer.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}
		remoteItems[key] = data
	}

	// Write to remote first for consistency
	if err := c.remote.MSet(ctx, remoteItems, ttl); err != nil {
		return fmt.Errorf("failed to set in remote cache: %w", err)
	}

	// Update local cache
	if !c.opts.InvalidateOnWrite {
		localTTL := c.opts.LocalTTL
		if ttl > 0 && ttl < localTTL {
			localTTL = ttl
		}
		for key, value := range items {
			c.local.SetWithTTL(key, value, c.opts.LocalCost, localTTL)
		}
	} else {
		// Invalidate local cache entries
		for key := range items {
			c.local.Del(key)
		}
	}

	return nil
}

// Close closes the cache and releases resources.
func (c *L2Cache[T]) Close() error {
	// Close local cache
	c.local.Close()

	// Close remote cache
	if err := c.remote.Close(); err != nil {
		return fmt.Errorf("failed to close remote cache: %w", err)
	}

	return nil
}

// Stats returns cache statistics if metrics are enabled.
func (c *L2Cache[T]) Stats() *CacheStats {
	if !c.opts.EnableMetrics || c.metrics == nil {
		return nil
	}

	ristrettoMetrics := c.local.Metrics

	localHits := c.metrics.localHits.Load()
	remoteHits := c.metrics.remoteHits.Load()
	remoteMisses := c.metrics.remoteMisses.Load()

	totalHits := localHits + remoteHits
	totalRequests := totalHits + remoteMisses

	var hitRatio float64
	if totalRequests > 0 {
		hitRatio = float64(totalHits) / float64(totalRequests)
	}

	var localHitRatio float64
	if totalRequests > 0 {
		localHitRatio = float64(localHits) / float64(totalRequests)
	}

	return &CacheStats{
		LocalHits:     localHits,
		LocalMisses:   remoteHits + remoteMisses, // Misses from local cache perspective
		RemoteHits:    remoteHits,
		RemoteMisses:  remoteMisses,
		LocalKeys:     ristrettoMetrics.KeysAdded() - ristrettoMetrics.KeysEvicted(),
		LocalCost:     int64(ristrettoMetrics.CostAdded() - ristrettoMetrics.CostEvicted()),
		Ratio:         ristrettoMetrics.Ratio(), // Ristretto's own hit ratio
		L2HitRatio:    hitRatio,                 // Overall L2 hit ratio
		LocalHitRatio: localHitRatio,            // Local cache hit ratio
	}
}

// CacheStats represents cache performance metrics.
type CacheStats struct {
	LocalHits     uint64  // Number of local cache hits
	LocalMisses   uint64  // Number of local cache misses (went to remote)
	RemoteHits    uint64  // Number of remote cache hits
	RemoteMisses  uint64  // Number of remote cache misses (total miss)
	LocalKeys     uint64  // Number of keys in local cache
	LocalCost     int64   // Total cost of local cache entries
	Ratio         float64 // Ristretto hit ratio
	L2HitRatio    float64 // Overall L2 cache hit ratio
	LocalHitRatio float64 // Percentage of requests served from local cache
}
