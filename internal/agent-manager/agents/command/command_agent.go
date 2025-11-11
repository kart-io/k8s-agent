package command

import (
	"context"
	"fmt"
	"time"

	"github.com/kart-io/k8s-agent/internal/agent-manager/command"
	agentcore "github.com/kart-io/k8s-agent/pkg/agent/core"
	"github.com/kart-io/k8s-agent/pkg/types"
	"github.com/kart-io/logger/core"
)

// CommandAgent 命令执行 Agent
// 负责将命令分发到目标集群并追踪执行结果
type CommandAgent struct {
	*agentcore.BaseAgent
	dispatcher *command.Dispatcher
	logger     core.Logger
}

// NewCommandAgent 创建命令 Agent
func NewCommandAgent(dispatcher *command.Dispatcher, logger core.Logger) *CommandAgent {
	return &CommandAgent{
		BaseAgent: agentcore.NewBaseAgent(
			"command-agent",
			"Dispatches and tracks command execution on target clusters",
			[]string{
				"command_dispatch",
				"execution_tracking",
				"result_polling",
				"timeout_management",
			},
		),
		dispatcher: dispatcher,
		logger:     logger.With("agent", "command"),
	}
}

// Execute 执行命令分发和追踪
func (a *CommandAgent) Execute(ctx context.Context, input *agentcore.AgentInput) (*agentcore.AgentOutput, error) {
	start := time.Now()

	// 解析命令
	cmd, ok := input.Context["command"].(*types.Command)
	if !ok {
		return nil, fmt.Errorf("invalid input: missing command")
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

	a.logger.Info("Dispatching command",
		"command_id", cmd.ID,
		"cluster_id", cmd.ClusterID,
		"tool", cmd.Tool,
		"action", cmd.Action)

	// 步骤 1: 分发命令
	dispatchStart := time.Now()
	if err := a.dispatcher.DispatchCommand(ctx, cmd); err != nil {
		output.ReasoningSteps = append(output.ReasoningSteps, agentcore.ReasoningStep{
			Step:        1,
			Action:      "dispatch_command",
			Description: "Dispatch command to target cluster",
			Duration:    time.Since(dispatchStart),
			Success:     false,
			Error:       err.Error(),
		})

		output.Status = "failed"
		output.Message = fmt.Sprintf("Failed to dispatch command: %v", err)
		output.Latency = time.Since(start)

		return output, fmt.Errorf("command dispatch failed: %w", err)
	}

	output.ReasoningSteps = append(output.ReasoningSteps, agentcore.ReasoningStep{
		Step:        1,
		Action:      "dispatch_command",
		Description: "Dispatch command to target cluster",
		Result:      fmt.Sprintf("Command %s dispatched to cluster %s", cmd.ID, cmd.ClusterID),
		Duration:    time.Since(dispatchStart),
		Success:     true,
	})

	// 步骤 2: 等待执行结果
	waitStart := time.Now()
	result, err := a.waitForResult(ctx, cmd.ID, cmd.Timeout)
	if err != nil {
		output.ReasoningSteps = append(output.ReasoningSteps, agentcore.ReasoningStep{
			Step:        2,
			Action:      "wait_for_result",
			Description: "Wait for command execution result",
			Duration:    time.Since(waitStart),
			Success:     false,
			Error:       err.Error(),
		})

		output.Status = "failed"
		output.Message = fmt.Sprintf("Failed to get command result: %v", err)
		output.Latency = time.Since(start)

		return output, fmt.Errorf("command execution failed: %w", err)
	}

	output.ReasoningSteps = append(output.ReasoningSteps, agentcore.ReasoningStep{
		Step:        2,
		Action:      "wait_for_result",
		Description: "Wait for command execution result",
		Result:      fmt.Sprintf("Command completed with status: %s", result.Status),
		Duration:    time.Since(waitStart),
		Success:     true,
	})

	// 记录工具调用
	output.ToolCalls = append(output.ToolCalls, agentcore.ToolCall{
		ToolName: cmd.Tool,
		Input: map[string]interface{}{
			"action":     cmd.Action,
			"args":       cmd.Args,
			"cluster_id": cmd.ClusterID,
		},
		Output:   result.Output,
		Duration: time.Since(start),
		Success:  result.Status == "success",
	})

	output.Latency = time.Since(start)
	output.Result = map[string]interface{}{
		"command_id":     cmd.ID,
		"status":         result.Status,
		"output":         result.Output,
		"error":          result.Error,
		"execution_time": result.ExecutionTime,
	}

	if result.Status == "success" {
		output.Message = "Command executed successfully"
	} else {
		output.Status = "failed"
		output.Message = fmt.Sprintf("Command execution failed: %s", result.Error)
	}

	a.logger.Info("Command execution completed",
		"command_id", cmd.ID,
		"status", result.Status,
		"duration", output.Latency)

	return output, nil
}

// waitForResult 等待命令执行结果
func (a *CommandAgent) waitForResult(ctx context.Context, commandID string, timeout time.Duration) (*types.CommandResult, error) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timeout waiting for command result")
			}

			// 轮询结果
			result, err := a.dispatcher.GetCommandResult(ctx, commandID)
			if err != nil {
				// 结果尚未准备好，继续轮询
				continue
			}

			return result, nil
		}
	}
}
