package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kart-io/k8s-agent/pkg/agent/cache"
	"github.com/kart-io/k8s-agent/pkg/agent/core"
	"github.com/kart-io/k8s-agent/pkg/agent/parsers"
)

// 示例 1: 使用 Runnable 接口和管道模式
func exampleRunnable() {
	fmt.Println("=== Example 1: Runnable Interface with Pipe ===")
	fmt.Println()

	// 创建一个简单的 Runnable 函数：文本转大写
	uppercaseRunnable := core.NewRunnableFunc(
		func(ctx context.Context, input string) (string, error) {
			return fmt.Sprintf("UPPERCASE: %s", input), nil
		},
	)

	// 使用 Pipe 连接两个 Runnable (类似 LangChain 的 LCEL)
	// 注意：由于 Go 泛型限制，Pipe 返回类型为 Runnable[string, any]
	pipeline := uppercaseRunnable.Pipe(core.NewRunnableFunc(
		func(ctx context.Context, input string) (any, error) {
			return fmt.Sprintf("PREFIX -> %s", input), nil
		},
	))

	// 执行
	result, err := pipeline.Invoke(context.Background(), "hello world")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Input: hello world\n")
	fmt.Printf("Output: %s\n\n", result)
}

// 示例 2: 使用 Callbacks 监控执行过程
func exampleCallbacks() {
	fmt.Println("=== Example 2: Callbacks for Monitoring ===")
	fmt.Println()

	// 创建 Stdout 回调（输出到控制台）
	stdoutCallback := core.NewStdoutCallback(true) // 启用颜色

	// 创建 Metrics 回调（收集指标）
	metricsCollector := &SimpleMetricsCollector{}
	metricsCallback := core.NewMetricsCallback(metricsCollector)

	// 创建带回调的 Runnable
	processRunnable := core.NewRunnableFunc(
		func(ctx context.Context, input string) (string, error) {
			// 模拟 LLM 调用
			stdoutCallback.OnLLMStart(ctx, []string{input}, "gpt-4")
			time.Sleep(100 * time.Millisecond) // 模拟延迟
			result := fmt.Sprintf("Processed: %s", input)
			stdoutCallback.OnLLMEnd(ctx, result, 50)
			return result, nil
		},
	).WithCallbacks(stdoutCallback, metricsCallback)

	// 执行
	result, _ := processRunnable.Invoke(context.Background(), "analyze system logs")
	fmt.Printf("\nResult: %s\n\n", result)

	// 查看收集的指标
	fmt.Printf("Metrics Collected:\n")
	fmt.Printf("- Total Calls: %d\n", metricsCollector.calls)
	fmt.Printf("- Total Latency: %v\n\n", metricsCollector.totalLatency)
}

// 示例 3: 使用 Output Parsers 解析结构化输出
func exampleOutputParsers() {
	fmt.Println("=== Example 3: Output Parsers for Structured Data ===")
	fmt.Println()

	// 定义输出结构
	type AnalysisResult struct {
		RootCause   string   `json:"root_cause" description:"The identified root cause"`
		Confidence  float64  `json:"confidence" description:"Confidence score 0-1"`
		Suggestions []string `json:"suggestions" description:"List of suggestions"`
	}

	// 创建 JSON 输出解析器
	parser := parsers.NewJSONOutputParser[AnalysisResult](false)

	// 模拟 LLM 返回的 JSON 输出（带 markdown 代码块）
	llmOutput := `
Based on the analysis, here are my findings:

` + "```json" + `
{
  "root_cause": "Database connection pool exhausted",
  "confidence": 0.95,
  "suggestions": [
    "Increase connection pool size",
    "Implement connection timeout",
    "Add connection retry logic"
  ]
}
` + "```" + `

These recommendations should resolve the issue.
`

	// 解析输出
	result, err := parser.Parse(context.Background(), llmOutput)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Parsed Result:\n")
	fmt.Printf("- Root Cause: %s\n", result.RootCause)
	fmt.Printf("- Confidence: %.2f\n", result.Confidence)
	fmt.Printf("- Suggestions:\n")
	for _, suggestion := range result.Suggestions {
		fmt.Printf("  * %s\n", suggestion)
	}
	fmt.Println()

	// 获取格式化指令（用于 prompt）
	fmt.Printf("Format Instructions:\n%s\n\n", parser.GetFormatInstructions())
}

