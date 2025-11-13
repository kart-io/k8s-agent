package options

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPostgresOptions(t *testing.T) {
	opts := NewPostgresOptions()

	assert.NotNil(t, opts)
	assert.Equal(t, "localhost", opts.Host)
	assert.Equal(t, 5432, opts.Port)
	assert.Equal(t, "postgres", opts.User)
	assert.Equal(t, "postgres", opts.Database)
	assert.Equal(t, "prefer", opts.SSLMode)
	assert.Equal(t, "public", opts.Schema)
	assert.Equal(t, 100, opts.MaxOpenConns)
	assert.Equal(t, 10, opts.MaxIdleConns)
	assert.Equal(t, time.Hour, opts.ConnMaxLifetime)
	assert.Equal(t, 10*time.Minute, opts.ConnMaxIdleTime)
	assert.Equal(t, "silent", opts.LogLevel)
	assert.Equal(t, 200*time.Millisecond, opts.SlowQueryThreshold)
	assert.False(t, opts.AutoMigrate)
	assert.False(t, opts.PreferSimpleProtocol)
	assert.Equal(t, "prepare", opts.StatementCacheMode)
}

func TestPostgresOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    *PostgresOptions
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid options",
			opts:    NewPostgresOptions(),
			wantErr: false,
		},
		{
			name: "empty host",
			opts: &PostgresOptions{
				Host:     "",
				Port:     5432,
				User:     "postgres",
				Database: "test",
			},
			wantErr: true,
			errMsg:  "PostgreSQL host",
		},
		{
			name: "invalid port",
			opts: &PostgresOptions{
				Host:     "localhost",
				Port:     99999,
				User:     "postgres",
				Database: "test",
			},
			wantErr: true,
			errMsg:  "port",
		},
		{
			name: "zero port",
			opts: &PostgresOptions{
				Host:     "localhost",
				Port:     0,
				User:     "postgres",
				Database: "test",
			},
			wantErr: true,
			errMsg:  "port",
		},
		{
			name: "empty user",
			opts: &PostgresOptions{
				Host:     "localhost",
				Port:     5432,
				User:     "",
				Database: "test",
			},
			wantErr: true,
			errMsg:  "PostgreSQL user",
		},
		{
			name: "empty database",
			opts: &PostgresOptions{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Database: "",
			},
			wantErr: true,
			errMsg:  "PostgreSQL database",
		},
		{
			name: "invalid SSL mode",
			opts: &PostgresOptions{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Database: "test",
				SSLMode:  "invalid",
			},
			wantErr: true,
			errMsg:  "invalid SSL mode",
		},
		{
			name: "valid SSL modes",
			opts: &PostgresOptions{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Database: "test",
				SSLMode:  "require",
			},
			wantErr: false,
		},
		{
			name: "invalid statement cache mode",
			opts: &PostgresOptions{
				Host:               "localhost",
				Port:               5432,
				User:               "postgres",
				Database:           "test",
				StatementCacheMode: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid statement cache mode",
		},
		{
			name: "max idle greater than max open",
			opts: &PostgresOptions{
				Host:         "localhost",
				Port:         5432,
				User:         "postgres",
				Database:     "test",
				MaxOpenConns: 10,
				MaxIdleConns: 20,
			},
			wantErr: true,
			errMsg:  "min_connections",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPostgresOptions_DSN(t *testing.T) {
	tests := []struct {
		name     string
		opts     *PostgresOptions
		expected string
	}{
		{
			name: "basic DSN",
			opts: &PostgresOptions{
				Host:     "localhost",
				Port:     5432,
				User:     "user",
				Password: "pass",
				Database: "mydb",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=user password=pass dbname=mydb sslmode=disable",
		},
		{
			name: "with timezone",
			opts: &PostgresOptions{
				Host:     "localhost",
				Port:     5432,
				User:     "user",
				Password: "pass",
				Database: "mydb",
				SSLMode:  "prefer",
				TimeZone: "UTC",
			},
			expected: "host=localhost port=5432 user=user password=pass dbname=mydb sslmode=prefer TimeZone=UTC",
		},
		{
			name: "with custom schema",
			opts: &PostgresOptions{
				Host:     "localhost",
				Port:     5432,
				User:     "user",
				Password: "pass",
				Database: "mydb",
				SSLMode:  "prefer",
				Schema:   "myschema",
			},
			expected: "host=localhost port=5432 user=user password=pass dbname=mydb sslmode=prefer search_path=myschema",
		},
		{
			name: "with prefer simple protocol",
			opts: &PostgresOptions{
				Host:                 "localhost",
				Port:                 5432,
				User:                 "user",
				Password:             "pass",
				Database:             "mydb",
				SSLMode:              "prefer",
				PreferSimpleProtocol: true,
			},
			expected: "host=localhost port=5432 user=user password=pass dbname=mydb sslmode=prefer prefer_simple_protocol=true",
		},
		{
			name: "with statement cache mode",
			opts: &PostgresOptions{
				Host:               "localhost",
				Port:               5432,
				User:               "user",
				Password:           "pass",
				Database:           "mydb",
				SSLMode:            "prefer",
				StatementCacheMode: "describe",
			},
			expected: "host=localhost port=5432 user=user password=pass dbname=mydb sslmode=prefer statement_cache_mode=describe",
		},
		{
			name: "full DSN",
			opts: &PostgresOptions{
				Host:                 "localhost",
				Port:                 5432,
				User:                 "user",
				Password:             "pass",
				Database:             "mydb",
				SSLMode:              "require",
				TimeZone:             "Asia/Shanghai",
				Schema:               "custom",
				PreferSimpleProtocol: true,
				StatementCacheMode:   "prepare",
			},
			expected: "host=localhost port=5432 user=user password=pass dbname=mydb sslmode=require TimeZone=Asia/Shanghai search_path=custom prefer_simple_protocol=true statement_cache_mode=prepare",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := tt.opts.DSN()
			assert.Equal(t, tt.expected, dsn)
		})
	}
}

func TestPostgresOptions_AddFlags(t *testing.T) {
	opts := NewPostgresOptions()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)

	opts.AddFlags(fs)

	// Check that all expected flags are registered
	flags := []string{
		"postgres.host",
		"postgres.port",
		"postgres.user",
		"postgres.password",
		"postgres.database",
		"postgres.ssl-mode",
		"postgres.timezone",
		"postgres.schema",
		"postgres.max-open-conns",
		"postgres.max-idle-conns",
		"postgres.conn-max-lifetime",
		"postgres.conn-max-idle-time",
		"postgres.log-level",
		"postgres.slow-query-threshold",
		"postgres.auto-migrate",
		"postgres.prefer-simple-protocol",
		"postgres.statement-cache-mode",
	}

	for _, flag := range flags {
		assert.NotNil(t, fs.Lookup(flag), "flag %s should be registered", flag)
	}

	// Test parsing
	args := []string{
		"--postgres.host=db.example.com",
		"--postgres.port=5433",
		"--postgres.user=testuser",
		"--postgres.database=testdb",
		"--postgres.ssl-mode=require",
	}

	err := fs.Parse(args)
	require.NoError(t, err)

	assert.Equal(t, "db.example.com", opts.Host)
	assert.Equal(t, 5433, opts.Port)
	assert.Equal(t, "testuser", opts.User)
	assert.Equal(t, "testdb", opts.Database)
	assert.Equal(t, "require", opts.SSLMode)
}

func TestPostgresOptions_Complete(t *testing.T) {
	tests := []struct {
		name     string
		initial  *PostgresOptions
		expected *PostgresOptions
	}{
		{
			name:    "empty options",
			initial: &PostgresOptions{},
			expected: &PostgresOptions{
				Host:               "localhost",
				Port:               5432,
				User:               "postgres",
				Database:           "postgres",
				SSLMode:            "prefer",
				TimeZone:           "Asia/Shanghai",
				Schema:             "public",
				LogLevel:           "silent",
				MaxOpenConns:       100,
				MaxIdleConns:       10,
				ConnMaxLifetime:    time.Hour,
				ConnMaxIdleTime:    10 * time.Minute,
				SlowQueryThreshold: 200 * time.Millisecond,
				StatementCacheMode: "prepare",
			},
		},
		{
			name: "invalid port",
			initial: &PostgresOptions{
				Port: -1,
			},
			expected: &PostgresOptions{
				Host:               "localhost",
				Port:               5432,
				User:               "postgres",
				Database:           "postgres",
				SSLMode:            "prefer",
				TimeZone:           "Asia/Shanghai",
				Schema:             "public",
				LogLevel:           "silent",
				MaxOpenConns:       100,
				MaxIdleConns:       10,
				ConnMaxLifetime:    time.Hour,
				ConnMaxIdleTime:    10 * time.Minute,
				SlowQueryThreshold: 200 * time.Millisecond,
				StatementCacheMode: "prepare",
			},
		},
		{
			name: "max idle greater than max open",
			initial: &PostgresOptions{
				MaxOpenConns: 10,
				MaxIdleConns: 20,
			},
			expected: &PostgresOptions{
				Host:               "localhost",
				Port:               5432,
				User:               "postgres",
				Database:           "postgres",
				SSLMode:            "prefer",
				TimeZone:           "Asia/Shanghai",
				Schema:             "public",
				LogLevel:           "silent",
				MaxOpenConns:       10,
				MaxIdleConns:       10, // Should be adjusted
				ConnMaxLifetime:    time.Hour,
				ConnMaxIdleTime:    10 * time.Minute,
				SlowQueryThreshold: 200 * time.Millisecond,
				StatementCacheMode: "prepare",
			},
		},
		{
			name: "zero timeouts",
			initial: &PostgresOptions{
				ConnMaxLifetime:    0,
				ConnMaxIdleTime:    0,
				SlowQueryThreshold: 0,
			},
			expected: &PostgresOptions{
				Host:               "localhost",
				Port:               5432,
				User:               "postgres",
				Database:           "postgres",
				SSLMode:            "prefer",
				TimeZone:           "Asia/Shanghai",
				Schema:             "public",
				LogLevel:           "silent",
				MaxOpenConns:       100,
				MaxIdleConns:       10,
				ConnMaxLifetime:    time.Hour,
				ConnMaxIdleTime:    10 * time.Minute,
				SlowQueryThreshold: 200 * time.Millisecond,
				StatementCacheMode: "prepare",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.initial.Complete()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, tt.initial)
		})
	}
}

