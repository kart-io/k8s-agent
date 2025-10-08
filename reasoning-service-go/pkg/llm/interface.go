package llm

import (
	"context"
)

// Provider represents an LLM provider type
type Provider string

const (
	ProviderOpenAI      Provider = "openai"
	ProviderGemini      Provider = "gemini"
	ProviderDeepSeek    Provider = "deepseek"
	ProviderOllama      Provider = "ollama"
	ProviderSiliconFlow Provider = "siliconflow"
	ProviderKimi        Provider = "kimi"
	ProviderCustom      Provider = "custom"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"`
}

// CompletionRequest represents a request for LLM completion
type CompletionRequest struct {
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Model       string    `json:"model,omitempty"`
}

// CompletionResponse represents the LLM completion response
type CompletionResponse struct {
	Content      string `json:"content"`
	Model        string `json:"model"`
	TokensUsed   int    `json:"tokens_used,omitempty"`
	FinishReason string `json:"finish_reason,omitempty"`
}

// Client interface for LLM providers
type Client interface {
	// Complete generates a completion for the given request
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// AnalyzeRootCause uses LLM to analyze root cause from context
	AnalyzeRootCause(ctx context.Context, event map[string]interface{}, logs string, metrics string) (string, error)

	// GenerateRecommendations uses LLM to generate recommendations
	GenerateRecommendations(ctx context.Context, rootCause string, context string) (string, error)

	// Provider returns the provider type
	Provider() Provider

	// IsAvailable checks if the LLM provider is available
	IsAvailable() bool
}

// Config represents the configuration for LLM client
type Config struct {
	Provider   Provider
	APIKey     string
	BaseURL    string // Optional custom base URL
	Model      string // Default model to use
	MaxTokens  int
	Temperature float64
	Timeout    int // Request timeout in seconds
}
