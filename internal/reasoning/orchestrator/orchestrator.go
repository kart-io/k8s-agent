package orchestrator

import (
	"context"
	"fmt"
	"log"
	"github.com/kart-io/k8s-agent/internal/reasoning/agents/k8s_tool"
	"github.com/kart-io/k8s-agent/internal/reasoning/agents/reasoning"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/description"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/root_cause"
	"github.com/kart-io/k8s-agent/internal/reasoning/memory"
	"time"
)

// NewOrchestrator 创建新的 Orchestrator
func NewOrchestrator(
	reasoningAgent *reasoning.ReasoningAgent,
	rootCauseChain *root_cause.RootCauseChain,
	descriptionChain *description.DescriptionChain,
	k8sTool *k8s_tool.K8sTool,
	memoryManager memory.Manager,
	config *OrchestratorConfig,
) (*Orchestrator, error) {
	if config == nil {
		config = DefaultOrchestratorConfig()
	}

	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 验证必需组件
	if reasoningAgent == nil {
		return nil, fmt.Errorf("reasoning agent is required")
	}

	if rootCauseChain == nil {
		return nil, fmt.Errorf("root cause chain is required")
	}

	if config.EnableDescription && descriptionChain == nil {
		return nil, fmt.Errorf("description chain is required when enable_description is true")
	}

	if config.EnableToolAgent && k8sTool == nil {
		return nil, fmt.Errorf("k8s tool is required when enable_tool_agent is true")
	}

	if config.EnableMemory && memoryManager == nil {
		return nil, fmt.Errorf("memory manager is required when enable_memory is true")
	}

	return &Orchestrator{
		reasoningAgent:   reasoningAgent,
		rootCauseChain:   rootCauseChain,
		descriptionChain: descriptionChain,
		k8sTool:          k8sTool,
		memoryManager:    memoryManager,
		config:           config,
	}, nil
}

// Analyze 执行完整的故障分析流程
func (o *Orchestrator) Analyze(ctx context.Context, req *AnalysisRequest) (*AnalysisResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}

	// 应用超时
	if o.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, o.config.Timeout)
		defer cancel()
	}

	start := time.Now()
	response := &AnalysisResponse{
		ExecutionSteps: make([]ExecutionStep, 0),
		Timestamp:      start,
	}

	log.Printf("Orchestrator: Starting analysis for %s/%s in namespace %s",
		req.ResourceType, req.ResourceName, req.Namespace)

	// 步骤 1: 加载历史上下文和相似案例
	var similarCases []*memory.CaseMemory
	var conversationCount int
	if o.config.EnableMemory {
		stepStart := time.Now()
		var err error
		similarCases, conversationCount, err = o.loadMemoryContext(ctx, req)

		step := ExecutionStep{
			Step:        1,
			Name:        "load_memory_context",
			Description: "Load history and similar cases from memory",
			Duration:    time.Since(stepStart),
		}

		if err != nil {
			log.Printf("Orchestrator: Failed to load memory context: %v", err)
			step.Status = "failed"
			step.Error = err.Error()
		} else {
			step.Status = "success"
			log.Printf("Orchestrator: Loaded %d similar cases, %d conversations",
				len(similarCases), conversationCount)
		}

		response.ExecutionSteps = append(response.ExecutionSteps, step)
		response.SimilarCases = similarCases
		response.ConversationCount = conversationCount
	}

	// 步骤 2: 执行根因分析
	stepStart := time.Now()
	rootCauseResult, err := o.analyzeRootCause(ctx, req, similarCases)

	step := ExecutionStep{
		Step:        2,
		Name:        "root_cause_analysis",
		Description: "Analyze root cause using LLM and rules",
		Duration:    time.Since(stepStart),
	}

	if err != nil {
		log.Printf("Orchestrator: Root cause analysis failed: %v", err)
		step.Status = "failed"
		step.Error = err.Error()
		response.ExecutionSteps = append(response.ExecutionSteps, step)
		response.Error = fmt.Sprintf("root cause analysis failed: %v", err)
		return response, err
	}

	step.Status = "success"
	response.ExecutionSteps = append(response.ExecutionSteps, step)
	response.RootCause = rootCauseResult
	log.Printf("Orchestrator: Root cause: %s (confidence: %.2f)",
		rootCauseResult.RootCause, rootCauseResult.Confidence)

	// 步骤 3: 生成故障描述（如果启用）
	if o.config.EnableDescription {
		stepStart := time.Now()
		descResult, err := o.generateDescription(ctx, req, rootCauseResult)

		step := ExecutionStep{
			Step:        3,
			Name:        "generate_description",
			Description: "Generate human-readable failure description",
			Duration:    time.Since(stepStart),
		}

		if err != nil {
			log.Printf("Orchestrator: Description generation failed: %v", err)
			step.Status = "failed"
			step.Error = err.Error()
			response.PartialError = fmt.Sprintf("description generation failed: %v", err)
		} else {
			step.Status = "success"
			response.Description = descResult
			log.Printf("Orchestrator: Generated description: %s", descResult.Title)
		}

		response.ExecutionSteps = append(response.ExecutionSteps, step)
	}

	// 步骤 4: 保存到 Memory（如果启用）
	if o.config.EnableMemory && o.config.SaveToMemory {
		stepStart := time.Now()
		err := o.saveToMemory(ctx, req, response)

		step := ExecutionStep{
			Step:        4,
			Name:        "save_to_memory",
			Description: "Save analysis result to memory",
			Duration:    time.Since(stepStart),
		}

		if err != nil {
			log.Printf("Orchestrator: Failed to save to memory: %v", err)
			step.Status = "failed"
			step.Error = err.Error()
		} else {
			step.Status = "success"
		}

		response.ExecutionSteps = append(response.ExecutionSteps, step)
	}

	response.TotalLatency = time.Since(start)
	log.Printf("Orchestrator: Analysis completed in %v", response.TotalLatency)

	return response, nil
}

