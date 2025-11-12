package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/kart-io/k8s-agent/internal/reasoning/agents/k8s_tool"
	"github.com/kart-io/k8s-agent/internal/reasoning/agents/reasoning"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/description"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/root_cause"
	"github.com/kart-io/k8s-agent/pkg/agent/core"
)

// ReasoningAgentAdapter adapts ReasoningAgent to core.Agent interface.
type ReasoningAgentAdapter struct {
	*core.BaseAgent
	reasoningAgent *reasoning.ReasoningAgent
}

// NewReasoningAgentAdapter creates a new adapter for ReasoningAgent.
func NewReasoningAgentAdapter(agent *reasoning.ReasoningAgent) *ReasoningAgentAdapter {
	return &ReasoningAgentAdapter{
		BaseAgent: core.NewBaseAgent(
			"reasoning_agent",
			"AI-driven Kubernetes failure analysis agent",
			[]string{"root_cause_analysis", "description_generation", "k8s_context_fetch"},
		),
		reasoningAgent: agent,
	}
}

// Execute implements core.Agent interface by adapting to ReasoningAgent's Analyze method.
func (a *ReasoningAgentAdapter) Execute(ctx context.Context, input *core.AgentInput) (*core.AgentOutput, error) {
	start := time.Now()

	// Convert core.AgentInput to reasoning.AnalysisInput
	analysisInput := a.convertInput(input)

	// Execute analysis using the wrapped ReasoningAgent
	analysisOutput, err := a.reasoningAgent.Analyze(ctx, analysisInput)
	if err != nil {
		return &core.AgentOutput{
			Status:    "failed",
			Message:   err.Error(),
			Latency:   time.Since(start),
		}, err
	}

	// Convert reasoning.AnalysisOutput to core.AgentOutput
	return a.convertOutput(analysisOutput, time.Since(start)), nil
}

// convertInput converts core.AgentInput to reasoning.AnalysisInput.
func (a *ReasoningAgentAdapter) convertInput(input *core.AgentInput) *reasoning.AnalysisInput {
	analysisInput := &reasoning.AnalysisInput{
		IncludeRootCause:   true,
		IncludeDescription: true,
		IncludeK8sContext:  true,
		Language:           "en",
		DetailLevel:        "normal",
		Timestamp:          input.Timestamp,
	}

	// Extract fields from Context map
	if failureType, ok := input.Context["failure_type"].(string); ok {
		analysisInput.FailureType = failureType
	}
	if resourceType, ok := input.Context["resource_type"].(string); ok {
		analysisInput.ResourceType = resourceType
	}
	if resourceName, ok := input.Context["resource_name"].(string); ok {
		analysisInput.ResourceName = resourceName
	}
	if namespace, ok := input.Context["namespace"].(string); ok {
		analysisInput.Namespace = namespace
	}
	if clusterID, ok := input.Context["cluster_id"].(string); ok {
		analysisInput.ClusterID = clusterID
	}
	if errorMessage, ok := input.Context["error_message"].(string); ok {
		analysisInput.ErrorMessage = errorMessage
	}

	// Extract K8s tool options
	if fetchLogs, ok := input.Context["fetch_logs"].(bool); ok {
		analysisInput.FetchLogs = fetchLogs
	}
	if fetchEvents, ok := input.Context["fetch_events"].(bool); ok {
		analysisInput.FetchEvents = fetchEvents
	}
	if fetchMetrics, ok := input.Context["fetch_metrics"].(bool); ok {
		analysisInput.FetchMetrics = fetchMetrics
	}

	// Extract language and detail level if provided
	if language, ok := input.Context["language"].(string); ok {
		analysisInput.Language = language
	}
	if detailLevel, ok := input.Context["detail_level"].(string); ok {
		analysisInput.DetailLevel = detailLevel
	}

	// Extract existing context data if provided
	if podLogs, ok := input.Context["pod_logs"].(string); ok {
		analysisInput.PodLogs = podLogs
	}
	if events, ok := input.Context["events"].([]k8s_tool.EventInfo); ok {
		analysisInput.Events = events
	}
	if metrics, ok := input.Context["metrics"].(*k8s_tool.MetricsInfo); ok {
		analysisInput.Metrics = metrics
	}
	if resourceStatus, ok := input.Context["resource_status"].(map[string]string); ok {
		analysisInput.ResourceStatus = resourceStatus
	}

	return analysisInput
}

