// Package factory provides a unified factory for creating cache instances.
package factory

import (
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/kart-io/k8s-agent/common/cache"
	"github.com/kart-io/k8s-agent/common/cache/l2"
	"github.com/kart-io/k8s-agent/common/cache/memory"
	rediscache "github.com/kart-io/k8s-agent/common/cache/redis"
)

// CacheType represents the type of cache backend.
type CacheType string

const (
	// TypeMemory represents in-memory cache.
	TypeMemory CacheType = "memory"

	// TypeRedis represents Redis cache.
	TypeRedis CacheType = "redis"

	// TypeL2 represents two-level cache (memory + remote).
	TypeL2 CacheType = "l2"
)

// Config is the unified configuration for all cache types.
type Config struct {
	// Type specifies which cache backend to use
	Type CacheType

	// Common options
	Options *cache.Options

	// Redis-specific configuration
	RedisClient goredis.UniversalClient

	// L2 cache-specific configuration
	L2Options *cache.L2Options
	L2Remote  cache.Cache // Remote cache for L2 (usually Redis)
}

// New creates a cache instance based on the configuration.
// This is the main factory function that provides a unified entry point.
//
// Example usage:
//
//	// Memory cache
//	memCache := cache.New(&cache.Config{
//	    Type: cache.TypeMemory,
//	    Options: cache.DefaultOptions(),
//	})
//
//	// Redis cache
//	redisCache := cache.New(&cache.Config{
//	    Type: cache.TypeRedis,
//	    RedisClient: redisClient,
//	    Options: cache.DefaultOptions(),
//	})
//
//	// L2 cache
//	l2Cache := cache.New(&cache.Config{
//	    Type: cache.TypeL2,
//	    L2Remote: redisCache,
//	    L2Options: cache.DefaultL2Options(),
//	})
func New(config *Config) (cache.Cache, error) {
	if config == nil {
		return nil, fmt.Errorf("cache config cannot be nil")
	}

	// Set default options if not provided
	if config.Options == nil {
		config.Options = cache.DefaultOptions()
	}

	switch config.Type {
	case TypeMemory:
		return memory.NewMemoryCache(
			cache.WithKeyPrefix(config.Options.KeyPrefix),
			cache.WithDefaultExpiration(config.Options.DefaultExpiration),
			cache.WithMaxRetries(config.Options.MaxRetries),
			cache.WithRetryDelay(config.Options.RetryDelay),
		), nil

	case TypeRedis:
		if config.RedisClient == nil {
			return nil, fmt.Errorf("redis client is required for Redis cache")
		}
		return rediscache.NewRedisCache(
			config.RedisClient,
			cache.WithKeyPrefix(config.Options.KeyPrefix),
			cache.WithDefaultExpiration(config.Options.DefaultExpiration),
			cache.WithMaxRetries(config.Options.MaxRetries),
			cache.WithRetryDelay(config.Options.RetryDelay),
			cache.WithCompression(config.Options.CompressionThreshold),
		), nil

	case TypeL2:
		if config.L2Remote == nil {
			return nil, fmt.Errorf("remote cache is required for L2 cache")
		}
		if config.L2Options == nil {
			config.L2Options = cache.DefaultL2Options()
		}

		// Convert L2Options to functional options
		l2Opts := []cache.L2Option{
			cache.WithLocalSize(config.L2Options.LocalSize),
			cache.WithLocalTTL(config.L2Options.LocalTTL),
			cache.WithLocalCost(config.L2Options.LocalCost),
			cache.WithLocalCounters(config.L2Options.LocalCounters),
			cache.WithWriteThrough(config.L2Options.WriteThrough),
			cache.WithInvalidateOnWrite(config.L2Options.InvalidateOnWrite),
			cache.WithMetrics(config.L2Options.EnableMetrics),
		}

		// Create L2 cache with raw type (using interface{})
		return l2.NewL2CacheRaw(config.L2Remote, l2Opts...)

	default:
		return nil, fmt.Errorf("unsupported cache type: %s", config.Type)
	}
}

// Builder provides a fluent API for building cache instances.
type Builder struct {
	config *Config
}

// NewBuilder creates a new cache builder.
func NewBuilder() *Builder {
	return &Builder{
		config: &Config{
			Options: cache.DefaultOptions(),
		},
	}
}

// Memory configures the builder to create a memory cache.
func (b *Builder) Memory() *Builder {
	b.config.Type = TypeMemory
	return b
}

// Redis configures the builder to create a Redis cache.
func (b *Builder) Redis(client goredis.UniversalClient) *Builder {
	b.config.Type = TypeRedis
	b.config.RedisClient = client
	return b
}

// L2 configures the builder to create an L2 cache.
func (b *Builder) L2(remote cache.Cache) *Builder {
	b.config.Type = TypeL2
	b.config.L2Remote = remote
	if b.config.L2Options == nil {
		b.config.L2Options = cache.DefaultL2Options()
	}
	return b
}

// WithPrefix sets the key prefix.
func (b *Builder) WithPrefix(prefix string) *Builder {
	b.config.Options.KeyPrefix = prefix
	return b
}

// WithExpiration sets the default expiration time.
func (b *Builder) WithExpiration(exp time.Duration) *Builder {
	b.config.Options.DefaultExpiration = exp
	return b
}

// WithL2LocalSize sets the L2 local cache size.
func (b *Builder) WithL2LocalSize(size int64) *Builder {
	if b.config.L2Options == nil {
		b.config.L2Options = cache.DefaultL2Options()
	}
	b.config.L2Options.LocalSize = size
	return b
}

// WithL2LocalTTL sets the L2 local cache TTL.
func (b *Builder) WithL2LocalTTL(ttl time.Duration) *Builder {
	if b.config.L2Options == nil {
		b.config.L2Options = cache.DefaultL2Options()
	}
	b.config.L2Options.LocalTTL = ttl
	return b
}

// Build creates the cache instance.
func (b *Builder) Build() (cache.Cache, error) {
	return New(b.config)
}

// MustBuild creates the cache instance or panics on error.
func (b *Builder) MustBuild() cache.Cache {
	cache, err := b.Build()
	if err != nil {
		panic(err)
	}
	return cache
}
