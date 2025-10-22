package k8s_tool

import (
	"context"
	"testing"
	"time"
)

func TestDefaultToolConfig(t *testing.T) {
	config := DefaultToolConfig()

	if config == nil {
		t.Fatal("DefaultToolConfig returned nil")
	}

	if config.DefaultNamespace != "default" {
		t.Errorf("Expected default_namespace 'default', got '%s'", config.DefaultNamespace)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", config.Timeout)
	}

	if config.MaxLogLines != 100 {
		t.Errorf("Expected max_log_lines 100, got %d", config.MaxLogLines)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *ToolConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &ToolConfig{
				Timeout:     30 * time.Second,
				MaxLogLines: 100,
			},
			wantErr: false,
		},
		{
			name: "negative timeout",
			config: &ToolConfig{
				Timeout:     -1 * time.Second,
				MaxLogLines: 100,
			},
			wantErr: true,
		},
		{
			name: "negative max_log_lines",
			config: &ToolConfig{
				Timeout:     30 * time.Second,
				MaxLogLines: -1,
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

func TestValidateInput(t *testing.T) {
	tests := []struct {
		name    string
		input   *ToolInput
		wantErr bool
	}{
		{
			name: "valid get pod input",
			input: &ToolInput{
				Action:       "get",
				ResourceType: "pod",
				ResourceName: "test-pod",
			},
			wantErr: false,
		},
		{
			name: "valid list pods input",
			input: &ToolInput{
				Action:       "list",
				ResourceType: "pod",
			},
			wantErr: false,
		},
		{
			name: "missing action",
			input: &ToolInput{
				ResourceType: "pod",
			},
			wantErr: true,
		},
		{
			name: "unsupported action",
			input: &ToolInput{
				Action:       "delete",
				ResourceType: "pod",
			},
			wantErr: true,
		},
		{
			name: "missing resource_type for get",
			input: &ToolInput{
				Action: "get",
			},
			wantErr: true,
		},
		{
			name: "unsupported resource_type",
			input: &ToolInput{
				Action:       "get",
				ResourceType: "unknown",
			},
			wantErr: true,
		},
		{
			name: "missing resource_name for get",
			input: &ToolInput{
				Action:       "get",
				ResourceType: "pod",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInput(tt.input)

			if tt.wantErr && err == nil {
				t.Error("validateInput() expected error but got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("validateInput() unexpected error: %v", err)
			}
		})
	}
}

func TestNewK8sTool(t *testing.T) {
	tests := []struct {
		name    string
		config  *ToolConfig
		wantErr bool
	}{
		{
			name:    "nil config uses default",
			config:  nil,
			wantErr: false,
		},
		{
			name:    "valid config",
			config:  DefaultToolConfig(),
			wantErr: false,
		},
		{
			name: "invalid config",
			config: &ToolConfig{
				Timeout: -1 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, err := NewK8sTool(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Error("NewK8sTool() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewK8sTool() unexpected error: %v", err)
				return
			}

			if tool == nil {
				t.Error("NewK8sTool() returned nil tool")
				return
			}

			if tool.Name() != "k8s_tool" {
				t.Errorf("Name() = %s, want 'k8s_tool'", tool.Name())
			}

			if tool.Description() == "" {
				t.Error("Description() returned empty string")
			}
		})
	}
}

func TestExecute_GetPod(t *testing.T) {
	tool, err := NewK8sTool(nil)
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	ctx := context.Background()
	input := &ToolInput{
		Action:       "get",
		ResourceType: "pod",
		ResourceName: "test-pod",
		Namespace:    "default",
	}

	output, err := tool.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if output == nil {
		t.Fatal("Execute() returned nil output")
	}

	if !output.Success {
		t.Errorf("Execute() failed: %s", output.ErrorMsg)
	}

	if output.Data == nil {
		t.Error("Execute() returned nil data")
	}

	if output.ToolName != "k8s_tool" {
		t.Errorf("ToolName = %s, want 'k8s_tool'", output.ToolName)
	}

	if output.Action != "get" {
		t.Errorf("Action = %s, want 'get'", output.Action)
	}

	// 验证返回的 Pod 信息
	podInfo, ok := output.Data.(*PodInfo)
	if !ok {
		t.Fatal("Data is not PodInfo type")
	}

	if podInfo.Name != "test-pod" {
		t.Errorf("Pod name = %s, want 'test-pod'", podInfo.Name)
	}

	if podInfo.Namespace != "default" {
		t.Errorf("Pod namespace = %s, want 'default'", podInfo.Namespace)
	}
}

func TestExecute_ListPods(t *testing.T) {
	tool, err := NewK8sTool(nil)
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	ctx := context.Background()
	input := &ToolInput{
		Action:       "list",
		ResourceType: "pod",
		Namespace:    "default",
	}

	output, err := tool.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if !output.Success {
		t.Errorf("Execute() failed: %s", output.ErrorMsg)
	}

	// 验证返回的 Pod 列表
	podList, ok := output.Data.([]PodInfo)
	if !ok {
		t.Fatal("Data is not []PodInfo type")
	}

	if len(podList) == 0 {
		t.Error("Pod list is empty")
	}

	t.Logf("Listed %d pods", len(podList))
}

func TestExecute_GetLogs(t *testing.T) {
	tool, err := NewK8sTool(nil)
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	ctx := context.Background()
	input := &ToolInput{
		Action:       "logs",
		ResourceType: "pod",
		ResourceName: "test-pod",
		Namespace:    "default",
		Parameters: map[string]string{
			"timestamps": "true",
		},
	}

	output, err := tool.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if !output.Success {
		t.Errorf("Execute() failed: %s", output.ErrorMsg)
	}

	// 验证返回的日志
	logData, ok := output.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Data is not map type")
	}

	if logData["logs"] == nil {
		t.Error("Logs field is missing")
	}

	t.Logf("Retrieved logs: %v", logData["logs"])
}

func TestExecute_GetEvents(t *testing.T) {
	tool, err := NewK8sTool(nil)
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	ctx := context.Background()
	input := &ToolInput{
		Action:       "events",
		ResourceType: "pod",
		ResourceName: "test-pod",
		Namespace:    "default",
	}

	output, err := tool.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if !output.Success {
		t.Errorf("Execute() failed: %s", output.ErrorMsg)
	}

	// 验证返回的事件
	events, ok := output.Data.([]EventInfo)
	if !ok {
		t.Fatal("Data is not []EventInfo type")
	}

	if len(events) == 0 {
		t.Error("Events list is empty")
	}

	t.Logf("Retrieved %d events", len(events))
}

func TestExecute_GetMetrics(t *testing.T) {
	tool, err := NewK8sTool(nil)
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	ctx := context.Background()
	input := &ToolInput{
		Action:       "metrics",
		ResourceType: "pod",
		ResourceName: "test-pod",
		Namespace:    "default",
	}

	output, err := tool.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if !output.Success {
		t.Errorf("Execute() failed: %s", output.ErrorMsg)
	}

	// 验证返回的指标
	metrics, ok := output.Data.(*MetricsInfo)
	if !ok {
		t.Fatal("Data is not *MetricsInfo type")
	}

	if metrics.ResourceName != "test-pod" {
		t.Errorf("Resource name = %s, want 'test-pod'", metrics.ResourceName)
	}

	if metrics.CPU.Utilization <= 0 {
		t.Error("CPU utilization should be positive")
	}

	t.Logf("CPU utilization: %.2f%%", metrics.CPU.Utilization)
	t.Logf("Memory utilization: %.2f%%", metrics.Memory.Utilization)
}

func TestExecute_InvalidAction(t *testing.T) {
	tool, err := NewK8sTool(nil)
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	ctx := context.Background()
	input := &ToolInput{
		Action:       "invalid",
		ResourceType: "pod",
		ResourceName: "test-pod",
	}

	output, err := tool.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if output.Success {
		t.Error("Execute() should fail for invalid action")
	}

	if output.ErrorMsg == "" {
		t.Error("ErrorMsg should not be empty for failed execution")
	}

	t.Logf("Error message: %s", output.ErrorMsg)
}

func TestExecute_NilInput(t *testing.T) {
	tool, err := NewK8sTool(nil)
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	ctx := context.Background()

	_, err = tool.Execute(ctx, nil)
	if err == nil {
		t.Error("Execute() should return error for nil input")
	}
}

func TestParseLogsOptions(t *testing.T) {
	tool, err := NewK8sTool(nil)
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	tests := []struct {
		name   string
		params map[string]string
		check  func(*LogsOptions) error
	}{
		{
			name:   "empty params",
			params: map[string]string{},
			check: func(opts *LogsOptions) error {
				if opts.Container != "" {
					t.Errorf("Container should be empty, got '%s'", opts.Container)
				}
				return nil
			},
		},
		{
			name: "with container",
			params: map[string]string{
				"container": "main",
			},
			check: func(opts *LogsOptions) error {
				if opts.Container != "main" {
					t.Errorf("Container = '%s', want 'main'", opts.Container)
				}
				return nil
			},
		},
		{
			name: "with previous",
			params: map[string]string{
				"previous": "true",
			},
			check: func(opts *LogsOptions) error {
				if !opts.Previous {
					t.Error("Previous should be true")
				}
				return nil
			},
		},
		{
			name: "with timestamps",
			params: map[string]string{
				"timestamps": "true",
			},
			check: func(opts *LogsOptions) error {
				if !opts.Timestamps {
					t.Error("Timestamps should be true")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tool.parseLogsOptions(tt.params)
			if err := tt.check(opts); err != nil {
				t.Error(err)
			}
		})
	}
}
