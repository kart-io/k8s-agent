package description

import (
	"context"
	"time"
)

// Chain 定义故障描述 Chain 的接口.
type Chain interface {
	// Generate 生成故障描述
	Generate(ctx context.Context, input *DescriptionInput) (*DescriptionOutput, error)
}

// DescriptionInput 故障描述生成输入.
type DescriptionInput struct {
	// 故障基本信息
	FailureType  string    `json:"failure_type"`  // 故障类型
	ResourceType string    `json:"resource_type"` // 资源类型
	ResourceName string    `json:"resource_name"` // 资源名称
	Namespace    string    `json:"namespace"`     // 命名空间
	ClusterID    string    `json:"cluster_id"`    // 集群 ID
	Timestamp    time.Time `json:"timestamp"`     // 故障时间

	// 故障详情
	ErrorMessage string   `json:"error_message"` // 错误信息
	Symptoms     []string `json:"symptoms"`      // 症状列表
	Impact       string   `json:"impact"`        // 影响范围

	// 上下文信息
	PodEvents      []PodEvent         `json:"pod_events"`      // Pod 事件
	PodLogs        string             `json:"pod_logs"`        // Pod 日志
	ResourceStatus map[string]string  `json:"resource_status"` // 资源状态
	Metrics        map[string]float64 `json:"metrics"`         // 指标数据

	// 根因分析结果 (可选,如果已完成根因分析)
	RootCause *RootCauseInfo `json:"root_cause"` // 根因信息

	// 生成配置
	Language        string `json:"language"`         // 目标语言: "zh", "en"
	DetailLevel     string `json:"detail_level"`     // 详细程度: "brief", "normal", "detailed"
	IncludeTimeline bool   `json:"include_timeline"` // 是否包含时间线
}

// DescriptionOutput 故障描述输出.
type DescriptionOutput struct {
	// 描述内容
	Title       string `json:"title"`       // 标题
	Summary     string `json:"summary"`     // 摘要
	Description string `json:"description"` // 详细描述
	Severity    string `json:"severity"`    // 严重程度: "critical", "high", "medium", "low"

	// 影响分析
	AffectedComponents []string `json:"affected_components"` // 受影响的组件
	UserImpact         string   `json:"user_impact"`         // 用户影响
	BusinessImpact     string   `json:"business_impact"`     // 业务影响

	// 时间线
	Timeline []TimelineEvent `json:"timeline"` // 事件时间线

	// 技术细节
	TechnicalDetails map[string]string `json:"technical_details"` // 技术细节

	// 元数据
	Language    string        `json:"language"`     // 语言
	GeneratedAt time.Time     `json:"generated_at"` // 生成时间
	Provider    string        `json:"provider"`     // LLM 提供商
	Model       string        `json:"model"`        // 模型
	TokensUsed  int           `json:"tokens_used"`  // Token 使用量
	Latency     time.Duration `json:"latency"`      // 延迟
}

// PodEvent Pod 事件.
type PodEvent struct {
	Type      string    `json:"type"`      // 事件类型
	Reason    string    `json:"reason"`    // 原因
	Message   string    `json:"message"`   // 消息
	Timestamp time.Time `json:"timestamp"` // 时间戳
	Source    string    `json:"source"`    // 来源
}

// RootCauseInfo 根因信息 (来自根因分析 Chain).
type RootCauseInfo struct {
	RootCause  string  `json:"root_cause"` // 根本原因
	Confidence float64 `json:"confidence"` // 置信度
	Category   string  `json:"category"`   // 类别
	Reasoning  string  `json:"reasoning"`  // 推理过程
}

// TimelineEvent 时间线事件.
type TimelineEvent struct {
	Timestamp time.Time `json:"timestamp"` // 时间戳
	Event     string    `json:"event"`     // 事件描述
	Severity  string    `json:"severity"`  // 严重程度
	Component string    `json:"component"` // 相关组件
}

// ChainConfig 故障描述 Chain 配置.
type ChainConfig struct {
	// LLM 配置
	Temperature float64 `json:"temperature"` // 温度参数
	MaxTokens   int     `json:"max_tokens"`  // 最大 token 数

	// 描述配置
	DefaultLanguage    string `json:"default_language"`     // 默认语言
	DefaultDetailLevel string `json:"default_detail_level"` // 默认详细程度
	IncludeTimeline    bool   `json:"include_timeline"`     // 默认是否包含时间线

	// Prompt 配置
	SystemPrompt string `json:"system_prompt"` // 系统提示

	// 超时配置
	Timeout time.Duration `json:"timeout"` // 超时时间
}

// SupportedLanguages 支持的语言.
var SupportedLanguages = map[string]string{
	"zh": "Chinese (Simplified)",
	"en": "English",
}

// SupportedDetailLevels 支持的详细程度.
var SupportedDetailLevels = []string{
	"brief",    // 简要描述
	"normal",   // 正常描述
	"detailed", // 详细描述
}

// SupportedSeverities 支持的严重程度.
var SupportedSeverities = []string{
	"critical", // 紧急
	"high",     // 高
	"medium",   // 中
	"low",      // 低
}
