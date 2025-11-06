package database

import (
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	assert.Equal(t, "localhost", opts.Host)
	assert.Equal(t, 3306, opts.Port)
	assert.Equal(t, "root", opts.User)
	assert.Equal(t, "", opts.Password)
	assert.Equal(t, "", opts.Database)
	assert.Equal(t, "disable", opts.SSLMode)

	assert.Equal(t, 100, opts.MaxOpenConns)
	assert.Equal(t, 10, opts.MaxIdleConns)
	assert.Equal(t, time.Hour, opts.ConnMaxLifetime)
	assert.Equal(t, 10*time.Minute, opts.ConnMaxIdleTime)

	assert.Equal(t, "error", opts.LogLevel)
	assert.False(t, opts.AutoMigrate)
}

func TestOptionsValidate(t *testing.T) {
	tests := []struct {
		name    string
		options *Options
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid options",
			options: &Options{
				Host:         "localhost",
				Port:         3306,
				User:         "root",
				Database:     "testdb",
				MaxOpenConns: 100,
				MaxIdleConns: 10,
			},
			wantErr: false,
		},
		{
			name: "missing host",
			options: &Options{
				Host:     "",
				Port:     3306,
				User:     "root",
				Database: "testdb",
			},
			wantErr: true,
			errMsg:  "database host is required",
		},
		{
			name: "invalid port - zero",
			options: &Options{
				Host:     "localhost",
				Port:     0,
				User:     "root",
				Database: "testdb",
			},
			wantErr: true,
			errMsg:  "invalid database port: 0",
		},
		{
			name: "invalid port - negative",
			options: &Options{
				Host:     "localhost",
				Port:     -1,
				User:     "root",
				Database: "testdb",
			},
			wantErr: true,
			errMsg:  "invalid database port: -1",
		},
		{
			name: "invalid port - too large",
			options: &Options{
				Host:     "localhost",
				Port:     65536,
				User:     "root",
				Database: "testdb",
			},
			wantErr: true,
			errMsg:  "invalid database port: 65536",
		},
		{
			name: "missing user",
			options: &Options{
				Host:     "localhost",
				Port:     3306,
				User:     "",
				Database: "testdb",
			},
			wantErr: true,
			errMsg:  "database user is required",
		},
		{
			name: "missing database name",
			options: &Options{
				Host:     "localhost",
				Port:     3306,
				User:     "root",
				Database: "",
			},
			wantErr: true,
			errMsg:  "database name is required",
		},
		{
			name: "negative max open connections",
			options: &Options{
				Host:         "localhost",
				Port:         3306,
				User:         "root",
				Database:     "testdb",
				MaxOpenConns: -1,
			},
			wantErr: true,
			errMsg:  "max open connections must be >= 0",
		},
		{
			name: "negative max idle connections",
			options: &Options{
				Host:         "localhost",
				Port:         3306,
				User:         "root",
				Database:     "testdb",
				MaxIdleConns: -1,
			},
			wantErr: true,
			errMsg:  "max idle connections must be >= 0",
		},
		{
			name: "max idle > max open",
			options: &Options{
				Host:         "localhost",
				Port:         3306,
				User:         "root",
				Database:     "testdb",
				MaxOpenConns: 10,
				MaxIdleConns: 20,
			},
			wantErr: true,
			errMsg:  "max idle connections must be <= max open connections",
		},
		{
			name: "boundary value - min valid port",
			options: &Options{
				Host:     "localhost",
				Port:     1,
				User:     "root",
				Database: "testdb",
			},
			wantErr: false,
		},
		{
			name: "boundary value - max valid port",
			options: &Options{
				Host:     "localhost",
				Port:     65535,
				User:     "root",
				Database: "testdb",
			},
			wantErr: false,
		},
		{
			name: "zero max open connections (unlimited)",
			options: &Options{
				Host:         "localhost",
				Port:         3306,
				User:         "root",
				Database:     "testdb",
				MaxOpenConns: 0,
				MaxIdleConns: 0,
			},
			wantErr: false,
		},
		{
			name: "idle equals open",
			options: &Options{
				Host:         "localhost",
				Port:         3306,
				User:         "root",
				Database:     "testdb",
				MaxOpenConns: 50,
				MaxIdleConns: 50,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOptionsComplete(t *testing.T) {
	tests := []struct {
		name     string
		input    *Options
		expected *Options
	}{
		{
			name:  "complete empty options",
			input: &Options{},
			expected: &Options{
				Host:            "localhost",
				Port:            3306,
				SSLMode:         "disable",
				MaxOpenConns:    100,
				MaxIdleConns:    10,
				ConnMaxLifetime: time.Hour,
				ConnMaxIdleTime: 10 * time.Minute,
				LogLevel:        "error",
			},
		},
		{
			name: "complete partial options",
			input: &Options{
				User:     "testuser",
				Database: "testdb",
			},
			expected: &Options{
				Host:            "localhost",
				Port:            3306,
				User:            "testuser",
				Database:        "testdb",
				SSLMode:         "disable",
				MaxOpenConns:    100,
				MaxIdleConns:    10,
				ConnMaxLifetime: time.Hour,
				ConnMaxIdleTime: 10 * time.Minute,
				LogLevel:        "error",
			},
		},
		{
			name: "do not override existing values",
			input: &Options{
				Host:            "192.168.1.1",
				Port:            5432,
				User:            "admin",
				Password:        "secret",
				Database:        "mydb",
				SSLMode:         "require",
				MaxOpenConns:    200,
				MaxIdleConns:    20,
				ConnMaxLifetime: 2 * time.Hour,
				ConnMaxIdleTime: 20 * time.Minute,
				LogLevel:        "debug",
				AutoMigrate:     true,
			},
			expected: &Options{
				Host:            "192.168.1.1",
				Port:            5432,
				User:            "admin",
				Password:        "secret",
				Database:        "mydb",
				SSLMode:         "require",
				MaxOpenConns:    200,
				MaxIdleConns:    20,
				ConnMaxLifetime: 2 * time.Hour,
				ConnMaxIdleTime: 20 * time.Minute,
				LogLevel:        "debug",
				AutoMigrate:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Complete()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}

func TestFunctionalOptions(t *testing.T) {
	tests := []struct {
		name     string
		options  []Option
		expected *Options
	}{
		{
			name:     "no options - defaults",
			options:  nil,
			expected: DefaultOptions(),
		},
		{
			name: "with host",
			options: []Option{
				WithHost("192.168.1.1"),
			},
			expected: &Options{
				Host:            "192.168.1.1",
				Port:            3306,
				User:            "root",
				Password:        "",
				Database:        "",
				SSLMode:         "disable",
				MaxOpenConns:    100,
				MaxIdleConns:    10,
				ConnMaxLifetime: time.Hour,
				ConnMaxIdleTime: 10 * time.Minute,
				LogLevel:        "error",
				AutoMigrate:     false,
			},
		},
		{
			name: "with port",
			options: []Option{
				WithPort(5432),
			},
			expected: &Options{
				Host:            "localhost",
				Port:            5432,
				User:            "root",
				Password:        "",
				Database:        "",
				SSLMode:         "disable",
				MaxOpenConns:    100,
				MaxIdleConns:    10,
				ConnMaxLifetime: time.Hour,
				ConnMaxIdleTime: 10 * time.Minute,
				LogLevel:        "error",
				AutoMigrate:     false,
			},
		},
		{
			name: "with user",
			options: []Option{
				WithUser("admin"),
			},
			expected: &Options{
				Host:            "localhost",
				Port:            3306,
				User:            "admin",
				Password:        "",
				Database:        "",
				SSLMode:         "disable",
				MaxOpenConns:    100,
				MaxIdleConns:    10,
				ConnMaxLifetime: time.Hour,
				ConnMaxIdleTime: 10 * time.Minute,
				LogLevel:        "error",
				AutoMigrate:     false,
			},
		},
		{
			name: "with password",
			options: []Option{
				WithPassword("secret"),
			},
			expected: &Options{
				Host:            "localhost",
				Port:            3306,
				User:            "root",
				Password:        "secret",
				Database:        "",
				SSLMode:         "disable",
				MaxOpenConns:    100,
				MaxIdleConns:    10,
				ConnMaxLifetime: time.Hour,
				ConnMaxIdleTime: 10 * time.Minute,
				LogLevel:        "error",
				AutoMigrate:     false,
			},
		},
		{
			name: "with database",
			options: []Option{
				WithDatabase("mydb"),
			},
			expected: &Options{
				Host:            "localhost",
				Port:            3306,
				User:            "root",
				Password:        "",
				Database:        "mydb",
				SSLMode:         "disable",
				MaxOpenConns:    100,
				MaxIdleConns:    10,
				ConnMaxLifetime: time.Hour,
				ConnMaxIdleTime: 10 * time.Minute,
				LogLevel:        "error",
				AutoMigrate:     false,
			},
		},
		{
			name: "with max open connections",
			options: []Option{
				WithMaxOpenConns(200),
			},
			expected: &Options{
				Host:            "localhost",
				Port:            3306,
				User:            "root",
				Password:        "",
				Database:        "",
				SSLMode:         "disable",
				MaxOpenConns:    200,
				MaxIdleConns:    10,
				ConnMaxLifetime: time.Hour,
				ConnMaxIdleTime: 10 * time.Minute,
				LogLevel:        "error",
				AutoMigrate:     false,
			},
		},
		{
			name: "with max idle connections",
			options: []Option{
				WithMaxIdleConns(20),
			},
			expected: &Options{
				Host:            "localhost",
				Port:            3306,
				User:            "root",
				Password:        "",
				Database:        "",
				SSLMode:         "disable",
				MaxOpenConns:    100,
				MaxIdleConns:    20,
				ConnMaxLifetime: time.Hour,
				ConnMaxIdleTime: 10 * time.Minute,
				LogLevel:        "error",
				AutoMigrate:     false,
			},
		},
		{
			name: "with auto migrate",
			options: []Option{
				WithAutoMigrate(true),
			},
			expected: &Options{
				Host:            "localhost",
				Port:            3306,
				User:            "root",
				Password:        "",
				Database:        "",
				SSLMode:         "disable",
				MaxOpenConns:    100,
				MaxIdleConns:    10,
				ConnMaxLifetime: time.Hour,
				ConnMaxIdleTime: 10 * time.Minute,
				LogLevel:        "error",
				AutoMigrate:     true,
			},
		},
		{
			name: "multiple options chained",
			options: []Option{
				WithHost("192.168.1.1"),
				WithPort(5432),
				WithUser("admin"),
				WithPassword("secret"),
				WithDatabase("mydb"),
				WithMaxOpenConns(200),
				WithMaxIdleConns(20),
				WithAutoMigrate(true),
			},
			expected: &Options{
				Host:            "192.168.1.1",
				Port:            5432,
				User:            "admin",
				Password:        "secret",
				Database:        "mydb",
				SSLMode:         "disable",
				MaxOpenConns:    200,
				MaxIdleConns:    20,
				ConnMaxLifetime: time.Hour,
				ConnMaxIdleTime: 10 * time.Minute,
				LogLevel:        "error",
				AutoMigrate:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultOptions()
			for _, opt := range tt.options {
				opt(opts)
			}
			assert.Equal(t, tt.expected, opts)
		})
	}
}

func TestAddFlags(t *testing.T) {
	tests := []struct {
		name     string
		prefixes []string
		verify   func(t *testing.T, fs *pflag.FlagSet, opts *Options)
	}{
		{
			name:     "no prefix",
			prefixes: nil,
			verify: func(t *testing.T, fs *pflag.FlagSet, opts *Options) {
				flag := fs.Lookup("db-host")
				require.NotNil(t, flag)
				assert.Equal(t, "localhost", flag.DefValue)

				flag = fs.Lookup("db-port")
				require.NotNil(t, flag)
				assert.Equal(t, "3306", flag.DefValue)

				flag = fs.Lookup("db-user")
				require.NotNil(t, flag)
				assert.Equal(t, "root", flag.DefValue)

				flag = fs.Lookup("db-auto-migrate")
				require.NotNil(t, flag)
				assert.Equal(t, "false", flag.DefValue)
			},
		},
		{
			name:     "with prefix",
			prefixes: []string{"mysql"},
			verify: func(t *testing.T, fs *pflag.FlagSet, opts *Options) {
				flag := fs.Lookup("mysql-db-host")
				require.NotNil(t, flag)
				assert.Equal(t, "localhost", flag.DefValue)

				flag = fs.Lookup("mysql-db-port")
				require.NotNil(t, flag)
				assert.Equal(t, "3306", flag.DefValue)

				flag = fs.Lookup("mysql-db-max-open-conns")
				require.NotNil(t, flag)
				assert.Equal(t, "100", flag.DefValue)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultOptions()
			fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
			opts.AddFlags(fs, tt.prefixes...)
			tt.verify(t, fs, opts)
		})
	}
}

func TestAddFlagsAndParse(t *testing.T) {
	opts := DefaultOptions()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	opts.AddFlags(fs)

	args := []string{
		"--db-host=192.168.1.1",
		"--db-port=5432",
		"--db-user=admin",
		"--db-password=secret",
		"--db-name=mydb",
		"--db-max-open-conns=200",
		"--db-max-idle-conns=20",
		"--db-auto-migrate=true",
	}

	err := fs.Parse(args)
	require.NoError(t, err)

	assert.Equal(t, "192.168.1.1", opts.Host)
	assert.Equal(t, 5432, opts.Port)
	assert.Equal(t, "admin", opts.User)
	assert.Equal(t, "secret", opts.Password)
	assert.Equal(t, "mydb", opts.Database)
	assert.Equal(t, 200, opts.MaxOpenConns)
	assert.Equal(t, 20, opts.MaxIdleConns)
	assert.True(t, opts.AutoMigrate)
}

func TestOptionsCompleteAndValidate(t *testing.T) {
	tests := []struct {
		name         string
		input        *Options
		expectValid  bool
		expectChange bool
	}{
		{
			name: "empty options - complete but fail validation",
			input: &Options{
				// Missing required fields: user, database
			},
			expectValid:  false,
			expectChange: true,
		},
		{
			name: "partial valid options - complete and validate",
			input: &Options{
				User:     "root",
				Database: "testdb",
			},
			expectValid:  true,
			expectChange: true,
		},
		{
			name: "invalid options - complete but fail validation",
			input: &Options{
				Port: -1,
			},
			expectValid:  false,
			expectChange: true,
		},
		{
			name: "valid options with user and database",
			input: &Options{
				User:     "testuser",
				Database: "testdb",
			},
			expectValid:  true,
			expectChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := *tt.input

			err := tt.input.Complete()
			require.NoError(t, err)

			if tt.expectChange {
				assert.NotEqual(t, &original, tt.input)
			}

			err = tt.input.Validate()
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestConnectionPoolValidation(t *testing.T) {
	tests := []struct {
		name         string
		maxOpenConns int
		maxIdleConns int
		wantErr      bool
	}{
		{
			name:         "idle less than open",
			maxOpenConns: 100,
			maxIdleConns: 10,
			wantErr:      false,
		},
		{
			name:         "idle equals open",
			maxOpenConns: 50,
			maxIdleConns: 50,
			wantErr:      false,
		},
		{
			name:         "idle greater than open",
			maxOpenConns: 10,
			maxIdleConns: 20,
			wantErr:      true,
		},
		{
			name:         "unlimited open (0), any idle",
			maxOpenConns: 0,
			maxIdleConns: 10,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{
				Host:         "localhost",
				Port:         3306,
				User:         "root",
				Database:     "testdb",
				MaxOpenConns: tt.maxOpenConns,
				MaxIdleConns: tt.maxIdleConns,
			}

			err := opts.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
