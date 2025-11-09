//go:build wireinject
// +build wireinject

// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/google/wire"

	"github.com/kart-io/k8s-agent/internal/agent-manager/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
)

// InitializerSet Wire dependency set for all initializers.
var InitializerSet = wire.NewSet(
	ProvideLogger,
	initializers.NewDatabaseInitializer,
	initializers.NewRedisInitializer,
	initializers.NewServiceInitializer,
	initializers.NewRegistryInitializer,
	initializers.NewNATSInitializer,
	initializers.NewDispatcherInitializer,
	initializers.NewHTTPServerInitializer,
	initializers.NewGRPCServerInitializer,
)

// HealthInitializerSet Wire dependency set for health check.
var HealthInitializerSet = wire.NewSet(
	pkginitializers.NewHealthCheckInitializer,
	wire.FieldsOf(new(*commonapp.StandardOptions), "Health"),
)

// InitializeAgentManagerContainer automatically injects all components using Wire.
func InitializeAgentManagerContainer(opts *commonapp.StandardOptions) (*AgentManagerContainer, error) {
	wire.Build(
		InitializerSet,
		HealthInitializerSet,
		NewAgentManagerContainer,
	)
	return nil, nil
}
