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

// KimiClient implements the Client interface for Kimi (Moonshot AI)
// Kimi provides OpenAI-compatible API with excellent long-text capability
type KimiClient struct {
	config     *Config
	httpClient *http.Client
}

// NewKimiClient creates a new Kimi client
func NewKimiClient(config *Config) (*KimiClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Kimi API key is required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.moonshot.cn/v1"
	}

	if config.Model == "" {
		config.Model = "moonshot-v1-8k" // Default to 8k context version
	}

	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}

	if config.Temperature == 0 {
		config.Temperature = 0.7
	}

	timeout := 30 * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	return &KimiClient{
		config: config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

type kimiRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type kimiResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Complete implements the Client interface
func (c *KimiClient) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	model := c.config.Model
	if req.Model != "" {
		model = req.Model
	}

	temperature := c.config.Temperature
	if req.Temperature > 0 {
		temperature = req.Temperature
	}

	maxTokens := c.config.MaxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}

	apiReq := kimiRequest{
		Model:       model,
		Messages:    req.Messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Kimi API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResp kimiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from Kimi")
	}

	return &CompletionResponse{
		Content:      apiResp.Choices[0].Message.Content,
		Model:        apiResp.Model,
		TokensUsed:   apiResp.Usage.TotalTokens,
		FinishReason: apiResp.Choices[0].FinishReason,
	}, nil
}

// AnalyzeRootCause uses Kimi to analyze root cause
func (c *KimiClient) AnalyzeRootCause(ctx context.Context, event map[string]interface{}, logs string, metrics string) (string, error) {
	eventJSON, _ := json.MarshalIndent(event, "", "  ")

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

// GenerateRecommendations uses Kimi to generate recommendations
func (c *KimiClient) GenerateRecommendations(ctx context.Context, rootCause string, contextInfo string) (string, error) {
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
func (c *KimiClient) Provider() Provider {
	return ProviderKimi
}

// IsAvailable checks if Kimi is available
func (c *KimiClient) IsAvailable() bool {
	return c.config.APIKey != ""
}
