package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/cluster-service/internal/handler"
	"github.com/sirupsen/logrus"
)

type ServerConfig struct {
	Port         int
	Mode         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	JWTSecret    string
}

type Server struct {
	config  *ServerConfig
	handler *handler.ClusterHandler
	log     *logrus.Logger
	engine  *gin.Engine
	server  *http.Server
}

func NewServer(config *ServerConfig, handler *handler.ClusterHandler, logger *logrus.Logger) *Server {
	gin.SetMode(config.Mode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(ginLogger(logger))

	s := &Server{
		config:  config,
		handler: handler,
		log:     logger,
		engine:  engine,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Health check
	s.engine.GET("/health", s.handler.HealthCheck)

	// API routes
	v1 := s.engine.Group("/api/v1")
	{
		// Cluster routes
		clusters := v1.Group("/clusters")
		{
			clusters.POST("", s.handler.AddCluster)
			clusters.GET("/:id/health", s.handler.GetClusterHealth)

			// Pod routes
			clusters.GET("/:cluster_id/namespaces/:namespace/pods", s.handler.GetPods)
		}
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.engine,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	s.log.WithField("port", s.config.Port).Info("Starting cluster service")
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("Shutting down server...")
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// ginLogger returns a gin middleware that logs requests
func ginLogger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		entry := logger.WithFields(logrus.Fields{
			"status":     status,
			"method":     c.Request.Method,
			"path":       path,
			"query":      query,
			"ip":         c.ClientIP(),
			"latency":    latency,
			"user_agent": c.Request.UserAgent(),
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.String())
		} else {
			entry.Info("Request completed")
		}
	}
}
