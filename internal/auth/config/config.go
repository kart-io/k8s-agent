package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config is an alias for Options for backward compatibility
// Deprecated: Use Options instead
type Config = Options

// Load loads configuration from file and environment variables
// Deprecated: This function is kept for backward compatibility with old code.
// New code should use pkg/app framework which handles config loading automatically.
func Load() (*Config, error) {
	return LoadFromPath("")
}

// LoadFromPath loads configuration from a specific file path
// Deprecated: This function is kept for backward compatibility with old code.
// New code should use pkg/app framework which handles config loading automatically.
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

	config := NewOptions()
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if errs := config.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("config validation failed: %v", errs)
	}

	return config, nil
}

// validate validates configuration
// Deprecated: Use Options.Validate() instead
func validate(cfg *Config) error {
	errs := cfg.Validate()
	if len(errs) > 0 {
		return fmt.Errorf("validation errors: %v", errs)
	}
	return nil
}
