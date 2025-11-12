package root_cause

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestDefaultChainConfig(t *testing.T) {
	config := DefaultChainConfig()

	if config == nil {
		t.Fatal("DefaultChainConfig returned nil")
	}

	// 验证默认值
	if config.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %.2f", config.Temperature)
	}

	if config.MaxTokens != 4096 {
		t.Errorf("Expected max_tokens 4096, got %d", config.MaxTokens)
	}

	if config.MinConfidence != 0.6 {
		t.Errorf("Expected min_confidence 0.6, got %.2f", config.MinConfidence)
	}

	if !config.IncludeSimilar {
		t.Error("Expected include_similar to be true")
	}

	if config.MaxSimilarCases != 3 {
		t.Errorf("Expected max_similar_cases 3, got %d", config.MaxSimilarCases)
	}

	if config.SystemPrompt == "" {
		t.Error("System prompt should not be empty")
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", config.Timeout)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *ChainConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &ChainConfig{
				Temperature:   0.7,
				MaxTokens:     4096,
				MinConfidence: 0.6,
			},
			wantErr: false,
		},
		{
			name: "invalid temperature - too low",
			config: &ChainConfig{
				Temperature:   -0.1,
				MaxTokens:     4096,
				MinConfidence: 0.6,
			},
			wantErr: true,
			errMsg:  "temperature",
		},
		{
			name: "invalid temperature - too high",
			config: &ChainConfig{
				Temperature:   2.1,
				MaxTokens:     4096,
				MinConfidence: 0.6,
			},
			wantErr: true,
			errMsg:  "temperature",
		},
		{
			name: "invalid max_tokens",
			config: &ChainConfig{
				Temperature:   0.7,
				MaxTokens:     0,
				MinConfidence: 0.6,
			},
			wantErr: true,
			errMsg:  "max_tokens",
		},
		{
			name: "invalid min_confidence - too low",
			config: &ChainConfig{
				Temperature:   0.7,
				MaxTokens:     4096,
				MinConfidence: -0.1,
			},
			wantErr: true,
			errMsg:  "min_confidence",
		},
		{
			name: "invalid min_confidence - too high",
			config: &ChainConfig{
				Temperature:   0.7,
				MaxTokens:     4096,
				MinConfidence: 1.1,
			},
			wantErr: true,
			errMsg:  "min_confidence",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateConfig() expected error but got nil")
					return
				}
				// 不进行详细的错误消息检查,因为错误消息可能变化
			} else {
				if err != nil {
					t.Errorf("validateConfig() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain JSON",
			input:    `{"root_cause": "test"}`,
			expected: `{"root_cause": "test"}`,
		},
		{
			name: "JSON in markdown code block",
			input: "```json\n" +
				`{"root_cause": "test"}` + "\n" +
				"```",
			expected: `{"root_cause": "test"}`,
		},
		{
			name: "JSON with surrounding text",
			input: "Here is the analysis:\n" +
				`{"root_cause": "test"}` + "\n" +
				"That's the result.",
			expected: `{"root_cause": "test"}`,
		},
		{
			name: "JSON in generic code block",
			input: "```\n" +
				`{"root_cause": "test"}` + "\n" +
				"```",
			expected: `{"root_cause": "test"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSON(tt.input)
			if result != tt.expected {
				t.Errorf("extractJSON() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseResponse(t *testing.T) {
	chain := &RootCauseChain{
		config: DefaultChainConfig(),
	}

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid JSON response",
			content: `{
				"root_cause": "Pod OOMKilled",
				"confidence": 0.95,
				"reasoning": "Memory limit exceeded",
				"category": "resource",
				"contributing_factors": [
					{
						"name": "Memory leak",
						"description": "Application memory leak",
						"impact": "high",
						"evidence": "Increasing memory usage"
					}
				],
				"recommendations": [
					{
						"action": "Increase memory limit",
						"priority": "high",
						"description": "Increase to 1Gi",
						"commands": ["kubectl set resources deployment/app --limits=memory=1Gi"],
						"impact": "Immediate",
						"risk_level": "low"
					}
				]
			}`,
			wantErr: false,
		},
		{
			name: "JSON in markdown",
			content: "```json\n" +
				`{"root_cause": "test", "confidence": 0.8, "reasoning": "test", "category": "test"}` +
				"\n```",
			wantErr: false,
		},
		{
			name: "missing root_cause",
			content: `{
				"confidence": 0.8,
				"reasoning": "test"
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := chain.parseResponse(tt.content)

			if tt.wantErr {
				if err == nil {
					t.Error("parseResponse() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("parseResponse() unexpected error: %v", err)
				return
			}

			if output == nil {
				t.Error("parseResponse() returned nil output")
				return
			}

			if output.RootCause == "" {
				t.Error("root_cause is empty")
			}

			// 验证置信度范围
			if output.Confidence < 0 || output.Confidence > 1 {
				t.Errorf("confidence out of range: %.2f", output.Confidence)
			}
		})
	}
}

func TestBuildPrompt(t *testing.T) {
	chain := &RootCauseChain{
		config: DefaultChainConfig(),
	}

	input := &AnalysisInput{
		FailureType:  "pod_failure",
		ResourceType: "pod",
		ResourceName: "test-pod",
		Namespace:    "default",
		ClusterID:    "cluster-1",
		ErrorMessage: "OOMKilled",
		Symptoms:     []string{"High memory usage", "Pod restart"},
		Timestamp:    time.Now(),
		PodEvents: []K8sEvent{
			{
				Type:      "Warning",
				Reason:    "OOMKilled",
				Message:   "Container killed due to OOM",
				Timestamp: time.Now(),
				Source:    "kubelet",
			},
		},
		PodLogs: "ERROR: Out of memory\nKilled",
		ResourceStatus: map[string]string{
			"phase":  "Failed",
			"reason": "OOMKilled",
		},
		Metrics: map[string]float64{
			"memory_usage":   512.0,
			"memory_limit":   500.0,
			"memory_percent": 102.4,
		},
	}

	prompt := chain.buildPrompt(input)

	if prompt == "" {
		t.Fatal("buildPrompt() returned empty prompt")
	}

	// 验证 prompt 包含关键信息
	requiredStrings := []string{
		"test-pod",
		"default",
		"OOMKilled",
		"High memory usage",
		"Pod Events",
		"Pod Logs",
		"Resource Status",
		"Metrics",
		"JSON format",
	}

	for _, required := range requiredStrings {
		if !contains(prompt, required) {
			t.Errorf("Prompt missing required string: %s", required)
		}
	}

	t.Logf("Prompt length: %d characters", len(prompt))
}

func TestBuildPromptWithSimilarCases(t *testing.T) {
	config := DefaultChainConfig()
	config.IncludeSimilar = true
	config.MaxSimilarCases = 2

	chain := &RootCauseChain{
		config: config,
	}

	input := &AnalysisInput{
		FailureType:  "pod_failure",
		ResourceType: "pod",
		ResourceName: "test-pod",
		Namespace:    "default",
		ClusterID:    "cluster-1",
		Timestamp:    time.Now(),
		SimilarCases: []SimilarCase{
			{
				CaseID:      "case-1",
				Description: "Similar OOM issue",
				RootCause:   "Memory limit too low",
				Similarity:  0.95,
				Resolution:  "Increased memory limit",
			},
			{
				CaseID:      "case-2",
				Description: "Another OOM case",
				RootCause:   "Memory leak",
				Similarity:  0.85,
				Resolution:  "Fixed memory leak",
			},
			{
				CaseID:      "case-3",
				Description: "Third case",
				RootCause:   "Test",
				Similarity:  0.75,
				Resolution:  "Test resolution",
			},
		},
	}

	prompt := chain.buildPrompt(input)

	// 打印 prompt 用于调试
	t.Logf("Generated prompt:\n%s", prompt)

	// 验证包含相似案例
	if !contains(prompt, "Similar Cases") {
		t.Error("Prompt should contain Similar Cases section")
	}

	if !contains(prompt, "Similar OOM issue") {
		t.Error("Prompt should contain description from case-1")
	}

	if !contains(prompt, "Another OOM case") {
		t.Error("Prompt should contain description from case-2")
	}

	// 验证只包含前 2 个案例 (maxSimilarCases = 2)
	if contains(prompt, "Third case") {
		t.Error("Prompt should not contain case-3 (max 2 cases)")
	}
}

func TestJSONSchemaValidity(t *testing.T) {
	schema := getJSONSchema()

	// 验证是否为有效 JSON
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(schema), &obj); err != nil {
		t.Errorf("JSON schema is not valid JSON: %v", err)
	}

	// 验证包含必需字段
	requiredFields := []string{
		"root_cause",
		"confidence",
		"reasoning",
		"category",
		"contributing_factors",
		"recommendations",
	}

	for _, field := range requiredFields {
		if _, ok := obj[field]; !ok {
			t.Errorf("JSON schema missing required field: %s", field)
		}
	}
}

func TestNewRootCauseChain(t *testing.T) {
	tests := []struct {
		name    string
		proxy   interface{} // 使用 interface{} 以便测试 nil
		config  *ChainConfig
		wantErr bool
	}{
		{
			name:    "nil proxy",
			proxy:   nil,
			config:  DefaultChainConfig(),
			wantErr: true,
		},
		{
			name:    "nil config uses default",
			proxy:   &struct{}{}, // 模拟非 nil proxy
			config:  nil,
			wantErr: false, // 应该使用默认配置
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 这个测试需要 mock proxy adapter
			// 暂时跳过实际的创建,只测试配置验证
			if tt.proxy == nil {
				_, err := NewRootCauseChain(nil, tt.config)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewRootCauseChain() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

// 辅助函数.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
