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

// ServerOptions defines options for cluster service
// This implements the pkg/app.Options interface.
type ServerOptions struct {
	Server   *commonoptions.ServerOptions   `json:"server" mapstructure:"server"`
	Database *commonoptions.DatabaseOptions `json:"database" mapstructure:"database"`
	JWT      *commonoptions.JWTOptions      `json:"jwt" mapstructure:"jwt"`
	Logging  *commonoptions.LoggingOptions  `json:"logging" mapstructure:"logging"`
}

// NewServerOptions creates a new ServerOptions instance with default values.
func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		Server:   commonoptions.NewServerOptions(),
		Database: commonoptions.NewDatabaseOptions(),
		JWT:      commonoptions.NewJWTOptions(),
		Logging:  commonoptions.NewLoggingOptions(),
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

// InitLogger initializes the logger based on logging options
// This method is required by the Bootstrap pattern.
func (o *ServerOptions) InitLogger() (core.Logger, error) {
	return loggerutil.InitFromOptions(o.Logging)
}

// GetServiceName returns the service name.
func (o *ServerOptions) GetServiceName() string {
	return "Cluster"
}

// GetLogFields returns log fields for initialization logging.
func (o *ServerOptions) GetLogFields() []interface{} {
	return []interface{}{
		"http_port", o.Server.Port,
	}
}

// GetHealthPort returns the health check port
// 简化版本：直接返回固定端口，不使用HealthOptions.
func (o *ServerOptions) GetHealthPort() int {
	return o.Server.Port // Cluster 健康检查端口
}
