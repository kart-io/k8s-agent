package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/kart-io/logger/core"
)

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

type RedisStorage struct {
	client      *redis.Client
	log         core.Logger
	ownedClient bool // Whether this storage owns the client and should close it
}

// NewRedisStorage creates a new Redis storage with its own client.
// This method creates its own Redis connection and should only be used
// when a connection is not already available.
func NewRedisStorage(config *RedisConfig, logger core.Logger) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	logger.Info("Redis storage initialized successfully")
	return &RedisStorage{
		client:      client,
		log:         logger,
		ownedClient: true, // We own this client
	}, nil
}

// NewRedisStorageWithClient creates a new Redis storage using an existing Redis client.
// This is the preferred method when reusing an existing Redis connection.
func NewRedisStorageWithClient(client *redis.Client, logger core.Logger) (*RedisStorage, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client cannot be nil")
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis client ping failed: %w", err)
	}

	logger.Info("Reusing existing Redis connection for monitor storage")
	return &RedisStorage{
		client:      client,
		log:         logger,
		ownedClient: false, // We don't own this client
	}, nil
}

func (r *RedisStorage) Close() error {
	// Only close if we own the client
	if r.ownedClient && r.client != nil {
		return r.client.Close()
	}
	// If we're reusing a connection, don't close it
	return nil
}

func (r *RedisStorage) Client() *redis.Client {
	return r.client
}

// SetMetricsSummary 缓存监控概览.
func (r *RedisStorage) SetMetricsSummary(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// GetMetricsSummary 获取监控概览.
func (r *RedisStorage) GetMetricsSummary(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// IncrementCounter 增加计数器.
func (r *RedisStorage) IncrementCounter(ctx context.Context, key string) error {
	return r.client.Incr(ctx, key).Err()
}

// SetExpire 设置过期时间.
func (r *RedisStorage) SetExpire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}
