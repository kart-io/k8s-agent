// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/spf13/pflag"

	"github.com/kart-io/k8s-agent/common/loggerutil"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	pkgoptions "github.com/kart-io/k8s-agent/pkg/options"
	"github.com/kart-io/logger/core"
)

// StandardOptions 提供统一的服务配置选项，消除各服务的重复代码。
// 使用建造者模式允许服务选择需要的配置组件。
//
// 用法示例:
//
//	opts := app.NewStandardOptions("agent-manager", "aetherius-agent-manager").
//	    WithDatabase().
//	    WithRedis().
//	    WithNATS().
//	    WithMetrics()
type StandardOptions struct {
	// serviceName 服务名称（用于日志）
	serviceName string
	// userAgent 用户代理字符串（用于 HTTP 请求标识）
	userAgent string

	// 必需字段（所有服务都包含）
	Server  *commonoptions.ServerOptions  `json:"server" mapstructure:"server"`
	GRPC    *commonoptions.GRPCOptions    `json:"grpc" mapstructure:"grpc"`
	Logging *commonoptions.LoggingOptions `json:"logging" mapstructure:"logging"`
	Health  *commonoptions.HealthOptions  `json:"health" mapstructure:"health"`

	// 可选字段（按需启用）
	Database   *commonoptions.MySQLOptions      `json:"database,omitempty" mapstructure:"database"`
	Redis      *commonoptions.RedisOptions      `json:"redis,omitempty" mapstructure:"redis"`
	NATS       *commonoptions.NATSOptions       `json:"nats,omitempty" mapstructure:"nats"`
	JWT        *commonoptions.JWTOptions        `json:"jwt,omitempty" mapstructure:"jwt"`
	Email      *pkgoptions.EmailOptions         `json:"email,omitempty" mapstructure:"email"`
	Metrics    *commonoptions.MetricsOptions    `json:"metrics,omitempty" mapstructure:"metrics"`
	Prometheus *commonoptions.PrometheusOptions `json:"prometheus,omitempty" mapstructure:"prometheus"`
	Alert      *pkgoptions.AlertOptions         `json:"alert,omitempty" mapstructure:"alert"`
	LLM        *pkgoptions.LLMOptions           `json:"llm,omitempty" mapstructure:"llm"`
	Agent      *pkgoptions.AgentOptions         `json:"agent,omitempty" mapstructure:"agent"`
	AI         *pkgoptions.AIOptions            `json:"ai,omitempty" mapstructure:"ai"`
	Workflow   *commonoptions.WorkflowOptions   `json:"workflow,omitempty" mapstructure:"workflow"`
	CORS       *commonoptions.CORSOptions       `json:"cors,omitempty" mapstructure:"cors"`
	RateLimit  *commonoptions.RateLimitOptions  `json:"rate_limit,omitempty" mapstructure:"rate_limit"`
}

// Ensure StandardOptions implements the Options interface.
var _ Options = (*StandardOptions)(nil)

// NewStandardOptions 创建一个 StandardOptions 实例，包含必需的基础配置。
//
// Parameters:
//   - serviceName: 服务名称（用于日志，如 "Agent Manager"）
//   - userAgent: 用户代理字符串（如 "aetherius-agent-manager"）
//
// Returns:
//   - *StandardOptions: 包含基础配置的选项实例
func NewStandardOptions(serviceName, userAgent string) *StandardOptions {
	return &StandardOptions{
		serviceName: serviceName,
		userAgent:   userAgent,
		// 所有服务都需要的基础配置
		Server:  commonoptions.NewServerOptions(),
		GRPC:    commonoptions.NewGRPCOptions(),
		Logging: commonoptions.NewLoggingOptions(),
		Health:  commonoptions.NewHealthOptions(),
	}
}

// WithDatabase 启用数据库配置（MySQL）。
func (o *StandardOptions) WithDatabase() *StandardOptions {
	if o.Database == nil {
		o.Database = commonoptions.NewMySQLOptions()
	}
	return o
}

// WithRedis 启用 Redis 配置。
func (o *StandardOptions) WithRedis() *StandardOptions {
	if o.Redis == nil {
		o.Redis = commonoptions.NewRedisOptions()
	}
	return o
}

// WithNATS 启用 NATS 消息队列配置。
func (o *StandardOptions) WithNATS() *StandardOptions {
	if o.NATS == nil {
		o.NATS = commonoptions.NewNATSOptions()
	}
	return o
}

// WithJWT 启用 JWT 认证配置。
func (o *StandardOptions) WithJWT() *StandardOptions {
	if o.JWT == nil {
		o.JWT = commonoptions.NewJWTOptions()
	}
	return o
}

// WithEmail 启用邮件发送配置。
func (o *StandardOptions) WithEmail() *StandardOptions {
	if o.Email == nil {
		o.Email = pkgoptions.NewEmailOptions()
	}
	return o
}

