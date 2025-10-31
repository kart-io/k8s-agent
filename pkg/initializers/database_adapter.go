package initializers

import (
	"context"

	"github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
	"gorm.io/gorm"
)

// DatabaseInitializerAdapter 通用数据库初始化器适配器
//
// 这个适配器包装了通用的 DatabaseInitializer，
// 并提供了多种返回类型的便捷方法，消除各服务的重复代码。
//
// 使用示例：
//
//	adapter := pkginitializers.NewDatabaseInitializerAdapter(opts.Database, logger)
//	adapter.WithAutoMigrate(&types.User{}, &types.Session{})
//	// 在服务的 RegisterComponents 中注册
//	bs.Register(adapter)
//
// 优势：
//  1. 消除了每个服务都需要编写的适配器代码
//  2. 提供统一的接口和行为
//  3. 支持自动迁移配置
//  4. 支持多种返回类型（Client, DB, 自定义 Store）
type DatabaseInitializerAdapter struct {
	dbInit *DatabaseInitializer

	// 可选：业务层 Store 包装函数
	// 在 Initialize 完成后调用，用于创建服务特定的 Store
	storeWrapper func(*db.MySQLClient) interface{}
	store        interface{}
}

// NewDatabaseInitializerAdapter 创建数据库初始化器适配器
func NewDatabaseInitializerAdapter(
	opts *options.DatabaseOptions,
	logger core.Logger,
) *DatabaseInitializerAdapter {
	return &DatabaseInitializerAdapter{
		dbInit: NewDatabaseInitializer(opts, logger),
	}
}

// WithAutoMigrate 设置自动迁移的模型
// 返回自身以支持链式调用
func (a *DatabaseInitializerAdapter) WithAutoMigrate(models ...interface{}) *DatabaseInitializerAdapter {
	a.dbInit.WithAutoMigrate(models...)
	return a
}

// WithStoreWrapper 设置业务层 Store 包装函数
//
// 这个函数在数据库初始化完成后被调用，用于创建服务特定的 Store。
//
// 使用示例：
//
//	adapter.WithStoreWrapper(func(client *db.MySQLClient) interface{} {
//	    return &storage.PostgresStore{MySQLClient: client}
//	})
func (a *DatabaseInitializerAdapter) WithStoreWrapper(wrapper func(*db.MySQLClient) interface{}) *DatabaseInitializerAdapter {
	a.storeWrapper = wrapper
	return a
}

// Name 实现 bootstrap.Initializer 接口
func (a *DatabaseInitializerAdapter) Name() string {
	return a.dbInit.Name()
}

// Priority 实现 bootstrap.Initializer 接口
func (a *DatabaseInitializerAdapter) Priority() int {
	return bootstrap.PriorityDatabase
}

// Initialize 实现 bootstrap.Initializer 接口
func (a *DatabaseInitializerAdapter) Initialize(ctx context.Context) error {
	// 委托给通用初始化器
	if err := a.dbInit.Initialize(ctx); err != nil {
		return err
	}

	// 如果设置了 Store 包装函数，创建业务层 Store
	if a.storeWrapper != nil {
		a.store = a.storeWrapper(a.dbInit.Client())
	}

	return nil
}

// Close 实现 bootstrap.Initializer 接口
func (a *DatabaseInitializerAdapter) Close(ctx context.Context) error {
	return a.dbInit.Close(ctx)
}

// HealthCheck 实现 bootstrap.Initializer 接口
func (a *DatabaseInitializerAdapter) HealthCheck(ctx context.Context) error {
	return a.dbInit.HealthCheck(ctx)
}

// Client 获取 MySQL 客户端
func (a *DatabaseInitializerAdapter) Client() *db.MySQLClient {
	return a.dbInit.Client()
}

// DB 获取 GORM DB 实例
func (a *DatabaseInitializerAdapter) DB() *gorm.DB {
	return a.dbInit.DB()
}

// Store 获取业务层 Store（如果设置了 StoreWrapper）
//
// 返回 interface{}，调用方需要类型断言：
//
//	store := adapter.Store().(*storage.PostgresStore)
func (a *DatabaseInitializerAdapter) Store() interface{} {
	return a.store
}
