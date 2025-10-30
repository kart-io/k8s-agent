// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package options

import (
	commonlogger "github.com/kart-io/k8s-agent/common/logger"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	clusterconfig "github.com/kart-io/k8s-agent/internal/cluster/config"
	"github.com/kart-io/logger/core"
	"github.com/spf13/pflag"
)

// ServerOptions defines options for cluster service
// This implements the pkg/app.Options interface
type ServerOptions struct {
	Server   *commonoptions.ServerOptions   `json:"server" mapstructure:"server"`
	Database *commonoptions.DatabaseOptions `json:"database" mapstructure:"database"`
	JWT      *commonoptions.JWTOptions      `json:"jwt" mapstructure:"jwt"`
	Logging  *commonoptions.LoggingOptions  `json:"logging" mapstructure:"logging"`
	Health   *commonoptions.HealthOptions   `json:"health" mapstructure:"health"`
}

// NewServerOptions creates a new ServerOptions instance with default values
func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		Server:   commonoptions.NewServerOptions(),
		Database: commonoptions.NewDatabaseOptions(),
		JWT:      commonoptions.NewJWTOptions(),
		Logging:  commonoptions.NewLoggingOptions(),
		Health:   commonoptions.NewHealthOptions(),
	}
}

// Validate validates all the required options
func (o *ServerOptions) Validate() []error {
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

	if err := o.Health.Validate(); err != nil {
		errs = append(errs, err)
	}

	return errs
}

// Complete fills in any fields not set that are required to have valid data
func (o *ServerOptions) Complete() error {
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

	if err := o.Health.Complete(); err != nil {
		return err
	}

	return nil
}

// AddFlags adds flags to the flag set
// Note: --config/-c flag is automatically added by pkg/app framework
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	o.Server.AddFlags(fs)
	o.Database.AddFlags(fs)
	o.JWT.AddFlags(fs)
	o.Logging.AddFlags(fs)
	o.Health.AddFlags(fs, "")
}

// InitLogger initializes the logger based on logging options
// This method is required by the Bootstrap pattern
func (o *ServerOptions) InitLogger() (core.Logger, error) {
	return commonlogger.InitFromOptions(o.Logging)
}

// GetHealthPort returns the health check port
// This method is required by the Bootstrap pattern
func (o *ServerOptions) GetHealthPort() int {
	return o.Health.Port
}

// Config converts ServerOptions to internal cluster config
// This method is required by the Bootstrap pattern
func (o *ServerOptions) Config() (*clusterconfig.Config, error) {
	// Convert to legacy config format for backward compatibility
	return &clusterconfig.Config{
		Server: clusterconfig.ServerConfig{
			Port:         o.Server.Port,
			Mode:         o.Server.Mode,
			ReadTimeout:  o.Server.ReadTimeout.String(),
			WriteTimeout: o.Server.WriteTimeout.String(),
		},
		Database: clusterconfig.DatabaseConfig{
			Host:         o.Database.Host,
			Port:         o.Database.Port,
			User:         o.Database.User,
			Password:     o.Database.Password,
			DBName:       o.Database.Database,
			SSLMode:      o.Database.SSLMode,
			MaxOpenConns: o.Database.MaxOpenConns,
			MaxIdleConns: o.Database.MaxIdleConns,
		},
		JWT: clusterconfig.JWTConfig{
			Secret: o.JWT.Secret,
		},
		Logging: clusterconfig.LoggingConfig{
			Level:  o.Logging.Level,
			Format: o.Logging.Format,
		},
	}, nil
}
