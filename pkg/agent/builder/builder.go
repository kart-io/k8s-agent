package builder

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/kart-io/k8s-agent/pkg/agent/core"
	"github.com/kart-io/k8s-agent/pkg/agent/llm"
	"github.com/kart-io/k8s-agent/pkg/agent/tools"
)

// AgentBuilder provides a fluent API for building agents with all features
//
// Inspired by LangChain's create_agent function, it integrates:
//   - LLM client configuration
//   - Tools registration
//   - State management
//   - Runtime context
//   - Store and Checkpointer
//   - Middleware stack
//   - System prompts
type AgentBuilder[C any, S core.State] struct {
	// Core components
	llmClient    llm.Client
	tools        []tools.Tool
	systemPrompt string

	// Phase 1 components
	state        S
	store        core.Store
	checkpointer core.Checkpointer
	context      C

	// Phase 2 components
	middlewares []core.Middleware

	// Configuration
	config *AgentConfig

	// Callbacks
	callbacks []core.Callback

	// Error handling
	errorHandler func(error) error

	// Metadata
	metadata map[string]interface{}
}

// AgentConfig holds agent configuration options
type AgentConfig struct {
	// MaxIterations limits the number of reasoning steps
	MaxIterations int

	// Timeout for agent execution
	Timeout time.Duration

	// EnableStreaming enables streaming responses
	EnableStreaming bool

	// EnableAutoSave automatically saves state after each step
	EnableAutoSave bool

	// SaveInterval for auto-save
	SaveInterval time.Duration

	// MaxTokens limits LLM response tokens
	MaxTokens int

	// Temperature for LLM sampling
	Temperature float64

	// SessionID for checkpointing
	SessionID string

	// Verbose enables detailed logging
	Verbose bool
}

// DefaultAgentConfig returns default configuration
func DefaultAgentConfig() *AgentConfig {
	return &AgentConfig{
		MaxIterations:   10,
		Timeout:         5 * time.Minute,
		EnableStreaming: false,
		EnableAutoSave:  true,
		SaveInterval:    30 * time.Second,
		MaxTokens:       2000,
		Temperature:     0.7,
		SessionID:       fmt.Sprintf("session-%d", time.Now().Unix()),
		Verbose:         false,
	}
}

// NewAgentBuilder creates a new agent builder
func NewAgentBuilder[C any, S core.State](llmClient llm.Client) *AgentBuilder[C, S] {
	return &AgentBuilder[C, S]{
		llmClient:   llmClient,
		tools:       []tools.Tool{},
		middlewares: []core.Middleware{},
		callbacks:   []core.Callback{},
		config:      DefaultAgentConfig(),
		metadata:    make(map[string]interface{}),
	}
}

// WithTools adds tools to the agent
func (b *AgentBuilder[C, S]) WithTools(tools ...tools.Tool) *AgentBuilder[C, S] {
	b.tools = append(b.tools, tools...)
	return b
}

// WithSystemPrompt sets the system prompt
func (b *AgentBuilder[C, S]) WithSystemPrompt(prompt string) *AgentBuilder[C, S] {
	b.systemPrompt = prompt
	return b
}

// WithState sets the agent state
func (b *AgentBuilder[C, S]) WithState(state S) *AgentBuilder[C, S] {
	b.state = state
	return b
}

// WithContext sets the application context
func (b *AgentBuilder[C, S]) WithContext(context C) *AgentBuilder[C, S] {
	b.context = context
	return b
}

// WithStore sets the long-term storage
func (b *AgentBuilder[C, S]) WithStore(store core.Store) *AgentBuilder[C, S] {
	b.store = store
	return b
}

// WithCheckpointer sets the session checkpointer
func (b *AgentBuilder[C, S]) WithCheckpointer(checkpointer core.Checkpointer) *AgentBuilder[C, S] {
	b.checkpointer = checkpointer
	return b
}

// WithMiddleware adds middleware to the chain
func (b *AgentBuilder[C, S]) WithMiddleware(middleware ...core.Middleware) *AgentBuilder[C, S] {
	b.middlewares = append(b.middlewares, middleware...)
	return b
}

// WithCallbacks adds callbacks for monitoring
func (b *AgentBuilder[C, S]) WithCallbacks(callbacks ...core.Callback) *AgentBuilder[C, S] {
	b.callbacks = append(b.callbacks, callbacks...)
	return b
}

