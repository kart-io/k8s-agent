// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/kart-io/k8s-agent/cmd/gateway/app/options"
	"github.com/kart-io/k8s-agent/internal/gateway/initializers"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// GatewayComponents contains all component initializers.
type GatewayComponents struct {
	Redis  *initializers.RedisInitializer
	HTTP   *initializers.HTTPServerInitializer
	Health *pkginitializers.HealthCheckInitializer
}

// NewGatewayComponents creates a new GatewayComponents.
func NewGatewayComponents(
	redis *initializers.RedisInitializer,
	http *initializers.HTTPServerInitializer,
	health *pkginitializers.HealthCheckInitializer,
) *GatewayComponents {
	return &GatewayComponents{
		Redis:  redis,
		HTTP:   http,
		Health: health,
	}
}

// ProvideLogger provides logger from options.
func ProvideLogger(opts *options.ServerOptions) (core.Logger, error) {
	return opts.InitLogger()
}


