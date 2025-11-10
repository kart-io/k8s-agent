package options

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

// AgentOptions defines options for collect-agent service
// This is a reusable option set that can be used across different agent implementations
type AgentOptions struct {
	// Cluster identification
	ClusterID   string `json:"cluster_id" mapstructure:"cluster_id"`
	ClusterName string `json:"cluster_name" mapstructure:"cluster_name"`

	// Central endpoint configuration
	CentralEndpoint string `json:"central_endpoint" mapstructure:"central_endpoint"`

	// Connection management
	ReconnectDelay    time.Duration `json:"reconnect_delay" mapstructure:"reconnect_delay"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval" mapstructure:"heartbeat_interval"`
	ConnectionTimeout time.Duration `json:"connection_timeout" mapstructure:"connection_timeout"`

	// Data collection intervals
	MetricsInterval time.Duration `json:"metrics_interval" mapstructure:"metrics_interval"`
	EventInterval   time.Duration `json:"event_interval" mapstructure:"event_interval"`

	// Buffer and queue configuration
	BufferSize       int `json:"buffer_size" mapstructure:"buffer_size"`
	EventQueueSize   int `json:"event_queue_size" mapstructure:"event_queue_size"`
	MetricsQueueSize int `json:"metrics_queue_size" mapstructure:"metrics_queue_size"`

	// Retry configuration
	MaxRetries      int           `json:"max_retries" mapstructure:"max_retries"`
	RetryBackoff    time.Duration `json:"retry_backoff" mapstructure:"retry_backoff"`
	MaxRetryBackoff time.Duration `json:"max_retry_backoff" mapstructure:"max_retry_backoff"`

	// Feature toggles
	EnableMetrics bool `json:"enable_metrics" mapstructure:"enable_metrics"`
	EnableEvents  bool `json:"enable_events" mapstructure:"enable_events"`
	EnableTracing bool `json:"enable_tracing" mapstructure:"enable_tracing"`

	// Health and monitoring
	HealthPort  int  `json:"health_port" mapstructure:"health_port"`
	EnablePprof bool `json:"enable_pprof" mapstructure:"enable_pprof"`
	PprofPort   int  `json:"pprof_port" mapstructure:"pprof_port"`

	// Resource limits
	MaxConcurrentRequests int           `json:"max_concurrent_requests" mapstructure:"max_concurrent_requests"`
	RequestTimeout        time.Duration `json:"request_timeout" mapstructure:"request_timeout"`
}

// NewAgentOptions creates a new AgentOptions instance with default values
func NewAgentOptions() *AgentOptions {
	return &AgentOptions{
		// Default cluster identification (should be overridden)
		ClusterID:   "",
		ClusterName: "",

		// Default central endpoint
		CentralEndpoint: "nats://localhost:4222",

		// Connection defaults
		ReconnectDelay:    5 * time.Second,
		HeartbeatInterval: 30 * time.Second,
		ConnectionTimeout: 10 * time.Second,

		// Collection intervals
		MetricsInterval: 60 * time.Second,
		EventInterval:   5 * time.Second,

		// Buffer sizes
		BufferSize:       1000,
		EventQueueSize:   500,
		MetricsQueueSize: 500,

		// Retry configuration
		MaxRetries:      10,
		RetryBackoff:    1 * time.Second,
		MaxRetryBackoff: 60 * time.Second,

		// Feature flags (all enabled by default)
		EnableMetrics: true,
		EnableEvents:  true,
		EnableTracing: false,

		// Health monitoring
		HealthPort:  8080,
		EnablePprof: false,
		PprofPort:   6060,

		// Resource limits
		MaxConcurrentRequests: 100,
		RequestTimeout:        30 * time.Second,
	}
}

