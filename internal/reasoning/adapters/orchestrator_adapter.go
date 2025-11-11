package adapters

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kart-io/k8s-agent/internal/reasoning/agents/k8s_tool"
	"github.com/kart-io/k8s-agent/internal/reasoning/agents/reasoning"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/description"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/root_cause"
	"github.com/kart-io/k8s-agent/internal/reasoning/memory"
	"github.com/kart-io/k8s-agent/internal/reasoning/orchestrator"
	"github.com/kart-io/k8s-agent/pkg/agent/core"
)

// OrchestratorAdapter adapts the orchestrator to use pkg/agent/core framework.
type OrchestratorAdapter struct {
	*core.BaseOrchestrator

	// Original orchestrator components
	reasoningAgent   *reasoning.ReasoningAgent
	rootCauseChain   *root_cause.RootCauseChain
	descriptionChain *description.DescriptionChain
	k8sTool          *k8s_tool.K8sTool
	memoryManager    memory.Manager
	config           *orchestrator.OrchestratorConfig

	// Adapters for framework
	reasoningAgentAdapter   *ReasoningAgentAdapter
	rootCauseChainAdapter   *RootCauseChainAdapter
	descriptionChainAdapter *DescriptionChainAdapter
	k8sToolAdapter          *K8sToolAdapter
}

// NewOrchestratorAdapter creates a new orchestrator adapter.
func NewOrchestratorAdapter(
	reasoningAgent *reasoning.ReasoningAgent,
	rootCauseChain *root_cause.RootCauseChain,
	descriptionChain *description.DescriptionChain,
	k8sTool *k8s_tool.K8sTool,
	memoryManager memory.Manager,
	config *orchestrator.OrchestratorConfig,
) (*OrchestratorAdapter, error) {
	if config == nil {
		config = orchestrator.DefaultOrchestratorConfig()
	}

	adapter := &OrchestratorAdapter{
		BaseOrchestrator: core.NewBaseOrchestrator("reasoning_orchestrator"),
		reasoningAgent:   reasoningAgent,
		rootCauseChain:   rootCauseChain,
		descriptionChain: descriptionChain,
		k8sTool:          k8sTool,
		memoryManager:    memoryManager,
		config:           config,
	}

	// Create adapters
	adapter.reasoningAgentAdapter = NewReasoningAgentAdapter(reasoningAgent)
	adapter.rootCauseChainAdapter = NewRootCauseChainAdapter(rootCauseChain)
	adapter.descriptionChainAdapter = NewDescriptionChainAdapter(descriptionChain)
	adapter.k8sToolAdapter = NewK8sToolAdapter(k8sTool)

	// Register components with base orchestrator
	if err := adapter.RegisterAgent("reasoning_agent", adapter.reasoningAgentAdapter); err != nil {
		return nil, fmt.Errorf("failed to register reasoning agent: %w", err)
	}
	if err := adapter.RegisterChain("root_cause_chain", adapter.rootCauseChainAdapter); err != nil {
		return nil, fmt.Errorf("failed to register root cause chain: %w", err)
	}
	if err := adapter.RegisterChain("description_chain", adapter.descriptionChainAdapter); err != nil {
		return nil, fmt.Errorf("failed to register description chain: %w", err)
	}
	if err := adapter.RegisterTool("k8s_tool", adapter.k8sToolAdapter); err != nil {
		return nil, fmt.Errorf("failed to register k8s tool: %w", err)
	}

	return adapter, nil
}