func TestPostgresOptions_ApplyTo(t *testing.T) {
	opts := NewPostgresOptions()
	opts.Host = "db.example.com"
	opts.Port = 5433
	opts.User = "testuser"
	opts.Database = "testdb"

	// Test applying to slice of interfaces
	var target []interface{}
	err := opts.ApplyTo(&target)
	require.NoError(t, err)
	require.Len(t, target, 1)

	config, ok := target[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "db.example.com", config["host"])
	assert.Equal(t, 5433, config["port"])
	assert.Equal(t, "testuser", config["user"])
	assert.Equal(t, "testdb", config["database"])
	assert.Contains(t, config["dsn"], "host=db.example.com")

	// Test with nil target
	err = opts.ApplyTo(nil)
	assert.NoError(t, err)
}

func TestPostgresOptions_GetConnectionInfo(t *testing.T) {
	opts := NewPostgresOptions()
	opts.Host = "db.example.com"
	opts.Port = 5433
	opts.User = "testuser"
	opts.Database = "testdb"
	opts.SSLMode = "require"
	opts.Schema = "custom"

	info := opts.GetConnectionInfo()

	assert.Equal(t, "db.example.com", info["host"])
	assert.Equal(t, 5433, info["port"])
	assert.Equal(t, "testuser", info["user"])
	assert.Equal(t, "testdb", info["database"])
	assert.Equal(t, "require", info["ssl_mode"])
	assert.Equal(t, "custom", info["schema"])
	assert.Equal(t, 100, info["max_open"])
	assert.Equal(t, 10, info["max_idle"])
}

