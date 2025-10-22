package orchestrator

import (
	"github.com/kart-io/k8s-agent/internal/reasoning/agents/k8s_tool"
	"github.com/kart-io/k8s-agent/internal/reasoning/agents/reasoning"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/description"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/root_cause"
	"github.com/kart-io/k8s-agent/internal/reasoning/memory"
	"time"
)

// Orchestrator 协调所有分析组件的主控制器
type Orchestrator struct {
	// Agents
	reasoningAgent *reasoning.ReasoningAgent

	// Chains
	rootCauseChain   *root_cause.RootCauseChain
	descriptionChain *description.DescriptionChain

	// Tools
	k8sTool *k8s_tool.K8sTool

	// Memory
	memoryManager memory.Manager

	// 配置
	config *OrchestratorConfig
}

// OrchestratorConfig Orchestrator 配置
type OrchestratorConfig struct {
	// 功能开关
	EnableMemory      bool `json:"enable_memory"`      // 是否启用 Memory 系统
	EnableToolAgent   bool `json:"enable_tool_agent"`  // 是否启用 Tool Agent
	EnableDescription bool `json:"enable_description"` // 是否启用故障描述生成

	// Memory 配置
	LoadHistoryContext bool `json:"load_history_context"` // 是否加载历史上下文
	LoadSimilarCases   bool `json:"load_similar_cases"`   // 是否加载相似案例
	SaveToMemory       bool `json:"save_to_memory"`       // 是否保存到 Memory
	MaxSimilarCases    int  `json:"max_similar_cases"`    // 最大相似案例数量

	// Agent 配置
	MaxToolCalls int `json:"max_tool_calls"` // 最大工具调用次数

	// 超时配置
	Timeout            time.Duration `json:"timeout"`             // 总超时时间
	RootCauseTimeout   time.Duration `json:"root_cause_timeout"`  // 根因分析超时
	DescriptionTimeout time.Duration `json:"description_timeout"` // 描述生成超时
	ToolAgentTimeout   time.Duration `json:"tool_agent_timeout"`  // Tool Agent 超时
}

// AnalysisRequest Orchestrator 分析请求
type AnalysisRequest struct {
	// 基本信息
	SessionID    string    `json:"session_id"`    // 会话 ID
	FailureType  string    `json:"failure_type"`  // 故障类型
	ResourceType string    `json:"resource_type"` // 资源类型
	ResourceName string    `json:"resource_name"` // 资源名称
	Namespace    string    `json:"namespace"`     // 命名空间
	ClusterID    string    `json:"cluster_id"`    // 集群 ID
	ErrorMessage string    `json:"error_message"` // 错误信息
	Timestamp    time.Time `json:"timestamp"`     // 故障时间

	// 分析选项
	Language    string `json:"language"`     // 描述语言
	DetailLevel string `json:"detail_level"` // 详细程度

	// 上下文信息（可选）
	PodLogs        string                `json:"pod_logs"`
	Events         []k8s_tool.EventInfo  `json:"events"`
	Metrics        *k8s_tool.MetricsInfo `json:"metrics"`
	ResourceStatus map[string]string     `json:"resource_status"`
}

// AnalysisResponse Orchestrator 分析响应
type AnalysisResponse struct {
	// 分析结果
	RootCause   *root_cause.AnalysisOutput     `json:"root_cause"`  // 根因分析
	Description *description.DescriptionOutput `json:"description"` // 故障描述

	// 上下文信息
	SimilarCases      []*memory.CaseMemory `json:"similar_cases"`      // 相似案例
	ConversationCount int                  `json:"conversation_count"` // 对话数量

	// 执行信息
	ExecutionSteps []ExecutionStep `json:"execution_steps"` // 执行步骤
	TotalLatency   time.Duration   `json:"total_latency"`   // 总延迟
	Timestamp      time.Time       `json:"timestamp"`       // 响应时间

	// 错误信息
	Error        string `json:"error,omitempty"`         // 错误信息
	PartialError string `json:"partial_error,omitempty"` // 部分失败信息
}

// ExecutionStep 执行步骤
type ExecutionStep struct {
	Step        int           `json:"step"`            // 步骤编号
	Name        string        `json:"name"`            // 步骤名称
	Description string        `json:"description"`     // 步骤描述
	Status      string        `json:"status"`          // 状态: success, failed, skipped
	Duration    time.Duration `json:"duration"`        // 耗时
	Error       string        `json:"error,omitempty"` // 错误信息
}

// DefaultOrchestratorConfig 返回默认配置
func DefaultOrchestratorConfig() *OrchestratorConfig {
	return &OrchestratorConfig{
		EnableMemory:       true,
		EnableToolAgent:    false, // 默认关闭
		EnableDescription:  true,
		LoadHistoryContext: true,
		LoadSimilarCases:   true,
		SaveToMemory:       true,
		MaxSimilarCases:    5,
		MaxToolCalls:       3,
		Timeout:            60 * time.Second,
		RootCauseTimeout:   30 * time.Second,
		DescriptionTimeout: 20 * time.Second,
		ToolAgentTimeout:   20 * time.Second,
	}
}
