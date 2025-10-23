package initializers

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/k8s-agent/internal/auth/config"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
	"gorm.io/gorm"
)

// DatabaseInitializer 数据库初始化器
type DatabaseInitializer struct {
	cfg    *config.Config
	logger core.Logger
	db     *gorm.DB
}

// NewDatabaseInitializer 创建数据库初始化器
func NewDatabaseInitializer(cfg *config.Config, logger core.Logger) *DatabaseInitializer {
	return &DatabaseInitializer{
		cfg:    cfg,
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
		"host", d.cfg.Database.Host,
		"dbname", d.cfg.Database.Database,
	)

	// 创建 MySQL 客户端（使用 common/db helper 函数）
	mysqlClient, err := db.NewMySQLFromOptions(d.logger, d.cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to create MySQL client: %w", err)
	}

	d.db = mysqlClient.DB

	// 自动迁移
	if d.cfg.Database.AutoMigrate {
		d.logger.Infow("Running database migrations")
		// Note: Models will be defined in internal/auth/models package
		// Skipping AutoMigrate for now, will be added when models are created
		d.logger.Infow("AutoMigrate disabled: models not yet implemented")
	}

	d.logger.Infow("Database initialized successfully")
	return nil
}

// Close 关闭数据库连接
func (d *DatabaseInitializer) Close(ctx context.Context) error {
	if d.db != nil {
		d.logger.Infow("Closing database connection")
		sqlDB, err := d.db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// HealthCheck 检查数据库健康状态
func (d *DatabaseInitializer) HealthCheck(ctx context.Context) error {
	if d.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return nil
}

// DB 获取数据库实例
func (d *DatabaseInitializer) DB() *gorm.DB {
	return d.db
}
