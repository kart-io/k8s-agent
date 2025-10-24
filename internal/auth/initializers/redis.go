package initializers

import (
	"context"

	"github.com/kart-io/k8s-agent/internal/auth/config"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
	"github.com/redis/go-redis/v9"
)

// RedisInitializer Redis 初始化器（适配器）
//
// 此初始化器是一个适配器，内部使用通用的 pkg/initializers.RedisInitializer，
// 但提供 Client() 方法返回 *redis.Client，以保持与现有代码的兼容性。
type RedisInitializer struct {
	cfg    *config.Config
	logger core.Logger

	// 使用通用初始化器
	redisInit *pkginitializers.RedisInitializer
}

// NewRedisInitializer 创建 Redis 初始化器
func NewRedisInitializer(cfg *config.Config, logger core.Logger) *RedisInitializer {
	// 创建通用 Redis 初始化器
	redisInit := pkginitializers.NewRedisInitializer(
		cfg.Redis,
		logger,
	)

	return &RedisInitializer{
		cfg:       cfg,
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
	return r.redisInit.Initialize(ctx)
}

// Close 关闭 Redis 连接
func (r *RedisInitializer) Close(ctx context.Context) error {
	return r.redisInit.Close(ctx)
}

// HealthCheck 检查 Redis 健康状态
func (r *RedisInitializer) HealthCheck(ctx context.Context) error {
	return r.redisInit.HealthCheck(ctx)
}

// Client 获取 Redis 客户端
func (r *RedisInitializer) Client() *redis.Client {
	return r.redisInit.Client()
}
