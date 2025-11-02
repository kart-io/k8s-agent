package contextx

import (
	"context"
	"time"
)

// TimeoutOptions defines options for timeout management.
type TimeoutOptions struct {
	// Default timeout for operations
	DefaultTimeout time.Duration

	// Maximum timeout allowed
	MaxTimeout time.Duration

	// Minimum timeout allowed
	MinTimeout time.Duration
}

// DefaultTimeoutOptions returns default timeout options.
func DefaultTimeoutOptions() *TimeoutOptions {
	return &TimeoutOptions{
		DefaultTimeout: 30 * time.Second,
		MaxTimeout:     5 * time.Minute,
		MinTimeout:     time.Second,
	}
}

// WithTimeout creates a context with timeout, respecting limits.
func (o *TimeoutOptions) WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	// Use default if timeout is zero
	if timeout == 0 {
		timeout = o.DefaultTimeout
	}

	// Enforce minimum timeout
	if timeout < o.MinTimeout {
		timeout = o.MinTimeout
	}

	// Enforce maximum timeout
	if timeout > o.MaxTimeout {
		timeout = o.MaxTimeout
	}

	// Copy context values
	info := ExtractInfo(parent)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	ctx = ApplyInfo(ctx, info)

	return ctx, cancel
}

// WithDeadline creates a context with deadline, respecting limits.
func (o *TimeoutOptions) WithDeadline(parent context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	timeout := time.Until(deadline)
	return o.WithTimeout(parent, timeout)
}

// GetDeadline returns the deadline from context or calculates default.
func GetDeadline(ctx context.Context) time.Time {
	if deadline, ok := ctx.Deadline(); ok {
		return deadline
	}
	return time.Now().Add(30 * time.Second)
}

// GetTimeout returns remaining timeout duration.
func GetTimeout(ctx context.Context) time.Duration {
	if deadline, ok := ctx.Deadline(); ok {
		return time.Until(deadline)
	}
	return 0
}

// HasTimeout checks if context has a deadline.
func HasTimeout(ctx context.Context) bool {
	_, ok := ctx.Deadline()
	return ok
}

// IsExpired checks if context is expired.
func IsExpired(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// WaitWithTimeout waits for operation with timeout.
func WaitWithTimeout(ctx context.Context, timeout time.Duration, fn func() error) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- fn()
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Detach creates a new context that is detached from parent's cancellation
// but retains parent's values.
func Detach(parent context.Context) context.Context {
	info := ExtractInfo(parent)
	return ApplyInfo(context.Background(), info)
}

// DetachWithTimeout creates a detached context with timeout.
func DetachWithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	info := ExtractInfo(parent)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	ctx = ApplyInfo(ctx, info)
	return ctx, cancel
}
