// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package orchestrator

import (
	"os"
	"time"

	commonoptions "github.com/kart-io/k8s-agent/common/options"
)

var (
	// Name is the name of the compiled software.
	Name = "aetherius-orchestrator"

	// ID is the hostname of the machine running the orchestrator service.
	ID, _ = os.Hostname()
)

// Config contains application-related configurations for the orchestrator service.
// This is the business-layer configuration, converted from startup-layer ServerOptions.
type Config struct {
	// Server options for HTTP server configuration
	Server *commonoptions.ServerOptions

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

	// AI options for AI service integration
	AI *AIConfig
}

// AIConfig contains AI service configuration.
type AIConfig struct {
	// ReasoningServiceURL is the URL of the reasoning service
	ReasoningServiceURL string

	// AgentManagerURL is the URL of the agent manager service
	AgentManagerURL string

	// Timeout is the request timeout for AI service calls
	Timeout time.Duration

	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
}
