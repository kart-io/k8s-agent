//go:build wireinject
// +build wireinject

// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/google/wire"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/internal/cluster/initializers"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
)

// InitializerSet Wire dependency set for all initializers.
var InitializerSet = wire.NewSet(
	ProvideLogger,
	initializers.NewDatabaseInitializer,
	initializers.NewHTTPServerInitializer,
	initializers.NewGRPCServerInitializer,
)

// HealthInitializerSet Wire dependency set for health check.
var HealthInitializerSet = wire.NewSet(
	pkginitializers.NewHealthCheckInitializer,
	wire.FieldsOf(new(*commonapp.StandardOptions), "Health"),
)

// InitializeClusterComponents automatically injects all components using Wire.
func InitializeClusterComponents(opts *commonapp.StandardOptions) (*ClusterComponents, error) {
	wire.Build(
		InitializerSet,
		HealthInitializerSet,
		NewClusterComponents,
	)
	return nil, nil
}

