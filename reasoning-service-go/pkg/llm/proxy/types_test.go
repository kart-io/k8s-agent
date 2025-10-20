package proxy

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCompletionRequestSerialization(t *testing.T) {
	req := CompletionRequest{
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello!"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
		StopWords:   []string{"\n", "END"},
	}

	// 测试 JSON 序列化
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal CompletionRequest: %v", err)
	}

	// 测试 JSON 反序列化
	var decoded CompletionRequest
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal CompletionRequest: %v", err)
	}

	// 验证数据
	if decoded.Temperature != req.Temperature {
		t.Errorf("Temperature mismatch: got %f, want %f", decoded.Temperature, req.Temperature)
	}
	if decoded.MaxTokens != req.MaxTokens {
		t.Errorf("MaxTokens mismatch: got %d, want %d", decoded.MaxTokens, req.MaxTokens)
	}
	if len(decoded.Messages) != len(req.Messages) {
		t.Errorf("Messages length mismatch: got %d, want %d", len(decoded.Messages), len(req.Messages))
	}
}

func TestCompletionResponseSerialization(t *testing.T) {
	resp := CompletionResponse{
		Content:      "Hello! How can I help you?",
		Provider:     "openai",
		Model:        "gpt-4",
		TokensUsed:   50,
		Cost:         0.0025,
		Latency:      500 * time.Millisecond,
		FinishReason: "stop",
	}

	// 测试 JSON 序列化
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal CompletionResponse: %v", err)
	}

	// 测试 JSON 反序列化
	var decoded CompletionResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal CompletionResponse: %v", err)
	}

	// 验证数据
	if decoded.Content != resp.Content {
		t.Errorf("Content mismatch: got %s, want %s", decoded.Content, resp.Content)
	}
	if decoded.Provider != resp.Provider {
		t.Errorf("Provider mismatch: got %s, want %s", decoded.Provider, resp.Provider)
	}
	if decoded.TokensUsed != resp.TokensUsed {
		t.Errorf("TokensUsed mismatch: got %d, want %d", decoded.TokensUsed, resp.TokensUsed)
	}
}

func TestUsageMetricsSerialization(t *testing.T) {
	metrics := UsageMetrics{
		TotalRequests:   100,
		SuccessfulCalls: 95,
		FailedCalls:     5,
		TotalCost:       2.5,
		ProviderStats: map[string]ProviderMetrics{
			"openai": {
				Calls:      50,
				Successes:  48,
				Failures:   2,
				TotalCost:  1.5,
				AvgLatency: 300 * time.Millisecond,
			},
			"gemini": {
				Calls:      50,
				Successes:  47,
				Failures:   3,
				TotalCost:  1.0,
				AvgLatency: 400 * time.Millisecond,
			},
		},
	}

	// 测试 JSON 序列化
	data, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("Failed to marshal UsageMetrics: %v", err)
	}

	// 测试 JSON 反序列化
	var decoded UsageMetrics
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal UsageMetrics: %v", err)
	}

	// 验证数据
	if decoded.TotalRequests != metrics.TotalRequests {
		t.Errorf("TotalRequests mismatch: got %d, want %d", decoded.TotalRequests, metrics.TotalRequests)
	}
	if len(decoded.ProviderStats) != len(metrics.ProviderStats) {
		t.Errorf("ProviderStats length mismatch: got %d, want %d", len(decoded.ProviderStats), len(metrics.ProviderStats))
	}
}

func TestProviderStatusSerialization(t *testing.T) {
	now := time.Now()
	status := ProviderStatus{
		Name:      "openai",
		Healthy:   true,
		LastError: "",
		LastCheck: now,
	}

	// 测试 JSON 序列化
	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("Failed to marshal ProviderStatus: %v", err)
	}

	// 测试 JSON 反序列化
	var decoded ProviderStatus
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal ProviderStatus: %v", err)
	}

	// 验证数据
	if decoded.Name != status.Name {
		t.Errorf("Name mismatch: got %s, want %s", decoded.Name, status.Name)
	}
	if decoded.Healthy != status.Healthy {
		t.Errorf("Healthy mismatch: got %v, want %v", decoded.Healthy, status.Healthy)
	}
}