// Execute implements core.Orchestrator interface using the framework.
func (o *OrchestratorAdapter) Execute(ctx context.Context, request *core.OrchestratorRequest) (*core.OrchestratorResponse, error) {
	start := time.Now()

	response := &core.OrchestratorResponse{
		Result:         nil,
		Status:         "success",
		Message:        "",
		ExecutionPlan:  make([]core.ExecutionStep, 0),
		ExecutionSteps: make([]core.ExecutionStep, 0),
		TotalLatency:   0,
		StartTime:      start,
		EndTime:        time.Time{},
		Metadata:       make(map[string]interface{}),
	}

	// Apply timeout
	if request.Strategy.GlobalTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, request.Strategy.GlobalTimeout)
		defer cancel()
	}

	log.Printf("OrchestratorAdapter: Starting execution for task: %s", request.TaskID)

	// Step 1: Load memory context if enabled
	var similarCases []*memory.CaseMemory
	var conversationCount int
	if o.config.EnableMemory {
		stepStart := time.Now()
		var err error
		similarCases, conversationCount, err = o.loadMemoryContext(ctx, request)

		step := core.ExecutionStep{
			Step:          1,
			Name:          "load_memory_context",
			Type:          "memory",
			Description:   "Load history and similar cases from memory",
			ComponentName: "memory_manager",
			Status:        "success",
			StartTime:     stepStart,
			EndTime:       time.Now(),
			Duration:      time.Since(stepStart),
		}

		if err != nil {
			log.Printf("OrchestratorAdapter: Failed to load memory context: %v", err)
			step.Status = "failed"
			step.Error = err.Error()
		} else {
			log.Printf("OrchestratorAdapter: Loaded %d similar cases, %d conversations",
				len(similarCases), conversationCount)
			step.Output = map[string]interface{}{
				"similar_cases_count": len(similarCases),
				"conversation_count":  conversationCount,
			}
		}

		response.ExecutionSteps = append(response.ExecutionSteps, step)
		response.Metadata["similar_cases"] = similarCases
		response.Metadata["conversation_count"] = conversationCount
	}

	// Step 2: Execute root cause analysis using chain
	stepStart := time.Now()
	rootCauseResult, err := o.executeRootCauseAnalysis(ctx, request, similarCases)

	step := core.ExecutionStep{
		Step:          2,
		Name:          "root_cause_analysis",
		Type:          "chain",
		Description:   "Analyze root cause using LLM and rules",
		ComponentName: "root_cause_chain",
		StartTime:     stepStart,
		EndTime:       time.Now(),
		Duration:      time.Since(stepStart),
	}

	if err != nil {
		log.Printf("OrchestratorAdapter: Root cause analysis failed: %v", err)
		step.Status = "failed"
		step.Error = err.Error()
		response.ExecutionSteps = append(response.ExecutionSteps, step)
		response.Status = "failed"
		response.Message = fmt.Sprintf("root cause analysis failed: %v", err)
		response.EndTime = time.Now()
		response.TotalLatency = time.Since(start)
		return response, err
	}

	step.Status = "success"
	step.Output = rootCauseResult
	response.ExecutionSteps = append(response.ExecutionSteps, step)
	response.Metadata["root_cause"] = rootCauseResult
	log.Printf("OrchestratorAdapter: Root cause: %s (confidence: %.2f)",
		rootCauseResult.RootCause, rootCauseResult.Confidence)

	// Step 3: Generate description if enabled
	if o.config.EnableDescription {
		stepStart := time.Now()
		descResult, err := o.executeDescriptionGeneration(ctx, request, rootCauseResult)

		step := core.ExecutionStep{
			Step:          3,
			Name:          "generate_description",
			Type:          "chain",
			Description:   "Generate human-readable failure description",
			ComponentName: "description_chain",
			StartTime:     stepStart,
			EndTime:       time.Now(),
			Duration:      time.Since(stepStart),
		}

		if err != nil {
			log.Printf("OrchestratorAdapter: Description generation failed: %v", err)
			step.Status = "failed"
			step.Error = err.Error()
			response.Status = "partial"
			response.Message = fmt.Sprintf("description generation failed: %v", err)
		} else {
			step.Status = "success"
			step.Output = descResult
			response.Metadata["description"] = descResult
			log.Printf("OrchestratorAdapter: Generated description: %s", descResult.Title)
		}

		response.ExecutionSteps = append(response.ExecutionSteps, step)
	}

	// Step 4: Save to memory if enabled
	if o.config.EnableMemory && o.config.SaveToMemory {
		stepStart := time.Now()
		err := o.saveToMemory(ctx, request, rootCauseResult, response.Metadata["description"])

		step := core.ExecutionStep{
			Step:          4,
			Name:          "save_to_memory",
			Type:          "memory",
			Description:   "Save analysis result to memory",
			ComponentName: "memory_manager",
			StartTime:     stepStart,
			EndTime:       time.Now(),
			Duration:      time.Since(stepStart),
		}

		if err != nil {
			log.Printf("OrchestratorAdapter: Failed to save to memory: %v", err)
			step.Status = "failed"
			step.Error = err.Error()
		} else {
			step.Status = "success"
		}

		response.ExecutionSteps = append(response.ExecutionSteps, step)
	}

	// Build final result
	response.Result = map[string]interface{}{
		"root_cause":  rootCauseResult,
		"description": response.Metadata["description"],
		"metadata":    response.Metadata,
	}

	response.EndTime = time.Now()
	response.TotalLatency = time.Since(start)
	log.Printf("OrchestratorAdapter: Execution completed in %v", response.TotalLatency)

	return response, nil
}

