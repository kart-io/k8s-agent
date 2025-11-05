package initializers

import (
	"context"

	"github.com/kart-io/k8s-agent/cmd/monitor/app/options"
	"github.com/kart-io/k8s-agent/internal/monitor/storage"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
)

// RedisInitializer handles Redis initialization.
// Note: Monitor uses a custom storage layer (storage.RedisStorage) for metrics caching,
// so we keep this custom implementation rather than using pkg/initializers.
type RedisInitializer struct {
	cfg     *options.ServerOptions
	logger  core.Logger
	storage *storage.RedisStorage
}

// NewRedisInitializer creates a new Redis initializer.
func NewRedisInitializer(cfg *options.ServerOptions, logger core.Logger) *RedisInitializer {
	return &RedisInitializer{
		cfg:    cfg,
		logger: logger,
	}
}

// Name returns the initializer name.
func (r *RedisInitializer) Name() string {
	return "monitor-redis"
}

// Priority returns initialization priority (lower runs first).
func (r *RedisInitializer) Priority() int {
	return bootstrap.PriorityRedis
}

// Initialize initializes the Redis connection.
func (r *RedisInitializer) Initialize(ctx context.Context) error {
	r.logger.Infow("Initializing Redis connection")

	storage, err := storage.NewRedisStorage(&storage.RedisConfig{
		Host:     r.cfg.Redis.Addr,
		Port:     0, // Port is included in Addr
		Password: r.cfg.Redis.Password,
		DB:       r.cfg.Redis.DB,
		PoolSize: r.cfg.Redis.PoolSize,
	}, r.logger)
	if err != nil {
		return err
	}

	r.storage = storage
	r.logger.Infow("Redis initialized successfully")
	return nil
}

// Run does nothing - Redis is passive.
func (r *RedisInitializer) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Close closes the Redis connection.
func (r *RedisInitializer) Close(ctx context.Context) error {
	if r.storage != nil {
		r.logger.Infow("Closing Redis connection")
		return r.storage.Close()
	}
	return nil
}

// Storage returns the initialized storage.
func (r *RedisInitializer) Storage() *storage.RedisStorage {
	return r.storage
}
