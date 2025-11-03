// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"

	"github.com/kart-io/k8s-agent/common/health"
	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

// HealthCheckFunc is a function that performs a health check
type HealthCheckFunc func() error

// NewHealthCheckServer creates a health check server
// This is a convenience wrapper for common/health
func NewHealthCheckServer(opts *options.HealthOptions, logger core.Logger) health.Server {
	return health.NewHTTPServer(opts, logger)
}

// GetHealthOptions gets health check options from configuration
func GetHealthOptions(opts Options) *options.HealthOptions {
	// Direct configuration - no more compatibility layers
	// Services should directly provide health options in their config
	if healthOpts := extractHealthOptions(opts); healthOpts != nil {
		return healthOpts
	}

	// Default configuration
	return options.NewHealthOptions()
}

// extractHealthOptions attempts to extract health options from the options interface
// This replaces the multiple compatibility layers with a single, simple extraction
func extractHealthOptions(opts interface{}) *options.HealthOptions {
	// Try to extract health options from the configuration struct
	// This assumes services embed HealthOptions in their config
	type healthConfig interface {
		GetHealthOptions() *options.HealthOptions
	}

	if cfg, ok := opts.(healthConfig); ok {
		return cfg.GetHealthOptions()
	}

	return nil
}

// DefaultHealthCheckFunc creates a simple health check function
func DefaultHealthCheckFunc(healthOpts *options.HealthOptions) HealthCheckFunc {
	return func() error {
		server := NewHealthCheckServer(healthOpts, nil)
		return server.Start(context.Background())
	}
}