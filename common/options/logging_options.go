package config

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

// LoggingOptions defines logging configuration (aligned with kart-io/logger)
type LoggingOptions struct {
	// Engine specifies which logging engine to use ("zap" or "slog")
	Engine string `json:"engine" mapstructure:"engine"`

	// Level sets the minimum logging level
	Level string `json:"level" mapstructure:"level"`

	// Format specifies output format ("json" or "console")
	Format string `json:"format" mapstructure:"format"`

	// OutputPaths specifies where logs should be written
	OutputPaths []string `json:"output_paths" mapstructure:"output_paths"`

	// InitialFields are fields added to every log entry (like service.name, service.version)
	InitialFields map[string]interface{} `json:"initial_fields" mapstructure:"initial_fields"`

	// OTLP configuration (flattened and nested)
	OTLPEndpoint string      `json:"otlp_endpoint" mapstructure:"otlp_endpoint" yaml:"otlp_endpoint"`
	OTLP         *OTLPOption `json:"otlp" mapstructure:"otlp" yaml:"otlp"`

	// Development mode enables caller info and stacktraces
	Development bool `json:"development" mapstructure:"development"`

	// DisableCaller disables automatic caller detection
	DisableCaller bool `json:"disable_caller" mapstructure:"disable_caller"`

	// DisableStacktrace disables automatic stacktrace capture
	DisableStacktrace bool `json:"disable_stacktrace" mapstructure:"disable_stacktrace"`

	// Rotation configuration for file output
	Rotation *RotationOption `json:"rotation" mapstructure:"rotation" yaml:"rotation"`
}

// OTLPOption contains OTLP-specific configuration
type OTLPOption struct {
	Enabled  *bool             `json:"enabled" mapstructure:"enabled" yaml:"enabled"`
	Endpoint string            `json:"endpoint" mapstructure:"endpoint" yaml:"endpoint"`
	Protocol string            `json:"protocol" mapstructure:"protocol" yaml:"protocol"`
	Timeout  time.Duration     `json:"timeout" mapstructure:"timeout" yaml:"timeout"`
	Headers  map[string]string `json:"headers" mapstructure:"headers" yaml:"headers"`
	Insecure bool              `json:"insecure" mapstructure:"insecure" yaml:"insecure"`
}

// RotationOption contains log file rotation configuration
type RotationOption struct {
	MaxSize        int    `json:"max_size" mapstructure:"max_size" yaml:"max_size"`
	MaxAge         int    `json:"max_age" mapstructure:"max_age" yaml:"max_age"`
	MaxBackups     int    `json:"max_backups" mapstructure:"max_backups" yaml:"max_backups"`
	Compress       bool   `json:"compress" mapstructure:"compress" yaml:"compress"`
	RotateInterval string `json:"rotate_interval" mapstructure:"rotate_interval" yaml:"rotate_interval"`
}

// NewLoggingOptions creates a new LoggingOptions instance with default values
func NewLoggingOptions() *LoggingOptions {
	return &LoggingOptions{
		Engine:            "zap",
		Level:             "info",
		Format:            "json",
		OutputPaths:       []string{"stdout"},
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		InitialFields:     make(map[string]interface{}),
		OTLP: &OTLPOption{
			Protocol: "grpc",
			Timeout:  10 * time.Second,
			Insecure: true,
		},
		Rotation: &RotationOption{
			MaxSize:        100,
			MaxAge:         15,
			MaxBackups:     30,
			Compress:       true,
			RotateInterval: "7d",
		},
	}
}

