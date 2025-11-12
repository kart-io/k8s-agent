package core

import (
	"context"
	"fmt"
	"time"
)

// ExampleAgent 示例 Agent 实现
//
// 展示如何实现一个完整的 Agent，包括：
// - 实现 Runnable 接口的核心方法 Invoke
// - 使用回调系统
// - 支持流式输出、批量处理、管道连接
type ExampleAgent struct {
	*BaseAgent
	processor func(context.Context, *AgentInput) (*AgentOutput, error)
}

// NewExampleAgent 创建示例 Agent
func NewExampleAgent(name, description string, capabilities []string, processor func(context.Context, *AgentInput) (*AgentOutput, error)) *ExampleAgent {
	return &ExampleAgent{
		BaseAgent: NewBaseAgent(name, description, capabilities),
		processor: processor,
	}
}

// Invoke 执行 Agent 逻辑
//
// 这是 Runnable 接口的核心方法，实现实际的业务逻辑
func (e *ExampleAgent) Invoke(ctx context.Context, input *AgentInput) (*AgentOutput, error) {
	startTime := time.Now()

	// 触发开始回调
	if err := e.triggerOnStart(ctx, input); err != nil {
		return nil, err
	}

	// 执行实际处理逻辑
	output, err := e.processor(ctx, input)
	if err != nil {
		// 触发错误回调
		_ = e.triggerOnError(ctx, err)
		return nil, err
	}

	// 设置延迟
	output.Latency = time.Since(startTime)
	output.Timestamp = time.Now()

	// 触发结束回调
	if err := e.triggerOnEnd(ctx, output); err != nil {
		return output, err
	}

	// 触发完成回调
	if err := e.triggerOnFinish(ctx, output); err != nil {
		return output, err
	}

	return output, nil
}

// StreamingExampleAgent 支持流式输出的 Agent
//
// 展示如何实现流式 Agent
type StreamingExampleAgent struct {
	*BaseAgent
	streamProcessor func(context.Context, *AgentInput, chan<- StreamChunk[*AgentOutput]) error
}

// NewStreamingExampleAgent 创建流式示例 Agent
func NewStreamingExampleAgent(
	name, description string,
	capabilities []string,
	processor func(context.Context, *AgentInput, chan<- StreamChunk[*AgentOutput]) error,
) *StreamingExampleAgent {
	return &StreamingExampleAgent{
		BaseAgent:       NewBaseAgent(name, description, capabilities),
		streamProcessor: processor,
	}
}

// Invoke 同步执行（收集所有流块）
func (s *StreamingExampleAgent) Invoke(ctx context.Context, input *AgentInput) (*AgentOutput, error) {
	// 通过 Stream 实现 Invoke
	stream, err := s.Stream(ctx, input)
	if err != nil {
		return nil, err
	}

	var lastOutput *AgentOutput
	for chunk := range stream {
		if chunk.Error != nil {
			return lastOutput, chunk.Error
		}
		lastOutput = chunk.Data
	}

	return lastOutput, nil
}

// Stream 流式执行
func (s *StreamingExampleAgent) Stream(ctx context.Context, input *AgentInput) (<-chan StreamChunk[*AgentOutput], error) {
	outChan := make(chan StreamChunk[*AgentOutput])

	go func() {
		defer close(outChan)

		// 触发开始回调
		if err := s.triggerOnStart(ctx, input); err != nil {
			outChan <- StreamChunk[*AgentOutput]{Error: err}
			return
		}

		// 执行流式处理
		if err := s.streamProcessor(ctx, input, outChan); err != nil {
			_ = s.triggerOnError(ctx, err)
			outChan <- StreamChunk[*AgentOutput]{Error: err}
			return
		}
	}()

	return outChan, nil
}

// SimpleTaskAgent 简单任务 Agent
//
// 一个实用的 Agent 实现，用于处理简单任务
type SimpleTaskAgent struct {
	*BaseAgent
}

