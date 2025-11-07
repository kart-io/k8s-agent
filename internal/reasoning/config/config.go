package config

import (
	"time"

	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
)

// 类型别名，用于向后兼容
type (
	LLMConfig         = commonoptions.LLMOptions
	LLMProviderConfig = commonoptions.LLMProviderConfig
)

// Config is a minimal wrapper around ServerOptions for backward compatibility
// This allows the business logic to continue working without major refactoring
type Config struct {
	Server      commonoptions.ServerOptions
	LLM         commonoptions.LLMOptions
	Memory      commonoptions.MemoryOptions
	Analysis    commonoptions.AnalysisOptions
	Prediction  commonoptions.PredictionOptions
	Learning    commonoptions.LearningOptions
	Performance commonoptions.PerformanceOptions
	Logging     commonoptions.LoggingOptions
}

// NewConfigFromOptions creates a Config from StandardOptions
func NewConfigFromOptions(opts *commonapp.StandardOptions) *Config {
	// Get LLM options, use defaults if not set
	llm := commonoptions.LLMOptions{}
	if opts.LLM != nil {
		llm = *opts.LLM
	}

	return &Config{
		Server:      *opts.Server,
		LLM:         llm,
		Memory:      commonoptions.MemoryOptions{}, // Use defaults
		Analysis:    commonoptions.AnalysisOptions{},
		Prediction:  commonoptions.PredictionOptions{},
		Learning:    commonoptions.LearningOptions{},
		Performance: commonoptions.PerformanceOptions{},
		Logging:     *opts.Logging,
	}
}

// GetRequestTimeout 返回请求超时时间，用于向后兼容
func (c *Config) GetRequestTimeout() time.Duration {
	if c.Server.ReadTimeout > 0 {
		return c.Server.ReadTimeout
	}
	return 30 * time.Second // 默认 30 秒
}
