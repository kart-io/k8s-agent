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

	"github.com/kart/k8s-agent/collect-agent/internal/agent"
	"github.com/kart/k8s-agent/collect-agent/internal/config"
)

var (
	configPath = flag.String("config", "/etc/aetherius/config.yaml", "path to configuration file")
	version    = flag.Bool("version", false, "print version information")
	healthPort = flag.Int("health-port", 8080, "port for health checks")
)

const (
	AppName    = "aetherius-collect-agent"
	AppVersion = "v1.0.0"
)

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("%s version %s\n", AppName, AppVersion)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := initLogger(cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Aetherius Collect Agent",
		zap.String("version", AppVersion),
		zap.String("config_path", *configPath),
		zap.String("cluster_id", cfg.ClusterID),
		zap.String("central_endpoint", cfg.CentralEndpoint))

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
		cancel()
	}()

	// Create and start agent
	agentInstance, err := agent.New(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create agent", zap.Error(err))
	}

	// Start health server
	healthServer := agent.NewHealthServer(agentInstance, *healthPort, logger)
	if err := healthServer.Start(); err != nil {
		logger.Fatal("Failed to start health server", zap.Error(err))
	}
	defer healthServer.Stop()

	// Start agent
	logger.Info("Starting agent services...")
	if err := agentInstance.Start(ctx); err != nil {
		logger.Fatal("Failed to start agent", zap.Error(err))
	}

	logger.Info("Agent shutdown complete")
}

// initLogger initializes the logger based on the log level
func initLogger(logLevel string) (*zap.Logger, error) {
	var level zapcore.Level
	switch logLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: level == zapcore.DebugLevel,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return config.Build()
}