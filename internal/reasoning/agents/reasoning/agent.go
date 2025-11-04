package reasoning

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kart-io/k8s-agent/internal/reasoning/agents/k8s_tool"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/description"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/root_cause"
)

// ReasoningAgent Reasoning Agent 实现.
type ReasoningAgent struct {
	// Chains
	rootCauseChain   *root_cause.RootCauseChain
	descriptionChain *description.DescriptionChain

	// Tools
	k8sTool *k8s_tool.K8sTool

	// 配置
	config *AgentConfig
}

// NewReasoningAgent 创建新的 Reasoning Agent.
func NewReasoningAgent(
	rootCauseChain *root_cause.RootCauseChain,
	descriptionChain *description.DescriptionChain,
	k8sTool *k8s_tool.K8sTool,
	config *AgentConfig,
) (*ReasoningAgent, error) {
	if config == nil {
		config = DefaultAgentConfig()
	}

	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 验证必需的组件
	if config.EnableRootCause && rootCauseChain == nil {
		return nil, fmt.Errorf("root cause chain is required when enable_root_cause is true")
	}
	if config.EnableDescription && descriptionChain == nil {
		return nil, fmt.Errorf("description chain is required when enable_description is true")
	}
	if config.EnableK8sTool && k8sTool == nil {
		return nil, fmt.Errorf("k8s tool is required when enable_k8s_tool is true")
	}

	return &ReasoningAgent{
		rootCauseChain:   rootCauseChain,
		descriptionChain: descriptionChain,
		k8sTool:          k8sTool,
		config:           config,
	}, nil
}

// Analyze 执行完整的故障分析推理.
func (a *ReasoningAgent) Analyze(ctx context.Context, input *AnalysisInput) (*AnalysisOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}

	// 应用超时
	if a.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, a.config.Timeout)
		defer cancel()
	}

	start := time.Now()

	// 应用默认值
	a.applyDefaults(input)

	// 初始化输出
	output := &AnalysisOutput{
		ReasoningSteps: make([]ReasoningStep, 0),
		Timestamp:      start,
	}

	log.Printf("Starting reasoning analysis for %s/%s in namespace %s",
		input.ResourceType, input.ResourceName, input.Namespace)

	// 步骤 1: 获取 K8s 上下文
	if input.IncludeK8sContext && a.config.EnableK8sTool {
		k8sCtx, err := a.fetchK8sContext(ctx, input)
		if err != nil {
			return nil, &AnalysisError{
				Phase:   PhaseK8sContext,
				Message: "failed to fetch K8s context",
				Err:     err,
			}
		}
		output.K8sContext = k8sCtx
		output.ReasoningSteps = append(output.ReasoningSteps, ReasoningStep{
			Step:        1,
			Action:      "fetch_k8s_context",
			Description: "Fetched K8s context (logs, events, metrics)",
			Result:      fmt.Sprintf("Fetched %d events", len(k8sCtx.Events)),
			Duration:    time.Since(start),
			Success:     true,
		})
	}

	// 步骤 2: 根因分析
	if input.IncludeRootCause && a.config.EnableRootCause {
		rootCauseStart := time.Now()
		rootCauseResult, err := a.AnalyzeRootCause(ctx, input)
		if err != nil {
			output.ReasoningSteps = append(output.ReasoningSteps, ReasoningStep{
				Step:        2,
				Action:      "root_cause_analysis",
				Description: "Analyze root cause of failure",
				Result:      "",
				Duration:    time.Since(rootCauseStart),
				Success:     false,
				Error:       err.Error(),
			})
			return nil, &AnalysisError{
				Phase:   PhaseRootCause,
				Message: "failed to analyze root cause",
				Err:     err,
			}
		}
		output.RootCause = rootCauseResult
		output.ReasoningSteps = append(output.ReasoningSteps, ReasoningStep{
			Step:        2,
			Action:      "root_cause_analysis",
			Description: "Analyze root cause of failure",
			Result:      fmt.Sprintf("Root cause: %s (confidence: %.2f)", rootCauseResult.RootCause, rootCauseResult.Confidence),
			Duration:    time.Since(rootCauseStart),
			Success:     true,
		})
	}

	// 步骤 3: 生成故障描述
	if input.IncludeDescription && a.config.EnableDescription {
		descStart := time.Now()
		descResult, err := a.GenerateDescription(ctx, input)
		if err != nil {
			output.ReasoningSteps = append(output.ReasoningSteps, ReasoningStep{
				Step:        3,
				Action:      "generate_description",
				Description: "Generate failure description",
				Result:      "",
				Duration:    time.Since(descStart),
				Success:     false,
				Error:       err.Error(),
			})
			return nil, &AnalysisError{
				Phase:   PhaseDescription,
				Message: "failed to generate description",
				Err:     err,
			}
		}
		output.Description = descResult
		output.ReasoningSteps = append(output.ReasoningSteps, ReasoningStep{
			Step:        3,
			Action:      "generate_description",
			Description: "Generate failure description",
			Result:      fmt.Sprintf("Generated description: %s (severity: %s)", descResult.Title, descResult.Severity),
			Duration:    time.Since(descStart),
			Success:     true,
		})
	}

	output.TotalLatency = time.Since(start)

	log.Printf("Reasoning analysis completed in %v", output.TotalLatency)

	return output, nil
}