// loadMemoryContext 加载历史上下文和相似案例
func (o *Orchestrator) loadMemoryContext(ctx context.Context, req *AnalysisRequest) ([]*memory.CaseMemory, int, error) {
	var similarCases []*memory.CaseMemory
	var conversationCount int

	// 加载对话历史
	if o.config.LoadHistoryContext && req.SessionID != "" {
		conversations, err := o.memoryManager.GetConversationHistory(ctx, req.SessionID, 10)
		if err != nil {
			log.Printf("Failed to get conversation history: %v", err)
		} else {
			conversationCount = len(conversations)
		}
	}

	// 加载相似案例
	if o.config.LoadSimilarCases {
		// 构建查询文本
		query := fmt.Sprintf("%s %s %s %s",
			req.FailureType, req.ResourceType, req.ResourceName, req.ErrorMessage)

		cases, err := o.memoryManager.SearchSimilarCases(ctx, query, o.config.MaxSimilarCases)
		if err != nil {
			log.Printf("Failed to search similar cases: %v", err)
			return similarCases, conversationCount, err
		}
		similarCases = cases
	}

	return similarCases, conversationCount, nil
}

// analyzeRootCause 执行根因分析
func (o *Orchestrator) analyzeRootCause(ctx context.Context, req *AnalysisRequest, similarCases []*memory.CaseMemory) (*root_cause.AnalysisOutput, error) {
	// 应用超时
	if o.config.RootCauseTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, o.config.RootCauseTimeout)
		defer cancel()
	}

	// 构建输入
	input := &root_cause.AnalysisInput{
		FailureType:    req.FailureType,
		ResourceType:   req.ResourceType,
		ResourceName:   req.ResourceName,
		Namespace:      req.Namespace,
		ClusterID:      req.ClusterID,
		ErrorMessage:   req.ErrorMessage,
		Timestamp:      req.Timestamp,
		PodLogs:        req.PodLogs,
		ResourceStatus: req.ResourceStatus,
	}

	// 添加相似案例到输入
	if len(similarCases) > 0 {
		input.SimilarCases = make([]root_cause.SimilarCase, len(similarCases))
		for i, c := range similarCases {
			input.SimilarCases[i] = root_cause.SimilarCase{
				Description: c.Description,
				RootCause:   c.RootCause,
				Similarity:  c.Similarity,
				Resolution:  c.Solution, // 使用 Resolution 字段
			}
		}
	}

	// 转换事件
	if len(req.Events) > 0 {
		input.PodEvents = make([]root_cause.K8sEvent, len(req.Events))
		for i, e := range req.Events {
			input.PodEvents[i] = root_cause.K8sEvent{
				Type:      e.Type,
				Reason:    e.Reason,
				Message:   e.Message,
				Timestamp: e.LastTimestamp,
				Source:    e.Source,
			}
		}
	}

	// 转换指标
	if req.Metrics != nil {
		input.Metrics = map[string]float64{
			"cpu_utilization":     req.Metrics.CPU.Utilization,
			"memory_utilization":  req.Metrics.Memory.Utilization,
			"storage_utilization": req.Metrics.Storage.Utilization,
		}
	}

	return o.rootCauseChain.Analyze(ctx, input)
}

