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

// HealthCheckFunc is a function that performs a health check.
type HealthCheckFunc func() error

// NewHealthCheckServer creates a health check server
// This is a convenience wrapper for common/health.
func NewHealthCheckServer(opts *options.HealthOptions, logger core.Logger) health.Server {
	return health.NewHTTPServer(opts, logger)
}

// DefaultHealthCheckFunc creates a simple health check function.
func DefaultHealthCheckFunc(healthOpts *options.HealthOptions) HealthCheckFunc {
	return func() error {
		server := NewHealthCheckServer(healthOpts, nil)
		return server.Start(context.Background())
	}
}
