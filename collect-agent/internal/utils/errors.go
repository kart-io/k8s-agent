package utils

import (
	"errors"
	"fmt"
	"time"
)

// Common error types for the agent
var (
	// Connection errors
	ErrNATSConnectionFailed  = errors.New("failed to connect to NATS")
	ErrNATSDisconnected      = errors.New("NATS connection lost")
	ErrSubscriptionFailed    = errors.New("failed to subscribe to subject")
	ErrPublishFailed         = errors.New("failed to publish message")

	// Configuration errors
	ErrInvalidConfig         = errors.New("invalid configuration")
	ErrMissingClusterID      = errors.New("cluster ID is required")
	ErrMissingEndpoint       = errors.New("central endpoint is required")
	ErrInvalidInterval       = errors.New("invalid interval value")

	// K8s API errors
	ErrK8sAPIFailed          = errors.New("kubernetes API call failed")
	ErrResourceNotFound      = errors.New("kubernetes resource not found")
	ErrInsufficientPermission = errors.New("insufficient RBAC permissions")

	// Command execution errors
	ErrCommandValidationFailed = errors.New("command validation failed")
	ErrCommandExecutionFailed  = errors.New("command execution failed")
	ErrCommandTimeout          = errors.New("command execution timeout")
	ErrUnsafeCommand           = errors.New("unsafe command detected")

	// Internal errors
	ErrChannelClosed         = errors.New("channel closed")
	ErrQueueFull             = errors.New("queue is full")
	ErrShutdown              = errors.New("agent is shutting down")
)

// AgentError wraps errors with additional context
type AgentError struct {
	Op        string    // Operation that failed
	Err       error     // Original error
	Timestamp time.Time // When the error occurred
	Retryable bool      // Whether this error is retryable
	Context   map[string]interface{} // Additional context
}

// Error implements the error interface
func (e *AgentError) Error() string {
	if len(e.Context) > 0 {
		return fmt.Sprintf("%s: %v (context: %+v)", e.Op, e.Err, e.Context)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Unwrap returns the underlying error
func (e *AgentError) Unwrap() error {
	return e.Err
}

// IsRetryable returns whether this error is retryable
func (e *AgentError) IsRetryable() bool {
	return e.Retryable
}

// NewAgentError creates a new AgentError
func NewAgentError(op string, err error, retryable bool) *AgentError {
	return &AgentError{
		Op:        op,
		Err:       err,
		Timestamp: time.Now(),
		Retryable: retryable,
		Context:   make(map[string]interface{}),
	}
}

// WithContext adds context to an error
func (e *AgentError) WithContext(key string, value interface{}) *AgentError {
	e.Context[key] = value
	return e
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	var agentErr *AgentError
	if errors.As(err, &agentErr) {
		return agentErr.IsRetryable()
	}

	// Network errors are generally retryable
	if errors.Is(err, ErrNATSConnectionFailed) ||
		errors.Is(err, ErrNATSDisconnected) ||
		errors.Is(err, ErrPublishFailed) {
		return true
	}

	return false
}

// ErrorType represents different error categories
type ErrorType int

const (
	ErrorTypeConnection ErrorType = iota
	ErrorTypeConfiguration
	ErrorTypeKubernetes
	ErrorTypeCommand
	ErrorTypeInternal
)

// CategorizeError determines the error category
func CategorizeError(err error) ErrorType {
	switch {
	case errors.Is(err, ErrNATSConnectionFailed),
		errors.Is(err, ErrNATSDisconnected),
		errors.Is(err, ErrSubscriptionFailed),
		errors.Is(err, ErrPublishFailed):
		return ErrorTypeConnection

	case errors.Is(err, ErrInvalidConfig),
		errors.Is(err, ErrMissingClusterID),
		errors.Is(err, ErrMissingEndpoint),
		errors.Is(err, ErrInvalidInterval):
		return ErrorTypeConfiguration

	case errors.Is(err, ErrK8sAPIFailed),
		errors.Is(err, ErrResourceNotFound),
		errors.Is(err, ErrInsufficientPermission):
		return ErrorTypeKubernetes

	case errors.Is(err, ErrCommandValidationFailed),
		errors.Is(err, ErrCommandExecutionFailed),
		errors.Is(err, ErrCommandTimeout),
		errors.Is(err, ErrUnsafeCommand):
		return ErrorTypeCommand

	default:
		return ErrorTypeInternal
	}
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries     int
	InitialDelay   time.Duration
	MaxDelay       time.Duration
	BackoffFactor  float64
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
	}
}

// RetryWithBackoff executes a function with exponential backoff
func RetryWithBackoff(config RetryConfig, fn func() error) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(delay)
			// Exponential backoff
			delay = time.Duration(float64(delay) * config.BackoffFactor)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry if not retryable
		if !IsRetryableError(err) {
			return err
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// ErrorStats tracks error statistics
type ErrorStats struct {
	TotalErrors    int64
	LastError      error
	LastErrorTime  time.Time
	ErrorsByType   map[ErrorType]int64
}

// NewErrorStats creates a new error statistics tracker
func NewErrorStats() *ErrorStats {
	return &ErrorStats{
		ErrorsByType: make(map[ErrorType]int64),
	}
}

// RecordError records an error occurrence
func (s *ErrorStats) RecordError(err error) {
	s.TotalErrors++
	s.LastError = err
	s.LastErrorTime = time.Now()

	errorType := CategorizeError(err)
	s.ErrorsByType[errorType]++
}

// GetStats returns error statistics
func (s *ErrorStats) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_errors":    s.TotalErrors,
		"last_error":      fmt.Sprintf("%v", s.LastError),
		"last_error_time": s.LastErrorTime,
		"errors_by_type":  s.ErrorsByType,
	}
}