package app

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/common/options"
	commonserver "github.com/kart-io/k8s-agent/common/server"
	httpserver "github.com/kart-io/k8s-agent/common/server/http"
	"github.com/kart-io/k8s-agent/internal/collect-agent/agent"
	"github.com/kart-io/k8s-agent/internal/collect-agent/config"
	"github.com/kart-io/logger/core"
)

// CollectAgentService represents the collect-agent service using common/server
type CollectAgentService struct {
	opts          *config.Options
	log           core.Logger
	agentInstance *agent.Agent
	healthServer  commonserver.Server
}

// NewServer creates a new collect-agent service (使用 common/server)
func NewServer(opts *config.Options, log core.Logger) (*CollectAgentService, error) {
	srv := &CollectAgentService{
		opts: opts,
		log:  log,
	}

	if err := srv.initialize(); err != nil {
		return nil, err
	}

	return srv, nil
}

// initialize initializes all server components
func (s *CollectAgentService) initialize() error {
	var err error

	// Convert Options to AgentConfig for backward compatibility with agent package
	agentConfig := s.opts.ToAgentConfig()

	// Create agent instance (now uses core.Logger directly)
	s.agentInstance, err = agent.New(agentConfig, s.log)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	// Determine health port
	port := s.opts.Agent.HealthPort
	if port == 0 {
		port = 8080 // default
	}

	// Create health server using common/server
	ginConfig := httpserver.NewGinServerConfig(&options.ServerOptions{
		Host: "",
		Port: port,
		Mode: "release",
	})

	ginServer := httpserver.NewGinServerFromFullConfig(s.log, ginConfig)
	engine := ginServer.GetEngine()

	// Register health check handlers
	s.setupHealthRoutes(engine)

	s.healthServer = ginServer

	return nil
}

// setupHealthRoutes sets up health check routes
func (s *CollectAgentService) setupHealthRoutes(engine *gin.Engine) {
	health := engine.Group("/health")
	{
		health.GET("/live", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "ok",
				"time":   "placeholder", // Can be enhanced
			})
		})

		health.GET("/ready", func(c *gin.Context) {
			if s.agentInstance == nil {
				c.JSON(503, gin.H{
					"status":  "not ready",
					"message": "agent not initialized",
				})
				return
			}
			c.JSON(200, gin.H{
				"status": "ready",
			})
		})

		health.GET("/status", func(c *gin.Context) {
			if s.agentInstance == nil {
				c.JSON(503, gin.H{
					"status":  "unhealthy",
					"message": "agent not initialized",
				})
				return
			}
			c.JSON(200, gin.H{
				"status": "healthy",
				"agent":  "collect-agent",
			})
		})
	}

	// Metrics endpoint (basic)
	engine.GET("/metrics", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"metrics": "placeholder", // Can be enhanced with Prometheus metrics
		})
	})

	s.log.Infow("Health check routes configured", "port", s.opts.Agent.HealthPort)
}

// Run starts the collect-agent service
func (s *CollectAgentService) Run(ctx context.Context) error {
	s.log.Infow("Starting Collect Agent Service",
		"health_port", s.opts.Agent.HealthPort,
	)

	// Create context for agent
	agentCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start agent in background
	agentErrChan := make(chan error, 1)
	go func() {
		s.log.Info("Starting agent services...")
		if err := s.agentInstance.Start(agentCtx); err != nil {
			agentErrChan <- err
		}
		close(agentErrChan)
	}()

	// Start health server using common/server
	// This will block until signal or error
	serverErrChan := make(chan error, 1)
	go func() {
		serverErrChan <- commonserver.Serve(ctx, s.healthServer, s.log)
	}()

	// Wait for either agent or server to fail, or context cancellation
	select {
	case err := <-agentErrChan:
		if err != nil {
			s.log.Errorw("Agent failed", "error", err)
			return fmt.Errorf("agent failed: %w", err)
		}
	case err := <-serverErrChan:
		if err != nil {
			s.log.Errorw("Health server failed", "error", err)
		}
		cancel() // Stop agent
		return err
	case <-ctx.Done():
		s.log.Info("Context cancelled, shutting down...")
		cancel() // Stop agent
		// Wait for agent to finish
		<-agentErrChan
	}

	s.log.Info("Collect Agent shutdown complete")
	return nil
}

// GetServer returns the health server instance (实现 ServerProvider 接口)
func (s *CollectAgentService) GetServer() commonserver.Server {
	return s.healthServer
}