// WithConfig sets custom configuration
func (b *AgentBuilder[C, S]) WithConfig(config *AgentConfig) *AgentBuilder[C, S] {
	b.config = config
	return b
}

// WithErrorHandler sets custom error handling
func (b *AgentBuilder[C, S]) WithErrorHandler(handler func(error) error) *AgentBuilder[C, S] {
	b.errorHandler = handler
	return b
}

// WithMetadata adds metadata to the agent
func (b *AgentBuilder[C, S]) WithMetadata(key string, value interface{}) *AgentBuilder[C, S] {
	b.metadata[key] = value
	return b
}

// ConfigureForRAG adds common RAG (Retrieval-Augmented Generation) components
func (b *AgentBuilder[C, S]) ConfigureForRAG() *AgentBuilder[C, S] {
	// Add common RAG middleware
	b.WithMiddleware(
		core.NewCacheMiddleware(5*time.Minute),
		core.NewDynamicPromptMiddleware(func(req *core.MiddlewareRequest) string {
			// Add context from retrieval
			return fmt.Sprintf("Use the following context to answer: %v", req.Input)
		}),
	)

	// Set appropriate config
	b.config.MaxTokens = 3000
	b.config.Temperature = 0.3 // Lower temperature for factual responses

	return b
}

// ConfigureForChatbot adds common chatbot components
func (b *AgentBuilder[C, S]) ConfigureForChatbot() *AgentBuilder[C, S] {
	// Add chatbot middleware
	b.WithMiddleware(
		core.NewRateLimiterMiddleware(20, time.Minute),
		core.NewValidationMiddleware(
			func(req *core.MiddlewareRequest) error {
				// Validate input length
				if len(fmt.Sprintf("%v", req.Input)) > 1000 {
					return fmt.Errorf("message too long")
				}
				return nil
			},
		),
	)

	// Enable streaming for better UX
	b.config.EnableStreaming = true
	b.config.Temperature = 0.8 // Higher temperature for creativity

	return b
}

// ConfigureForAnalysis adds components for data analysis tasks
func (b *AgentBuilder[C, S]) ConfigureForAnalysis() *AgentBuilder[C, S] {
	// Add analysis middleware
	b.WithMiddleware(
		core.NewTimingMiddleware(),
		core.NewTransformMiddleware(
			nil, // No input transform
			func(output interface{}) (interface{}, error) {
				// Format output as structured data
				return map[string]interface{}{
					"analysis":  output,
					"timestamp": time.Now(),
				}, nil
			},
		),
	)

	// Configure for accuracy
	b.config.Temperature = 0.1  // Very low temperature for consistency
	b.config.MaxIterations = 20 // More iterations for complex analysis

	return b
}

// Build constructs the final agent
func (b *AgentBuilder[C, S]) Build() (*ConfigurableAgent[C, S], error) {
	// Validate required components
	if b.llmClient == nil {
		return nil, fmt.Errorf("LLM client is required")
	}

	// Set defaults if not provided
	// Check if state is zero value
	var zero S
	if reflect.DeepEqual(b.state, zero) {
		// Try to create a default state if S is *core.AgentState
		if _, ok := any(zero).(*core.AgentState); ok {
			b.state = any(core.NewAgentState()).(S)
		} else {
			return nil, fmt.Errorf("state is required")
		}
	}

	if b.store == nil {
		b.store = core.NewInMemoryStore()
	}

	if b.checkpointer == nil {
		b.checkpointer = core.NewInMemorySaver()
	}

	// Create runtime
	runtime := core.NewRuntime(
		b.context,
		b.state,
		b.store,
		b.checkpointer,
		b.config.SessionID,
	)

	// Build middleware chain
	handler := b.createHandler(runtime)
	chain := core.NewMiddlewareChain(handler)

	// Add default middleware if verbose
	if b.config.Verbose {
		chain.Use(core.NewLoggingMiddleware(nil))
		chain.Use(core.NewTimingMiddleware())
	}

	// Add user-specified middleware
	chain.Use(b.middlewares...)

	// Create the agent
	agent := &ConfigurableAgent[C, S]{
		llmClient:    b.llmClient,
		tools:        b.tools,
		systemPrompt: b.systemPrompt,
		runtime:      runtime,
		chain:        chain,
		config:       b.config,
		callbacks:    b.callbacks,
		errorHandler: b.errorHandler,
		metadata:     b.metadata,
	}

	// Initialize if needed
	if err := agent.Initialize(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize agent: %w", err)
	}

	return agent, nil
}

