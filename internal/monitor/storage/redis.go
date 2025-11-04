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
	client *redis.Client
	log    core.Logger
}

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
		client: client,
		log:    logger,
	}, nil
}

func (r *RedisStorage) Close() error {
	return r.client.Close()
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
