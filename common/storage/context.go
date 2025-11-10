package storage

import (
	"context"
	"time"
)

const (
	// DefaultTimeout is the default context timeout for database operations.
	DefaultTimeout = 5 * time.Second
)

// WithTimeout adds timeout to context if it doesn't already have a deadline.
// If ctx already has a deadline, it returns ctx unchanged.
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		// Context already has a deadline, don't add another one
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}

// WithDefaultTimeout adds default timeout (5s) to context if it doesn't have a deadline.
func WithDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return WithTimeout(ctx, DefaultTimeout)
}
