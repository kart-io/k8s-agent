package cache

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements Cache interface using Redis.
type RedisCache struct {
	client  redis.UniversalClient
	options *Options
}

// NewRedisCache creates a new Redis cache instance.
func NewRedisCache(client redis.UniversalClient, opts ...Option) Cache {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	return &RedisCache{
		client:  client,
		options: options,
	}
}

// buildKey constructs the full cache key with prefix.
func (r *RedisCache) buildKey(key string) string {
	if r.options.KeyPrefix == "" {
		return key
	}
	return r.options.KeyPrefix + ":" + key
}

// compress compresses data if enabled and above threshold.
func (r *RedisCache) compress(data []byte) ([]byte, error) {
	if !r.options.EnableCompression || len(data) < r.options.CompressionThreshold {
		return data, nil
	}

	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	if _, err := gzWriter.Write(data); err != nil {
		return nil, err
	}
	if err := gzWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// decompress decompresses data if compression is enabled.
func (r *RedisCache) decompress(data []byte) ([]byte, error) {
	if !r.options.EnableCompression {
		return data, nil
	}

	// Check if data is gzip compressed by magic number
	if len(data) < 2 || data[0] != 0x1f || data[1] != 0x8b {
		return data, nil
	}

	gzReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	return io.ReadAll(gzReader)
}

// Get retrieves a value from cache by key.
func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	fullKey := r.buildKey(key)

	result, err := r.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrKeyNotFound
		}
		return nil, fmt.Errorf("redis get failed: %w", err)
	}

	return r.decompress(result)
}

// Set stores a value in cache with expiration.
func (r *RedisCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	fullKey := r.buildKey(key)

	if expiration == 0 {
		expiration = r.options.DefaultExpiration
	}

	compressed, err := r.compress(value)
	if err != nil {
		return fmt.Errorf("compression failed: %w", err)
	}

	err = r.client.Set(ctx, fullKey, compressed, expiration).Err()
	if err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}

	return nil
}

// Delete removes a value from cache.
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := r.buildKey(key)

	err := r.client.Del(ctx, fullKey).Err()
	if err != nil {
		return fmt.Errorf("redis delete failed: %w", err)
	}

	return nil
}

// Exists checks if a key exists in cache.
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := r.buildKey(key)

	count, err := r.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists failed: %w", err)
	}

	return count > 0, nil
}

// Expire sets expiration time for a key.
func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	fullKey := r.buildKey(key)

	ok, err := r.client.Expire(ctx, fullKey, expiration).Result()
	if err != nil {
		return fmt.Errorf("redis expire failed: %w", err)
	}
	if !ok {
		return ErrKeyNotFound
	}

	return nil
}

// GetWithTTL retrieves value and remaining TTL.
func (r *RedisCache) GetWithTTL(ctx context.Context, key string) ([]byte, time.Duration, error) {
	fullKey := r.buildKey(key)

	// Use pipeline for atomic operation
	pipe := r.client.Pipeline()
	getCmd := pipe.Get(ctx, fullKey)
	ttlCmd := pipe.TTL(ctx, fullKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		if err == redis.Nil {
			return nil, 0, ErrKeyNotFound
		}
		return nil, 0, fmt.Errorf("redis pipeline failed: %w", err)
	}

	data, err := getCmd.Bytes()
	if err != nil {
		return nil, 0, fmt.Errorf("get command failed: %w", err)
	}

	ttl, err := ttlCmd.Result()
	if err != nil {
		return nil, 0, fmt.Errorf("ttl command failed: %w", err)
	}

	decompressed, err := r.decompress(data)
	if err != nil {
		return nil, 0, fmt.Errorf("decompression failed: %w", err)
	}

	return decompressed, ttl, nil
}

// MGet retrieves multiple values by keys.
func (r *RedisCache) MGet(ctx context.Context, keys ...string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return make(map[string][]byte), nil
	}

	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.buildKey(key)
	}

	results, err := r.client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("redis mget failed: %w", err)
	}

	output := make(map[string][]byte, len(keys))
	for i, result := range results {
		if result == nil {
			continue
		}

		data, ok := result.(string)
		if !ok {
			continue
		}

		decompressed, err := r.decompress([]byte(data))
		if err != nil {
			continue
		}

		output[keys[i]] = decompressed
	}

	return output, nil
}

// MSet stores multiple key-value pairs.
func (r *RedisCache) MSet(ctx context.Context, items map[string][]byte, expiration time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	if expiration == 0 {
		expiration = r.options.DefaultExpiration
	}

	pipe := r.client.Pipeline()
	for key, value := range items {
		fullKey := r.buildKey(key)
		compressed, err := r.compress(value)
		if err != nil {
			return fmt.Errorf("compression failed for key %s: %w", key, err)
		}
		pipe.Set(ctx, fullKey, compressed, expiration)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis mset pipeline failed: %w", err)
	}

	return nil
}

// Increment atomically increments a counter.
func (r *RedisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	fullKey := r.buildKey(key)

	result, err := r.client.IncrBy(ctx, fullKey, delta).Result()
	if err != nil {
		return 0, fmt.Errorf("redis incr failed: %w", err)
	}

	return result, nil
}

// Decrement atomically decrements a counter.
func (r *RedisCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	fullKey := r.buildKey(key)

	result, err := r.client.DecrBy(ctx, fullKey, delta).Result()
	if err != nil {
		return 0, fmt.Errorf("redis decr failed: %w", err)
	}

	return result, nil
}

// Clear removes all keys from cache (use with caution).
func (r *RedisCache) Clear(ctx context.Context) error {
	// If prefix is set, only clear keys with that prefix
	if r.options.KeyPrefix != "" {
		pattern := r.buildKey("*")
		iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()

		keys := make([]string, 0)
		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}

		if err := iter.Err(); err != nil {
			return fmt.Errorf("redis scan failed: %w", err)
		}

		if len(keys) > 0 {
			err := r.client.Del(ctx, keys...).Err()
			if err != nil {
				return fmt.Errorf("redis delete failed: %w", err)
			}
		}

		return nil
	}

	// Clear all keys (dangerous!)
	err := r.client.FlushDB(ctx).Err()
	if err != nil {
		return fmt.Errorf("redis flushdb failed: %w", err)
	}

	return nil
}

// Close closes the cache connection.
func (r *RedisCache) Close() error {
	return r.client.Close()
}