// createHandler creates the main execution handler
func (b *AgentBuilder[C, S]) createHandler(runtime *core.Runtime[C, S]) core.Handler {
	return func(ctx context.Context, request *core.MiddlewareRequest) (*core.MiddlewareResponse, error) {
		// Extract input
		input := fmt.Sprintf("%v", request.Input)

		// Add system prompt if available
		if b.systemPrompt != "" {
			input = fmt.Sprintf("%s\n\n%s", b.systemPrompt, input)
		}

		// Create LLM request
		llmReq := &llm.CompletionRequest{
			Messages: []llm.Message{
				{
					Role:    "system",
					Content: b.systemPrompt,
				},
				{
					Role:    "user",
					Content: fmt.Sprintf("%v", request.Input),
				},
			},
			MaxTokens:   b.config.MaxTokens,
			Temperature: b.config.Temperature,
		}

		// Call LLM
		response, err := b.llmClient.Complete(ctx, llmReq)
		if err != nil {
			return nil, fmt.Errorf("LLM error: %w", err)
		}

		// Update state if needed
		if request.State != nil {
			request.State.Set("last_response", response.Content)
			request.State.Set("last_timestamp", time.Now())
		}

		// Save checkpoint if auto-save is enabled
		if b.config.EnableAutoSave && runtime.Checkpointer != nil {
			runtime.SaveState(ctx)
		}

		// Create response
		return &core.MiddlewareResponse{
			Output:   response.Content,
			State:    request.State,
			Metadata: request.Metadata,
		}, nil
	}
}

// ConfigurableAgent is the built agent with full configuration
type ConfigurableAgent[C any, S core.State] struct {
	llmClient    llm.Client
	tools        []tools.Tool
	systemPrompt string
	runtime      *core.Runtime[C, S]
	chain        *core.MiddlewareChain
	config       *AgentConfig
	callbacks    []core.Callback
	errorHandler func(error) error
	metadata     map[string]interface{}
	mu           sync.RWMutex
}

// Initialize prepares the agent for execution
func (a *ConfigurableAgent[C, S]) Initialize(ctx context.Context) error {
	// Load previous state if exists
	if a.runtime.Checkpointer != nil {
		if exists, _ := a.runtime.Checkpointer.Exists(ctx, a.config.SessionID); exists {
			state, err := a.runtime.Checkpointer.Load(ctx, a.config.SessionID)
			if err == nil {
				// Update runtime state
				a.runtime.State = state.(S)
			}
		}
	}

	// Notify callbacks
	for _, cb := range a.callbacks {
		if err := cb.OnStart(ctx, a.metadata); err != nil {
			return err
		}
	}

	return nil
}

// Execute runs the agent with the given input
func (a *ConfigurableAgent[C, S]) Execute(ctx context.Context, input interface{}) (*AgentOutput, error) {
	// Apply timeout if configured
	if a.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, a.config.Timeout)
		defer cancel()
	}

	// Create request
	request := &core.MiddlewareRequest{
		Input:     input,
		State:     a.runtime.State,
		Runtime:   a.runtime,
		Metadata:  make(map[string]interface{}),
		Headers:   make(map[string]string),
		Timestamp: time.Now(),
	}

	// Add metadata
	for k, v := range a.metadata {
		request.Metadata[k] = v
	}

	// Execute through middleware chain
	response, err := a.chain.Execute(ctx, request)
	if err != nil {
		// Handle error
		if a.errorHandler != nil {
			err = a.errorHandler(err)
		}

		// Notify callbacks
		for _, cb := range a.callbacks {
			cb.OnError(ctx, err)
		}

		return nil, err
	}

	// Create output
	output := &AgentOutput{
		Result:    response.Output,
		State:     response.State,
		Metadata:  response.Metadata,
		Duration:  response.Duration,
		Timestamp: time.Now(),
	}

	// Notify callbacks
	for _, cb := range a.callbacks {
		if err := cb.OnEnd(ctx, output); err != nil {
			return output, err
		}
	}

	return output, nil
}

