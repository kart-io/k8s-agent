// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package options contains flags and options for initializing the orchestrator service
package options

import (
	"time"

	"github.com/spf13/pflag"

	commonlogger "github.com/kart-io/k8s-agent/common/logger"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	orchestrator "github.com/kart-io/k8s-agent/internal/orchestrator"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/logger/core"
)

const (
	// UserAgent is the userAgent name when starting orchestrator server.
	UserAgent = "aetherius-orchestrator"
)

// ServerOptions contains the configuration options for the orchestrator server.
type ServerOptions struct {
	// Server options for configuring HTTP server related options.
	Server *commonoptions.ServerOptions `json:"server" mapstructure:"server"`
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
	// AI options for configuring AI service integration.
	AI *AIOptions `json:"ai" mapstructure:"ai"`
}

// AIOptions contains AI service configuration options.
type AIOptions struct {
	// ReasoningServiceURL is the URL of the reasoning service.
	ReasoningServiceURL string `json:"reasoning_service_url" mapstructure:"reasoning_service_url"`
	// AgentManagerURL is the URL of the agent manager service.
	AgentManagerURL string `json:"agent_manager_url" mapstructure:"agent_manager_url"`
	// Timeout is the request timeout for AI service calls.
	Timeout time.Duration `json:"timeout" mapstructure:"timeout"`
	// MaxRetries is the maximum number of retry attempts.
	MaxRetries int `json:"max_retries" mapstructure:"max_retries"`
}

// NewAIOptions creates default AIOptions.
func NewAIOptions() *AIOptions {
	return &AIOptions{
		ReasoningServiceURL: "http://localhost:8082",
		AgentManagerURL:     "http://localhost:8080",
		Timeout:             30 * time.Second,
		MaxRetries:          3,
	}
}

// AddFlags adds flags for AI options.
func (o *AIOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ReasoningServiceURL, "ai.reasoning-service-url", o.ReasoningServiceURL,
		"URL of the reasoning service")
	fs.StringVar(&o.AgentManagerURL, "ai.agent-manager-url", o.AgentManagerURL,
		"URL of the agent manager service")
	fs.DurationVar(&o.Timeout, "ai.timeout", o.Timeout,
		"Request timeout for AI service calls")
	fs.IntVar(&o.MaxRetries, "ai.max-retries", o.MaxRetries,
		"Maximum number of retry attempts for AI service calls")
}

// Validate validates AI options.
func (o *AIOptions) Validate() error {
	return nil // Add validation logic if needed
}

// Complete fills in any fields not set that are required to have valid data.
func (o *AIOptions) Complete() error {
	if o.ReasoningServiceURL == "" {
		o.ReasoningServiceURL = "http://localhost:8082"
	}
	if o.AgentManagerURL == "" {
		o.AgentManagerURL = "http://localhost:8080"
	}
	if o.Timeout == 0 {
		o.Timeout = 30 * time.Second
	}
	if o.MaxRetries == 0 {
		o.MaxRetries = 3
	}
	return nil
}

// Ensure ServerOptions implements the commonapp.Options interface.
var _ commonapp.Options = (*ServerOptions)(nil)

// NewServerOptions creates a ServerOptions instance with default values.
func NewServerOptions() *ServerOptions {
	healthOpts := commonoptions.NewHealthOptions()
	healthOpts.Port = 8092 // Orchestrator 健康检查端口

	return &ServerOptions{
		Server:   commonoptions.NewServerOptions(),
		Database: commonoptions.NewDatabaseOptions(),
		Redis:    commonoptions.NewRedisOptions(),
		NATS:     commonoptions.NewNATSOptions(),
		Logging:  commonoptions.NewLoggingOptions(),
		Metrics:  commonoptions.NewMetricsOptions(),
		Health:   healthOpts,
		AI:       NewAIOptions(),
	}
}

// GetHealthPort 实现 commonapp.HealthPortProvider 接口
func (o *ServerOptions) GetHealthPort() int {
	if o.Health != nil {
		return o.Health.Port
	}
	return 8092 // 默认端口
}

// AddFlags adds flags to the specified FlagSet.
// This method implements the commonapp.NamedFlagSetOptions interface.
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	// Add all sub-options flags to the main flag set
	o.Server.AddFlags(fs)
	o.Database.AddFlags(fs)
	o.Redis.AddFlags(fs)
	o.NATS.AddFlags(fs)
	o.Logging.AddFlags(fs)
	o.Metrics.AddFlags(fs)
	if o.Health != nil {
		o.Health.AddFlags(fs, "")
	}
	if o.AI != nil {
		o.AI.AddFlags(fs)
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

	if err := o.NATS.Complete(); err != nil {
		return err
	}

	if err := o.Logging.Complete(); err != nil {
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

	if o.AI != nil {
		if err := o.AI.Complete(); err != nil {
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

	if err := o.NATS.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Logging.Validate(); err != nil {
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

	if o.AI != nil {
		if err := o.AI.Validate(); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// Config builds an orchestrator.Config based on ServerOptions.
// This method converts startup-layer configuration to business-layer configuration.
func (o *ServerOptions) Config() (*orchestrator.Config, error) {
	return &orchestrator.Config{
		Server:   o.Server,
		Database: o.Database,
		Redis:    o.Redis,
		NATS:     o.NATS,
		Logging:  o.Logging,
		Metrics:  o.Metrics,
		AI: &orchestrator.AIConfig{
			ReasoningServiceURL: o.AI.ReasoningServiceURL,
			AgentManagerURL:     o.AI.AgentManagerURL,
			Timeout:             o.AI.Timeout,
			MaxRetries:          o.AI.MaxRetries,
		},
	}, nil
}

// InitLogger initializes logger based on the options.
func (o *ServerOptions) InitLogger() (core.Logger, error) {
	// Use the common logger initialization
	return commonlogger.InitFromOptions(o.Logging)
}
