// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package agentmanager

import (
	"os"

	commonoptions "github.com/kart-io/k8s-agent/common/options"
)

var (
	// Name is the name of the compiled software.
	Name = "aetherius-agent-manager"

	// ID is the hostname of the machine running the agent-manager service.
	ID, _ = os.Hostname()
)

// Config contains application-related configurations for the agent-manager service.
// This is the business-layer configuration, converted from startup-layer ServerOptions.
type Config struct {
	// Server options for HTTP server configuration
	Server *commonoptions.ServerOptions

	// GRPC options for gRPC server configuration
	GRPC *commonoptions.GRPCOptions

	// Database options for MySQL database configuration
	Database *commonoptions.DatabaseOptions

	// Redis options for Redis cache configuration
	Redis *commonoptions.RedisOptions

	// NATS options for NATS message queue configuration
	NATS *commonoptions.NATSOptions

	// Logging options for logging configuration
	Logging *commonoptions.LoggingOptions

	// Metrics options for Prometheus metrics configuration
	Metrics *commonoptions.MetricsOptions
}
