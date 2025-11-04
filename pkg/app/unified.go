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

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	commoncore "github.com/kart-io/k8s-agent/common/core"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger"
	"github.com/kart-io/logger/core"
	"github.com/kart-io/version"
)

// Application defines the unified application interface.
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

// Options defines application configuration options.
type Options interface {
	// Complete completes the configuration with defaults
	Complete() error
	// Validate validates the configuration
	Validate() []error
	// AddFlags adds command-line flags
	AddFlags(fs *pflag.FlagSet)
}

// Config defines the application configuration.
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

// Run runs an application with the given configuration.
func Run(app Application, opts Options, cfg Config) {
	var configFile string

	cmd := &cobra.Command{
		Use:           cfg.Use,
		Short:         cfg.Short,
		Long:          cfg.Long,
		SilenceUsage:  true, // 出错时不显示帮助文档
		SilenceErrors: true, // 我们自己用 logger 打印错误
		RunE: func(cmd *cobra.Command, args []string) error {
			// Print version if requested
			version.PrintAndExitIfRequested()

			// Initialize a basic logger first for error reporting
			logger := initLogger(app.Name())

			// Load config file if specified
			if configFile != "" {
				// 检查 opts 是否实现了 commoncore.LoadableConfig 接口
				if loadable, ok := opts.(commoncore.LoadableConfig); ok {
					if err := commoncore.LoadOptions(loadable, configFile, nil); err != nil {
						logger.Errorw("Failed to load config file", "error", err, "config_file", configFile)
						return fmt.Errorf("failed to load config file: %w", err)
					}
				} else {
					logger.Errorw("Options does not implement LoadableConfig interface")
					return fmt.Errorf("options does not implement LoadableConfig interface")
				}
			} else {
				// Complete configuration
				if err := opts.Complete(); err != nil {
					logger.Errorw("Failed to complete config", "error", err)
					return fmt.Errorf("failed to complete config: %w", err)
				}

				// Validate configuration
				if errs := opts.Validate(); len(errs) > 0 {
					logger.Errorw("Config validation failed", "errors", errs)
					return fmt.Errorf("config validation failed: %v", errs)
				}
			}

			// Create context with cancel
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Setup signal handling
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			// Initialize application
			logger.Infow("Initializing application", "app", app.Name())
			if err := app.Initialize(ctx, opts); err != nil {
				logger.Errorw("Application initialization failed", "error", err, "app", app.Name())
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
				logger.Errorw("Application error", "error", err, "app", app.Name())
				return fmt.Errorf("application error: %w", err)
			}

			logger.Infow("Application stopped")
			return nil
		},
	}

	// Add --config flag
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file (YAML)")

	// Add flags
	version.AddFlags(cmd.Flags())
	opts.AddFlags(cmd.Flags())

	// Execute command
	// Note: SilenceErrors is true, so we handle error logging ourselves
	if err := cmd.Execute(); err != nil {
		// Error has already been logged by the RunE function
		os.Exit(1)
	}
}

// RunWithBootstrap runs an application with bootstrap support.
func RunWithBootstrap(app Application, opts Options, cfg Config, registrar func(*bootstrap.Bootstrap) error) {
	// Wrap the application with bootstrap
	wrappedApp := &bootstrapApp{
		app:       app,
		registrar: registrar,
	}

	Run(wrappedApp, opts, cfg)
}

// bootstrapApp wraps an application with bootstrap support.
type bootstrapApp struct {
	app       Application
	bootstrap *bootstrap.Bootstrap
	registrar func(*bootstrap.Bootstrap) error
}

func (b *bootstrapApp) Name() string {
	return b.app.Name()
}

func (b *bootstrapApp) Initialize(ctx context.Context, opts Options) error {
	// Initialize app first so it can set up its internal state
	if err := b.app.Initialize(ctx, opts); err != nil {
		return fmt.Errorf("app initialization failed: %w", err)
	}

	// Create bootstrap
	b.bootstrap = bootstrap.New(nil)

	// Register components (app is now initialized, so opts/logger are available)
	if b.registrar != nil {
		if err := b.registrar(b.bootstrap); err != nil {
			return fmt.Errorf("failed to register components: %w", err)
		}
	}

	// Note: Do NOT call b.bootstrap.Initialize(ctx) here!
	// Bootstrap.Run() will call Initialize() automatically

	return nil
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

// initLogger initializes a basic logger.
func initLogger(appName string) core.Logger {
	// Create a basic logger for startup/error messages
	// This is used before the full logger is configured
	logger, err := logger.NewWithDefaults()
	if err != nil {
		// Fallback to stderr if logger creation fails
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		return core.NewNoOpLogger(nil)
	}
	return logger
}