// AnalyzeRootCause 仅执行根因分析.
func (a *ReasoningAgent) AnalyzeRootCause(ctx context.Context, input *AnalysisInput) (*root_cause.AnalysisOutput, error) {
	if !a.config.EnableRootCause {
		return nil, fmt.Errorf("root cause analysis is disabled")
	}

	// 构建根因分析输入
	rcInput := a.buildRootCauseInput(input)

	// 执行根因分析
	return a.rootCauseChain.Analyze(ctx, rcInput)
}

// GenerateDescription 仅生成故障描述.
func (a *ReasoningAgent) GenerateDescription(ctx context.Context, input *AnalysisInput) (*description.DescriptionOutput, error) {
	if !a.config.EnableDescription {
		return nil, fmt.Errorf("description generation is disabled")
	}

	// 构建描述生成输入
	descInput := a.buildDescriptionInput(input)

	// 执行描述生成
	return a.descriptionChain.Generate(ctx, descInput)
}

// fetchK8sContext 获取 K8s 上下文信息.
func (a *ReasoningAgent) fetchK8sContext(ctx context.Context, input *AnalysisInput) (*K8sContext, error) {
	k8sCtx := &K8sContext{
		FetchedAt: time.Now(),
	}

	// 获取资源信息
	if input.ResourceType == "pod" {
		podInput := &k8s_tool.ToolInput{
			Action:       "get",
			ResourceType: "pod",
			ResourceName: input.ResourceName,
			Namespace:    input.Namespace,
			ClusterID:    input.ClusterID,
		}

		output, err := a.k8sTool.Execute(ctx, podInput)
		if err == nil && output.Success {
			if podInfo, ok := output.Data.(*k8s_tool.PodInfo); ok {
				k8sCtx.PodInfo = podInfo
			}
		}
	}

	// 获取日志
	if input.FetchLogs && input.ResourceType == "pod" {
		logsInput := &k8s_tool.ToolInput{
			Action:       "logs",
			ResourceType: "pod",
			ResourceName: input.ResourceName,
			Namespace:    input.Namespace,
			ClusterID:    input.ClusterID,
			Parameters: map[string]string{
				"timestamps": "true",
			},
		}

		output, err := a.k8sTool.Execute(ctx, logsInput)
		if err == nil && output.Success {
			if logData, ok := output.Data.(map[string]interface{}); ok {
				if logs, ok := logData["logs"].(string); ok {
					k8sCtx.Logs = logs
				}
			}
		}
	}

	// 获取事件
	if input.FetchEvents {
		eventsInput := &k8s_tool.ToolInput{
			Action:       "events",
			ResourceType: input.ResourceType,
			ResourceName: input.ResourceName,
			Namespace:    input.Namespace,
			ClusterID:    input.ClusterID,
		}

		output, err := a.k8sTool.Execute(ctx, eventsInput)
		if err == nil && output.Success {
			if events, ok := output.Data.([]k8s_tool.EventInfo); ok {
				k8sCtx.Events = events
			}
		}
	}

	// 获取指标
	if input.FetchMetrics {
		metricsInput := &k8s_tool.ToolInput{
			Action:       "metrics",
			ResourceType: input.ResourceType,
			ResourceName: input.ResourceName,
			Namespace:    input.Namespace,
			ClusterID:    input.ClusterID,
		}

		output, err := a.k8sTool.Execute(ctx, metricsInput)
		if err == nil && output.Success {
			if metrics, ok := output.Data.(*k8s_tool.MetricsInfo); ok {
				k8sCtx.Metrics = metrics
			}
		}
	}

	return k8sCtx, nil
}

