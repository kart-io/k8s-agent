package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/kart-io/k8s-agent/agent-manager/internal/agent"
	"github.com/kart-io/k8s-agent/agent-manager/internal/api"
	"github.com/kart-io/k8s-agent/agent-manager/internal/command"
	"github.com/kart-io/k8s-agent/agent-manager/internal/config"
	"github.com/kart-io/k8s-agent/agent-manager/internal/event"
	"github.com/kart-io/k8s-agent/agent-manager/internal/nats"
	"github.com/kart-io/k8s-agent/agent-manager/internal/storage"
	"github.com/kart-io/k8s-agent/agent-manager/pkg/types"
)

var (
	version = "1.0.0"
)

func main() {
	// Parse command-line flags
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to configuration file (defaults to ./configs/config.yaml)")
	flag.StringVar(&configPath, "c", "", "Path to configuration file (shorthand)")
	flag.Parse()

	// Load configuration from config file
	var cfg *types.Config
	var err error

	if configPath != "" {
		cfg, err = config.LoadFromPath(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load configuration from %s: %v\n", configPath, err)
			os.Exit(1)
		}
		fmt.Printf("Loaded configuration from: %s\n", configPath)
	} else {
		cfg, err = config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
			os.Exit(1)
		}
	}

	// Initialize logger
	logger, err := initLogger(cfg.Logging)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Aetherius Agent Manager",
		zap.String("version", version))

	// Run application
	if err := run(cfg, logger); err != nil {
		logger.Fatal("Application error", zap.Error(err))
	}

	logger.Info("Agent Manager stopped successfully")
}

func run(config *types.Config, logger *zap.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize PostgreSQL storage
	logger.Info("Initializing PostgreSQL storage")
	pgStore, err := storage.NewPostgresStore(config.Database, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}
	defer pgStore.Close()

	// Initialize Redis cache
	logger.Info("Initializing Redis cache")
	redisStore, err := storage.NewRedisStore(config.Redis, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize Redis: %w", err)
	}
	defer redisStore.Close()

	// Initialize agent registry
	logger.Info("Initializing agent registry")
	registry := agent.NewRegistry(pgStore, redisStore, logger)
	if err := registry.Start(ctx); err != nil {
		return fmt.Errorf("failed to start registry: %w", err)
	}
	defer registry.Stop()

	// Initialize event processor
	logger.Info("Initializing event processor")
	eventProcessor := event.NewProcessor(pgStore, redisStore, nil, logger)

	// Initialize NATS server
	logger.Info("Initializing NATS server")
	natsServer := nats.NewServer(config.NATS, registry, eventProcessor, logger)
	if err := natsServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start NATS server: %w", err)
	}
	defer natsServer.Stop()

	// Update event processor with NATS connection
	// eventProcessor.SetNATS(natsServer.GetConnection())

	// Initialize command dispatcher
	logger.Info("Initializing command dispatcher")
	dispatcher := command.NewDispatcher(pgStore, redisStore, registry, natsServer, logger)

	// Initialize API server
	logger.Info("Initializing API server")
	apiServer := api.NewServer(
		config.Server,
		registry,
		eventProcessor,
		dispatcher,
		pgStore,
		redisStore,
		logger,
	)

	// Start API server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := apiServer.Start(); err != nil {
			errChan <- fmt.Errorf("API server error: %w", err)
		}
	}()

	// Wait for interrupt signal or error
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		logger.Error("Application error", zap.Error(err))
		return err
	case sig := <-sigCh:
		logger.Info("Received signal, shutting down", zap.String("signal", sig.String()))
	}

	// Graceful shutdown
	logger.Info("Initiating graceful shutdown")

	// Stop API server
	if err := apiServer.Stop(); err != nil {
		logger.Error("Error stopping API server", zap.Error(err))
	}

	return nil
}

func initLogger(config types.LoggingConfig) (*zap.Logger, error) {
	// Parse log level
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(config.Level)); err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	// Configure encoder
	var encoderConfig zapcore.EncoderConfig
	if config.Format == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Build logger
	zapConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      config.Format != "json",
		Encoding:         config.Format,
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{config.OutputPath},
		ErrorOutputPaths: []string{"stderr"},
	}

	return zapConfig.Build()
}
