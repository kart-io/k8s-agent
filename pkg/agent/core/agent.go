package core

import (
	"context"
	"time"
)

// Agent 定义通用 AI Agent 接口
//
// Agent 是一个具有推理能力的智能体，能够：
// - 接收输入并进行处理
// - 调用工具获取额外信息
// - 使用 LLM 进行推理
// - 返回结构化输出
type Agent interface {
	// Execute 执行 Agent 的主要逻辑
	Execute(ctx context.Context, input *AgentInput) (*AgentOutput, error)

	// Name 返回 Agent 的名称
	Name() string

	// Description 返回 Agent 的描述
	Description() string

	// Capabilities 返回 Agent 的能力列表
	Capabilities() []string
}

// AgentInput Agent 输入
type AgentInput struct {
	// 任务描述
	Task        string                 `json:"task"`        // 任务描述
	Instruction string                 `json:"instruction"` // 具体指令
	Context     map[string]interface{} `json:"context"`     // 上下文信息

	// 执行选项
	Options AgentOptions `json:"options"` // 执行选项

	// 元数据
	SessionID string    `json:"session_id"` // 会话 ID
	Timestamp time.Time `json:"timestamp"`  // 时间戳
}

// AgentOutput Agent 输出
type AgentOutput struct {
	// 执行结果
	Result  interface{} `json:"result"`  // 结果数据
	Status  string      `json:"status"`  // 状态: "success", "failed", "partial"
	Message string      `json:"message"` // 结果消息

	// 推理过程
	ReasoningSteps []ReasoningStep `json:"reasoning_steps"` // 推理步骤
	ToolCalls      []ToolCall      `json:"tool_calls"`      // 工具调用记录

	// 元数据
	Latency   time.Duration          `json:"latency"`   // 执行延迟
	Timestamp time.Time              `json:"timestamp"` // 时间戳
	Metadata  map[string]interface{} `json:"metadata"`  // 额外元数据
}

// AgentOptions Agent 执行选项
type AgentOptions struct {
	// LLM 配置
	Temperature float64 `json:"temperature,omitempty"` // LLM 温度参数
	MaxTokens   int     `json:"max_tokens,omitempty"`  // 最大 token 数
	Model       string  `json:"model,omitempty"`       // LLM 模型

	// 工具配置
	EnableTools  bool     `json:"enable_tools,omitempty"`   // 是否启用工具
	AllowedTools []string `json:"allowed_tools,omitempty"`  // 允许的工具列表
	MaxToolCalls int      `json:"max_tool_calls,omitempty"` // 最大工具调用次数

	// 记忆配置
	EnableMemory     bool `json:"enable_memory,omitempty"`      // 是否启用记忆
	LoadHistory      bool `json:"load_history,omitempty"`       // 是否加载历史
	SaveToMemory     bool `json:"save_to_memory,omitempty"`     // 是否保存到记忆
	MaxHistoryLength int  `json:"max_history_length,omitempty"` // 最大历史长度

	// 超时配置
	Timeout time.Duration `json:"timeout,omitempty"` // 超时时间
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

// ToolCall 工具调用记录
type ToolCall struct {
	ToolName string                 `json:"tool_name"` // 工具名称
	Input    map[string]interface{} `json:"input"`     // 输入参数
	Output   interface{}            `json:"output"`    // 输出结果
	Duration time.Duration          `json:"duration"`  // 耗时
	Success  bool                   `json:"success"`   // 是否成功
	Error    string                 `json:"error"`     // 错误信息
}

// BaseAgent 提供 Agent 的基础实现
type BaseAgent struct {
	name         string
	description  string
	capabilities []string
}

// NewBaseAgent 创建基础 Agent
func NewBaseAgent(name, description string, capabilities []string) *BaseAgent {
	return &BaseAgent{
		name:         name,
		description:  description,
		capabilities: capabilities,
	}
}

// Name 返回 Agent 名称
func (a *BaseAgent) Name() string {
	return a.name
}

// Description 返回 Agent 描述
func (a *BaseAgent) Description() string {
	return a.description
}

// Capabilities 返回 Agent 能力列表
func (a *BaseAgent) Capabilities() []string {
	return a.capabilities
}

// Execute 需要由具体 Agent 实现
func (a *BaseAgent) Execute(ctx context.Context, input *AgentInput) (*AgentOutput, error) {
	panic("Execute method must be implemented by concrete agent")
}

// DefaultAgentOptions 返回默认的 Agent 选项
func DefaultAgentOptions() AgentOptions {
	return AgentOptions{
		Temperature:      0.7,
		MaxTokens:        2000,
		EnableTools:      true,
		MaxToolCalls:     5,
		EnableMemory:     false,
		LoadHistory:      false,
		SaveToMemory:     false,
		MaxHistoryLength: 10,
		Timeout:          60 * time.Second,
	}
}
