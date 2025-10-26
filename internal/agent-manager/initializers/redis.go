package initializers

import (
	"context"

	"github.com/kart-io/k8s-agent/cmd/agent-manager/app/options"
	"github.com/kart-io/k8s-agent/internal/agent-manager/storage"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// RedisInitializer Redis 初始化器（适配器）
//
// 此初始化器是一个适配器，内部使用通用的 pkg/initializers.RedisInitializer，
// 但提供 Store() 方法返回业务特定的 storage.RedisStore，
// 以保持与现有代码的兼容性。
type RedisInitializer struct {
	opts   *options.ServerOptions
	logger core.Logger

	// 使用通用初始化器
	redisInit *pkginitializers.RedisInitializer
	store     *storage.RedisStore
}

// NewRedisInitializer 创建 Redis 初始化器
func NewRedisInitializer(opts *options.ServerOptions, logger core.Logger) *RedisInitializer {
	// 创建通用 Redis 初始化器
	redisInit := pkginitializers.NewRedisInitializer(
		opts.Redis,
		logger,
	)

	return &RedisInitializer{
		opts:      opts,
		logger:    logger,
		redisInit: redisInit,
	}
}

// Name 返回初始化器名称
func (r *RedisInitializer) Name() string {
	return r.redisInit.Name()
}

// Priority 返回初始化优先级
func (r *RedisInitializer) Priority() int {
	return bootstrap.PriorityCache
}

// Initialize 执行初始化
func (r *RedisInitializer) Initialize(ctx context.Context) error {
	// 委托给通用初始化器
	if err := r.redisInit.Initialize(ctx); err != nil {
		return err
	}

	// 创建业务存储层（包装通用客户端）
	r.store = &storage.RedisStore{
		RedisClient: r.redisInit.RedisClient(),
	}

	return nil
}

// Close 关闭 Redis 连接
func (r *RedisInitializer) Close(ctx context.Context) error {
	return r.redisInit.Close(ctx)
}

// HealthCheck 检查 Redis 健康状态
func (r *RedisInitializer) HealthCheck(ctx context.Context) error {
	return r.redisInit.HealthCheck(ctx)
}

// Store 获取存储实例
//
// 返回业务特定的 RedisStore，供其他组件使用。
func (r *RedisInitializer) Store() *storage.RedisStore {
	return r.store
}
