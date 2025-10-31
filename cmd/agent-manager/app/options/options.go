// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package options contains flags and options for initializing the agent-manager service
package options

import (
	"github.com/spf13/pflag"

	commonlogger "github.com/kart-io/k8s-agent/common/logger"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	agentmanager "github.com/kart-io/k8s-agent/internal/agent-manager"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/logger/core"
)

const (
	// UserAgent is the userAgent name when starting agent-manager server.
	UserAgent = "aetherius-agent-manager"
)

// ServerOptions contains the configuration options for the agent-manager server.
type ServerOptions struct {
	// Server options for configuring HTTP server related options.
	Server *commonoptions.ServerOptions `json:"server" mapstructure:"server"`
	// GRPC options for configuring gRPC server related options.
	GRPC *commonoptions.GRPCOptions `json:"grpc" mapstructure:"grpc"`
	// Database options for configuring MySQL database related options.
	Database *commonoptions.DatabaseOptions `json:"database" mapstructure:"database"`
	// Redis options for configuring Redis related options.
	Redis *commonoptions.RedisOptions `json:"redis" mapstructure:"redis"`
	// NATS options for configuring NATS message queue related options.
	NATS *commonoptions.NATSOptions `json:"nats" mapstructure:"nats"`
	// Logging options for configuring logging related options.
	Logging *commonoptions.LoggingOptions `json:"logging" mapstructure:"logging"`
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
	healthOpts.Port = 8091 // Agent Manager 健康检查端口

	return &ServerOptions{
		Server:   commonoptions.NewServerOptions(),
		GRPC:     commonoptions.NewGRPCOptions(),
		Database: commonoptions.NewDatabaseOptions(),
		Redis:    commonoptions.NewRedisOptions(),
		NATS:     commonoptions.NewNATSOptions(),
		Logging:  commonoptions.NewLoggingOptions(),
		Metrics:  commonoptions.NewMetricsOptions(),
		Health:   healthOpts,
	}
}

// GetHealthPort 实现 commonapp.HealthPortProvider 接口
func (o *ServerOptions) GetHealthPort() int {
	if o.Health != nil {
		return o.Health.Port
	}
	return 8091 // 默认端口
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

// Config builds an agentmanager.Config based on ServerOptions.
// This method converts startup-layer configuration to business-layer configuration.
func (o *ServerOptions) Config() (*agentmanager.Config, error) {
	return &agentmanager.Config{
		Server:   o.Server,
		GRPC:     o.GRPC,
		Database: o.Database,
		Redis:    o.Redis,
		NATS:     o.NATS,
		Logging:  o.Logging,
		Metrics:  o.Metrics,
	}, nil
}

// GetServiceName returns the service name
// This method is required by the BootstrapConfig interface
func (o *ServerOptions) GetServiceName() string {
	return "Agent Manager"
}

// GetLogFields returns log fields for initialization logging
// This method is required by the BootstrapConfig interface
func (o *ServerOptions) GetLogFields() []interface{} {
	return []interface{}{
		"http_port", o.Server.Port,
		"grpc_enabled", o.GRPC.Enable,
		"grpc_port", o.GRPC.Port,
		"health_port", o.Health.Port,
	}
}

// InitLogger initializes logger based on the options.
func (o *ServerOptions) InitLogger() (core.Logger, error) {
	// Use the common logger initialization
	return commonlogger.InitFromOptions(o.Logging)
}
