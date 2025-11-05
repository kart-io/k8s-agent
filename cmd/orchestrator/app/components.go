// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/kart-io/k8s-agent/cmd/orchestrator/app/options"
	"github.com/kart-io/k8s-agent/internal/orchestrator/initializers"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// OrchestratorComponents contains all component initializers.
type OrchestratorComponents struct {
	DB         *initializers.DatabaseInitializer
	Redis      *initializers.RedisInitializer
	NATS       *initializers.NATSInitializer
	Workflow   *initializers.WorkflowInitializer
	Strategy   *initializers.StrategyInitializer
	Subscriber *initializers.SubscriberInitializer
	GRPC       *initializers.GRPCServerInitializer
	HTTP       *initializers.HTTPServerInitializer
	Health     *pkginitializers.HealthCheckInitializer
}

// NewOrchestratorComponents creates a new OrchestratorComponents.
func NewOrchestratorComponents(
	db *initializers.DatabaseInitializer,
	redis *initializers.RedisInitializer,
	nats *initializers.NATSInitializer,
	workflow *initializers.WorkflowInitializer,
	strategy *initializers.StrategyInitializer,
	subscriber *initializers.SubscriberInitializer,
	grpc *initializers.GRPCServerInitializer,
	http *initializers.HTTPServerInitializer,
	health *pkginitializers.HealthCheckInitializer,
) *OrchestratorComponents {
	return &OrchestratorComponents{
		DB:         db,
		Redis:      redis,
		NATS:       nats,
		Workflow:   workflow,
		Strategy:   strategy,
		Subscriber: subscriber,
		GRPC:       grpc,
		HTTP:       http,
		Health:     health,
	}
}

// ProvideLogger provides logger from options.
func ProvideLogger(opts *options.ServerOptions) (core.Logger, error) {
	return opts.InitLogger()
}

