package config

import (
	"github.com/kart-io/k8s-agent/cmd/reasoning/app/options"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
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
