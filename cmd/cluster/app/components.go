// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/internal/cluster/initializers"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// ClusterComponents contains all component initializers.
type ClusterComponents struct {
	DB     *initializers.DatabaseInitializer
	HTTP   *initializers.HTTPServerInitializer
	GRPC   *initializers.GRPCServerInitializer
	Health *pkginitializers.HealthCheckInitializer
}

// NewClusterComponents creates a new ClusterComponents.
func NewClusterComponents(
	db *initializers.DatabaseInitializer,
	http *initializers.HTTPServerInitializer,
	grpc *initializers.GRPCServerInitializer,
	health *pkginitializers.HealthCheckInitializer,
) *ClusterComponents {
	return &ClusterComponents{
		DB:     db,
		HTTP:   http,
		GRPC:   grpc,
		Health: health,
	}
}

// ProvideLogger provides logger from options.
func ProvideLogger(opts *commonapp.StandardOptions) (core.Logger, error) {
	return opts.InitLogger()
}


