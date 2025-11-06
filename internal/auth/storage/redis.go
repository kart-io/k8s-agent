package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	commonoptions "github.com/kart-io/k8s-agent/common/options"
)

// RedisClient wraps Redis client.
type RedisClient struct {
	Client *redis.Client
}

// NewRedisClient creates a new Redis client.
func NewRedisClient(cfg *commonoptions.RedisOptions) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to ping redis: %w (close error: %w)", err, closeErr)
		}
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &RedisClient{Client: client}, nil
}

// Close closes the Redis connection.
func (r *RedisClient) Close() error {
	if r.Client != nil {
		return r.Client.Close()
	}
	return nil
}

// Ping checks Redis health.
func (r *RedisClient) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return r.Client.Ping(ctx).Err()
}

// BlacklistToken adds a token to the blacklist with expiration.
func (r *RedisClient) BlacklistToken(ctx context.Context, token string, expiration time.Duration) error {
	key := fmt.Sprintf("blacklist:%s", token)
	return r.Client.Set(ctx, key, "1", expiration).Err()
}

// IsTokenBlacklisted checks if a token is blacklisted.
func (r *RedisClient) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", token)
	val, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil // Token not in blacklist
	}
	if err != nil {
		return false, err
	}
	return val != "", nil
}

// StoreRefreshToken stores a refresh token in Redis with user association.
// Key format: refresh_token:{jti} -> user_id
func (r *RedisClient) StoreRefreshToken(ctx context.Context, jti, userID string, expiration time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s", jti)
	return r.Client.Set(ctx, key, userID, expiration).Err()
}

// GetRefreshTokenOwner retrieves the user ID associated with a refresh token JTI.
func (r *RedisClient) GetRefreshTokenOwner(ctx context.Context, jti string) (string, error) {
	key := fmt.Sprintf("refresh_token:%s", jti)
	userID, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("refresh token not found or expired")
	}
	if err != nil {
		return "", err
	}
	return userID, nil
}

// RevokeRefreshToken removes a refresh token from Redis (token rotation).
func (r *RedisClient) RevokeRefreshToken(ctx context.Context, jti string) error {
	key := fmt.Sprintf("refresh_token:%s", jti)
	return r.Client.Del(ctx, key).Err()
}

// BlacklistRefreshToken adds a refresh token to the blacklist.
// This is used when we want to explicitly revoke a token before it expires naturally.
func (r *RedisClient) BlacklistRefreshToken(ctx context.Context, jti string, expiration time.Duration) error {
	key := fmt.Sprintf("blacklist:refresh:%s", jti)
	return r.Client.Set(ctx, key, "1", expiration).Err()
}

// IsRefreshTokenBlacklisted checks if a refresh token JTI is blacklisted.
func (r *RedisClient) IsRefreshTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("blacklist:refresh:%s", jti)
	val, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val != "", nil
}
