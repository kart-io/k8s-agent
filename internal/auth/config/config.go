package config

import commoncore "github.com/kart-io/k8s-agent/common/core"

// Config is an alias for Options for backward compatibility
// Deprecated: Use Options instead.
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
	config := NewOptions()

	// 使用通用配置加载器
	envBindings := map[string]string{
		"jwt.secret":        "JWT_SECRET",
		"jwt.expires_hours": "JWT_EXPIRES_HOURS",
	}

	if err := commoncore.LoadOptions(config, configPath, envBindings); err != nil {
		return nil, err
	}

	return config, nil
}
