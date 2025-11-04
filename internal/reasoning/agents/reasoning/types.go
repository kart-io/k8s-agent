package reasoning

import (
	"context"
	"time"

	"github.com/kart-io/k8s-agent/internal/reasoning/agents/k8s_tool"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/description"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/root_cause"
)

// Agent 定义 Reasoning Agent 接口
type Agent interface {
	// Analyze 执行完整的故障分析推理
	Analyze(ctx context.Context, input *AnalysisInput) (*AnalysisOutput, error)

	// AnalyzeRootCause 仅执行根因分析
	AnalyzeRootCause(ctx context.Context, input *AnalysisInput) (*root_cause.AnalysisOutput, error)

	// GenerateDescription 仅生成故障描述
	GenerateDescription(ctx context.Context, input *AnalysisInput) (*description.DescriptionOutput, error)
}

// AnalysisInput Reasoning Agent 分析输入
type AnalysisInput struct {
	// 故障基本信息
	FailureType  string    `json:"failure_type"`  // 故障类型
	ResourceType string    `json:"resource_type"` // 资源类型
	ResourceName string    `json:"resource_name"` // 资源名称
	Namespace    string    `json:"namespace"`     // 命名空间
	ClusterID    string    `json:"cluster_id"`    // 集群 ID
	ErrorMessage string    `json:"error_message"` // 错误信息
	Timestamp    time.Time `json:"timestamp"`     // 故障时间

	// 分析选项
	IncludeRootCause   bool   `json:"include_root_cause"`  // 是否包含根因分析
	IncludeDescription bool   `json:"include_description"` // 是否包含故障描述
	IncludeK8sContext  bool   `json:"include_k8s_context"` // 是否获取 K8s 上下文
	Language           string `json:"language"`            // 描述语言: "zh", "en"
	DetailLevel        string `json:"detail_level"`        // 详细程度

	// K8s 工具选项
	FetchLogs    bool  `json:"fetch_logs"`    // 是否获取日志
	FetchEvents  bool  `json:"fetch_events"`  // 是否获取事件
	FetchMetrics bool  `json:"fetch_metrics"` // 是否获取指标
	MaxLogLines  int64 `json:"max_log_lines"` // 最大日志行数

	// 上下文信息（可选，如果已有可直接提供）
	PodLogs        string                `json:"pod_logs"`
	Events         []k8s_tool.EventInfo  `json:"events"`
	Metrics        *k8s_tool.MetricsInfo `json:"metrics"`
	ResourceStatus map[string]string     `json:"resource_status"`
}

// AnalysisOutput Reasoning Agent 分析输出
type AnalysisOutput struct {
	// 分析结果
	RootCause   *root_cause.AnalysisOutput     `json:"root_cause"`  // 根因分析结果
	Description *description.DescriptionOutput `json:"description"` // 故障描述

	// K8s 上下文
	K8sContext *K8sContext `json:"k8s_context"` // K8s 上下文信息

	// 推理过程
	ReasoningSteps []ReasoningStep `json:"reasoning_steps"` // 推理步骤

	// 元数据
	TotalLatency time.Duration `json:"total_latency"` // 总延迟
	Timestamp    time.Time     `json:"timestamp"`     // 时间戳
}

// K8sContext K8s 上下文信息
type K8sContext struct {
	// 资源信息
	PodInfo        *k8s_tool.PodInfo        `json:"pod_info"`
	DeploymentInfo *k8s_tool.DeploymentInfo `json:"deployment_info"`
	ServiceInfo    *k8s_tool.ServiceInfo    `json:"service_info"`

	// 日志和事件
	Logs   string               `json:"logs"`
	Events []k8s_tool.EventInfo `json:"events"`

	// 指标
	Metrics *k8s_tool.MetricsInfo `json:"metrics"`

	// 获取时间
	FetchedAt time.Time `json:"fetched_at"`
}

// ReasoningStep 推理步骤
type ReasoningStep struct {
	Step        int           `json:"step"`        // 步骤编号
	Action      string        `json:"action"`      // 执行的操作
	Description string        `json:"description"` // 操作描述
	Result      string        `json:"result"`      // 操作结果
	Duration    time.Duration `json:"duration"`    // 耗时
	Success     bool          `json:"success"`     // 是否成功
	Error       string        `json:"error"`       // 错误信息
}

// AgentConfig Reasoning Agent 配置
type AgentConfig struct {
	// Chains 配置
	RootCauseConfig   *root_cause.ChainConfig  `json:"root_cause_config"`
	DescriptionConfig *description.ChainConfig `json:"description_config"`

	// K8s Tool 配置
	K8sToolConfig *k8s_tool.ToolConfig `json:"k8s_tool_config"`

	// Agent 配置
	EnableK8sTool     bool `json:"enable_k8s_tool"`    // 是否启用 K8s 工具
	EnableRootCause   bool `json:"enable_root_cause"`  // 是否启用根因分析
	EnableDescription bool `json:"enable_description"` // 是否启用故障描述

	// 默认选项
	DefaultLanguage    string `json:"default_language"`     // 默认语言
	DefaultDetailLevel string `json:"default_detail_level"` // 默认详细程度

	// 超时配置
	Timeout time.Duration `json:"timeout"` // 总超时时间
}

// AnalysisPhase 分析阶段
type AnalysisPhase string

const (
	PhaseInit        AnalysisPhase = "init"        // 初始化
	PhaseK8sContext  AnalysisPhase = "k8s_context" // 获取 K8s 上下文
	PhaseRootCause   AnalysisPhase = "root_cause"  // 根因分析
	PhaseDescription AnalysisPhase = "description" // 故障描述
	PhaseComplete    AnalysisPhase = "complete"    // 完成
)

// AnalysisError 分析错误
type AnalysisError struct {
	Phase   AnalysisPhase `json:"phase"`   // 发生错误的阶段
	Message string        `json:"message"` // 错误信息
	Err     error         `json:"-"`       // 原始错误
}

func (e *AnalysisError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AnalysisError) Unwrap() error {
	return e.Err
}
