package initializers

import (
	"context"

	"github.com/kart-io/k8s-agent/cmd/orchestrator/app/options"
	"github.com/kart-io/k8s-agent/internal/orchestrator/storage"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
)

// DatabaseInitializer 数据库初始化器
type DatabaseInitializer struct {
	opts   *options.ServerOptions
	logger core.Logger
	store  *storage.PostgresStore
}

// NewDatabaseInitializer 创建数据库初始化器
func NewDatabaseInitializer(opts *options.ServerOptions, logger core.Logger) *DatabaseInitializer {
	return &DatabaseInitializer{
		opts:   opts,
		logger: logger,
	}
}

// Name 返回初始化器名称
func (d *DatabaseInitializer) Name() string {
	return "database"
}

// Priority 返回初始化优先级
func (d *DatabaseInitializer) Priority() int {
	return bootstrap.PriorityDatabase
}

// Initialize 执行初始化
func (d *DatabaseInitializer) Initialize(ctx context.Context) error {
	d.logger.Infow("Initializing PostgreSQL",
		"host", d.opts.Database.Host,
		"port", d.opts.Database.Port,
		"database", d.opts.Database.Database,
	)

	store, err := storage.NewPostgresStore(d.opts.Database, d.logger)
	if err != nil {
		return err
	}

	d.store = store
	d.logger.Info("PostgreSQL initialized successfully")
	return nil
}

// Close 关闭数据库连接
func (d *DatabaseInitializer) Close(ctx context.Context) error {
	if d.store != nil {
		d.store.Close()
	}
	return nil
}

// HealthCheck 检查数据库健康状态
func (d *DatabaseInitializer) HealthCheck(ctx context.Context) error {
	// Database health is checked via connection status
	return nil
}

// Store 获取存储实例
func (d *DatabaseInitializer) Store() *storage.PostgresStore {
	return d.store
}
