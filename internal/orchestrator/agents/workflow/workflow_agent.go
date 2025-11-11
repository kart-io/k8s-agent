package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/kart-io/k8s-agent/internal/orchestrator/types"
	"github.com/kart-io/k8s-agent/internal/orchestrator/workflow"
	agentcore "github.com/kart-io/k8s-agent/pkg/agent/core"
	"github.com/kart-io/logger/core"
)

// WorkflowAgent 工作流执行 Agent
// 负责编排和执行完整的工作流
type WorkflowAgent struct {
	*agentcore.BaseAgent
	executor *workflow.Executor
	logger   core.Logger
}

// NewWorkflowAgent 创建工作流 Agent
func NewWorkflowAgent(executor *workflow.Executor, logger core.Logger) *WorkflowAgent {
	return &WorkflowAgent{
		BaseAgent: agentcore.NewBaseAgent(
			"workflow-agent",
			"Orchestrates and executes multi-step diagnostic workflows",
			[]string{
				"workflow_execution",
				"step_orchestration",
				"context_management",
				"error_recovery",
			},
		),
		executor: executor,
		logger:   logger.With("agent", "workflow"),
	}
}

// Execute 执行工作流
func (a *WorkflowAgent) Execute(ctx context.Context, input *agentcore.AgentInput) (*agentcore.AgentOutput, error) {
	start := time.Now()

	// 解析输入
	execution, ok := input.Context["execution"].(*types.WorkflowExecution)
	if !ok {
		return nil, fmt.Errorf("invalid input: missing workflow execution")
	}

	steps, ok := input.Context["steps"].([]types.WorkflowStep)
	if !ok {
		return nil, fmt.Errorf("invalid input: missing workflow steps")
	}

	// 应用超时
	if input.Options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, input.Options.Timeout)
		defer cancel()
	}

	// 初始化输出
	output := &agentcore.AgentOutput{
		Status:         "success",
		ReasoningSteps: make([]agentcore.ReasoningStep, 0),
		ToolCalls:      make([]agentcore.ToolCall, 0),
		Timestamp:      start,
	}

	a.logger.Info("Starting workflow execution",
		"execution_id", execution.ID,
		"workflow_id", execution.WorkflowID,
		"steps", len(steps))

	// 执行工作流步骤
	for i, step := range steps {
		stepStart := time.Now()

		a.logger.Info("Executing workflow step",
			"step", i+1,
			"step_id", step.ID,
			"step_type", step.Type)

		// 创建步骤 Agent
		stepAgent := NewStepAgent(a.executor, a.logger)

		// 执行步骤
		stepInput := &agentcore.AgentInput{
			Task:        fmt.Sprintf("Execute %s step", step.Type),
			Instruction: step.ID,
			Context: map[string]interface{}{
				"execution": execution,
				"step":      step,
			},
			Options: input.Options,
		}

		stepOutput, err := stepAgent.Execute(ctx, stepInput)
		if err != nil {
			// 记录失败步骤
			output.ReasoningSteps = append(output.ReasoningSteps, agentcore.ReasoningStep{
				Step:        i + 1,
				Action:      fmt.Sprintf("execute_%s_step", step.Type),
				Description: fmt.Sprintf("Execute %s step: %s", step.Type, step.ID),
				Duration:    time.Since(stepStart),
				Success:     false,
				Error:       err.Error(),
			})

			output.Status = "failed"
			output.Message = fmt.Sprintf("Workflow failed at step %d: %v", i+1, err)
			output.Latency = time.Since(start)

			return output, fmt.Errorf("step %d failed: %w", i+1, err)
		}

		// 记录成功步骤
		output.ReasoningSteps = append(output.ReasoningSteps, agentcore.ReasoningStep{
			Step:        i + 1,
			Action:      fmt.Sprintf("execute_%s_step", step.Type),
			Description: fmt.Sprintf("Execute %s step: %s", step.Type, step.ID),
			Result:      fmt.Sprintf("Step completed: %v", stepOutput.Result),
			Duration:    time.Since(stepStart),
			Success:     true,
		})

		// 合并工具调用记录
		output.ToolCalls = append(output.ToolCalls, stepOutput.ToolCalls...)

		// 更新执行上下文
		if stepResult, ok := stepOutput.Result.(map[string]interface{}); ok {
			for k, v := range stepResult {
				execution.Context[fmt.Sprintf("step_%d_%s", i+1, k)] = v
			}
		}
	}

	output.Latency = time.Since(start)
	output.Message = fmt.Sprintf("Workflow completed successfully with %d steps", len(steps))
	output.Result = map[string]interface{}{
		"execution_id": execution.ID,
		"steps_count":  len(steps),
		"context":      execution.Context,
	}

	a.logger.Info("Workflow execution completed",
		"execution_id", execution.ID,
		"duration", output.Latency,
		"steps", len(steps))

	return output, nil
}
