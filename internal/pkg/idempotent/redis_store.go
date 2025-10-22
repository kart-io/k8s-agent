package idempotent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore implements Store interface using Redis.
type RedisStore struct {
	client redis.UniversalClient
	prefix string
}

// NewRedisStore creates a new Redis-based idempotency store.
func NewRedisStore(client redis.UniversalClient, prefix string) *RedisStore {
	if prefix == "" {
		prefix = "idempotent"
	}

	return &RedisStore{
		client: client,
		prefix: prefix,
	}
}

// buildKey constructs the Redis key.
func (r *RedisStore) buildKey(key string) string {
	return fmt.Sprintf("%s:%s", r.prefix, key)
}

// buildLockKey constructs the lock key.
func (r *RedisStore) buildLockKey(key string) string {
	return fmt.Sprintf("%s:lock:%s", r.prefix, key)
}

// Get retrieves a record by key.
func (r *RedisStore) Get(ctx context.Context, key string) (*Record, error) {
	redisKey := r.buildKey(key)

	data, err := r.client.Get(ctx, redisKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("record not found")
		}
		return nil, fmt.Errorf("redis get failed: %w", err)
	}

	var record Record
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record: %w", err)
	}

	// Check if expired
	if !record.ExpiresAt.IsZero() && time.Now().After(record.ExpiresAt) {
		return nil, ErrKeyExpired
	}

	return &record, nil
}

// Set stores a record.
func (r *RedisStore) Set(ctx context.Context, record *Record) error {
	redisKey := r.buildKey(record.Key)

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	// Calculate TTL
	var ttl time.Duration
	if !record.ExpiresAt.IsZero() {
		ttl = time.Until(record.ExpiresAt)
		if ttl < 0 {
			ttl = 0
		}
	}

	if err := r.client.Set(ctx, redisKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}

	return nil
}

// Delete removes a record.
func (r *RedisStore) Delete(ctx context.Context, key string) error {
	redisKey := r.buildKey(key)

	if err := r.client.Del(ctx, redisKey).Err(); err != nil {
		return fmt.Errorf("redis delete failed: %w", err)
	}

	return nil
}

// Acquire attempts to acquire a lock for processing.
func (r *RedisStore) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	lockKey := r.buildLockKey(key)

	// Try to set lock with NX (only if not exists)
	result, err := r.client.SetNX(ctx, lockKey, "locked", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("redis setnx failed: %w", err)
	}

	return result, nil
}

// Release releases a lock.
func (r *RedisStore) Release(ctx context.Context, key string) error {
	lockKey := r.buildLockKey(key)

	if err := r.client.Del(ctx, lockKey).Err(); err != nil {
		return fmt.Errorf("redis delete lock failed: %w", err)
	}

	return nil
}
