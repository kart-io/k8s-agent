package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kart-io/k8s-agent/pkg/agent/agents"
	"github.com/kart-io/k8s-agent/pkg/agent/core"
	"github.com/kart-io/k8s-agent/pkg/agent/llm"
	"github.com/kart-io/k8s-agent/pkg/agent/llm/providers"
	"github.com/kart-io/k8s-agent/pkg/agent/middleware"
	"github.com/kart-io/k8s-agent/pkg/agent/store"
	"github.com/kart-io/k8s-agent/pkg/agent/stream"
	"github.com/kart-io/k8s-agent/pkg/agent/tools"
	"github.com/kart-io/k8s-agent/pkg/agent/tools/practical"
)

// ComprehensiveIntegrationExample demonstrates full integration of all components
func ComprehensiveIntegrationExample() {
	ctx := context.Background()

	// Initialize LLM providers
	openAIConfig := &llm.Config{
		APIKey:      "your-openai-key",
		Model:       "gpt-4",
		MaxTokens:   2000,
		Temperature: 0.7,
	}

	geminiConfig := &llm.Config{
		APIKey:      "your-gemini-key",
		Model:       "gemini-pro",
		MaxTokens:   2000,
		Temperature: 0.7,
	}

	deepseekConfig := &llm.Config{
		APIKey:      "your-deepseek-key",
		Model:       "deepseek-chat",
		MaxTokens:   2000,
		Temperature: 0.7,
		BaseURL:     "https://api.deepseek.com/v1",
	}

	// Create providers
	openAI, err := providers.NewOpenAI(openAIConfig)
	if err != nil {
		log.Fatal("Failed to create OpenAI provider:", err)
	}

	gemini, err := providers.NewGemini(geminiConfig)
	if err != nil {
		log.Fatal("Failed to create Gemini provider:", err)
	}

	deepseek, err := providers.NewDeepSeek(deepseekConfig)
	if err != nil {
		log.Fatal("Failed to create DeepSeek provider:", err)
	}

	// Create LangGraph store for persistent memory
	langGraphStore := store.NewInMemoryLangGraphStore()

	// Create cached store
	cachedStore := store.NewStoreWithCache(langGraphStore, 5*time.Minute)

	// Initialize practical tools
	webScraper := practical.NewWebScraperTool()
	apiCaller := practical.NewAPICallerTool()
	dbQuery := practical.NewDatabaseQueryTool()
	fileOps := practical.NewFileOperationsTool("/tmp/agent-workspace")

	// Create tool registry
	toolRegistry := tools.NewRegistry()
	toolRegistry.Register(webScraper)
	toolRegistry.Register(apiCaller)
	toolRegistry.Register(dbQuery)
	toolRegistry.Register(fileOps)

	// Setup tool executor
	toolExecutor := tools.NewToolExecutor(
		tools.WithMaxConcurrency(10),
		tools.WithTimeout(30*time.Second),
		tools.WithRetryPolicy(&tools.RetryPolicy{
			MaxRetries:   3,
			InitialDelay: 1 * time.Second,
			MaxDelay:     10 * time.Second,
			Multiplier:   2.0,
			RetryableErrors: []string{
				"timeout",
				"rate_limit",
				"temporary_failure",
			},
		}),
	)

	// Configure multi-mode streaming
	streamConfig := &stream.StreamConfig{
		Modes:           []stream.StreamMode{stream.StreamModeMessages, stream.StreamModeUpdates, stream.StreamModeCustom},
		BufferSize:      100,
		IncludeMetadata: true,
	}

	ctx := context.Background()
	multiStream := stream.NewMultiModeStream(ctx, streamConfig)

	// Create middleware chain
	toolSelectorMiddleware := &middleware.LLMToolSelectorMiddleware{
		Model:         openAI,
		MaxTools:      5,
		AlwaysInclude: []string{"web_scraper", "api_caller"},
		CacheTTL:      5 * time.Minute,
	}

	dynamicPromptMiddleware := &middleware.DynamicPromptMiddleware{
		Templates: map[string]string{
			"research": "You are a research assistant. Focus on finding accurate information from reliable sources.",
			"coding":   "You are a coding assistant. Write clean, efficient, and well-documented code.",
			"analysis": "You are a data analyst. Provide insights and patterns from the data.",
		},
		Selector:       openAI,
		ContextHistory: 5,
	}

	adaptiveMiddleware := &middleware.AdaptiveMiddleware{
		Model:            gemini,
		PerformanceStore: cachedStore,
		ConfigStore:      cachedStore,
		MetricsWindow:    100,
	}

	// Create specialized agents
	researchAgent := &core.BasicAgent{
		ID:          "research-agent",
		Name:        "Research Specialist",
		Description: "Specializes in web research and data gathering",
		Capabilities: []string{
			"web_research",
			"api_integration",
			"data_extraction",
		},
		Tools: []tools.Tool{webScraper, apiCaller},
		Model: openAI,
	}

	analysisAgent := &core.BasicAgent{
		ID:          "analysis-agent",
		Name:        "Data Analyst",
		Description: "Analyzes data and provides insights",
		Capabilities: []string{
			"data_analysis",
			"pattern_recognition",
			"report_generation",
		},
		Tools: []tools.Tool{dbQuery, fileOps},
		Model: gemini,
	}

	codeAgent := &core.BasicAgent{
		ID:          "code-agent",
		Name:        "Code Generator",
		Description: "Generates and modifies code",
		Capabilities: []string{
			"code_generation",
			"code_review",
			"refactoring",
		},
		Tools: []tools.Tool{fileOps},
		Model: deepseek,
	}

	// Create supervisor agent
	supervisor := &agents.SupervisorAgent{
		Name:        "Master Coordinator",
		Description: "Coordinates multiple specialized agents",
		SubAgents: map[string]core.Agent{
			"research": researchAgent,
			"analysis": analysisAgent,
			"code":     codeAgent,
		},
		Router: &agents.HybridRouter{
			LLMRouter: &agents.LLMRouter{
				Model: openAI,
				PromptTemplate: `Given this task: {{.Task}}

Available agents:
{{range .Agents}}
- {{.Name}}: {{.Description}}
  Capabilities: {{.Capabilities}}
{{end}}

Select the best agent for this task and explain why.`,
			},
			RuleRouter: &agents.RuleBasedRouter{
				Rules: []agents.RoutingRule{
					{
						Pattern:   "research|find|search|discover",
						AgentName: "research",
						Priority:  10,
					},
					{
						Pattern:   "analyze|insight|pattern|report",
						AgentName: "analysis",
						Priority:  10,
					},
					{
						Pattern:   "code|program|implement|develop",
						AgentName: "code",
						Priority:  10,
					},
				},
			},
			LLMWeight:  0.7,
			RuleWeight: 0.3,
		},
		Orchestrator: &agents.TaskOrchestrator{
			MaxConcurrentTasks: 5,
			TaskTimeout:        5 * time.Minute,
			EnableCaching:      true,
			CacheStore:         cachedStore,
		},
		ResultAggregator: &agents.ConsensusAggregator{
			MinAgreement:     0.7,
			RequiredAgents:   2,
			ConflictResolver: agents.MajorityVoting,
			WeightedVoting:   true,
			AgentWeights: map[string]float64{
				"research": 1.0,
				"analysis": 1.2,
				"code":     1.0,
			},
		},
		MemoryStore:     cachedStore,
		EnableStreaming: true,
		StreamWriter: func(event agents.SupervisorEvent) {
			multiStream.Send(stream.StreamEvent{
				Type:      "supervisor_event",
				Data:      event,
				Timestamp: time.Now(),
			})
		},
	}

	// Example 1: Web Research with Tool Orchestration
	fmt.Println("=== Example 1: Web Research with Tool Orchestration ===")
	researchExample(ctx, supervisor, multiStream)

	// Example 2: Data Analysis Pipeline
	fmt.Println("\n=== Example 2: Data Analysis Pipeline ===")
	analysisExample(ctx, supervisor, multiStream)

	// Example 3: Code Generation with Review
	fmt.Println("\n=== Example 3: Code Generation with Review ===")
	codeGenExample(ctx, supervisor, multiStream)

	// Example 4: Multi-Agent Collaboration
	fmt.Println("\n=== Example 4: Multi-Agent Collaboration ===")
	collaborationExample(ctx, supervisor, multiStream)

	// Example 5: Adaptive Learning
	fmt.Println("\n=== Example 5: Adaptive Learning ===")
	adaptiveLearningExample(ctx, supervisor, adaptiveMiddleware, multiStream)

	// Cleanup
	// multiStream doesn't have Stop method - context cancellation handles this
	langGraphStore.Close(ctx)
	if closer, ok := dbQuery.(*practical.DatabaseQueryRuntimeTool); ok {
		closer.DatabaseQueryTool.Close()
	}

	fmt.Println("\n=== All examples completed successfully ===")
}

