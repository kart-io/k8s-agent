package description

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

	if config.Temperature != 0.5 {
		t.Errorf("Expected temperature 0.5, got %.2f", config.Temperature)
	}

	if config.MaxTokens != 2048 {
		t.Errorf("Expected max_tokens 2048, got %d", config.MaxTokens)
	}

	if config.DefaultLanguage != "en" {
		t.Errorf("Expected default_language 'en', got '%s'", config.DefaultLanguage)
	}

	if config.DefaultDetailLevel != "normal" {
		t.Errorf("Expected default_detail_level 'normal', got '%s'", config.DefaultDetailLevel)
	}

	if !config.IncludeTimeline {
		t.Error("Expected include_timeline to be true")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *ChainConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &ChainConfig{
				Temperature:        0.5,
				MaxTokens:          2048,
				DefaultLanguage:    "en",
				DefaultDetailLevel: "normal",
			},
			wantErr: false,
		},
		{
			name: "invalid temperature",
			config: &ChainConfig{
				Temperature: 3.0,
				MaxTokens:   2048,
			},
			wantErr: true,
		},
		{
			name: "invalid max_tokens",
			config: &ChainConfig{
				Temperature: 0.5,
				MaxTokens:   0,
			},
			wantErr: true,
		},
		{
			name: "unsupported language",
			config: &ChainConfig{
				Temperature:     0.5,
				MaxTokens:       2048,
				DefaultLanguage: "fr",
			},
			wantErr: true,
		},
		{
			name: "unsupported detail level",
			config: &ChainConfig{
				Temperature:        0.5,
				MaxTokens:          2048,
				DefaultDetailLevel: "extreme",
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
		input   *DescriptionInput
		wantErr bool
	}{
		{
			name: "valid input with zh language",
			input: &DescriptionInput{
				Language:    "zh",
				DetailLevel: "normal",
			},
			wantErr: false,
		},
		{
			name: "valid input with en language",
			input: &DescriptionInput{
				Language:    "en",
				DetailLevel: "detailed",
			},
			wantErr: false,
		},
		{
			name: "unsupported language",
			input: &DescriptionInput{
				Language: "jp",
			},
			wantErr: true,
		},
		{
			name: "unsupported detail level",
			input: &DescriptionInput{
				DetailLevel: "ultra",
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

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain JSON",
			input:    `{"title": "test"}`,
			expected: `{"title": "test"}`,
		},
		{
			name: "JSON in markdown code block",
			input: "```json\n" +
				`{"title": "test"}` + "\n" +
				"```",
			expected: `{"title": "test"}`,
		},
		{
			name: "JSON with surrounding text",
			input: "Here is the description:\n" +
				`{"title": "test"}` + "\n" +
				"End of description.",
			expected: `{"title": "test"}`,
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
	chain := &DescriptionChain{
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
				"title": "Pod OOMKilled",
				"summary": "Pod terminated due to out of memory",
				"description": "The pod exceeded its memory limit and was killed by Kubernetes",
				"severity": "high",
				"affected_components": ["api-server"],
				"user_impact": "API requests failing",
				"business_impact": "Service degradation",
				"technical_details": {}
			}`,
			wantErr: false,
		},
		{
			name: "JSON in markdown",
			content: "```json\n" +
				`{"title": "test", "summary": "test", "description": "test", "severity": "low"}` +
				"\n```",
			wantErr: false,
		},
		{
			name: "missing title",
			content: `{
				"summary": "test",
				"description": "test"
			}`,
			wantErr: true,
		},
		{
			name: "missing description",
			content: `{
				"title": "test",
				"summary": "test"
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

			if output.Title == "" {
				t.Error("title is empty")
			}

			if output.Description == "" {
				t.Error("description is empty")
			}

			// 验证严重程度
			if !isValidSeverity(output.Severity) {
				t.Errorf("invalid severity: %s", output.Severity)
			}
		})
	}
}

func TestIsValidSeverity(t *testing.T) {
	tests := []struct {
		severity string
		expected bool
	}{
		{"critical", true},
		{"high", true},
		{"medium", true},
		{"low", true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			result := isValidSeverity(tt.severity)
			if result != tt.expected {
				t.Errorf("isValidSeverity(%q) = %v, want %v", tt.severity, result, tt.expected)
			}
		})
	}
}

func TestBuildPrompt_English(t *testing.T) {
	chain := &DescriptionChain{
		config: DefaultChainConfig(),
	}

	input := &DescriptionInput{
		FailureType:     "pod_failure",
		ResourceType:    "pod",
		ResourceName:    "test-pod",
		Namespace:       "default",
		ClusterID:       "cluster-1",
		Timestamp:       time.Now(),
		ErrorMessage:    "OOMKilled",
		Symptoms:        []string{"High memory usage", "Pod restart"},
		Language:        "en",
		DetailLevel:     "normal",
		IncludeTimeline: true,
	}

	prompt, err := chain.buildPrompt(input)
	if err != nil {
		t.Fatalf("buildPrompt() error: %v", err)
	}

	if prompt == "" {
		t.Fatal("buildPrompt() returned empty prompt")
	}

	// 验证包含关键英文内容
	requiredStrings := []string{
		"Kubernetes Failure Description Generation",
		"test-pod",
		"default",
		"OOMKilled",
		"High memory usage",
		"Language: English",
		"JSON format",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(prompt, required) {
			t.Errorf("Prompt missing required string: %s", required)
		}
	}

	t.Logf("English prompt length: %d characters", len(prompt))
}

func TestBuildPrompt_Chinese(t *testing.T) {
	chain := &DescriptionChain{
		config: DefaultChainConfig(),
	}

	input := &DescriptionInput{
		FailureType:     "pod_failure",
		ResourceType:    "pod",
		ResourceName:    "test-pod",
		Namespace:       "default",
		ClusterID:       "cluster-1",
		Timestamp:       time.Now(),
		ErrorMessage:    "内存不足",
		Symptoms:        []string{"内存使用过高", "Pod 重启"},
		Language:        "zh",
		DetailLevel:     "detailed",
		IncludeTimeline: false,
	}

	prompt, err := chain.buildPrompt(input)
	if err != nil {
		t.Fatalf("buildPrompt() error: %v", err)
	}

	if prompt == "" {
		t.Fatal("buildPrompt() returned empty prompt")
	}

	// 验证包含关键中文内容
	requiredStrings := []string{
		"Kubernetes 故障描述生成",
		"故障信息",
		"test-pod",
		"default",
		"内存不足",
		"语言: 中文",
		"JSON",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(prompt, required) {
			t.Errorf("Prompt missing required string: %s", required)
		}
	}

	t.Logf("Chinese prompt length: %d characters", len(prompt))
}

func TestBuildPromptWithRootCause(t *testing.T) {
	chain := &DescriptionChain{
		config: DefaultChainConfig(),
	}

	input := &DescriptionInput{
		FailureType:  "pod_failure",
		ResourceType: "pod",
		ResourceName: "test-pod",
		Namespace:    "default",
		ClusterID:    "cluster-1",
		Timestamp:    time.Now(),
		Language:     "en",
		DetailLevel:  "normal",
		RootCause: &RootCauseInfo{
			RootCause:  "Memory limit too low",
			Confidence: 0.95,
			Category:   "resource",
			Reasoning:  "Pod memory usage exceeded limit",
		},
	}

	prompt, err := chain.buildPrompt(input)
	if err != nil {
		t.Fatalf("buildPrompt() error: %v", err)
	}

	// 验证包含根因信息
	if !strings.Contains(prompt, "Root Cause Analysis") {
		t.Error("Prompt should contain Root Cause Analysis section")
	}

	if !strings.Contains(prompt, "Memory limit too low") {
		t.Error("Prompt should contain root cause")
	}

	if !strings.Contains(prompt, "0.95") {
		t.Error("Prompt should contain confidence")
	}
}

func TestGetSystemPrompt(t *testing.T) {
	chain := &DescriptionChain{
		config: DefaultChainConfig(),
	}

	tests := []struct {
		language string
		contains []string
	}{
		{
			language: "en",
			contains: []string{"Kubernetes", "failure descriptions", "JSON"},
		},
		{
			language: "zh",
			contains: []string{"Kubernetes", "故障描述", "JSON"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			prompt := chain.getSystemPrompt(tt.language)

			if prompt == "" {
				t.Error("getSystemPrompt() returned empty string")
			}

			for _, required := range tt.contains {
				if !strings.Contains(prompt, required) {
					t.Errorf("System prompt missing required string: %s", required)
				}
			}
		})
	}
}

func TestGetJSONSchema(t *testing.T) {
	chain := &DescriptionChain{
		config: DefaultChainConfig(),
	}

	tests := []struct {
		name            string
		language        string
		includeTimeline bool
		contains        []string
	}{
		{
			name:            "English with timeline",
			language:        "en",
			includeTimeline: true,
			contains:        []string{"title", "summary", "description", "timeline"},
		},
		{
			name:            "English without timeline",
			language:        "en",
			includeTimeline: false,
			contains:        []string{"title", "summary", "description"},
		},
		{
			name:            "Chinese with timeline",
			language:        "zh",
			includeTimeline: true,
			contains:        []string{"故障标题", "简短摘要", "详细描述", "timeline"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := chain.getJSONSchema(tt.language, tt.includeTimeline)

			if schema == "" {
				t.Error("getJSONSchema() returned empty string")
			}

			// 验证是否为有效 JSON
			var obj map[string]interface{}
			if err := json.Unmarshal([]byte(schema), &obj); err != nil {
				t.Errorf("Schema is not valid JSON: %v", err)
			}

			// 验证包含所需字段
			for _, required := range tt.contains {
				if !strings.Contains(schema, required) {
					t.Errorf("Schema missing required field: %s", required)
				}
			}

			// 验证 timeline 字段的存在性
			hasTimeline := strings.Contains(schema, "timeline")
			if tt.includeTimeline && !hasTimeline {
				t.Error("Schema should include timeline")
			}
			if !tt.includeTimeline && hasTimeline {
				t.Error("Schema should not include timeline")
			}
		})
	}
}

func TestNewDescriptionChain(t *testing.T) {
	tests := []struct {
		name    string
		proxy   interface{}
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
			proxy:   &struct{}{},
			config:  nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.proxy == nil {
				_, err := NewDescriptionChain(nil, tt.config)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewDescriptionChain() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
