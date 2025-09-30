package config

import (
	"os"
	"testing"
	"time"

	"github.com/kart/k8s-agent/collect-agent/internal/types"
)

func TestLoadConfig_DefaultConfig(t *testing.T) {
	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if config == nil {
		t.Fatal("LoadConfig returned nil config")
	}

	// Check default values
	if config.BufferSize != 1000 {
		t.Errorf("BufferSize = %v, want %v", config.BufferSize, 1000)
	}
}

func TestValidateConfig_Valid(t *testing.T) {
	config := &types.AgentConfig{
		CentralEndpoint:   "nats://localhost:4222",
		ReconnectDelay:    5 * time.Second,
		HeartbeatInterval: 30 * time.Second,
		MetricsInterval:   60 * time.Second,
		BufferSize:        100,
		MaxRetries:        10,
		LogLevel:          "info",
	}

	err := validateConfig(config)
	if err != nil {
		t.Errorf("validateConfig failed for valid config: %v", err)
	}
}

func TestValidateConfig_MissingEndpoint(t *testing.T) {
	config := &types.AgentConfig{
		CentralEndpoint: "",
		ReconnectDelay:  5 * time.Second,
	}

	err := validateConfig(config)
	if err == nil {
		t.Error("validateConfig should fail for missing central_endpoint")
	}
}

func TestValidateConfig_InvalidReconnectDelay(t *testing.T) {
	config := &types.AgentConfig{
		CentralEndpoint: "nats://localhost:4222",
		ReconnectDelay:  500 * time.Millisecond, // Less than 1 second
		HeartbeatInterval: 30 * time.Second,
		MetricsInterval:   60 * time.Second,
		BufferSize:        100,
		MaxRetries:        10,
		LogLevel:          "info",
	}

	err := validateConfig(config)
	if err == nil {
		t.Error("validateConfig should fail for reconnect_delay < 1s")
	}
}

func TestValidateConfig_InvalidHeartbeatInterval(t *testing.T) {
	config := &types.AgentConfig{
		CentralEndpoint:   "nats://localhost:4222",
		ReconnectDelay:    5 * time.Second,
		HeartbeatInterval: 5 * time.Second, // Less than 10 seconds
		MetricsInterval:   60 * time.Second,
		BufferSize:        100,
		MaxRetries:        10,
		LogLevel:          "info",
	}

	err := validateConfig(config)
	if err == nil {
		t.Error("validateConfig should fail for heartbeat_interval < 10s")
	}
}

func TestValidateConfig_InvalidMetricsInterval(t *testing.T) {
	config := &types.AgentConfig{
		CentralEndpoint:   "nats://localhost:4222",
		ReconnectDelay:    5 * time.Second,
		HeartbeatInterval: 30 * time.Second,
		MetricsInterval:   20 * time.Second, // Less than 30 seconds
		BufferSize:        100,
		MaxRetries:        10,
		LogLevel:          "info",
	}

	err := validateConfig(config)
	if err == nil {
		t.Error("validateConfig should fail for metrics_interval < 30s")
	}
}

func TestValidateConfig_InvalidBufferSize(t *testing.T) {
	config := &types.AgentConfig{
		CentralEndpoint:   "nats://localhost:4222",
		ReconnectDelay:    5 * time.Second,
		HeartbeatInterval: 30 * time.Second,
		MetricsInterval:   60 * time.Second,
		BufferSize:        5, // Less than 10
		MaxRetries:        10,
		LogLevel:          "info",
	}

	err := validateConfig(config)
	if err == nil {
		t.Error("validateConfig should fail for buffer_size < 10")
	}
}

func TestValidateConfig_InvalidMaxRetries(t *testing.T) {
	config := &types.AgentConfig{
		CentralEndpoint:   "nats://localhost:4222",
		ReconnectDelay:    5 * time.Second,
		HeartbeatInterval: 30 * time.Second,
		MetricsInterval:   60 * time.Second,
		BufferSize:        100,
		MaxRetries:        0, // Less than 1
		LogLevel:          "info",
	}

	err := validateConfig(config)
	if err == nil {
		t.Error("validateConfig should fail for max_retries < 1")
	}
}

func TestValidateConfig_InvalidLogLevel(t *testing.T) {
	config := &types.AgentConfig{
		CentralEndpoint:   "nats://localhost:4222",
		ReconnectDelay:    5 * time.Second,
		HeartbeatInterval: 30 * time.Second,
		MetricsInterval:   60 * time.Second,
		BufferSize:        100,
		MaxRetries:        10,
		LogLevel:          "invalid",
	}

	err := validateConfig(config)
	if err == nil {
		t.Error("validateConfig should fail for invalid log_level")
	}
}

func TestOverrideWithEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("CLUSTER_ID", "env-cluster")
	os.Setenv("CENTRAL_ENDPOINT", "nats://env:4222")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("RECONNECT_DELAY", "10s")
	os.Setenv("ENABLE_METRICS", "false")
	defer func() {
		os.Unsetenv("CLUSTER_ID")
		os.Unsetenv("CENTRAL_ENDPOINT")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("RECONNECT_DELAY")
		os.Unsetenv("ENABLE_METRICS")
	}()

	config := types.DefaultConfig()
	overrideWithEnv(config)

	if config.ClusterID != "env-cluster" {
		t.Errorf("ClusterID = %v, want %v", config.ClusterID, "env-cluster")
	}

	if config.CentralEndpoint != "nats://env:4222" {
		t.Errorf("CentralEndpoint = %v, want %v", config.CentralEndpoint, "nats://env:4222")
	}

	if config.LogLevel != "debug" {
		t.Errorf("LogLevel = %v, want %v", config.LogLevel, "debug")
	}

	if config.ReconnectDelay != 10*time.Second {
		t.Errorf("ReconnectDelay = %v, want %v", config.ReconnectDelay, 10*time.Second)
	}

	if config.EnableMetrics != false {
		t.Errorf("EnableMetrics = %v, want %v", config.EnableMetrics, false)
	}
}

func TestGetDefaultConfigYAML(t *testing.T) {
	yaml := GetDefaultConfigYAML()

	if yaml == "" {
		t.Error("GetDefaultConfigYAML returned empty string")
	}

	// Check if YAML contains expected keys
	expectedKeys := []string{
		"central_endpoint",
		"reconnect_delay",
		"heartbeat_interval",
		"metrics_interval",
		"buffer_size",
		"max_retries",
		"log_level",
	}

	for _, key := range expectedKeys {
		if !contains(yaml, key) {
			t.Errorf("YAML missing key: %s", key)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && s != substr && (len(s) >= len(substr)) &&
		(s[:len(substr)] == substr || contains(s[1:], substr))
}