func researchExample(ctx context.Context, supervisor *agents.SupervisorAgent, stream *stream.MultiModeStream) {
	// Complex research task
	task := agents.SupervisorTask{
		ID:          "research-001",
		Description: "Research the latest developments in quantum computing and their potential impact on cryptography",
		Type:        agents.TaskTypeResearch,
		Priority:    agents.PriorityHigh,
		Requirements: []string{
			"Find recent academic papers",
			"Identify key players and companies",
			"Analyze potential risks to current encryption",
			"Provide timeline estimates",
		},
		Context: map[string]interface{}{
			"depth":      "comprehensive",
			"sources":    []string{"academic", "industry", "news"},
			"time_frame": "last 6 months",
		},
	}

	// Execute task
	result, err := supervisor.Execute(ctx, task)
	if err != nil {
		log.Printf("Research task failed: %v", err)
		return
	}

	// Display results
	fmt.Printf("Research Results:\n")
	fmt.Printf("- Task ID: %s\n", result.TaskID)
	fmt.Printf("- Status: %s\n", result.Status)
	fmt.Printf("- Agents Used: %v\n", result.AgentsUsed)
	fmt.Printf("- Duration: %v\n", result.Duration)

	// Subscribe to streaming updates
	messages := stream.Subscribe(stream.StreamModeMessages)
	go func() {
		for msg := range messages {
			fmt.Printf("Stream: %+v\n", msg)
		}
	}()
}

