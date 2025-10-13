package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/orchestrator-service/internal/config"
)

// RedisStore implements Redis caching
type RedisStore struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisStore creates a new Redis store
func NewRedisStore(cfg config.RedisConfig, log *zap.Logger) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	store := &RedisStore{
		client: client,
		logger: log.With(zap.String("component", "redis")),
	}

	store.logger.Info("Redis store initialized")
	return store, nil
}

func (s *RedisStore) Close() error {
	return s.client.Close()
}

func (s *RedisStore) Health(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}
