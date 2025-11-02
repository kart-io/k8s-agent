package config

import (
	"fmt"
	"time"

	commoncore "github.com/kart-io/k8s-agent/common/core"
)

// Config holds all configuration
type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	Log         LogConfig         `mapstructure:"log"`
	Redis       RedisConfig       `mapstructure:"redis"`
	JWT         JWTConfig         `mapstructure:"jwt"`
	RateLimit   RateLimitConfig   `mapstructure:"rate_limit"`
	CORS        CORSConfig        `mapstructure:"cors"`
	Services    ServicesConfig    `mapstructure:"services"`
	Routes      []RouteConfig     `mapstructure:"routes"`
	HealthCheck HealthCheckConfig `mapstructure:"health_check"`
	Metrics     MetricsConfig     `mapstructure:"metrics"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"` // debug, release, test
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret       string `mapstructure:"secret"`
	ExpiresHours int    `mapstructure:"expires_hours"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerSecond int  `mapstructure:"requests_per_second"`
	Burst             int  `mapstructure:"burst"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Enabled          bool          `mapstructure:"enabled"`
	AllowOrigins     []string      `mapstructure:"allow_origins"`
	AllowMethods     []string      `mapstructure:"allow_methods"`
	AllowHeaders     []string      `mapstructure:"allow_headers"`
	ExposeHeaders    []string      `mapstructure:"expose_headers"`
	AllowCredentials bool          `mapstructure:"allow_credentials"`
	MaxAge           time.Duration `mapstructure:"max_age"`
}

// ServiceConfig holds individual service configuration
type ServiceConfig struct {
	Name        string        `mapstructure:"name"`
	URL         string        `mapstructure:"url"`
	Timeout     time.Duration `mapstructure:"timeout"`
	Retry       int           `mapstructure:"retry"`
	HealthCheck string        `mapstructure:"health_check"`
}

// ServicesConfig holds all backend services configuration
type ServicesConfig struct {
	Auth         ServiceConfig `mapstructure:"auth"`
	AgentManager ServiceConfig `mapstructure:"agent_manager"`
	Orchestrator ServiceConfig `mapstructure:"orchestrator"`
	Reasoning    ServiceConfig `mapstructure:"reasoning"`
}

// RouteConfig holds route configuration
type RouteConfig struct {
	Path         string `mapstructure:"path"`
	Service      string `mapstructure:"service"`
	StripPrefix  bool   `mapstructure:"strip_prefix"`
	AuthRequired bool   `mapstructure:"auth_required"`
}

// HealthCheckConfig holds health check configuration
type HealthCheckConfig struct {
	Enabled  bool          `mapstructure:"enabled"`
	Interval time.Duration `mapstructure:"interval"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
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
		"jwt.expires_hours": "JWT_EXPIRES_HOURS",
		"server.port":       "GATEWAY_PORT",
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
	// gateway 的 Config 不需要特殊的 Complete 逻辑
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
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret is required")
	}
	if cfg.Redis.Host == "" {
		return fmt.Errorf("redis.host is required")
	}
	return nil
}
