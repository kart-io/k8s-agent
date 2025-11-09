package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

// RedisStore implements Redis caching.
type RedisStore struct {
	client      *redis.Client
	logger      core.Logger
}

// NewRedisStore creates a new Redis store using common/db.
func NewRedisStore(opts *options.RedisOptions, log core.Logger) (*RedisStore, error) {
	// 直接使用 db 包创建 Redis 客户端
	redisClient, err := opts.ConnectRedis(log)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	store := &RedisStore{
		client:      redisClient,
		logger:      log,
	}

	store.logger.Info("Redis store initialized successfully")
	return store, nil
}

func (s *RedisStore) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

func (s *RedisStore) Health(ctx context.Context) error {
	if s.client != nil {
		if err := s.client.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("failed to ping Redis: %w", err)
		}
		return nil
	}
	return errors.New("redis client not initialized")
}
