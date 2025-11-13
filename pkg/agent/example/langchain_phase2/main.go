package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kart-io/k8s-agent/pkg/agent/core"
	"github.com/kart-io/k8s-agent/pkg/agent/store/memory"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║  LangChain-Inspired Phase 2: Middleware System            ║")
	fmt.Println("║  Demonstrating Request/Response Interceptors              ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 1. Basic Middleware Chain
	fmt.Println("1. Basic Middleware Chain")
	fmt.Println("-------------------------")
	basicMiddlewareDemo()

	// 2. Dynamic Prompt Middleware
	fmt.Println("\n2. Dynamic Prompt Middleware")
	fmt.Println("----------------------------")
	dynamicPromptDemo()

	// 3. Tool Selector Middleware
	fmt.Println("\n3. Tool Selector Middleware")
	fmt.Println("---------------------------")
	toolSelectorDemo()

	// 4. Rate Limiter Middleware
	fmt.Println("\n4. Rate Limiter Middleware")
	fmt.Println("--------------------------")
	rateLimiterDemo()

	// 5. Complete Agent with Middleware Stack
	fmt.Println("\n5. Complete Agent with Middleware Stack")
	fmt.Println("----------------------------------------")
	completeAgentDemo()

	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║  All Middleware Examples Completed Successfully! 🎉        ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
}