// Validate validates all the required options
func (o *AgentOptions) Validate() error {
	if o.CentralEndpoint == "" {
		return fmt.Errorf("central_endpoint is required")
	}

	if o.ReconnectDelay < time.Second {
		return fmt.Errorf("reconnect_delay must be at least 1 second, got %v", o.ReconnectDelay)
	}

	if o.HeartbeatInterval < 10*time.Second {
		return fmt.Errorf("heartbeat_interval must be at least 10 seconds, got %v", o.HeartbeatInterval)
	}

	if o.MetricsInterval < 30*time.Second {
		return fmt.Errorf("metrics_interval must be at least 30 seconds, got %v", o.MetricsInterval)
	}

	if o.BufferSize < 10 {
		return fmt.Errorf("buffer_size must be at least 10, got %d", o.BufferSize)
	}

	if o.EventQueueSize < 10 {
		return fmt.Errorf("event_queue_size must be at least 10, got %d", o.EventQueueSize)
	}

	if o.MetricsQueueSize < 10 {
		return fmt.Errorf("metrics_queue_size must be at least 10, got %d", o.MetricsQueueSize)
	}

	if o.MaxRetries < 1 {
		return fmt.Errorf("max_retries must be at least 1, got %d", o.MaxRetries)
	}

	if o.HealthPort <= 0 || o.HealthPort > 65535 {
		return fmt.Errorf("health_port must be between 1-65535, got %d", o.HealthPort)
	}

	if o.EnablePprof && (o.PprofPort <= 0 || o.PprofPort > 65535) {
		return fmt.Errorf("pprof_port must be between 1-65535 when pprof is enabled, got %d", o.PprofPort)
	}

	if o.MaxConcurrentRequests < 1 {
		return fmt.Errorf("max_concurrent_requests must be at least 1, got %d", o.MaxConcurrentRequests)
	}

	if o.RequestTimeout < time.Second {
		return fmt.Errorf("request_timeout must be at least 1 second, got %v", o.RequestTimeout)
	}

	if o.ConnectionTimeout < time.Second {
		return fmt.Errorf("connection_timeout must be at least 1 second, got %v", o.ConnectionTimeout)
	}

	return nil
}

// Complete fills in any fields not set that are required to have valid data
func (o *AgentOptions) Complete() error {
	// Set defaults if not configured
	if o.CentralEndpoint == "" {
		o.CentralEndpoint = "nats://localhost:4222"
	}

	if o.ReconnectDelay == 0 {
		o.ReconnectDelay = 5 * time.Second
	}

	if o.HeartbeatInterval == 0 {
		o.HeartbeatInterval = 30 * time.Second
	}

	if o.ConnectionTimeout == 0 {
		o.ConnectionTimeout = 10 * time.Second
	}

	if o.MetricsInterval == 0 {
		o.MetricsInterval = 60 * time.Second
	}

	if o.EventInterval == 0 {
		o.EventInterval = 5 * time.Second
	}

	if o.BufferSize == 0 {
		o.BufferSize = 1000
	}

	if o.EventQueueSize == 0 {
		o.EventQueueSize = 500
	}

	if o.MetricsQueueSize == 0 {
		o.MetricsQueueSize = 500
	}

	if o.MaxRetries == 0 {
		o.MaxRetries = 10
	}

	if o.RetryBackoff == 0 {
		o.RetryBackoff = 1 * time.Second
	}

	if o.MaxRetryBackoff == 0 {
		o.MaxRetryBackoff = 60 * time.Second
	}

	if o.HealthPort == 0 {
		o.HealthPort = 8080
	}

	if o.PprofPort == 0 {
		o.PprofPort = 6060
	}

	if o.MaxConcurrentRequests == 0 {
		o.MaxConcurrentRequests = 100
	}

	if o.RequestTimeout == 0 {
		o.RequestTimeout = 30 * time.Second
	}

	return nil
}