// buildRootCauseInput 构建根因分析输入.
func (a *ReasoningAgent) buildRootCauseInput(input *AnalysisInput) *root_cause.AnalysisInput {
	rcInput := &root_cause.AnalysisInput{
		FailureType:    input.FailureType,
		ResourceType:   input.ResourceType,
		ResourceName:   input.ResourceName,
		Namespace:      input.Namespace,
		ClusterID:      input.ClusterID,
		ErrorMessage:   input.ErrorMessage,
		Timestamp:      input.Timestamp,
		ResourceStatus: input.ResourceStatus,
	}

	// 使用已获取的 K8s 上下文或提供的数据
	if input.PodLogs != "" {
		rcInput.PodLogs = input.PodLogs
	}

	if len(input.Events) > 0 {
		rcInput.PodEvents = convertEvents(input.Events)
	}

	if input.Metrics != nil {
		rcInput.Metrics = convertMetrics(input.Metrics)
	}

	return rcInput
}

// buildDescriptionInput 构建描述生成输入.
func (a *ReasoningAgent) buildDescriptionInput(input *AnalysisInput) *description.DescriptionInput {
	descInput := &description.DescriptionInput{
		FailureType:     input.FailureType,
		ResourceType:    input.ResourceType,
		ResourceName:    input.ResourceName,
		Namespace:       input.Namespace,
		ClusterID:       input.ClusterID,
		ErrorMessage:    input.ErrorMessage,
		Timestamp:       input.Timestamp,
		Language:        input.Language,
		DetailLevel:     input.DetailLevel,
		IncludeTimeline: true,
	}

	// 使用已获取的 K8s 上下文或提供的数据
	if input.PodLogs != "" {
		descInput.PodLogs = input.PodLogs
	}

	if len(input.Events) > 0 {
		descInput.PodEvents = convertEventsToDesc(input.Events)
	}

	if input.Metrics != nil {
		descInput.Metrics = convertMetrics(input.Metrics)
	}

	descInput.ResourceStatus = input.ResourceStatus

	return descInput
}

// applyDefaults 应用默认值.
func (a *ReasoningAgent) applyDefaults(input *AnalysisInput) {
	if input.Language == "" {
		input.Language = a.config.DefaultLanguage
	}
	if input.DetailLevel == "" {
		input.DetailLevel = a.config.DefaultDetailLevel
	}
	if input.Timestamp.IsZero() {
		input.Timestamp = time.Now()
	}
}

// convertEvents 转换事件格式.
func convertEvents(events []k8s_tool.EventInfo) []root_cause.K8sEvent {
	result := make([]root_cause.K8sEvent, len(events))
	for i, e := range events {
		result[i] = root_cause.K8sEvent{
			Type:      e.Type,
			Reason:    e.Reason,
			Message:   e.Message,
			Timestamp: e.LastTimestamp,
			Source:    e.Source,
		}
	}
	return result
}

// convertEventsToDesc 转换事件格式到描述.
func convertEventsToDesc(events []k8s_tool.EventInfo) []description.PodEvent {
	result := make([]description.PodEvent, len(events))
	for i, e := range events {
		result[i] = description.PodEvent{
			Type:      e.Type,
			Reason:    e.Reason,
			Message:   e.Message,
			Timestamp: e.LastTimestamp,
			Source:    e.Source,
		}
	}
	return result
}

// convertMetrics 转换指标格式.
func convertMetrics(metrics *k8s_tool.MetricsInfo) map[string]float64 {
	return map[string]float64{
		"cpu_utilization":     metrics.CPU.Utilization,
		"memory_utilization":  metrics.Memory.Utilization,
		"storage_utilization": metrics.Storage.Utilization,
		"network_rx_bytes":    float64(metrics.Network.RxBytes),
		"network_tx_bytes":    float64(metrics.Network.TxBytes),
	}
}

// validateConfig 验证配置.
func validateConfig(config *AgentConfig) error {
	if config.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	// 至少启用一个分析功能
	if !config.EnableRootCause && !config.EnableDescription {
		return fmt.Errorf("at least one analysis feature must be enabled")
	}

	return nil
}

// DefaultAgentConfig 返回默认配置.
func DefaultAgentConfig() *AgentConfig {
	return &AgentConfig{
		RootCauseConfig:    root_cause.DefaultChainConfig(),
		DescriptionConfig:  description.DefaultChainConfig(),
		K8sToolConfig:      k8s_tool.DefaultToolConfig(),
		EnableK8sTool:      true,
		EnableRootCause:    true,
		EnableDescription:  true,
		DefaultLanguage:    "en",
		DefaultDetailLevel: "normal",
		Timeout:            60 * time.Second,
	}
}
