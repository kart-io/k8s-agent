package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CustomClient implements a generic OpenAI-compatible LLM client
// This allows connecting to any custom LLM service that follows OpenAI's API format
type CustomClient struct {
	config     *Config
	httpClient *http.Client
}

// NewCustomClient creates a new custom LLM client
func NewCustomClient(config *Config) (*CustomClient, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base_url is required for custom LLM provider")
	}
	if config.Model == "" {
		return nil, fmt.Errorf("model is required for custom LLM provider")
	}

	timeout := 30 * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	return &CustomClient{
		config: config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

type customRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

type customResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Complete sends a completion request to the custom LLM service
func (c *CustomClient) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	model := c.config.Model
	if req.Model != "" {
		model = req.Model
	}

	temperature := 0.7
	if c.config.Temperature > 0 {
		temperature = c.config.Temperature
	}
	if req.Temperature > 0 {
		temperature = req.Temperature
	}

	maxTokens := 4096
	if c.config.MaxTokens > 0 {
		maxTokens = c.config.MaxTokens
	}
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}

	apiReq := customRequest{
		Model:       model,
		Messages:    req.Messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      false,
	}

	jsonData, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("custom LLM API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp customResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &CompletionResponse{
		Content:      apiResp.Choices[0].Message.Content,
		Model:        apiResp.Model,
		TokensUsed:   apiResp.Usage.TotalTokens,
		FinishReason: apiResp.Choices[0].FinishReason,
	}, nil
}

// AnalyzeRootCause analyzes root cause using the custom LLM
func (c *CustomClient) AnalyzeRootCause(ctx context.Context, event map[string]interface{}, logs string, metrics string) (string, error) {
	eventJSON, _ := json.Marshal(event)

	systemPrompt := `You are an expert Kubernetes troubleshooting assistant. Analyze the provided event, logs, and metrics to identify the root cause of the issue. Provide a clear, concise analysis with:
1. Root cause type (e.g., OOMKiller, CPUThrottling, NetworkError, ConfigError, etc.)
2. Confidence level (0.0-1.0)
3. Key evidence supporting your analysis
4. Brief explanation

Format your response as JSON with fields: root_cause_type, confidence, evidence (array), explanation.`

	userPrompt := fmt.Sprintf(`Analyze the following Kubernetes issue:

Event:
%s

Logs:
%s

Metrics:
%s

Provide your root cause analysis.`, string(eventJSON), logs, metrics)

	resp, err := c.Complete(ctx, &CompletionRequest{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	})

	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// GenerateRecommendations generates recommendations using the custom LLM
func (c *CustomClient) GenerateRecommendations(ctx context.Context, rootCause string, contextInfo string) (string, error) {
	systemPrompt := `You are an expert Kubernetes operations engineer. Based on the identified root cause, provide actionable recommendations to fix the issue. For each recommendation, include:
1. Action name
2. Description
3. Risk level (low/medium/high)
4. Impact description
5. Step-by-step instructions
6. Rollback steps
7. Estimated duration

Format your response as JSON array of recommendations.`

	userPrompt := fmt.Sprintf(`Root Cause: %s

Context:
%s

Provide recommended actions to fix this issue.`, rootCause, contextInfo)

	resp, err := c.Complete(ctx, &CompletionRequest{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	})

	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// Provider returns the provider type
func (c *CustomClient) Provider() Provider {
	return ProviderCustom
}

// IsAvailable checks if the custom LLM service is available
func (c *CustomClient) IsAvailable() bool {
	// Try to ping the service with a simple request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if base URL is accessible
	testReq := customRequest{
		Model: c.config.Model,
		Messages: []Message{
			{Role: "user", Content: "ping"},
		},
		MaxTokens: 10,
	}

	jsonData, err := json.Marshal(testReq)
	if err != nil {
		return false
	}

	url := fmt.Sprintf("%s/chat/completions", c.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return false
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Accept any 2xx or 4xx status (4xx means service is up but our request might be invalid)
	return resp.StatusCode >= 200 && resp.StatusCode < 500
}
