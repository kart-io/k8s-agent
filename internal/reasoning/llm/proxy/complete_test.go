package proxy

import (
	"context"
	"testing"
	"time"

	"github.com/kart-io/k8s-agent/internal/reasoning/config"
)

func TestCompleteWithMockGollm(t *testing.T) {
	// 注意: 这个测试需要真实的 API key 才能运行
	// 在 CI/CD 环境中应该跳过或使用 mock
	t.Skip("Skipping test that requires real API keys")

	cfg := &config.LLMConfig{
		Enabled: true,
		Providers: []config.LLMProviderConfig{
			{
				Name:        "openai",
				APIKey:      "test-key",
				Model:       "gpt-4",
				MaxTokens:   100,
				Temperature: 0.7,
				Timeout:     30,
				Priority:    1,
			},
		},
	}

	adapter, err := NewProxyAdapter(cfg)
	if err != nil {
		t.Fatalf("NewProxyAdapter() failed: %v", err)
	}

	req := &CompletionRequest{
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Say hello!"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := adapter.Complete(ctx, req)
	if err != nil {
		t.Fatalf("Complete() failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Complete() returned nil response")
	}

	if resp.Content == "" {
		t.Error("Response content is empty")
	}

	if resp.Provider == "" {
		t.Error("Response provider is empty")
	}

	t.Logf("Response: %+v", resp)
}

func TestBuildGollmPrompt(t *testing.T) {
	tests := []struct {
		name    string
		req     *CompletionRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &CompletionRequest{
				Messages: []Message{
					{Role: "system", Content: "You are helpful."},
					{Role: "user", Content: "Hello!"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty messages",
			req: &CompletionRequest{
				Messages: []Message{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, err := buildGollmPrompt(tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("buildGollmPrompt() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("buildGollmPrompt() unexpected error: %v", err)
				return
			}

			if prompt == nil {
				t.Error("buildGollmPrompt() returned nil prompt")
			}
		})
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		text     string
		expected int
	}{
		{"Hello", 1},          // 5 / 4 = 1
		{"Hello World", 2},    // 11 / 4 = 2
		{"This is a test", 3}, // 14 / 4 = 3
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			tokens := estimateTokens(tt.text)
			if tokens != tt.expected {
				t.Errorf("estimateTokens(%q) = %d, want %d", tt.text, tokens, tt.expected)
			}
		})
	}
}

func TestCalculateCost(t *testing.T) {
	tests := []struct {
		provider   string
		tokensUsed int
		minCost    float64 // 最小预期成本
		maxCost    float64 // 最大预期成本
	}{
		{"openai", 1000, 0.01, 0.03},
		{"gemini", 1000, 0.005, 0.015},
		{"deepseek", 1000, 0.0001, 0.002},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			cost := calculateCost(tt.provider, tt.tokensUsed)
			if cost < tt.minCost || cost > tt.maxCost {
				t.Errorf("calculateCost(%s, %d) = %f, want between %f and %f",
					tt.provider, tt.tokensUsed, cost, tt.minCost, tt.maxCost)
			}
		})
	}
}
