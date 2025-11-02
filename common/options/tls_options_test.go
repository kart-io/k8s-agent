package options

import (
	"crypto/tls"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTLSOptions(t *testing.T) {
	opts := NewTLSOptions()

	assert.NotNil(t, opts)
	assert.False(t, opts.UseTLS)
	assert.False(t, opts.InsecureSkipVerify)
	assert.Equal(t, "", opts.CaCert)
	assert.Equal(t, "", opts.Cert)
	assert.Equal(t, "", opts.Key)
	assert.Equal(t, uint16(tls.VersionTLS12), opts.MinVersion)
	assert.Equal(t, uint16(tls.VersionTLS13), opts.MaxVersion)
}

func TestTLSOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    *TLSOptions
		wantErr bool
	}{
		{
			name:    "valid default options (TLS disabled)",
			opts:    NewTLSOptions(),
			wantErr: false,
		},
		{
			name: "TLS disabled - skip validation",
			opts: &TLSOptions{
				UseTLS: false,
			},
			wantErr: false,
		},
		{
			name: "cert without key",
			opts: &TLSOptions{
				UseTLS: true,
				Cert:   "/tmp/cert.pem",
				Key:    "",
			},
			wantErr: true,
		},
		{
			name: "key without cert",
			opts: &TLSOptions{
				UseTLS: true,
				Cert:   "",
				Key:    "/tmp/key.pem",
			},
			wantErr: true,
		},
		{
			name: "min_version > max_version",
			opts: &TLSOptions{
				UseTLS:     true,
				MinVersion: tls.VersionTLS13,
				MaxVersion: tls.VersionTLS12,
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

func TestTLSOptions_Complete(t *testing.T) {
	tests := []struct {
		name     string
		opts     *TLSOptions
		expected *TLSOptions
	}{
		{
			name: "complete with default TLS versions",
			opts: &TLSOptions{
				UseTLS:     true,
				MinVersion: 0,
				MaxVersion: 0,
			},
			expected: &TLSOptions{
				UseTLS:     true,
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS13,
			},
		},
		{
			name: "disabled - skip completion",
			opts: &TLSOptions{
				UseTLS: false,
			},
			expected: &TLSOptions{
				UseTLS: false,
			},
		},
		{
			name: "already has TLS versions",
			opts: &TLSOptions{
				UseTLS:     true,
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS13,
			},
			expected: &TLSOptions{
				UseTLS:     true,
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS13,
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

func TestTLSOptions_WithFunctions(t *testing.T) {
	opts := NewTLSOptions()

	// Test WithUseTLS
	WithUseTLS(true)(opts)
	assert.True(t, opts.UseTLS)

	// Test WithInsecureSkipVerify
	WithInsecureSkipVerify(true)(opts)
	assert.True(t, opts.InsecureSkipVerify)

	// Test WithCACert
	WithCACert("/path/to/ca.pem")(opts)
	assert.Equal(t, "/path/to/ca.pem", opts.CaCert)

	// Test WithTLSCert
	WithTLSCert("/path/to/cert.pem")(opts)
	assert.Equal(t, "/path/to/cert.pem", opts.Cert)

	// Test WithTLSKey
	WithTLSKey("/path/to/key.pem")(opts)
	assert.Equal(t, "/path/to/key.pem", opts.Key)

	// Test WithTLSMinVersion
	WithTLSMinVersion(tls.VersionTLS13)(opts)
	assert.Equal(t, uint16(tls.VersionTLS13), opts.MinVersion)

	// Test WithTLSMaxVersion
	WithTLSMaxVersion(tls.VersionTLS13)(opts)
	assert.Equal(t, uint16(tls.VersionTLS13), opts.MaxVersion)
}

func TestTLSOptions_ApplyTo(t *testing.T) {
	opts := &TLSOptions{
		UseTLS:             true,
		InsecureSkipVerify: false,
		CaCert:             "/path/to/ca.pem",
		Cert:               "/path/to/cert.pem",
		Key:                "/path/to/key.pem",
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
	}

	var target []interface{}
	err := opts.ApplyTo(&target)

	require.NoError(t, err)
	require.Len(t, target, 1)

	config, ok := target[0].(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, true, config["useTLS"])
	assert.Equal(t, false, config["insecureSkipVerify"])
	assert.Equal(t, "/path/to/ca.pem", config["caCert"])
	assert.Equal(t, "/path/to/cert.pem", config["cert"])
	assert.Equal(t, "/path/to/key.pem", config["key"])
	assert.Equal(t, uint16(tls.VersionTLS12), config["minVersion"])
	assert.Equal(t, uint16(tls.VersionTLS13), config["maxVersion"])
}

func TestTLSOptions_ApplyTo_Nil(t *testing.T) {
	opts := NewTLSOptions()
	err := opts.ApplyTo(nil)
	assert.NoError(t, err)
}

func TestTLSOptions_Scheme(t *testing.T) {
	tests := []struct {
		name     string
		useTLS   bool
		expected string
	}{
		{
			name:     "TLS enabled - https",
			useTLS:   true,
			expected: "https",
		},
		{
			name:     "TLS disabled - http",
			useTLS:   false,
			expected: "http",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &TLSOptions{UseTLS: tt.useTLS}
			assert.Equal(t, tt.expected, opts.Scheme())
		})
	}
}

func TestTLSOptions_TLSConfig_Disabled(t *testing.T) {
	opts := &TLSOptions{
		UseTLS: false,
	}

	config, err := opts.TLSConfig()
	assert.NoError(t, err)
	assert.Nil(t, config)
}

func TestTLSOptions_TLSConfig_Basic(t *testing.T) {
	opts := &TLSOptions{
		UseTLS:             true,
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
	}

	config, err := opts.TLSConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.True(t, config.InsecureSkipVerify)
	assert.Equal(t, uint16(tls.VersionTLS12), config.MinVersion)
	assert.Equal(t, uint16(tls.VersionTLS13), config.MaxVersion)
}

func TestTLSOptions_MustTLSConfig(t *testing.T) {
	opts := &TLSOptions{
		UseTLS:             true,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
	}

	config := opts.MustTLSConfig()
	assert.NotNil(t, config)
}

// TestTLSOptions_TLSConfig_WithCerts 测试加载证书
// 注意：这个测试需要临时证书文件，在实际环境中运行
func TestTLSOptions_TLSConfig_WithCerts(t *testing.T) {
	t.Skip("Skipping test that requires actual cert files")

	// 这是一个示例，展示如何测试证书加载
	// 在实际使用中，你需要准备测试证书文件

	// 创建临时目录
	tmpDir := t.TempDir()

	// 创建测试证书文件路径
	certFile := filepath.Join(tmpDir, "cert.pem")
	keyFile := filepath.Join(tmpDir, "key.pem")
	caFile := filepath.Join(tmpDir, "ca.pem")

	// 这里应该生成或复制测试证书
	// generateTestCertificates(certFile, keyFile, caFile)

	opts := &TLSOptions{
		UseTLS: true,
		Cert:   certFile,
		Key:    keyFile,
		CaCert: caFile,
	}

	config, err := opts.TLSConfig()
	if err != nil {
		t.Logf("Expected error in test environment: %v", err)
		return
	}

	assert.NotNil(t, config)
}
