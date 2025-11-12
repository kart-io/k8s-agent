package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kart-io/k8s-agent/pkg/agent/core"
	"github.com/kart-io/k8s-agent/pkg/agent/stream/agents"
)

//nolint:gocyclo // Example code with comprehensive demo flow
func main() {
	fmt.Println("=== Data Processing Pipeline Example ===")
	fmt.Println()

	// 创建数据管道 Agent
	config := &agents.DataPipelineConfig{
		BatchSize:        50,
		ProcessDelay:     50 * time.Millisecond,
		EnableProgress:   true,
		ProgressInterval: 500 * time.Millisecond,
		MaxWorkers:       4,
	}

	agent := agents.NewDataPipelineAgent(config)

	// 准备大数据集
	dataSize := 500
	dataSource := make([]interface{}, dataSize)
	for i := 0; i < dataSize; i++ {
		dataSource[i] = map[string]interface{}{
			"id":    i,
			"value": fmt.Sprintf("item-%d", i),
			"data":  []int{i, i * 2, i * 3},
		}
	}

	// 准备输入
	input := &core.AgentInput{
		Task:        "Process large dataset in streaming fashion",
		Instruction: "Process data in batches and report progress",
		Context: map[string]interface{}{
			"data_source": dataSource,
		},
		Options:   core.DefaultAgentOptions(),
		Timestamp: time.Now(),
	}

	ctx := context.Background()

	// 演示 1: 基本流式处理
	fmt.Println("=== Example 1: Basic Stream Processing ===")
	fmt.Println()

	streamOutput, err := agent.ExecuteStream(ctx, input)
	if err != nil {
		log.Fatalf("Failed to execute stream: %v", err)
	}
	defer streamOutput.Close()

	fmt.Printf("Processing %d items in batches of %d...\n\n", dataSize, config.BatchSize)

	batchCount := 0
	itemsProcessed := 0
	startTime := time.Now()

	for {
		chunk, err := streamOutput.Next()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("Error reading chunk: %v", err)
			break
		}

		switch chunk.Type {
		case core.ChunkTypeJSON:
			batchCount++
			result, ok := chunk.Data.(map[string]interface{})
			if ok {
				batchSize := int(result["batch_size"].(float64))
				itemsProcessed += batchSize
				fmt.Printf("Batch %d: processed %d items (Total: %d/%d)\n",
					batchCount, batchSize, itemsProcessed, dataSize)
			}

		case core.ChunkTypeProgress:
			fmt.Printf("Progress: %.1f%% (%s)\n",
				chunk.Metadata.Progress, chunk.Data.(map[string]interface{})["message"])

		case core.ChunkTypeStatus:
			fmt.Printf("[Status] %s\n", chunk.Data)
		}
	}

	elapsed := time.Since(startTime)
	throughput := float64(itemsProcessed) / elapsed.Seconds()

	fmt.Printf("\nProcessing complete!\n")
	fmt.Printf("Total batches: %d\n", batchCount)
	fmt.Printf("Total items: %d\n", itemsProcessed)
	fmt.Printf("Time elapsed: %v\n", elapsed)
	fmt.Printf("Throughput: %.2f items/sec\n", throughput)

	// 演示 2: 使用转换函数
	fmt.Println("\n=== Example 2: Stream with Transform ===")
	fmt.Println()

	simpleData := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		simpleData[i] = i
	}

	// 定义转换函数：平方每个数字
	transformFunc := func(item interface{}) (interface{}, error) {
		num := item.(int)
		return map[string]interface{}{
			"original": num,
			"squared":  num * num,
		}, nil
	}

	streamOutput2, err := agent.ProcessWithTransform(ctx, simpleData, transformFunc)
	if err != nil {
		log.Fatalf("Failed to execute transform stream: %v", err)
	}
	defer streamOutput2.Close()

	fmt.Println("Transforming numbers (showing first 10):")
	count := 0
	for {
		chunk, err := streamOutput2.Next()
		if err != nil {
			break
		}

		if chunk.Type == core.ChunkTypeJSON && count < 10 {
			result := chunk.Data.(map[string]interface{})
			fmt.Printf("  %v -> squared = %v\n", result["original"], result["squared"])
			count++
		}
	}

	// 演示 3: 流式过滤
	fmt.Println("\n=== Example 3: Stream Filtering ===")
	fmt.Println()

	streamOutput3, err := agent.ExecuteStream(ctx, input)
	if err != nil {
		log.Fatalf("Failed to execute stream: %v", err)
	}

	// 过滤器：只保留进度更新
	filterFunc := func(chunk *core.LegacyStreamChunk) bool {
		return chunk.Type == core.ChunkTypeProgress
	}

	filteredStream, err := agent.StreamFilter(ctx, streamOutput3, filterFunc)
	if err != nil {
		log.Fatalf("Failed to create filtered stream: %v", err)
	}
	defer filteredStream.Close()

	fmt.Println("Progress updates only:")
	for {
		chunk, err := filteredStream.Next()
		if err != nil {
			break
		}
		fmt.Printf("  Progress: %.1f%%\n", chunk.Metadata.Progress)
	}

	// 演示 4: 流式映射
	fmt.Println("\n=== Example 4: Stream Mapping ===")
	fmt.Println()

	streamOutput4, err := agent.ExecuteStream(ctx, input)
	if err != nil {
		log.Fatalf("Failed to execute stream: %v", err)
	}

	// 映射器：添加时间戳到每个块
	mapperFunc := func(chunk *core.LegacyStreamChunk) (*core.LegacyStreamChunk, error) {
		chunk.Metadata.Extra = map[string]interface{}{
			"processed_at": time.Now().Format(time.RFC3339),
			"processor":    "stream-mapper",
		}
		return chunk, nil
	}

	mappedStream, err := agent.StreamMap(ctx, streamOutput4, mapperFunc)
	if err != nil {
		log.Fatalf("Failed to create mapped stream: %v", err)
	}
	defer mappedStream.Close()

	fmt.Println("Mapped stream (showing first 5 chunks):")
	count = 0
	for {
		chunk, err := mappedStream.Next()
		if err != nil {
			break
		}

		if count < 5 && chunk.Metadata.Extra != nil {
			fmt.Printf("  Type: %s, Processed at: %s\n",
				chunk.Type, chunk.Metadata.Extra["processed_at"])
			count++
		}
	}

	// 演示 5: 流式归约
	fmt.Println("\n=== Example 5: Stream Reduce ===")
	fmt.Println()

	numbers := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		numbers[i] = i + 1
	}

	sumInput := &core.AgentInput{
		Task: "Sum numbers",
		Context: map[string]interface{}{
			"data_source": numbers,
		},
		Timestamp: time.Now(),
	}

	streamOutput5, err := agent.ExecuteStream(ctx, sumInput)
	if err != nil {
		log.Fatalf("Failed to execute stream: %v", err)
	}

	// 归约器：计算总和
	reducerFunc := func(accumulator, current interface{}) (interface{}, error) {
		sum := accumulator.(int)
		if data, ok := current.(map[string]interface{}); ok {
			if items, ok := data["items"].([]interface{}); ok {
				for _, item := range items {
					if num, ok := item.(int); ok {
						sum += num
					}
				}
			}
		}
		return sum, nil
	}

	result, err := agent.StreamReduce(ctx, streamOutput5, 0, reducerFunc)
	if err != nil {
		log.Fatalf("Failed to reduce stream: %v", err)
	}

	fmt.Printf("Sum of numbers 1-100: %v (expected: 5050)\n", result)

	fmt.Println("\n=== All Examples Completed ===")
}
