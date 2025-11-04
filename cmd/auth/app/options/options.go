// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package options contains flags and options for initializing the auth service
package options

import (
	"github.com/spf13/pflag"

	"github.com/kart-io/k8s-agent/common/loggerutil"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
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
}

// Ensure ServerOptions implements the commonapp.Options interface.
var _ commonapp.Options = (*ServerOptions)(nil)

// NewServerOptions creates a ServerOptions instance with default values.
func NewServerOptions() *ServerOptions {
	o := &ServerOptions{
		Server:   commonoptions.NewServerOptions(),
		Database: commonoptions.NewDatabaseOptions(),
		Redis:    commonoptions.NewRedisOptions(),
		JWT:      commonoptions.NewJWTOptions(),
		Logging:  commonoptions.NewLoggingOptions(),
		Email:    commonoptions.NewEmailOptions(),
		Metrics:  commonoptions.NewMetricsOptions(),
	}

	return o
}

// AddFlags adds flags to the specified FlagSet.
// This method implements the commonapp.NamedFlagSetOptions interface.
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	// 使用通用工具函数统一添加所有子选项的 flags
	commonoptions.AddFlagsAll(o, fs)
}

// Complete completes all the required options.
func (o *ServerOptions) Complete() error {
	// 使用通用工具函数：设置服务名称并完成所有子选项
	return commonoptions.CompleteWithServiceName(o, o.Logging, UserAgent)
}

// Validate checks whether the options in ServerOptions are valid.
func (o *ServerOptions) Validate() []error {
	// 使用通用工具函数统一验证所有子选项
	return commonoptions.ValidateAll(o)
}

// GetServiceName returns the service name.
func (o *ServerOptions) GetServiceName() string {
	return "Auth"
}

// GetLogFields returns log fields for initialization logging.
func (o *ServerOptions) GetLogFields() []interface{} {
	return []interface{}{
		"http_port", o.Server.Port,
		"email_enabled", o.Email.Enabled,
	}
}

// InitLogger initializes logger based on the options.
func (o *ServerOptions) InitLogger() (core.Logger, error) {
	// Use the common logger initialization
	return loggerutil.InitFromOptions(o.Logging)
}
