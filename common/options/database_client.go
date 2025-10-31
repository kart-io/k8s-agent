package options

import (
	"github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/logger/core"
)

// NewMySQLClient 根据配置创建 MySQL 客户端
// 这个方法直接在 Options 中实现数据库连接创建，简化调用链
//
// 之前的调用方式：
//
//	client, err := db.NewMySQLFromOptions(logger, opts.Database)
//
// 现在的调用方式：
//
//	client, err := opts.Database.NewMySQLClient(logger)
//
// 使用示例：
//
//	opts := cmd/service/app/options.NewServerOptions()
//	client, err := opts.Database.NewMySQLClient(logger)
//	if err != nil {
//	    return err
//	}
func (o *DatabaseOptions) NewMySQLClient(log core.Logger) (*db.MySQLClient, error) {
	return db.NewMySQL(log,
		db.WithHost(o.Host),
		db.WithPort(o.Port),
		db.WithUser(o.User),
		db.WithPassword(o.Password),
		db.WithDatabase(o.Database),
		db.WithMaxOpenConns(o.MaxOpenConns),
		db.WithMaxIdleConns(o.MaxIdleConns),
		db.WithConnMaxLifetime(o.ConnMaxLifetime),
	)
}
