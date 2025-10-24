package storage

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	commondb "github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

// RedisStore implements Redis caching
type RedisStore struct {
	client      *redis.Client
	logger      core.Logger
	redisClient *commondb.RedisClient
}

// NewRedisStore creates a new Redis store using common/db
func NewRedisStore(opts *options.RedisOptions, log core.Logger) (*RedisStore, error) {
	// 使用 common/db helper 函数创建 Redis 客户端
	redisClient, err := commondb.NewRedisFromOptions(log, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	store := &RedisStore{
		client:      redisClient.Client,
		logger:      log,
		redisClient: redisClient,
	}

	store.logger.Info("Redis store initialized successfully")
	return store, nil
}

func (s *RedisStore) Close() error {
	if s.redisClient != nil {
		return s.redisClient.Close()
	}
	return nil
}

func (s *RedisStore) Health(ctx context.Context) error {
	if s.redisClient != nil {
		return s.redisClient.Health(ctx)
	}
	return fmt.Errorf("redis client not initialized")
}
