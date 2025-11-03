// Package cache provides a unified interface for caching with multiple backend support.
package cache

import (
	"context"
	"time"

	"github.com/kart-io/k8s-agent/common/serializers"
)


// Cache defines the interface for cache operations.
type Cache interface {
	// Get retrieves a value from cache by key.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in cache with expiration.
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error

	// Delete removes a value from cache.
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in cache.
	Exists(ctx context.Context, key string) (bool, error)

	// Expire sets expiration time for a key.
	Expire(ctx context.Context, key string, expiration time.Duration) error

	// GetWithTTL retrieves value and remaining TTL.
	GetWithTTL(ctx context.Context, key string) ([]byte, time.Duration, error)

	// MGet retrieves multiple values by keys.
	MGet(ctx context.Context, keys ...string) (map[string][]byte, error)

	// MSet stores multiple key-value pairs.
	MSet(ctx context.Context, items map[string][]byte, expiration time.Duration) error

	// Increment atomically increments a counter.
	Increment(ctx context.Context, key string, delta int64) (int64, error)

	// Decrement atomically decrements a counter.
	Decrement(ctx context.Context, key string, delta int64) (int64, error)

	// Clear removes all keys from cache (use with caution).
	Clear(ctx context.Context) error

	// Close closes the cache connection.
	Close() error
}

// Options defines cache configuration options.
type Options struct {
	// Prefix for all cache keys
	KeyPrefix string

	// Default expiration time
	DefaultExpiration time.Duration

	// Maximum number of retries
	MaxRetries int

	// Retry delay
	RetryDelay time.Duration

	// Enable compression for large values
	EnableCompression bool

	// Compression threshold in bytes
	CompressionThreshold int
}

// DefaultOptions returns default cache options.
func DefaultOptions() *Options {
	return &Options{
		KeyPrefix:            "",
		DefaultExpiration:    time.Hour,
		MaxRetries:           3,
		RetryDelay:           time.Millisecond * 100,
		EnableCompression:    false,
		CompressionThreshold: 1024, // 1KB
	}
}

// Option is a functional option for configuring cache.
type Option func(*Options)

// WithKeyPrefix sets the key prefix.
func WithKeyPrefix(prefix string) Option {
	return func(o *Options) {
		o.KeyPrefix = prefix
	}
}

// WithDefaultExpiration sets default expiration.
func WithDefaultExpiration(duration time.Duration) Option {
	return func(o *Options) {
		o.DefaultExpiration = duration
	}
}

// WithMaxRetries sets maximum retry attempts.
func WithMaxRetries(retries int) Option {
	return func(o *Options) {
		o.MaxRetries = retries
	}
}

// WithRetryDelay sets retry delay.
func WithRetryDelay(delay time.Duration) Option {
	return func(o *Options) {
		o.RetryDelay = delay
	}
}

// WithCompression enables compression.
func WithCompression(threshold int) Option {
	return func(o *Options) {
		o.EnableCompression = true
		o.CompressionThreshold = threshold
	}
}

// BuildKey constructs a cache key with optional prefix.
// If prefix is empty, returns the key as-is.
// Otherwise, returns "prefix:key".
func BuildKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + ":" + key
}
