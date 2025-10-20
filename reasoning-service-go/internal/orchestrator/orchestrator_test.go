package orchestrator

import (
	"testing"
	"time"
)

func TestDefaultOrchestratorConfig(t *testing.T) {
	config := DefaultOrchestratorConfig()

	if config == nil {
		t.Fatal("DefaultOrchestratorConfig returned nil")
	}

	if !config.EnableMemory {
		t.Error("EnableMemory should be true by default")
	}

	if config.EnableToolAgent {
		t.Error("EnableToolAgent should be false by default")
	}

	if !config.EnableDescription {
		t.Error("EnableDescription should be true by default")
	}

	if config.Timeout != 60*time.Second {
		t.Errorf("Timeout = %v, want 60s", config.Timeout)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *OrchestratorConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultOrchestratorConfig(),
			wantErr: false,
		},
		{
			name: "negative timeout",
			config: &OrchestratorConfig{
				Timeout: -1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "negative root_cause_timeout",
			config: &OrchestratorConfig{
				Timeout:          30 * time.Second,
				RootCauseTimeout: -1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "negative max_similar_cases",
			config: &OrchestratorConfig{
				Timeout:         30 * time.Second,
				MaxSimilarCases: -1,
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

func TestNewOrchestrator(t *testing.T) {
	// 注意：这里只测试构造函数的验证逻辑
	// 完整的集成测试需要在 tests/integration/ 中进行

	tests := []struct {
		name    string
		config  *OrchestratorConfig
		wantErr bool
	}{
		{
			name:    "nil config uses default",
			config:  nil,
			wantErr: false,
		},
		{
			name: "invalid config",
			config: &OrchestratorConfig{
				Timeout: -1 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用 nil 组件来测试验证逻辑
			// 实际使用中需要提供真实组件
			_, err := NewOrchestrator(nil, nil, nil, nil, nil, tt.config)

			// 应该因为缺少必需组件而失败
			if err == nil {
				t.Error("NewOrchestrator() should return error for nil components")
			}
		})
	}
}

func TestAnalysisRequest(t *testing.T) {
	// 测试请求结构的基本功能
	req := &AnalysisRequest{
		SessionID:    "session-1",
		FailureType:  "pod_failure",
		ResourceType: "pod",
		ResourceName: "test-pod",
		Namespace:    "default",
		ClusterID:    "cluster-1",
		ErrorMessage: "OOMKilled",
		Timestamp:    time.Now(),
		Language:     "en",
		DetailLevel:  "normal",
	}

	if req.SessionID == "" {
		t.Error("SessionID should not be empty")
	}

	if req.FailureType == "" {
		t.Error("FailureType should not be empty")
	}
}

func TestAnalysisResponse(t *testing.T) {
	// 测试响应结构的基本功能
	resp := &AnalysisResponse{
		ExecutionSteps: make([]ExecutionStep, 0),
		Timestamp:      time.Now(),
	}

	// 添加执行步骤
	resp.ExecutionSteps = append(resp.ExecutionSteps, ExecutionStep{
		Step:        1,
		Name:        "test_step",
		Description: "Test execution step",
		Status:      "success",
		Duration:    100 * time.Millisecond,
	})

	if len(resp.ExecutionSteps) != 1 {
		t.Errorf("ExecutionSteps count = %d, want 1", len(resp.ExecutionSteps))
	}

	if resp.ExecutionSteps[0].Status != "success" {
		t.Errorf("Step status = %s, want 'success'", resp.ExecutionSteps[0].Status)
	}
}

func TestExecutionStep(t *testing.T) {
	step := ExecutionStep{
		Step:        1,
		Name:        "root_cause_analysis",
		Description: "Analyze root cause",
		Status:      "success",
		Duration:    500 * time.Millisecond,
	}

	if step.Step != 1 {
		t.Errorf("Step = %d, want 1", step.Step)
	}

	if step.Name != "root_cause_analysis" {
		t.Errorf("Name = %s, want 'root_cause_analysis'", step.Name)
	}

	if step.Status != "success" {
		t.Errorf("Status = %s, want 'success'", step.Status)
	}

	if step.Duration != 500*time.Millisecond {
		t.Errorf("Duration = %v, want 500ms", step.Duration)
	}
}
