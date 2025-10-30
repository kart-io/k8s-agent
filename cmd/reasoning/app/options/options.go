// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package options

import (
	commonlogger "github.com/kart-io/k8s-agent/common/logger"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	reasoningconfig "github.com/kart-io/k8s-agent/internal/reasoning/config"
	"github.com/kart-io/logger/core"
	"github.com/spf13/pflag"
)

// ServerOptions defines options for reasoning service
// This implements the pkg/app.Options interface
type ServerOptions struct {
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

// NewServerOptions creates a new ServerOptions instance with default values
func NewServerOptions() *ServerOptions {
	healthOpts := commonoptions.NewHealthOptions()
	healthOpts.Port = 8093 // Reasoning 健康检查端口

	return &ServerOptions{
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
func (o *ServerOptions) Validate() []error {
	var errs []error

	if err := o.Server.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Logging.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Health.Validate(); err != nil {
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
func (o *ServerOptions) Complete() error {
	// Complete all sub-options
	if err := o.Server.Complete(); err != nil {
		return err
	}

	if err := o.Logging.Complete(); err != nil {
		return err
	}

	if err := o.Health.Complete(); err != nil {
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

	// Note: LLM environment variable overrides are applied in internal/reasoning/config
	// during Config() conversion if needed

	return nil
}

// AddFlags adds flags to the flag set
// Note: --config/-c flag is automatically added by pkg/app framework
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
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

// InitLogger initializes the logger based on logging options
// This method is required by the Bootstrap pattern
func (o *ServerOptions) InitLogger() (core.Logger, error) {
	return commonlogger.InitFromOptions(o.Logging)
}

// GetHealthPort returns the health check port
// This method is required by the Bootstrap pattern
func (o *ServerOptions) GetHealthPort() int {
	if o.Health != nil {
		return o.Health.Port
	}
	return 8093 // 默认端口
}

// Config converts ServerOptions to internal reasoning config
// This method is required by the Bootstrap pattern
func (o *ServerOptions) Config() *reasoningconfig.Config {
	return &reasoningconfig.Config{
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
