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
	logger.Info("==========================================================")
	logger.Info("     Aetherius Orchestrator Service - Initialization      ")
	logger.Info("==========================================================")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize PostgreSQL
	logger.Info("📦 [1/6] Initializing PostgreSQL",
		zap.String("host", config.Database.Host),
		zap.Int("port", config.Database.Port),
		zap.String("database", config.Database.Database))
	pgStore, err := storage.NewPostgresStore(config.Database, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}
	defer pgStore.Close()
	logger.Info("✅ PostgreSQL initialized successfully")

	// Initialize Redis
	logger.Info("📦 [2/6] Initializing Redis",
		zap.String("addr", config.Redis.Addr))
	redisStore, err := storage.NewRedisStore(config.Redis, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize Redis: %w", err)
	}
	defer redisStore.Close()
	logger.Info("✅ Redis initialized successfully")

	// Connect to NATS
	logger.Info("📡 [3/6] Connecting to NATS",
		zap.String("url", config.NATS.URL))
	natsConn, err := nats.Connect(config.NATS.URL,
		nats.Name("orchestrator-service"),
		nats.MaxReconnects(config.NATS.MaxReconnect),
		nats.ReconnectWait(config.NATS.ReconnectWait))
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}
	defer natsConn.Close()
	logger.Info("✅ NATS connected successfully",
		zap.String("server_info", natsConn.ConnectedUrl()))

	// Initialize workflow components
	logger.Info("⚙️  [4/6] Initializing workflow engine",
		zap.String("agent_manager_url", config.AI.AgentManagerURL),
		zap.String("reasoning_service_url", config.AI.ReasoningServiceURL))
	executor := workflow.NewExecutor(
		config.AI.AgentManagerURL,
		config.AI.ReasoningServiceURL,
		logger)

	engine := workflow.NewEngine(pgStore, redisStore, executor, logger)
	logger.Info("✅ Workflow engine initialized")

	// Initialize strategy manager
	logger.Info("🎯 [5/6] Initializing strategy manager")
	strategyManager := strategy.NewManager(pgStore, engine, logger)
	logger.Info("✅ Strategy manager initialized")

	// Initialize event subscriber
	logger.Info("📬 [6/6] Initializing event subscriber")
	eventSubscriber := subscriber.NewSubscriber(natsConn, strategyManager, logger)
	if err := eventSubscriber.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event subscriber: %w", err)
	}
	defer eventSubscriber.Stop()

	logger.Info("==========================================================")
	logger.Info("✅ Orchestrator Service started successfully!")
	logger.Info("==========================================================")
	logger.Info("🎧 Listening for events on NATS channels:")
	logger.Info("   - internal.event.critical")
	logger.Info("   - internal.event.anomaly")
	logger.Info("   - internal.event.* (debug)")
	logger.Info("==========================================================")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	logger.Info("🛑 Shutdown signal received")
	logger.Info("Shutting down gracefully...")
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
	if agentURL := os.Getenv("AGENT_MANAGER_URL"); agentURL != "" {
		config.AI.AgentManagerURL = agentURL
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
