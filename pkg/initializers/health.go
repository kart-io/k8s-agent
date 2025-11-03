// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package initializers

import (
	"context"

	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/k8s-agent/common/health"
	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

// HealthCheckInitializer Bootstrap 适配器
type HealthCheckInitializer struct {
	server health.Server
}

// NewHealthCheckInitializer 创建初始化器
func NewHealthCheckInitializer(opts *options.HealthOptions, logger core.Logger) *HealthCheckInitializer {
	return &HealthCheckInitializer{
		server: health.NewHTTPServer(opts, logger),
	}
}

func (h *HealthCheckInitializer) Name() string                         { return "HealthCheck" }
func (h *HealthCheckInitializer) Priority() int                        { return bootstrap.PriorityLowest }
func (h *HealthCheckInitializer) Initialize(ctx context.Context) error { return h.server.Start(ctx) }
func (h *HealthCheckInitializer) Shutdown(ctx context.Context) error   { return h.server.Stop(ctx) }
