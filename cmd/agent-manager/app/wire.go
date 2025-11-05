//go:build wireinject
// +build wireinject

// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/google/wire"
	"github.com/kart-io/k8s-agent/cmd/agent-manager/app/options"
	"github.com/kart-io/k8s-agent/internal/agent-manager/initializers"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
)

// BaseProviderSet Wire dependency set for base infrastructure.
var BaseProviderSet = wire.NewSet(
	ProvideLogger,
	initializers.NewDatabaseInitializer,
	initializers.NewRedisInitializer,
)

// BusinessProviderSet Wire dependency set for business logic.
var BusinessProviderSet = wire.NewSet(
	BaseProviderSet,
	initializers.NewRegistryInitializer,
	initializers.NewNATSInitializer,
	initializers.NewDispatcherInitializer,
)

// ServerProviderSet Wire dependency set for servers.
var ServerProviderSet = wire.NewSet(
	BusinessProviderSet,
	initializers.NewHTTPServerInitializer,
)

// HealthProviderSet Wire dependency set for health check.
var HealthProviderSet = wire.NewSet(
	pkginitializers.NewHealthCheckInitializer,
	wire.FieldsOf(new(*options.ServerOptions), "Health"),
)

// InitializeAgentManagerComponents automatically injects all components using Wire.
func InitializeAgentManagerComponents(opts *options.ServerOptions) (*AgentManagerComponents, error) {
	wire.Build(
		ServerProviderSet,
		HealthProviderSet,
		NewAgentManagerComponents,
	)
	return nil, nil
}

