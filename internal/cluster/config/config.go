package config

import (
	"fmt"

	commoncore "github.com/kart-io/k8s-agent/common/core"
)

// Config holds all configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"` // debug, release, test
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

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // json, text
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	return LoadFromPath("")
}

// LoadFromPath loads configuration from a specific file path
func LoadFromPath(configPath string) (*Config, error) {
	config := &Config{}

	// 使用通用配置加载器
	// 注意：Config 结构体需要实现 Complete() 和 Validate() 方法
	// 这里为了兼容，使用一个包装类型
	wrapper := &configWrapper{Config: config}

	envBindings := map[string]string{
		"jwt.secret":        "JWT_SECRET",
		"database.host":     "DB_HOST",
		"database.port":     "DB_PORT",
		"database.user":     "DB_USER",
		"database.password": "DB_PASSWORD",
		"database.dbname":   "DB_NAME",
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
	// cluster 服务的 Config 不需要特殊的 Complete 逻辑
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
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret is required")
	}
	return nil
}
