package config

import (
	"time"

	"github.com/spf13/pflag"

	commonoptions "github.com/kart-io/k8s-agent/common/options"
)

// Options defines options for cluster service
// This implements the pkg/app.Options interface.
type Options struct {
	Server   *commonoptions.ServerOptions   `json:"server" mapstructure:"server"`
	Database *commonoptions.DatabaseOptions `json:"database" mapstructure:"database"`
	JWT      *commonoptions.JWTOptions      `json:"jwt" mapstructure:"jwt"`
	Logging  *commonoptions.LoggingOptions  `json:"logging" mapstructure:"logging"`
}

// NewOptions creates a new Options instance with default values.
func NewOptions() *Options {
	return &Options{
		Server:   commonoptions.NewServerOptions(),
		Database: commonoptions.NewDatabaseOptions(),
		JWT:      commonoptions.NewJWTOptions(),
		Logging:  commonoptions.NewLoggingOptions(),
	}
}

// Validate validates all the required options.
func (o *Options) Validate() []error {
	var errs []error

	if err := o.Server.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Database.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.JWT.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Logging.Validate(); err != nil {
		errs = append(errs, err)
	}

	return errs
}

// Complete fills in any fields not set that are required to have valid data.
func (o *Options) Complete() error {
	if err := o.Server.Complete(); err != nil {
		return err
	}

	if err := o.Database.Complete(); err != nil {
		return err
	}

	if err := o.JWT.Complete(); err != nil {
		return err
	}

	if err := o.Logging.Complete(); err != nil {
		return err
	}

	return nil
}

// AddFlags adds flags to the flag set
// Note: --config/-c flag is automatically added by pkg/app framework.
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.Server.AddFlags(fs)
	o.Database.AddFlags(fs)
	o.JWT.AddFlags(fs)
	o.Logging.AddFlags(fs)
}

// ToLegacyConfig converts Options to legacy Config for backward compatibility
// Deprecated: Use Options directly in new code.
func (o *Options) ToLegacyConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         o.Server.Port,
			Mode:         o.Server.Mode,
			ReadTimeout:  o.Server.ReadTimeout.String(),
			WriteTimeout: o.Server.WriteTimeout.String(),
		},
		Database: DatabaseConfig{
			Host:         o.Database.Host,
			Port:         o.Database.Port,
			User:         o.Database.User,
			Password:     o.Database.Password,
			DBName:       o.Database.Database,
			SSLMode:      o.Database.SSLMode,
			MaxOpenConns: o.Database.MaxOpenConns,
			MaxIdleConns: o.Database.MaxIdleConns,
		},
		JWT: JWTConfig{
			Secret: o.JWT.Secret,
		},
		Logging: LoggingConfig{
			Level:  o.Logging.Level,
			Format: o.Logging.Format,
		},
	}
}

// FromLegacyConfig creates Options from legacy Config
// Deprecated: Use NewOptions() instead.
func FromLegacyConfig(cfg *Config) *Options {
	opts := NewOptions()

	// Server
	opts.Server.Port = cfg.Server.Port
	opts.Server.Mode = cfg.Server.Mode
	if rt, err := time.ParseDuration(cfg.Server.ReadTimeout); err == nil {
		opts.Server.ReadTimeout = rt
	}
	if wt, err := time.ParseDuration(cfg.Server.WriteTimeout); err == nil {
		opts.Server.WriteTimeout = wt
	}

	// Database
	opts.Database.Host = cfg.Database.Host
	opts.Database.Port = cfg.Database.Port
	opts.Database.User = cfg.Database.User
	opts.Database.Password = cfg.Database.Password
	opts.Database.Database = cfg.Database.DBName
	opts.Database.SSLMode = cfg.Database.SSLMode
	opts.Database.MaxOpenConns = cfg.Database.MaxOpenConns
	opts.Database.MaxIdleConns = cfg.Database.MaxIdleConns

	// JWT
	opts.JWT.Secret = cfg.JWT.Secret

	// Logging
	opts.Logging.Level = cfg.Logging.Level
	opts.Logging.Format = cfg.Logging.Format

	return opts
}
