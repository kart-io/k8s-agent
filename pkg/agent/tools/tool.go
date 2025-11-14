package tools

import (
	"context"

	agentcore "github.com/kart-io/k8s-agent/pkg/agent/core"
	"github.com/kart-io/k8s-agent/pkg/agent/interfaces"
)

// Tool 定义工具接口
//
// Deprecated: The Tool interface is defined in interfaces.Tool.
// While this package retains the Tool interface for backward compatibility,
// new code should use interfaces.Tool directly.
//
// Migration: import "github.com/kart-io/k8s-agent/pkg/agent/interfaces"
// See: pkg/agent/docs/refactoring/migration-guide.md
//
// Note: This interface extends agentcore.Runnable which is aliased to interfaces.Runnable.
// The implementation is maintained in this package for tool-specific functionality.
type Tool interface {
	// 继承 Runnable 接口，工具是可执行的组件
	agentcore.Runnable[*ToolInput, *ToolOutput]

	// Tool 特有方法
	// Name 返回工具名称
	Name() string

	// Description 返回工具描述
	Description() string

	// ArgsSchema 返回参数的 JSON Schema
	// 用于 LLM 理解如何调用此工具
	ArgsSchema() string
}

// ToolInput 工具输入
//
// Deprecated: Use interfaces.ToolInput instead for new code.
// This type alias provides backward compatibility. It will be removed in v1.0.0.
//
// Note: This struct is retained for compatibility with existing tool implementations
// that reference tools.ToolInput. The canonical definition is now in interfaces.ToolInput.
type ToolInput = interfaces.ToolInput

// ToolOutput 工具输出
//
// Deprecated: Use interfaces.ToolOutput instead for new code.
// This type alias provides backward compatibility. It will be removed in v1.0.0.
//
// Note: This struct is retained for compatibility with existing tool implementations
// that reference tools.ToolOutput. The canonical definition is now in interfaces.ToolOutput.
type ToolOutput = interfaces.ToolOutput

// BaseTool 提供 Tool 的基础实现
//
// 实现了 Tool 接口的通用功能
// 具体的执行逻辑通过 RunFunc 函数提供
type BaseTool struct {
	*agentcore.BaseRunnable[*ToolInput, *ToolOutput]
	name        string
	description string
	argsSchema  string
	runFunc     func(context.Context, *ToolInput) (*ToolOutput, error)
}

// NewBaseTool 创建基础工具
func NewBaseTool(
	name string,
	description string,
	argsSchema string,
	runFunc func(context.Context, *ToolInput) (*ToolOutput, error),
) *BaseTool {
	return &BaseTool{
		BaseRunnable: agentcore.NewBaseRunnable[*ToolInput, *ToolOutput](),
		name:         name,
		description:  description,
		argsSchema:   argsSchema,
		runFunc:      runFunc,
	}
}

// Name 返回工具名称
func (t *BaseTool) Name() string {
	return t.name
}

// Description 返回工具描述
func (t *BaseTool) Description() string {
	return t.description
}

// ArgsSchema 返回参数 JSON Schema
func (t *BaseTool) ArgsSchema() string {
	return t.argsSchema
}

// Invoke 执行工具
func (t *BaseTool) Invoke(ctx context.Context, input *ToolInput) (*ToolOutput, error) {
	// 触发回调
	config := t.GetConfig()
	for _, cb := range config.Callbacks {
		if err := cb.OnToolStart(ctx, t.name, input.Args); err != nil {
			return nil, err
		}
	}

	// 执行工具函数
	output, err := t.runFunc(ctx, input)

	// 触发回调
	for _, cb := range config.Callbacks {
		if err != nil {
			_ = cb.OnToolError(ctx, t.name, err)
		} else {
			_ = cb.OnToolEnd(ctx, t.name, output.Result)
		}
	}

	return output, err
}

// Stream 流式执行工具（默认实现：包装成单个块）
func (t *BaseTool) Stream(ctx context.Context, input *ToolInput) (<-chan agentcore.StreamChunk[*ToolOutput], error) {
	outChan := make(chan agentcore.StreamChunk[*ToolOutput], 1)

	go func() {
		defer close(outChan)

		output, err := t.Invoke(ctx, input)
		outChan <- agentcore.StreamChunk[*ToolOutput]{
			Data:  output,
			Error: err,
			Done:  true,
		}
	}()

	return outChan, nil
}

// Batch 批量执行工具
func (t *BaseTool) Batch(ctx context.Context, inputs []*ToolInput) ([]*ToolOutput, error) {
	return t.BaseRunnable.Batch(ctx, inputs, t.Invoke)
}

// Pipe 连接到另一个 Runnable
func (t *BaseTool) Pipe(next agentcore.Runnable[*ToolOutput, any]) agentcore.Runnable[*ToolInput, any] {
	return agentcore.NewRunnablePipe[*ToolInput, *ToolOutput, any](t, next)
}

// WithCallbacks 添加回调（重写以返回正确的类型）
func (t *BaseTool) WithCallbacks(callbacks ...agentcore.Callback) agentcore.Runnable[*ToolInput, *ToolOutput] {
	newTool := *t
	newTool.BaseRunnable = t.BaseRunnable.WithCallbacks(callbacks...)
	return &newTool
}

// WithConfig 配置工具（重写以返回正确的类型）
func (t *BaseTool) WithConfig(config agentcore.RunnableConfig) agentcore.Runnable[*ToolInput, *ToolOutput] {
	newTool := *t
	newTool.BaseRunnable = t.BaseRunnable.WithConfig(config)
	return &newTool
}

// ToolError 工具执行错误
type ToolError struct {
	ToolName string
	Message  string
	Err      error
}

// Error 实现 error 接口
func (e *ToolError) Error() string {
	if e.Err != nil {
		return e.ToolName + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.ToolName + ": " + e.Message
}

// Unwrap 支持 errors.Unwrap
func (e *ToolError) Unwrap() error {
	return e.Err
}

// NewToolError 创建工具错误
func NewToolError(toolName, message string, err error) *ToolError {
	return &ToolError{
		ToolName: toolName,
		Message:  message,
		Err:      err,
	}
}
