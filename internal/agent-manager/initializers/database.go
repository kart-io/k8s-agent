package initializers

import (
	"github.com/kart-io/k8s-agent/cmd/agent-manager/app/options"
	"github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/k8s-agent/internal/agent-manager/storage"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/k8s-agent/pkg/types"
	"github.com/kart-io/logger/core"
)

// DatabaseInitializer 数据库初始化器（使用通用适配器）
//
// 现在使用 pkg/initializers.DatabaseInitializerAdapter 来消除重复代码。
type DatabaseInitializer struct {
	*pkginitializers.DatabaseInitializerAdapter
}

// NewDatabaseInitializer 创建数据库初始化器
func NewDatabaseInitializer(opts *options.ServerOptions, logger core.Logger) *DatabaseInitializer {
	// 创建通用适配器
	adapter := pkginitializers.NewDatabaseInitializerAdapter(opts.Database, logger)

	// 配置自动迁移
	if opts.Database.AutoMigrate {
		adapter.WithAutoMigrate(
			&types.Agent{},
			&types.Event{},
			&types.Metrics{},
			&types.Command{},
			&types.CommandResult{},
			&types.Cluster{},
			&types.AlertRule{},
			&types.Alert{},
		)
	}

	// 配置 Store 包装函数
	adapter.WithStoreWrapper(func(client *db.MySQLClient) interface{} {
		return &storage.PostgresStore{
			MySQLClient: client,
		}
	})

	return &DatabaseInitializer{
		DatabaseInitializerAdapter: adapter,
	}
}

// Store 获取存储实例（类型安全的便捷方法）
//
// 返回业务特定的 PostgresStore，供其他组件使用。
func (d *DatabaseInitializer) Store() *storage.PostgresStore {
	if store := d.DatabaseInitializerAdapter.Store(); store != nil {
		return store.(*storage.PostgresStore)
	}
	return nil
}
