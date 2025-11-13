package benchmarks

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/kart-io/k8s-agent/pkg/agent/agents"
	"github.com/kart-io/k8s-agent/pkg/agent/core"
	"github.com/kart-io/k8s-agent/pkg/agent/llm"
	"github.com/kart-io/k8s-agent/pkg/agent/llm/providers"
	"github.com/kart-io/k8s-agent/pkg/agent/store"
	"github.com/kart-io/k8s-agent/pkg/agent/stream"
	"github.com/kart-io/k8s-agent/pkg/agent/tools"
	"github.com/kart-io/k8s-agent/pkg/agent/tools/practical"
)

// BenchmarkLLMProviders compares performance of different LLM providers
func BenchmarkLLMProviders(b *testing.B) {
	ctx := context.Background()

	// Initialize providers
	openAI, _ := providers.NewOpenAI(&llm.Config{
		APIKey:    "test-key",
		Model:     "gpt-4",
		MaxTokens: 100,
	})

	gemini, _ := providers.NewGemini(&llm.Config{
		APIKey:    "test-key",
		Model:     "gemini-pro",
		MaxTokens: 100,
	})

	deepseek, _ := providers.NewDeepSeek(&llm.Config{
		APIKey:    "test-key",
		Model:     "deepseek-chat",
		MaxTokens: 100,
	})

	providers := map[string]llm.LLM{
		"OpenAI":   openAI,
		"Gemini":   gemini,
		"DeepSeek": deepseek,
	}

	prompt := "What is the capital of France?"

	for name, provider := range providers {
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				req := &llm.CompletionRequest{
					Messages: []llm.Message{
						{Role: "user", Content: prompt},
					},
				}
				_, _ = provider.Complete(ctx, req)
			}
		})
	}
}

// BenchmarkStreamingModes compares different streaming modes
func BenchmarkStreamingModes(b *testing.B) {
	modes := []stream.StreamMode{
		stream.ModeMessages,
		stream.ModeUpdates,
		stream.ModeCustom,
		stream.ModeValues,
	}

	for _, mode := range modes {
		b.Run(string(mode), func(b *testing.B) {
			config := &stream.StreamConfig{
				Modes:         []stream.StreamMode{mode},
				BufferSize:    1000,
				FlushInterval: 10 * time.Millisecond,
			}

			multiStream := stream.NewMultiModeStream(config)
			multiStream.Start()
			defer multiStream.Stop()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				event := stream.StreamEvent{
					Type:      "test",
					Data:      fmt.Sprintf("event_%d", i),
					Timestamp: time.Now(),
					Mode:      mode,
				}
				multiStream.Send(event)
			}
		})
	}
}

// BenchmarkToolExecution tests tool execution performance
func BenchmarkToolExecution(b *testing.B) {
	executor := tools.NewToolExecutor(tools.WithMaxConcurrency(10))

	// Create mock tools
	mockTool := &MockTool{
		delay: 10 * time.Millisecond,
	}

	toolCounts := []int{1, 5, 10, 20, 50}

	for _, count := range toolCounts {
		b.Run(fmt.Sprintf("Tools_%d", count), func(b *testing.B) {
			calls := make([]*tools.ToolCall, count)
			for i := 0; i < count; i++ {
				calls[i] = &tools.ToolCall{
					ID:    fmt.Sprintf("tool_%d", i),
					Tool:  mockTool,
					Input: &tools.ToolInput{Data: map[string]interface{}{"index": i}},
				}
			}

			ctx := context.Background()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = executor.ExecuteParallel(ctx, calls)
			}
		})
	}
}

