package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/kart-io/k8s-agent/orchestrator-service/internal/config"
	"github.com/kart-io/k8s-agent/orchestrator-service/internal/storage"
	"github.com/kart-io/k8s-agent/orchestrator-service/internal/strategy"
	"github.com/kart-io/k8s-agent/orchestrator-service/internal/subscriber"
	"github.com/kart-io/k8s-agent/orchestrator-service/internal/workflow"
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
	var cfg *config.Config
	var err error

	if configPath != "" {
		cfg, err = config.LoadFromPath(configPath)
		if err != nil {
			log.Fatalf("Failed to load configuration from %s: %v", configPath, err)
		}
		log.Printf("Loaded configuration from: %s", configPath)
	} else {
		cfg, err = config.Load()
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
	}

	logger, err := initLogger(cfg.Logging)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Aetherius Orchestrator Service",
		zap.String("version", version))

	if err := run(cfg, logger); err != nil {
		logger.Fatal("Application error", zap.Error(err))
	}

	logger.Info("Orchestrator Service stopped successfully")
}

func run(cfg *config.Config, logger *zap.Logger) error {
	logger.Info("==========================================================")
	logger.Info("     Aetherius Orchestrator Service - Initialization      ")
	logger.Info("==========================================================")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize PostgreSQL
	logger.Info("📦 [1/6] Initializing PostgreSQL",
		zap.String("host", cfg.Database.Host),
		zap.Int("port", cfg.Database.Port),
		zap.String("database", cfg.Database.Database))
	pgStore, err := storage.NewPostgresStore(cfg.Database, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}
	defer pgStore.Close()
	logger.Info("✅ PostgreSQL initialized successfully")

	// Initialize Redis
	logger.Info("📦 [2/6] Initializing Redis",
		zap.String("addr", cfg.Redis.Addr))
	redisStore, err := storage.NewRedisStore(cfg.Redis, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize Redis: %w", err)
	}
	defer redisStore.Close()
	logger.Info("✅ Redis initialized successfully")

	// Connect to NATS
	logger.Info("📡 [3/6] Connecting to NATS",
		zap.String("url", cfg.NATS.URL))
	natsConn, err := nats.Connect(cfg.NATS.URL,
		nats.Name("orchestrator-service"),
		nats.MaxReconnects(cfg.NATS.MaxReconnect),
		nats.ReconnectWait(cfg.NATS.ReconnectWait))
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}
	defer natsConn.Close()
	logger.Info("✅ NATS connected successfully",
		zap.String("server_info", natsConn.ConnectedUrl()))

	// Initialize workflow components
	logger.Info("⚙️  [4/6] Initializing workflow engine",
		zap.String("agent_manager_url", cfg.AI.AgentManagerURL),
		zap.String("reasoning_service_url", cfg.AI.ReasoningServiceURL))
	executor := workflow.NewExecutor(
		cfg.AI.AgentManagerURL,
		cfg.AI.ReasoningServiceURL,
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

func initLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	var encoderConfig zapcore.EncoderConfig
	if cfg.Format == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      cfg.Format != "json",
		Encoding:         cfg.Format,
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{cfg.OutputPath},
		ErrorOutputPaths: []string{"stderr"},
	}

	return zapConfig.Build()
}
