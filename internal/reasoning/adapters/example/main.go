package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kart-io/k8s-agent/internal/reasoning/adapters"
	"github.com/kart-io/k8s-agent/internal/reasoning/agents/k8s_tool"
	"github.com/kart-io/k8s-agent/internal/reasoning/agents/reasoning"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/description"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/root_cause"
	"github.com/kart-io/k8s-agent/internal/reasoning/llm/proxy"
	"github.com/kart-io/k8s-agent/internal/reasoning/memory"
	"github.com/kart-io/k8s-agent/internal/reasoning/orchestrator"
	"github.com/kart-io/k8s-agent/pkg/agent/core"
)

// This example demonstrates how to use the adapters to integrate
// the reasoning service with the pkg/agent framework.

func main() {
	ctx := context.Background()

	// Initialize components
	components, err := initializeComponents()
	if err != nil {
		log.Fatalf("Failed to initialize components: %v", err)
	}

	// Example 1: Using ReasoningAgentAdapter directly
	fmt.Println("=== Example 1: Using ReasoningAgentAdapter ===")
	if err := exampleReasoningAgent(ctx, components); err != nil {
		log.Printf("Example 1 failed: %v", err)
	}
	fmt.Println()

	// Example 2: Using OrchestratorAdapter for full workflow
	fmt.Println("=== Example 2: Using OrchestratorAdapter ===")
	if err := exampleOrchestrator(ctx, components); err != nil {
		log.Printf("Example 2 failed: %v", err)
	}
	fmt.Println()

	// Example 3: Using individual chain adapters
	fmt.Println("=== Example 3: Using RootCauseChainAdapter ===")
	if err := exampleRootCauseChain(ctx, components); err != nil {
		log.Printf("Example 3 failed: %v", err)
	}
}

// Components holds all initialized reasoning service components
type Components struct {
	ReasoningAgent   *reasoning.ReasoningAgent
	RootCauseChain   *root_cause.RootCauseChain
	DescriptionChain *description.DescriptionChain
	K8sTool          *k8s_tool.K8sTool
	MemoryManager    memory.Manager
	LLMProxy         *proxy.ProxyAdapter
}

// initializeComponents creates all necessary components for the reasoning service
func initializeComponents() (*Components, error) {
	// Create LLM proxy (mock for demonstration)
	llmProxy := &proxy.ProxyAdapter{}

	// Create K8s tool
	k8sToolConfig := k8s_tool.DefaultToolConfig()
	k8sTool, err := k8s_tool.NewK8sTool(k8sToolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create K8s tool: %w", err)
	}

	// Create chains
	rootCauseConfig := root_cause.DefaultChainConfig()
	rootCauseChain, err := root_cause.NewRootCauseChain(llmProxy, rootCauseConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create root cause chain: %w", err)
	}

	descriptionConfig := description.DefaultChainConfig()
	descriptionChain, err := description.NewDescriptionChain(llmProxy, descriptionConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create description chain: %w", err)
	}

	// Create reasoning agent
	agentConfig := reasoning.DefaultAgentConfig()
	reasoningAgent, err := reasoning.NewReasoningAgent(
		rootCauseChain,
		descriptionChain,
		k8sTool,
		agentConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create reasoning agent: %w", err)
	}

	// Create memory manager
	memoryConfig := &memory.ManagerConfig{
		EnableConversation:    true,
		MaxConversationLength: 10,
		EnableVectorStore:     true,
		VectorStoreType:       "memory",
		EmbeddingDimension:    768,
		DefaultSearchLimit:    5,
		SimilarityThreshold:   0.7,
	}
	memoryManager, err := memory.NewManager(memoryConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory manager: %w", err)
	}

	return &Components{
		ReasoningAgent:   reasoningAgent,
		RootCauseChain:   rootCauseChain,
		DescriptionChain: descriptionChain,
		K8sTool:          k8sTool,
		MemoryManager:    memoryManager,
		LLMProxy:         llmProxy,
	}, nil
}

