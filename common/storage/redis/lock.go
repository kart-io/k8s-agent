package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Lock represents a distributed lock using Redis.
// It uses SetNX for acquiring locks and Lua script for safe release.
type Lock struct {
	client *Client
	key    string
	value  string
	ttl    time.Duration
}

// Lua script for atomic lock release (check value before delete).
// This prevents accidentally releasing someone else's lock.
var releaseLockScript = redis.NewScript(`
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end
`)

// AcquireLock attempts to acquire a distributed lock.
// It uses Redis SetNX to atomically set the key if it doesn't exist.
// The lock value is a unique UUID to prevent accidental release by other processes.
//
// Parameters:
//   - ctx: Context for timeout control
//   - key: Lock key name (will be prefixed with "lock:")
//   - ttl: Time-to-live for the lock (auto-expires after this duration)
//
// Returns:
//   - *Lock: Lock object if acquired successfully
//   - error: Error if lock is already held or operation fails
//
// Example usage:
//
//	lock, err := client.AcquireLock(ctx, "my-resource", 10*time.Second)
//	if err != nil {
//	    // Lock already held or error
//	    return err
//	}
//	defer lock.Release(ctx)
//	// ... critical section ...
func (c *Client) AcquireLock(ctx context.Context, key string, ttl time.Duration) (*Lock, error) {
	lockKey := fmt.Sprintf("lock:%s", key)
	lockValue := uuid.New().String() // Unique value to prevent accidental release

	// Try to acquire lock using SetNX (SET if Not eXists)
	acquired, err := c.client.SetNX(ctx, lockKey, lockValue, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !acquired {
		return nil, fmt.Errorf("lock already held: %s", key)
	}

	c.logger.Debugw("Lock acquired",
		"key", key,
		"ttl", ttl,
		"value", lockValue,
	)

	return &Lock{
		client: c,
		key:    lockKey,
		value:  lockValue,
		ttl:    ttl,
	}, nil
}

// Release releases the lock.
// It uses a Lua script to atomically check the lock value before deleting,
// ensuring we don't accidentally release someone else's lock.
//
// Returns:
//   - error: Error if lock value doesn't match (already expired/released) or operation fails
func (l *Lock) Release(ctx context.Context) error {
	// Use Lua script to atomically check value and delete
	result, err := releaseLockScript.Run(ctx, l.client.client, []string{l.key}, l.value).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if result == int64(0) {
		return fmt.Errorf("lock value mismatch (already expired or released): %s", l.key)
	}

	l.client.logger.Debugw("Lock released", "key", l.key)
	return nil
}

// Extend extends the lock TTL.
// This is useful for long-running operations that need to keep the lock longer.
//
// Returns:
//   - error: Error if lock value doesn't match or operation fails
func (l *Lock) Extend(ctx context.Context, additionalTTL time.Duration) error {
	// Get current value to verify we still own the lock
	currentValue, err := l.client.client.Get(ctx, l.key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("lock expired: %s", l.key)
		}
		return fmt.Errorf("failed to get lock: %w", err)
	}

	if currentValue != l.value {
		return fmt.Errorf("lock value mismatch (lost ownership): %s", l.key)
	}

	// Extend TTL
	if err := l.client.client.Expire(ctx, l.key, l.ttl+additionalTTL).Err(); err != nil {
		return fmt.Errorf("failed to extend lock: %w", err)
	}

	l.ttl += additionalTTL
	l.client.logger.Debugw("Lock extended",
		"key", l.key,
		"new_ttl", l.ttl,
	)

	return nil
}
