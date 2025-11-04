package initializers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
)

// DatabaseInitializer 通用数据库初始化器
//
// 负责初始化MySQL数据库连接，提供以下功能：
// - 数据库连接初始化
// - 连接池配置
// - 可选的自动迁移
// - 健康检查
// - 优雅关闭.
type DatabaseInitializer struct {
	opts   *options.DatabaseOptions
	logger core.Logger
	client *db.MySQLClient

	// 可选的自动迁移
	autoMigrate bool
	models      []interface{}
}

// NewDatabaseInitializer 创建数据库初始化器
//
// 参数：
//   - opts: 数据库配置选项
//   - logger: 日志记录器
//
// 返回：
//   - *DatabaseInitializer: 数据库初始化器实例
func NewDatabaseInitializer(
	opts *options.DatabaseOptions,
	logger core.Logger,
) *DatabaseInitializer {
	return &DatabaseInitializer{
		opts:   opts,
		logger: logger,
	}
}

// WithAutoMigrate 启用自动迁移
//
// 设置需要自动迁移的数据库模型。
// 在 Initialize() 时会自动执行 GORM 的 AutoMigrate。
//
// 参数：
//   - models: 需要迁移的模型列表
//
// 返回：
//   - *DatabaseInitializer: 返回自身以支持链式调用
//
// 示例：
//
//	dbInit := NewDatabaseInitializer(opts, logger).WithAutoMigrate(
//	    &types.Agent{},
//	    &types.Event{},
//	    &types.Command{},
//	)
func (d *DatabaseInitializer) WithAutoMigrate(models ...interface{}) *DatabaseInitializer {
	d.autoMigrate = true
	d.models = models
	return d
}

// Name 返回初始化器名称.
func (d *DatabaseInitializer) Name() string {
	return "database"
}

// Priority 返回初始化优先级
//
// 数据库初始化器应该在配置和日志之后，业务逻辑之前初始化.
func (d *DatabaseInitializer) Priority() int {
	return bootstrap.PriorityDatabase
}

// Initialize 执行初始化
//
// 初始化流程：
//  1. 创建 MySQL 客户端
//  2. 测试数据库连接
//  3. 执行自动迁移（如果启用）
//
// 参数：
//   - ctx: 上下文对象
//
// 返回：
//   - error: 初始化错误
func (d *DatabaseInitializer) Initialize(ctx context.Context) error {
	d.logger.Infow("Initializing database connection",
		"host", d.opts.Host,
		"port", d.opts.Port,
		"database", d.opts.Database,
	)

	// 直接使用 db 包创建客户端
	client, err := db.NewMySQL(d.logger,
		db.WithHost(d.opts.Host),
		db.WithPort(d.opts.Port),
		db.WithUser(d.opts.User),
		db.WithPassword(d.opts.Password),
		db.WithDatabase(d.opts.Database),
		db.WithMaxOpenConns(d.opts.MaxOpenConns),
		db.WithMaxIdleConns(d.opts.MaxIdleConns),
		db.WithConnMaxLifetime(d.opts.ConnMaxLifetime),
		db.WithLogLevel(d.opts.LogLevel),
	)
	if err != nil {
		return fmt.Errorf("failed to create MySQL client: %w", err)
	}

	d.client = client

	// 执行自动迁移（如果启用）
	if d.autoMigrate && len(d.models) > 0 {
		d.logger.Infow("Running database migrations",
			"models", len(d.models),
		)
		if err := d.client.AutoMigrate(d.models...); err != nil {
			return fmt.Errorf("failed to auto-migrate: %w", err)
		}
		d.logger.Infow("Database migrations completed successfully")
	}

	d.logger.Infow("Database initialized successfully")
	return nil
}

// Close 关闭数据库连接
//
// 优雅关闭数据库连接池，释放所有资源。
//
// 参数：
//   - ctx: 上下文对象（用于超时控制）
//
// 返回：
//   - error: 关闭错误
func (d *DatabaseInitializer) Close(ctx context.Context) error {
	if d.client != nil {
		d.logger.Infow("Closing database connection")
		return d.client.Close()
	}
	return nil
}

// HealthCheck 检查数据库健康状态
//
// 通过执行 Ping 操作检查数据库连接是否正常。
//
// 参数：
//   - ctx: 上下文对象（用于超时控制）
//
// 返回：
//   - error: 健康检查错误
func (d *DatabaseInitializer) HealthCheck(ctx context.Context) error {
	if d.client == nil {
		return fmt.Errorf("database not initialized")
	}
	return d.client.Health(ctx)
}

// DB 获取数据库实例
//
// 返回 GORM DB 实例，供业务逻辑使用。
//
// 返回：
//   - *gorm.DB: GORM 数据库实例
func (d *DatabaseInitializer) DB() *gorm.DB {
	if d.client == nil {
		return nil
	}
	return d.client.DB
}

// Client 获取完整的 MySQL 客户端
//
// 返回完整的 MySQLClient 实例，包含了额外的辅助方法。
//
// 返回：
//   - *db.MySQLClient: MySQL 客户端实例
func (d *DatabaseInitializer) Client() *db.MySQLClient {
	return d.client
}
