package logger

import (
	"github.com/kart-io/k8s-agent/internal/auth/config"
	"github.com/kart-io/logger"
	"github.com/kart-io/logger/core"
	"github.com/kart-io/logger/option"
)

var globalLogger core.Logger

// Init initializes the kart-io/logger with configuration
func Init(cfg *config.LoggingConfig) error {
	// Create logger configuration
	logOption := &option.LogOption{
		Engine: cfg.Engine, // "zap" or "slog"
		Level:  cfg.Level,  // "debug", "info", "warn", "error"
		Format: cfg.Format, // "json" or "console"
	}

	// Set output paths
	switch cfg.Output {
	case "", "stdout":
		logOption.OutputPaths = []string{"stdout"}
	case "stderr":
		logOption.OutputPaths = []string{"stderr"}
	default:
		// File path
		logOption.OutputPaths = []string{cfg.Output}
	}

	// Configure OTLP if enabled
	if cfg.OTLP.Enabled {
		logOption.OTLP = &option.OTLPOption{
			Endpoint: cfg.OTLP.Endpoint,
			Insecure: cfg.OTLP.Insecure,
		}
	}

	// Create logger
	log, err := logger.New(logOption)
	if err != nil {
		return err
	}

	// Set as global logger
	logger.SetGlobal(log)
	globalLogger = log

	logger.Infow("Logger initialized",
		"engine", cfg.Engine,
		"level", cfg.Level,
		"format", cfg.Format,
		"otlp_enabled", cfg.OTLP.Enabled,
	)

	return nil
}

// GetLogger returns the global logger instance
// This function maintains compatibility with existing code
func GetLogger() core.Logger {
	if globalLogger == nil {
		// Initialize with default settings if not initialized
		defaultOption := option.DefaultLogOption()
		log, _ := logger.New(defaultOption)
		logger.SetGlobal(log)
		globalLogger = log
	}
	return globalLogger
}