func TestPostgresOptions_Clone(t *testing.T) {
	original := NewPostgresOptions()
	original.Host = "db.example.com"
	original.Port = 5433
	original.User = "testuser"
	original.Password = "testpass"
	original.Database = "testdb"
	original.SSLMode = "require"
	original.Schema = "custom"
	original.AutoMigrate = true

	cloned := original.Clone()

	// Check that all fields are copied
	assert.Equal(t, original, cloned)

	// Check that they are different instances
	assert.NotSame(t, original, cloned)

	// Modify original and ensure clone is not affected
	original.Host = "modified.example.com"
	assert.NotEqual(t, original.Host, cloned.Host)
	assert.Equal(t, "db.example.com", cloned.Host)
}

func TestPostgresOptions_WithFunctions(t *testing.T) {
	opts := NewPostgresOptions()

	WithPostgresHost("db.example.com")(opts)
	assert.Equal(t, "db.example.com", opts.Host)

	WithPostgresPort(5433)(opts)
	assert.Equal(t, 5433, opts.Port)

	WithPostgresUser("testuser")(opts)
	assert.Equal(t, "testuser", opts.User)

	WithPostgresPassword("testpass")(opts)
	assert.Equal(t, "testpass", opts.Password)

	WithPostgresDatabase("testdb")(opts)
	assert.Equal(t, "testdb", opts.Database)

	WithPostgresSSLMode("require")(opts)
	assert.Equal(t, "require", opts.SSLMode)

	WithPostgresSchema("custom")(opts)
	assert.Equal(t, "custom", opts.Schema)

	WithPostgresMaxOpenConns(50)(opts)
	assert.Equal(t, 50, opts.MaxOpenConns)

	WithPostgresMaxIdleConns(5)(opts)
	assert.Equal(t, 5, opts.MaxIdleConns)

	WithPostgresAutoMigrate(true)(opts)
	assert.True(t, opts.AutoMigrate)
}

func TestPostgresOptions_getGormLogLevel(t *testing.T) {
	tests := []struct {
		logLevel string
		expected int
	}{
		{"silent", 1},
		{"error", 2},
		{"warn", 3},
		{"info", 4},
		{"unknown", 1},
		{"", 1},
	}

	for _, tt := range tests {
		t.Run(tt.logLevel, func(t *testing.T) {
			opts := &PostgresOptions{LogLevel: tt.logLevel}
			level := opts.getGormLogLevel()
			assert.Equal(t, tt.expected, int(level))
		})
	}
}

func TestPostgresOptions_Health_NoConnection(t *testing.T) {
	// Test health check with invalid connection
	opts := NewPostgresOptions()
	opts.Host = "invalid.host.that.does.not.exist"
	opts.Port = 99999

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := opts.Health(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to PostgreSQL")
}