func analysisExample(ctx context.Context, supervisor *agents.SupervisorAgent, stream *stream.MultiModeStream) {
	// Data analysis task
	task := agents.SupervisorTask{
		ID:          "analysis-001",
		Description: "Analyze customer behavior patterns from the last quarter's sales data",
		Type:        agents.TaskTypeProcess,
		Priority:    agents.PriorityMedium,
		Requirements: []string{
			"Load sales data from database",
			"Identify purchasing patterns",
			"Segment customers",
			"Generate visualizations",
			"Create executive summary",
		},
		Context: map[string]interface{}{
			"database":    "sales_db",
			"time_period": "Q4 2024",
			"metrics": []string{
				"purchase_frequency",
				"average_order_value",
				"customer_lifetime_value",
			},
		},
		Dependencies: []string{}, // No dependencies for this task
	}

	// Execute with progress tracking
	progressChan := make(chan agents.TaskProgress, 10)
	go func() {
		for progress := range progressChan {
			fmt.Printf("Progress: %s - %.2f%% complete\n",
				progress.Stage, progress.Percentage)
		}
	}()

	result, err := supervisor.ExecuteWithProgress(ctx, task, progressChan)
	if err != nil {
		log.Printf("Analysis task failed: %v", err)
		return
	}

	fmt.Printf("Analysis completed: %+v\n", result)
}

func codeGenExample(ctx context.Context, supervisor *agents.SupervisorAgent, stream *stream.MultiModeStream) {
	// Code generation task
	task := agents.SupervisorTask{
		ID:          "code-001",
		Description: "Generate a REST API client for the weather service with retry logic and caching",
		Type:        agents.TaskTypeGenerate,
		Priority:    agents.PriorityLow,
		Requirements: []string{
			"Create client interface",
			"Implement HTTP methods",
			"Add exponential backoff retry",
			"Implement response caching",
			"Write unit tests",
			"Generate documentation",
		},
		Context: map[string]interface{}{
			"language":    "go",
			"api_spec":    "openweather_api_v2.5",
			"cache_ttl":   "5m",
			"max_retries": 3,
		},
		Metadata: map[string]interface{}{
			"reviewer":      "code",
			"test_coverage": 80,
		},
	}

	// Execute task
	result, err := supervisor.Execute(ctx, task)
	if err != nil {
		log.Printf("Code generation failed: %v", err)
		return
	}

	// Display generated code info
	fmt.Printf("Code Generation Results:\n")
	if code, ok := result.Result.(map[string]interface{})["generated_code"]; ok {
		fmt.Printf("Generated code preview: %.200s...\n", code)
	}
	if tests, ok := result.Result.(map[string]interface{})["tests"]; ok {
		fmt.Printf("Tests generated: %v\n", tests)
	}
}

