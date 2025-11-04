// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package options

import (
	"github.com/spf13/pflag"

	"github.com/kart-io/k8s-agent/common/loggerutil"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

// ServerOptions defines options for reasoning service
// This implements the pkg/app.Options interface.
type ServerOptions struct {
	Server      *commonoptions.ServerOptions      `json:"server" mapstructure:"server"`
	GRPC        *commonoptions.GRPCOptions        `json:"grpc" mapstructure:"grpc"`
	Logging     *commonoptions.LoggingOptions     `json:"logging" mapstructure:"logging"`
	LLM         *commonoptions.LLMOptions         `json:"llm" mapstructure:"llm"`
	Memory      *commonoptions.MemoryOptions      `json:"memory" mapstructure:"memory"`
	Analysis    *commonoptions.AnalysisOptions    `json:"analysis" mapstructure:"analysis"`
	Prediction  *commonoptions.PredictionOptions  `json:"prediction" mapstructure:"prediction"`
	Learning    *commonoptions.LearningOptions    `json:"learning" mapstructure:"learning"`
	Performance *commonoptions.PerformanceOptions `json:"performance" mapstructure:"performance"`
}

// NewServerOptions creates a new ServerOptions instance with default values.
func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		Server:      commonoptions.NewServerOptions(),
		GRPC:        commonoptions.NewGRPCOptions(),
		Logging:     commonoptions.NewLoggingOptions(),
		LLM:         commonoptions.NewLLMOptions(),
		Memory:      commonoptions.NewMemoryOptions(),
		Analysis:    commonoptions.NewAnalysisOptions(),
		Prediction:  commonoptions.NewPredictionOptions(),
		Learning:    commonoptions.NewLearningOptions(),
		Performance: commonoptions.NewPerformanceOptions(),
	}
}

// Validate validates all the required options.
func (o *ServerOptions) Validate() []error {
	// 使用通用工具函数统一验证所有子选项
	return commonoptions.ValidateAll(o)
}

// Complete fills in any fields not set that are required to have valid data.
func (o *ServerOptions) Complete() error {
	// 使用通用工具函数统一完成所有子选项
	return commonoptions.CompleteAll(o)
}

// AddFlags adds flags to the flag set
// Note: --config/-c flag is automatically added by pkg/app framework.
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	// 使用通用工具函数统一添加所有子选项的 flags
	commonoptions.AddFlagsAll(o, fs)
}

// GetServiceName returns the service name.
func (o *ServerOptions) GetServiceName() string {
	return "Reasoning"
}

// GetLogFields returns log fields for initialization logging.
func (o *ServerOptions) GetLogFields() []interface{} {
	return []interface{}{
		"http_port", o.Server.Port,
		"health_port", o.GetHealthPort(),
		"llm_enabled", o.LLM.Enabled,
		"memory_enabled", o.Memory.EnableVectorStore,
	}
}

// InitLogger initializes the logger based on logging options
// This method is required by the Bootstrap pattern.
func (o *ServerOptions) InitLogger() (core.Logger, error) {
	return loggerutil.InitFromOptions(o.Logging)
}

// GetHealthPort returns the health check port
// This method is required by the Bootstrap pattern
// 简化版本：直接返回固定端口，不使用HealthOptions.
func (o *ServerOptions) GetHealthPort() int {
	return o.Server.Port // Reasoning 健康检查端口
}