// ExecuteWithTools runs the agent with tool execution capability
func (a *ConfigurableAgent[C, S]) ExecuteWithTools(ctx context.Context, input interface{}) (*AgentOutput, error) {
	iterations := 0
	var lastOutput *AgentOutput

	for iterations < a.config.MaxIterations {
		// Execute one step
		output, err := a.Execute(ctx, input)
		if err != nil {
			return nil, err
		}

		lastOutput = output

		// Check if we need to use tools
		toolCalls := a.extractToolCalls(output.Result)
		if len(toolCalls) == 0 {
			// No tools needed, return result
			return output, nil
		}

		// Execute tools
		toolResults := make([]interface{}, 0, len(toolCalls))
		for _, call := range toolCalls {
			result, err := a.executeToolCall(ctx, call)
			if err != nil {
				return nil, fmt.Errorf("tool execution failed: %w", err)
			}
			toolResults = append(toolResults, result)
		}

		// Update input with tool results for next iteration
		input = map[string]interface{}{
			"previous_output": output.Result,
			"tool_results":    toolResults,
		}

		iterations++
	}

	// Max iterations reached
	return lastOutput, fmt.Errorf("max iterations (%d) reached", a.config.MaxIterations)
}

// extractToolCalls extracts tool calls from LLM output
func (a *ConfigurableAgent[C, S]) extractToolCalls(output interface{}) []ToolCall {
	// Simplified tool call extraction
	// In production, use proper parsing
	return []ToolCall{}
}

// executeToolCall executes a single tool call
func (a *ConfigurableAgent[C, S]) executeToolCall(ctx context.Context, call ToolCall) (interface{}, error) {
	// Find tool
	for _, tool := range a.tools {
		if tool.Name() == call.Name {
			// Create tool input
			toolInput := &tools.ToolInput{
				Args:    call.Input,
				Context: ctx,
			}

			// Execute tool
			output, err := tool.Invoke(ctx, toolInput)
			if err != nil {
				return nil, err
			}
			return output.Result, nil
		}
	}
	return nil, fmt.Errorf("tool not found: %s", call.Name)
}

// GetState returns the current state
func (a *ConfigurableAgent[C, S]) GetState() S {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.runtime.State
}

// GetMetrics returns agent metrics
func (a *ConfigurableAgent[C, S]) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Add basic metrics
	metrics["session_id"] = a.config.SessionID
	metrics["tools_count"] = len(a.tools)
	// Note: middleware count not exposed by MiddlewareChain

	// Add state size if available
	if state, ok := any(a.runtime.State).(*core.AgentState); ok {
		metrics["state_size"] = state.Size()
	}

	return metrics
}

// Shutdown gracefully shuts down the agent
func (a *ConfigurableAgent[C, S]) Shutdown(ctx context.Context) error {
	// Save final state
	if a.runtime.Checkpointer != nil {
		if err := a.runtime.SaveState(ctx); err != nil {
			return fmt.Errorf("failed to save final state: %w", err)
		}
	}

	// Notify callbacks
	for _, cb := range a.callbacks {
		if shutdown, ok := cb.(interface{ OnShutdown(context.Context) error }); ok {
			if err := shutdown.OnShutdown(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

// ToolCall represents a tool invocation request
type ToolCall struct {
	Name  string
	Input map[string]interface{}
}

// AgentOutput represents the agent execution result
type AgentOutput struct {
	Result    interface{}
	State     core.State
	Metadata  map[string]interface{}
	Duration  time.Duration
	Timestamp time.Time
}

// QuickAgent creates a simple agent with minimal configuration
func QuickAgent(llmClient llm.Client, systemPrompt string) (*ConfigurableAgent[any, *core.AgentState], error) {
	return NewAgentBuilder[any, *core.AgentState](llmClient).
		WithSystemPrompt(systemPrompt).
		WithState(core.NewAgentState()).
		Build()
}

// RAGAgent creates a pre-configured RAG agent
func RAGAgent(llmClient llm.Client, retriever interface{}) (*ConfigurableAgent[any, *core.AgentState], error) {
	return NewAgentBuilder[any, *core.AgentState](llmClient).
		WithSystemPrompt("You are a helpful assistant. Answer questions based on the provided context.").
		ConfigureForRAG().
		WithState(core.NewAgentState()).
		WithMetadata("type", "rag").
		Build()
}

// ChatAgent creates a pre-configured chatbot agent
func ChatAgent(llmClient llm.Client, userName string) (*ConfigurableAgent[any, *core.AgentState], error) {
	state := core.NewAgentState()
	state.Set("user_name", userName)

	return NewAgentBuilder[any, *core.AgentState](llmClient).
		WithSystemPrompt(fmt.Sprintf("You are a helpful assistant chatting with %s.", userName)).
		WithState(state).
		ConfigureForChatbot().
		WithMetadata("type", "chatbot").
		Build()
}
