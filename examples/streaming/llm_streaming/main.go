package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kart-io/k8s-agent/pkg/agent/core"
	"github.com/kart-io/k8s-agent/pkg/agent/llm"
	"github.com/kart-io/k8s-agent/pkg/agent/stream/agents"
)

// MockLLMClient 模拟 LLM 客户端（用于演示）
type MockLLMClient struct{}

func (m *MockLLMClient) Complete(ctx context.Context, req *llm.CompletionRequest) (*llm.CompletionResponse, error) {
	return &llm.CompletionResponse{
		Content: "This is a simulated LLM response that will be streamed character by character to demonstrate the streaming capabilities of the agent framework.",
		Model:   "mock-model",
	}, nil
}

func (m *MockLLMClient) Chat(ctx context.Context, messages []llm.Message) (*llm.CompletionResponse, error) {
	return m.Complete(ctx, &llm.CompletionRequest{Messages: messages})
}

func (m *MockLLMClient) Provider() llm.Provider {
	return llm.ProviderCustom
}

func (m *MockLLMClient) IsAvailable() bool {
	return true
}

func main() {
	fmt.Println("=== LLM Streaming Example ===")
	fmt.Println()

	// 创建 LLM 流式 Agent
	llmClient := &MockLLMClient{}
	config := &agents.StreamingLLMConfig{
		ChunkSize:        5,                      // 每次发送 5 个字符
		ChunkDelay:       100 * time.Millisecond, // 模拟打字效果
		EnableProgress:   true,
		ProgressInterval: time.Second,
	}

	agent := agents.NewStreamingLLMAgent(llmClient, config)

	// 准备输入
	input := &core.AgentInput{
		Task:        "Explain the benefits of streaming in AI applications",
		Instruction: "You are a helpful AI assistant.",
		Options:     core.DefaultAgentOptions(),
		Timestamp:   time.Now(),
	}

	ctx := context.Background()

	// 执行流式任务
	fmt.Println("Starting streaming execution...")
	streamOutput, err := agent.ExecuteStream(ctx, input)
	if err != nil {
		log.Fatalf("Failed to execute stream: %v", err)
	}
	defer streamOutput.Close()

	// 实时读取和显示流数据
	fmt.Println("\nStreaming output:")
	fmt.Println("----------------------------------------")

	var fullText string
	chunkCount := 0

	for {
		chunk, err := streamOutput.Next()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("Error reading chunk: %v", err)
			break
		}

		chunkCount++

		switch chunk.Type {
		case core.ChunkTypeText:
			// 实时显示文本（模拟打字效果）
			fmt.Print(chunk.Text)
			fullText += chunk.Text

		case core.ChunkTypeProgress:
			// 显示进度
			fmt.Printf("\n[Progress: %.1f%%]", chunk.Metadata.Progress)

		case core.ChunkTypeStatus:
			// 显示状态更新
			fmt.Printf("\n[Status: %s]", chunk.Data)

		case core.ChunkTypeError:
			// 显示错误
			fmt.Printf("\n[Error: %v]", chunk.Error)
		}
	}

	fmt.Println("\n----------------------------------------")
	fmt.Printf("\nStreaming completed!\n")
	fmt.Printf("Total chunks received: %d\n", chunkCount)
	fmt.Printf("Full text length: %d characters\n", len(fullText))

	// 演示同步执行（收集所有流数据）
	fmt.Println("\n=== Synchronous Execution (Collecting All Data) ===")
	fmt.Println()

	output, err := agent.Execute(ctx, input)
	if err != nil {
		log.Fatalf("Failed to execute: %v", err)
	}

	fmt.Printf("Result: %s\n", output.Result)
	fmt.Printf("Status: %s\n", output.Status)
	fmt.Printf("Latency: %v\n", output.Latency)

	// 演示流消费者
	fmt.Println("\n=== Stream Consumer Example ===")
	fmt.Println()

	consumer := &agents.SimpleStreamConsumer{
		OnStartFunc: func() error {
			fmt.Println("Consumer: Stream started")
			return nil
		},
		OnChunkFunc: func(chunk *core.LegacyStreamChunk) error {
			if chunk.Type == core.ChunkTypeText {
				fmt.Printf("Consumer received: %s\n", chunk.Text)
			}
			return nil
		},
		OnCompleteFunc: func() error {
			fmt.Println("Consumer: Stream completed")
			return nil
		},
		OnErrorFunc: func(err error) error {
			fmt.Printf("Consumer: Error occurred: %v\n", err)
			return err
		},
	}

	// 使用文本累积消费者
	accumulator := agents.NewTextAccumulatorConsumer()

	streamOutput2, err := agent.ExecuteStream(ctx, input)
	if err != nil {
		log.Fatalf("Failed to execute stream: %v", err)
	}
	defer streamOutput2.Close()

	accumulator.OnStart()

	for {
		chunk, err := streamOutput2.Next()
		if err != nil {
			break
		}

		consumer.OnChunk(chunk)
		accumulator.OnChunk(chunk)
	}

	consumer.OnComplete()
	accumulator.OnComplete()

	fmt.Printf("\nAccumulated text: %s\n", accumulator.Text())

	fmt.Println("\n=== Example Completed ===")
}
