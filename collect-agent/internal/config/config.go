package config

import (
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/kart/k8s-agent/collect-agent/internal/types"
)

// LoadConfig loads configuration from a file path or creates default config
func LoadConfig(configPath string) (*types.AgentConfig, error) {
	if configPath == "" {
		return types.DefaultConfig(), nil
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file %s: %w", configPath, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := types.DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables if present
	overrideWithEnv(config)

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// overrideWithEnv overrides configuration values with environment variables
func overrideWithEnv(config *types.AgentConfig) {
	if val := os.Getenv("CLUSTER_ID"); val != "" {
		config.ClusterID = val
	}

	if val := os.Getenv("CENTRAL_ENDPOINT"); val != "" {
		config.CentralEndpoint = val
	}

	if val := os.Getenv("LOG_LEVEL"); val != "" {
		config.LogLevel = val
	}

	if val := os.Getenv("RECONNECT_DELAY"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.ReconnectDelay = duration
		}
	}

	if val := os.Getenv("HEARTBEAT_INTERVAL"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.HeartbeatInterval = duration
		}
	}

	if val := os.Getenv("METRICS_INTERVAL"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.MetricsInterval = duration
		}
	}

	if val := os.Getenv("ENABLE_METRICS"); val != "" {
		config.EnableMetrics = val == "true" || val == "1"
	}

	if val := os.Getenv("ENABLE_EVENTS"); val != "" {
		config.EnableEvents = val == "true" || val == "1"
	}
}

// validateConfig validates the configuration values
func validateConfig(config *types.AgentConfig) error {
	if config.CentralEndpoint == "" {
		return fmt.Errorf("central_endpoint is required")
	}

	if config.ReconnectDelay < time.Second {
		return fmt.Errorf("reconnect_delay must be at least 1 second")
	}

	if config.HeartbeatInterval < 10*time.Second {
		return fmt.Errorf("heartbeat_interval must be at least 10 seconds")
	}

	if config.MetricsInterval < 30*time.Second {
		return fmt.Errorf("metrics_interval must be at least 30 seconds")
	}

	if config.BufferSize < 10 {
		return fmt.Errorf("buffer_size must be at least 10")
	}

	if config.MaxRetries < 1 {
		return fmt.Errorf("max_retries must be at least 1")
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}

	if !validLogLevels[config.LogLevel] {
		return fmt.Errorf("invalid log_level: %s, must be one of: debug, info, warn, error, fatal", config.LogLevel)
	}

	return nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(config *types.AgentConfig, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDefaultConfigYAML returns the default configuration as YAML string
func GetDefaultConfigYAML() string {
	config := types.DefaultConfig()
	data, _ := yaml.Marshal(config)
	return string(data)
}