package config

import (
	"fmt"

	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	commoncore "github.com/kart-io/k8s-agent/common/core"
	"github.com/spf13/pflag"
)

// Options 实现 commonapp.Options 接口
type Options struct {
	ConfigFile string // 配置文件路径
	*Config           // 嵌入原有的 Config 结构
}

// NewOptions 创建配置选项
func NewOptions() *Options {
	return &Options{
		Config: &Config{
			Server: ServerConfig{
				Port:         8095,
				Mode:         "release",
				ReadTimeout:  "60s",
				WriteTimeout: "60s",
			},
			Database: DatabaseConfig{
				Host:         "localhost",
				Port:         3306,
				User:         "aetherius",
				Password:     "aetherius",
				DBName:       "aetherius_monitor",
				MaxOpenConns: 100,
				MaxIdleConns: 10,
			},
			Redis: RedisConfig{
				Host:     "localhost",
				Port:     6379,
				DB:       0,
				PoolSize: 100,
			},
			Prometheus: PrometheusConfig{
				Enabled: true,
				Port:    9095,
			},
			JWT: JWTConfig{
				Secret:     "change-this-secret",
				Expiration: "24h",
			},
			Logging: LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "stdout",
			},
		},
	}
}

// AddFlags 添加命令行标志
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.ConfigFile, "config", "c", "", "配置文件路径")
	fs.IntVar(&o.Server.Port, "port", o.Server.Port, "服务器端口")
	fs.StringVar(&o.Database.Host, "db-host", o.Database.Host, "数据库主机")
	fs.IntVar(&o.Database.Port, "db-port", o.Database.Port, "数据库端口")
}

// Complete 完成配置初始化
func (o *Options) Complete() error {
	if o.ConfigFile != "" {
		cfg, err := LoadFromPath(o.ConfigFile)
		if err != nil {
			return err
		}
		o.Config = cfg
	}
	return nil
}

// Validate 验证配置
func (o *Options) Validate() []error {
	if err := validate(o.Config); err != nil {
		return []error{err}
	}
	return nil
}

// GetHealthPort 实现 commonapp.HealthPortProvider 接口
// 简化版本：直接返回固定端口，不使用HealthOptions
func (o *Options) GetHealthPort() int {
	return 8096 // Monitor 健康检查端口
}

// 确保 Options 实现 commonapp.Options 接口
var _ commonapp.Options = (*Options)(nil)

// 确保 Options 实现 commonapp.HealthPortProvider 接口
var _ commonapp.HealthPortProvider = (*Options)(nil)

// Config holds all configuration
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Logging    LoggingConfig    `mapstructure:"logging"`
	Alert      AlertConfig      `mapstructure:"alert"`
	Metrics    MetricsConfig    `mapstructure:"metrics"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`
	ReadTimeout  string `mapstructure:"read_timeout"`
	WriteTimeout string `mapstructure:"write_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	SSLMode      string `mapstructure:"sslmode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// PrometheusConfig holds Prometheus configuration
type PrometheusConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Port    int  `mapstructure:"port"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	Expiration string `mapstructure:"expiration"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// AlertConfig holds alert configuration
type AlertConfig struct {
	CheckInterval string              `mapstructure:"check_interval"`
	Channels      AlertChannelsConfig `mapstructure:"channels"`
}

// AlertChannelsConfig holds alert channels configuration
type AlertChannelsConfig struct {
	Email   EmailAlertConfig   `mapstructure:"email"`
	Webhook WebhookAlertConfig `mapstructure:"webhook"`
	Slack   SlackAlertConfig   `mapstructure:"slack"`
}

// EmailAlertConfig holds email alert configuration
type EmailAlertConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	SMTPHost string `mapstructure:"smtp_host"`
	SMTPPort int    `mapstructure:"smtp_port"`
	From     string `mapstructure:"from"`
}

// WebhookAlertConfig holds webhook alert configuration
type WebhookAlertConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	URL     string `mapstructure:"url"`
}

// SlackAlertConfig holds Slack alert configuration
type SlackAlertConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	WebhookURL string `mapstructure:"webhook_url"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	RetentionDays       int    `mapstructure:"retention_days"`
	AggregationInterval string `mapstructure:"aggregation_interval"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	return LoadFromPath("")
}

// LoadFromPath loads configuration from a specific file path
func LoadFromPath(configPath string) (*Config, error) {
	config := &Config{}

	// 使用通用配置加载器
	wrapper := &configWrapper{Config: config}

	envBindings := map[string]string{
		"jwt.secret":        "JWT_SECRET",
		"database.password": "DB_PASSWORD",
		"redis.password":    "REDIS_PASSWORD",
	}

	if err := commoncore.LoadOptions(wrapper, configPath, envBindings); err != nil {
		return nil, err
	}

	return config, nil
}

// configWrapper 包装 Config 以实现 Options 接口
type configWrapper struct {
	*Config
}

// Complete 实现 Options 接口
func (w *configWrapper) Complete() error {
	// monitor 的 Config 不需要特殊的 Complete 逻辑
	return nil
}

// Validate 实现 Options 接口
func (w *configWrapper) Validate() []error {
	if err := validate(w.Config); err != nil {
		return []error{err}
	}
	return nil
}

// validate validates configuration
func validate(cfg *Config) error {
	if cfg.Server.Port == 0 {
		return fmt.Errorf("server.port is required")
	}
	if cfg.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if cfg.Database.DBName == "" {
		return fmt.Errorf("database.dbname is required")
	}
	if cfg.Redis.Host == "" {
		return fmt.Errorf("redis.host is required")
	}
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret is required")
	}
	return nil
}
