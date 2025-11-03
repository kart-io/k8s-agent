// Package cache provides cache configuration options.
package cache

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

// Type defines the cache backend type.
type Type string

const (
	// TypeMemory represents in-memory cache.
	TypeMemory Type = "memory"
	// TypeRedis represents Redis cache.
	TypeRedis Type = "redis"
	// TypeL2 represents two-level cache (memory + redis).
	TypeL2 Type = "l2"
)

// Options defines cache configuration.
type Options struct {
	// Backend type
	Type Type `json:"type" mapstructure:"type"`

	// Common options
	KeyPrefix         string        `json:"keyPrefix" mapstructure:"key_prefix"`
	DefaultExpiration time.Duration `json:"defaultExpiration" mapstructure:"default_expiration"`

	// Memory cache options
	MemoryMaxSize      int           `json:"memoryMaxSize" mapstructure:"memory_max_size"`
	MemoryCleanupInterval time.Duration `json:"memoryCleanupInterval" mapstructure:"memory_cleanup_interval"`

	// Redis cache options
	RedisAddr     string `json:"redisAddr" mapstructure:"redis_addr"`
	RedisPassword string `json:"redisPassword" mapstructure:"redis_password"`
	RedisDB       int    `json:"redisDB" mapstructure:"redis_db"`
	RedisPoolSize int    `json:"redisPoolSize" mapstructure:"redis_pool_size"`

	// L2 cache options
	L2LocalTTL  time.Duration `json:"l2LocalTTL" mapstructure:"l2_local_ttl"`
	L2RemoteTTL time.Duration `json:"l2RemoteTTL" mapstructure:"l2_remote_ttl"`
}

// DefaultOptions returns default cache options.
func DefaultOptions() *Options {
	return &Options{
		Type:                  TypeMemory,
		KeyPrefix:             "",
		DefaultExpiration:     time.Hour,
		MemoryMaxSize:         1000,
		MemoryCleanupInterval: 10 * time.Minute,
		RedisAddr:             "localhost:6379",
		RedisPassword:         "",
		RedisDB:               0,
		RedisPoolSize:         10,
		L2LocalTTL:            5 * time.Minute,
		L2RemoteTTL:           time.Hour,
	}
}

// Validate validates the cache options.
func (o *Options) Validate() error {
	switch o.Type {
	case TypeMemory, TypeRedis, TypeL2:
		// Valid types
	default:
		return fmt.Errorf("invalid cache type: %s", o.Type)
	}

	if o.Type == TypeRedis || o.Type == TypeL2 {
		if o.RedisAddr == "" {
			return fmt.Errorf("redis address is required for %s cache", o.Type)
		}
		if o.RedisDB < 0 || o.RedisDB > 15 {
			return fmt.Errorf("redis DB must be between 0 and 15")
		}
		if o.RedisPoolSize <= 0 {
			return fmt.Errorf("redis pool size must be > 0")
		}
	}

	if o.Type == TypeMemory || o.Type == TypeL2 {
		if o.MemoryMaxSize <= 0 {
			return fmt.Errorf("memory max size must be > 0")
		}
	}

	return nil
}

// Complete completes the cache options with defaults.
func (o *Options) Complete() error {
	if o.Type == "" {
		o.Type = TypeMemory
	}
	if o.DefaultExpiration == 0 {
		o.DefaultExpiration = time.Hour
	}
	if o.MemoryMaxSize == 0 {
		o.MemoryMaxSize = 1000
	}
	if o.MemoryCleanupInterval == 0 {
		o.MemoryCleanupInterval = 10 * time.Minute
	}
	if o.RedisAddr == "" {
		o.RedisAddr = "localhost:6379"
	}
	if o.RedisPoolSize == 0 {
		o.RedisPoolSize = 10
	}
	if o.L2LocalTTL == 0 {
		o.L2LocalTTL = 5 * time.Minute
	}
	if o.L2RemoteTTL == 0 {
		o.L2RemoteTTL = time.Hour
	}
	return nil
}

// AddFlags adds cache configuration flags.
func (o *Options) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	prefix := ""
	if len(prefixes) > 0 {
		prefix = prefixes[0] + "-"
	}

	fs.StringVar((*string)(&o.Type), prefix+"cache-type", string(o.Type), "Cache backend type (memory, redis, l2)")
	fs.StringVar(&o.KeyPrefix, prefix+"cache-key-prefix", o.KeyPrefix, "Cache key prefix")
	fs.DurationVar(&o.DefaultExpiration, prefix+"cache-default-expiration", o.DefaultExpiration, "Default cache expiration")

	fs.IntVar(&o.MemoryMaxSize, prefix+"cache-memory-max-size", o.MemoryMaxSize, "Memory cache max size")
	fs.DurationVar(&o.MemoryCleanupInterval, prefix+"cache-memory-cleanup-interval", o.MemoryCleanupInterval, "Memory cache cleanup interval")

	fs.StringVar(&o.RedisAddr, prefix+"cache-redis-addr", o.RedisAddr, "Redis address")
	fs.StringVar(&o.RedisPassword, prefix+"cache-redis-password", o.RedisPassword, "Redis password")
	fs.IntVar(&o.RedisDB, prefix+"cache-redis-db", o.RedisDB, "Redis database")
	fs.IntVar(&o.RedisPoolSize, prefix+"cache-redis-pool-size", o.RedisPoolSize, "Redis pool size")

	fs.DurationVar(&o.L2LocalTTL, prefix+"cache-l2-local-ttl", o.L2LocalTTL, "L2 cache local TTL")
	fs.DurationVar(&o.L2RemoteTTL, prefix+"cache-l2-remote-ttl", o.L2RemoteTTL, "L2 cache remote TTL")
}

// Functional options

// Option is a functional option for cache configuration.
type Option func(*Options)

// WithType sets the cache type.
func WithType(t Type) Option {
	return func(o *Options) {
		o.Type = t
	}
}

// WithKeyPrefix sets the key prefix.
func WithKeyPrefix(prefix string) Option {
	return func(o *Options) {
		o.KeyPrefix = prefix
	}
}

// WithDefaultExpiration sets the default expiration.
func WithDefaultExpiration(d time.Duration) Option {
	return func(o *Options) {
		o.DefaultExpiration = d
	}
}

// WithRedisAddr sets the Redis address.
func WithRedisAddr(addr string) Option {
	return func(o *Options) {
		o.RedisAddr = addr
	}
}

// WithRedisPassword sets the Redis password.
func WithRedisPassword(password string) Option {
	return func(o *Options) {
		o.RedisPassword = password
	}
}