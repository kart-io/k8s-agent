package initializers

import (
	"github.com/kart-io/k8s-agent/cmd/agent-manager/app/options"
	"github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/k8s-agent/internal/agent-manager/storage"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// RedisInitializer Redis 初始化器（使用通用适配器）
//
// 现在使用 pkg/initializers.RedisInitializerAdapter 来消除重复代码。
type RedisInitializer struct {
	*pkginitializers.RedisInitializerAdapter
}

// NewRedisInitializer 创建 Redis 初始化器
func NewRedisInitializer(opts *options.ServerOptions, logger core.Logger) *RedisInitializer {
	// 创建通用适配器
	adapter := pkginitializers.NewRedisInitializerAdapter(opts.Redis, logger)

	// 配置 Store 包装函数
	adapter.WithStoreWrapper(func(client *db.RedisClient) interface{} {
		return &storage.RedisStore{
			RedisClient: client,
		}
	})

	return &RedisInitializer{
		RedisInitializerAdapter: adapter,
	}
}

// Store 获取存储实例（类型安全的便捷方法）
//
// 返回业务特定的 RedisStore，供其他组件使用。
func (r *RedisInitializer) Store() *storage.RedisStore {
	if store := r.RedisInitializerAdapter.Store(); store != nil {
		return store.(*storage.RedisStore)
	}
	return nil
}
