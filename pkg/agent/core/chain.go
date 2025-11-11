package core

import (
	"context"
	"time"
)

// Chain 定义链式处理接口
//
// Chain 是一种串行执行的处理模式，适用于：
// - 多步骤的数据处理流程
// - 需要按顺序执行的分析任务
// - 每个步骤依赖前一步骤的输出
type Chain interface {
	// Process 执行链式处理
	Process(ctx context.Context, input *ChainInput) (*ChainOutput, error)

	// Name 返回 Chain 的名称
	Name() string

	// Steps 返回包含的步骤数量
	Steps() int
}

// ChainInput Chain 输入
type ChainInput struct {
	// 输入数据
	Data interface{}            `json:"data"` // 输入数据
	Vars map[string]interface{} `json:"vars"` // 变量集合
	Tags []string               `json:"tags"` // 标签

	// 执行选项
	Options ChainOptions `json:"options"` // 执行选项

	// 元数据
	Timestamp time.Time `json:"timestamp"` // 时间戳
}

// ChainOutput Chain 输出
type ChainOutput struct {
	// 输出数据
	Data   interface{}            `json:"data"`   // 输出数据
	Result map[string]interface{} `json:"result"` // 结果集合

	// 执行信息
	StepsExecuted []StepExecution `json:"steps_executed"` // 执行的步骤
	TotalLatency  time.Duration   `json:"total_latency"`  // 总延迟
	Status        string          `json:"status"`         // 状态: "success", "failed", "partial"

	// 元数据
	Timestamp time.Time              `json:"timestamp"` // 时间戳
	Metadata  map[string]interface{} `json:"metadata"`  // 额外元数据
}

// ChainOptions Chain 执行选项
type ChainOptions struct {
	// 执行控制
	StopOnError bool          `json:"stop_on_error,omitempty"` // 出错时是否停止
	Timeout     time.Duration `json:"timeout,omitempty"`       // 超时时间
	Parallel    bool          `json:"parallel,omitempty"`      // 是否并行执行（如果可能）

	// 步骤控制
	SkipSteps []int `json:"skip_steps,omitempty"` // 跳过的步骤编号
	OnlySteps []int `json:"only_steps,omitempty"` // 仅执行的步骤编号

	// 额外选项
	Extra map[string]interface{} `json:"extra,omitempty"` // 额外选项
}

// StepExecution 步骤执行记录
type StepExecution struct {
	StepNumber  int           `json:"step_number"` // 步骤编号
	StepName    string        `json:"step_name"`   // 步骤名称
	Description string        `json:"description"` // 步骤描述
	Input       interface{}   `json:"input"`       // 输入
	Output      interface{}   `json:"output"`      // 输出
	Duration    time.Duration `json:"duration"`    // 耗时
	Success     bool          `json:"success"`     // 是否成功
	Error       string        `json:"error"`       // 错误信息
	Skipped     bool          `json:"skipped"`     // 是否跳过
}

// Step 定义 Chain 中的单个步骤接口
type Step interface {
	// Execute 执行步骤
	Execute(ctx context.Context, input interface{}) (interface{}, error)

	// Name 返回步骤名称
	Name() string

	// Description 返回步骤描述
	Description() string
}

// BaseChain 提供 Chain 的基础实现
type BaseChain struct {
	name  string
	steps []Step
}

// NewBaseChain 创建基础 Chain
func NewBaseChain(name string, steps []Step) *BaseChain {
	return &BaseChain{
		name:  name,
		steps: steps,
	}
}

// Name 返回 Chain 名称
func (c *BaseChain) Name() string {
	return c.name
}

// Steps 返回步骤数量
func (c *BaseChain) Steps() int {
	return len(c.steps)
}

// Process 执行链式处理
func (c *BaseChain) Process(ctx context.Context, input *ChainInput) (*ChainOutput, error) {
	start := time.Now()

	output := &ChainOutput{
		StepsExecuted: make([]StepExecution, 0),
		Status:        "success",
		Timestamp:     start,
		Metadata:      make(map[string]interface{}),
	}

	// 应用超时
	if input.Options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, input.Options.Timeout)
		defer cancel()
	}

	// 执行步骤
	currentData := input.Data
	for i, step := range c.steps {
		// 检查是否跳过
		if shouldSkipStep(i+1, input.Options) {
			output.StepsExecuted = append(output.StepsExecuted, StepExecution{
				StepNumber: i + 1,
				StepName:   step.Name(),
				Skipped:    true,
			})
			continue
		}

		// 执行步骤
		stepStart := time.Now()
		result, err := step.Execute(ctx, currentData)
		duration := time.Since(stepStart)

		execution := StepExecution{
			StepNumber:  i + 1,
			StepName:    step.Name(),
			Description: step.Description(),
			Input:       currentData,
			Output:      result,
			Duration:    duration,
			Success:     err == nil,
		}

		if err != nil {
			execution.Error = err.Error()
			output.StepsExecuted = append(output.StepsExecuted, execution)

			if input.Options.StopOnError {
				output.Status = "failed"
				output.TotalLatency = time.Since(start)
				return output, err
			}

			output.Status = "partial"
		} else {
			currentData = result
		}

		output.StepsExecuted = append(output.StepsExecuted, execution)
	}

	output.Data = currentData
	output.TotalLatency = time.Since(start)

	return output, nil
}

// shouldSkipStep 检查是否应该跳过步骤
func shouldSkipStep(stepNum int, options ChainOptions) bool {
	// 如果指定了 OnlySteps，只执行这些步骤
	if len(options.OnlySteps) > 0 {
		for _, only := range options.OnlySteps {
			if only == stepNum {
				return false
			}
		}
		return true
	}

	// 检查 SkipSteps
	for _, skip := range options.SkipSteps {
		if skip == stepNum {
			return true
		}
	}

	return false
}

// DefaultChainOptions 返回默认的 Chain 选项
func DefaultChainOptions() ChainOptions {
	return ChainOptions{
		StopOnError: true,
		Timeout:     60 * time.Second,
		Parallel:    false,
		Extra:       make(map[string]interface{}),
	}
}
