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

	"github.com/kart-io/k8s-agent/common/server"
	"github.com/kart-io/logger/core"
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

// ServerProvider represents a component that provides a server instance.
// Initializers can implement this interface to register servers that should
// be started automatically by the bootstrap system.
type ServerProvider interface {
	// GetServer returns the server instance from common/server package.
	// Returns nil if no server is provided by this initializer.
	GetServer() server.Server
}

// Bootstrap manages application lifecycle.
type Bootstrap struct {
	initializers []Initializer
	closers      []Closer
	healthChecks []HealthChecker
	logger       core.Logger
	mu           sync.RWMutex
	initialized  bool
}

// New creates a new Bootstrap instance.
func New(logger core.Logger) *Bootstrap {
	if logger == nil {
		// 如果没有提供 logger,使用默认的 noop logger
		logger = core.NewNoOpLogger(nil)
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
		b.logger.Infow("Initializing component",
			"name", init.Name(),
			"priority", init.Priority(),
		)

		start := time.Now()
		if err := init.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize %s: %w", init.Name(), err)
		}

		b.logger.Infow("Component initialized",
			"name", init.Name(),
			"duration", time.Since(start),
		)
	}

	b.mu.Lock()
	b.initialized = true
	b.mu.Unlock()

	b.logger.Infow("All components initialized successfully")
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

		b.logger.Infow("Shutting down component", "name", name)

		start := time.Now()
		if err := closer.Close(ctx); err != nil {
			b.logger.Errorw("Failed to close component",
				"name", name,
				"error", err,
			)
			errors = append(errors, err)
		} else {
			b.logger.Infow("Component closed",
				"name", name,
				"duration", time.Since(start),
			)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown completed with %d errors", len(errors))
	}

	b.logger.Infow("All components shut down successfully")
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

	// Collect all servers from initializers that implement ServerProvider
	var servers []server.Server
	b.mu.RLock()
	for _, init := range b.initializers {
		if provider, ok := init.(ServerProvider); ok {
			if srv := provider.GetServer(); srv != nil {
				servers = append(servers, srv)

				// Log server registration
				name := init.Name()
				b.logger.Infow("Registered server from initializer", "initializer", name)
			}
		}
	}
	b.mu.RUnlock()

	// Start all servers in background
	if len(servers) > 0 {
		b.logger.Infow("Starting servers", "count", len(servers))
		for _, srv := range servers {
			srv := srv // avoid closure issue
			go srv.RunOrDie()
		}
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
			b.logger.Errorw("Application error", "error", err)
		}
	case sig := <-sigChan:
		b.logger.Infow("Received signal", "signal", sig.String())
	}

	// Graceful shutdown
	b.logger.Infow("Starting graceful shutdown...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop all servers first
	if len(servers) > 0 {
		b.logger.Infow("Stopping servers", "count", len(servers))
		for _, srv := range servers {
			go srv.GracefulStop(shutdownCtx)
		}
		// Give servers time to stop
		time.Sleep(1 * time.Second)
	}

	// Then shutdown other components
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
