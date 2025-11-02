package initializers

import (
	pkginitializers "github.com/kart-io/k8s-agent/common/initializers"
	"github.com/kart-io/k8s-agent/internal/auth/config"
	"github.com/kart-io/logger/core"
	"github.com/redis/go-redis/v9"
)

// RedisInitializer Redis 初始化器（使用通用适配器）
//
// 现在使用 pkg/initializers.RedisInitializerAdapter 来消除重复代码。
type RedisInitializer struct {
	*pkginitializers.RedisInitializerAdapter
}

// NewRedisInitializer 创建 Redis 初始化器
func NewRedisInitializer(cfg *config.Config, logger core.Logger) *RedisInitializer {
	// 创建通用适配器
	adapter := pkginitializers.NewRedisInitializerAdapter(cfg.Redis, logger)

	return &RedisInitializer{
		RedisInitializerAdapter: adapter,
	}
}

// Client 获取 Redis 客户端（类型安全的便捷方法）
func (r *RedisInitializer) Client() *redis.Client {
	return r.RedisInitializerAdapter.Client()
}
