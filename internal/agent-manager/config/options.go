package config

import (
	configoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/spf13/pflag"
)

// Options defines options for agent-manager service
type Options struct {
	Server   *configoptions.ServerOptions   `json:"server" mapstructure:"server"`
	GRPC     *configoptions.GRPCOptions     `json:"grpc" mapstructure:"grpc"`
	Database *configoptions.DatabaseOptions `json:"database" mapstructure:"database"`
	Redis    *configoptions.RedisOptions    `json:"redis" mapstructure:"redis"`
	NATS     *configoptions.NATSOptions     `json:"nats" mapstructure:"nats"`
	Logging  *configoptions.LoggingOptions  `json:"logging" mapstructure:"logging"`
	Metrics  *configoptions.MetricsOptions  `json:"metrics" mapstructure:"metrics"`
}

// NewOptions creates a new Options instance with default values
func NewOptions() *Options {
	return &Options{
		Server:   configoptions.NewServerOptions(),
		GRPC:     configoptions.NewGRPCOptions(),
		Database: configoptions.NewDatabaseOptions(),
		Redis:    configoptions.NewRedisOptions(),
		NATS:     configoptions.NewNATSOptions(),
		Logging:  configoptions.NewLoggingOptions(),
		Metrics:  configoptions.NewMetricsOptions(),
	}
}

// Validate validates all the required options
func (o *Options) Validate() []error {
	var errs []error

	if err := o.Server.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.GRPC.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Database.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Redis.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.NATS.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Logging.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Metrics.Validate(); err != nil {
		errs = append(errs, err)
	}

	return errs
}

// Complete fills in any fields not set that are required to have valid data
func (o *Options) Complete() error {
	// Complete all sub-options
	if err := o.Server.Complete(); err != nil {
		return err
	}

	if err := o.GRPC.Complete(); err != nil {
		return err
	}

	if err := o.Database.Complete(); err != nil {
		return err
	}

	if err := o.Redis.Complete(); err != nil {
		return err
	}

	if err := o.NATS.Complete(); err != nil {
		return err
	}

	if err := o.Logging.Complete(); err != nil {
		return err
	}

	if err := o.Metrics.Complete(); err != nil {
		return err
	}

	return nil
}

// AddFlags adds flags to the flag set
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.Server.AddFlags(fs)
	o.GRPC.AddFlags(fs)
	o.Database.AddFlags(fs)
	o.Redis.AddFlags(fs)
	o.NATS.AddFlags(fs)
	o.Logging.AddFlags(fs)
	o.Metrics.AddFlags(fs)
}