// convertOutput converts reasoning.AnalysisOutput to core.AgentOutput.
func (a *ReasoningAgentAdapter) convertOutput(output *reasoning.AnalysisOutput, latency time.Duration) *core.AgentOutput {
	agentOutput := &core.AgentOutput{
		Result:         output,
		Status:         "success",
		Message:        "Analysis completed successfully",
		ReasoningSteps: make([]core.ReasoningStep, 0),
		ToolCalls:      make([]core.ToolCall, 0),
		Latency:        latency,
		Metadata:       make(map[string]interface{}),
	}

	// Convert reasoning steps
	for _, step := range output.ReasoningSteps {
		agentOutput.ReasoningSteps = append(agentOutput.ReasoningSteps, core.ReasoningStep{
			Step:        step.Step,
			Action:      step.Action,
			Description: step.Description,
			Result:      step.Result,
			Duration:    step.Duration,
			Success:     step.Success,
			Error:       step.Error,
		})
	}

	// Add metadata
	if output.RootCause != nil {
		agentOutput.Metadata["root_cause"] = output.RootCause.RootCause
		agentOutput.Metadata["confidence"] = output.RootCause.Confidence
		agentOutput.Metadata["category"] = output.RootCause.Category
	}
	if output.Description != nil {
		agentOutput.Metadata["description_title"] = output.Description.Title
		agentOutput.Metadata["severity"] = output.Description.Severity
	}
	if output.K8sContext != nil {
		agentOutput.Metadata["k8s_context_fetched"] = true
	}

	return agentOutput
}

// K8sToolAdapter adapts K8sTool to core.Tool interface.
type K8sToolAdapter struct {
	k8sTool *k8s_tool.K8sTool
}

// NewK8sToolAdapter creates a new adapter for K8sTool.
func NewK8sToolAdapter(tool *k8s_tool.K8sTool) *K8sToolAdapter {
	return &K8sToolAdapter{
		k8sTool: tool,
	}
}

// Name implements core.Tool interface.
func (t *K8sToolAdapter) Name() string {
	return t.k8sTool.Name()
}

// Description implements core.Tool interface.
func (t *K8sToolAdapter) Description() string {
	return t.k8sTool.Description()
}

// Parameters implements core.Tool interface.
func (t *K8sToolAdapter) Parameters() []core.ToolParameter {
	return []core.ToolParameter{
		{
			Name:        "action",
			Type:        "string",
			Description: "Action to perform: get, describe, logs, events, list, metrics, top",
			Required:    true,
		},
		{
			Name:        "resource_type",
			Type:        "string",
			Description: "Resource type: pod, deployment, service, node",
			Required:    true,
		},
		{
			Name:        "resource_name",
			Type:        "string",
			Description: "Resource name",
			Required:    false,
		},
		{
			Name:        "namespace",
			Type:        "string",
			Description: "Namespace",
			Required:    false,
			Default:     "default",
		},
	}
}

