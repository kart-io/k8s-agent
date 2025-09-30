package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/agent-manager/internal/agent"
	"github.com/kart-io/k8s-agent/agent-manager/internal/command"
	"github.com/kart-io/k8s-agent/agent-manager/internal/event"
	"github.com/kart-io/k8s-agent/agent-manager/internal/storage"
	"github.com/kart-io/k8s-agent/agent-manager/pkg/types"
)

// Server represents the API server
type Server struct {
	config     types.ServerConfig
	router     *gin.Engine
	httpServer *http.Server
	logger     *zap.Logger

	// Components
	registry       *agent.Registry
	eventProcessor *event.Processor
	dispatcher     *command.Dispatcher
	store          *storage.PostgresStore
	cache          *storage.RedisStore

	// State
	startTime time.Time
}

// NewServer creates a new API server
func NewServer(
	config types.ServerConfig,
	registry *agent.Registry,
	eventProcessor *event.Processor,
	dispatcher *command.Dispatcher,
	store *storage.PostgresStore,
	cache *storage.RedisStore,
	logger *zap.Logger,
) *Server {
	// Set gin mode
	if logger.Core().Enabled(zap.DebugLevel) {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	return &Server{
		config:         config,
		router:         gin.New(),
		logger:         logger.With(zap.String("component", "api-server")),
		registry:       registry,
		eventProcessor: eventProcessor,
		dispatcher:     dispatcher,
		store:          store,
		cache:          cache,
		startTime:      time.Now(),
	}
}

// Start starts the API server
func (s *Server) Start() error {
	// Setup middlewares
	s.setupMiddlewares()

	// Setup routes
	s.setupRoutes()

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	s.logger.Info("Starting API server", zap.String("addr", addr))

	// Start server
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Stop stops the API server gracefully
func (s *Server) Stop() error {
	s.logger.Info("Stopping API server")

	ctx, cancel := context.WithTimeout(context.Background(), s.config.GracefulStop)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	s.logger.Info("API server stopped")

	return nil
}

// setupMiddlewares sets up middleware chain
func (s *Server) setupMiddlewares() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Logger middleware
	s.router.Use(s.loggingMiddleware())

	// CORS middleware
	s.router.Use(s.corsMiddleware())

	// Request ID middleware
	s.router.Use(s.requestIDMiddleware())
}

// setupRoutes sets up API routes
func (s *Server) setupRoutes() {
	// Health endpoints
	health := s.router.Group("/health")
	{
		health.GET("/live", s.handleLiveness)
		health.GET("/ready", s.handleReadiness)
		health.GET("/status", s.handleStatus)
	}

	// Metrics endpoint
	s.router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1
	v1 := s.router.Group("/api/v1")
	{
		// Agent management
		agents := v1.Group("/agents")
		{
			agents.GET("", s.handleListAgents)
			agents.GET("/:id", s.handleGetAgent)
			agents.DELETE("/:id", s.handleDeleteAgent)
		}

		// Cluster management
		clusters := v1.Group("/clusters")
		{
			clusters.GET("", s.handleListClusters)
			clusters.GET("/:id", s.handleGetCluster)
			clusters.POST("", s.handleCreateCluster)
			clusters.PUT("/:id", s.handleUpdateCluster)
			clusters.DELETE("/:id", s.handleDeleteCluster)
			clusters.GET("/:id/health", s.handleClusterHealth)
		}

		// Event management
		events := v1.Group("/events")
		{
			events.GET("", s.handleListEvents)
			events.GET("/:id", s.handleGetEvent)
			events.POST("/search", s.handleSearchEvents)
		}

		// Command management
		commands := v1.Group("/commands")
		{
			commands.POST("", s.handleSendCommand)
			commands.GET("/:id", s.handleGetCommand)
			commands.GET("/:id/result", s.handleGetCommandResult)
			commands.GET("", s.handleListPendingCommands)
		}
	}
}

// Health handlers

func (s *Server) handleLiveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

func (s *Server) handleReadiness(c *gin.Context) {
	// Check database
	if err := s.store.Health(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"reason": "database unavailable",
		})
		return
	}

	// Check Redis
	if err := s.cache.Health(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"reason": "redis unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

func (s *Server) handleStatus(c *gin.Context) {
	ctx := c.Request.Context()

	onlineAgents, _ := s.registry.GetAgentCount(ctx, ptrAgentStatus(types.AgentStatusOnline))
	totalClusters, _ := s.store.ListClusters(ctx)

	status := types.HealthStatus{
		Status:           "healthy",
		Version:          "1.0.0",
		Uptime:           time.Since(s.startTime),
		ActiveAgents:     onlineAgents,
		TotalClusters:    len(totalClusters),
		EventsProcessed:  s.eventProcessor.GetStatistics()["events_processed"].(int64),
		CommandsIssued:   s.dispatcher.GetStatistics()["commands_issued"].(int64),
		Timestamp:        time.Now(),
		Components: map[string]interface{}{
			"registry":        s.registry.GetStatistics(),
			"event_processor": s.eventProcessor.GetStatistics(),
			"dispatcher":      s.dispatcher.GetStatistics(),
		},
	}

	c.JSON(http.StatusOK, status)
}

// Agent handlers

func (s *Server) handleListAgents(c *gin.Context) {
	var status *types.AgentStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s := types.AgentStatus(statusStr)
		status = &s
	}

	agents, err := s.registry.ListAgents(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": agents,
		"count":  len(agents),
	})
}

