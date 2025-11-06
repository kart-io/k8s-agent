// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/kart-io/k8s-agent/cmd/monitor/app/options"
	"github.com/kart-io/k8s-agent/internal/monitor/initializers"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// MonitorComponents contains all component initializers.
type MonitorComponents struct {
	DB     *initializers.DatabaseInitializer
	Redis  *initializers.RedisInitializer
	HTTP   *initializers.HTTPServerInitializer
	GRPC   *initializers.GRPCServerInitializer
	Health *pkginitializers.HealthCheckInitializer
}

// NewMonitorComponents creates a new MonitorComponents.
func NewMonitorComponents(
	db *initializers.DatabaseInitializer,
	redis *initializers.RedisInitializer,
	http *initializers.HTTPServerInitializer,
	grpc *initializers.GRPCServerInitializer,
	health *pkginitializers.HealthCheckInitializer,
) *MonitorComponents {
	return &MonitorComponents{
		DB:     db,
		Redis:  redis,
		HTTP:   http,
		GRPC:   grpc,
		Health: health,
	}
}

// ProvideLogger provides logger from options.
func ProvideLogger(opts *options.ServerOptions) (core.Logger, error) {
	return opts.InitLogger()
}


