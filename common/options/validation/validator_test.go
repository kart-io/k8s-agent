package validation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name      string
		port      int
		fieldName string
		wantErr   bool
	}{
		{"valid port 80", 80, "server", false},
		{"valid port 8080", 8080, "server", false},
		{"valid port 65535", 65535, "server", false},
		{"invalid port 0", 0, "server", true},
		{"invalid port -1", -1, "server", true},
		{"invalid port 65536", 65536, "server", true},
		{"invalid port 100000", 100000, "server", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePort(tt.port, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateHostPort(t *testing.T) {
	tests := []struct {
		name        string
		host        string
		port        int
		serviceName string
		wantErr     bool
	}{
		{"valid localhost:8080", "localhost", 8080, "server", false},
		{"valid 0.0.0.0:80", "0.0.0.0", 80, "server", false},
		{"empty host", "", 8080, "server", true},
		{"invalid port", "localhost", 0, "server", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHostPort(tt.host, tt.port, tt.serviceName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAddr(t *testing.T) {
	tests := []struct {
		name      string
		addr      string
		fieldName string
		wantErr   bool
	}{
		{"valid localhost:6379", "localhost:6379", "redis", false},
		{"valid 127.0.0.1:3306", "127.0.0.1:3306", "mysql", false},
		{"valid with domain", "example.com:443", "https", false},
		{"empty address", "", "redis", true},
		{"missing port", "localhost", "redis", true},
		{"invalid port", "localhost:abc", "redis", true},
		{"invalid port range", "localhost:99999", "redis", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAddr(tt.addr, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePositiveInt(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		fieldName string
		wantErr   bool
	}{
		{"positive 1", 1, "pool_size", false},
		{"positive 100", 100, "pool_size", false},
		{"zero", 0, "pool_size", true},
		{"negative", -1, "pool_size", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePositiveInt(tt.value, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateNonNegativeInt(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		fieldName string
		wantErr   bool
	}{
		{"positive", 10, "max_conns", false},
		{"zero", 0, "max_conns", false},
		{"negative", -1, "max_conns", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNonNegativeInt(tt.value, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateIntRange(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		min       int
		max       int
		fieldName string
		wantErr   bool
	}{
		{"within range", 5, 1, 10, "db_index", false},
		{"at min", 1, 1, 10, "db_index", false},
		{"at max", 10, 1, 10, "db_index", false},
		{"below min", 0, 1, 10, "db_index", true},
		{"above max", 11, 1, 10, "db_index", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIntRange(tt.value, tt.min, tt.max, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePositiveDuration(t *testing.T) {
	tests := []struct {
		name      string
		duration  time.Duration
		fieldName string
		wantErr   bool
	}{
		{"positive duration", 5 * time.Second, "timeout", false},
		{"zero duration", 0, "timeout", true},
		{"negative duration", -1 * time.Second, "timeout", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePositiveDuration(tt.duration, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDurationRange(t *testing.T) {
	tests := []struct {
		name      string
		duration  time.Duration
		min       time.Duration
		max       time.Duration
		fieldName string
		wantErr   bool
	}{
		{"within range", 5 * time.Second, 1 * time.Second, 10 * time.Second, "timeout", false},
		{"at min", 1 * time.Second, 1 * time.Second, 10 * time.Second, "timeout", false},
		{"at max", 10 * time.Second, 1 * time.Second, 10 * time.Second, "timeout", false},
		{"below min", 0, 1 * time.Second, 10 * time.Second, "timeout", true},
		{"above max", 11 * time.Second, 1 * time.Second, 10 * time.Second, "timeout", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDurationRange(tt.duration, tt.min, tt.max, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		wantErr   bool
	}{
		{"valid string", "test", "username", false},
		{"empty string", "", "username", true},
		{"whitespace only", "   ", "username", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequired(tt.value, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEnum(t *testing.T) {
	allowedModes := []string{"debug", "release", "test"}

	tests := []struct {
		name          string
		value         string
		fieldName     string
		allowedValues []string
		wantErr       bool
	}{
		{"valid debug", "debug", "mode", allowedModes, false},
		{"valid release", "release", "mode", allowedModes, false},
		{"invalid mode", "production", "mode", allowedModes, true},
		{"empty value", "", "mode", allowedModes, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnum(tt.value, tt.fieldName, tt.allowedValues)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateConnectionPool(t *testing.T) {
	tests := []struct {
		name        string
		maxConns    int
		minConns    int
		serviceName string
		wantErr     bool
	}{
		{"valid pool", 100, 10, "database", false},
		{"equal conns", 10, 10, "database", false},
		{"min > max", 20, 100, "database", true},
		{"negative max", -1, 10, "database", true},
		{"negative min", 100, -1, "database", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConnectionPool(tt.maxConns, tt.minConns, tt.serviceName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		fieldName string
		wantErr   bool
	}{
		{"valid http", "http://example.com", "url", false},
		{"valid https", "https://example.com:443", "url", false},
		{"valid nats", "nats://localhost:4222", "nats_url", false},
		{"valid tcp", "tcp://localhost:3306", "db_url", false},
		{"empty url", "", "url", true},
		{"invalid scheme", "ftp://example.com", "url", true},
		{"no scheme", "example.com", "url", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRedisDB(t *testing.T) {
	tests := []struct {
		name    string
		db      int
		wantErr bool
	}{
		{"valid 0", 0, false},
		{"valid 8", 8, false},
		{"valid 15", 15, false},
		{"invalid -1", -1, true},
		{"invalid 16", 16, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRedisDB(tt.db)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTimeouts(t *testing.T) {
	tests := []struct {
		name        string
		dial        time.Duration
		read        time.Duration
		write       time.Duration
		serviceName string
		wantErr     bool
	}{
		{"valid timeouts", 5 * time.Second, 3 * time.Second, 3 * time.Second, "redis", false},
		{"zero dial", 0, 3 * time.Second, 3 * time.Second, "redis", true},
		{"zero read", 5 * time.Second, 0, 3 * time.Second, "redis", true},
		{"zero write", 5 * time.Second, 3 * time.Second, 0, "redis", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimeouts(tt.dial, tt.read, tt.write, tt.serviceName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMaxValue(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		max       int
		fieldName string
		wantErr   bool
	}{
		{"below max", 50, 100, "count", false},
		{"at max", 100, 100, "count", false},
		{"above max", 101, 100, "count", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMaxValue(tt.value, tt.max, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMinValue(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		min       int
		fieldName string
		wantErr   bool
	}{
		{"above min", 50, 10, "count", false},
		{"at min", 10, 10, "count", false},
		{"below min", 5, 10, "count", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMinValue(tt.value, tt.min, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
