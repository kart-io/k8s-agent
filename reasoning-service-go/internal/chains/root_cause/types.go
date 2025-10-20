package root_cause

import (
	"context"
	"time"
)

// Chain 定义根因分析 Chain 的接口
type Chain interface {
	// Analyze 执行根因分析
	Analyze(ctx context.Context, input *AnalysisInput) (*AnalysisOutput, error)
}

// AnalysisInput 根因分析输入
type AnalysisInput struct {
	// 故障信息
	FailureType  string    `json:"failure_type"`  // 故障类型: "pod_failure", "service_down", etc.
	ResourceType string    `json:"resource_type"` // 资源类型: "pod", "deployment", "service", etc.
	ResourceName string    `json:"resource_name"` // 资源名称
	Namespace    string    `json:"namespace"`     // 命名空间
	ClusterID    string    `json:"cluster_id"`    // 集群 ID
	ErrorMessage string    `json:"error_message"` // 错误信息
	Symptoms     []string  `json:"symptoms"`      // 症状列表
	Timestamp    time.Time `json:"timestamp"`     // 故障发生时间

	// K8s 上下文信息
	PodEvents      []K8sEvent         `json:"pod_events"`      // Pod 事件
	PodLogs        string             `json:"pod_logs"`        // Pod 日志
	ResourceStatus map[string]string  `json:"resource_status"` // 资源状态
	Metrics        map[string]float64 `json:"metrics"`         // 指标数据

	// 额外上下文
	SimilarCases   []SimilarCase          `json:"similar_cases"`   // 相似案例
	HistoricalData map[string]interface{} `json:"historical_data"` // 历史数据
}

// AnalysisOutput 根因分析输出
type AnalysisOutput struct {
	// 根因分析结果
	RootCause  string  `json:"root_cause"` // 根本原因描述
	Confidence float64 `json:"confidence"` // 置信度 (0.0-1.0)
	Reasoning  string  `json:"reasoning"`  // 推理过程
	Category   string  `json:"category"`   // 根因类别

	// 贡献因素
	ContributingFactors []Factor `json:"contributing_factors"` // 贡献因素列表

	// 建议
	Recommendations []Recommendation `json:"recommendations"` // 修复建议

	// 元数据
	AnalysisTime time.Time     `json:"analysis_time"` // 分析时间
	Provider     string        `json:"provider"`      // 使用的 LLM 提供商
	Model        string        `json:"model"`         // 使用的模型
	TokensUsed   int           `json:"tokens_used"`   // 使用的 token 数
	Latency      time.Duration `json:"latency"`       // 延迟
}

// K8sEvent K8s 事件
type K8sEvent struct {
	Type      string    `json:"type"`      // 事件类型: "Normal", "Warning"
	Reason    string    `json:"reason"`    // 原因
	Message   string    `json:"message"`   // 消息
	Timestamp time.Time `json:"timestamp"` // 时间戳
	Source    string    `json:"source"`    // 来源
}

// SimilarCase 相似案例
type SimilarCase struct {
	CaseID      string  `json:"case_id"`     // 案例 ID
	Description string  `json:"description"` // 描述
	RootCause   string  `json:"root_cause"`  // 根因
	Similarity  float64 `json:"similarity"`  // 相似度
	Resolution  string  `json:"resolution"`  // 解决方案
}

// Factor 贡献因素
type Factor struct {
	Name        string `json:"name"`        // 因素名称
	Description string `json:"description"` // 描述
	Impact      string `json:"impact"`      // 影响: "high", "medium", "low"
	Evidence    string `json:"evidence"`    // 证据
}

// Recommendation 修复建议
type Recommendation struct {
	Action      string   `json:"action"`      // 建议操作
	Priority    string   `json:"priority"`    // 优先级: "high", "medium", "low"
	Description string   `json:"description"` // 详细描述
	Commands    []string `json:"commands"`    // kubectl 命令
	Impact      string   `json:"impact"`      // 影响评估
	RiskLevel   string   `json:"risk_level"`  // 风险等级
}

// ChainConfig 根因分析 Chain 配置
type ChainConfig struct {
	// LLM 配置
	Temperature float64 `json:"temperature"` // 温度参数
	MaxTokens   int     `json:"max_tokens"`  // 最大 token 数

	// 分析配置
	MinConfidence   float64 `json:"min_confidence"`    // 最小置信度阈值
	IncludeSimilar  bool    `json:"include_similar"`   // 是否包含相似案例
	MaxSimilarCases int     `json:"max_similar_cases"` // 最多相似案例数

	// Prompt 配置
	PromptTemplate string `json:"prompt_template"` // Prompt 模板路径
	SystemPrompt   string `json:"system_prompt"`   // 系统提示

	// 超时配置
	Timeout time.Duration `json:"timeout"` // 超时时间
}
