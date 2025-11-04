package reasoning

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kart-io/k8s-agent/internal/reasoning/agents/k8s_tool"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/description"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/root_cause"
)

func TestDefaultAgentConfig(t *testing.T) {
	config := DefaultAgentConfig()

	if config == nil {
		t.Fatal("DefaultAgentConfig returned nil")
	}

	if !config.EnableK8sTool {
		t.Error("EnableK8sTool should be true by default")
	}

	if !config.EnableRootCause {
		t.Error("EnableRootCause should be true by default")
	}

	if !config.EnableDescription {
		t.Error("EnableDescription should be true by default")
	}

	if config.DefaultLanguage != "en" {
		t.Errorf("DefaultLanguage = %s, want 'en'", config.DefaultLanguage)
	}

	if config.Timeout != 60*time.Second {
		t.Errorf("Timeout = %v, want 60s", config.Timeout)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *AgentConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &AgentConfig{
				EnableRootCause:   true,
				EnableDescription: false,
				Timeout:           30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "negative timeout",
			config: &AgentConfig{
				EnableRootCause: true,
				Timeout:         -1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "no features enabled",
			config: &AgentConfig{
				EnableRootCause:   false,
				EnableDescription: false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)

			if tt.wantErr && err == nil {
				t.Error("validateConfig() expected error but got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("validateConfig() unexpected error: %v", err)
			}
		})
	}
}

func TestNewReasoningAgent(t *testing.T) {
	// 创建必需的组件（使用 nil 作为占位符）
	tests := []struct {
		name             string
		rootCauseChain   *root_cause.RootCauseChain
		descriptionChain *description.DescriptionChain
		k8sTool          *k8s_tool.K8sTool
		config           *AgentConfig
		wantErr          bool
	}{
		{
			name:             "nil config uses default",
			rootCauseChain:   &root_cause.RootCauseChain{},
			descriptionChain: &description.DescriptionChain{},
			k8sTool:          &k8s_tool.K8sTool{},
			config:           nil,
			wantErr:          false,
		},
		{
			name:             "missing root cause chain",
			rootCauseChain:   nil,
			descriptionChain: &description.DescriptionChain{},
			k8sTool:          &k8s_tool.K8sTool{},
			config: &AgentConfig{
				EnableRootCause:   true,
				EnableDescription: true,
			},
			wantErr: true,
		},
		{
			name:             "missing description chain",
			rootCauseChain:   &root_cause.RootCauseChain{},
			descriptionChain: nil,
			k8sTool:          &k8s_tool.K8sTool{},
			config: &AgentConfig{
				EnableRootCause:   true,
				EnableDescription: true,
			},
			wantErr: true,
		},
		{
			name:             "missing k8s tool",
			rootCauseChain:   &root_cause.RootCauseChain{},
			descriptionChain: &description.DescriptionChain{},
			k8sTool:          nil,
			config: &AgentConfig{
				EnableK8sTool:     true,
				EnableRootCause:   true,
				EnableDescription: false,
			},
			wantErr: true,
		},
		{
			name:             "invalid config",
			rootCauseChain:   &root_cause.RootCauseChain{},
			descriptionChain: &description.DescriptionChain{},
			k8sTool:          &k8s_tool.K8sTool{},
			config: &AgentConfig{
				Timeout: -1 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := NewReasoningAgent(
				tt.rootCauseChain,
				tt.descriptionChain,
				tt.k8sTool,
				tt.config,
			)

			if tt.wantErr {
				if err == nil {
					t.Error("NewReasoningAgent() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewReasoningAgent() unexpected error: %v", err)
				return
			}

			if agent == nil {
				t.Error("NewReasoningAgent() returned nil agent")
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	agent := &ReasoningAgent{
		config: DefaultAgentConfig(),
	}

	input := &AnalysisInput{
		ResourceName: "test-pod",
	}

	agent.applyDefaults(input)

	if input.Language != "en" {
		t.Errorf("Language = %s, want 'en'", input.Language)
	}

	if input.DetailLevel != "normal" {
		t.Errorf("DetailLevel = %s, want 'normal'", input.DetailLevel)
	}

	if input.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero after applying defaults")
	}
}

func TestConvertEvents(t *testing.T) {
	events := []k8s_tool.EventInfo{
		{
			Type:          "Normal",
			Reason:        "Scheduled",
			Message:       "Pod scheduled",
			LastTimestamp: time.Now(),
			Source:        "scheduler",
		},
		{
			Type:          "Warning",
			Reason:        "Failed",
			Message:       "Failed to pull image",
			LastTimestamp: time.Now(),
			Source:        "kubelet",
		},
	}

	converted := convertEvents(events)

	if len(converted) != len(events) {
		t.Errorf("Converted events count = %d, want %d", len(converted), len(events))
	}

	for i, e := range converted {
		if e.Type != events[i].Type {
			t.Errorf("Event %d Type = %s, want %s", i, e.Type, events[i].Type)
		}
		if e.Reason != events[i].Reason {
			t.Errorf("Event %d Reason = %s, want %s", i, e.Reason, events[i].Reason)
		}
		if e.Message != events[i].Message {
			t.Errorf("Event %d Message = %s, want %s", i, e.Message, events[i].Message)
		}
	}
}

func TestConvertEventsToDesc(t *testing.T) {
	events := []k8s_tool.EventInfo{
		{
			Type:          "Normal",
			Reason:        "Started",
			Message:       "Container started",
			LastTimestamp: time.Now(),
			Source:        "kubelet",
		},
	}

	converted := convertEventsToDesc(events)

	if len(converted) != len(events) {
		t.Errorf("Converted events count = %d, want %d", len(converted), len(events))
	}

	if converted[0].Type != events[0].Type {
		t.Errorf("Type = %s, want %s", converted[0].Type, events[0].Type)
	}
}

func TestConvertMetrics(t *testing.T) {
	metrics := &k8s_tool.MetricsInfo{
		CPU: k8s_tool.ResourceMetric{
			Utilization: 30.5,
		},
		Memory: k8s_tool.ResourceMetric{
			Utilization: 50.2,
		},
		Storage: k8s_tool.ResourceMetric{
			Utilization: 10.1,
		},
		Network: k8s_tool.NetworkMetric{
			RxBytes: 1024,
			TxBytes: 2048,
		},
	}

	converted := convertMetrics(metrics)

	if converted["cpu_utilization"] != 30.5 {
		t.Errorf("cpu_utilization = %.2f, want 30.5", converted["cpu_utilization"])
	}

	if converted["memory_utilization"] != 50.2 {
		t.Errorf("memory_utilization = %.2f, want 50.2", converted["memory_utilization"])
	}

	if converted["network_rx_bytes"] != 1024.0 {
		t.Errorf("network_rx_bytes = %.2f, want 1024.0", converted["network_rx_bytes"])
	}
}

func TestAnalysisError(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	analysisErr := &AnalysisError{
		Phase:   PhaseRootCause,
		Message: "root cause failed",
		Err:     originalErr,
	}

	errMsg := analysisErr.Error()
	expectedMsg := "root cause failed: original error"
	if errMsg != expectedMsg {
		t.Errorf("Error() = %s, want %s", errMsg, expectedMsg)
	}

	unwrapped := analysisErr.Unwrap()
	if unwrapped != originalErr {
		t.Error("Unwrap() should return original error")
	}
}

func TestBuildRootCauseInput(t *testing.T) {
	agent := &ReasoningAgent{
		config: DefaultAgentConfig(),
	}

	now := time.Now()
	input := &AnalysisInput{
		FailureType:  "pod_failure",
		ResourceType: "pod",
		ResourceName: "test-pod",
		Namespace:    "default",
		ClusterID:    "cluster-1",
		ErrorMessage: "OOMKilled",
		Timestamp:    now,
		PodLogs:      "ERROR: Out of memory",
		ResourceStatus: map[string]string{
			"phase": "Failed",
		},
	}

	rcInput := agent.buildRootCauseInput(input)

	if rcInput.FailureType != input.FailureType {
		t.Errorf("FailureType = %s, want %s", rcInput.FailureType, input.FailureType)
	}

	if rcInput.ResourceName != input.ResourceName {
		t.Errorf("ResourceName = %s, want %s", rcInput.ResourceName, input.ResourceName)
	}

	if rcInput.PodLogs != input.PodLogs {
		t.Errorf("PodLogs = %s, want %s", rcInput.PodLogs, input.PodLogs)
	}

	if rcInput.ResourceStatus["phase"] != "Failed" {
		t.Error("ResourceStatus not properly copied")
	}
}

func TestBuildDescriptionInput(t *testing.T) {
	agent := &ReasoningAgent{
		config: DefaultAgentConfig(),
	}

	now := time.Now()
	input := &AnalysisInput{
		FailureType:  "pod_failure",
		ResourceType: "pod",
		ResourceName: "test-pod",
		Namespace:    "default",
		ClusterID:    "cluster-1",
		ErrorMessage: "OOMKilled",
		Timestamp:    now,
		Language:     "zh",
		DetailLevel:  "detailed",
		PodLogs:      "ERROR: Out of memory",
	}

	descInput := agent.buildDescriptionInput(input)

	if descInput.FailureType != input.FailureType {
		t.Errorf("FailureType = %s, want %s", descInput.FailureType, input.FailureType)
	}

	if descInput.Language != input.Language {
		t.Errorf("Language = %s, want %s", descInput.Language, input.Language)
	}

	if descInput.DetailLevel != input.DetailLevel {
		t.Errorf("DetailLevel = %s, want %s", descInput.DetailLevel, input.DetailLevel)
	}

	if !descInput.IncludeTimeline {
		t.Error("IncludeTimeline should be true")
	}
}

// 由于完整的集成测试需要真实的 LLM Proxy 和所有组件，
// 这里只测试核心逻辑和数据转换
// 完整的集成测试应该在 tests/integration/ 中进行

func TestAnalyzeRootCause_Disabled(t *testing.T) {
	agent := &ReasoningAgent{
		config: &AgentConfig{
			EnableRootCause:   false,
			EnableDescription: true,
		},
	}

	ctx := context.Background()
	input := &AnalysisInput{
		ResourceName: "test-pod",
	}

	_, err := agent.AnalyzeRootCause(ctx, input)
	if err == nil {
		t.Error("AnalyzeRootCause() should return error when disabled")
	}
}

func TestGenerateDescription_Disabled(t *testing.T) {
	agent := &ReasoningAgent{
		config: &AgentConfig{
			EnableRootCause:   true,
			EnableDescription: false,
		},
	}

	ctx := context.Background()
	input := &AnalysisInput{
		ResourceName: "test-pod",
	}

	_, err := agent.GenerateDescription(ctx, input)
	if err == nil {
		t.Error("GenerateDescription() should return error when disabled")
	}
}