// loadMemoryContext loads historical context and similar cases.
func (o *OrchestratorAdapter) loadMemoryContext(ctx context.Context, req *core.OrchestratorRequest) ([]*memory.CaseMemory, int, error) {
	var similarCases []*memory.CaseMemory
	var conversationCount int

	// Extract session ID
	sessionID := req.SessionID

	// Load conversation history
	if o.config.LoadHistoryContext && sessionID != "" {
		conversations, err := o.memoryManager.GetConversationHistory(ctx, sessionID, 10)
		if err != nil {
			log.Printf("Failed to get conversation history: %v", err)
		} else {
			conversationCount = len(conversations)
		}
	}

	// Load similar cases
	if o.config.LoadSimilarCases {
		// Build query from parameters
		query := fmt.Sprintf("%v %v %v %v",
			req.Parameters["failure_type"],
			req.Parameters["resource_type"],
			req.Parameters["resource_name"],
			req.Parameters["error_message"])

		cases, err := o.memoryManager.SearchSimilarCases(ctx, query, o.config.MaxSimilarCases)
		if err != nil {
			log.Printf("Failed to search similar cases: %v", err)
			return similarCases, conversationCount, err
		}
		similarCases = cases
	}

	return similarCases, conversationCount, nil
}

// executeRootCauseAnalysis executes root cause analysis using the chain.
func (o *OrchestratorAdapter) executeRootCauseAnalysis(ctx context.Context, req *core.OrchestratorRequest, similarCases []*memory.CaseMemory) (*root_cause.AnalysisOutput, error) {
	// Apply timeout
	if o.config.RootCauseTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, o.config.RootCauseTimeout)
		defer cancel()
	}

	// Build input from request parameters
	input := &root_cause.AnalysisInput{
		FailureType:  getStringParam(req.Parameters, "failure_type"),
		ResourceType: getStringParam(req.Parameters, "resource_type"),
		ResourceName: getStringParam(req.Parameters, "resource_name"),
		Namespace:    getStringParam(req.Parameters, "namespace"),
		ClusterID:    getStringParam(req.Parameters, "cluster_id"),
		ErrorMessage: getStringParam(req.Parameters, "error_message"),
		Timestamp:    req.Timestamp,
		PodLogs:      getStringParam(req.Parameters, "pod_logs"),
	}

	// Add resource status
	if rs, ok := req.Parameters["resource_status"].(map[string]string); ok {
		input.ResourceStatus = rs
	}

	// Add similar cases
	if len(similarCases) > 0 {
		input.SimilarCases = make([]root_cause.SimilarCase, len(similarCases))
		for i, c := range similarCases {
			input.SimilarCases[i] = root_cause.SimilarCase{
				Description: c.Description,
				RootCause:   c.RootCause,
				Similarity:  c.Similarity,
				Resolution:  c.Solution,
			}
		}
	}

	// Add events
	if events, ok := req.Parameters["events"].([]k8s_tool.EventInfo); ok {
		input.PodEvents = make([]root_cause.K8sEvent, len(events))
		for i, e := range events {
			input.PodEvents[i] = root_cause.K8sEvent{
				Type:      e.Type,
				Reason:    e.Reason,
				Message:   e.Message,
				Timestamp: e.LastTimestamp,
				Source:    e.Source,
			}
		}
	}

	// Add metrics
	if metrics, ok := req.Parameters["metrics"].(*k8s_tool.MetricsInfo); ok {
		input.Metrics = map[string]float64{
			"cpu_utilization":     metrics.CPU.Utilization,
			"memory_utilization":  metrics.Memory.Utilization,
			"storage_utilization": metrics.Storage.Utilization,
		}
	}

	return o.rootCauseChain.Analyze(ctx, input)
}

