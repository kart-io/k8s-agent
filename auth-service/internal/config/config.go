package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release, test
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
	AutoMigrate  bool   `mapstructure:"auto_migrate"` // Enable GORM AutoMigrate (dev/test only)
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

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Engine string     `mapstructure:"engine"` // zap, slog
	Level  string     `mapstructure:"level"`  // debug, info, warn, error
	Format string     `mapstructure:"format"` // json, text
	Output string     `mapstructure:"output"` // stdout, stderr, file path
	OTLP   OTLPConfig `mapstructure:"otlp"`   // OTLP configuration
}

// OTLPConfig holds OTLP exporter configuration
type OTLPConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"` // e.g., http://localhost:4317
	Insecure bool   `mapstructure:"insecure"` // Use insecure connection
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	v := viper.New()

	// Set configuration file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")

	// Read configuration file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Allow environment variable overrides
	v.AutomaticEnv()

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
	if cfg.Database.DBName == "" {
		return fmt.Errorf("database.dbname is required")
	}
	if cfg.Redis.Host == "" {
		return fmt.Errorf("redis.host is required")
	}
	if cfg.JWT.Secret == "" || len(cfg.JWT.Secret) < 32 {
		return fmt.Errorf("jwt.secret must be at least 32 characters")
	}
	return nil
}
