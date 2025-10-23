package initializers

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/k8s-agent/internal/agent-manager/config"
	"github.com/kart-io/k8s-agent/internal/agent-manager/storage"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/k8s-agent/pkg/types"
	"github.com/kart-io/logger/core"
)

// DatabaseInitializer 数据库初始化器
type DatabaseInitializer struct {
	opts   *config.Options
	logger core.Logger
	store  *storage.PostgresStore
}

// NewDatabaseInitializer 创建数据库初始化器
func NewDatabaseInitializer(opts *config.Options, logger core.Logger) *DatabaseInitializer {
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
	d.logger.Infow("Initializing database connection",
		"host", d.opts.Database.Host,
		"database", d.opts.Database.Database,
	)

	// 创建 MySQL 客户端
	mysqlClient, err := db.NewMySQL(d.logger,
		db.WithHost(d.opts.Database.Host),
		db.WithPort(d.opts.Database.Port),
		db.WithUser(d.opts.Database.User),
		db.WithPassword(d.opts.Database.Password),
		db.WithDatabase(d.opts.Database.Database),
		db.WithMaxOpenConns(d.opts.Database.MaxOpenConns),
		db.WithMaxIdleConns(d.opts.Database.MaxIdleConns),
		db.WithConnMaxLifetime(d.opts.Database.ConnMaxLifetime),
	)
	if err != nil {
		return fmt.Errorf("failed to create MySQL client: %w", err)
	}

	// 创建存储层
	d.store = &storage.PostgresStore{
		MySQLClient: mysqlClient,
	}

	// 自动迁移
	if d.opts.Database.AutoMigrate {
		d.logger.Infow("Running database migrations")
		if err := d.store.AutoMigrate(
			&types.Agent{},
			&types.Event{},
			&types.Metrics{},
			&types.Command{},
			&types.CommandResult{},
			&types.Cluster{},
			&types.AlertRule{},
			&types.Alert{},
		); err != nil {
			return fmt.Errorf("failed to auto-migrate: %w", err)
		}
	}

	d.logger.Infow("Database initialized successfully")
	return nil
}

// Close 关闭数据库连接
func (d *DatabaseInitializer) Close(ctx context.Context) error {
	if d.store != nil {
		d.logger.Infow("Closing database connection")
		return d.store.Close()
	}
	return nil
}

// HealthCheck 检查数据库健康状态
func (d *DatabaseInitializer) HealthCheck(ctx context.Context) error {
	if d.store == nil {
		return fmt.Errorf("database not initialized")
	}
	// 简单检查:database store 是否存在
	return nil
}

// Store 获取存储实例
func (d *DatabaseInitializer) Store() *storage.PostgresStore {
	return d.store
}
