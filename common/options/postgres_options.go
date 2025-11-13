package options

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/spf13/pflag"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/kart-io/k8s-agent/common/options/validation"
	"github.com/kart-io/logger/core"
)

// PostgresOptions PostgreSQL 数据库配置
type PostgresOptions struct {
	Host                 string        `mapstructure:"host" yaml:"host" json:"host"`
	Port                 int           `mapstructure:"port" yaml:"port" json:"port"`
	User                 string        `mapstructure:"user" yaml:"user" json:"user"`
	Password             string        `mapstructure:"password" yaml:"password" json:"-"` // 密码不序列化到 JSON
	Database             string        `mapstructure:"database" yaml:"database" json:"database"`
	SSLMode              string        `mapstructure:"ssl_mode" yaml:"ssl_mode" json:"ssl_mode"` // disable, allow, prefer, require, verify-ca, verify-full
	TimeZone             string        `mapstructure:"timezone" yaml:"timezone" json:"timezone"` // 时区设置
	Schema               string        `mapstructure:"schema" yaml:"schema" json:"schema"`       // PostgreSQL schema
	MaxOpenConns         int           `mapstructure:"max_open_conns" yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns         int           `mapstructure:"max_idle_conns" yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime      time.Duration `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
	LogLevel             string        `mapstructure:"log_level" yaml:"log_level" json:"log_level"` // silent, error, warn, info
	SlowQueryThreshold   time.Duration `mapstructure:"slow_query_threshold" yaml:"slow_query_threshold" json:"slow_query_threshold"`
	AutoMigrate          bool          `mapstructure:"auto_migrate" yaml:"auto_migrate" json:"auto_migrate"`
	PreferSimpleProtocol bool          `mapstructure:"prefer_simple_protocol" yaml:"prefer_simple_protocol" json:"prefer_simple_protocol"` // 禁用 prepared statement 缓存

	// Connection pool tuning
	ConnMaxIdleTime    time.Duration `mapstructure:"conn_max_idle_time" yaml:"conn_max_idle_time" json:"conn_max_idle_time"`
	StatementCacheMode string        `mapstructure:"statement_cache_mode" yaml:"statement_cache_mode" json:"statement_cache_mode"` // describe, prepare
}

// NewPostgresOptions 创建默认的 PostgreSQL 数据库配置
func NewPostgresOptions() *PostgresOptions {
	return &PostgresOptions{
		Host:                 "localhost",
		Port:                 5432,
		User:                 "postgres",
		Password:             "",
		Database:             "postgres",
		SSLMode:              "prefer",
		TimeZone:             "Asia/Shanghai",
		Schema:               "public",
		MaxOpenConns:         100,
		MaxIdleConns:         10,
		ConnMaxLifetime:      1 * time.Hour,
		ConnMaxIdleTime:      10 * time.Minute,
		LogLevel:             "silent", // 默认静默模式
		SlowQueryThreshold:   200 * time.Millisecond,
		AutoMigrate:          false,
		PreferSimpleProtocol: false,
		StatementCacheMode:   "prepare",
	}
}

// Validate 验证配置
func (o *PostgresOptions) Validate() error {
	// 使用通用验证器
	if err := validation.ValidateRequired(o.Host, "PostgreSQL host"); err != nil {
		return err
	}

	if err := validation.ValidatePort(o.Port, "PostgreSQL"); err != nil {
		return err
	}

	if err := validation.ValidateRequired(o.User, "PostgreSQL user"); err != nil {
		return err
	}

	if err := validation.ValidateRequired(o.Database, "PostgreSQL database name"); err != nil {
		return err
	}

	// 验证 SSL 模式
	validSSLModes := map[string]bool{
		"disable":     true,
		"allow":       true,
		"prefer":      true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}
	if o.SSLMode != "" && !validSSLModes[o.SSLMode] {
		return fmt.Errorf("invalid SSL mode: %s, must be one of: disable, allow, prefer, require, verify-ca, verify-full", o.SSLMode)
	}

	// 验证 statement cache mode
	if o.StatementCacheMode != "" && o.StatementCacheMode != "describe" && o.StatementCacheMode != "prepare" {
		return fmt.Errorf("invalid statement cache mode: %s, must be 'describe' or 'prepare'", o.StatementCacheMode)
	}

	if err := validation.ValidateConnectionPool(o.MaxOpenConns, o.MaxIdleConns, "PostgreSQL"); err != nil {
		return err
	}

	return nil
}

// DSN 返回 PostgreSQL 连接字符串
func (o *PostgresOptions) DSN() string {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		o.Host, o.Port, o.User, o.Password, o.Database, o.SSLMode)

	if o.TimeZone != "" {
		dsn += fmt.Sprintf(" TimeZone=%s", o.TimeZone)
	}

	if o.Schema != "" && o.Schema != "public" {
		dsn += fmt.Sprintf(" search_path=%s", o.Schema)
	}

	if o.PreferSimpleProtocol {
		dsn += " prefer_simple_protocol=true"
	}

	if o.StatementCacheMode != "" {
		dsn += fmt.Sprintf(" statement_cache_mode=%s", o.StatementCacheMode)
	}

	return dsn
}

// AddFlags 添加命令行参数
func (o *PostgresOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Host, "postgres.host", o.Host, "PostgreSQL host address")
	fs.IntVar(&o.Port, "postgres.port", o.Port, "PostgreSQL port")
	fs.StringVar(&o.User, "postgres.user", o.User, "PostgreSQL user")
	fs.StringVar(&o.Password, "postgres.password", o.Password, "PostgreSQL password")
	fs.StringVar(&o.Database, "postgres.database", o.Database, "PostgreSQL database name")
	fs.StringVar(&o.SSLMode, "postgres.ssl-mode", o.SSLMode, "PostgreSQL SSL mode (disable, allow, prefer, require, verify-ca, verify-full)")
	fs.StringVar(&o.TimeZone, "postgres.timezone", o.TimeZone, "PostgreSQL timezone")
	fs.StringVar(&o.Schema, "postgres.schema", o.Schema, "PostgreSQL schema")
	fs.IntVar(&o.MaxOpenConns, "postgres.max-open-conns", o.MaxOpenConns, "Maximum number of open connections")
	fs.IntVar(&o.MaxIdleConns, "postgres.max-idle-conns", o.MaxIdleConns, "Maximum number of idle connections")
	fs.DurationVar(&o.ConnMaxLifetime, "postgres.conn-max-lifetime", o.ConnMaxLifetime, "Maximum connection lifetime")
	fs.DurationVar(&o.ConnMaxIdleTime, "postgres.conn-max-idle-time", o.ConnMaxIdleTime, "Maximum connection idle time")
	fs.StringVar(&o.LogLevel, "postgres.log-level", o.LogLevel, "GORM log level (silent, error, warn, info)")
	fs.DurationVar(&o.SlowQueryThreshold, "postgres.slow-query-threshold", o.SlowQueryThreshold, "GORM slow query threshold")
	fs.BoolVar(&o.AutoMigrate, "postgres.auto-migrate", o.AutoMigrate, "Enable automatic PostgreSQL database migration")
	fs.BoolVar(&o.PreferSimpleProtocol, "postgres.prefer-simple-protocol", o.PreferSimpleProtocol, "Disable prepared statement cache")
	fs.StringVar(&o.StatementCacheMode, "postgres.statement-cache-mode", o.StatementCacheMode, "Statement cache mode (describe, prepare)")
}

// ApplyTo 将配置应用到目标接口
func (o *PostgresOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	// 类型断言，检查是否为函数选项切片指针
	switch v := target.(type) {
	case *[]interface{}:
		// 将配置转换为通用选项
		*v = append(*v,
			map[string]interface{}{
				"dsn":                  o.DSN(),
				"host":                 o.Host,
				"port":                 o.Port,
				"user":                 o.User,
				"password":             o.Password,
				"database":             o.Database,
				"sslMode":              o.SSLMode,
				"timezone":             o.TimeZone,
				"schema":               o.Schema,
				"maxOpenConns":         o.MaxOpenConns,
				"maxIdleConns":         o.MaxIdleConns,
				"connMaxLifetime":      o.ConnMaxLifetime,
				"connMaxIdleTime":      o.ConnMaxIdleTime,
				"logLevel":             o.LogLevel,
				"slowQueryThreshold":   o.SlowQueryThreshold,
				"autoMigrate":          o.AutoMigrate,
				"preferSimpleProtocol": o.PreferSimpleProtocol,
				"statementCacheMode":   o.StatementCacheMode,
			},
		)
	}

	return nil
}

// Complete 完成配置初始化
func (o *PostgresOptions) Complete() error {
	// 如果地址为空，设置默认值
	if o.Host == "" {
		o.Host = "localhost"
	}

	// 确保端口在有效范围内
	if o.Port <= 0 || o.Port > 65535 {
		o.Port = 5432
	}

	// 设置默认用户
	if o.User == "" {
		o.User = "postgres"
	}

	// 设置默认数据库
	if o.Database == "" {
		o.Database = "postgres"
	}

	// 设置默认 SSL 模式
	if o.SSLMode == "" {
		o.SSLMode = "prefer"
	}

	// 设置默认时区
	if o.TimeZone == "" {
		o.TimeZone = "Asia/Shanghai"
	}

	// 设置默认 schema
	if o.Schema == "" {
		o.Schema = "public"
	}

	// 设置默认日志级别
	if o.LogLevel == "" {
		o.LogLevel = "silent"
	}

	// 确保连接池配置合理
	if o.MaxOpenConns <= 0 {
		o.MaxOpenConns = 100
	}

	if o.MaxIdleConns <= 0 {
		o.MaxIdleConns = 10
	}

	// 确保 MaxIdleConns 不大于 MaxOpenConns
	if o.MaxIdleConns > o.MaxOpenConns {
		o.MaxIdleConns = o.MaxOpenConns
	}

	// 确保超时时间有合理的值
	if o.ConnMaxLifetime <= 0 {
		o.ConnMaxLifetime = time.Hour
	}

	if o.ConnMaxIdleTime <= 0 {
		o.ConnMaxIdleTime = 10 * time.Minute
	}

	if o.SlowQueryThreshold <= 0 {
		o.SlowQueryThreshold = 200 * time.Millisecond
	}

	// 设置默认 statement cache mode
	if o.StatementCacheMode == "" {
		o.StatementCacheMode = "prepare"
	}

	return nil
}

// WithPostgresHost 设置 PostgreSQL 地址
func WithPostgresHost(host string) func(*PostgresOptions) {
	return func(o *PostgresOptions) {
		o.Host = host
	}
}

// WithPostgresPort 设置 PostgreSQL 端口
func WithPostgresPort(port int) func(*PostgresOptions) {
	return func(o *PostgresOptions) {
		o.Port = port
	}
}

// WithPostgresUser 设置 PostgreSQL 用户
func WithPostgresUser(user string) func(*PostgresOptions) {
	return func(o *PostgresOptions) {
		o.User = user
	}
}

// WithPostgresPassword 设置 PostgreSQL 密码
func WithPostgresPassword(password string) func(*PostgresOptions) {
	return func(o *PostgresOptions) {
		o.Password = password
	}
}

// WithPostgresDatabase 设置 PostgreSQL 数据库
func WithPostgresDatabase(database string) func(*PostgresOptions) {
	return func(o *PostgresOptions) {
		o.Database = database
	}
}

// WithPostgresSSLMode 设置 SSL 模式
func WithPostgresSSLMode(mode string) func(*PostgresOptions) {
	return func(o *PostgresOptions) {
		o.SSLMode = mode
	}
}

// WithPostgresSchema 设置 schema
func WithPostgresSchema(schema string) func(*PostgresOptions) {
	return func(o *PostgresOptions) {
		o.Schema = schema
	}
}

// WithPostgresMaxOpenConns 设置最大打开连接数
func WithPostgresMaxOpenConns(n int) func(*PostgresOptions) {
	return func(o *PostgresOptions) {
		o.MaxOpenConns = n
	}
}

// WithPostgresMaxIdleConns 设置最大空闲连接数
func WithPostgresMaxIdleConns(n int) func(*PostgresOptions) {
	return func(o *PostgresOptions) {
		o.MaxIdleConns = n
	}
}

// WithPostgresAutoMigrate 设置自动迁移
func WithPostgresAutoMigrate(enable bool) func(*PostgresOptions) {
	return func(o *PostgresOptions) {
		o.AutoMigrate = enable
	}
}

// getGormLogLevel 转换日志级别
func (o *PostgresOptions) getGormLogLevel() logger.LogLevel {
	switch o.LogLevel {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Silent
	}
}

// ConnectPostgres 连接 PostgreSQL 数据库
func (o *PostgresOptions) ConnectPostgres(log core.Logger) (*gorm.DB, error) {
	// Create GORM config
	config := &gorm.Config{
		Logger: logger.Default.LogMode(o.getGormLogLevel()),
	}

	// Set slow query threshold
	if o.SlowQueryThreshold > 0 {
		config.Logger = logger.Default.LogMode(o.getGormLogLevel()).
			LogMode(o.getGormLogLevel())
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(o.DSN()), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Get underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL database: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(o.MaxOpenConns)
	sqlDB.SetMaxIdleConns(o.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(o.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(o.ConnMaxIdleTime)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	if log != nil {
		log.Info("Successfully connected to PostgreSQL",
			"host", o.Host,
			"port", o.Port,
			"database", o.Database,
			"user", o.User,
			"ssl_mode", o.SSLMode,
		)
	}

	return db, nil
}

// ConnectRawPostgres 创建原生 PostgreSQL 连接（不使用 GORM）
func (o *PostgresOptions) ConnectRawPostgres(log core.Logger) (*sql.DB, error) {
	// Build DSN for lib/pq
	dsn := o.DSN()

	// Open database connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(o.MaxOpenConns)
	db.SetMaxIdleConns(o.MaxIdleConns)
	db.SetConnMaxLifetime(o.ConnMaxLifetime)
	db.SetConnMaxIdleTime(o.ConnMaxIdleTime)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	if log != nil {
		log.Info("Successfully connected to PostgreSQL (raw)",
			"host", o.Host,
			"port", o.Port,
			"database", o.Database,
			"user", o.User,
		)
	}

	return db, nil
}

// Health 检查 PostgreSQL 健康状态
func (o *PostgresOptions) Health(ctx context.Context, log core.Logger) error {
	// Try to connect using raw connection for health check
	db, err := o.ConnectRawPostgres(log)
	if err != nil {
		return errors.New("failed to connect to PostgreSQL: " + err.Error())
	}
	defer db.Close()

	// Use context for timeout control
	if err := db.PingContext(ctx); err != nil {
		return errors.New("failed to ping PostgreSQL: " + err.Error())
	}

	// Check database is accessible
	var result int
	err = db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return errors.New("failed to query PostgreSQL: " + err.Error())
	}

	if result != 1 {
		return errors.New("unexpected result from PostgreSQL health check")
	}

	return nil
}

// GetConnectionInfo 获取连接信息（用于日志和调试）
func (o *PostgresOptions) GetConnectionInfo() map[string]interface{} {
	return map[string]interface{}{
		"host":     o.Host,
		"port":     o.Port,
		"database": o.Database,
		"user":     o.User,
		"ssl_mode": o.SSLMode,
		"schema":   o.Schema,
		"max_open": o.MaxOpenConns,
		"max_idle": o.MaxIdleConns,
	}
}

// Clone 创建配置的副本
func (o *PostgresOptions) Clone() *PostgresOptions {
	return &PostgresOptions{
		Host:                 o.Host,
		Port:                 o.Port,
		User:                 o.User,
		Password:             o.Password,
		Database:             o.Database,
		SSLMode:              o.SSLMode,
		TimeZone:             o.TimeZone,
		Schema:               o.Schema,
		MaxOpenConns:         o.MaxOpenConns,
		MaxIdleConns:         o.MaxIdleConns,
		ConnMaxLifetime:      o.ConnMaxLifetime,
		ConnMaxIdleTime:      o.ConnMaxIdleTime,
		LogLevel:             o.LogLevel,
		SlowQueryThreshold:   o.SlowQueryThreshold,
		AutoMigrate:          o.AutoMigrate,
		PreferSimpleProtocol: o.PreferSimpleProtocol,
		StatementCacheMode:   o.StatementCacheMode,
	}
}
