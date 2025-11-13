package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kart-io/k8s-agent/pkg/agent/core"
	"github.com/kart-io/k8s-agent/pkg/agent/stream"
)

//nolint:gocyclo // Example code with comprehensive demo flow
func main() {
	fmt.Println("=== Progress Tracking Example ===")
	fmt.Println()

	// 创建进度 Agent
	config := &stream.ProgressConfig{
		EnableProgress:   true,
		ProgressInterval: 200 * time.Millisecond,
		EnableETA:        true,
		EnablePhases:     true,
	}

	agent := stream.NewProgressAgent(config)

	// 演示 1: 基本进度跟踪
	fmt.Println("=== Example 1: Basic Progress Tracking ===")
	fmt.Println()

	input := &core.AgentInput{
		Task:        "Long running task with progress tracking",
		Instruction: "Execute task and report progress in real-time",
		Context: map[string]interface{}{
			"total_steps":   100,
			"step_duration": 50 * time.Millisecond,
		},
		Options:   core.DefaultAgentOptions(),
		Timestamp: time.Now(),
	}

	ctx := context.Background()

	streamOutput, err := agent.ExecuteStream(ctx, input)
	if err != nil {
		log.Fatalf("Failed to execute stream: %v", err)
	}
	defer streamOutput.Close()

	fmt.Println("Executing task with real-time progress...")
	fmt.Println()

	lastProgress := 0.0
	lastPhase := ""

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
		case core.ChunkTypeProgress:
			data := chunk.Data.(map[string]interface{})
			progress := chunk.Metadata.Progress
			step := int(data["step"].(float64))
			total := int(data["total"].(float64))
			phase := data["phase"].(string)
			eta := chunk.Metadata.ETA

			// 只在进度有显著变化时打印（每 10%）
			if progress-lastProgress >= 10.0 || progress == 100.0 {
				fmt.Printf("[Step %d/%d] Phase: %s | Progress: %.1f%% | ETA: %v\n",
					step, total, phase, progress, eta)
				lastProgress = progress
			}

		case core.ChunkTypeStatus:
			status := chunk.Data.(string)
			if status != lastPhase {
				fmt.Printf("\n>>> %s\n\n", status)
				lastPhase = status
			}

		case core.ChunkTypeJSON:
			fmt.Println("\nFinal result received:")
			result := chunk.Data.(map[string]interface{})
			for key, value := range result {
				fmt.Printf("  %s: %v\n", key, value)
			}
		}
	}

	fmt.Println("\nTask completed!")

	// 演示 2: 进度条可视化
	fmt.Println("\n=== Example 2: Progress Bar Visualization ===")
	fmt.Println()

	input2 := &core.AgentInput{
		Task: "Task with visual progress bar",
		Context: map[string]interface{}{
			"total_steps":   50,
			"step_duration": 100 * time.Millisecond,
		},
		Timestamp: time.Now(),
	}

	streamOutput2, err := agent.ExecuteStream(ctx, input2)
	if err != nil {
		log.Fatalf("Failed to execute stream: %v", err)
	}
	defer streamOutput2.Close()

	fmt.Println("Task execution with progress bar:")
	fmt.Println()

	for {
		chunk, err := streamOutput2.Next()
		if err != nil {
			break
		}

		if chunk.Type == core.ChunkTypeProgress {
			progress := chunk.Metadata.Progress
			drawProgressBar(progress)
		}
	}

	fmt.Println("\n\nTask completed!")

	// 演示 3: 多任务进度跟踪
	fmt.Println("\n=== Example 3: Multiple Task Progress ===")
	fmt.Println()

	tasks := []struct {
		name  string
		steps int
	}{
		{"Task A", 30},
		{"Task B", 50},
		{"Task C", 20},
	}

	fmt.Println("Executing multiple tasks concurrently:")
	fmt.Println()

	type taskProgress struct {
		name     string
		progress float64
		phase    string
	}

	progressChan := make(chan taskProgress, len(tasks))

	// 启动多个任务
	for _, task := range tasks {
		go func(taskName string, steps int) {
			input := &core.AgentInput{
				Task: taskName,
				Context: map[string]interface{}{
					"total_steps":   steps,
					"step_duration": 50 * time.Millisecond,
				},
				Timestamp: time.Now(),
			}

			stream, err := agent.ExecuteStream(ctx, input)
			if err != nil {
				log.Printf("Failed to start %s: %v", taskName, err)
				return
			}
			defer stream.Close()

			for {
				chunk, err := stream.Next()
				if err != nil {
					break
				}

				if chunk.Type == core.ChunkTypeProgress {
					data := chunk.Data.(map[string]interface{})
					progressChan <- taskProgress{
						name:     taskName,
						progress: chunk.Metadata.Progress,
						phase:    data["phase"].(string),
					}
				}
			}
		}(task.name, task.steps)
	}

	// 收集并显示所有任务的进度
	taskStates := make(map[string]float64)
	completedTasks := 0

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for completedTasks < len(tasks) {
		select {
		case progress := <-progressChan:
			taskStates[progress.name] = progress.progress

			if progress.progress >= 100.0 {
				completedTasks++
			}

		case <-ticker.C:
			// 定期更新显示
			fmt.Print("\033[H\033[2J") // 清屏
			fmt.Println("=== Multiple Task Progress ===")
			fmt.Println()

			for _, task := range tasks {
				progress := taskStates[task.name]
				fmt.Printf("%s: ", task.name)
				drawProgressBar(progress)
				fmt.Println()
			}

			fmt.Printf("\nCompleted: %d/%d tasks\n", completedTasks, len(tasks))
		}
	}

	fmt.Println("\n\nAll tasks completed!")

	// 演示 4: 自定义进度跟踪器
	fmt.Println("\n=== Example 4: Custom Progress Tracker ===")
	fmt.Println()

	customInput := &core.AgentInput{
		Task: "Custom progress tracking",
		Context: map[string]interface{}{
			"total_steps":   200,
			"step_duration": 20 * time.Millisecond,
		},
		Timestamp: time.Now(),
	}

	streamOutput4, err := agent.ExecuteStream(ctx, customInput)
	if err != nil {
		log.Fatalf("Failed to execute stream: %v", err)
	}
	defer streamOutput4.Close()

	fmt.Println("Processing with custom progress metrics:")
	fmt.Println()

	var (
		totalChunks     int
		progressUpdates int
		statusUpdates   int
		startTime       = time.Now()
	)

	for {
		chunk, err := streamOutput4.Next()
		if err != nil {
			break
		}

		totalChunks++

		switch chunk.Type {
		case core.ChunkTypeProgress:
			progressUpdates++
			if progressUpdates%10 == 0 {
				progress := chunk.Metadata.Progress
				elapsed := time.Since(startTime)
				rate := float64(progressUpdates) / elapsed.Seconds()
				fmt.Printf("Progress: %.1f%% | Updates: %d | Rate: %.2f updates/sec\n",
					progress, progressUpdates, rate)
			}

		case core.ChunkTypeStatus:
			statusUpdates++
		}
	}

	elapsed := time.Since(startTime)
	fmt.Printf("\nMetrics:\n")
	fmt.Printf("  Total chunks: %d\n", totalChunks)
	fmt.Printf("  Progress updates: %d\n", progressUpdates)
	fmt.Printf("  Status updates: %d\n", statusUpdates)
	fmt.Printf("  Total time: %v\n", elapsed)
	fmt.Printf("  Average chunk rate: %.2f chunks/sec\n", float64(totalChunks)/elapsed.Seconds())

	fmt.Println("\n=== All Examples Completed ===")
}

// drawProgressBar 绘制进度条
func drawProgressBar(progress float64) {
	barWidth := 40
	filled := int(progress / 100.0 * float64(barWidth))

	fmt.Print("[")
	for i := 0; i < barWidth; i++ {
		if i < filled {
			fmt.Print("=")
		} else if i == filled {
			fmt.Print(">")
		} else {
			fmt.Print(" ")
		}
	}
	fmt.Printf("] %.1f%%", progress)
}