// BenchmarkLangGraphStore tests store performance
func BenchmarkLangGraphStore(b *testing.B) {
	store := store.NewInMemoryLangGraphStore()
	ctx := context.Background()

	b.Run("Put", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			namespace := []string{"bench", fmt.Sprintf("ns_%d", i%100)}
			key := fmt.Sprintf("key_%d", i)
			value := map[string]interface{}{
				"index":     i,
				"data":      fmt.Sprintf("value_%d", i),
				"timestamp": time.Now(),
			}
			store.Put(ctx, namespace, key, value)
		}
	})

	// Populate store for read tests
	for i := 0; i < 1000; i++ {
		namespace := []string{"bench", fmt.Sprintf("ns_%d", i%100)}
		key := fmt.Sprintf("key_%d", i)
		value := map[string]interface{}{
			"index": i,
			"data":  fmt.Sprintf("value_%d", i),
		}
		store.Put(ctx, namespace, key, value)
	}

	b.Run("Get", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			namespace := []string{"bench", fmt.Sprintf("ns_%d", i%100)}
			key := fmt.Sprintf("key_%d", i%1000)
			store.Get(ctx, namespace, key)
		}
	})

	b.Run("Search", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			namespace := []string{"bench", fmt.Sprintf("ns_%d", i%100)}
			pattern := fmt.Sprintf("value_%d", i%100)
			store.Search(ctx, namespace, pattern, 10)
		}
	})

	b.Run("List", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			namespace := []string{"bench", fmt.Sprintf("ns_%d", i%100)}
			store.List(ctx, namespace)
		}
	})
}

// BenchmarkCachedStore compares cached vs non-cached store
func BenchmarkCachedStore(b *testing.B) {
	ctx := context.Background()
	backend := store.NewInMemoryLangGraphStore()
	cached := store.NewStoreWithCache(backend, 5*time.Minute)

	// Populate stores
	for i := 0; i < 1000; i++ {
		namespace := []string{"bench"}
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		backend.Put(ctx, namespace, key, value)
	}

	b.Run("DirectStore", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			namespace := []string{"bench"}
			key := fmt.Sprintf("key_%d", i%1000)
			backend.Get(ctx, namespace, key)
		}
	})

	b.Run("CachedStore", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			namespace := []string{"bench"}
			key := fmt.Sprintf("key_%d", i%1000)
			cached.Get(ctx, namespace, key)
		}
	})
}

// BenchmarkAgentRouting tests different routing strategies
func BenchmarkAgentRouting(b *testing.B) {
	// Create mock agents
	mockAgents := make(map[string]core.Agent)
	for i := 0; i < 10; i++ {
		mockAgents[fmt.Sprintf("agent_%d", i)] = &MockAgent{
			id:           fmt.Sprintf("agent_%d", i),
			capabilities: []string{"cap1", "cap2", "cap3"},
		}
	}

	routers := map[string]agents.AgentRouter{
		"RoundRobin": &agents.RoundRobinRouter{},
		"Random":     &agents.RandomRouter{},
		"Capability": &agents.CapabilityRouter{
			AgentCapabilities: map[string][]string{
				"agent_0": {"search", "analyze"},
				"agent_1": {"code", "review"},
				"agent_2": {"data", "process"},
			},
		},
		"LoadBalancing": &agents.LoadBalancingRouter{
			CurrentLoads: make(map[string]int),
			MaxLoad:      10,
		},
	}

	ctx := context.Background()
	task := agents.SupervisorTask{
		ID:          "test_task",
		Description: "Test routing task",
		Type:        agents.TaskTypeProcess,
	}

	for name, router := range routers {
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				router.Route(ctx, task, mockAgents)
			}
		})
	}
}

