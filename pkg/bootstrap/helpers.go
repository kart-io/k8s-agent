package bootstrap

import (
	"context"
	"time"
)

// Priority constants for common initialization orders.
const (
	PriorityConfig      = 100  // Configuration loading
	PriorityLogger      = 200  // Logging setup
	PriorityDatabase    = 300  // Database connections
	PriorityRedis       = 400  // Redis connections (alias for Cache)
	PriorityCache       = 400  // Cache connections
	PriorityNATS        = 500  // NATS message queue
	PriorityMQ          = 500  // Message queue connections
	PriorityHTTP        = 1000 // HTTP server
	PriorityGRPC        = 1100 // gRPC server
	PriorityApplication = 1500 // Application logic (for background tasks)
	PriorityHealthCheck = 2000 // Health check server (should initialize last)
	PriorityLowest      = 2000 // Components that should initialize last
)

// FuncInitializer wraps a function as an Initializer.
type FuncInitializer struct {
	name      string
	priority  int
	initFunc  func(context.Context) error
	closeFunc func(context.Context) error
}

// NewFuncInitializer creates a new function-based initializer.
func NewFuncInitializer(name string, priority int, initFunc func(context.Context) error, closeFunc func(context.Context) error) *FuncInitializer {
	return &FuncInitializer{
		name:      name,
		priority:  priority,
		initFunc:  initFunc,
		closeFunc: closeFunc,
	}
}

// Name returns the initializer name.
func (f *FuncInitializer) Name() string {
	return f.name
}

// Initialize performs initialization.
func (f *FuncInitializer) Initialize(ctx context.Context) error {
	if f.initFunc != nil {
		return f.initFunc(ctx)
	}
	return nil
}

// Priority returns initialization priority.
func (f *FuncInitializer) Priority() int {
	return f.priority
}

// Close performs cleanup.
func (f *FuncInitializer) Close(ctx context.Context) error {
	if f.closeFunc != nil {
		return f.closeFunc(ctx)
	}
	return nil
}

// SimpleCloser wraps a close function as a Closer.
type SimpleCloser struct {
	closeFunc func(context.Context) error
}

// NewSimpleCloser creates a new simple closer.
func NewSimpleCloser(closeFunc func(context.Context) error) *SimpleCloser {
	return &SimpleCloser{closeFunc: closeFunc}
}

// Close performs cleanup.
func (s *SimpleCloser) Close(ctx context.Context) error {
	if s.closeFunc != nil {
		return s.closeFunc(ctx)
	}
	return nil
}

// RetryInitializer wraps an initializer with retry logic.
type RetryInitializer struct {
	wrapped    Initializer
	maxRetries int
	retryDelay time.Duration
}

// NewRetryInitializer creates an initializer with retry logic.
func NewRetryInitializer(wrapped Initializer, maxRetries int, retryDelay time.Duration) *RetryInitializer {
	return &RetryInitializer{
		wrapped:    wrapped,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
	}
}

// Name returns the initializer name.
func (r *RetryInitializer) Name() string {
	return r.wrapped.Name()
}

// Initialize performs initialization with retries.
func (r *RetryInitializer) Initialize(ctx context.Context) error {
	var lastErr error

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(r.retryDelay):
			}
		}

		if err := r.wrapped.Initialize(ctx); err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	return lastErr
}

// Priority returns initialization priority.
func (r *RetryInitializer) Priority() int {
	return r.wrapped.Priority()
}

// Close performs cleanup if wrapped implements Closer.
func (r *RetryInitializer) Close(ctx context.Context) error {
	if closer, ok := r.wrapped.(Closer); ok {
		return closer.Close(ctx)
	}
	return nil
}

// HealthCheck performs health check if wrapped implements HealthChecker.
func (r *RetryInitializer) HealthCheck(ctx context.Context) error {
	if checker, ok := r.wrapped.(HealthChecker); ok {
		return checker.HealthCheck(ctx)
	}
	return nil
}
