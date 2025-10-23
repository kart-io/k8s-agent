package config

import (
	"fmt"

	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/spf13/viper"
)

// Config holds all configuration
type Config struct {
	Server   commonoptions.ServerOptions   `mapstructure:"server"`
	Database commonoptions.DatabaseOptions `mapstructure:"database"`
	Redis    commonoptions.RedisOptions    `mapstructure:"redis"`
	JWT      commonoptions.JWTOptions      `mapstructure:"jwt"`
	Logging  commonoptions.LoggingOptions  `mapstructure:"logging"`
	Email    commonoptions.EmailOptions    `mapstructure:"email"`
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
	v.BindEnv("jwt.secret", "JWT_SECRET")
	v.BindEnv("jwt.expires_hours", "JWT_EXPIRES_HOURS")

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
	if err := cfg.Server.Validate(); err != nil {
		return fmt.Errorf("server validation failed: %w", err)
	}
	if err := cfg.Database.Validate(); err != nil {
		return fmt.Errorf("database validation failed: %w", err)
	}
	if err := cfg.Redis.Validate(); err != nil {
		return fmt.Errorf("redis validation failed: %w", err)
	}
	if err := cfg.JWT.Validate(); err != nil {
		return fmt.Errorf("jwt validation failed: %w", err)
	}
	if err := cfg.Logging.Validate(); err != nil {
		return fmt.Errorf("logging validation failed: %w", err)
	}
	if err := cfg.Email.Validate(); err != nil {
		return fmt.Errorf("email validation failed: %w", err)
	}
	return nil
}
