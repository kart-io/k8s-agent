package options

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

// DatabaseOptions 数据库配置
type DatabaseOptions struct {
	Host            string        `mapstructure:"host" yaml:"host" json:"host"`
	Port            int           `mapstructure:"port" yaml:"port" json:"port"`
	User            string        `mapstructure:"user" yaml:"user" json:"user"`
	Password        string        `mapstructure:"password" yaml:"password" json:"password"`
	Database        string        `mapstructure:"database" yaml:"database" json:"database"`
	SSLMode         string        `mapstructure:"ssl_mode" yaml:"ssl_mode" json:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns" yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
	AutoMigrate     bool          `mapstructure:"auto_migrate" yaml:"auto_migrate" json:"auto_migrate"`
}

// NewDatabaseOptions 创建默认的数据库配置
func NewDatabaseOptions() *DatabaseOptions {
	return &DatabaseOptions{
		Host:            "localhost",
		Port:            3306,
		User:            "root",
		Password:        "",
		Database:        "test",
		SSLMode:         "disable",
		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: 1 * time.Hour,
		AutoMigrate:     false,
	}
}

// Validate 验证配置
func (o *DatabaseOptions) Validate() error {
	if o.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if o.Port < 1 || o.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", o.Port)
	}
	if o.User == "" {
		return fmt.Errorf("database user is required")
	}
	if o.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if o.MaxOpenConns < 0 {
		return fmt.Errorf("max_open_conns must be >= 0")
	}
	if o.MaxIdleConns < 0 {
		return fmt.Errorf("max_idle_conns must be >= 0")
	}
	return nil
}

// DSN 返回数据库连接字符串
func (o *DatabaseOptions) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		o.User, o.Password, o.Host, o.Port, o.Database)
}

// AddFlags 添加命令行参数
func (o *DatabaseOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Host, "db.host", o.Host, "Database host address")
	fs.IntVar(&o.Port, "db.port", o.Port, "Database port")
	fs.StringVar(&o.User, "db.user", o.User, "Database user")
	fs.StringVar(&o.Password, "db.password", o.Password, "Database password")
	fs.StringVar(&o.Database, "db.database", o.Database, "Database name")
	fs.StringVar(&o.SSLMode, "db.ssl-mode", o.SSLMode, "Database SSL mode")
	fs.IntVar(&o.MaxOpenConns, "db.max-open-conns", o.MaxOpenConns, "Maximum number of open connections")
	fs.IntVar(&o.MaxIdleConns, "db.max-idle-conns", o.MaxIdleConns, "Maximum number of idle connections")
	fs.DurationVar(&o.ConnMaxLifetime, "db.conn-max-lifetime", o.ConnMaxLifetime, "Maximum connection lifetime")
	fs.BoolVar(&o.AutoMigrate, "db.auto-migrate", o.AutoMigrate, "Enable automatic database migration")
}

// ApplyTo 将配置应用到目标接口
// 接受一个函数切片指针，将配置转换为 db.MySQLOption 函数选项
func (o *DatabaseOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	// 类型断言，检查是否为函数选项切片指针
	switch v := target.(type) {
	case *[]interface{}:
		// 将配置转换为通用选项
		*v = append(*v,
			map[string]interface{}{
				"host":            o.Host,
				"port":            o.Port,
				"user":            o.User,
				"password":        o.Password,
				"database":        o.Database,
				"maxOpenConns":    o.MaxOpenConns,
				"maxIdleConns":    o.MaxIdleConns,
				"connMaxLifetime": o.ConnMaxLifetime,
			},
		)
	}

	return nil
}

// Complete 完成配置初始化
// 设置默认值和计算派生值
func (o *DatabaseOptions) Complete() error {
	// 如果 Host 为空，设置默认值
	if o.Host == "" {
		o.Host = "localhost"
	}

	// 确保端口在有效范围内
	if o.Port <= 0 {
		o.Port = 3306
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

	// 确保连接最大生命周期有合理的值
	if o.ConnMaxLifetime <= 0 {
		o.ConnMaxLifetime = 1 * time.Hour
	}

	return nil
}

// WithDBHost 设置数据库地址
func WithDBHost(host string) func(*DatabaseOptions) {
	return func(o *DatabaseOptions) {
		o.Host = host
	}
}

// WithDBPort 设置数据库端口
func WithDBPort(port int) func(*DatabaseOptions) {
	return func(o *DatabaseOptions) {
		o.Port = port
	}
}

// WithDBUser 设置数据库用户名
func WithDBUser(user string) func(*DatabaseOptions) {
	return func(o *DatabaseOptions) {
		o.User = user
	}
}

// WithDBPassword 设置数据库密码
func WithDBPassword(password string) func(*DatabaseOptions) {
	return func(o *DatabaseOptions) {
		o.Password = password
	}
}

// WithDBName 设置数据库名称
func WithDBName(database string) func(*DatabaseOptions) {
	return func(o *DatabaseOptions) {
		o.Database = database
	}
}

// WithDBSSLMode 设置SSL模式
func WithDBSSLMode(sslMode string) func(*DatabaseOptions) {
	return func(o *DatabaseOptions) {
		o.SSLMode = sslMode
	}
}

// WithDBMaxOpenConns 设置最大打开连接数
func WithDBMaxOpenConns(max int) func(*DatabaseOptions) {
	return func(o *DatabaseOptions) {
		o.MaxOpenConns = max
	}
}

// WithDBMaxIdleConns 设置最大空闲连接数
func WithDBMaxIdleConns(max int) func(*DatabaseOptions) {
	return func(o *DatabaseOptions) {
		o.MaxIdleConns = max
	}
}

// WithDBConnMaxLifetime 设置连接最大生命周期
func WithDBConnMaxLifetime(lifetime time.Duration) func(*DatabaseOptions) {
	return func(o *DatabaseOptions) {
		o.ConnMaxLifetime = lifetime
	}
}

// WithDBAutoMigrate 设置是否自动迁移
func WithDBAutoMigrate(autoMigrate bool) func(*DatabaseOptions) {
	return func(o *DatabaseOptions) {
		o.AutoMigrate = autoMigrate
	}
}