func collaborationExample(ctx context.Context, supervisor *agents.SupervisorAgent, stream *stream.MultiModeStream) {
	// Multi-agent collaboration task
	task := agents.SupervisorTask{
		ID:          "collab-001",
		Description: "Create a comprehensive market analysis report with code examples and data visualizations",
		Type:        agents.TaskTypeAnalyze,
		Priority:    agents.PriorityHigh,
		Requirements: []string{
			"Research market trends",
			"Analyze competitor data",
			"Generate statistical models",
			"Create visualization code",
			"Compile final report",
		},
		Context: map[string]interface{}{
			"market":      "AI/ML tools",
			"competitors": []string{"OpenAI", "Google", "Anthropic", "Cohere"},
			"metrics": []string{
				"market_share",
				"growth_rate",
				"feature_comparison",
				"pricing_analysis",
			},
		},
		Dependencies: []string{}, // Tasks will be automatically orchestrated
	}

	// Execute with parallel agent coordination
	result, err := supervisor.Execute(ctx, task)
	if err != nil {
		log.Printf("Collaboration task failed: %v", err)
		return
	}

	// Display collaboration metrics
	fmt.Printf("Collaboration Results:\n")
	fmt.Printf("- Agents Involved: %v\n", result.AgentsUsed)
	fmt.Printf("- Inter-agent Communications: %d\n",
		result.Metadata["communications"])
	fmt.Printf("- Parallel Tasks: %d\n",
		result.Metadata["parallel_tasks"])
	fmt.Printf("- Total Execution Time: %v\n", result.Duration)
}

func adaptiveLearningExample(ctx context.Context, supervisor *agents.SupervisorAgent, adaptive *middleware.AdaptiveMiddleware, stream *stream.MultiModeStream) {
	// Series of tasks to demonstrate adaptive learning
	tasks := []agents.SupervisorTask{
		{
			ID:          "adaptive-001",
			Description: "Simple web search task",
			Type:        agents.TaskTypeResearch,
			Priority:    agents.PriorityLow,
		},
		{
			ID:          "adaptive-002",
			Description: "Complex data analysis with multiple sources",
			Type:        agents.TaskTypeAnalyze,
			Priority:    agents.PriorityMedium,
		},
		{
			ID:          "adaptive-003",
			Description: "Code refactoring with performance optimization",
			Type:        agents.TaskTypeProcess,
			Priority:    agents.PriorityHigh,
		},
	}

	// Execute tasks and let the system learn
	for _, task := range tasks {
		startTime := time.Now()

		// Apply adaptive middleware
		adaptedTask := adaptive.Process(ctx, task)

		// Execute task
		result, err := supervisor.Execute(ctx, adaptedTask)

		// Record performance metrics
		metrics := middleware.PerformanceMetrics{
			TaskID:        task.ID,
			ExecutionTime: time.Since(startTime),
			Success:       err == nil,
			ErrorRate:     0.0,
		}

		if err != nil {
			metrics.ErrorRate = 1.0
			log.Printf("Task %s failed: %v", task.ID, err)
		} else {
			fmt.Printf("Task %s completed in %v\n", task.ID, metrics.ExecutionTime)
		}

		// Update adaptive learning
		adaptive.UpdateMetrics(task.ID, metrics)

		// Learn from results
		if err == nil {
			adaptive.Learn(ctx, task, result.Result)
		}
	}

	// Display learning statistics
	stats := adaptive.GetStatistics()
	fmt.Printf("\nAdaptive Learning Statistics:\n")
	fmt.Printf("- Tasks Processed: %d\n", stats["tasks_processed"])
	fmt.Printf("- Average Success Rate: %.2f%%\n", stats["success_rate"])
	fmt.Printf("- Performance Improvement: %.2f%%\n", stats["performance_improvement"])
	fmt.Printf("- Adaptations Made: %d\n", stats["adaptations_made"])
}
