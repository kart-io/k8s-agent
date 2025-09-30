package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"

	"github.com/kart-io/k8s-agent/orchestrator-service/internal/storage"
	"github.com/kart-io/k8s-agent/orchestrator-service/internal/strategy"
	"github.com/kart-io/k8s-agent/orchestrator-service/internal/subscriber"
	"github.com/kart-io/k8s-agent/orchestrator-service/internal/workflow"
	"github.com/kart-io/k8s-agent/orchestrator-service/pkg/types"
)

var (
	configFile = flag.String("config", "configs/config.yaml", "Path to configuration file")
	version    = "1.0.0"
)

func main() {
	flag.Parse()

	config, err := loadConfig(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	logger, err := initLogger(config.Logging)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Aetherius Orchestrator Service",
		zap.String("version", version))

	if err := run(config, logger); err != nil {
		logger.Fatal("Application error", zap.Error(err))
	}

	logger.Info("Orchestrator Service stopped successfully")
}

func run(config *types.Config, logger *zap.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize PostgreSQL
	logger.Info("Initializing PostgreSQL")
	pgStore, err := storage.NewPostgresStore(config.Database, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}
	defer pgStore.Close()

	// Initialize Redis
	logger.Info("Initializing Redis")
	redisStore, err := storage.NewRedisStore(config.Redis, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize Redis: %w", err)
	}
	defer redisStore.Close()

	// Connect to NATS
	logger.Info("Connecting to NATS")
	natsConn, err := nats.Connect(config.NATS.URL,
		nats.Name("orchestrator-service"),
		nats.MaxReconnects(config.NATS.MaxReconnect),
		nats.ReconnectWait(config.NATS.ReconnectWait))
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}
	defer natsConn.Close()

	// Initialize workflow components
	logger.Info("Initializing workflow engine")
	executor := workflow.NewExecutor(
		"http://agent-manager:8080",
		config.AI.ReasoningServiceURL,
		logger)

	engine := workflow.NewEngine(pgStore, redisStore, executor, logger)

	// Initialize strategy manager
	logger.Info("Initializing strategy manager")
	strategyManager := strategy.NewManager(pgStore, engine, logger)

	// Initialize event subscriber
	logger.Info("Initializing event subscriber")
	eventSubscriber := subscriber.NewSubscriber(natsConn, strategyManager, logger)
	if err := eventSubscriber.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event subscriber: %w", err)
	}
	defer eventSubscriber.Stop()

	logger.Info("Orchestrator Service started successfully")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	logger.Info("Shutting down gracefully")
	return nil
}

func loadConfig(path string) (*types.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config types.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	applyEnvOverrides(&config)
	return &config, nil
}

func applyEnvOverrides(config *types.Config) {
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		config.Database.Host = dbHost
	}
	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		fmt.Sscanf(dbPort, "%d", &config.Database.Port)
	}
	if natsURL := os.Getenv("NATS_URL"); natsURL != "" {
		config.NATS.URL = natsURL
	}
	if aiURL := os.Getenv("AI_SERVICE_URL"); aiURL != "" {
		config.AI.ReasoningServiceURL = aiURL
	}
}

func initLogger(config types.LoggingConfig) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(config.Level)); err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	var encoderConfig zapcore.EncoderConfig
	if config.Format == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

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