package llm

import "fmt"

// NewClient creates a new LLM client based on the config
func NewClient(config *Config) (Client, error) {
	switch config.Provider {
	case ProviderOpenAI:
		return NewOpenAIClient(config)
	case ProviderGemini:
		return NewGeminiClient(config)
	case ProviderDeepSeek:
		return NewDeepSeekClient(config)
	case ProviderOllama:
		return NewOllamaClient(config)
	case ProviderSiliconFlow:
		return NewSiliconFlowClient(config)
	case ProviderKimi:
		return NewKimiClient(config)
	case ProviderCustom:
		return NewCustomClient(config)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", config.Provider)
	}
}

// NewMultiClient creates multiple LLM clients for fallback
func NewMultiClient(configs []*Config) ([]Client, error) {
	var clients []Client
	for _, config := range configs {
		client, err := NewClient(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create client for %s: %w", config.Provider, err)
		}
		clients = append(clients, client)
	}
	return clients, nil
}
