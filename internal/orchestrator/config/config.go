package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents orchestrator configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	NATS     NATSConfig     `mapstructure:"nats"`
	Temporal TemporalConfig `mapstructure:"temporal"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	AI       AIConfig       `mapstructure:"ai"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	HealthPort   int           `mapstructure:"health_port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	GracefulStop time.Duration `mapstructure:"graceful_stop"`
}

// NATSConfig represents NATS configuration
type NATSConfig struct {
	URL           string        `mapstructure:"url"`
	MaxReconnect  int           `mapstructure:"max_reconnect"`
	ReconnectWait time.Duration `mapstructure:"reconnect_wait"`
}

// TemporalConfig represents Temporal workflow engine configuration
type TemporalConfig struct {
	HostPort  string `mapstructure:"host_port"`
	Namespace string `mapstructure:"namespace"`
	TaskQueue string `mapstructure:"task_queue"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Addr         string        `mapstructure:"addr"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
}

// AIConfig represents AI service configuration
type AIConfig struct {
	ReasoningServiceURL string        `mapstructure:"reasoning_service_url"`
	AgentManagerURL     string        `mapstructure:"agent_manager_url"`
	Timeout             time.Duration `mapstructure:"timeout"`
	MaxRetries          int           `mapstructure:"max_retries"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	OutputPath string `mapstructure:"output_path"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	return LoadFromPath("")
}

// LoadFromPath loads configuration from a specific file path
func LoadFromPath(configPath string) (*Config, error) {
	v := viper.New()

	if configPath != "" {
		// Use specified config file path
		v.SetConfigFile(configPath)
	} else {
		// Use default config file search
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	// Read configuration file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Allow environment variable overrides
	v.AutomaticEnv()

	// Bind specific environment variables with custom names
	v.BindEnv("database.host", "DB_HOST")
	v.BindEnv("database.port", "DB_PORT")
	v.BindEnv("database.user", "DB_USER")
	v.BindEnv("database.password", "DB_PASSWORD")
	v.BindEnv("database.database", "DB_DATABASE")
	v.BindEnv("nats.url", "NATS_URL")
	v.BindEnv("ai.reasoning_service_url", "AI_SERVICE_URL")
	v.BindEnv("ai.agent_manager_url", "AGENT_MANAGER_URL")
	v.BindEnv("redis.addr", "REDIS_ADDR")
	v.BindEnv("redis.password", "REDIS_PASSWORD")

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// validate validates configuration
func validate(cfg *Config) error {
	if cfg.Server.Port == 0 {
		return fmt.Errorf("server.port is required")
	}
	if cfg.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if cfg.Database.Database == "" {
		return fmt.Errorf("database.database is required")
	}
	if cfg.Redis.Addr == "" {
		return fmt.Errorf("redis.addr is required")
	}
	if cfg.NATS.URL == "" {
		return fmt.Errorf("nats.url is required")
	}
	if cfg.AI.AgentManagerURL == "" {
		return fmt.Errorf("ai.agent_manager_url is required")
	}
	if cfg.AI.ReasoningServiceURL == "" {
		return fmt.Errorf("ai.reasoning_service_url is required")
	}
	return nil
}
