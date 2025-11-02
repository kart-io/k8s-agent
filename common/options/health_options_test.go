package options

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHealthOptions(t *testing.T) {
	opts := NewHealthOptions()

	assert.NotNil(t, opts)
	assert.True(t, opts.Enable)
	assert.Equal(t, "0.0.0.0", opts.Host)
	assert.Equal(t, 20250, opts.Port)
	assert.Equal(t, "/healthz", opts.Path)
	assert.Equal(t, "/readyz", opts.ReadinessPath)
	assert.Equal(t, "/livez", opts.LivenessPath)
	assert.False(t, opts.EnablePprof)
	assert.Equal(t, 5*time.Second, opts.ShutdownTimeout)
	assert.Equal(t, 10*time.Second, opts.HealthCheckInterval)
}

func TestHealthOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    *HealthOptions
		wantErr bool
	}{
		{
			name:    "valid default options",
			opts:    NewHealthOptions(),
			wantErr: false,
		},
		{
			name: "disabled health check - skip validation",
			opts: &HealthOptions{
				Enable: false,
			},
			wantErr: false,
		},
		{
			name: "invalid port - too low",
			opts: &HealthOptions{
				Enable: true,
				Host:   "localhost",
				Port:   0,
				Path:   "/health",
			},
			wantErr: true,
		},
		{
			name: "invalid port - too high",
			opts: &HealthOptions{
				Enable: true,
				Host:   "localhost",
				Port:   70000,
				Path:   "/health",
			},
			wantErr: true,
		},
		{
			name: "empty host",
			opts: &HealthOptions{
				Enable: true,
				Host:   "",
				Port:   8080,
				Path:   "/health",
			},
			wantErr: true,
		},
		{
			name: "empty path",
			opts: &HealthOptions{
				Enable:        true,
				Host:          "localhost",
				Port:          8080,
				Path:          "",
				ReadinessPath: "/ready",
				LivenessPath:  "/live",
			},
			wantErr: true,
		},
		{
			name: "empty readiness path",
			opts: &HealthOptions{
				Enable:       true,
				Host:         "localhost",
				Port:         8080,
				Path:         "/health",
				LivenessPath: "/live",
			},
			wantErr: true,
		},
		{
			name: "negative shutdown timeout",
			opts: &HealthOptions{
				Enable:              true,
				Host:                "localhost",
				Port:                8080,
				Path:                "/health",
				ReadinessPath:       "/ready",
				LivenessPath:        "/live",
				ShutdownTimeout:     -1 * time.Second,
				HealthCheckInterval: 10 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHealthOptions_Complete(t *testing.T) {
	tests := []struct {
		name     string
		opts     *HealthOptions
		expected *HealthOptions
	}{
		{
			name: "complete with empty host",
			opts: &HealthOptions{
				Enable: true,
				Host:   "",
				Port:   8080,
			},
			expected: &HealthOptions{
				Enable:              true,
				Host:                "0.0.0.0",
				Port:                8080,
				Path:                "/healthz",
				ReadinessPath:       "/readyz",
				LivenessPath:        "/livez",
				ShutdownTimeout:     5 * time.Second,
				HealthCheckInterval: 10 * time.Second,
			},
		},
		{
			name: "complete with invalid port",
			opts: &HealthOptions{
				Enable: true,
				Host:   "localhost",
				Port:   0,
			},
			expected: &HealthOptions{
				Enable:              true,
				Host:                "localhost",
				Port:                20250,
				Path:                "/healthz",
				ReadinessPath:       "/readyz",
				LivenessPath:        "/livez",
				ShutdownTimeout:     5 * time.Second,
				HealthCheckInterval: 10 * time.Second,
			},
		},
		{
			name: "complete with path without leading slash",
			opts: &HealthOptions{
				Enable:        true,
				Host:          "localhost",
				Port:          8080,
				Path:          "health",
				ReadinessPath: "ready",
				LivenessPath:  "live",
			},
			expected: &HealthOptions{
				Enable:              true,
				Host:                "localhost",
				Port:                8080,
				Path:                "/health",
				ReadinessPath:       "/ready",
				LivenessPath:        "/live",
				ShutdownTimeout:     5 * time.Second,
				HealthCheckInterval: 10 * time.Second,
			},
		},
		{
			name: "disabled - skip completion",
			opts: &HealthOptions{
				Enable: false,
			},
			expected: &HealthOptions{
				Enable: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Complete()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.opts)
		})
	}
}

func TestHealthOptions_WithFunctions(t *testing.T) {
	opts := NewHealthOptions()

	// Test WithHealthEnable
	WithHealthEnable(false)(opts)
	assert.False(t, opts.Enable)

	// Test WithHealthHost
	WithHealthHost("127.0.0.1")(opts)
	assert.Equal(t, "127.0.0.1", opts.Host)

	// Test WithHealthPort
	WithHealthPort(9090)(opts)
	assert.Equal(t, 9090, opts.Port)

	// Test WithHealthPath
	WithHealthPath("/status")(opts)
	assert.Equal(t, "/status", opts.Path)

	// Test WithReadinessPath
	WithReadinessPath("/ready")(opts)
	assert.Equal(t, "/ready", opts.ReadinessPath)

	// Test WithLivenessPath
	WithLivenessPath("/alive")(opts)
	assert.Equal(t, "/alive", opts.LivenessPath)

	// Test WithPprofEnabled
	WithPprofEnabled(true)(opts)
	assert.True(t, opts.EnablePprof)

	// Test WithShutdownTimeout
	WithShutdownTimeout(30 * time.Second)(opts)
	assert.Equal(t, 30*time.Second, opts.ShutdownTimeout)

	// Test WithHealthCheckInterval
	WithHealthCheckInterval(15 * time.Second)(opts)
	assert.Equal(t, 15*time.Second, opts.HealthCheckInterval)
}

func TestHealthOptions_ApplyTo(t *testing.T) {
	opts := &HealthOptions{
		Enable:              true,
		Host:                "localhost",
		Port:                8080,
		Path:                "/health",
		ReadinessPath:       "/ready",
		LivenessPath:        "/live",
		EnablePprof:         true,
		ShutdownTimeout:     5 * time.Second,
		HealthCheckInterval: 10 * time.Second,
	}

	var target []interface{}
	err := opts.ApplyTo(&target)

	require.NoError(t, err)
	require.Len(t, target, 1)

	config, ok := target[0].(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, true, config["enable"])
	assert.Equal(t, "localhost", config["host"])
	assert.Equal(t, 8080, config["port"])
	assert.Equal(t, "/health", config["path"])
	assert.Equal(t, "/ready", config["readinessPath"])
	assert.Equal(t, "/live", config["livenessPath"])
	assert.Equal(t, true, config["enablePprof"])
	assert.Equal(t, 5*time.Second, config["shutdownTimeout"])
	assert.Equal(t, 10*time.Second, config["healthCheckInterval"])
}

func TestHealthOptions_ApplyTo_Nil(t *testing.T) {
	opts := NewHealthOptions()
	err := opts.ApplyTo(nil)
	assert.NoError(t, err)
}
