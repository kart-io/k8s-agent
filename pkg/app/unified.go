// Package app provides a unified application framework for all services.
// This simplified version consolidates the previous three modes (Simple, Runner, Bootstrap)
// into a single, consistent approach.
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
	"github.com/kart-io/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Application defines the unified application interface
type Application interface {
	// Name returns the application name
	Name() string

	// Initialize initializes the application
	Initialize(ctx context.Context, opts Options) error

	// Run runs the application (blocks until shutdown)
	Run(ctx context.Context) error

	// Shutdown gracefully shuts down the application
	Shutdown(ctx context.Context) error
}

// Options defines application configuration options
type Options interface {
	// Complete completes the configuration with defaults
	Complete() error
	// Validate validates the configuration
	Validate() []error
	// AddFlags adds command-line flags
	AddFlags(fs *pflag.FlagSet)
}

// Config defines the application configuration
type Config struct {
	// Use is the command use string
	Use string
	// Short is the short description
	Short string
	// Long is the long description
	Long string
	// EnvPrefix is the environment variable prefix
	EnvPrefix string
}

// Run runs an application with the given configuration
func Run(app Application, opts Options, cfg Config) {
	cmd := &cobra.Command{
		Use:   cfg.Use,
		Short: cfg.Short,
		Long:  cfg.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Print version if requested
			version.PrintAndExitIfRequested()

			// Complete configuration
			if err := opts.Complete(); err != nil {
				return fmt.Errorf("failed to complete config: %w", err)
			}

			// Validate configuration
			if errs := opts.Validate(); errs != nil && len(errs) > 0 {
				return fmt.Errorf("config validation failed: %v", errs)
			}

			// Initialize logger
			logger := initLogger(app.Name())

			// Create context with cancel
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Setup signal handling
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			// Initialize application
			logger.Infow("Initializing application", "app", app.Name())
			if err := app.Initialize(ctx, opts); err != nil {
				return fmt.Errorf("failed to initialize: %w", err)
			}

			// Start shutdown handler
			go func() {
				<-sigChan
				logger.Infow("Received shutdown signal")
				cancel()

				// Graceful shutdown with timeout
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer shutdownCancel()

				if err := app.Shutdown(shutdownCtx); err != nil {
					logger.Errorw("Shutdown error", "error", err)
				}
			}()

			// Run application (blocks)
			logger.Infow("Starting application", "app", app.Name())
			if err := app.Run(ctx); err != nil {
				return fmt.Errorf("application error: %w", err)
			}

			logger.Infow("Application stopped")
			return nil
		},
	}

	// Add flags
	version.AddFlags(cmd.Flags())
	opts.AddFlags(cmd.Flags())

	// Execute
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// RunWithBootstrap runs an application with bootstrap support
func RunWithBootstrap(app Application, opts Options, cfg Config, registrar func(*bootstrap.Bootstrap) error) {
	// Wrap the application with bootstrap
	wrappedApp := &bootstrapApp{
		app:       app,
		registrar: registrar,
	}

	Run(wrappedApp, opts, cfg)
}

// bootstrapApp wraps an application with bootstrap support
type bootstrapApp struct {
	app       Application
	bootstrap *bootstrap.Bootstrap
	registrar func(*bootstrap.Bootstrap) error
}

func (b *bootstrapApp) Name() string {
	return b.app.Name()
}

func (b *bootstrapApp) Initialize(ctx context.Context, opts Options) error {
	// Create bootstrap
	b.bootstrap = bootstrap.New(nil)

	// Register components
	if b.registrar != nil {
		if err := b.registrar(b.bootstrap); err != nil {
			return fmt.Errorf("failed to register components: %w", err)
		}
	}

	// Initialize bootstrap
	if err := b.bootstrap.Initialize(ctx); err != nil {
		return fmt.Errorf("bootstrap initialization failed: %w", err)
	}

	// Initialize app
	return b.app.Initialize(ctx, opts)
}

func (b *bootstrapApp) Run(ctx context.Context) error {
	// Bootstrap handles running servers
	return b.bootstrap.Run(ctx, func() error {
		// Run the wrapped application
		return b.app.Run(ctx)
	})
}

func (b *bootstrapApp) Shutdown(ctx context.Context) error {
	// Shutdown bootstrap
	if err := b.bootstrap.Shutdown(ctx); err != nil {
		return err
	}

	// Shutdown app
	return b.app.Shutdown(ctx)
}

// initLogger initializes a basic logger
func initLogger(appName string) core.Logger {
	// This is a simplified version
	// In real implementation, use configuration to setup logger
	return core.NewNoOpLogger(nil)
}