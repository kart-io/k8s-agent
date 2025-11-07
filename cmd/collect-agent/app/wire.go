//go:build wireinject
// +build wireinject

// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/google/wire"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/internal/collect-agent/initializers"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
)

// InitializerSet Wire dependency set for all initializers.
var InitializerSet = wire.NewSet(
	ProvideLogger,
	initializers.NewAgentInitializer,
)

// HealthInitializerSet Wire dependency set for health check.
var HealthInitializerSet = wire.NewSet(
	pkginitializers.NewHealthCheckInitializer,
	wire.FieldsOf(new(*commonapp.StandardOptions), "Health"),
)

// InitializeCollectAgentComponents automatically injects all components using Wire.
func InitializeCollectAgentComponents(opts *commonapp.StandardOptions) (*CollectAgentComponents, error) {
	wire.Build(
		InitializerSet,
		HealthInitializerSet,
		NewCollectAgentComponents,
	)
	return nil, nil
}