// basicMiddlewareDemo demonstrates basic middleware chain execution
func basicMiddlewareDemo() {
	// Create main handler (simulates LLM call)
	handler := func(ctx context.Context, req *core.MiddlewareRequest) (*core.MiddlewareResponse, error) {
		fmt.Printf("  [Handler] Processing: %v\n", req.Input)
		time.Sleep(100 * time.Millisecond) // Simulate processing

		return &core.MiddlewareResponse{
			Output:   fmt.Sprintf("Processed: %v", req.Input),
			State:    req.State,
			Metadata: req.Metadata,
		}, nil
	}

	// Create middleware chain
	chain := core.NewMiddlewareChain(handler)

	// Add logging middleware
	chain.Use(core.NewLoggingMiddleware(func(msg string) {
		fmt.Printf("  [LOG] %s\n", msg)
	}))

	// Add timing middleware
	chain.Use(core.NewTimingMiddleware())

	// Add cache middleware
	chain.Use(core.NewCacheMiddleware(5 * time.Second))

	// Execute request
	state := core.NewAgentState()
	state.Set("user", "Alice")

	request := &core.MiddlewareRequest{
		Input:     "What is Kubernetes?",
		State:     state,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	fmt.Println("  First request (cache miss):")
	resp1, err := chain.Execute(context.Background(), request)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  Response: %v\n", resp1.Output)
	fmt.Printf("  Duration: %v\n", resp1.Duration)

	fmt.Println("\n  Second request (cache hit):")
	resp2, err := chain.Execute(context.Background(), request)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  Response: %v\n", resp2.Output)
	fmt.Printf("  Duration: %v (cached)\n", resp2.Duration)
}

// dynamicPromptDemo demonstrates dynamic prompt modification
func dynamicPromptDemo() {
	// Create prompt modifier function
	promptModifier := func(req *core.MiddlewareRequest) string {
		if req.State != nil {
			if role, ok := req.State.Get("role"); ok {
				original := fmt.Sprintf("%v", req.Input)
				return fmt.Sprintf("You are a %s assistant. %s", role, original)
			}
		}
		return fmt.Sprintf("%v", req.Input)
	}

	// Create dynamic prompt middleware
	dynamicPrompt := core.NewDynamicPromptMiddleware(promptModifier)

	// Create handler that echoes the modified prompt
	handler := func(ctx context.Context, req *core.MiddlewareRequest) (*core.MiddlewareResponse, error) {
		return &core.MiddlewareResponse{
			Output: fmt.Sprintf("Received prompt: %v", req.Input),
		}, nil
	}

	chain := core.NewMiddlewareChain(handler)
	chain.Use(dynamicPrompt)

	// Test with different roles
	for _, role := range []string{"technical", "creative", "analytical"} {
		state := core.NewAgentState()
		state.Set("role", role)
		state.Set("user_name", "Bob")

		request := &core.MiddlewareRequest{
			Input: "Explain cloud computing",
			State: state,
		}

		resp, _ := chain.Execute(context.Background(), request)
		fmt.Printf("  Role: %s\n", role)
		fmt.Printf("  %v\n\n", resp.Output)
	}
}

// toolSelectorDemo demonstrates tool selection middleware
func toolSelectorDemo() {
	// Available tools
	tools := []string{
		"calculator",
		"search",
		"database",
		"file_reader",
		"file_writer",
		"web_scraper",
		"translator",
		"code_runner",
	}

	// Create tool selector middleware
	toolSelector := core.NewToolSelectorMiddleware(tools, 3)

	// Handler that shows selected tools
	handler := func(ctx context.Context, req *core.MiddlewareRequest) (*core.MiddlewareResponse, error) {
		selectedTools := req.Metadata["selected_tools"]
		return &core.MiddlewareResponse{
			Output: fmt.Sprintf("Selected tools: %v", selectedTools),
		}, nil
	}

	chain := core.NewMiddlewareChain(handler)
	chain.Use(toolSelector)

	// Test different queries
	queries := []string{
		"Calculate 2 + 2",
		"Search for Python tutorials",
		"Read the config file",
		"Translate this to Spanish",
	}

	for _, query := range queries {
		request := &core.MiddlewareRequest{
			Input:    query,
			Metadata: make(map[string]interface{}),
		}

		resp, _ := chain.Execute(context.Background(), request)
		fmt.Printf("  Query: %s\n", query)
		fmt.Printf("  %v\n\n", resp.Output)
	}
}

// rateLimiterDemo demonstrates rate limiting
func rateLimiterDemo() {
	// Create rate limiter (3 requests per 5 seconds)
	rateLimiter := core.NewRateLimiterMiddleware(3, 5*time.Second)

	// Simple handler
	handler := func(ctx context.Context, req *core.MiddlewareRequest) (*core.MiddlewareResponse, error) {
		return &core.MiddlewareResponse{
			Output: fmt.Sprintf("Request %v processed", req.Input),
		}, nil
	}

	chain := core.NewMiddlewareChain(handler)
	chain.Use(rateLimiter)

	// Simulate multiple requests
	state := core.NewAgentState()
	state.Set("user_id", "user123")

	for i := 1; i <= 5; i++ {
		request := &core.MiddlewareRequest{
			Input:    fmt.Sprintf("Request %d", i),
			State:    state,
			Metadata: make(map[string]interface{}),
		}

		resp, err := chain.Execute(context.Background(), request)
		if err != nil {
			fmt.Printf("  Request %d: BLOCKED - %v\n", i, err)
		} else {
			remaining := request.Metadata["rate_limit_remaining"]
			fmt.Printf("  Request %d: SUCCESS - %v (remaining: %v)\n", i, resp.Output, remaining)
		}
	}
}

// completeAgentDemo demonstrates a complete agent with full middleware stack
func completeAgentDemo() {
	fmt.Println("  Building AI Agent with Complete Middleware Stack...")
	fmt.Println()

	// Initialize Phase 1 components
	store := memory.New()
	checkpointer := core.NewInMemorySaver()
	ctx := context.Background()

	// Create agent state
	state := core.NewAgentState()
	state.Set("session_id", "agent-session-001")
	state.Set("user_name", "Charlie")
	state.Set("user_role", "developer")
	state.Set("conversation_count", 0)

	// Main agent handler (simulates LLM with context)
	agentHandler := func(ctx context.Context, req *core.MiddlewareRequest) (*core.MiddlewareResponse, error) {
		// Increment conversation count
		if req.State != nil {
			count, _ := req.State.Get("conversation_count")
			if c, ok := count.(int); ok {
				req.State.Set("conversation_count", c+1)
			}
		}

		// Process with selected tools
		tools := "none"
		if req.Metadata != nil {
			if selected, ok := req.Metadata["selected_tools"]; ok {
				tools = fmt.Sprintf("%v", selected)
			}
		}

		output := fmt.Sprintf("Processing '%v' with tools: %s", req.Input, tools)

		// Store in long-term memory
		store.Put(ctx, []string{"conversations", "agent-session-001"},
			fmt.Sprintf("turn-%d", time.Now().Unix()),
			map[string]interface{}{
				"input":  req.Input,
				"output": output,
				"tools":  tools,
			})

		return &core.MiddlewareResponse{
			Output:   output,
			State:    req.State,
			Metadata: req.Metadata,
		}, nil
	}

	// Build middleware stack
	chain := core.NewMiddlewareChain(agentHandler)

	// 1. Authentication middleware
	authMiddleware := core.NewAuthenticationMiddleware(func(ctx context.Context, req *core.MiddlewareRequest) (bool, error) {
		// Check if user is authenticated
		if req.State != nil {
			if userName, ok := req.State.Get("user_name"); ok && userName != "" {
				return true, nil
			}
		}
		return false, nil
	})

	// 2. Validation middleware
	validationMiddleware := core.NewValidationMiddleware(
		func(req *core.MiddlewareRequest) error {
			// Validate input is not empty
			if req.Input == nil || fmt.Sprintf("%v", req.Input) == "" {
				return fmt.Errorf("input cannot be empty")
			}
			return nil
		},
		func(req *core.MiddlewareRequest) error {
			// Validate input length
			if len(fmt.Sprintf("%v", req.Input)) > 1000 {
				return fmt.Errorf("input too long (max 1000 characters)")
			}
			return nil
		},
	)

	// 3. Dynamic prompt middleware
	dynamicPrompt := core.NewDynamicPromptMiddleware(func(req *core.MiddlewareRequest) string {
		if req.State != nil {
			if role, ok := req.State.Get("user_role"); ok {
				return fmt.Sprintf("[Role: %v] %v", role, req.Input)
			}
		}
		return fmt.Sprintf("%v", req.Input)
	})

	// 4. Tool selector middleware
	toolSelector := core.NewToolSelectorMiddleware([]string{
		"code_runner", "database", "file_reader", "search",
	}, 2)

	// 5. Transform middleware (uppercase for demonstration)
	transformMiddleware := core.NewTransformMiddleware(
		func(input interface{}) (interface{}, error) {
			// Transform input to uppercase
			return strings.ToUpper(fmt.Sprintf("%v", input)), nil
		},
		func(output interface{}) (interface{}, error) {
			// Add timestamp to output
			return fmt.Sprintf("%v [%s]", output, time.Now().Format("15:04:05")), nil
		},
	)

	// 6. Rate limiter
	rateLimiter := core.NewRateLimiterMiddleware(10, 1*time.Minute)

	// 7. Circuit breaker
	circuitBreaker := core.NewCircuitBreakerMiddleware(3, 30*time.Second)

	// 8. Logging
	logger := core.NewLoggingMiddleware(func(msg string) {
		fmt.Printf("    [LOG] %s\n", msg)
	})

	// 9. Timing
	timing := core.NewTimingMiddleware()

	// 10. Cache
	cache := core.NewCacheMiddleware(30 * time.Second)

	// Add all middleware to chain (order matters!)
	chain.Use(
		logger,               // Log all requests/responses
		timing,               // Track timing
		authMiddleware,       // Check authentication first
		rateLimiter,          // Apply rate limits
		circuitBreaker,       // Circuit breaker protection
		validationMiddleware, // Validate input
		cache,                // Check cache
		dynamicPrompt,        // Modify prompt
		toolSelector,         // Select tools
		transformMiddleware,  // Transform input/output
	)

	// Execute agent requests
	fmt.Println("  Executing Agent Requests:")
	fmt.Println()

	requests := []string{
		"Write a Python function",
		"Query the user database",
		"Search for documentation",
	}

	for i, input := range requests {
		fmt.Printf("  Request %d: %s\n", i+1, input)

		request := &core.MiddlewareRequest{
			Input:     input,
			State:     state,
			Metadata:  make(map[string]interface{}),
			Timestamp: time.Now(),
		}

		resp, err := chain.Execute(ctx, request)
		if err != nil {
			fmt.Printf("    Error: %v\n", err)
			continue
		}

		fmt.Printf("    Output: %v\n", resp.Output)
		fmt.Printf("    Duration: %v\n", resp.Duration)

		// Save checkpoint after each turn
		checkpointer.Save(ctx, "agent-session-001", state)

		fmt.Println()
		time.Sleep(100 * time.Millisecond) // Small delay between requests
	}

	// Show final statistics
	fmt.Println("  Final Statistics:")
	count, _ := state.Get("conversation_count")
	fmt.Printf("    - Conversation turns: %v\n", count)

	// Get average latency from timing middleware
	avgLatency := timing.GetAverageLatency()
	fmt.Printf("    - Average latency: %v\n", avgLatency)

	// Check cache size
	fmt.Printf("    - Cache size: %d entries\n", cache.Size())

	// Check stored conversations
	keys, _ := store.List(ctx, []string{"conversations", "agent-session-001"})
	fmt.Printf("    - Stored conversations: %d\n", len(keys))

	// Show checkpointed state
	loadedState, _ := checkpointer.Load(ctx, "agent-session-001")
	if loadedState != nil {
		userName, _ := loadedState.Get("user_name")
		fmt.Printf("    - Checkpointed user: %v\n", userName)
	}
}
