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

// SiliconFlowClient implements the Client interface for SiliconFlow
// SiliconFlow provides OpenAI-compatible API for various models
type SiliconFlowClient struct {
	config     *Config
	httpClient *http.Client
}

// NewSiliconFlowClient creates a new SiliconFlow client
func NewSiliconFlowClient(config *Config) (*SiliconFlowClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("SiliconFlow API key is required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.siliconflow.cn/v1"
	}

	if config.Model == "" {
		config.Model = "Qwen/Qwen2.5-7B-Instruct" // Default to Qwen 2.5
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

	return &SiliconFlowClient{
		config: config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

type siliconFlowRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type siliconFlowResponse struct {
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
func (c *SiliconFlowClient) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
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

	apiReq := siliconFlowRequest{
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
		return nil, fmt.Errorf("SiliconFlow API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResp siliconFlowResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from SiliconFlow")
	}

	return &CompletionResponse{
		Content:      apiResp.Choices[0].Message.Content,
		Model:        apiResp.Model,
		TokensUsed:   apiResp.Usage.TotalTokens,
		FinishReason: apiResp.Choices[0].FinishReason,
	}, nil
}

// AnalyzeRootCause uses SiliconFlow to analyze root cause
func (c *SiliconFlowClient) AnalyzeRootCause(ctx context.Context, event map[string]interface{}, logs string, metrics string) (string, error) {
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

// GenerateRecommendations uses SiliconFlow to generate recommendations
func (c *SiliconFlowClient) GenerateRecommendations(ctx context.Context, rootCause string, contextInfo string) (string, error) {
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
func (c *SiliconFlowClient) Provider() Provider {
	return ProviderSiliconFlow
}

// IsAvailable checks if SiliconFlow is available
func (c *SiliconFlowClient) IsAvailable() bool {
	return c.config.APIKey != ""
}