// BenchmarkPracticalTools tests performance of practical tools
func BenchmarkPracticalTools(b *testing.B) {
	ctx := context.Background()

	b.Run("WebScraper", func(b *testing.B) {
		tool := practical.NewWebScraperTool()
		input := map[string]interface{}{
			"url": "https://example.com",
			"selectors": map[string]interface{}{
				"title":   "h1",
				"content": "p",
			},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tool.Execute(ctx, input)
		}
	})

	b.Run("APICallerWithCache", func(b *testing.B) {
		tool := practical.NewAPICallerTool()
		input := map[string]interface{}{
			"url":    "https://api.example.com/data",
			"method": "GET",
			"cache":  true,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tool.Execute(ctx, input)
		}
	})

	b.Run("FileOperations", func(b *testing.B) {
		tool := practical.NewFileOperationsTool("/tmp")

		b.Run("Write", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				input := map[string]interface{}{
					"operation": "write",
					"path":      fmt.Sprintf("/tmp/bench_%d.txt", i),
					"content":   fmt.Sprintf("benchmark content %d", i),
				}
				tool.Execute(ctx, input)
			}
		})

		b.Run("Read", func(b *testing.B) {
			// Create test file
			input := map[string]interface{}{
				"operation": "write",
				"path":      "/tmp/bench_read.txt",
				"content":   "test content for reading",
			}
			tool.Execute(ctx, input)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				input := map[string]interface{}{
					"operation": "read",
					"path":      "/tmp/bench_read.txt",
				}
				tool.Execute(ctx, input)
			}
		})
	})
}

// BenchmarkMemoryUsage tests memory consumption
func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("StoreMemory", func(b *testing.B) {
		b.ReportAllocs()
		store := store.NewInMemoryLangGraphStore()
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			namespace := []string{"mem", "test"}
			key := fmt.Sprintf("key_%d", i)
			value := map[string]interface{}{
				"data": make([]byte, 1024), // 1KB per entry
			}
			store.Put(ctx, namespace, key, value)
		}
	})

	b.Run("StreamMemory", func(b *testing.B) {
		b.ReportAllocs()
		config := &stream.StreamConfig{
			Modes:         []stream.StreamMode{stream.ModeMessages},
			BufferSize:    1000,
			FlushInterval: 100 * time.Millisecond,
		}
		multiStream := stream.NewMultiModeStream(config)
		multiStream.Start()
		defer multiStream.Stop()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			event := stream.StreamEvent{
				Type: "memory_test",
				Data: make([]byte, 1024), // 1KB per event
			}
			multiStream.Send(event)
		}
	})
}

// BenchmarkConcurrency tests concurrent operations
func BenchmarkConcurrency(b *testing.B) {
	store := store.NewInMemoryLangGraphStore()
	ctx := context.Background()

	b.Run("ConcurrentWrites", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				namespace := []string{"concurrent", fmt.Sprintf("ns_%d", i%10)}
				key := fmt.Sprintf("key_%d", i)
				value := fmt.Sprintf("value_%d", i)
				store.Put(ctx, namespace, key, value)
				i++
			}
		})
	})

	// Populate for reads
	for i := 0; i < 1000; i++ {
		namespace := []string{"concurrent", fmt.Sprintf("ns_%d", i%10)}
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		store.Put(ctx, namespace, key, value)
	}

	b.Run("ConcurrentReads", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				namespace := []string{"concurrent", fmt.Sprintf("ns_%d", i%10)}
				key := fmt.Sprintf("key_%d", i%1000)
				store.Get(ctx, namespace, key)
				i++
			}
		})
	})

	b.Run("MixedOperations", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				operation := i % 3
				namespace := []string{"mixed", fmt.Sprintf("ns_%d", i%10)}

				switch operation {
				case 0: // Write
					key := fmt.Sprintf("key_%d", i)
					value := fmt.Sprintf("value_%d", i)
					store.Put(ctx, namespace, key, value)
				case 1: // Read
					key := fmt.Sprintf("key_%d", i%100)
					store.Get(ctx, namespace, key)
				case 2: // List
					store.List(ctx, namespace)
				}
				i++
			}
		})
	})
}

