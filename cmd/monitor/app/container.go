// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/kart-io/k8s-agent/internal/monitor/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// MonitorContainer contains all component initializers.
//
// SIMPLIFIED ARCHITECTURE:
// This service uses a flat Wire configuration without nested provider sets.
// All dependencies are injected directly in wire.go for clarity and simplicity.
type MonitorContainer struct {
	DB     *initializers.DatabaseInitializer
	Redis  *initializers.RedisInitializer
	HTTP   *initializers.HTTPServerInitializer
	GRPC   *initializers.GRPCServerInitializer
	Health *pkginitializers.HealthCheckInitializer
}

// NewMonitorContainer creates a new MonitorContainer.
func NewMonitorContainer(
	db *initializers.DatabaseInitializer,
	redis *initializers.RedisInitializer,
	http *initializers.HTTPServerInitializer,
	grpc *initializers.GRPCServerInitializer,
	health *pkginitializers.HealthCheckInitializer,
) *MonitorContainer {
	return &MonitorContainer{
		DB:     db,
		Redis:  redis,
		HTTP:   http,
		GRPC:   grpc,
		Health: health,
	}
}

// ProvideLogger provides logger from options.
func ProvideLogger(opts *commonapp.StandardOptions) (core.Logger, error) {
	return opts.InitLogger()
}
