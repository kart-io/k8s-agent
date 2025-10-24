// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package options contains flags and options for initializing the auth service
package options

import (
	"github.com/spf13/pflag"

	commonlogger "github.com/kart-io/k8s-agent/common/logger"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/internal/auth"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/logger/core"
)

const (
	// UserAgent is the userAgent name when starting auth server.
	UserAgent = "aetherius-auth"
)

// ServerOptions contains the configuration options for the auth server.
type ServerOptions struct {
	// Server options for configuring HTTP server related options.
	Server *commonoptions.ServerOptions `json:"server" mapstructure:"server"`
	// Database options for configuring MySQL database related options.
	Database *commonoptions.DatabaseOptions `json:"database" mapstructure:"database"`
	// Redis options for configuring Redis related options.
	Redis *commonoptions.RedisOptions `json:"redis" mapstructure:"redis"`
	// JWT options for configuring JWT authentication related options.
	JWT *commonoptions.JWTOptions `json:"jwt" mapstructure:"jwt"`
	// Logging options for configuring logging related options.
	Logging *commonoptions.LoggingOptions `json:"logging" mapstructure:"logging"`
	// Email options for configuring email related options.
	Email *commonoptions.EmailOptions `json:"email" mapstructure:"email"`
	// Metrics options for configuring metrics related options.
	Metrics *commonoptions.MetricsOptions `json:"metrics" mapstructure:"metrics"`
	// Health options for configuring health check related options.
	Health *commonoptions.HealthOptions `json:"health" mapstructure:"health"`
}

// Ensure ServerOptions implements the commonapp.Options interface.
var _ commonapp.Options = (*ServerOptions)(nil)

// NewServerOptions creates a ServerOptions instance with default values.
func NewServerOptions() *ServerOptions {
	healthOpts := commonoptions.NewHealthOptions()
	healthOpts.Port = 8090 // Auth 健康检查端口

	o := &ServerOptions{
		Server:   commonoptions.NewServerOptions(),
		Database: commonoptions.NewDatabaseOptions(),
		Redis:    commonoptions.NewRedisOptions(),
		JWT:      commonoptions.NewJWTOptions(),
		Logging:  commonoptions.NewLoggingOptions(),
		Email:    commonoptions.NewEmailOptions(),
		Metrics:  commonoptions.NewMetricsOptions(),
		Health:   healthOpts,
	}

	return o
}

// GetHealthPort 实现 commonapp.HealthPortProvider 接口
func (o *ServerOptions) GetHealthPort() int {
	if o.Health != nil {
		return o.Health.Port
	}
	return 8090 // 默认端口
}

// AddFlags adds flags to the specified FlagSet.
// This method implements the commonapp.NamedFlagSetOptions interface.
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	// Add all sub-options flags to the main flag set
	o.Server.AddFlags(fs)
	o.Database.AddFlags(fs)
	o.Redis.AddFlags(fs)
	o.JWT.AddFlags(fs)
	o.Logging.AddFlags(fs)
	o.Email.AddFlags(fs)
	o.Metrics.AddFlags(fs)
	if o.Health != nil {
		o.Health.AddFlags(fs, "")
	}
}

// Complete completes all the required options.
func (o *ServerOptions) Complete() error {
	// Set default service name in initial fields if not specified
	if o.Logging.InitialFields == nil {
		o.Logging.InitialFields = make(map[string]interface{})
	}
	if _, ok := o.Logging.InitialFields["service.name"]; !ok {
		o.Logging.InitialFields["service.name"] = UserAgent
	}

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

	if err := o.Metrics.Complete(); err != nil {
		return err
	}

	if o.Health != nil {
		if err := o.Health.Complete(); err != nil {
			return err
		}
	}

	return nil
}

// Validate checks whether the options in ServerOptions are valid.
func (o *ServerOptions) Validate() []error {
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

	if err := o.Metrics.Validate(); err != nil {
		errs = append(errs, err)
	}

	if o.Health != nil {
		if err := o.Health.Validate(); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// Config builds an auth.Config based on ServerOptions.
func (o *ServerOptions) Config() (*auth.Config, error) {
	return &auth.Config{
		Server:   o.Server,
		Database: o.Database,
		Redis:    o.Redis,
		JWT:      o.JWT,
		Logging:  o.Logging,
		Email:    o.Email,
		Metrics:  o.Metrics,
	}, nil
}

// InitLogger initializes logger based on the options.
func (o *ServerOptions) InitLogger() (core.Logger, error) {
	// Use the common logger initialization
	return commonlogger.InitFromOptions(o.Logging)
}