// AddFlags adds flags to the flag set
func (o *AgentOptions) AddFlags(fs *pflag.FlagSet) {
	// Cluster identification
	fs.StringVar(&o.ClusterID, "cluster-id", o.ClusterID,
		"Unique identifier for this cluster (required)")

	fs.StringVar(&o.ClusterName, "cluster-name", o.ClusterName,
		"Human-readable name for this cluster")

	// Central endpoint
	fs.StringVar(&o.CentralEndpoint, "central-endpoint", o.CentralEndpoint,
		"NATS endpoint for central agent manager (e.g., nats://localhost:4222)")

	// Connection management
	fs.DurationVar(&o.ReconnectDelay, "reconnect-delay", o.ReconnectDelay,
		"Delay between reconnection attempts (minimum 1s)")

	fs.DurationVar(&o.HeartbeatInterval, "heartbeat-interval", o.HeartbeatInterval,
		"Interval for sending heartbeat messages (minimum 10s)")

	fs.DurationVar(&o.ConnectionTimeout, "connection-timeout", o.ConnectionTimeout,
		"Timeout for establishing connections (minimum 1s)")

	// Data collection
	fs.DurationVar(&o.MetricsInterval, "metrics-interval", o.MetricsInterval,
		"Interval for collecting and sending metrics (minimum 30s)")

	fs.DurationVar(&o.EventInterval, "event-interval", o.EventInterval,
		"Interval for processing events")

	// Buffers and queues
	fs.IntVar(&o.BufferSize, "buffer-size", o.BufferSize,
		"Size of main buffer queue (minimum 10)")

	fs.IntVar(&o.EventQueueSize, "event-queue-size", o.EventQueueSize,
		"Size of event queue (minimum 10)")

	fs.IntVar(&o.MetricsQueueSize, "metrics-queue-size", o.MetricsQueueSize,
		"Size of metrics queue (minimum 10)")

	// Retry configuration
	fs.IntVar(&o.MaxRetries, "max-retries", o.MaxRetries,
		"Maximum number of retries for failed operations (minimum 1)")

	fs.DurationVar(&o.RetryBackoff, "retry-backoff", o.RetryBackoff,
		"Initial backoff duration for retries")

	fs.DurationVar(&o.MaxRetryBackoff, "max-retry-backoff", o.MaxRetryBackoff,
		"Maximum backoff duration for retries")

	// Feature toggles
	fs.BoolVar(&o.EnableMetrics, "enable-metrics", o.EnableMetrics,
		"Enable metrics collection and reporting")

	fs.BoolVar(&o.EnableEvents, "enable-events", o.EnableEvents,
		"Enable event monitoring and reporting")

	fs.BoolVar(&o.EnableTracing, "enable-tracing", o.EnableTracing,
		"Enable distributed tracing")

	// Health and monitoring
	fs.IntVar(&o.HealthPort, "health-port", o.HealthPort,
		"Port for health check endpoint")

	fs.BoolVar(&o.EnablePprof, "enable-pprof", o.EnablePprof,
		"Enable pprof profiling endpoints")

	fs.IntVar(&o.PprofPort, "pprof-port", o.PprofPort,
		"Port for pprof profiling endpoints")

	// Resource limits
	fs.IntVar(&o.MaxConcurrentRequests, "max-concurrent-requests", o.MaxConcurrentRequests,
		"Maximum number of concurrent requests")

	fs.DurationVar(&o.RequestTimeout, "request-timeout", o.RequestTimeout,
		"Timeout for individual requests")
}

// String returns a string representation of the options (for debugging)
func (o *AgentOptions) String() string {
	return fmt.Sprintf("AgentOptions{ClusterID=%s, ClusterName=%s, CentralEndpoint=%s, "+
		"HeartbeatInterval=%v, MetricsInterval=%v, BufferSize=%d, EnableMetrics=%v, EnableEvents=%v}",
		o.ClusterID, o.ClusterName, o.CentralEndpoint,
		o.HeartbeatInterval, o.MetricsInterval, o.BufferSize,
		o.EnableMetrics, o.EnableEvents)
}