// exampleReasoningAgent demonstrates using the ReasoningAgentAdapter
func exampleReasoningAgent(ctx context.Context, components *Components) error {
	// Create adapter
	adapter := adapters.NewReasoningAgentAdapter(components.ReasoningAgent)

	fmt.Println("Agent Name:", adapter.Name())
	fmt.Println("Agent Description:", adapter.Description())
	fmt.Println("Agent Capabilities:", adapter.Capabilities())
	fmt.Println()

	// Prepare input using framework types
	input := &core.AgentInput{
		Task:        "Analyze Kubernetes pod failure",
		Instruction: "Identify root cause and generate description",
		Context: map[string]interface{}{
			"failure_type":  "CrashLoopBackOff",
			"resource_type": "pod",
			"resource_name": "my-app-pod-xyz123",
			"namespace":     "production",
			"cluster_id":    "cluster-01",
			"error_message": "Error: OOMKilled - container exceeded memory limit",
			"fetch_logs":    true,
			"fetch_events":  true,
			"fetch_metrics": true,
			"language":      "en",
			"detail_level":  "normal",
		},
		Options: core.AgentOptions{
			Temperature:  0.7,
			MaxTokens:    4096,
			EnableTools:  true,
			EnableMemory: false,
			Timeout:      60 * time.Second,
		},
		SessionID: "session-123",
		Timestamp: time.Now(),
	}

	// Execute through framework interface
	output, err := adapter.Execute(ctx, input)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	// Print results
	fmt.Printf("Status: %s\n", output.Status)
	fmt.Printf("Message: %s\n", output.Message)
	fmt.Printf("Latency: %v\n", output.Latency)
	fmt.Printf("Reasoning Steps: %d\n", len(output.ReasoningSteps))
	fmt.Println()

	// Print reasoning steps
	for _, step := range output.ReasoningSteps {
		fmt.Printf("  Step %d: %s - %s\n", step.Step, step.Action, step.Description)
		if step.Error != "" {
			fmt.Printf("    Error: %s\n", step.Error)
		}
	}
	fmt.Println()

	// Print metadata
	fmt.Println("Metadata:")
	for key, value := range output.Metadata {
		fmt.Printf("  %s: %v\n", key, value)
	}

	return nil
}

// exampleOrchestrator demonstrates using the OrchestratorAdapter
func exampleOrchestrator(ctx context.Context, components *Components) error {
	// Create orchestrator config
	config := orchestrator.DefaultOrchestratorConfig()

	// Create orchestrator adapter
	adapter, err := adapters.NewOrchestratorAdapter(
		components.ReasoningAgent,
		components.RootCauseChain,
		components.DescriptionChain,
		components.K8sTool,
		components.MemoryManager,
		config,
	)
	if err != nil {
		return fmt.Errorf("failed to create orchestrator adapter: %w", err)
	}

	fmt.Println("Orchestrator Name:", adapter.Name())
	fmt.Println()

	// Prepare request using framework types
	request := &core.OrchestratorRequest{
		TaskID:      "analysis-task-001",
		TaskType:    "failure_analysis",
		Description: "Complete failure analysis workflow for pod crash",
		Parameters: map[string]interface{}{
			"failure_type":  "CrashLoopBackOff",
			"resource_type": "pod",
			"resource_name": "my-app-pod-xyz123",
			"namespace":     "production",
			"cluster_id":    "cluster-01",
			"error_message": "Error: OOMKilled",
			"language":      "en",
			"detail_level":  "normal",
		},
		Strategy: core.OrchestratorStrategy{
			Mode:          "sequential",
			EnableRetry:   false,
			MaxRetries:    3,
			RetryBackoff:  2,
			FailurePolicy: "stop",
			GlobalTimeout: 5 * time.Minute,
			StepTimeout:   60 * time.Second,
		},
		Options: core.OrchestratorOptions{
			EnableLogging: true,
			EnableMetrics: true,
			EnableTracing: false,
		},
		SessionID: "session-456",
		Timestamp: time.Now(),
	}

	// Execute workflow
	response, err := adapter.Execute(ctx, request)
	if err != nil {
		return fmt.Errorf("orchestration failed: %w", err)
	}

	// Print response summary
	fmt.Printf("Status: %s\n", response.Status)
	fmt.Printf("Message: %s\n", response.Message)
	fmt.Printf("Total Latency: %v\n", response.TotalLatency)
	fmt.Printf("Start Time: %s\n", response.StartTime.Format(time.RFC3339))
	fmt.Printf("End Time: %s\n", response.EndTime.Format(time.RFC3339))
	fmt.Printf("Steps Executed: %d\n", len(response.ExecutionSteps))
	fmt.Println()

	// Print execution steps
	fmt.Println("Execution Steps:")
	for _, step := range response.ExecutionSteps {
		fmt.Printf("  Step %d [%s]: %s\n", step.Step, step.Type, step.Name)
		fmt.Printf("    Description: %s\n", step.Description)
		fmt.Printf("    Status: %s\n", step.Status)
		fmt.Printf("    Duration: %v\n", step.Duration)
		if step.Error != "" {
			fmt.Printf("    Error: %s\n", step.Error)
		}
	}
	fmt.Println()

	// Print metadata
	fmt.Println("Metadata Keys:", getMapKeys(response.Metadata))

	return nil
}

