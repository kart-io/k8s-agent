package utils

import (
	"errors"
	"testing"
	"time"
)

func TestNewAgentError(t *testing.T) {
	originalErr := errors.New("test error")
	agentErr := NewAgentError("test_operation", originalErr, true)

	if agentErr.Op != "test_operation" {
		t.Errorf("Expected Op to be 'test_operation', got '%s'", agentErr.Op)
	}

	if agentErr.Err != originalErr {
		t.Errorf("Expected Err to be original error")
	}

	if !agentErr.Retryable {
		t.Errorf("Expected Retryable to be true")
	}

	if agentErr.Context == nil {
		t.Errorf("Expected Context to be initialized")
	}
}

func TestAgentErrorWithContext(t *testing.T) {
	agentErr := NewAgentError("test_operation", errors.New("test"), true)
	agentErr.WithContext("cluster_id", "test-cluster")
	agentErr.WithContext("attempt", 1)

	if len(agentErr.Context) != 2 {
		t.Errorf("Expected Context to have 2 entries, got %d", len(agentErr.Context))
	}

	if agentErr.Context["cluster_id"] != "test-cluster" {
		t.Errorf("Expected cluster_id context")
	}
}

func TestAgentErrorError(t *testing.T) {
	agentErr := NewAgentError("connect", ErrNATSConnectionFailed, true)
	errorStr := agentErr.Error()

	if errorStr == "" {
		t.Errorf("Expected non-empty error string")
	}

	// With context
	agentErr.WithContext("endpoint", "nats://localhost:4222")
	errorStr = agentErr.Error()

	if errorStr == "" {
		t.Errorf("Expected non-empty error string with context")
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{
			name:      "AgentError retryable",
			err:       NewAgentError("test", errors.New("test"), true),
			retryable: true,
		},
		{
			name:      "AgentError not retryable",
			err:       NewAgentError("test", errors.New("test"), false),
			retryable: false,
		},
		{
			name:      "NATS connection failed",
			err:       ErrNATSConnectionFailed,
			retryable: true,
		},
		{
			name:      "NATS disconnected",
			err:       ErrNATSDisconnected,
			retryable: true,
		},
		{
			name:      "Publish failed",
			err:       ErrPublishFailed,
			retryable: true,
		},
		{
			name:      "Invalid config",
			err:       ErrInvalidConfig,
			retryable: false,
		},
		{
			name:      "Generic error",
			err:       errors.New("generic error"),
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableError(tt.err)
			if result != tt.retryable {
				t.Errorf("Expected retryable=%v for %s, got %v", tt.retryable, tt.name, result)
			}
		})
	}
}

func TestCategorizeError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorType
	}{
		{
			name:     "Connection error",
			err:      ErrNATSConnectionFailed,
			expected: ErrorTypeConnection,
		},
		{
			name:     "Configuration error",
			err:      ErrInvalidConfig,
			expected: ErrorTypeConfiguration,
		},
		{
			name:     "Kubernetes error",
			err:      ErrK8sAPIFailed,
			expected: ErrorTypeKubernetes,
		},
		{
			name:     "Command error",
			err:      ErrCommandValidationFailed,
			expected: ErrorTypeCommand,
		},
		{
			name:     "Unknown error",
			err:      errors.New("unknown"),
			expected: ErrorTypeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CategorizeError(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRetryWithBackoff_Success(t *testing.T) {
	config := RetryConfig{
		MaxRetries:    3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	attempts := 0
	fn := func() error {
		attempts++
		if attempts < 3 {
			return NewAgentError("test", errors.New("retry"), true)
		}
		return nil
	}

	err := RetryWithBackoff(config, fn)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetryWithBackoff_MaxRetriesExceeded(t *testing.T) {
	config := RetryConfig{
		MaxRetries:    2,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	attempts := 0
	fn := func() error {
		attempts++
		return NewAgentError("test", errors.New("always fail"), true)
	}

	err := RetryWithBackoff(config, fn)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	// MaxRetries + 1 initial attempt
	expectedAttempts := config.MaxRetries + 1
	if attempts != expectedAttempts {
		t.Errorf("Expected %d attempts, got %d", expectedAttempts, attempts)
	}
}

func TestRetryWithBackoff_NonRetryableError(t *testing.T) {
	config := DefaultRetryConfig()

	attempts := 0
	fn := func() error {
		attempts++
		return NewAgentError("test", ErrInvalidConfig, false)
	}

	err := RetryWithBackoff(config, fn)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt for non-retryable error, got %d", attempts)
	}
}

func TestErrorStats(t *testing.T) {
	stats := NewErrorStats()

	if stats.TotalErrors != 0 {
		t.Errorf("Expected initial TotalErrors to be 0")
	}

	// Record connection error
	stats.RecordError(ErrNATSConnectionFailed)
	if stats.TotalErrors != 1 {
		t.Errorf("Expected TotalErrors to be 1")
	}
	if stats.ErrorsByType[ErrorTypeConnection] != 1 {
		t.Errorf("Expected 1 connection error")
	}

	// Record another connection error
	stats.RecordError(ErrNATSDisconnected)
	if stats.TotalErrors != 2 {
		t.Errorf("Expected TotalErrors to be 2")
	}
	if stats.ErrorsByType[ErrorTypeConnection] != 2 {
		t.Errorf("Expected 2 connection errors")
	}

	// Record configuration error
	stats.RecordError(ErrInvalidConfig)
	if stats.TotalErrors != 3 {
		t.Errorf("Expected TotalErrors to be 3")
	}
	if stats.ErrorsByType[ErrorTypeConfiguration] != 1 {
		t.Errorf("Expected 1 configuration error")
	}

	// Check stats
	result := stats.GetStats()
	if result["total_errors"] != int64(3) {
		t.Errorf("Expected total_errors to be 3 in stats")
	}
}

func TestErrorStatsLastError(t *testing.T) {
	stats := NewErrorStats()

	err1 := errors.New("first error")
	stats.RecordError(err1)

	time.Sleep(10 * time.Millisecond)

	err2 := errors.New("second error")
	stats.RecordError(err2)

	if stats.LastError != err2 {
		t.Errorf("Expected LastError to be the second error")
	}

	if stats.LastErrorTime.IsZero() {
		t.Errorf("Expected LastErrorTime to be set")
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries <= 0 {
		t.Errorf("Expected MaxRetries to be positive")
	}

	if config.InitialDelay <= 0 {
		t.Errorf("Expected InitialDelay to be positive")
	}

	if config.MaxDelay <= 0 {
		t.Errorf("Expected MaxDelay to be positive")
	}

	if config.BackoffFactor <= 1.0 {
		t.Errorf("Expected BackoffFactor to be > 1.0")
	}
}