// WithMetrics 启用 Prometheus 指标配置。
func (o *StandardOptions) WithMetrics() *StandardOptions {
	if o.Metrics == nil {
		o.Metrics = commonoptions.NewMetricsOptions()
	}
	return o
}

// WithLLM 启用 LLM 配置（AI 模型）。
func (o *StandardOptions) WithLLM() *StandardOptions {
	if o.LLM == nil {
		o.LLM = pkgoptions.NewLLMOptions()
	}
	return o
}

// WithAgent 启用 Agent 配置（Collect Agent）。
func (o *StandardOptions) WithAgent() *StandardOptions {
	if o.Agent == nil {
		o.Agent = pkgoptions.NewAgentOptions()
	}
	return o
}

// WithAI 启用 AI 配置（Orchestrator）。
func (o *StandardOptions) WithAI() *StandardOptions {
	if o.AI == nil {
		o.AI = pkgoptions.NewAIOptions()
	}
	return o
}

// WithWorkflow 启用 Workflow 配置（Orchestrator）。
func (o *StandardOptions) WithWorkflow() *StandardOptions {
	if o.Workflow == nil {
		o.Workflow = commonoptions.NewWorkflowOptions()
	}
	return o
}

// WithCORS 启用 CORS 配置。
func (o *StandardOptions) WithCORS() *StandardOptions {
	if o.CORS == nil {
		o.CORS = commonoptions.NewCORSOptions()
	}
	return o
}

// WithRateLimit 启用 RateLimit 配置。
func (o *StandardOptions) WithRateLimit() *StandardOptions {
	if o.RateLimit == nil {
		o.RateLimit = commonoptions.NewRateLimitOptions()
	}
	return o
}

// WithPrometheus 启用 Prometheus 指标配置。
func (o *StandardOptions) WithPrometheus() *StandardOptions {
	if o.Prometheus == nil {
		o.Prometheus = commonoptions.NewPrometheusOptions()
	}
	return o
}

// WithAlert 启用告警配置。
func (o *StandardOptions) WithAlert() *StandardOptions {
	if o.Alert == nil {
		o.Alert = pkgoptions.NewAlertOptions()
	}
	return o
}

// AddFlags 添加命令行标志到指定的 FlagSet。
// 实现 Options 接口。
func (o *StandardOptions) AddFlags(fs *pflag.FlagSet) {
	// 使用通用工具函数统一添加所有子选项的 flags
	commonoptions.AddFlagsAll(o, fs)
}

// Complete 完成所有必需的选项配置。
// 实现 Options 接口。
func (o *StandardOptions) Complete() error {
	// 使用通用工具函数：设置服务名称并完成所有子选项
	return commonoptions.CompleteWithServiceName(o, o.Logging, o.userAgent)
}

// Validate 验证 StandardOptions 中的选项是否有效。
// 实现 Options 接口。
func (o *StandardOptions) Validate() []error {
	// 使用通用工具函数统一验证所有子选项
	return commonoptions.ValidateAll(o)
}

// GetServiceName 返回服务名称。
func (o *StandardOptions) GetServiceName() string {
	return o.serviceName
}

// GetLogFields 返回用于初始化日志的字段。
// 返回通用字段，各服务可在自己的 app.go 中添加额外字段。
func (o *StandardOptions) GetLogFields() []interface{} {
	fields := []interface{}{
		"http_port", o.Server.Port,
		"grpc_enabled", o.GRPC.Enable,
	}

	if o.GRPC.Enable {
		fields = append(fields, "grpc_port", o.GRPC.Port)
	}

	if o.Health != nil {
		fields = append(fields, "health_port", o.Health.Port)
	}

	return fields
}

// InitLogger 根据配置选项初始化日志器。
func (o *StandardOptions) InitLogger() (core.Logger, error) {
	// 使用通用日志初始化工具
	return loggerutil.InitFromOptions(o.Logging)
}

// GetHealthPort 返回健康检查端口。
func (o *StandardOptions) GetHealthPort() int {
	if o.Health == nil {
		return 0
	}
	return o.Health.Port
}

// HasDatabase 检查是否启用了数据库。
func (o *StandardOptions) HasDatabase() bool {
	return o.Database != nil
}

// HasRedis 检查是否启用了 Redis。
func (o *StandardOptions) HasRedis() bool {
	return o.Redis != nil
}

// HasNATS 检查是否启用了 NATS。
func (o *StandardOptions) HasNATS() bool {
	return o.NATS != nil
}

// HasJWT 检查是否启用了 JWT。
func (o *StandardOptions) HasJWT() bool {
	return o.JWT != nil
}

// HasEmail 检查是否启用了邮件。
func (o *StandardOptions) HasEmail() bool {
	return o.Email != nil
}

// HasMetrics 检查是否启用了指标。
func (o *StandardOptions) HasMetrics() bool {
	return o.Metrics != nil
}
