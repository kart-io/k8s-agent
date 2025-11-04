package initializers

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/options"
	commonserver "github.com/kart-io/k8s-agent/common/server"
	httpserver "github.com/kart-io/k8s-agent/common/server/http"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
)

// HTTPServerConfig holds configuration for the HTTP server initializer.
type HTTPServerConfig struct {
	Name       string
	Priority   int
	Config     *options.ServerOptions
	RouteSetup func(*gin.Engine) error

	// Middleware configurations (optional)
	CORS      *options.CORSOptions
	JWT       *options.JWTOptions
	RateLimit *options.RateLimitOptions
}

// HTTPServerInitializer initializes a Gin HTTP server and integrates it with the bootstrap lifecycle.
// It implements bootstrap.Initializer and bootstrap.ServerProvider.
type HTTPServerInitializer struct {
	config *HTTPServerConfig
	logger core.Logger
	server commonserver.Server
}

// NewHTTPServerInitializer creates a new HTTPServerInitializer.
func NewHTTPServerInitializer(config *HTTPServerConfig, logger core.Logger) *HTTPServerInitializer {
	// Set default name and priority if not provided
	if config.Name == "" {
		config.Name = "http-server"
	}
	if config.Priority == 0 {
		config.Priority = bootstrap.PriorityHTTP
	}
	return &HTTPServerInitializer{
		config: config,
		logger: logger,
	}
}

// Name returns the initializer name.
func (i *HTTPServerInitializer) Name() string {
	return i.config.Name
}

// Priority returns the initializer priority.
func (i *HTTPServerInitializer) Priority() int {
	return i.config.Priority
}

// Initialize creates and configures the Gin server.
func (i *HTTPServerInitializer) Initialize(ctx context.Context) error {
	i.logger.Infow("Initializing HTTP server", "name", i.config.Name)

	// Create Gin server config with middleware options
	ginConfig := httpserver.NewGinServerConfig(i.config.Config)

	// Add middleware configurations if provided
	if i.config.CORS != nil {
		ginConfig.WithCORS(i.config.CORS)
	}

	if i.config.JWT != nil {
		ginConfig.WithJWT(i.config.JWT)
	}

	if i.config.RateLimit != nil {
		ginConfig.WithRateLimit(i.config.RateLimit)
	}

	// Create Gin server from full config
	ginServer := httpserver.NewGinServerFromFullConfig(i.logger, ginConfig)
	i.server = ginServer

	// Setup routes
	if i.config.RouteSetup != nil {
		if err := i.config.RouteSetup(ginServer.GetEngine()); err != nil {
			return fmt.Errorf("failed to setup HTTP routes for %s: %w", i.config.Name, err)
		}
		i.logger.Infow("HTTP routes configured", "name", i.config.Name)
	}

	i.logger.Infow("HTTP server initialized successfully",
		"name", i.config.Name,
		"addr", ginServer.Addr(),
		"cors_enabled", i.config.CORS != nil && i.config.CORS.Enabled,
		"jwt_configured", i.config.JWT != nil,
	)

	return nil
}

// GetServer returns the server instance to be managed by bootstrap.
// This implements the bootstrap.ServerProvider interface.
func (i *HTTPServerInitializer) GetServer() commonserver.Server {
	return i.server
}

// Close is a no-op for this initializer because the server's lifecycle
// is managed by the bootstrap's Run method, which calls GracefulStop.
func (i *HTTPServerInitializer) Close(ctx context.Context) error {
	return nil
}
