// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
)

// ================= 预定义的常用中间件 =================
// 注意：这些中间件是示例，实际使用时需要在服务层实现具体的初始化器

// MetricsMiddleware 添加 Prometheus 指标收集中间件
// 优先级: 100 (非常早，在其他组件之前)
// 注意：需要在服务层提供 NewMetricsInitializer 实现
func MetricsMiddleware(port int, newMetricsInit func(int, core.Logger) bootstrap.Initializer) MiddlewareConfig {
	return Middleware("Metrics", 100, func(bs *bootstrap.Bootstrap, logger core.Logger, opts Options) error {
		logger.Infow("Registering Metrics middleware", "port", port)
		metricsInit := newMetricsInit(port, logger)
		bs.Register(metricsInit)
		return nil
	})
}

// TracingMiddleware 添加分布式追踪中间件
// 优先级: 150 (早期，在业务组件之前)
func TracingMiddleware(serviceName string, endpoint string) MiddlewareConfig {
	return Middleware("Tracing", 150, func(bs *bootstrap.Bootstrap, logger core.Logger, opts Options) error {
		logger.Infow("Registering Tracing middleware",
			"service", serviceName,
			"endpoint", endpoint,
		)
		// 实际实现由服务层提供
		return nil
	})
}

// ProfilingMiddleware 添加性能分析中间件 (pprof)
// 优先级: 200 (早期)
func ProfilingMiddleware(port int) MiddlewareConfig {
	return Middleware("Profiling", 200, func(bs *bootstrap.Bootstrap, logger core.Logger, opts Options) error {
		logger.Infow("Registering Profiling middleware", "port", port)
		// 实际实现由服务层提供
		return nil
	})
}

// RateLimitMiddleware 添加限流中间件
// 优先级: 250 (中等优先级，在业务组件之前)
func RateLimitMiddleware(requestsPerSecond int) MiddlewareConfig {
	return Middleware("RateLimit", 250, func(bs *bootstrap.Bootstrap, logger core.Logger, opts Options) error {
		logger.Infow("Registering RateLimit middleware", "rps", requestsPerSecond)
		// 实际实现由服务层提供
		return nil
	})
}

// ================= 自定义中间件辅助函数 =================

// CustomMiddleware 创建自定义中间件的便捷函数
// 示例：
//
//	app.WithMiddleware(CustomMiddleware("MyMiddleware", 500, func(bs, logger, opts) {
//	    // 你的中间件逻辑
//	    init := initializers.NewXXXInitializer(...)
//	    bs.Register(init)
//	}))
func CustomMiddleware(name string, priority int, fn func(bs *bootstrap.Bootstrap, logger core.Logger, opts Options) error) MiddlewareConfig {
	return Middleware(name, priority, fn)
}

// ConditionalMiddleware 创建条件中间件（仅在条件满足时执行）
// 示例：
//
//	ConditionalMiddleware("DevTools", 100,
//	    func(opts) bool { return opts.GetEnv() == "development" },
//	    func(bs, logger, opts) { /* 开发工具初始化 */ })
func ConditionalMiddleware(name string, priority int, condition func(opts Options) bool, fn MiddlewareFunc) MiddlewareConfig {
	return Middleware(name, priority, func(bs *bootstrap.Bootstrap, logger core.Logger, opts Options) error {
		if condition(opts) {
			logger.Infow("Condition met, applying middleware", "name", name)
			return fn(bs, logger, opts)
		}
		logger.Infow("Condition not met, skipping middleware", "name", name)
		return nil
	})
}