// AddFlags adds flags to the flag set
func (o *LoggingOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Engine, "logging.engine", o.Engine, "Logging engine (zap|slog)")
	fs.StringVar(&o.Level, "logging.level", o.Level, "Log level (DEBUG|INFO|WARN|ERROR|FATAL)")
	fs.StringVar(&o.Format, "logging.format", o.Format, "Log format (json|console)")
	fs.StringSliceVar(&o.OutputPaths, "logging.output-paths", o.OutputPaths, "Output paths for logs")
	fs.StringVar(&o.OTLPEndpoint, "logging.otlp-endpoint", "", "OTLP endpoint URL")
	fs.BoolVar(&o.Development, "logging.development", o.Development, "Enable development mode")
	fs.BoolVar(&o.DisableCaller, "logging.disable-caller", o.DisableCaller, "Disable caller detection")
	fs.BoolVar(&o.DisableStacktrace, "logging.disable-stacktrace", o.DisableStacktrace, "Disable stacktrace capture")

	// OTLP nested options
	if o.OTLP == nil {
		o.OTLP = &OTLPOption{}
	}
	fs.StringVar(&o.OTLP.Endpoint, "logging.otlp.endpoint", "", "OTLP nested endpoint URL")
	fs.StringVar(&o.OTLP.Protocol, "logging.otlp.protocol", "grpc", "OTLP protocol (grpc|http)")
	fs.DurationVar(&o.OTLP.Timeout, "logging.otlp.timeout", 10*time.Second, "OTLP timeout duration")

	// Rotation options
	if o.Rotation == nil {
		o.Rotation = &RotationOption{}
	}
	fs.IntVar(&o.Rotation.MaxSize, "logging.rotation.max-size", 100, "Maximum size in MB of the log file before rotation")
	fs.IntVar(&o.Rotation.MaxAge, "logging.rotation.max-age", 15, "Maximum number of days to retain old log files")
	fs.IntVar(&o.Rotation.MaxBackups, "logging.rotation.max-backups", 30, "Maximum number of old log files to retain")
	fs.BoolVar(&o.Rotation.Compress, "logging.rotation.compress", true, "Compress rotated log files using gzip")
	fs.StringVar(&o.Rotation.RotateInterval, "logging.rotation.rotate-interval", "7d", "Log rotation interval (e.g., 1h, 24h, 7d)")
}

// Validate validates the logging options
func (o *LoggingOptions) Validate() error {
	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
		"DEBUG": true,
		"INFO":  true,
		"WARN":  true,
		"ERROR": true,
		"FATAL": true,
	}

	if !validLevels[o.Level] {
		return fmt.Errorf("invalid log level: %s, must be one of: debug, info, warn, error, fatal", o.Level)
	}

	// Validate log format
	if o.Format != "json" && o.Format != "console" && o.Format != "text" {
		return fmt.Errorf("invalid log format: %s, must be json or console", o.Format)
	}

	// Convert "text" to "console" for compatibility
	if o.Format == "text" {
		o.Format = "console"
	}

	// Validate logger engine
	if o.Engine != "zap" && o.Engine != "slog" {
		return fmt.Errorf("invalid logger engine: %s, must be zap or slog", o.Engine)
	}

	// Validate output paths
	if len(o.OutputPaths) == 0 {
		return fmt.Errorf("at least one output path is required")
	}

	// Validate OTLP configuration
	if err := o.validateOTLPConfig(); err != nil {
		return err
	}

	// Validate rotation configuration
	if err := o.validateRotationConfig(); err != nil {
		return err
	}

	return nil
}

// validateOTLPConfig validates OTLP configuration
func (o *LoggingOptions) validateOTLPConfig() error {
	if o.OTLP == nil {
		o.OTLP = &OTLPOption{}
	}

	// Apply flattened configuration logic
	if o.OTLPEndpoint != "" {
		if o.OTLP.Enabled != nil && !*o.OTLP.Enabled {
			return nil
		}
		if o.OTLP.Enabled == nil {
			enabled := true
			o.OTLP.Enabled = &enabled
		}
		o.OTLP.Endpoint = o.OTLPEndpoint
	} else if o.OTLP.Enabled == nil && o.OTLP.Endpoint != "" {
		enabled := true
		o.OTLP.Enabled = &enabled
	}

	// Apply defaults for enabled OTLP
	if o.OTLP.Enabled != nil && *o.OTLP.Enabled {
		if o.OTLP.Protocol == "" {
			o.OTLP.Protocol = "grpc"
		}
		if o.OTLP.Timeout == 0 {
			o.OTLP.Timeout = 10 * time.Second
		}
	}

	return nil
}

