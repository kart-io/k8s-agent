package initializers

import (
	pkginitializers "github.com/kart-io/k8s-agent/common/initializers"
	"github.com/kart-io/k8s-agent/internal/auth/config"
	"github.com/kart-io/logger/core"
	"gorm.io/gorm"
)

// DatabaseInitializer 数据库初始化器（使用通用适配器）
//
// 现在使用 pkg/initializers.DatabaseInitializerAdapter 来消除重复代码。
type DatabaseInitializer struct {
	*pkginitializers.DatabaseInitializerAdapter
}

// NewDatabaseInitializer 创建数据库初始化器
func NewDatabaseInitializer(cfg *config.Config, logger core.Logger) *DatabaseInitializer {
	// 创建通用适配器
	adapter := pkginitializers.NewDatabaseInitializerAdapter(cfg.Database, logger)

	// Note: Models will be defined in internal/auth/models package
	// When models are created, uncomment the following:
	// if cfg.Database.AutoMigrate {
	//     adapter.WithAutoMigrate(&models.User{}, &models.Session{}, ...)
	// }

	return &DatabaseInitializer{
		DatabaseInitializerAdapter: adapter,
	}
}

// DB 获取数据库实例（类型安全的便捷方法）
func (d *DatabaseInitializer) DB() *gorm.DB {
	return d.DatabaseInitializerAdapter.DB()
}
