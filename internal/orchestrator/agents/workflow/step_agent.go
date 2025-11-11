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

// StepAgent 工作流步骤执行 Agent
// 负责执行单个工作流步骤
type StepAgent struct {
	*agentcore.BaseAgent
	executor *workflow.Executor
	logger   core.Logger
}

// NewStepAgent 创建步骤 Agent
func NewStepAgent(executor *workflow.Executor, logger core.Logger) *StepAgent {
	return &StepAgent{
		BaseAgent: agentcore.NewBaseAgent(
			"step-agent",
			"Executes individual workflow steps (Command/AI/Decision/Remediation/Notification/Wait)",
			[]string{
				"command_execution",
				"ai_analysis",
				"decision_making",
				"remediation",
				"notification",
				"timing_control",
			},
		),
		executor: executor,
		logger:   logger.With("agent", "step"),
	}
}

// Execute 执行工作流步骤
func (a *StepAgent) Execute(ctx context.Context, input *agentcore.AgentInput) (*agentcore.AgentOutput, error) {
	start := time.Now()

	// 解析输入
	execution, ok := input.Context["execution"].(*types.WorkflowExecution)
	if !ok {
		return nil, fmt.Errorf("invalid input: missing execution")
	}

	step, ok := input.Context["step"].(types.WorkflowStep)
	if !ok {
		return nil, fmt.Errorf("invalid input: missing step")
	}

	// 初始化输出
	output := &agentcore.AgentOutput{
		Status:    "success",
		ToolCalls: make([]agentcore.ToolCall, 0),
		Timestamp: start,
	}

	a.logger.Info("Executing step",
		"step_id", step.ID,
		"step_type", step.Type)

	// 根据步骤类型执行
	var result map[string]interface{}
	var err error
	var toolName string

	switch step.Type {
	case "command":
		toolName = "kubectl_command"
		result, err = a.executeCommand(ctx, execution, step)
	case "ai":
		toolName = "ai_analysis"
		result, err = a.executeAIAnalysis(ctx, execution, step)
	case "decision":
		toolName = "decision_evaluator"
		result, err = a.executeDecision(ctx, execution, step)
	case "remediation":
		toolName = "remediation_executor"
		result, err = a.executeRemediation(ctx, execution, step)
	case "notification":
		toolName = "notification_sender"
		result, err = a.executeNotification(ctx, execution, step)
	case "wait":
		toolName = "timer"
		result, err = a.executeWait(ctx, execution, step)
	default:
		return nil, fmt.Errorf("unknown step type: %s", step.Type)
	}

	// 记录工具调用
	toolCall := agentcore.ToolCall{
		ToolName: toolName,
		Input: map[string]interface{}{
			"step_id":   step.ID,
			"step_type": step.Type,
			"config":    step.Config,
		},
		Duration: time.Since(start),
		Success:  err == nil,
	}

	if err != nil {
		toolCall.Error = err.Error()
		output.Status = "failed"
		output.Message = fmt.Sprintf("Step execution failed: %v", err)
	} else {
		toolCall.Output = result
		output.Result = result
		output.Message = fmt.Sprintf("Step %s executed successfully", step.Type)
	}

	output.ToolCalls = append(output.ToolCalls, toolCall)
	output.Latency = time.Since(start)

	return output, err
}

// executeCommand 执行命令步骤
func (a *StepAgent) executeCommand(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep) (map[string]interface{}, error) {
	a.logger.Info("Executing command step", "step_id", step.ID)
	return a.executor.ExecuteCommand(ctx, execution, step)
}

// executeAIAnalysis 执行 AI 分析步骤
func (a *StepAgent) executeAIAnalysis(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep) (map[string]interface{}, error) {
	a.logger.Info("Executing AI analysis step", "step_id", step.ID)
	return a.executor.ExecuteAIAnalysis(ctx, execution, step)
}

// executeDecision 执行决策步��
func (a *StepAgent) executeDecision(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep) (map[string]interface{}, error) {
	a.logger.Info("Executing decision step", "step_id", step.ID)
	return a.executor.ExecuteDecision(ctx, execution, step)
}

// executeRemediation 执行修复步骤
func (a *StepAgent) executeRemediation(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep) (map[string]interface{}, error) {
	a.logger.Info("Executing remediation step", "step_id", step.ID)
	return a.executor.ExecuteRemediation(ctx, execution, step)
}

// executeNotification 执行通知步骤
func (a *StepAgent) executeNotification(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep) (map[string]interface{}, error) {
	a.logger.Info("Executing notification step", "step_id", step.ID)
	return a.executor.ExecuteNotification(ctx, execution, step)
}

// executeWait 执行等待步骤
func (a *StepAgent) executeWait(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep) (map[string]interface{}, error) {
	a.logger.Info("Executing wait step", "step_id", step.ID)
	return a.executor.ExecuteWait(ctx, execution, step)
}
