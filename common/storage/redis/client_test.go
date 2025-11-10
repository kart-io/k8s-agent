package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger"
	"github.com/kart-io/logger/core"
)

// mockLogger creates a mock logger for testing
func mockLogger() core.Logger {
	log, err := logger.NewWithDefaults()
	if err != nil {
		panic(err)
	}
	return log
}

func TestNewConfig(t *testing.T) {
	redisOpts := options.NewRedisOptions()

	assert.Equal(t, "localhost:6379", redisOpts.Addr)
	assert.Equal(t, "", redisOpts.Password)
	assert.Equal(t, 0, redisOpts.DB)
	assert.Equal(t, 10, redisOpts.PoolSize)
	assert.Equal(t, 5, redisOpts.MinIdleConns)
	assert.Equal(t, 5*time.Second, redisOpts.DialTimeout)
	assert.Equal(t, 3*time.Second, redisOpts.ReadTimeout)
	assert.Equal(t, 3*time.Second, redisOpts.WriteTimeout)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *options.RedisOptions
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &options.RedisOptions{
				Addr:         "localhost:6379",
				DB:           0,
				PoolSize:     10,
				MinIdleConns: 5,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "empty address",
			config: &options.RedisOptions{
				Addr: "",
			},
			wantErr: true,
			errMsg:  "redis address is required",
		},
		{
			name: "invalid DB index negative",
			config: &options.RedisOptions{
				Addr: "localhost:6379",
				DB:   -1,
			},
			wantErr: true,
			errMsg:  "database index must be between 0 and 15",
		},
		{
			name: "invalid DB index too high",
			config: &options.RedisOptions{
				Addr: "localhost:6379",
				DB:   16,
			},
			wantErr: true,
			errMsg:  "database index must be between 0 and 15",
		},
		{
			name: "invalid pool size",
			config: &options.RedisOptions{
				Addr:     "localhost:6379",
				DB:       0,
				PoolSize: 0,
			},
			wantErr: true,
			errMsg:  "redis pool_size must be > 0",
		},
		{
			name: "negative min idle conns",
			config: &options.RedisOptions{
				Addr:         "localhost:6379",
				DB:           0,
				PoolSize:     10,
				MinIdleConns: -1,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
			},
			wantErr: true,
			errMsg:  "redis min_idle_conns must be >= 0",
		},
		{
			name: "invalid dial timeout",
			config: &options.RedisOptions{
				Addr:         "localhost:6379",
				DB:           0,
				PoolSize:     10,
				MinIdleConns: 5,
				DialTimeout:  0,
			},
			wantErr: true,
			errMsg:  "redis dial_timeout must be > 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func setupMiniRedis(t *testing.T) (*miniredis.Miniredis, *Client) {
	// Start miniredis
	mr := miniredis.RunT(t)

	// Create config using options.RedisOptions
	redisOpts := &options.RedisOptions{
		Addr:         mr.Addr(),
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	log := mockLogger()

	// Create client
	client, err := NewClient(redisOpts, log)
	require.NoError(t, err)
	require.NotNil(t, client)

	return mr, client
}

func TestNewClient_Success(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	assert.NotNil(t, client.Client())
}

func TestNewClient_ConnectionFailure(t *testing.T) {
	redisOpts := &options.RedisOptions{
		Addr:         "localhost:9999", // Non-existent server
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  100 * time.Millisecond, // Short timeout
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
	}

	log := mockLogger()

	_, err := NewClient(redisOpts, log)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to ping Redis")
}

func TestClient_Health(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()

	// Test Health
	err := client.Health(ctx)
	require.NoError(t, err)

	// Close miniredis and test again
	mr.Close()
	err = client.Health(ctx)
	require.Error(t, err)
}

func TestClient_HealthCheck(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	// Test HealthCheck (uses default timeout)
	err := client.HealthCheck()
	require.NoError(t, err)
}

func TestClient_PoolStats(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()
	defer client.Close()

	stats := client.PoolStats()
	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.TotalConns, uint32(0))
}

func TestClient_Close(t *testing.T) {
	mr, client := setupMiniRedis(t)
	defer mr.Close()

	// Close client
	err := client.Close()
	require.NoError(t, err)

	// Trying to use closed client should fail
	ctx := context.Background()
	err = client.Health(ctx)
	require.Error(t, err)
}

func TestNewClient_InvalidConfig(t *testing.T) {
	log := mockLogger()
	redisOpts := &options.RedisOptions{
		Addr: "", // Invalid: empty address
	}

	_, err := NewClient(redisOpts, log)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config")
}
