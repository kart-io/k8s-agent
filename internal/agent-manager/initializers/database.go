package initializers

import (
	"context"

	"github.com/kart-io/k8s-agent/internal/agent-manager/config"
	"github.com/kart-io/k8s-agent/internal/agent-manager/storage"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/k8s-agent/pkg/types"
	"github.com/kart-io/logger/core"
)

// DatabaseInitializer 数据库初始化器（适配器）
//
// 此初始化器是一个适配器，内部使用通用的 pkg/initializers.DatabaseInitializer，
// 但提供 Store() 方法返回业务特定的 storage.PostgresStore，
// 以保持与现有代码的兼容性。
type DatabaseInitializer struct {
	opts   *config.Options
	logger core.Logger

	// 使用通用初始化器
	dbInit *pkginitializers.DatabaseInitializer
	store  *storage.PostgresStore
}

// NewDatabaseInitializer 创建数据库初始化器
func NewDatabaseInitializer(opts *config.Options, logger core.Logger) *DatabaseInitializer {
	// 创建通用数据库初始化器
	dbInit := pkginitializers.NewDatabaseInitializer(
		opts.Database,
		logger,
	)

	// 如果配置了自动迁移，设置模型
	if opts.Database.AutoMigrate {
		dbInit.WithAutoMigrate(
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

	return &DatabaseInitializer{
		opts:   opts,
		logger: logger,
		dbInit: dbInit,
	}
}

// Name 返回初始化器名称
func (d *DatabaseInitializer) Name() string {
	return d.dbInit.Name()
}

// Priority 返回初始化优先级
func (d *DatabaseInitializer) Priority() int {
	return bootstrap.PriorityDatabase
}

// Initialize 执行初始化
func (d *DatabaseInitializer) Initialize(ctx context.Context) error {
	// 委托给通用初始化器
	if err := d.dbInit.Initialize(ctx); err != nil {
		return err
	}

	// 创建业务存储层（包装通用客户端）
	d.store = &storage.PostgresStore{
		MySQLClient: d.dbInit.Client(),
	}

	return nil
}

// Close 关闭数据库连接
func (d *DatabaseInitializer) Close(ctx context.Context) error {
	return d.dbInit.Close(ctx)
}

// HealthCheck 检查数据库健康状态
func (d *DatabaseInitializer) HealthCheck(ctx context.Context) error {
	return d.dbInit.HealthCheck(ctx)
}

// Store 获取存储实例
//
// 返回业务特定的 PostgresStore，供其他组件使用。
func (d *DatabaseInitializer) Store() *storage.PostgresStore {
	return d.store
}
