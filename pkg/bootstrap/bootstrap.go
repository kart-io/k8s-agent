// Package bootstrap provides application initialization and startup management.
package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// Initializer represents a component that needs initialization.
type Initializer interface {
	// Name returns the initializer name for logging.
	Name() string

	// Initialize performs the initialization.
	Initialize(ctx context.Context) error

	// Priority returns initialization priority (lower runs first).
	Priority() int
}

// Closer represents a component that needs cleanup.
type Closer interface {
	// Close performs cleanup operations.
	Close(ctx context.Context) error
}

// HealthChecker represents a component that can report health status.
type HealthChecker interface {
	// HealthCheck returns health status.
	HealthCheck(ctx context.Context) error
}

// Bootstrap manages application lifecycle.
type Bootstrap struct {
	initializers []Initializer
	closers      []Closer
	healthChecks []HealthChecker
	logger       *logrus.Logger
	mu           sync.RWMutex
	initialized  bool
}

// New creates a new Bootstrap instance.
func New(logger *logrus.Logger) *Bootstrap {
	if logger == nil {
		logger = logrus.New()
	}

	return &Bootstrap{
		initializers: make([]Initializer, 0),
		closers:      make([]Closer, 0),
		healthChecks: make([]HealthChecker, 0),
		logger:       logger,
	}
}

// Register registers an initializer.
func (b *Bootstrap) Register(init Initializer) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.initializers = append(b.initializers, init)

	// Register as closer if implements Closer interface
	if closer, ok := init.(Closer); ok {
		b.closers = append(b.closers, closer)
	}

	// Register as health checker if implements HealthChecker interface
	if checker, ok := init.(HealthChecker); ok {
		b.healthChecks = append(b.healthChecks, checker)
	}
}

// RegisterCloser registers a closer for cleanup.
func (b *Bootstrap) RegisterCloser(closer Closer) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.closers = append(b.closers, closer)
}

// Initialize initializes all registered components in priority order.
func (b *Bootstrap) Initialize(ctx context.Context) error {
	b.mu.Lock()
	if b.initialized {
		b.mu.Unlock()
		return fmt.Errorf("already initialized")
	}

	// Sort by priority
	sortedInits := make([]Initializer, len(b.initializers))
	copy(sortedInits, b.initializers)
	b.mu.Unlock()

	// Simple bubble sort by priority
	for i := 0; i < len(sortedInits); i++ {
		for j := i + 1; j < len(sortedInits); j++ {
			if sortedInits[i].Priority() > sortedInits[j].Priority() {
				sortedInits[i], sortedInits[j] = sortedInits[j], sortedInits[i]
			}
		}
	}

	// Initialize each component
	for _, init := range sortedInits {
		b.logger.Infof("Initializing %s (priority: %d)", init.Name(), init.Priority())

		start := time.Now()
		if err := init.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize %s: %w", init.Name(), err)
		}

		b.logger.Infof("Initialized %s in %v", init.Name(), time.Since(start))
	}

	b.mu.Lock()
	b.initialized = true
	b.mu.Unlock()

	b.logger.Info("All components initialized successfully")
	return nil
}

// Shutdown gracefully shuts down all components.
func (b *Bootstrap) Shutdown(ctx context.Context) error {
	b.mu.RLock()
	closers := make([]Closer, len(b.closers))
	copy(closers, b.closers)
	b.mu.RUnlock()

	// Close in reverse order
	var errors []error
	for i := len(closers) - 1; i >= 0; i-- {
		closer := closers[i]

		// Try to get name if closer is also an Initializer
		name := fmt.Sprintf("component-%d", i)
		if init, ok := closer.(Initializer); ok {
			name = init.Name()
		}

		b.logger.Infof("Shutting down %s", name)

		start := time.Now()
		if err := closer.Close(ctx); err != nil {
			b.logger.Errorf("Failed to close %s: %v", name, err)
			errors = append(errors, err)
		} else {
			b.logger.Infof("Closed %s in %v", name, time.Since(start))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown completed with %d errors", len(errors))
	}

	b.logger.Info("All components shut down successfully")
	return nil
}

// HealthCheck checks health of all registered components.
func (b *Bootstrap) HealthCheck(ctx context.Context) error {
	b.mu.RLock()
	checkers := make([]HealthChecker, len(b.healthChecks))
	copy(checkers, b.healthChecks)
	b.mu.RUnlock()

	for _, checker := range checkers {
		if err := checker.HealthCheck(ctx); err != nil {
			// Try to get name if checker is also an Initializer
			name := "unknown"
			if init, ok := checker.(Initializer); ok {
				name = init.Name()
			}
			return fmt.Errorf("health check failed for %s: %w", name, err)
		}
	}

	return nil
}

// Run initializes, runs until signal, then shuts down gracefully.
func (b *Bootstrap) Run(ctx context.Context, runFunc func() error) error {
	// Initialize all components
	if err := b.Initialize(ctx); err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run main function in goroutine
	errChan := make(chan error, 1)
	go func() {
		if runFunc != nil {
			errChan <- runFunc()
		} else {
			// Block until signal
			<-sigChan
			errChan <- nil
		}
	}()

	// Wait for completion or signal
	select {
	case err := <-errChan:
		if err != nil {
			b.logger.Errorf("Application error: %v", err)
		}
	case sig := <-sigChan:
		b.logger.Infof("Received signal: %v", sig)
	}

	// Graceful shutdown
	b.logger.Info("Starting graceful shutdown...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := b.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown failed: %w", err)
	}

	return nil
}

// IsInitialized returns whether bootstrap is initialized.
func (b *Bootstrap) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}
