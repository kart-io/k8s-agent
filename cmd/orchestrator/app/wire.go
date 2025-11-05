//go:build wireinject
// +build wireinject

// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/google/wire"
	"github.com/kart-io/k8s-agent/cmd/orchestrator/app/options"
	"github.com/kart-io/k8s-agent/internal/orchestrator/initializers"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
)

// BaseProviderSet Wire dependency set for base infrastructure.
var BaseProviderSet = wire.NewSet(
	ProvideLogger,
	initializers.NewDatabaseInitializer,
	initializers.NewRedisInitializer,
	initializers.NewNATSInitializer,
)

// BusinessProviderSet Wire dependency set for business logic.
var BusinessProviderSet = wire.NewSet(
	BaseProviderSet,
	initializers.NewWorkflowInitializer,
	initializers.NewStrategyInitializer,
	initializers.NewSubscriberInitializer,
)

// ServerProviderSet Wire dependency set for servers.
var ServerProviderSet = wire.NewSet(
	BusinessProviderSet,
	initializers.NewGRPCServerInitializer,
	initializers.NewHTTPServerInitializer,
)

// HealthProviderSet Wire dependency set for health check.
var HealthProviderSet = wire.NewSet(
	pkginitializers.NewHealthCheckInitializer,
	wire.FieldsOf(new(*options.ServerOptions), "Health"),
)

// InitializeOrchestratorComponents automatically injects all components using Wire.
func InitializeOrchestratorComponents(opts *options.ServerOptions) (*OrchestratorComponents, error) {
	wire.Build(
		ServerProviderSet,
		HealthProviderSet,
		NewOrchestratorComponents,
	)
	return nil, nil
}

