package server

import (
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	assert.Equal(t, "0.0.0.0", opts.HTTPHost)
	assert.Equal(t, 8080, opts.HTTPPort)
	assert.Equal(t, "release", opts.HTTPMode)
	assert.Equal(t, 30*time.Second, opts.HTTPReadTimeout)
	assert.Equal(t, 30*time.Second, opts.HTTPWriteTimeout)
	assert.Equal(t, 1<<20, opts.HTTPMaxHeaderBytes)

	assert.Equal(t, "0.0.0.0", opts.GRPCHost)
	assert.Equal(t, 9090, opts.GRPCPort)

	assert.False(t, opts.EnableProfiling)
	assert.True(t, opts.EnableMetrics)
}

func TestOptionsValidate(t *testing.T) {
	tests := []struct {
		name    string
		options *Options
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid default options",
			options: DefaultOptions(),
			wantErr: false,
		},
		{
			name: "valid custom options",
			options: &Options{
				HTTPPort: 8888,
				GRPCPort: 9999,
				HTTPMode: "debug",
			},
			wantErr: false,
		},
		{
			name: "invalid HTTP port - zero",
			options: &Options{
				HTTPPort: 0,
				GRPCPort: 9090,
				HTTPMode: "release",
			},
			wantErr: true,
			errMsg:  "invalid HTTP port: 0",
		},
		{
			name: "invalid HTTP port - negative",
			options: &Options{
				HTTPPort: -1,
				GRPCPort: 9090,
				HTTPMode: "release",
			},
			wantErr: true,
			errMsg:  "invalid HTTP port: -1",
		},
		{
			name: "invalid HTTP port - too large",
			options: &Options{
				HTTPPort: 65536,
				GRPCPort: 9090,
				HTTPMode: "release",
			},
			wantErr: true,
			errMsg:  "invalid HTTP port: 65536",
		},
		{
			name: "invalid gRPC port - zero",
			options: &Options{
				HTTPPort: 8080,
				GRPCPort: 0,
				HTTPMode: "release",
			},
			wantErr: true,
			errMsg:  "invalid gRPC port: 0",
		},
		{
			name: "invalid gRPC port - negative",
			options: &Options{
				HTTPPort: 8080,
				GRPCPort: -100,
				HTTPMode: "release",
			},
			wantErr: true,
			errMsg:  "invalid gRPC port: -100",
		},
		{
			name: "invalid gRPC port - too large",
			options: &Options{
				HTTPPort: 8080,
				GRPCPort: 70000,
				HTTPMode: "release",
			},
			wantErr: true,
			errMsg:  "invalid gRPC port: 70000",
		},
		{
			name: "invalid HTTP mode",
			options: &Options{
				HTTPPort: 8080,
				GRPCPort: 9090,
				HTTPMode: "production",
			},
			wantErr: true,
			errMsg:  "invalid HTTP mode: production",
		},
		{
			name: "valid HTTP mode - debug",
			options: &Options{
				HTTPPort: 8080,
				GRPCPort: 9090,
				HTTPMode: "debug",
			},
			wantErr: false,
		},
		{
			name: "valid HTTP mode - test",
			options: &Options{
				HTTPPort: 8080,
				GRPCPort: 9090,
				HTTPMode: "test",
			},
			wantErr: false,
		},
		{
			name: "valid HTTP mode - release",
			options: &Options{
				HTTPPort: 8080,
				GRPCPort: 9090,
				HTTPMode: "release",
			},
			wantErr: false,
		},
		{
			name: "boundary value - min valid port",
			options: &Options{
				HTTPPort: 1,
				GRPCPort: 1,
				HTTPMode: "release",
			},
			wantErr: false,
		},
		{
			name: "boundary value - max valid port",
			options: &Options{
				HTTPPort: 65535,
				GRPCPort: 65535,
				HTTPMode: "release",
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
				HTTPHost:           "0.0.0.0",
				HTTPPort:           8080,
				HTTPMode:           "release",
				HTTPReadTimeout:    30 * time.Second,
				HTTPWriteTimeout:   30 * time.Second,
				HTTPMaxHeaderBytes: 1 << 20,
				GRPCHost:           "0.0.0.0",
				GRPCPort:           9090,
			},
		},
		{
			name: "complete partial options",
			input: &Options{
				HTTPPort: 9000,
				GRPCPort: 9999,
			},
			expected: &Options{
				HTTPHost:           "0.0.0.0",
				HTTPPort:           9000,
				HTTPMode:           "release",
				HTTPReadTimeout:    30 * time.Second,
				HTTPWriteTimeout:   30 * time.Second,
				HTTPMaxHeaderBytes: 1 << 20,
				GRPCHost:           "0.0.0.0",
				GRPCPort:           9999,
			},
		},
		{
			name: "do not override existing values",
			input: &Options{
				HTTPHost:           "127.0.0.1",
				HTTPPort:           8888,
				HTTPMode:           "debug",
				HTTPReadTimeout:    60 * time.Second,
				HTTPWriteTimeout:   60 * time.Second,
				HTTPMaxHeaderBytes: 2 << 20,
				GRPCHost:           "localhost",
				GRPCPort:           9999,
			},
			expected: &Options{
				HTTPHost:           "127.0.0.1",
				HTTPPort:           8888,
				HTTPMode:           "debug",
				HTTPReadTimeout:    60 * time.Second,
				HTTPWriteTimeout:   60 * time.Second,
				HTTPMaxHeaderBytes: 2 << 20,
				GRPCHost:           "localhost",
				GRPCPort:           9999,
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
			name:    "no options - defaults",
			options: nil,
			expected: &Options{
				HTTPHost:           "0.0.0.0",
				HTTPPort:           8080,
				HTTPMode:           "release",
				HTTPReadTimeout:    30 * time.Second,
				HTTPWriteTimeout:   30 * time.Second,
				HTTPMaxHeaderBytes: 1 << 20,
				GRPCHost:           "0.0.0.0",
				GRPCPort:           9090,
				EnableProfiling:    false,
				EnableMetrics:      true,
			},
		},
		{
			name: "with HTTP port",
			options: []Option{
				WithHTTPPort(9999),
			},
			expected: &Options{
				HTTPHost:           "0.0.0.0",
				HTTPPort:           9999,
				HTTPMode:           "release",
				HTTPReadTimeout:    30 * time.Second,
				HTTPWriteTimeout:   30 * time.Second,
				HTTPMaxHeaderBytes: 1 << 20,
				GRPCHost:           "0.0.0.0",
				GRPCPort:           9090,
				EnableProfiling:    false,
				EnableMetrics:      true,
			},
		},
		{
			name: "with HTTP host",
			options: []Option{
				WithHTTPHost("127.0.0.1"),
			},
			expected: &Options{
				HTTPHost:           "127.0.0.1",
				HTTPPort:           8080,
				HTTPMode:           "release",
				HTTPReadTimeout:    30 * time.Second,
				HTTPWriteTimeout:   30 * time.Second,
				HTTPMaxHeaderBytes: 1 << 20,
				GRPCHost:           "0.0.0.0",
				GRPCPort:           9090,
				EnableProfiling:    false,
				EnableMetrics:      true,
			},
		},
		{
			name: "with gRPC port",
			options: []Option{
				WithGRPCPort(8888),
			},
			expected: &Options{
				HTTPHost:           "0.0.0.0",
				HTTPPort:           8080,
				HTTPMode:           "release",
				HTTPReadTimeout:    30 * time.Second,
				HTTPWriteTimeout:   30 * time.Second,
				HTTPMaxHeaderBytes: 1 << 20,
				GRPCHost:           "0.0.0.0",
				GRPCPort:           8888,
				EnableProfiling:    false,
				EnableMetrics:      true,
			},
		},
		{
			name: "with gRPC host",
			options: []Option{
				WithGRPCHost("localhost"),
			},
			expected: &Options{
				HTTPHost:           "0.0.0.0",
				HTTPPort:           8080,
				HTTPMode:           "release",
				HTTPReadTimeout:    30 * time.Second,
				HTTPWriteTimeout:   30 * time.Second,
				HTTPMaxHeaderBytes: 1 << 20,
				GRPCHost:           "localhost",
				GRPCPort:           9090,
				EnableProfiling:    false,
				EnableMetrics:      true,
			},
		},
		{
			name: "with HTTP mode",
			options: []Option{
				WithHTTPMode("debug"),
			},
			expected: &Options{
				HTTPHost:           "0.0.0.0",
				HTTPPort:           8080,
				HTTPMode:           "debug",
				HTTPReadTimeout:    30 * time.Second,
				HTTPWriteTimeout:   30 * time.Second,
				HTTPMaxHeaderBytes: 1 << 20,
				GRPCHost:           "0.0.0.0",
				GRPCPort:           9090,
				EnableProfiling:    false,
				EnableMetrics:      true,
			},
		},
		{
			name: "with enable profiling",
			options: []Option{
				WithEnableProfiling(true),
			},
			expected: &Options{
				HTTPHost:           "0.0.0.0",
				HTTPPort:           8080,
				HTTPMode:           "release",
				HTTPReadTimeout:    30 * time.Second,
				HTTPWriteTimeout:   30 * time.Second,
				HTTPMaxHeaderBytes: 1 << 20,
				GRPCHost:           "0.0.0.0",
				GRPCPort:           9090,
				EnableProfiling:    true,
				EnableMetrics:      true,
			},
		},
		{
			name: "with enable metrics",
			options: []Option{
				WithEnableMetrics(false),
			},
			expected: &Options{
				HTTPHost:           "0.0.0.0",
				HTTPPort:           8080,
				HTTPMode:           "release",
				HTTPReadTimeout:    30 * time.Second,
				HTTPWriteTimeout:   30 * time.Second,
				HTTPMaxHeaderBytes: 1 << 20,
				GRPCHost:           "0.0.0.0",
				GRPCPort:           9090,
				EnableProfiling:    false,
				EnableMetrics:      false,
			},
		},
		{
			name: "multiple options chained",
			options: []Option{
				WithHTTPHost("127.0.0.1"),
				WithHTTPPort(9999),
				WithHTTPMode("debug"),
				WithGRPCHost("localhost"),
				WithGRPCPort(8888),
				WithEnableProfiling(true),
				WithEnableMetrics(false),
			},
			expected: &Options{
				HTTPHost:           "127.0.0.1",
				HTTPPort:           9999,
				HTTPMode:           "debug",
				HTTPReadTimeout:    30 * time.Second,
				HTTPWriteTimeout:   30 * time.Second,
				HTTPMaxHeaderBytes: 1 << 20,
				GRPCHost:           "localhost",
				GRPCPort:           8888,
				EnableProfiling:    true,
				EnableMetrics:      false,
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
				flag := fs.Lookup("http-port")
				require.NotNil(t, flag)
				assert.Equal(t, "8080", flag.DefValue)

				flag = fs.Lookup("grpc-port")
				require.NotNil(t, flag)
				assert.Equal(t, "9090", flag.DefValue)

				flag = fs.Lookup("http-mode")
				require.NotNil(t, flag)
				assert.Equal(t, "release", flag.DefValue)
			},
		},
		{
			name:     "with prefix",
			prefixes: []string{"server"},
			verify: func(t *testing.T, fs *pflag.FlagSet, opts *Options) {
				flag := fs.Lookup("server-http-port")
				require.NotNil(t, flag)
				assert.Equal(t, "8080", flag.DefValue)

				flag = fs.Lookup("server-grpc-port")
				require.NotNil(t, flag)
				assert.Equal(t, "9090", flag.DefValue)

				flag = fs.Lookup("server-enable-profiling")
				require.NotNil(t, flag)
				assert.Equal(t, "false", flag.DefValue)
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
		"--http-port=9999",
		"--http-host=127.0.0.1",
		"--http-mode=debug",
		"--grpc-port=8888",
		"--grpc-host=localhost",
		"--enable-profiling=true",
		"--enable-metrics=false",
	}

	err := fs.Parse(args)
	require.NoError(t, err)

	assert.Equal(t, "127.0.0.1", opts.HTTPHost)
	assert.Equal(t, 9999, opts.HTTPPort)
	assert.Equal(t, "debug", opts.HTTPMode)
	assert.Equal(t, "localhost", opts.GRPCHost)
	assert.Equal(t, 8888, opts.GRPCPort)
	assert.True(t, opts.EnableProfiling)
	assert.False(t, opts.EnableMetrics)
}

func TestOptionsCompleteAndValidate(t *testing.T) {
	tests := []struct {
		name         string
		input        *Options
		expectValid  bool
		expectChange bool
	}{
		{
			name:         "empty options - complete and validate",
			input:        &Options{},
			expectValid:  true,
			expectChange: true,
		},
		{
			name: "partial options - complete and validate",
			input: &Options{
				HTTPPort: 9999,
			},
			expectValid:  true,
			expectChange: true,
		},
		{
			name: "invalid options - complete but fail validation",
			input: &Options{
				HTTPPort: -1,
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