// exampleRootCauseChain demonstrates using the RootCauseChainAdapter
func exampleRootCauseChain(ctx context.Context, components *Components) error {
	// Create adapter
	adapter := adapters.NewRootCauseChainAdapter(components.RootCauseChain)

	fmt.Println("Chain Name:", adapter.Name())
	fmt.Println("Chain Steps:", adapter.Steps())
	fmt.Println()

	// Prepare input (domain-specific)
	analysisInput := &root_cause.AnalysisInput{
		FailureType:  "pod_failure",
		ResourceType: "pod",
		ResourceName: "my-pod",
		Namespace:    "default",
		ClusterID:    "cluster-01",
		ErrorMessage: "OOMKilled - container exceeded memory limit",
		Timestamp:    time.Now(),
		PodLogs:      "2024-01-01 10:00:00 Starting application...\n2024-01-01 10:00:05 Out of memory error",
		PodEvents: []root_cause.K8sEvent{
			{
				Type:      "Warning",
				Reason:    "OOMKilled",
				Message:   "Container exceeded memory limit and was killed",
				Timestamp: time.Now().Add(-5 * time.Minute),
				Source:    "kubelet",
			},
		},
		ResourceStatus: map[string]string{
			"phase":  "CrashLoopBackOff",
			"reason": "OOMKilled",
		},
		Metrics: map[string]float64{
			"memory_utilization": 98.5,
			"cpu_utilization":    45.2,
		},
	}

	// Wrap in framework ChainInput
	chainInput := &core.ChainInput{
		Data:    analysisInput,
		Vars:    make(map[string]interface{}),
		Tags:    []string{"kubernetes", "pod_failure", "oom"},
		Options: core.DefaultChainOptions(),
	}

	// Execute through framework interface
	output, err := adapter.Process(ctx, chainInput)
	if err != nil {
		return fmt.Errorf("chain processing failed: %w", err)
	}

	// Print results
	fmt.Printf("Status: %s\n", output.Status)
	fmt.Printf("Total Latency: %v\n", output.TotalLatency)
	fmt.Printf("Steps Executed: %d\n", len(output.StepsExecuted))
	fmt.Println()

	// Print step execution details
	for _, step := range output.StepsExecuted {
		fmt.Printf("  Step %d: %s - %s (duration: %v)\n",
			step.StepNumber, step.StepName, step.Description, step.Duration)
	}
	fmt.Println()

	// Access domain-specific output
	if analysisOutput, ok := output.Data.(*root_cause.AnalysisOutput); ok {
		fmt.Println("Root Cause Analysis Result:")
		fmt.Printf("  Root Cause: %s\n", analysisOutput.RootCause)
		fmt.Printf("  Confidence: %.2f\n", analysisOutput.Confidence)
		fmt.Printf("  Category: %s\n", analysisOutput.Category)
		fmt.Printf("  Reasoning: %s\n", analysisOutput.Reasoning)
		fmt.Printf("  Recommendations: %d\n", len(analysisOutput.Recommendations))
		fmt.Println()

		// Print recommendations
		if len(analysisOutput.Recommendations) > 0 {
			fmt.Println("Recommendations:")
			for i, rec := range analysisOutput.Recommendations {
				fmt.Printf("  %d. [%s] %s\n", i+1, rec.Priority, rec.Action)
				fmt.Printf("     %s\n", rec.Description)
				if len(rec.Commands) > 0 {
					fmt.Printf("     Commands: %v\n", rec.Commands)
				}
			}
		}
	}

	return nil
}

// getMapKeys returns the keys of a map as a slice
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// prettyPrint prints a struct as formatted JSON
// nolint:unused,deadcode // Example function
func prettyPrint(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println(string(b))
}
