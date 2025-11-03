// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"

	"github.com/kart-io/k8s-agent/common/health"
	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

// NewHealthCheckServer 创建健康检查服务器
// 这是 common/health 的便捷包装
func NewHealthCheckServer(opts *options.HealthOptions, logger core.Logger) health.Server {
	return health.NewHTTPServer(opts, logger)
}

// GetHealthOptions 从 Options 中获取健康检查配置
func GetHealthOptions(opts Options) *options.HealthOptions {
	// 1. HealthOptionsProvider
	if provider, ok := opts.(HealthOptionsProvider); ok {
		if healthOpts := provider.GetHealthOptions(); healthOpts != nil {
			return healthOpts
		}
	}

	// 2. HealthPortProvider (向后兼容)
	if provider, ok := opts.(HealthPortProvider); ok {
		port := provider.GetHealthPort()
		if port > 0 {
			healthOpts := options.NewHealthOptions()
			healthOpts.Port = port
			return healthOpts
		}
	}

	// 3. 默认配置
	return options.NewHealthOptions()
}

// DefaultHealthCheckFuncFromOptions 从 Options 创建健康检查函数
func DefaultHealthCheckFuncFromOptions(opts Options) HealthCheckFunc {
	healthOpts := GetHealthOptions(opts)
	return func() error {
		server := NewHealthCheckServer(healthOpts, nil)
		return server.Start(context.Background())
	}
}
