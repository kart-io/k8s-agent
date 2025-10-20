package proxy

import "time"

// CompletionRequest 补全请求结构
type CompletionRequest struct {
	Messages    []Message // 消息列表
	Temperature float64   // 温度参数 (0.0-1.0)
	MaxTokens   int       // 最大 token 数
	StopWords   []string  // 停止词列表
}

// Message 消息结构
type Message struct {
	Role    string // 角色: "system", "user", "assistant"
	Content string // 消息内容
}

// CompletionResponse 补全响应结构
type CompletionResponse struct {
	Content      string        // 生成的内容
	Provider     string        // 实际使用的提供商
	Model        string        // 实际使用的模型
	TokensUsed   int           // 使用的 token 数
	Cost         float64       // 调用成本
	Latency      time.Duration // 调用延迟
	FinishReason string        // 完成原因
}

// UsageMetrics 使用指标
type UsageMetrics struct {
	TotalRequests   int                        // 总请求数
	SuccessfulCalls int                        // 成功调用数
	FailedCalls     int                        // 失败调用数
	TotalCost       float64                    // 总成本
	ProviderStats   map[string]ProviderMetrics // 各提供商的统计数据
}

// ProviderMetrics 提供商指标
type ProviderMetrics struct {
	Calls      int           // 调用次数
	Successes  int           // 成功次数
	Failures   int           // 失败次数
	TotalCost  float64       // 总成本
	AvgLatency time.Duration // 平均延迟
}

// ProviderStatus 提供商状态
type ProviderStatus struct {
	Name      string    // 提供商名称
	Healthy   bool      // 健康状态
	LastError string    // 最后一次错误信息
	LastCheck time.Time // 最后一次检查时间
}
