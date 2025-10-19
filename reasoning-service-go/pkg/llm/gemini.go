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

// GeminiClient implements the Client interface for Google Gemini
type GeminiClient struct {
	config     *Config
	httpClient *http.Client
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient(config *Config) (*GeminiClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://generativelanguage.googleapis.com/v1beta"
	}

	if config.Model == "" {
		config.Model = "gemini-1.5-flash" // Default to Gemini 1.5 Flash
	}

	if config.MaxTokens == 0 {
		config.MaxTokens = 8192
	}

	if config.Temperature == 0 {
		config.Temperature = 0.7
	}

	timeout := 30 * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	return &GeminiClient{
		config: config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

type geminiRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
		Role string `json:"role,omitempty"`
	} `json:"contents"`
	GenerationConfig struct {
		Temperature   float64 `json:"temperature,omitempty"`
		MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	} `json:"generationConfig,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

// Complete implements the Client interface
func (c *GeminiClient) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
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

	// Convert messages to Gemini format
	var apiReq geminiRequest
	for _, msg := range req.Messages {
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}
		// Gemini doesn't have separate system role, merge into user message
		apiReq.Contents = append(apiReq.Contents, struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role,omitempty"`
		}{
			Parts: []struct {
				Text string `json:"text"`
			}{{Text: msg.Content}},
			Role: role,
		})
	}

	apiReq.GenerationConfig.Temperature = temperature
	apiReq.GenerationConfig.MaxOutputTokens = maxTokens

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.config.BaseURL, model, c.config.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Gemini API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(apiResp.Candidates) == 0 || len(apiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content returned from Gemini")
	}

	return &CompletionResponse{
		Content:      apiResp.Candidates[0].Content.Parts[0].Text,
		Model:        model,
		TokensUsed:   apiResp.UsageMetadata.TotalTokenCount,
		FinishReason: apiResp.Candidates[0].FinishReason,
	}, nil
}

// AnalyzeRootCause uses Gemini to analyze root cause
func (c *GeminiClient) AnalyzeRootCause(ctx context.Context, event map[string]interface{}, logs string, metrics string) (string, error) {
	eventJSON, _ := json.MarshalIndent(event, "", "  ")

	userPrompt := fmt.Sprintf(`%s

Analyze the following Kubernetes issue:

Event:
%s

Logs:
%s

Metrics:
%s

Provide your root cause analysis.`, RootCauseAnalysisSystemPrompt, string(eventJSON), logs, metrics)

	resp, err := c.Complete(ctx, &CompletionRequest{
		Messages: []Message{
			{Role: "user", Content: userPrompt},
		},
	})

	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// GenerateRecommendations uses Gemini to generate recommendations
func (c *GeminiClient) GenerateRecommendations(ctx context.Context, rootCause string, contextInfo string) (string, error) {
	userPrompt := fmt.Sprintf(`%s

Root Cause: %s

Context:
%s

Provide recommended actions to fix this issue.`, RecommendationsSystemPrompt, rootCause, contextInfo)

	resp, err := c.Complete(ctx, &CompletionRequest{
		Messages: []Message{
			{Role: "user", Content: userPrompt},
		},
	})

	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// Provider returns the provider type
func (c *GeminiClient) Provider() Provider {
	return ProviderGemini
}

// IsAvailable checks if Gemini is available
func (c *GeminiClient) IsAvailable() bool {
	return c.config.APIKey != ""
}