// 示例 4: 使用 Caching 优化性能
func exampleCaching() {
	fmt.Println("=== Example 4: Caching for Performance ===")
	fmt.Println()

	// 创建内存缓存
	cacheInstance := cache.NewInMemoryCache(
		100,           // maxSize
		5*time.Minute, // defaultTTL
		1*time.Minute, // cleanupInterval
	)

	// 创建缓存键生成器
	keyGen := cache.NewCacheKeyGenerator("llm")

	// 模拟 LLM 调用（耗时操作）
	expensiveLLMCall := func(ctx context.Context, prompt string) string {
		// 生成缓存键
		key := keyGen.GenerateKeySimple(prompt)

		// 检查缓存
		if cached, err := cacheInstance.Get(ctx, key); err == nil {
			fmt.Printf("🚀 Cache HIT for prompt: %s\n", prompt)
			return cached.(string)
		}

		fmt.Printf("💤 Cache MISS for prompt: %s\n", prompt)

		// 模拟 LLM 调用延迟
		time.Sleep(500 * time.Millisecond)
		result := fmt.Sprintf("LLM Response for: %s", prompt)

		// 缓存结果
		_ = cacheInstance.Set(ctx, key, result, 0)

		return result
	}

	// 第一次调用（缓存未命中）
	start := time.Now()
	result1 := expensiveLLMCall(context.Background(), "What is Kubernetes?")
	fmt.Printf("First call took: %v\n", time.Since(start))
	fmt.Printf("Result: %s\n\n", result1)

	// 第二次调用相同问题（缓存命中）
	start = time.Now()
	result2 := expensiveLLMCall(context.Background(), "What is Kubernetes?")
	fmt.Printf("Second call took: %v\n", time.Since(start))
	fmt.Printf("Result: %s\n\n", result2)

	// 显示缓存统计
	stats := cacheInstance.GetStats()
	fmt.Printf("Cache Stats:\n")
	fmt.Printf("- Hits: %d\n", stats.Hits)
	fmt.Printf("- Misses: %d\n", stats.Misses)
	fmt.Printf("- Hit Rate: %.2f%%\n", stats.HitRate*100)
	fmt.Printf("- Size: %d/%d\n\n", stats.Size, stats.MaxSize)
}

