package logger

import (
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger"
	"github.com/kart-io/logger/core"
	"github.com/kart-io/logger/option"
)

var globalLogger core.Logger

// Init initializes the kart-io/logger with configuration
func Init(cfg *commonoptions.LoggingOptions) error {
	// Create logger configuration
	logOption := &option.LogOption{
		Engine: cfg.Engine, // "zap" or "slog"
		Level:  cfg.Level,  // "debug", "info", "warn", "error"
		Format: cfg.Format, // "json" or "console"
	}

	// Set output paths
	if len(cfg.OutputPaths) > 0 {
		logOption.OutputPaths = cfg.OutputPaths
	} else {
		logOption.OutputPaths = []string{"stdout"}
	}

	// Configure OTLP if enabled
	if cfg.OTLP != nil && cfg.OTLP.Enabled != nil && *cfg.OTLP.Enabled {
		logOption.OTLP = &option.OTLPOption{
			Endpoint: cfg.OTLP.Endpoint,
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

	otlpEnabled := false
	if cfg.OTLP != nil && cfg.OTLP.Enabled != nil {
		otlpEnabled = *cfg.OTLP.Enabled
	}

	logger.Infow("Logger initialized",
		"engine", cfg.Engine,
		"level", cfg.Level,
		"format", cfg.Format,
		"otlp_enabled", otlpEnabled,
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