// generateDescription 生成故障描述
func (o *Orchestrator) generateDescription(ctx context.Context, req *AnalysisRequest, rootCause *root_cause.AnalysisOutput) (*description.DescriptionOutput, error) {
	// 应用超时
	if o.config.DescriptionTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, o.config.DescriptionTimeout)
		defer cancel()
	}

	// 构建输入
	input := &description.DescriptionInput{
		FailureType:     req.FailureType,
		ResourceType:    req.ResourceType,
		ResourceName:    req.ResourceName,
		Namespace:       req.Namespace,
		ClusterID:       req.ClusterID,
		ErrorMessage:    req.ErrorMessage,
		Timestamp:       req.Timestamp,
		Language:        req.Language,
		DetailLevel:     req.DetailLevel,
		PodLogs:         req.PodLogs,
		ResourceStatus:  req.ResourceStatus,
		IncludeTimeline: true,
	}

	// 转换事件
	if len(req.Events) > 0 {
		input.PodEvents = make([]description.PodEvent, len(req.Events))
		for i, e := range req.Events {
			input.PodEvents[i] = description.PodEvent{
				Type:      e.Type,
				Reason:    e.Reason,
				Message:   e.Message,
				Timestamp: e.LastTimestamp,
				Source:    e.Source,
			}
		}
	}

	// 转换指标
	if req.Metrics != nil {
		input.Metrics = map[string]float64{
			"cpu_utilization":    req.Metrics.CPU.Utilization,
			"memory_utilization": req.Metrics.Memory.Utilization,
		}
	}

	// 添加根因信息
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

// saveToMemory 保存分析结果到 Memory
func (o *Orchestrator) saveToMemory(ctx context.Context, req *AnalysisRequest, resp *AnalysisResponse) error {
	// 保存对话
	if req.SessionID != "" {
		// 用户查询
		userConv := &memory.Conversation{
			SessionID: req.SessionID,
			Role:      "user",
			Content:   fmt.Sprintf("Analyze failure: %s/%s - %s", req.ResourceType, req.ResourceName, req.ErrorMessage),
			Timestamp: time.Now(),
		}
		if err := o.memoryManager.AddConversation(ctx, userConv); err != nil {
			log.Printf("Failed to save user conversation: %v", err)
		}

		// AI 响应
		if resp.RootCause != nil {
			assistantConv := &memory.Conversation{
				SessionID: req.SessionID,
				Role:      "assistant",
				Content:   fmt.Sprintf("Root cause: %s (confidence: %.2f)", resp.RootCause.RootCause, resp.RootCause.Confidence),
				Timestamp: time.Now(),
			}
			if err := o.memoryManager.AddConversation(ctx, assistantConv); err != nil {
				log.Printf("Failed to save assistant conversation: %v", err)
			}
		}
	}

	// 保存案例（如果置信度足够高）
	if resp.RootCause != nil && resp.RootCause.Confidence >= 0.7 {
		caseMemory := &memory.CaseMemory{
			Description:  req.ErrorMessage,
			RootCause:    resp.RootCause.RootCause,
			Solution:     resp.RootCause.Reasoning,
			FailureType:  req.FailureType,
			ResourceType: req.ResourceType,
			Tags:         []string{req.FailureType, req.ResourceType},
		}

		// 添加描述信息
		if resp.Description != nil {
			caseMemory.Description = resp.Description.Title
			caseMemory.Symptoms = resp.Description.AffectedComponents
		}

		if err := o.memoryManager.AddCase(ctx, caseMemory); err != nil {
			log.Printf("Failed to save case to memory: %v", err)
			return err
		}

		log.Printf("Saved case to memory: %s", caseMemory.ID)
	}

	return nil
}

// validateConfig 验证配置
func validateConfig(config *OrchestratorConfig) error {
	if config.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	if config.RootCauseTimeout < 0 {
		return fmt.Errorf("root_cause_timeout cannot be negative")
	}

	if config.DescriptionTimeout < 0 {
		return fmt.Errorf("description_timeout cannot be negative")
	}

	if config.ToolAgentTimeout < 0 {
		return fmt.Errorf("tool_agent_timeout cannot be negative")
	}

	if config.MaxSimilarCases < 0 {
		return fmt.Errorf("max_similar_cases cannot be negative")
	}

	if config.MaxToolCalls < 0 {
		return fmt.Errorf("max_tool_calls cannot be negative")
	}

	return nil
}