// 示例 5: 综合应用 - LangChain 风格的 RAG Pipeline
func exampleRAGPipeline() {
	fmt.Println("=== Example 5: LangChain-style RAG Pipeline ===")
	fmt.Println()

	// 定义结构化输出
	type RAGResponse struct {
		Answer     string   `json:"answer"`
		Sources    []string `json:"sources"`
		Confidence float64  `json:"confidence"`
	}

	// 创建组件
	cacheInstance := cache.NewInMemoryCache(100, 5*time.Minute, 1*time.Minute)
	parser := parsers.NewJSONOutputParser[RAGResponse](false)
	stdoutCallback := core.NewStdoutCallback(true)

	// Step 1: 检索相关文档 (Retrieval)
	retrievalStep := core.NewRunnableFunc(
		func(ctx context.Context, query string) ([]string, error) {
			stdoutCallback.OnToolStart(ctx, "VectorStoreRetriever", query)
			// 模拟从向量数据库检索
			docs := []string{
				"Kubernetes is a container orchestration platform",
				"K8s provides automated deployment and scaling",
				"Pods are the smallest deployable units in Kubernetes",
			}
			stdoutCallback.OnToolEnd(ctx, "VectorStoreRetriever", docs)
			return docs, nil
		},
	)

	// Step 2: 生成 Prompt (Augment)
	promptStep := core.NewRunnableFunc(
		func(ctx context.Context, input []string) (string, error) {
			query := "What is Kubernetes?" // 简化示例
			docs := input
			prompt := fmt.Sprintf(
				"Context:\n%s\n\nQuestion: %s\n\n%s",
				fmt.Sprintf("%v", docs),
				query,
				parser.GetFormatInstructions(),
			)
			return prompt, nil
		},
	)

	// Step 3: 调用 LLM (Generate)
	llmStep := core.NewRunnableFunc(
		func(ctx context.Context, prompt string) (string, error) {
			stdoutCallback.OnLLMStart(ctx, []string{prompt}, "gpt-4")

			// 检查缓存
			keyGen := cache.NewCacheKeyGenerator("rag")
			key := keyGen.GenerateKeySimple(prompt)

			if cached, err := cacheInstance.Get(ctx, key); err == nil {
				stdoutCallback.OnLLMEnd(ctx, cached.(string), 0)
				return cached.(string), nil
			}

			// 模拟 LLM 调用
			time.Sleep(200 * time.Millisecond)
			response := `
` + "```json" + `
{
  "answer": "Kubernetes is a container orchestration platform that provides automated deployment, scaling, and management of containerized applications.",
  "sources": ["doc1", "doc2"],
  "confidence": 0.92
}
` + "```" + `
`
			cacheInstance.Set(ctx, key, response, 0)
			stdoutCallback.OnLLMEnd(ctx, response, 150)
			return response, nil
		},
	)

	// Step 4: 解析输出
	parseStep := core.NewRunnableFunc(
		func(ctx context.Context, llmOutput string) (RAGResponse, error) {
			return parser.Parse(ctx, llmOutput)
		},
	)

	// 构建 RAG 管道 (类似 LangChain 的 LCEL)
	// 注意：由于 Go 泛型限制，需要将中间步骤的返回类型调整为 any
	ragPipeline := retrievalStep.
		Pipe(core.NewRunnableFunc(func(ctx context.Context, input []string) (any, error) {
			return promptStep.Invoke(ctx, input)
		})).
		Pipe(core.NewRunnableFunc(func(ctx context.Context, input any) (any, error) {
			return llmStep.Invoke(ctx, input.(string))
		})).
		Pipe(core.NewRunnableFunc(func(ctx context.Context, input any) (any, error) {
			return parseStep.Invoke(ctx, input.(string))
		})).
		WithCallbacks(stdoutCallback)

	// 执行 RAG 查询
	fmt.Println("Executing RAG Pipeline...")
	fmt.Println()
	result, err := ragPipeline.Invoke(context.Background(), "What is Kubernetes?")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n✅ RAG Pipeline Result:\n")
	ragResp := result.(RAGResponse)
	fmt.Printf("Answer: %s\n", ragResp.Answer)
	fmt.Printf("Sources: %v\n", ragResp.Sources)
	fmt.Printf("Confidence: %.2f\n\n", ragResp.Confidence)
}

// 示例 6: 批处理和并发执行
func exampleBatchProcessing() {
	fmt.Println("=== Example 6: Batch Processing ===")
	fmt.Println()

	// 创建处理 Runnable
	processor := core.NewRunnableFunc(
		func(ctx context.Context, input string) (string, error) {
			// 模拟处理延迟
			time.Sleep(100 * time.Millisecond)
			return fmt.Sprintf("Processed: %s", input), nil
		},
	)

	// 批量输入
	inputs := []string{
		"task 1",
		"task 2",
		"task 3",
		"task 4",
		"task 5",
	}

	// 批量执行（并发处理）
	fmt.Printf("Processing %d tasks in batch...\n", len(inputs))
	start := time.Now()

	results, err := processor.Batch(context.Background(), inputs)
	if err != nil {
		log.Fatal(err)
	}

	duration := time.Since(start)
	fmt.Printf("Completed in: %v\n", duration)
	fmt.Printf("Results:\n")
	for i, result := range results {
		fmt.Printf("  %d. %s\n", i+1, result)
	}
	fmt.Println()
}

// SimpleMetricsCollector 简单的指标收集器实现
type SimpleMetricsCollector struct {
	calls        int64
	totalLatency time.Duration
}

func (m *SimpleMetricsCollector) IncrementCounter(name string, value int64, tags map[string]string) {
	m.calls += value
}

func (m *SimpleMetricsCollector) RecordHistogram(name string, value float64, tags map[string]string) {
	if name == "llm.latency" {
		m.totalLatency += time.Duration(value * float64(time.Second))
	}
}

func (m *SimpleMetricsCollector) RecordGauge(name string, value float64, tags map[string]string) {
	// No-op for this example
}

func main() {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║  LangChain-Inspired pkg/agent Examples                   ║")
	fmt.Println("║  Demonstrating Runnable, Callbacks, Parsers, and Cache   ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 运行所有示例
	exampleRunnable()
	exampleCallbacks()
	exampleOutputParsers()
	exampleCaching()
	exampleRAGPipeline()
	exampleBatchProcessing()

	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║  All Examples Completed Successfully! 🎉                  ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
}