// NewSimpleTaskAgent 创建简单任务 Agent
func NewSimpleTaskAgent() *SimpleTaskAgent {
	return &SimpleTaskAgent{
		BaseAgent: NewBaseAgent(
			"SimpleTaskAgent",
			"A simple agent that processes tasks and returns results",
			[]string{"task-processing", "analysis"},
		),
	}
}

// Invoke 执行任务
func (s *SimpleTaskAgent) Invoke(ctx context.Context, input *AgentInput) (*AgentOutput, error) {
	startTime := time.Now()

	// 简单的任务处理逻辑
	result := fmt.Sprintf("Processed task: %s with instruction: %s", input.Task, input.Instruction)

	output := &AgentOutput{
		Result:  result,
		Status:  "success",
		Message: "Task completed successfully",
		ReasoningSteps: []ReasoningStep{
			{
				Step:        1,
				Action:      "parse_input",
				Description: "Parsing input task and instruction",
				Result:      "Input parsed successfully",
				Duration:    10 * time.Millisecond,
				Success:     true,
			},
			{
				Step:        2,
				Action:      "process_task",
				Description: "Processing the task",
				Result:      result,
				Duration:    50 * time.Millisecond,
				Success:     true,
			},
		},
		Latency:   time.Since(startTime),
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"agent_name": s.Name(),
			"task_type":  "simple",
		},
	}

	return output, nil
}

// ExampleUsage 展示 Agent 的使用方式
func ExampleUsage() {
	// 1. 创建一个简单的 Agent
	simpleAgent := NewSimpleTaskAgent()

	// 2. 准备输入
	input := &AgentInput{
		Task:        "Analyze system logs",
		Instruction: "Find error patterns",
		Context: map[string]interface{}{
			"log_level": "error",
			"timeframe": "last_hour",
		},
		Options:   DefaultAgentOptions(),
		SessionID: "session-123",
		Timestamp: time.Now(),
	}

	// 3. 使用 Invoke 执行
	ctx := context.Background()
	output, err := simpleAgent.Invoke(ctx, input)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Status: %s\n", output.Status)
	fmt.Printf("Result: %v\n", output.Result)
	fmt.Printf("Latency: %v\n", output.Latency)

	// 4. 使用回调
	loggingCallback := NewStdoutCallback(true)
	agentWithCallbacks := simpleAgent.WithCallbacks(loggingCallback)

	_, err = agentWithCallbacks.Invoke(ctx, input)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// 5. 批量执行
	inputs := []*AgentInput{
		{Task: "Task 1", Options: DefaultAgentOptions()},
		{Task: "Task 2", Options: DefaultAgentOptions()},
		{Task: "Task 3", Options: DefaultAgentOptions()},
	}

	outputs, err := simpleAgent.Batch(ctx, inputs)
	if err != nil {
		fmt.Printf("Batch error: %v\n", err)
		return
	}

	fmt.Printf("Processed %d tasks\n", len(outputs))

	// 6. 使用 ChainableAgent 串联多个 Agent
	agent1 := NewSimpleTaskAgent()
	agent2 := NewSimpleTaskAgent()

	chainAgent := NewChainableAgent(
		"ChainedAgent",
		"A chain of multiple agents",
		agent1,
		agent2,
	)

	output, err = chainAgent.Invoke(ctx, input)
	if err != nil {
		fmt.Printf("Chain error: %v\n", err)
		return
	}

	fmt.Printf("Chain result: %v\n", output.Result)

	// 7. 使用 AgentExecutor 添加重试和超时
	executor := NewAgentExecutor(
		simpleAgent,
		WithMaxRetries(3),
		WithTimeout(30*time.Second),
	)

	output, err = executor.Execute(ctx, input)
	if err != nil {
		fmt.Printf("Executor error: %v\n", err)
		return
	}

	fmt.Printf("Executor result: %v\n", output.Result)
}
