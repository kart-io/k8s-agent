package options

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"github.com/kart-io/k8s-agent/common/options/validation"
)

// DatabaseOptions 数据库配置
type DatabaseOptions struct {
	Host            string        `mapstructure:"host" yaml:"host" json:"host"`
	Port            int           `mapstructure:"port" yaml:"port" json:"port"`
	User            string        `mapstructure:"user" yaml:"user" json:"user"`
	Password        string        `mapstructure:"password" yaml:"password" json:"-"` // 密码不序列化到 JSON
	Database        string        `mapstructure:"database" yaml:"database" json:"database"`
	SSLMode         string        `mapstructure:"ssl_mode" yaml:"ssl_mode" json:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns" yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
	LogLevel        string        `mapstructure:"log_level" yaml:"log_level" json:"log_level"` // silent, error, warn, info
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
		LogLevel:        "silent", // 默认静默模式
		AutoMigrate:     false,
	}
}

// Validate 验证配置
func (o *DatabaseOptions) Validate() error {
	// 使用通用验证器
	if err := validation.ValidateRequired(o.Host, "database host"); err != nil {
		return err
	}

	if err := validation.ValidatePort(o.Port, "database"); err != nil {
		return err
	}

	if err := validation.ValidateRequired(o.User, "database user"); err != nil {
		return err
	}

	if err := validation.ValidateRequired(o.Database, "database name"); err != nil {
		return err
	}

	if err := validation.ValidateConnectionPool(o.MaxOpenConns, o.MaxIdleConns, "database"); err != nil {
		return err
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
	fs.StringVar(&o.LogLevel, "db.log-level", o.LogLevel, "GORM log level (silent, error, warn, info)")
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

// NewDB 创建 MySQL 数据库连接
// 参考 OneX 项目设计，提供便捷的数据库初始化方法
// 需要传入 logger 实例用于日志记录
//
// 使用示例:
//
//	import (
//	    "github.com/kart-io/k8s-agent/common/db"
//	    "github.com/kart-io/logger"
//	)
//
//	log := logger.New(logger.Config{...})
//	dbClient, err := dbOpts.NewDB(log)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer dbClient.Close()
func (o *DatabaseOptions) NewDB(log interface{}) (interface{}, error) {
	// 导入 db 包会产生循环依赖，这里返回错误提示用户直接使用 db.NewMySQL()
	// 这个方法主要是为了与 OneX 风格保持一致，实际使用时推荐:
	//   import "github.com/kart-io/k8s-agent/common/db"
	//   dbClient, err := db.NewMySQL(log,
	//       db.WithHost(opts.Host),
	//       db.WithPort(opts.Port),
	//       db.WithUser(opts.User),
	//       db.WithPassword(opts.Password),
	//       db.WithDatabase(opts.Database),
	//       db.WithMaxOpenConns(opts.MaxOpenConns),
	//       db.WithMaxIdleConns(opts.MaxIdleConns),
	//       db.WithConnMaxLifetime(opts.ConnMaxLifetime),
	//       db.WithLogLevel(opts.LogLevel),
	//   )
	return nil, fmt.Errorf("NewDB() method is a reference implementation. " +
		"Please use db.NewMySQL() or db.NewMySQLFromDSN() directly to avoid circular dependencies. " +
		"Example: db.NewMySQL(log, db.WithHost(opts.Host), db.WithPort(opts.Port), ...)")
}
