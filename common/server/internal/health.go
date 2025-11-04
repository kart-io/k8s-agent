// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package internal provides shared internal utilities for server implementations.
package internal

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/health"
	"github.com/kart-io/k8s-agent/common/options"
)

// RegisterHealthEndpoints 注册健康检查端点到 Gin Engine
// 这是 common/health 到 Gin 的适配器
// 使用 HealthOptions 配置端点路径
func RegisterHealthEndpoints(engine *gin.Engine, manager *health.Manager, opts *options.HealthOptions) {
	if opts == nil {
		opts = options.NewHealthOptions()
	}

	// 完成配置以确保路径正确
	opts.Complete()

	// 使用统一的 health manager 和可配置的路径
	engine.GET(opts.Path, func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	engine.GET(opts.ReadinessPath, func(c *gin.Context) {
		results, allHealthy := manager.CheckAll(c.Request.Context())

		if !allHealthy {
			c.JSON(503, gin.H{"status": "not ready", "checks": results})
			return
		}

		c.JSON(200, gin.H{"status": "ready", "checks": results})
	})

	engine.GET(opts.LivenessPath, func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "alive"})
	})
}
