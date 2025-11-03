// Package cache provides L2 cache implementation with local and remote levels.
package cache

import (
	"time"

	"github.com/kart-io/k8s-agent/common/serializers"
)

// L2Options configures the L2Cache behavior.
type L2Options struct {
	// Local cache options
	LocalSize     int64         // Maximum number of items in local cache
	LocalTTL      time.Duration // TTL for local cache entries
	LocalCost     int64         // Cost per item (for Ristretto)
	LocalCounters int64         // Number of counters for Ristretto

	// Synchronization options
	SyncInterval      time.Duration // How often to sync local with remote
	WriteThrough      bool          // Whether writes go through to remote immediately
	InvalidateOnWrite bool          // Whether to invalidate local on write

	// Metrics
	EnableMetrics bool // Whether to collect cache metrics

	// Serialization
	Serializer serializers.Serializer // Serializer for encoding/decoding values (defaults to JSON)
}

// DefaultL2Options returns sensible defaults for L2Cache.
func DefaultL2Options() *L2Options {
	return &L2Options{
		LocalSize:         10000,            // 10k items
		LocalTTL:          5 * time.Minute,  // 5 minutes local cache
		LocalCost:         1,                // 1 unit per item
		LocalCounters:     10000,            // Same as size
		SyncInterval:      30 * time.Second, // Sync every 30s
		WriteThrough:      true,             // Always write to remote
		InvalidateOnWrite: true,             // Invalidate local on write
		EnableMetrics:     true,             // Enable metrics by default
		Serializer:        nil,              // Will use JSON by default
	}
}

// L2Option is a functional option for configuring L2Cache.
type L2Option func(*L2Options)

// WithLocalSize sets the maximum number of items in local cache.
func WithLocalSize(size int64) L2Option {
	return func(opts *L2Options) {
		opts.LocalSize = size
	}
}

// WithLocalTTL sets the TTL for local cache entries.
func WithLocalTTL(ttl time.Duration) L2Option {
	return func(opts *L2Options) {
		opts.LocalTTL = ttl
	}
}

// WithLocalCost sets the cost per item for Ristretto.
func WithLocalCost(cost int64) L2Option {
	return func(opts *L2Options) {
		opts.LocalCost = cost
	}
}

// WithLocalCounters sets the number of counters for Ristretto.
func WithLocalCounters(counters int64) L2Option {
	return func(opts *L2Options) {
		opts.LocalCounters = counters
	}
}

// WithSyncInterval sets how often to sync local with remote.
func WithSyncInterval(interval time.Duration) L2Option {
	return func(opts *L2Options) {
		opts.SyncInterval = interval
	}
}

// WithWriteThrough enables/disables write-through to remote.
func WithWriteThrough(enabled bool) L2Option {
	return func(opts *L2Options) {
		opts.WriteThrough = enabled
	}
}

// WithInvalidateOnWrite enables/disables local invalidation on write.
func WithInvalidateOnWrite(enabled bool) L2Option {
	return func(opts *L2Options) {
		opts.InvalidateOnWrite = enabled
	}
}

// WithMetrics enables/disables cache metrics collection.
func WithMetrics(enabled bool) L2Option {
	return func(opts *L2Options) {
		opts.EnableMetrics = enabled
	}
}

// WithSerializer sets the serializer for encoding/decoding values.
// If not set, defaults to JSON serializer.
// Example:
//
//	import "github.com/kart-io/k8s-agent/common/serializers"
//	WithSerializer(serializers.NewMsgpackSerializer())
func WithSerializer(serializer serializers.Serializer) L2Option {
	return func(opts *L2Options) {
		opts.Serializer = serializer
	}
}