// executeDescriptionGeneration generates failure description using the chain.
func (o *OrchestratorAdapter) executeDescriptionGeneration(ctx context.Context, req *core.OrchestratorRequest, rootCause *root_cause.AnalysisOutput) (*description.DescriptionOutput, error) {
	// Apply timeout
	if o.config.DescriptionTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, o.config.DescriptionTimeout)
		defer cancel()
	}

	// Build input
	input := &description.DescriptionInput{
		FailureType:     getStringParam(req.Parameters, "failure_type"),
		ResourceType:    getStringParam(req.Parameters, "resource_type"),
		ResourceName:    getStringParam(req.Parameters, "resource_name"),
		Namespace:       getStringParam(req.Parameters, "namespace"),
		ClusterID:       getStringParam(req.Parameters, "cluster_id"),
		ErrorMessage:    getStringParam(req.Parameters, "error_message"),
		Timestamp:       req.Timestamp,
		Language:        getStringParam(req.Parameters, "language", "en"),
		DetailLevel:     getStringParam(req.Parameters, "detail_level", "normal"),
		PodLogs:         getStringParam(req.Parameters, "pod_logs"),
		IncludeTimeline: true,
	}

	// Add resource status
	if rs, ok := req.Parameters["resource_status"].(map[string]string); ok {
		input.ResourceStatus = rs
	}

	// Add events
	if events, ok := req.Parameters["events"].([]k8s_tool.EventInfo); ok {
		input.PodEvents = make([]description.PodEvent, len(events))
		for i, e := range events {
			input.PodEvents[i] = description.PodEvent{
				Type:      e.Type,
				Reason:    e.Reason,
				Message:   e.Message,
				Timestamp: e.LastTimestamp,
				Source:    e.Source,
			}
		}
	}

	// Add metrics
	if metrics, ok := req.Parameters["metrics"].(*k8s_tool.MetricsInfo); ok {
		input.Metrics = map[string]float64{
			"cpu_utilization":    metrics.CPU.Utilization,
			"memory_utilization": metrics.Memory.Utilization,
		}
	}

	// Add root cause info
	if rootCause != nil {
		input.RootCause = &description.RootCauseInfo{
			RootCause:  rootCause.RootCause,
			Confidence: rootCause.Confidence,
			Category:   rootCause.Category,
			Reasoning:  rootCause.Reasoning,
		}
	}

	return o.descriptionChain.Generate(ctx, input)
}

// saveToMemory saves analysis results to memory.
func (o *OrchestratorAdapter) saveToMemory(ctx context.Context, req *core.OrchestratorRequest, rootCause *root_cause.AnalysisOutput, desc interface{}) error {
	// Save conversation
	sessionID := req.SessionID
	if sessionID != "" {
		// User query
		userConv := &memory.Conversation{
			SessionID: sessionID,
			Role:      "user",
			Content: fmt.Sprintf("Analyze failure: %s/%s - %s",
				req.Parameters["resource_type"],
				req.Parameters["resource_name"],
				req.Parameters["error_message"]),
			Timestamp: time.Now(),
		}
		if err := o.memoryManager.AddConversation(ctx, userConv); err != nil {
			log.Printf("Failed to save user conversation: %v", err)
		}

		// AI response
		if rootCause != nil {
			assistantConv := &memory.Conversation{
				SessionID: sessionID,
				Role:      "assistant",
				Content:   fmt.Sprintf("Root cause: %s (confidence: %.2f)", rootCause.RootCause, rootCause.Confidence),
				Timestamp: time.Now(),
			}
			if err := o.memoryManager.AddConversation(ctx, assistantConv); err != nil {
				log.Printf("Failed to save assistant conversation: %v", err)
			}
		}
	}

	// Save case if confidence is high enough
	if rootCause != nil && rootCause.Confidence >= 0.7 {
		caseMemory := &memory.CaseMemory{
			Description:  getStringParam(req.Parameters, "error_message"),
			RootCause:    rootCause.RootCause,
			Solution:     rootCause.Reasoning,
			FailureType:  getStringParam(req.Parameters, "failure_type"),
			ResourceType: getStringParam(req.Parameters, "resource_type"),
			Tags:         []string{getStringParam(req.Parameters, "failure_type"), getStringParam(req.Parameters, "resource_type")},
		}

		// Add description info if available
		if descOutput, ok := desc.(*description.DescriptionOutput); ok && descOutput != nil {
			caseMemory.Description = descOutput.Title
			caseMemory.Symptoms = descOutput.AffectedComponents
		}

		if err := o.memoryManager.AddCase(ctx, caseMemory); err != nil {
			log.Printf("Failed to save case to memory: %v", err)
			return err
		}

		log.Printf("Saved case to memory: %s", caseMemory.ID)
	}

	return nil
}

// getStringParam is a helper to extract string parameters with optional default.
func getStringParam(params map[string]interface{}, key string, defaultVal ...string) string {
	if val, ok := params[key].(string); ok {
		return val
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return ""
}
