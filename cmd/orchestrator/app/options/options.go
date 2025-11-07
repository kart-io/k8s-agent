// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package options contains flags and options for initializing the orchestrator service
package options

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/kart-io/k8s-agent/common/loggerutil"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/logger/core"
)

const (
	// UserAgent is the userAgent name when starting orchestrator server.
	UserAgent = "aetherius-orchestrator"

	// Default timeout and retry settings
	defaultAITimeout   = 30 * time.Second
	defaultMaxRetries  = 3
	defaultHTTPTimeout = 30 * time.Second

	// Default workflow timeout settings
	defaultGlobalTimeout      = 30 * time.Minute
	defaultStepDefaultTimeout = 5 * time.Minute
	defaultWorkflowMaxRetries = 3
)

// ServerOptions contains the configuration options for the orchestrator server.
type ServerOptions struct {
	// Server options for configuring HTTP server related options.
	Server *commonoptions.ServerOptions `json:"server" mapstructure:"server"`
	// Database options for configuring MySQL database related options.
	Database *commonoptions.MySQLOptions `json:"database" mapstructure:"database"`
	// Redis options for configuring Redis related options.
	Redis *commonoptions.RedisOptions `json:"redis" mapstructure:"redis"`
	// NATS options for configuring NATS message queue related options.
	NATS *commonoptions.NATSOptions `json:"nats" mapstructure:"nats"`
	// GRPC options for configuring gRPC server related options.
	GRPC *commonoptions.GRPCOptions `json:"grpc" mapstructure:"grpc"`
	// Logging options for configuring logging related options.
	Logging *commonoptions.LoggingOptions `json:"logging" mapstructure:"logging"`
	// Health options for configuring health check related options.
	Health *commonoptions.HealthOptions `json:"health" mapstructure:"health"`
	// Metrics options for configuring metrics related options.
	Metrics *commonoptions.MetricsOptions `json:"metrics" mapstructure:"metrics"`
	// AI options for configuring AI service integration.
	AI *AIOptions `json:"ai" mapstructure:"ai"`
	// Workflow options for configuring workflow timeout settings.
	Workflow *WorkflowOptions `json:"workflow" mapstructure:"workflow"`
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

// WorkflowOptions contains workflow timeout configuration options.
type WorkflowOptions struct {
	// GlobalTimeout is the maximum time a workflow can run.
	GlobalTimeout time.Duration `json:"global_timeout" mapstructure:"global_timeout"`
	// StepDefaultTimeout is the default timeout for workflow steps.
	StepDefaultTimeout time.Duration `json:"step_default_timeout" mapstructure:"step_default_timeout"`
	// RetryOnTimeout indicates whether to retry workflows that timeout.
	RetryOnTimeout bool `json:"retry_on_timeout" mapstructure:"retry_on_timeout"`
	// MaxRetries is the maximum number of retry attempts for timed-out workflows.
	MaxRetries int `json:"max_retries" mapstructure:"max_retries"`
}

// NewWorkflowOptions creates default WorkflowOptions.
func NewWorkflowOptions() *WorkflowOptions {
	return &WorkflowOptions{
		GlobalTimeout:      defaultGlobalTimeout,
		StepDefaultTimeout: defaultStepDefaultTimeout,
		RetryOnTimeout:     true,
		MaxRetries:         defaultWorkflowMaxRetries,
	}
}

// AddFlags adds flags for workflow options.
func (o *WorkflowOptions) AddFlags(fs *pflag.FlagSet) {
	fs.DurationVar(&o.GlobalTimeout, "workflow.global-timeout", o.GlobalTimeout,
		"Maximum time a workflow can run")
	fs.DurationVar(&o.StepDefaultTimeout, "workflow.step-default-timeout", o.StepDefaultTimeout,
		"Default timeout for workflow steps")
	fs.BoolVar(&o.RetryOnTimeout, "workflow.retry-on-timeout", o.RetryOnTimeout,
		"Whether to retry workflows that timeout")
	fs.IntVar(&o.MaxRetries, "workflow.max-retries", o.MaxRetries,
		"Maximum number of retry attempts for timed-out workflows")
}

// Validate validates workflow options.
func (o *WorkflowOptions) Validate() error {
	if o.GlobalTimeout <= 0 {
		o.GlobalTimeout = defaultGlobalTimeout
	}
	if o.StepDefaultTimeout <= 0 {
		o.StepDefaultTimeout = defaultStepDefaultTimeout
	}
	if o.MaxRetries < 0 {
		o.MaxRetries = 0
	}
	return nil
}

// Complete fills in any fields not set that are required to have valid data.
func (o *WorkflowOptions) Complete() error {
	if o.GlobalTimeout == 0 {
		o.GlobalTimeout = defaultGlobalTimeout
	}
	if o.StepDefaultTimeout == 0 {
		o.StepDefaultTimeout = defaultStepDefaultTimeout
	}
	if o.MaxRetries == 0 {
		o.MaxRetries = defaultWorkflowMaxRetries
	}
	return nil
}

func NewAIOptions() *AIOptions {
	return &AIOptions{
		ReasoningServiceURL: "http://localhost:8083", // Reasoning service 端口
		AgentManagerURL:     "http://localhost:8081", // Agent Manager 端口
		Timeout:             defaultAITimeout,
		MaxRetries:          defaultMaxRetries,
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
		o.ReasoningServiceURL = "http://localhost:8083"
	}
	if o.AgentManagerURL == "" {
		o.AgentManagerURL = "http://localhost:8081"
	}
	if o.Timeout == 0 {
		o.Timeout = defaultAITimeout
	}
	if o.MaxRetries == 0 {
		o.MaxRetries = defaultMaxRetries
	}
	return nil
}

// Ensure ServerOptions implements the commonapp.Options interface.
var _ commonapp.Options = (*ServerOptions)(nil)

// NewServerOptions creates a ServerOptions instance with default values.
func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		Server:   commonoptions.NewServerOptions(),
		Database: commonoptions.NewMySQLOptions(),
		Redis:    commonoptions.NewRedisOptions(),
		NATS:     commonoptions.NewNATSOptions(),
		GRPC:     commonoptions.NewGRPCOptions(),
		Logging:  commonoptions.NewLoggingOptions(),
		Health:   commonoptions.NewHealthOptions(),
		Metrics:  commonoptions.NewMetricsOptions(),
		AI:       NewAIOptions(),
		Workflow: NewWorkflowOptions(),
	}
}

// GetHealthPort returns the health check port
func (o *ServerOptions) GetHealthPort() int {
	return o.Health.Port
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
	return "Orchestrator"
}

// GetLogFields returns log fields for initialization logging.
func (o *ServerOptions) GetLogFields() []interface{} {
	return []interface{}{
		"http_port", o.Server.Port,
		"health_port", o.GetHealthPort(),
		"reasoning_service_url", o.AI.ReasoningServiceURL,
		"agent_manager_url", o.AI.AgentManagerURL,
	}
}

// InitLogger initializes logger based on the options.
func (o *ServerOptions) InitLogger() (core.Logger, error) {
	// Use the common logger initialization
	return loggerutil.InitFromOptions(o.Logging)
}
