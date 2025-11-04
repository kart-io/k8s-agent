package config

import (
	"github.com/spf13/pflag"

	commonoptions "github.com/kart-io/k8s-agent/common/options"
)

// Options defines options for auth service
// This replaces the old Config struct to align with the pkg/app framework
type Options struct {
	Server   *commonoptions.ServerOptions   `json:"server" mapstructure:"server"`
	Database *commonoptions.DatabaseOptions `json:"database" mapstructure:"database"`
	Redis    *commonoptions.RedisOptions    `json:"redis" mapstructure:"redis"`
	JWT      *commonoptions.JWTOptions      `json:"jwt" mapstructure:"jwt"`
	Logging  *commonoptions.LoggingOptions  `json:"logging" mapstructure:"logging"`
	Email    *commonoptions.EmailOptions    `json:"email" mapstructure:"email"`
}

// NewOptions creates a new Options instance with default values
func NewOptions() *Options {
	return &Options{
		Server:   commonoptions.NewServerOptions(),
		Database: commonoptions.NewDatabaseOptions(),
		Redis:    commonoptions.NewRedisOptions(),
		JWT:      commonoptions.NewJWTOptions(),
		Logging:  commonoptions.NewLoggingOptions(),
		Email:    commonoptions.NewEmailOptions(),
	}
}

// Validate validates all the required options
func (o *Options) Validate() []error {
	var errs []error

	if err := o.Server.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Database.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Redis.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.JWT.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Logging.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Email.Validate(); err != nil {
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

	if err := o.Database.Complete(); err != nil {
		return err
	}

	if err := o.Redis.Complete(); err != nil {
		return err
	}

	if err := o.JWT.Complete(); err != nil {
		return err
	}

	if err := o.Logging.Complete(); err != nil {
		return err
	}

	if err := o.Email.Complete(); err != nil {
		return err
	}

	return nil
}

// AddFlags adds flags to the flag set
// Note: --config/-c flag is automatically added by pkg/app framework
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.Server.AddFlags(fs)
	o.Database.AddFlags(fs)
	o.Redis.AddFlags(fs)
	o.JWT.AddFlags(fs)
	o.Logging.AddFlags(fs)
	o.Email.AddFlags(fs)
}
