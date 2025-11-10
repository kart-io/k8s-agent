package config

import (
	"time"

	commonoptions "github.com/kart-io/k8s-agent/common/options"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	pkgoptions "github.com/kart-io/k8s-agent/pkg/options"
)

// 类型别名，用于向后兼容
type (
	LLMConfig         = pkgoptions.LLMOptions
	LLMProviderConfig = pkgoptions.LLMProviderConfig
)

// ReasoningOptions is a minimal wrapper around ServerOptions for backward compatibility
// This allows the business logic to continue working without major refactoring
type ReasoningOptions struct {
	Server      commonoptions.ServerOptions
	LLM         pkgoptions.LLMOptions
	Memory      commonoptions.MemoryOptions
	Analysis    pkgoptions.AnalysisOptions
	Prediction  pkgoptions.PredictionOptions
	Learning    pkgoptions.LearningOptions
	Performance commonoptions.PerformanceOptions
	Logging     commonoptions.LoggingOptions
}

// Backward compatibility alias
type Config = ReasoningOptions

// NewReasoningOptionsFromStandardOptions creates ReasoningOptions from StandardOptions
func NewReasoningOptionsFromStandardOptions(opts *commonapp.StandardOptions) *ReasoningOptions {
	// Get LLM options, use defaults if not set
	llm := pkgoptions.LLMOptions{}
	if opts.LLM != nil {
		llm = *opts.LLM
	}

	return &ReasoningOptions{
		Server:      *opts.Server,
		LLM:         llm,
		Memory:      commonoptions.MemoryOptions{}, // Use defaults
		Analysis:    pkgoptions.AnalysisOptions{},
		Prediction:  pkgoptions.PredictionOptions{},
		Learning:    pkgoptions.LearningOptions{},
		Performance: commonoptions.PerformanceOptions{},
		Logging:     *opts.Logging,
	}
}

// NewConfigFromOptions creates a Config from StandardOptions (backward compatibility)
func NewConfigFromOptions(opts *commonapp.StandardOptions) *Config {
	return NewReasoningOptionsFromStandardOptions(opts)
}

// GetRequestTimeout 返回请求超时时间，用于向后兼容
func (o *ReasoningOptions) GetRequestTimeout() time.Duration {
	if o.Server.ReadTimeout > 0 {
		return o.Server.ReadTimeout
	}
	return 30 * time.Second // 默认 30 秒
}
