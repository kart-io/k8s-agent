package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStoreConfig holds configuration for Redis store
type RedisStoreConfig struct {
	// Addr is the Redis server address (host:port)
	Addr string

	// Password for Redis authentication
	Password string

	// DB is the Redis database number
	DB int

	// Prefix is the key prefix for all store keys
	Prefix string

	// TTL is the default time-to-live for keys (0 = no expiration)
	TTL time.Duration

	// PoolSize is the maximum number of connections
	PoolSize int

	// MinIdleConns is the minimum number of idle connections
	MinIdleConns int

	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// DialTimeout is the timeout for establishing connections
	DialTimeout time.Duration

	// ReadTimeout is the timeout for read operations
	ReadTimeout time.Duration

	// WriteTimeout is the timeout for write operations
	WriteTimeout time.Duration
}

// DefaultRedisStoreConfig returns default Redis configuration
func DefaultRedisStoreConfig() *RedisStoreConfig {
	return &RedisStoreConfig{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		Prefix:       "agent:store:",
		TTL:          0, // No expiration
		PoolSize:     10,
		MinIdleConns: 2,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// RedisStore is a Redis-backed implementation of the Store interface.
//
// Features:
//   - Connection pooling for high performance
//   - Namespace-based key organization
//   - Optional TTL for automatic expiration
//   - JSON serialization for complex values
//   - Thread-safe operations
//
// Suitable for:
//   - Production deployments
//   - Distributed systems
//   - High-throughput scenarios
//   - Shared state across services
type RedisStore struct {
	client *redis.Client
	config *RedisStoreConfig
}

// NewRedisStore creates a new Redis-backed store
func NewRedisStore(config *RedisStoreConfig) (*RedisStore, error) {
	if config == nil {
		config = DefaultRedisStoreConfig()
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStore{
		client: client,
		config: config,
	}, nil
}

// NewRedisStoreFromClient creates a RedisStore from an existing client
func NewRedisStoreFromClient(client *redis.Client, config *RedisStoreConfig) *RedisStore {
	if config == nil {
		config = DefaultRedisStoreConfig()
	}

	return &RedisStore{
		client: client,
		config: config,
	}
}

// Put stores a value with the given namespace and key
func (s *RedisStore) Put(ctx context.Context, namespace []string, key string, value interface{}) error {
	redisKey := s.makeKey(namespace, key)

	// Get existing value to preserve created timestamp
	existing, err := s.Get(ctx, namespace, key)
	now := time.Now()

	var created time.Time
	var metadata map[string]interface{}

	if err == nil && existing != nil {
		created = existing.Created
		metadata = existing.Metadata
	} else {
		created = now
		metadata = make(map[string]interface{})
	}

	// Create store value
	storeValue := &StoreValue{
		Value:     value,
		Metadata:  metadata,
		Created:   created,
		Updated:   now,
		Namespace: namespace,
		Key:       key,
	}

	// Serialize to JSON
	data, err := json.Marshal(storeValue)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}

	// Store in Redis
	if s.config.TTL > 0 {
		err = s.client.Set(ctx, redisKey, data, s.config.TTL).Err()
	} else {
		err = s.client.Set(ctx, redisKey, data, 0).Err()
	}

	if err != nil {
		return fmt.Errorf("failed to store value in Redis: %w", err)
	}

	return nil
}

// Get retrieves a value by namespace and key
func (s *RedisStore) Get(ctx context.Context, namespace []string, key string) (*StoreValue, error) {
	redisKey := s.makeKey(namespace, key)

	// Get from Redis
	data, err := s.client.Get(ctx, redisKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("key not found: %s", key)
		}
		return nil, fmt.Errorf("failed to get value from Redis: %w", err)
	}

	// Deserialize
	var storeValue StoreValue
	if err := json.Unmarshal(data, &storeValue); err != nil {
		return nil, fmt.Errorf("failed to deserialize value: %w", err)
	}

	return &storeValue, nil
}

// Delete removes a value by namespace and key
func (s *RedisStore) Delete(ctx context.Context, namespace []string, key string) error {
	redisKey := s.makeKey(namespace, key)

	err := s.client.Del(ctx, redisKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key from Redis: %w", err)
	}

	return nil
}

// Search finds values matching the filter within a namespace
func (s *RedisStore) Search(ctx context.Context, namespace []string, filter map[string]interface{}) ([]*StoreValue, error) {
	// Get all keys in namespace
	pattern := s.makePattern(namespace)
	keys, err := s.scanKeys(ctx, pattern)
	if err != nil {
		return nil, err
	}

	results := make([]*StoreValue, 0)

	// Iterate through keys and filter
	for _, redisKey := range keys {
		data, err := s.client.Get(ctx, redisKey).Bytes()
		if err != nil {
			continue // Skip keys that can't be read
		}

		var storeValue StoreValue
		if err := json.Unmarshal(data, &storeValue); err != nil {
			continue // Skip invalid data
		}

		// Apply filter
		if matchesFilter(&storeValue, filter) {
			results = append(results, &storeValue)
		}
	}

	return results, nil
}

// List returns all keys within a namespace
func (s *RedisStore) List(ctx context.Context, namespace []string) ([]string, error) {
	pattern := s.makePattern(namespace)
	redisKeys, err := s.scanKeys(ctx, pattern)
	if err != nil {
		return nil, err
	}

	// Extract actual keys from Redis keys
	keys := make([]string, 0, len(redisKeys))
	for _, redisKey := range redisKeys {
		// Remove prefix and namespace to get the key
		key := s.extractKey(namespace, redisKey)
		if key != "" {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// Clear removes all values within a namespace
func (s *RedisStore) Clear(ctx context.Context, namespace []string) error {
	pattern := s.makePattern(namespace)
	keys, err := s.scanKeys(ctx, pattern)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	// Delete all keys in batch
	err = s.client.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("failed to clear namespace: %w", err)
	}

	return nil
}

// Close closes the Redis connection
func (s *RedisStore) Close() error {
	return s.client.Close()
}

// makeKey creates a Redis key from namespace and key
func (s *RedisStore) makeKey(namespace []string, key string) string {
	nsKey := s.config.Prefix + namespaceToKey(namespace)
	if !strings.HasSuffix(nsKey, "/") {
		nsKey += "/"
	}
	return nsKey + key
}

// makePattern creates a Redis pattern for scanning namespace keys
func (s *RedisStore) makePattern(namespace []string) string {
	nsKey := s.config.Prefix + namespaceToKey(namespace)
	if !strings.HasSuffix(nsKey, "/") {
		nsKey += "/"
	}
	return nsKey + "*"
}

// extractKey extracts the key part from a full Redis key
func (s *RedisStore) extractKey(namespace []string, redisKey string) string {
	prefix := s.config.Prefix + namespaceToKey(namespace)
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	if !strings.HasPrefix(redisKey, prefix) {
		return ""
	}

	return strings.TrimPrefix(redisKey, prefix)
}

// scanKeys scans Redis keys matching the pattern
func (s *RedisStore) scanKeys(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	var cursor uint64

	for {
		var scanKeys []string
		var err error

		scanKeys, cursor, err = s.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan keys: %w", err)
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

// Ping tests the connection to Redis
func (s *RedisStore) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

// Size returns the approximate number of keys in all namespaces
func (s *RedisStore) Size(ctx context.Context) (int, error) {
	pattern := s.config.Prefix + "*"
	keys, err := s.scanKeys(ctx, pattern)
	if err != nil {
		return 0, err
	}
	return len(keys), nil
}
