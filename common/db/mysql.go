package db

import (
	"context"
	"fmt"
	"time"

	"github.com/kart-io/logger/core"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MySQLOptions MySQL 配置选项
type MySQLOptions struct {
	host            string
	port            int
	user            string
	password        string
	database        string
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
	logLevel        string
}

// MySQLOption MySQL 配置选项函数
type MySQLOption func(*MySQLOptions)

// WithHost 设置主机地址
func WithHost(host string) MySQLOption {
	return func(o *MySQLOptions) {
		o.host = host
	}
}

// WithPort 设置端口
func WithPort(port int) MySQLOption {
	return func(o *MySQLOptions) {
		o.port = port
	}
}

// WithUser 设置用户名
func WithUser(user string) MySQLOption {
	return func(o *MySQLOptions) {
		o.user = user
	}
}

// WithPassword 设置密码
func WithPassword(password string) MySQLOption {
	return func(o *MySQLOptions) {
		o.password = password
	}
}

// WithDatabase 设置数据库名
func WithDatabase(database string) MySQLOption {
	return func(o *MySQLOptions) {
		o.database = database
	}
}

// WithMaxOpenConns 设置最大打开连接数
func WithMaxOpenConns(n int) MySQLOption {
	return func(o *MySQLOptions) {
		o.maxOpenConns = n
	}
}

// WithMaxIdleConns 设置最大空闲连接数
func WithMaxIdleConns(n int) MySQLOption {
	return func(o *MySQLOptions) {
		o.maxIdleConns = n
	}
}

// WithConnMaxLifetime 设置连接最大生命周期
func WithConnMaxLifetime(d time.Duration) MySQLOption {
	return func(o *MySQLOptions) {
		o.connMaxLifetime = d
	}
}

// WithLogLevel 设置日志级别 (silent, error, warn, info)
func WithLogLevel(level string) MySQLOption {
	return func(o *MySQLOptions) {
		o.logLevel = level
	}
}

// defaultMySQLOptions 返回默认 MySQL 配置
func defaultMySQLOptions() *MySQLOptions {
	return &MySQLOptions{
		host:            "localhost",
		port:            3306,
		user:            "root",
		password:        "",
		database:        "test",
		maxOpenConns:    100,
		maxIdleConns:    10,
		connMaxLifetime: time.Hour,
		logLevel:        "silent",
	}
}

// MySQLClient MySQL 客户端
type MySQLClient struct {
	DB     *gorm.DB
	logger core.Logger
}

// NewMySQL 创建 MySQL 连接
func NewMySQL(log core.Logger, opts ...MySQLOption) (*MySQLClient, error) {
	// 应用默认配置
	options := defaultMySQLOptions()

	// 应用用户配置
	for _, opt := range opts {
		opt(options)
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		options.user, options.password, options.host, options.port, options.database,
	)

	// GORM 日志级别
	gormLogLevel := logger.Silent
	switch options.logLevel {
	case "error":
		gormLogLevel = logger.Error
	case "warn":
		gormLogLevel = logger.Warn
	case "info":
		gormLogLevel = logger.Info
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// 连接池配置
	sqlDB.SetMaxOpenConns(options.maxOpenConns)
	sqlDB.SetMaxIdleConns(options.maxIdleConns)
	sqlDB.SetConnMaxLifetime(options.connMaxLifetime)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL: %w", err)
	}

	log.Infow("MySQL connected",
		"host", options.host,
		"port", options.port,
		"database", options.database,
	)

	return &MySQLClient{
		DB:     db,
		logger: log.With("component", "mysql"),
	}, nil
}

// Health 检查数据库健康
func (c *MySQLClient) Health(ctx context.Context) error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// Close 关闭连接
func (c *MySQLClient) Close() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}

	c.logger.Info("Closing MySQL connection")
	return sqlDB.Close()
}

// Shutdown 实现 ShutdownHandler 接口
func (c *MySQLClient) Shutdown(ctx context.Context) error {
	return c.Close()
}

// AutoMigrate 自动迁移数据库表结构
func (c *MySQLClient) AutoMigrate(models ...interface{}) error {
	c.logger.Info("Running database migrations")
	if err := c.DB.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}
	c.logger.Info("Database migrations completed successfully")
	return nil
}

// NewMySQLFromDSN 从 DSN 字符串创建 MySQL 连接
// dsn 格式: user:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local
func NewMySQLFromDSN(log core.Logger, dsn string, opts ...MySQLOption) (*MySQLClient, error) {
	// 应用默认配置
	options := defaultMySQLOptions()

	// 应用用户配置
	for _, opt := range opts {
		opt(options)
	}

	// GORM 日志级别
	gormLogLevel := logger.Silent
	switch options.logLevel {
	case "error":
		gormLogLevel = logger.Error
	case "warn":
		gormLogLevel = logger.Warn
	case "info":
		gormLogLevel = logger.Info
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// 连接池配置
	sqlDB.SetMaxOpenConns(options.maxOpenConns)
	sqlDB.SetMaxIdleConns(options.maxIdleConns)
	sqlDB.SetConnMaxLifetime(options.connMaxLifetime)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL: %w", err)
	}

	log.Infow("MySQL connected", "dsn", dsn)

	return &MySQLClient{
		DB:     db,
		logger: log.With("component", "mysql"),
	}, nil
}