// BenchmarkEndToEnd tests complete workflows
func BenchmarkEndToEnd(b *testing.B) {
	ctx := context.Background()

	// Setup complete system
	llmProvider, _ := providers.NewOpenAI(&llm.Config{
		APIKey:    "test-key",
		Model:     "gpt-4",
		MaxTokens: 100,
	})

	langStore := store.NewInMemoryLangGraphStore()
	cachedStore := store.NewStoreWithCache(langStore, 5*time.Minute)

	toolRegistry := tools.NewRegistry()
	toolRegistry.Register(practical.NewWebScraperRuntimeTool())
	toolRegistry.Register(practical.NewAPICallerRuntimeTool())

	executor := tools.NewParallelToolExecutor()
	executor.WithMaxConcurrency(10)

	// Create agents
	agent1 := &MockAgent{
		id:           "agent1",
		capabilities: []string{"search", "analyze"},
	}
	agent2 := &MockAgent{
		id:           "agent2",
		capabilities: []string{"process", "generate"},
	}

	supervisor := &agents.SupervisorAgent{
		Name: "Benchmark Supervisor",
		SubAgents: map[string]core.Agent{
			"agent1": agent1,
			"agent2": agent2,
		},
		Router:      &agents.RoundRobinRouter{},
		MemoryStore: cachedStore,
	}

	b.Run("SimpleWorkflow", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			task := agents.SupervisorTask{
				ID:          fmt.Sprintf("task_%d", i),
				Description: "Simple benchmark task",
				Type:        agents.TaskTypeProcess,
			}
			supervisor.Execute(ctx, task)
		}
	})

	b.Run("ComplexWorkflow", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			task := agents.SupervisorTask{
				ID:          fmt.Sprintf("complex_%d", i),
				Description: "Complex multi-step task",
				Type:        agents.TaskTypeAnalyze,
				Requirements: []string{
					"Step 1: Gather data",
					"Step 2: Process information",
					"Step 3: Generate report",
				},
				Context: map[string]interface{}{
					"depth":      "comprehensive",
					"iterations": 5,
				},
			}
			supervisor.Execute(ctx, task)
		}
	})
}

// Mock implementations for benchmarking

type MockTool struct {
	delay time.Duration
}

func (t *MockTool) GetName() string        { return "mock_tool" }
func (t *MockTool) GetDescription() string { return "Mock tool for benchmarking" }
func (t *MockTool) GetInputSchema() interface{} {
	return map[string]interface{}{"type": "object"}
}

func (t *MockTool) GetOutputSchema() interface{} {
	return map[string]interface{}{"type": "object"}
}

func (t *MockTool) Execute(ctx context.Context, input *tools.ToolInput) (*tools.ToolOutput, error) {
	if t.delay > 0 {
		time.Sleep(t.delay)
	}
	return &tools.ToolOutput{Result: map[string]interface{}{"result": "success"}}, nil
}

type MockAgent struct {
	id           string
	capabilities []string
	delay        time.Duration
}

func (a *MockAgent) GetID() string                        { return a.id }
func (a *MockAgent) GetName() string                      { return a.id }
func (a *MockAgent) GetDescription() string               { return "Mock agent" }
func (a *MockAgent) GetCapabilities() []string            { return a.capabilities }
func (a *MockAgent) GetStatus() core.AgentStatus          { return core.StatusReady }
func (a *MockAgent) GetMemory() core.Memory               { return nil }
func (a *MockAgent) GetTools() []tools.Tool               { return nil }
func (a *MockAgent) GetModel() llm.LLM                    { return nil }
func (a *MockAgent) Initialize(ctx context.Context) error { return nil }
func (a *MockAgent) Execute(ctx context.Context, task interface{}) (interface{}, error) {
	if a.delay > 0 {
		time.Sleep(a.delay)
	}
	return map[string]interface{}{"agent": a.id, "result": "completed"}, nil
}
func (a *MockAgent) Shutdown(ctx context.Context) error { return nil }

// Benchmark result analysis and reporting

type BenchmarkReport struct {
	Timestamp   time.Time
	Results     map[string]BenchmarkResult
	Comparisons map[string]Comparison
}

type BenchmarkResult struct {
	Name        string
	Operations  int
	NsPerOp     int64
	AllocsPerOp int
	BytesPerOp  int
}

