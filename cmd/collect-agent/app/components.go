// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/internal/collect-agent/initializers"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// CollectAgentComponents contains all component initializers.
type CollectAgentComponents struct {
	Agent  *initializers.AgentInitializer
	Health *pkginitializers.HealthCheckInitializer
}

// NewCollectAgentComponents creates a new CollectAgentComponents.
func NewCollectAgentComponents(
	agent *initializers.AgentInitializer,
	health *pkginitializers.HealthCheckInitializer,
) *CollectAgentComponents {
	return &CollectAgentComponents{
		Agent:  agent,
		Health: health,
	}
}

// ProvideLogger provides logger from options.
func ProvideLogger(opts *commonapp.StandardOptions) (core.Logger, error) {
	return opts.InitLogger()
}
