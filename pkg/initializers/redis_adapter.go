package initializers

import (
	"context"

	"github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
	"github.com/redis/go-redis/v9"
)

// RedisInitializerAdapter 通用 Redis 初始化器适配器
//
// 这个适配器包装了通用的 RedisInitializer，
// 并提供了多种返回类型的便捷方法，消除各服务的重复代码。
//
// 使用示例：
//
//	adapter := pkginitializers.NewRedisInitializerAdapter(opts.Redis, logger)
//	// 在服务的 RegisterComponents 中注册
//	bs.Register(adapter)
//
// 优势：
//  1. 消除了每个服务都需要编写的适配器代码
//  2. 提供统一的接口和行为
//  3. 支持多种返回类型（RedisClient, Client, 自定义 Store）
type RedisInitializerAdapter struct {
	redisInit *RedisInitializer

	// 可选：业务层 Store 包装函数
	// 在 Initialize 完成后调用，用于创建服务特定的 Store
	storeWrapper func(*db.RedisClient) interface{}
	store        interface{}
}

// NewRedisInitializerAdapter 创建 Redis 初始化器适配器
func NewRedisInitializerAdapter(
	opts *options.RedisOptions,
	logger core.Logger,
) *RedisInitializerAdapter {
	return &RedisInitializerAdapter{
		redisInit: NewRedisInitializer(opts, logger),
	}
}

// WithStoreWrapper 设置业务层 Store 包装函数
//
// 这个函数在 Redis 初始化完成后被调用，用于创建服务特定的 Store。
//
// 使用示例：
//
//	adapter.WithStoreWrapper(func(client *db.RedisClient) interface{} {
//	    return &storage.RedisStore{RedisClient: client}
//	})
func (a *RedisInitializerAdapter) WithStoreWrapper(wrapper func(*db.RedisClient) interface{}) *RedisInitializerAdapter {
	a.storeWrapper = wrapper
	return a
}

// Name 实现 bootstrap.Initializer 接口
func (a *RedisInitializerAdapter) Name() string {
	return a.redisInit.Name()
}

// Priority 实现 bootstrap.Initializer 接口
func (a *RedisInitializerAdapter) Priority() int {
	return bootstrap.PriorityCache
}

// Initialize 实现 bootstrap.Initializer 接口
func (a *RedisInitializerAdapter) Initialize(ctx context.Context) error {
	// 委托给通用初始化器
	if err := a.redisInit.Initialize(ctx); err != nil {
		return err
	}

	// 如果设置了 Store 包装函数，创建业务层 Store
	if a.storeWrapper != nil {
		a.store = a.storeWrapper(a.redisInit.RedisClient())
	}

	return nil
}

// Close 实现 bootstrap.Initializer 接口
func (a *RedisInitializerAdapter) Close(ctx context.Context) error {
	return a.redisInit.Close(ctx)
}

// HealthCheck 实现 bootstrap.Initializer 接口
func (a *RedisInitializerAdapter) HealthCheck(ctx context.Context) error {
	return a.redisInit.HealthCheck(ctx)
}

// Client 获取原生 Redis 客户端
func (a *RedisInitializerAdapter) Client() *redis.Client {
	return a.redisInit.Client()
}

// RedisClient 获取完整的 RedisClient（带辅助方法）
func (a *RedisInitializerAdapter) RedisClient() *db.RedisClient {
	return a.redisInit.RedisClient()
}

// Store 获取业务层 Store（如果设置了 StoreWrapper）
//
// 返回 interface{}，调用方需要类型断言：
//
//	store := adapter.Store().(*storage.RedisStore)
func (a *RedisInitializerAdapter) Store() interface{} {
	return a.store
}