// validateRotationConfig validates rotation configuration
func (o *LoggingOptions) validateRotationConfig() error {
	if o.Rotation == nil {
		return nil
	}

	rotation := o.Rotation

	if rotation.MaxSize < 0 {
		return fmt.Errorf("rotation max_size must be non-negative, got %d", rotation.MaxSize)
	}

	if rotation.MaxAge < 0 {
		return fmt.Errorf("rotation max_age must be non-negative, got %d", rotation.MaxAge)
	}

	if rotation.MaxBackups < 0 {
		return fmt.Errorf("rotation max_backups must be non-negative, got %d", rotation.MaxBackups)
	}

	// Apply defaults
	if rotation.MaxSize <= 0 {
		rotation.MaxSize = 100
	}
	if rotation.MaxAge <= 0 {
		rotation.MaxAge = 15
	}
	if rotation.MaxBackups <= 0 {
		rotation.MaxBackups = 30
	}
	if rotation.RotateInterval == "" {
		rotation.RotateInterval = "7d"
	}

	return nil
}

// Complete fills in any fields not set that are required to have valid data
func (o *LoggingOptions) Complete() error {
	// Set default log level if not specified
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
		"DEBUG": true,
		"INFO":  true,
		"WARN":  true,
		"ERROR": true,
		"FATAL": true,
	}

	if !validLevels[o.Level] {
		o.Level = "info"
	}

	// Set default log format if not specified
	if o.Format != "json" && o.Format != "console" && o.Format != "text" {
		o.Format = "json"
	}

	// Convert "text" to "console" for compatibility
	if o.Format == "text" {
		o.Format = "console"
	}

	// Set default logger engine if not specified
	if o.Engine != "zap" && o.Engine != "slog" {
		o.Engine = "zap"
	}

	// Set default output if not specified
	if len(o.OutputPaths) == 0 {
		o.OutputPaths = []string{"stdout"}
	}

	// Initialize OTLP if nil
	if o.OTLP == nil {
		o.OTLP = &OTLPOption{
			Protocol: "grpc",
			Timeout:  10 * time.Second,
			Insecure: true,
		}
	}

	// Initialize Rotation if nil
	if o.Rotation == nil {
		o.Rotation = &RotationOption{
			MaxSize:        100,
			MaxAge:         15,
			MaxBackups:     30,
			Compress:       true,
			RotateInterval: "7d",
		}
	}

	// Initialize InitialFields if nil
	if o.InitialFields == nil {
		o.InitialFields = make(map[string]interface{})
	}

	return nil
}

// ApplyTo applies the logging configuration to the target interface
func (o *LoggingOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	// Type assertion to check if target is a configuration structure pointer
	switch v := target.(type) {
	case *[]interface{}:
		*v = append(*v, map[string]interface{}{
			"engine":             o.Engine,
			"level":              o.Level,
			"format":             o.Format,
			"output_paths":       o.OutputPaths,
			"initial_fields":     o.InitialFields,
			"otlp_endpoint":      o.OTLPEndpoint,
			"development":        o.Development,
			"disable_caller":     o.DisableCaller,
			"disable_stacktrace": o.DisableStacktrace,
		})
	}

	return nil
}

// WithInitialFields adds or updates fields in InitialFields map
func (o *LoggingOptions) WithInitialFields(fields map[string]interface{}) *LoggingOptions {
	if o.InitialFields == nil {
		o.InitialFields = make(map[string]interface{})
	}

	for key, value := range fields {
		o.InitialFields[key] = value
	}

	return o
}

// AddInitialField adds a single field to InitialFields
func (o *LoggingOptions) AddInitialField(key string, value interface{}) *LoggingOptions {
	if o.InitialFields == nil {
		o.InitialFields = make(map[string]interface{})
	}

	o.InitialFields[key] = value
	return o
}

// IsOTLPEnabled returns true if OTLP is enabled after configuration resolution
func (o *LoggingOptions) IsOTLPEnabled() bool {
	return o.OTLP != nil && o.OTLP.Enabled != nil && *o.OTLP.Enabled && o.OTLP.Endpoint != ""
}
