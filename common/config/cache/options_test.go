package cache

import (
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	assert.Equal(t, TypeMemory, opts.Type)
	assert.Equal(t, "", opts.KeyPrefix)
	assert.Equal(t, time.Hour, opts.DefaultExpiration)

	assert.Equal(t, 1000, opts.MemoryMaxSize)
	assert.Equal(t, 10*time.Minute, opts.MemoryCleanupInterval)

	assert.Equal(t, "localhost:6379", opts.RedisAddr)
	assert.Equal(t, "", opts.RedisPassword)
	assert.Equal(t, 0, opts.RedisDB)
	assert.Equal(t, 10, opts.RedisPoolSize)

	assert.Equal(t, 5*time.Minute, opts.L2LocalTTL)
	assert.Equal(t, time.Hour, opts.L2RemoteTTL)
}

func TestOptionsValidate(t *testing.T) {
	tests := []struct {
		name    string
		options *Options
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid memory cache",
			options: &Options{
				Type:          TypeMemory,
				MemoryMaxSize: 1000,
			},
			wantErr: false,
		},
		{
			name: "valid redis cache",
			options: &Options{
				Type:          TypeRedis,
				RedisAddr:     "localhost:6379",
				RedisDB:       0,
				RedisPoolSize: 10,
			},
			wantErr: false,
		},
		{
			name: "valid L2 cache",
			options: &Options{
				Type:          TypeL2,
				MemoryMaxSize: 1000,
				RedisAddr:     "localhost:6379",
				RedisDB:       0,
				RedisPoolSize: 10,
			},
			wantErr: false,
		},
		{
			name: "invalid cache type",
			options: &Options{
				Type: Type("invalid"),
			},
			wantErr: true,
			errMsg:  "invalid cache type",
		},
		{
			name: "redis cache - missing address",
			options: &Options{
				Type:      TypeRedis,
				RedisAddr: "",
			},
			wantErr: true,
			errMsg:  "redis address is required",
		},
		{
			name: "redis cache - invalid DB number (negative)",
			options: &Options{
				Type:      TypeRedis,
				RedisAddr: "localhost:6379",
				RedisDB:   -1,
			},
			wantErr: true,
			errMsg:  "redis DB must be between 0 and 15",
		},
		{
			name: "redis cache - invalid DB number (too large)",
			options: &Options{
				Type:      TypeRedis,
				RedisAddr: "localhost:6379",
				RedisDB:   16,
			},
			wantErr: true,
			errMsg:  "redis DB must be between 0 and 15",
		},
		{
			name: "redis cache - invalid pool size",
			options: &Options{
				Type:          TypeRedis,
				RedisAddr:     "localhost:6379",
				RedisDB:       0,
				RedisPoolSize: 0,
			},
			wantErr: true,
			errMsg:  "redis pool size must be > 0",
		},
		{
			name: "memory cache - invalid max size",
			options: &Options{
				Type:          TypeMemory,
				MemoryMaxSize: 0,
			},
			wantErr: true,
			errMsg:  "memory max size must be > 0",
		},
		{
			name: "L2 cache - missing redis address",
			options: &Options{
				Type:          TypeL2,
				MemoryMaxSize: 1000,
				RedisAddr:     "",
			},
			wantErr: true,
			errMsg:  "redis address is required",
		},
		{
			name: "L2 cache - invalid memory max size",
			options: &Options{
				Type:          TypeL2,
				MemoryMaxSize: -1,
				RedisAddr:     "localhost:6379",
				RedisPoolSize: 10,
			},
			wantErr: true,
			errMsg:  "memory max size must be > 0",
		},
		{
			name: "boundary value - redis DB 0",
			options: &Options{
				Type:          TypeRedis,
				RedisAddr:     "localhost:6379",
				RedisDB:       0,
				RedisPoolSize: 1,
			},
			wantErr: false,
		},
		{
			name: "boundary value - redis DB 15",
			options: &Options{
				Type:          TypeRedis,
				RedisAddr:     "localhost:6379",
				RedisDB:       15,
				RedisPoolSize: 1,
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
				Type:                  TypeMemory,
				DefaultExpiration:     time.Hour,
				MemoryMaxSize:         1000,
				MemoryCleanupInterval: 10 * time.Minute,
				RedisAddr:             "localhost:6379",
				RedisPoolSize:         10,
				L2LocalTTL:            5 * time.Minute,
				L2RemoteTTL:           time.Hour,
			},
		},
		{
			name: "complete partial options",
			input: &Options{
				Type:      TypeRedis,
				RedisAddr: "192.168.1.1:6379",
			},
			expected: &Options{
				Type:                  TypeRedis,
				DefaultExpiration:     time.Hour,
				MemoryMaxSize:         1000,
				MemoryCleanupInterval: 10 * time.Minute,
				RedisAddr:             "192.168.1.1:6379",
				RedisPoolSize:         10,
				L2LocalTTL:            5 * time.Minute,
				L2RemoteTTL:           time.Hour,
			},
		},
		{
			name: "do not override existing values",
			input: &Options{
				Type:                  TypeL2,
				KeyPrefix:             "myapp:",
				DefaultExpiration:     2 * time.Hour,
				MemoryMaxSize:         2000,
				MemoryCleanupInterval: 20 * time.Minute,
				RedisAddr:             "redis.example.com:6379",
				RedisPassword:         "secret",
				RedisDB:               5,
				RedisPoolSize:         50,
				L2LocalTTL:            10 * time.Minute,
				L2RemoteTTL:           2 * time.Hour,
			},
			expected: &Options{
				Type:                  TypeL2,
				KeyPrefix:             "myapp:",
				DefaultExpiration:     2 * time.Hour,
				MemoryMaxSize:         2000,
				MemoryCleanupInterval: 20 * time.Minute,
				RedisAddr:             "redis.example.com:6379",
				RedisPassword:         "secret",
				RedisDB:               5,
				RedisPoolSize:         50,
				L2LocalTTL:            10 * time.Minute,
				L2RemoteTTL:           2 * time.Hour,
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
			name: "with type",
			options: []Option{
				WithType(TypeRedis),
			},
			expected: &Options{
				Type:                  TypeRedis,
				KeyPrefix:             "",
				DefaultExpiration:     time.Hour,
				MemoryMaxSize:         1000,
				MemoryCleanupInterval: 10 * time.Minute,
				RedisAddr:             "localhost:6379",
				RedisPassword:         "",
				RedisDB:               0,
				RedisPoolSize:         10,
				L2LocalTTL:            5 * time.Minute,
				L2RemoteTTL:           time.Hour,
			},
		},
		{
			name: "with key prefix",
			options: []Option{
				WithKeyPrefix("myapp:"),
			},
			expected: &Options{
				Type:                  TypeMemory,
				KeyPrefix:             "myapp:",
				DefaultExpiration:     time.Hour,
				MemoryMaxSize:         1000,
				MemoryCleanupInterval: 10 * time.Minute,
				RedisAddr:             "localhost:6379",
				RedisPassword:         "",
				RedisDB:               0,
				RedisPoolSize:         10,
				L2LocalTTL:            5 * time.Minute,
				L2RemoteTTL:           time.Hour,
			},
		},
		{
			name: "with default expiration",
			options: []Option{
				WithDefaultExpiration(30 * time.Minute),
			},
			expected: &Options{
				Type:                  TypeMemory,
				KeyPrefix:             "",
				DefaultExpiration:     30 * time.Minute,
				MemoryMaxSize:         1000,
				MemoryCleanupInterval: 10 * time.Minute,
				RedisAddr:             "localhost:6379",
				RedisPassword:         "",
				RedisDB:               0,
				RedisPoolSize:         10,
				L2LocalTTL:            5 * time.Minute,
				L2RemoteTTL:           time.Hour,
			},
		},
		{
			name: "with redis addr",
			options: []Option{
				WithRedisAddr("192.168.1.1:6379"),
			},
			expected: &Options{
				Type:                  TypeMemory,
				KeyPrefix:             "",
				DefaultExpiration:     time.Hour,
				MemoryMaxSize:         1000,
				MemoryCleanupInterval: 10 * time.Minute,
				RedisAddr:             "192.168.1.1:6379",
				RedisPassword:         "",
				RedisDB:               0,
				RedisPoolSize:         10,
				L2LocalTTL:            5 * time.Minute,
				L2RemoteTTL:           time.Hour,
			},
		},
		{
			name: "with redis password",
			options: []Option{
				WithRedisPassword("secret"),
			},
			expected: &Options{
				Type:                  TypeMemory,
				KeyPrefix:             "",
				DefaultExpiration:     time.Hour,
				MemoryMaxSize:         1000,
				MemoryCleanupInterval: 10 * time.Minute,
				RedisAddr:             "localhost:6379",
				RedisPassword:         "secret",
				RedisDB:               0,
				RedisPoolSize:         10,
				L2LocalTTL:            5 * time.Minute,
				L2RemoteTTL:           time.Hour,
			},
		},
		{
			name: "multiple options chained",
			options: []Option{
				WithType(TypeL2),
				WithKeyPrefix("app:"),
				WithDefaultExpiration(2 * time.Hour),
				WithRedisAddr("redis.example.com:6379"),
				WithRedisPassword("secret"),
			},
			expected: &Options{
				Type:                  TypeL2,
				KeyPrefix:             "app:",
				DefaultExpiration:     2 * time.Hour,
				MemoryMaxSize:         1000,
				MemoryCleanupInterval: 10 * time.Minute,
				RedisAddr:             "redis.example.com:6379",
				RedisPassword:         "secret",
				RedisDB:               0,
				RedisPoolSize:         10,
				L2LocalTTL:            5 * time.Minute,
				L2RemoteTTL:           time.Hour,
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
				flag := fs.Lookup("cache-type")
				require.NotNil(t, flag)
				assert.Equal(t, "memory", flag.DefValue)

				flag = fs.Lookup("cache-redis-addr")
				require.NotNil(t, flag)
				assert.Equal(t, "localhost:6379", flag.DefValue)

				flag = fs.Lookup("cache-memory-max-size")
				require.NotNil(t, flag)
				assert.Equal(t, "1000", flag.DefValue)
			},
		},
		{
			name:     "with prefix",
			prefixes: []string{"app"},
			verify: func(t *testing.T, fs *pflag.FlagSet, opts *Options) {
				flag := fs.Lookup("app-cache-type")
				require.NotNil(t, flag)
				assert.Equal(t, "memory", flag.DefValue)

				flag = fs.Lookup("app-cache-redis-db")
				require.NotNil(t, flag)
				assert.Equal(t, "0", flag.DefValue)

				flag = fs.Lookup("app-cache-redis-pool-size")
				require.NotNil(t, flag)
				assert.Equal(t, "10", flag.DefValue)
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
		"--cache-type=redis",
		"--cache-key-prefix=myapp:",
		"--cache-redis-addr=192.168.1.1:6379",
		"--cache-redis-password=secret",
		"--cache-redis-db=5",
		"--cache-redis-pool-size=50",
		"--cache-default-expiration=2h",
	}

	err := fs.Parse(args)
	require.NoError(t, err)

	assert.Equal(t, TypeRedis, opts.Type)
	assert.Equal(t, "myapp:", opts.KeyPrefix)
	assert.Equal(t, "192.168.1.1:6379", opts.RedisAddr)
	assert.Equal(t, "secret", opts.RedisPassword)
	assert.Equal(t, 5, opts.RedisDB)
	assert.Equal(t, 50, opts.RedisPoolSize)
	assert.Equal(t, 2*time.Hour, opts.DefaultExpiration)
}

func TestOptionsCompleteAndValidate(t *testing.T) {
	tests := []struct {
		name         string
		input        *Options
		expectValid  bool
		expectChange bool
	}{
		{
			name:         "empty options - complete and validate (memory cache)",
			input:        &Options{},
			expectValid:  true,
			expectChange: true,
		},
		{
			name: "redis cache - complete and validate",
			input: &Options{
				Type:      TypeRedis,
				RedisAddr: "localhost:6379",
			},
			expectValid:  true,
			expectChange: true,
		},
		{
			name: "invalid options - complete but fail validation",
			input: &Options{
				Type: Type("invalid"),
			},
			expectValid:  false,
			expectChange: true,
		},
		{
			name:         "default options - already valid",
			input:        DefaultOptions(),
			expectValid:  true,
			expectChange: false,
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

func TestCacheTypes(t *testing.T) {
	tests := []struct {
		name      string
		cacheType Type
		wantValid bool
	}{
		{
			name:      "memory cache",
			cacheType: TypeMemory,
			wantValid: true,
		},
		{
			name:      "redis cache",
			cacheType: TypeRedis,
			wantValid: true,
		},
		{
			name:      "L2 cache",
			cacheType: TypeL2,
			wantValid: true,
		},
		{
			name:      "invalid cache type",
			cacheType: Type("memcached"),
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{
				Type:          tt.cacheType,
				MemoryMaxSize: 1000,
			}

			// Add required fields for redis/L2
			if tt.cacheType == TypeRedis || tt.cacheType == TypeL2 {
				opts.RedisAddr = "localhost:6379"
				opts.RedisPoolSize = 10
			}

			err := opts.Validate()
			if tt.wantValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestRedisDBValidation(t *testing.T) {
	tests := []struct {
		name    string
		db      int
		wantErr bool
	}{
		{name: "DB 0", db: 0, wantErr: false},
		{name: "DB 1", db: 1, wantErr: false},
		{name: "DB 8", db: 8, wantErr: false},
		{name: "DB 15", db: 15, wantErr: false},
		{name: "DB -1 (invalid)", db: -1, wantErr: true},
		{name: "DB 16 (invalid)", db: 16, wantErr: true},
		{name: "DB 100 (invalid)", db: 100, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{
				Type:          TypeRedis,
				RedisAddr:     "localhost:6379",
				RedisDB:       tt.db,
				RedisPoolSize: 10,
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
