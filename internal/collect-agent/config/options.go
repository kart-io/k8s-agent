package config

import (
	"time"

	"github.com/kart-io/k8s-agent/common/options"
	"github.com/spf13/pflag"
)

// Options defines options for collect-agent service
// This implements the pkg/app.Options interface
type Options struct {
	// Common options from common/options package
	Logging *options.LoggingOptions `json:"logging" mapstructure:"logging"`
	Agent   *options.AgentOptions   `json:"agent" mapstructure:"agent"`

	// Collect-agent specific options (if any)
	// Add service-specific options here if needed
}

// NewOptions creates a new Options instance with default values
func NewOptions() *Options {
	return &Options{
		Logging: options.NewLoggingOptions(),
		Agent:   options.NewAgentOptions(),
	}
}

// Validate validates all the required options
func (o *Options) Validate() []error {
	var errs []error

	// Validate logging options
	if err := o.Logging.Validate(); err != nil {
		errs = append(errs, err)
	}

	// Validate agent options
	if err := o.Agent.Validate(); err != nil {
		errs = append(errs, err)
	}

	return errs
}

// Complete fills in any fields not set that are required to have valid data
func (o *Options) Complete() error {
	if err := o.Logging.Complete(); err != nil {
		return err
	}

	if err := o.Agent.Complete(); err != nil {
		return err
	}

	return nil
}

// AddFlags adds flags to the flag set
// Note: --config/-c flag is automatically added by pkg/app framework
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.Logging.AddFlags(fs)
	o.Agent.AddFlags(fs)
}

// Deprecated compatibility methods - map to new Agent options structure

// GetClusterID returns the cluster ID (for backward compatibility)
func (o *Options) GetClusterID() string {
	return o.Agent.ClusterID
}

// GetClusterName returns the cluster name (for backward compatibility)
func (o *Options) GetClusterName() string {
	return o.Agent.ClusterName
}

// GetCentralEndpoint returns the central endpoint (for backward compatibility)
func (o *Options) GetCentralEndpoint() string {
	return o.Agent.CentralEndpoint
}

// GetReconnectDelay returns the reconnect delay (for backward compatibility)
func (o *Options) GetReconnectDelay() time.Duration {
	return o.Agent.ReconnectDelay
}

// GetHeartbeatInterval returns the heartbeat interval (for backward compatibility)
func (o *Options) GetHeartbeatInterval() time.Duration {
	return o.Agent.HeartbeatInterval
}

// GetMetricsInterval returns the metrics interval (for backward compatibility)
func (o *Options) GetMetricsInterval() time.Duration {
	return o.Agent.MetricsInterval
}

// GetBufferSize returns the buffer size (for backward compatibility)
func (o *Options) GetBufferSize() int {
	return o.Agent.BufferSize
}

// GetMaxRetries returns the max retries (for backward compatibility)
func (o *Options) GetMaxRetries() int {
	return o.Agent.MaxRetries
}

// IsMetricsEnabled returns whether metrics are enabled (for backward compatibility)
func (o *Options) IsMetricsEnabled() bool {
	return o.Agent.EnableMetrics
}

// IsEventsEnabled returns whether events are enabled (for backward compatibility)
func (o *Options) IsEventsEnabled() bool {
	return o.Agent.EnableEvents
}

// GetHealthPort returns the health port (for backward compatibility)
func (o *Options) GetHealthPort() int {
	return o.Agent.HealthPort
}
