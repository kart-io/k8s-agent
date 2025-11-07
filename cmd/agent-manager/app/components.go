// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/kart-io/k8s-agent/internal/agent-manager/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// AgentManagerComponents contains all component initializers.
type AgentManagerComponents struct {
	DB         *initializers.DatabaseInitializer
	Redis      *initializers.RedisInitializer
	Registry   *initializers.RegistryInitializer
	NATS       *initializers.NATSInitializer
	Dispatcher *initializers.DispatcherInitializer
	HTTP       *initializers.HTTPServerInitializer
	GRPC       *initializers.GRPCServerInitializer // Can be nil if disabled
	Health     *pkginitializers.HealthCheckInitializer
}

// NewAgentManagerComponents creates a new AgentManagerComponents.
func NewAgentManagerComponents(
	db *initializers.DatabaseInitializer,
	redis *initializers.RedisInitializer,
	registry *initializers.RegistryInitializer,
	nats *initializers.NATSInitializer,
	dispatcher *initializers.DispatcherInitializer,
	http *initializers.HTTPServerInitializer,
	health *pkginitializers.HealthCheckInitializer,
) *AgentManagerComponents {
	return &AgentManagerComponents{
		DB:         db,
		Redis:      redis,
		Registry:   registry,
		NATS:       nats,
		Dispatcher: dispatcher,
		HTTP:       http,
		Health:     health,
	}
}

// ProvideLogger provides logger from StandardOptions.
func ProvideLogger(opts *commonapp.StandardOptions) (core.Logger, error) {
	return opts.InitLogger()
}