func (s *Server) handleGetAgent(c *gin.Context) {
	agentID := c.Param("id")

	agent, err := s.registry.GetAgent(c.Request.Context(), agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	c.JSON(http.StatusOK, agent)
}

func (s *Server) handleDeleteAgent(c *gin.Context) {
	agentID := c.Param("id")

	if err := s.registry.UnregisterAgent(c.Request.Context(), agentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "agent deleted"})
}

// Cluster handlers

func (s *Server) handleListClusters(c *gin.Context) {
	clusters, err := s.store.ListClusters(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"clusters": clusters,
		"count":    len(clusters),
	})
}

func (s *Server) handleGetCluster(c *gin.Context) {
	clusterID := c.Param("id")

	cluster, err := s.store.GetCluster(c.Request.Context(), clusterID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}

	c.JSON(http.StatusOK, cluster)
}

func (s *Server) handleCreateCluster(c *gin.Context) {
	var cluster types.Cluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cluster.CreatedAt = time.Now()
	cluster.UpdatedAt = time.Now()

	if err := s.store.SaveCluster(c.Request.Context(), &cluster); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, cluster)
}

func (s *Server) handleUpdateCluster(c *gin.Context) {
	clusterID := c.Param("id")

	var cluster types.Cluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cluster.ID = clusterID
	cluster.UpdatedAt = time.Now()

	if err := s.store.SaveCluster(c.Request.Context(), &cluster); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cluster)
}

func (s *Server) handleDeleteCluster(c *gin.Context) {
	clusterID := c.Param("id")

	if err := s.store.DeleteCluster(c.Request.Context(), clusterID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cluster deleted"})
}

func (s *Server) handleClusterHealth(c *gin.Context) {
	clusterID := c.Param("id")

	agent, err := s.registry.GetAgentByClusterID(c.Request.Context(), clusterID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster agent not found"})
		return
	}

	health := gin.H{
		"cluster_id": clusterID,
		"agent_status": agent.Status,
		"last_heartbeat": agent.LastHeartbeat,
		"healthy": agent.Status == types.AgentStatusOnline,
	}

	c.JSON(http.StatusOK, health)
}

// Event handlers

func (s *Server) handleListEvents(c *gin.Context) {
	filter := storage.EventFilter{
		ClusterID: c.Query("cluster_id"),
		Severity:  c.Query("severity"),
		Namespace: c.Query("namespace"),
		Limit:     100,
	}

	events, err := s.store.ListEvents(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"count":  len(events),
	})
}

func (s *Server) handleGetEvent(c *gin.Context) {
	eventID := c.Param("id")

	event, err := s.store.GetEvent(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

func (s *Server) handleSearchEvents(c *gin.Context) {
	var filter storage.EventFilter
	if err := c.ShouldBindJSON(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if filter.Limit == 0 {
		filter.Limit = 100
	}

	events, err := s.store.ListEvents(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"count":  len(events),
	})
}

// Command handlers

func (s *Server) handleSendCommand(c *gin.Context) {
	var cmd types.Command
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.dispatcher.DispatchCommand(c.Request.Context(), &cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, cmd)
}

func (s *Server) handleGetCommand(c *gin.Context) {
	commandID := c.Param("id")

	cmd, err := s.dispatcher.GetCommand(c.Request.Context(), commandID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "command not found"})
		return
	}

	c.JSON(http.StatusOK, cmd)
}

func (s *Server) handleGetCommandResult(c *gin.Context) {
	commandID := c.Param("id")

	result, err := s.dispatcher.GetCommandResult(c.Request.Context(), commandID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "result not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (s *Server) handleListPendingCommands(c *gin.Context) {
	commands := s.dispatcher.GetPendingCommands()

	c.JSON(http.StatusOK, gin.H{
		"commands": commands,
		"count":    len(commands),
	})
}

// Middlewares

func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		s.logger.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()))
	}
}

func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func (s *Server) requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
		}

		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		c.Next()
	}
}

// Helper functions

func ptrAgentStatus(status types.AgentStatus) *types.AgentStatus {
	return &status
}