// Execute implements core.Tool interface.
func (t *K8sToolAdapter) Execute(ctx context.Context, input *core.ToolInput) (*core.ToolOutput, error) {
	// Convert core.ToolInput to k8s_tool.ToolInput
	toolInput := &k8s_tool.ToolInput{
		Action:     input.Action,
		Parameters: make(map[string]string),
		ClusterID:  "",
	}

	// Extract parameters
	if resourceType, ok := input.Parameters["resource_type"].(string); ok {
		toolInput.ResourceType = resourceType
	}
	if resourceName, ok := input.Parameters["resource_name"].(string); ok {
		toolInput.ResourceName = resourceName
	}
	if namespace, ok := input.Parameters["namespace"].(string); ok {
		toolInput.Namespace = namespace
	}
	if clusterID, ok := input.Parameters["cluster_id"].(string); ok {
		toolInput.ClusterID = clusterID
	}

	// Convert other parameters to string map
	for k, v := range input.Parameters {
		if strVal, ok := v.(string); ok {
			toolInput.Parameters[k] = strVal
		}
	}

	// Execute k8s tool
	output, err := t.k8sTool.Execute(ctx, toolInput)
	if err != nil {
		return &core.ToolOutput{
			Success: false,
			Data:    nil,
			Message: "",
			Error:   err.Error(),
		}, err
	}

	// Convert k8s_tool.ToolOutput to core.ToolOutput
	return &core.ToolOutput{
		Success: output.Success,
		Data:    output.Data,
		Message: "",
		Error:   output.ErrorMsg,
	}, nil
}

// RootCauseChainAdapter adapts RootCauseChain to core.Chain interface.
type RootCauseChainAdapter struct {
	*core.BaseChain
	chain *root_cause.RootCauseChain
}

// NewRootCauseChainAdapter creates a new adapter for RootCauseChain.
func NewRootCauseChainAdapter(chain *root_cause.RootCauseChain) *RootCauseChainAdapter {
	return &RootCauseChainAdapter{
		BaseChain: core.NewBaseChain("root_cause_chain", nil),
		chain:     chain,
	}
}

// Process implements core.Chain interface.
func (c *RootCauseChainAdapter) Process(ctx context.Context, input *core.ChainInput) (*core.ChainOutput, error) {
	start := time.Now()

	// Convert input
	analysisInput, ok := input.Data.(*root_cause.AnalysisInput)
	if !ok {
		return nil, fmt.Errorf("invalid input type, expected *root_cause.AnalysisInput")
	}

	// Execute analysis
	output, err := c.chain.Analyze(ctx, analysisInput)
	if err != nil {
		return &core.ChainOutput{
			Data:         nil,
			Status:       "failed",
			TotalLatency: time.Since(start),
		}, err
	}

	// Convert output
	return &core.ChainOutput{
		Data: output,
		StepsExecuted: []core.StepExecution{
			{
				StepNumber:  1,
				StepName:    "root_cause_analysis",
				Description: "Analyze root cause using LLM",
				Input:       analysisInput,
				Output:      output,
				Duration:    output.Latency,
				Success:     true,
			},
		},
		Status:       "success",
		TotalLatency: time.Since(start),
	}, nil
}

// DescriptionChainAdapter adapts DescriptionChain to core.Chain interface.
type DescriptionChainAdapter struct {
	*core.BaseChain
	chain *description.DescriptionChain
}

// NewDescriptionChainAdapter creates a new adapter for DescriptionChain.
func NewDescriptionChainAdapter(chain *description.DescriptionChain) *DescriptionChainAdapter {
	return &DescriptionChainAdapter{
		BaseChain: core.NewBaseChain("description_chain", nil),
		chain:     chain,
	}
}

// Process implements core.Chain interface.
func (c *DescriptionChainAdapter) Process(ctx context.Context, input *core.ChainInput) (*core.ChainOutput, error) {
	start := time.Now()

	// Convert input
	descInput, ok := input.Data.(*description.DescriptionInput)
	if !ok {
		return nil, fmt.Errorf("invalid input type, expected *description.DescriptionInput")
	}

	// Execute generation
	output, err := c.chain.Generate(ctx, descInput)
	if err != nil {
		return &core.ChainOutput{
			Data:         nil,
			Status:       "failed",
			TotalLatency: time.Since(start),
		}, err
	}

	// Convert output
	return &core.ChainOutput{
		Data: output,
		StepsExecuted: []core.StepExecution{
			{
				StepNumber:  1,
				StepName:    "description_generation",
				Description: "Generate failure description",
				Input:       descInput,
				Output:      output,
				Duration:    output.Latency,
				Success:     true,
			},
		},
		Status:       "success",
		TotalLatency: time.Since(start),
	}, nil
}
