//go:build wireinject
// +build wireinject

// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/google/wire"
	"github.com/kart-io/k8s-agent/cmd/gateway/app/options"
	"github.com/kart-io/k8s-agent/internal/gateway/initializers"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
)

// InitializerSet Wire dependency set for all initializers.
var InitializerSet = wire.NewSet(
	ProvideLogger,
	initializers.NewRedisInitializer,
	initializers.NewHTTPServerInitializer,
)

// HealthInitializerSet Wire dependency set for health check.
var HealthInitializerSet = wire.NewSet(
	pkginitializers.NewHealthCheckInitializer,
	wire.FieldsOf(new(*options.ServerOptions), "Health"),
)

// InitializeGatewayComponents automatically injects all components using Wire.
func InitializeGatewayComponents(opts *options.ServerOptions) (*GatewayComponents, error) {
	wire.Build(
		InitializerSet,
		HealthInitializerSet,
		NewGatewayComponents,
	)
	return nil, nil
}


