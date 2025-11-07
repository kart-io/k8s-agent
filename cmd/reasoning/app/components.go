// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/internal/reasoning/initializers"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// ReasoningComponents contains all component initializers.
type ReasoningComponents struct {
	LLM           *initializers.LLMInitializer
	UnifiedServer *initializers.UnifiedServerInitializer
	Health        *pkginitializers.HealthCheckInitializer
}

// NewReasoningComponents creates a new ReasoningComponents.
func NewReasoningComponents(
	llm *initializers.LLMInitializer,
	unifiedServer *initializers.UnifiedServerInitializer,
	health *pkginitializers.HealthCheckInitializer,
) *ReasoningComponents {
	return &ReasoningComponents{
		LLM:           llm,
		UnifiedServer: unifiedServer,
		Health:        health,
	}
}

// ProvideLogger provides logger from options.
func ProvideLogger(opts *commonapp.StandardOptions) (core.Logger, error) {
	return opts.InitLogger()
}

