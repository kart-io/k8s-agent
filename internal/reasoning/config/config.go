package config

import (
	"time"

	"github.com/kart-io/k8s-agent/cmd/reasoning/app/options"
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

// NewConfigFromOptions creates a Config from ServerOptions
func NewConfigFromOptions(opts *options.ServerOptions) *Config {
	return &Config{
		Server:      *opts.Server,
		LLM:         *opts.LLM,
		Memory:      *opts.Memory,
		Analysis:    *opts.Analysis,
		Prediction:  *opts.Prediction,
		Learning:    *opts.Learning,
		Performance: *opts.Performance,
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
