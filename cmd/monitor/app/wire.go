//go:build wireinject

// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/google/wire"

	"github.com/kart-io/k8s-agent/internal/monitor/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
)

// InitializeMonitorContainer automatically injects all components using Wire.
// Simplified: flat provider list without nested sets.
func InitializeMonitorContainer(opts *commonapp.StandardOptions) (*MonitorContainer, error) {
	wire.Build(
		// Logger provider
		ProvideLogger,

		// Extract Health options from StandardOptions
		wire.FieldsOf(new(*commonapp.StandardOptions), "Health"),

		// Initializers - flat list for simplicity
		initializers.NewDatabaseInitializer,
		initializers.NewRedisInitializer,
		initializers.NewHTTPServerInitializer,
		initializers.NewGRPCServerInitializer,
		pkginitializers.NewHealthCheckInitializer,

		// Component container
		NewMonitorContainer,
	)
	return nil, nil
}
