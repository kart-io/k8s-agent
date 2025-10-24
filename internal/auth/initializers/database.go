package initializers

import (
	"context"

	"github.com/kart-io/k8s-agent/internal/auth/config"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
	"gorm.io/gorm"
)

// DatabaseInitializer 数据库初始化器（适配器）
//
// 此初始化器是一个适配器，内部使用通用的 pkg/initializers.DatabaseInitializer，
// 但提供 DB() 方法返回 *gorm.DB，以保持与现有代码的兼容性。
type DatabaseInitializer struct {
	cfg    *config.Config
	logger core.Logger

	// 使用通用初始化器
	dbInit *pkginitializers.DatabaseInitializer
}

// NewDatabaseInitializer 创建数据库初始化器
func NewDatabaseInitializer(cfg *config.Config, logger core.Logger) *DatabaseInitializer {
	// 创建通用数据库初始化器
	dbInit := pkginitializers.NewDatabaseInitializer(
		cfg.Database,
		logger,
	)

	// Note: Models will be defined in internal/auth/models package
	// When models are created, uncomment the following:
	// if cfg.Database.AutoMigrate {
	//     dbInit.WithAutoMigrate(&models.User{}, &models.Session{}, ...)
	// }

	return &DatabaseInitializer{
		cfg:    cfg,
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
	return d.dbInit.Initialize(ctx)
}

// Close 关闭数据库连接
func (d *DatabaseInitializer) Close(ctx context.Context) error {
	return d.dbInit.Close(ctx)
}

// HealthCheck 检查数据库健康状态
func (d *DatabaseInitializer) HealthCheck(ctx context.Context) error {
	return d.dbInit.HealthCheck(ctx)
}

// DB 获取数据库实例
func (d *DatabaseInitializer) DB() *gorm.DB {
	return d.dbInit.DB()
}
