package initializers

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/k8s-agent/internal/agent-manager/config"
	"github.com/kart-io/k8s-agent/internal/agent-manager/storage"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
)

// RedisInitializer Redis 初始化器
type RedisInitializer struct {
	opts   *config.Options
	logger core.Logger
	store  *storage.RedisStore
}

// NewRedisInitializer 创建 Redis 初始化器
func NewRedisInitializer(opts *config.Options, logger core.Logger) *RedisInitializer {
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
	r.logger.Infow("Initializing Redis connection",
		"addr", r.opts.Redis.Addr,
	)

	// 创建 Redis 客户端
	redisClient, err := db.NewRedis(r.logger,
		db.WithAddr(r.opts.Redis.Addr),
		db.WithRedisPassword(r.opts.Redis.Password),
		db.WithRedisDB(r.opts.Redis.DB),
		db.WithRedisPoolSize(r.opts.Redis.PoolSize),
		db.WithRedisMinIdleConns(r.opts.Redis.MinIdleConns),
		db.WithRedisDialTimeout(r.opts.Redis.DialTimeout),
		db.WithRedisReadTimeout(r.opts.Redis.ReadTimeout),
		db.WithRedisWriteTimeout(r.opts.Redis.WriteTimeout),
	)
	if err != nil {
		return fmt.Errorf("failed to create Redis client: %w", err)
	}

	// 创建存储层
	r.store = &storage.RedisStore{
		RedisClient: redisClient,
	}

	r.logger.Infow("Redis initialized successfully")
	return nil
}

// Close 关闭 Redis 连接
func (r *RedisInitializer) Close(ctx context.Context) error {
	if r.store != nil {
		r.logger.Infow("Closing Redis connection")
		return r.store.Close()
	}
	return nil
}

// HealthCheck 检查 Redis 健康状态
func (r *RedisInitializer) HealthCheck(ctx context.Context) error {
	if r.store == nil {
		return fmt.Errorf("redis not initialized")
	}
	// 简单检查:redis store 是否存在
	return nil
}

// Store 获取存储实例
func (r *RedisInitializer) Store() *storage.RedisStore {
	return r.store
}
