package initializers

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
)

// RedisInitializer 通用 Redis 初始化器
//
// 负责初始化Redis连接，提供以下功能：
// - Redis 连接初始化
// - 连接池配置
// - 健康检查
// - 优雅关闭.
type RedisInitializer struct {
	opts   *options.RedisOptions
	logger core.Logger
	client *redis.Client
}

// NewRedisInitializer 创建 Redis 初始化器
//
// 参数：
//   - opts: Redis 配置选项
//   - logger: 日志记录器
//
// 返回：
//   - *RedisInitializer: Redis 初始化器实例
func NewRedisInitializer(
	opts *options.RedisOptions,
	logger core.Logger,
) *RedisInitializer {
	return &RedisInitializer{
		opts:   opts,
		logger: logger,
	}
}

// Name 返回初始化器名称.
func (r *RedisInitializer) Name() string {
	return "redis"
}

// Priority 返回初始化优先级
//
// Redis 初始化器应该在数据库之后，业务逻辑之前初始化.
func (r *RedisInitializer) Priority() int {
	return bootstrap.PriorityCache
}

// Initialize 执行初始化
//
// 初始化流程：
//  1. 创建 Redis 客户端
//  2. 测试 Redis 连接（Ping）
//
// 参数：
//   - ctx: 上下文对象
//
// 返回：
//   - error: 初始化错误
func (r *RedisInitializer) Initialize(ctx context.Context) error {
	r.logger.Infow("Initializing Redis connection",
		"addr", r.opts.Addr,
		"db", r.opts.DB,
	)

	// 直接使用 db 包创建客户端
	client, err := r.opts.ConnectRedis(r.logger)
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	r.client = client

	return nil
}

// HealthCheck 检查 Redis 健康状态
//
// 通过执行 Ping 操作检查 Redis 连接是否正常。
//
// 参数：
//   - ctx: 上下文对象（用于超时控制）
//
// 返回：
//   - error: 健康检查错误
func (r *RedisInitializer) HealthCheck(ctx context.Context) error {
	if r.client == nil {
		return fmt.Errorf("redis not initialized")
	}
	return r.opts.Health(ctx, r.logger)
}

// Client 获取 Redis 客户端
//
// 返回原生 Redis 客户端实例，供业务逻辑使用。
//
// 返回：
//   - *redis.Client: Redis 客户端实例
func (r *RedisInitializer) Client() *redis.Client {
	if r.client == nil {
		return nil
	}
	return r.client
}

// RedisClient 获取完整的 Redis 客户端
//
// 返回完整的 RedisClient 实例，包含了额外的辅助方法。
//
// 返回：
//   - *db.RedisClient: Redis 客户端实例
func (r *RedisInitializer) RedisClient() *redis.Client {
	return r.client
}
