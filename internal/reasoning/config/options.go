package config

import (
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/spf13/pflag"
)

// Options defines options for reasoning service
type Options struct {
	Server      *commonoptions.ServerOptions      `json:"server" mapstructure:"server"`
	Logging     *commonoptions.LoggingOptions     `json:"logging" mapstructure:"logging"`
	Health      *commonoptions.HealthOptions      `json:"health" mapstructure:"health"`
	LLM         *commonoptions.LLMOptions         `json:"llm" mapstructure:"llm"`
	Memory      *commonoptions.MemoryOptions      `json:"memory" mapstructure:"memory"`
	Analysis    *commonoptions.AnalysisOptions    `json:"analysis" mapstructure:"analysis"`
	Prediction  *commonoptions.PredictionOptions  `json:"prediction" mapstructure:"prediction"`
	Learning    *commonoptions.LearningOptions    `json:"learning" mapstructure:"learning"`
	Performance *commonoptions.PerformanceOptions `json:"performance" mapstructure:"performance"`
}

// NewOptions creates a new Options instance with default values
func NewOptions() *Options {
	healthOpts := commonoptions.NewHealthOptions()
	healthOpts.Port = 8093 // Reasoning 健康检查端口

	return &Options{
		Server:      commonoptions.NewServerOptions(),
		Logging:     commonoptions.NewLoggingOptions(),
		Health:      healthOpts,
		LLM:         commonoptions.NewLLMOptions(),
		Memory:      commonoptions.NewMemoryOptions(),
		Analysis:    commonoptions.NewAnalysisOptions(),
		Prediction:  commonoptions.NewPredictionOptions(),
		Learning:    commonoptions.NewLearningOptions(),
		Performance: commonoptions.NewPerformanceOptions(),
	}
}

// Validate validates all the required options
func (o *Options) Validate() []error {
	var errs []error

	if err := o.Server.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Logging.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.LLM.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Memory.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Analysis.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Prediction.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Learning.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Performance.Validate(); err != nil {
		errs = append(errs, err)
	}

	return errs
}

// Complete fills in any fields not set that are required to have valid data
func (o *Options) Complete() error {
	// Complete all sub-options
	if err := o.Server.Complete(); err != nil {
		return err
	}

	if err := o.Logging.Complete(); err != nil {
		return err
	}

	if err := o.LLM.Complete(); err != nil {
		return err
	}

	if err := o.Memory.Complete(); err != nil {
		return err
	}

	if err := o.Analysis.Complete(); err != nil {
		return err
	}

	if err := o.Prediction.Complete(); err != nil {
		return err
	}

	if err := o.Learning.Complete(); err != nil {
		return err
	}

	if err := o.Performance.Complete(); err != nil {
		return err
	}

	// Apply LLM API key overrides from environment variables
	cfg := &Config{
		LLM: *o.LLM,
	}
	applyLLMEnvOverrides(cfg)
	o.LLM = &cfg.LLM

	return nil
}

// GetHealthPort 实现 commonapp.HealthPortProvider 接口
func (o *Options) GetHealthPort() int {
	if o.Health != nil {
		return o.Health.Port
	}
	return 8093 // 默认端口
}

// AddFlags adds flags to the specified FlagSet
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.Server.AddFlags(fs)
	o.Logging.AddFlags(fs)
	if o.Health != nil {
		o.Health.AddFlags(fs, "")
	}
	o.LLM.AddFlags(fs)
	o.Memory.AddFlags(fs)
	o.Analysis.AddFlags(fs)
	o.Prediction.AddFlags(fs)
	o.Learning.AddFlags(fs)
	o.Performance.AddFlags(fs)
}

// Config returns a Config struct based on Options
func (o *Options) Config() *Config {
	return &Config{
		Server:      *o.Server,
		LLM:         *o.LLM,
		Memory:      *o.Memory,
		Analysis:    *o.Analysis,
		Prediction:  *o.Prediction,
		Learning:    *o.Learning,
		Performance: *o.Performance,
		Logging:     *o.Logging,
	}
}