type Comparison struct {
	Baseline    string
	Challenger  string
	Improvement float64 // Percentage
}

// GenerateBenchmarkReport creates a comprehensive performance report
func GenerateBenchmarkReport(results []testing.BenchmarkResult) *BenchmarkReport {
	report := &BenchmarkReport{
		Timestamp:   time.Now(),
		Results:     make(map[string]BenchmarkResult),
		Comparisons: make(map[string]Comparison),
	}

	// Process results
	for _, result := range results {
		br := BenchmarkResult{
			Name:        result.Name,
			Operations:  result.N,
			NsPerOp:     result.NsPerOp(),
			AllocsPerOp: int(result.AllocsPerOp()),
			BytesPerOp:  int(result.AllocedBytesPerOp()),
		}
		report.Results[result.Name] = br
	}

	// Generate comparisons
	report.generateComparisons()

	return report
}

func (r *BenchmarkReport) generateComparisons() {
	// Compare cached vs non-cached
	if cached, ok := r.Results["CachedStore"]; ok {
		if direct, ok := r.Results["DirectStore"]; ok {
			improvement := float64(direct.NsPerOp-cached.NsPerOp) / float64(direct.NsPerOp) * 100
			r.Comparisons["CacheEfficiency"] = Comparison{
				Baseline:    "DirectStore",
				Challenger:  "CachedStore",
				Improvement: improvement,
			}
		}
	}

	// Compare parallel execution
	if single, ok := r.Results["Tools_1"]; ok {
		if parallel, ok := r.Results["Tools_10"]; ok {
			// Calculate throughput improvement
			singleThroughput := float64(single.Operations) / float64(single.NsPerOp)
			parallelThroughput := float64(parallel.Operations) / float64(parallel.NsPerOp)
			improvement := (parallelThroughput - singleThroughput) / singleThroughput * 100
			r.Comparisons["ParallelScaling"] = Comparison{
				Baseline:    "Sequential",
				Challenger:  "Parallel",
				Improvement: improvement,
			}
		}
	}
}

func (r *BenchmarkReport) PrintSummary() {
	fmt.Println("=== Performance Benchmark Report ===")
	fmt.Printf("Generated: %s\n\n", r.Timestamp.Format(time.RFC3339))

	fmt.Println("Top Performers:")
	// Sort and display top 5 fastest operations
	for name, result := range r.Results {
		fmt.Printf("- %s: %d ns/op (%d allocs, %d bytes)\n",
			name, result.NsPerOp, result.AllocsPerOp, result.BytesPerOp)
	}

	fmt.Println("\nKey Comparisons:")
	for name, comp := range r.Comparisons {
		fmt.Printf("- %s: %.2f%% improvement (%s vs %s)\n",
			name, comp.Improvement, comp.Challenger, comp.Baseline)
	}

	fmt.Println("\nMemory Efficiency:")
	// Find most memory-efficient operations
	for name, result := range r.Results {
		if result.BytesPerOp < 1024 {
			fmt.Printf("- %s: %d bytes/op (excellent)\n", name, result.BytesPerOp)
		}
	}
}

// PerformanceMonitor tracks runtime performance
type PerformanceMonitor struct {
	metrics sync.Map
	mu      sync.RWMutex
}

func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{}
}

func (p *PerformanceMonitor) RecordLatency(operation string, duration time.Duration) {
	key := fmt.Sprintf("%s_latency", operation)
	p.metrics.Store(key, duration)
}

func (p *PerformanceMonitor) RecordThroughput(operation string, count int, duration time.Duration) {
	throughput := float64(count) / duration.Seconds()
	key := fmt.Sprintf("%s_throughput", operation)
	p.metrics.Store(key, throughput)
}

func (p *PerformanceMonitor) GetMetrics() map[string]interface{} {
	result := make(map[string]interface{})
	p.metrics.Range(func(key, value interface{}) bool {
		result[key.(string)] = value
		return true
	})
	return result
}
