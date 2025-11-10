package mysql

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

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
	config := options.NewMySQLOptions()

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 3306, config.Port)
	assert.Equal(t, "root", config.User)
	assert.Equal(t, "test", config.Database)
	assert.Equal(t, "utf8mb4", config.Charset)
	assert.Equal(t, 100, config.MaxOpenConns)
	assert.Equal(t, 10, config.MaxIdleConns)
	assert.Equal(t, time.Hour, config.ConnMaxLifetime)
	assert.Equal(t, "silent", config.LogLevel)
	assert.Equal(t, 200*time.Millisecond, config.SlowQueryThreshold)
}

func TestConfig_DSN(t *testing.T) {
	config := &options.MySQLOptions{
		User:     "testuser",
		Password: "testpass",
		Host:     "testhost",
		Port:     3307,
		Database: "testdb",
		Charset:  "utf8mb4",
	}

	expected := "testuser:testpass@tcp(testhost:3307)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	assert.Equal(t, expected, config.DSN())
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *options.MySQLOptions
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &options.MySQLOptions{
				Host:               "localhost",
				Port:               3306,
				User:               "root",
				Database:           "test",
				MaxOpenConns:       100,
				MaxIdleConns:       10,
				ConnMaxLifetime:    time.Hour,
				LogLevel:           "info",
				SlowQueryThreshold: 200 * time.Millisecond,
			},
			wantErr: false,
		},
		{
			name: "empty host",
			config: &options.MySQLOptions{
				Host:     "",
				Port:     3306,
				User:     "root",
				Database: "test",
			},
			wantErr: true,
			errMsg:  "host",
		},
		{
			name: "invalid port",
			config: &options.MySQLOptions{
				Host:     "localhost",
				Port:     0,
				User:     "root",
				Database: "test",
			},
			wantErr: true,
			errMsg:  "port",
		},
		{
			name: "empty user",
			config: &options.MySQLOptions{
				Host:     "localhost",
				Port:     3306,
				User:     "",
				Database: "test",
			},
			wantErr: true,
			errMsg:  "user",
		},
		{
			name: "empty database",
			config: &options.MySQLOptions{
				Host:     "localhost",
				Port:     3306,
				User:     "root",
				Database: "",
			},
			wantErr: true,
			errMsg:  "database",
		},
		{
			name: "idle conns exceeds max",
			config: &options.MySQLOptions{
				Host:               "localhost",
				Port:               3306,
				User:               "root",
				Database:           "test",
				MaxOpenConns:       10,
				MaxIdleConns:       20,
				ConnMaxLifetime:    time.Hour,
				LogLevel:           "info",
				SlowQueryThreshold: 200 * time.Millisecond,
			},
			wantErr: true,
			errMsg:  "min_connections",
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

func TestClient_WithMockDB(t *testing.T) {
	// Create mock SQL database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Expect ping
	mock.ExpectPing()

	// Create GORM DB from mock
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	log := mockLogger()
	config := options.NewMySQLOptions()

	// Create client with mock DB
	client := &Client{
		db:     gormDB,
		logger: log.With("component", "mysql"),
		config: *config,
	}

	// Test DB() method
	assert.NotNil(t, client.DB())
	assert.Equal(t, gormDB, client.DB())

	// Test Health() method
	ctx := context.Background()
	err = client.Health(ctx)
	require.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClient_Stats(t *testing.T) {
	// Create mock SQL database
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Create GORM DB from mock
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	log := mockLogger()
	config := options.NewMySQLOptions()

	client := &Client{
		db:     gormDB,
		logger: log.With("component", "mysql"),
		config: *config,
	}

	// Test Stats() method
	stats, err := client.Stats()
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.MaxOpenConnections, 0)
}

func TestClient_HealthCheck(t *testing.T) {
	// Create mock SQL database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Expect ping
	mock.ExpectPing()

	// Create GORM DB from mock
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	log := mockLogger()
	config := options.NewMySQLOptions()

	client := &Client{
		db:     gormDB,
		logger: log.With("component", "mysql"),
		config: *config,
	}

	// Test HealthCheck() method (uses default timeout)
	err = client.HealthCheck()
	require.NoError(t, err)

	// Verify expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNewClient_InvalidConfig(t *testing.T) {
	log := mockLogger()
	config := &options.MySQLOptions{
		Host: "", // Invalid: empty host
	}

	_, err := NewClient(config, log)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config")
}
