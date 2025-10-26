package initializers

import (
	"context"

	"github.com/kart-io/k8s-agent/cmd/orchestrator/app/options"
	"github.com/kart-io/k8s-agent/internal/orchestrator/storage"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
)

// RedisInitializer Redis初始化器
type RedisInitializer struct {
	opts   *options.ServerOptions
	logger core.Logger
	store  *storage.RedisStore
}

// NewRedisInitializer 创建Redis初始化器
func NewRedisInitializer(opts *options.ServerOptions, logger core.Logger) *RedisInitializer {
	return &RedisInitializer{
		opts:   opts,
		logger: logger,
	}
}

// Name 返回初始化器名称
func (r *RedisInitializer) Name() string {
	return "redis"
}

// Priority 返回初始化优先级
func (r *RedisInitializer) Priority() int {
	return bootstrap.PriorityCache
}

// Initialize 执行初始化
func (r *RedisInitializer) Initialize(ctx context.Context) error {
	r.logger.Infow("Initializing Redis",
		"addr", r.opts.Redis.Addr,
	)

	store, err := storage.NewRedisStore(r.opts.Redis, r.logger)
	if err != nil {
		return err
	}

	r.store = store
	r.logger.Info("Redis initialized successfully")
	return nil
}

// Close 关闭Redis连接
func (r *RedisInitializer) Close(ctx context.Context) error {
	if r.store != nil {
		r.store.Close()
	}
	return nil
}

// HealthCheck 检查Redis健康状态
func (r *RedisInitializer) HealthCheck(ctx context.Context) error {
	// Redis health is checked via connection status
	return nil
}

// Store 获取存储实例
func (r *RedisInitializer) Store() *storage.RedisStore {
	return r.store
}
