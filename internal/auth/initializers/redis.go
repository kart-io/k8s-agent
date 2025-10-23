package initializers

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/k8s-agent/internal/auth/config"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
	"github.com/redis/go-redis/v9"
)

// RedisInitializer Redis 初始化器
type RedisInitializer struct {
	cfg    *config.Config
	logger core.Logger
	client *redis.Client
}

// NewRedisInitializer 创建 Redis 初始化器
func NewRedisInitializer(cfg *config.Config, logger core.Logger) *RedisInitializer {
	return &RedisInitializer{
		cfg:    cfg,
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
		"addr", r.cfg.Redis.Addr,
	)

	// 创建 Redis 客户端
	redisClient, err := db.NewRedis(r.logger,
		db.WithAddr(r.cfg.Redis.Addr),
		db.WithRedisPassword(r.cfg.Redis.Password),
		db.WithRedisDB(r.cfg.Redis.DB),
		db.WithRedisPoolSize(r.cfg.Redis.PoolSize),
	)
	if err != nil {
		return fmt.Errorf("failed to create Redis client: %w", err)
	}

	r.client = redisClient.Client

	r.logger.Infow("Redis initialized successfully")
	return nil
}

// Close 关闭 Redis 连接
func (r *RedisInitializer) Close(ctx context.Context) error {
	if r.client != nil {
		r.logger.Infow("Closing Redis connection")
		return r.client.Close()
	}
	return nil
}

// HealthCheck 检查 Redis 健康状态
func (r *RedisInitializer) HealthCheck(ctx context.Context) error {
	if r.client == nil {
		return fmt.Errorf("redis not initialized")
	}
	return nil
}

// Client 获取 Redis 客户端
func (r *RedisInitializer) Client() *redis.Client {
	return r.client
}
