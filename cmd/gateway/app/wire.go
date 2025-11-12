//go:build wireinject

// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/google/wire"

	"github.com/kart-io/k8s-agent/cmd/gateway/app/options"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// InitializeGatewayContainer uses Wire to inject dependencies.
func InitializeGatewayContainer(opts *options.ServerOptions, logger core.Logger) (*GatewayContainer, error) {
	wire.Build(
		// Simple initializers (all defined in container.go)
		NewRedisInitializer,
		NewHTTPServerInitializer,

		// Health check from pkg
		pkginitializers.NewHealthCheckInitializer,
		wire.FieldsOf(new(*options.ServerOptions), "Health"),

		// Bundle all together
		NewGatewayContainer,
	)
	return nil, nil
}
