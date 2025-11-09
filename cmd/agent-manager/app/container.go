// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/kart-io/k8s-agent/internal/agent-manager/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// AgentManagerContainer contains all initialized components for the agent-manager service.
type AgentManagerContainer struct {
	db         *initializers.DatabaseInitializer
	redis      *initializers.RedisInitializer
	service    *initializers.ServiceInitializer
	registry   *initializers.RegistryInitializer
	nats       *initializers.NATSInitializer
	dispatcher *initializers.DispatcherInitializer
	http       *initializers.HTTPServerInitializer
	grpc       *initializers.GRPCServerInitializer
	health     *pkginitializers.HealthCheckInitializer
}

// NewAgentManagerContainer creates a new AgentManagerContainer instance.
// This is called by Wire with all dependencies automatically injected.
func NewAgentManagerContainer(
	db *initializers.DatabaseInitializer,
	redis *initializers.RedisInitializer,
	service *initializers.ServiceInitializer,
	registry *initializers.RegistryInitializer,
	nats *initializers.NATSInitializer,
	dispatcher *initializers.DispatcherInitializer,
	http *initializers.HTTPServerInitializer,
	grpc *initializers.GRPCServerInitializer,
	health *pkginitializers.HealthCheckInitializer,
) *AgentManagerContainer {
	return &AgentManagerContainer{
		db:         db,
		redis:      redis,
		service:    service,
		registry:   registry,
		nats:       nats,
		dispatcher: dispatcher,
		http:       http,
		grpc:       grpc,
		health:     health,
	}
}

// GetInitializers returns all initializers for Bootstrap framework.
// Bootstrap will automatically register them in priority order.
func (c *AgentManagerContainer) GetInitializers() []bootstrap.Initializer {
	return []bootstrap.Initializer{
		c.db,         // Priority 300
		c.redis,      // Priority 400
		c.service,    // Priority 600 - CRITICAL: Service layer MUST come before servers
		c.registry,   // Priority 450 (delegates to Service)
		c.nats,       // Priority 500
		c.dispatcher, // Priority 550 (delegates to Service)
		c.http,       // Priority 1000
		c.grpc,       // Priority 900 (may be nil if disabled)
		c.health,     // Priority 2000
	}
}

// ProvideLogger creates a logger instance from options.
func ProvideLogger(opts *commonapp.StandardOptions) (core.Logger, error) {
	return opts.InitLogger()
}
