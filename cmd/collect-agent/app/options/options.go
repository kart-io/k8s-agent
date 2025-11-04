package options

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/kart-io/k8s-agent/common/loggerutil"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/internal/collect-agent/types"
	"github.com/kart-io/logger/core"
)

// ServerOptions 定义 collect-agent 服务的配置选项
// 实现 pkg/app.Options 接口
type ServerOptions struct {
	Logging *commonoptions.LoggingOptions `json:"logging" mapstructure:"logging"`
	Health  *commonoptions.HealthOptions  `json:"health" mapstructure:"health"`
	Agent   *commonoptions.AgentOptions   `json:"agent" mapstructure:"agent"`
}

// NewServerOptions 创建新的 ServerOptions 实例，使用默认值
func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		Logging: commonoptions.NewLoggingOptions(),
		Health:  commonoptions.NewHealthOptions(),
		Agent:   commonoptions.NewAgentOptions(),
	}
}

// Validate 验证所有必需的配置选项
func (o *ServerOptions) Validate() []error {
	// 使用通用工具函数统一验证所有子选项
	return commonoptions.ValidateAll(o)
}

// Complete 填充未设置但需要有效数据的字段
func (o *ServerOptions) Complete() error {
	// 使用通用工具函数统一完成所有子选项
	return commonoptions.CompleteAll(o)
}

// AddFlags 添加 flags 到 flag set
// 注意: --config/-c flag 由 pkg/app 框架自动添加
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	// 使用通用工具函数统一添加所有子选项的 flags
	commonoptions.AddFlagsAll(o, fs)
}

// InitLogger 基于配置初始化 logger
func (o *ServerOptions) InitLogger() (core.Logger, error) {
	return loggerutil.InitFromOptions(o.Logging)
}

// GetServiceName 返回服务名称
func (o *ServerOptions) GetServiceName() string {
	return "CollectAgent"
}

// GetLogFields 返回初始化日志的字段
func (o *ServerOptions) GetLogFields() []interface{} {
	return []interface{}{
		"cluster_id", o.Agent.ClusterID,
		"central_endpoint", o.Agent.CentralEndpoint,
		"health_port", o.Health.Port,
	}
}

// 向后兼容方法 - 映射到新的 Agent options 结构

// GetClusterID 返回集群 ID (向后兼容)
func (o *ServerOptions) GetClusterID() string {
	return o.Agent.ClusterID
}

// GetClusterName 返回集群名称 (向后兼容)
func (o *ServerOptions) GetClusterName() string {
	return o.Agent.ClusterName
}

// GetCentralEndpoint 返回中央端点 (向后兼容)
func (o *ServerOptions) GetCentralEndpoint() string {
	return o.Agent.CentralEndpoint
}

// GetReconnectDelay 返回重连延迟 (向后兼容)
func (o *ServerOptions) GetReconnectDelay() time.Duration {
	return o.Agent.ReconnectDelay
}

// GetHeartbeatInterval 返回心跳间隔 (向后兼容)
func (o *ServerOptions) GetHeartbeatInterval() time.Duration {
	return o.Agent.HeartbeatInterval
}

// GetMetricsInterval 返回指标间隔 (向后兼容)
func (o *ServerOptions) GetMetricsInterval() time.Duration {
	return o.Agent.MetricsInterval
}

// GetBufferSize 返回缓冲区大小 (向后兼容)
func (o *ServerOptions) GetBufferSize() int {
	return o.Agent.BufferSize
}

// GetMaxRetries 返回最大重试次数 (向后兼容)
func (o *ServerOptions) GetMaxRetries() int {
	return o.Agent.MaxRetries
}

// IsMetricsEnabled 返回是否启用指标 (向后兼容)
func (o *ServerOptions) IsMetricsEnabled() bool {
	return o.Agent.EnableMetrics
}

// IsEventsEnabled 返回是否启用事件 (向后兼容)
func (o *ServerOptions) IsEventsEnabled() bool {
	return o.Agent.EnableEvents
}

// GetHealthPort 返回健康检查端口 (向后兼容)
func (o *ServerOptions) GetHealthPort() int {
	return o.Health.Port
}

// ToAgentConfig 将 Options 转换为 types.AgentConfig (向后兼容)
func (o *ServerOptions) ToAgentConfig() *types.AgentConfig {
	return &types.AgentConfig{
		ClusterID:         o.Agent.ClusterID,
		ClusterName:       o.Agent.ClusterName,
		CentralEndpoint:   o.Agent.CentralEndpoint,
		ReconnectDelay:    o.Agent.ReconnectDelay,
		HeartbeatInterval: o.Agent.HeartbeatInterval,
		MetricsInterval:   o.Agent.MetricsInterval,
		BufferSize:        o.Agent.BufferSize,
		MaxRetries:        o.Agent.MaxRetries,
		LogLevel:          o.Logging.Level,
		EnableMetrics:     o.Agent.EnableMetrics,
		EnableEvents:      o.Agent.EnableEvents,
		HealthPort:        o.Health.Port,
	}
}
