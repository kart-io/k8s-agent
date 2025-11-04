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

// OpenAIClient implements the Client interface for OpenAI
type OpenAIClient struct {
	config     *Config
	httpClient *http.Client
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(config *Config) (*OpenAIClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}

	if config.Model == "" {
		config.Model = "gpt-4o-mini" // Default to GPT-4o-mini for cost efficiency
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

	return &OpenAIClient{
		config: config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

type openAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type openAIResponse struct {
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
func (c *OpenAIClient) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
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

	apiReq := openAIRequest{
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
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from OpenAI")
	}

	return &CompletionResponse{
		Content:      apiResp.Choices[0].Message.Content,
		Model:        apiResp.Model,
		TokensUsed:   apiResp.Usage.TotalTokens,
		FinishReason: apiResp.Choices[0].FinishReason,
	}, nil
}

// AnalyzeRootCause uses OpenAI to analyze root cause
func (c *OpenAIClient) AnalyzeRootCause(ctx context.Context, event map[string]interface{}, logs string, metrics string) (string, error) {
	eventJSON, _ := json.MarshalIndent(event, "", "  ")

	systemPrompt := RootCauseAnalysisSystemPrompt

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

// GenerateRecommendations uses OpenAI to generate recommendations
func (c *OpenAIClient) GenerateRecommendations(ctx context.Context, rootCause string, contextInfo string) (string, error) {
	systemPrompt := RecommendationsSystemPrompt

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
func (c *OpenAIClient) Provider() Provider {
	return ProviderOpenAI
}

// IsAvailable checks if OpenAI is available
func (c *OpenAIClient) IsAvailable() bool {
	return c.config.APIKey != ""
}
