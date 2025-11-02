package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/logger/core"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/kart-io/k8s-agent/common/idempotent"
	"github.com/kart-io/k8s-agent/common/middleware"
	"github.com/kart-io/k8s-agent/internal/agent-manager/agent"
	"github.com/kart-io/k8s-agent/internal/agent-manager/command"
	"github.com/kart-io/k8s-agent/internal/agent-manager/event"
	"github.com/kart-io/k8s-agent/internal/agent-manager/storage"
	"github.com/kart-io/k8s-agent/pkg/types"
)

// Server represents the API server
type Server struct {
	config     types.ServerConfig
	router     *gin.Engine
	httpServer *http.Server
	logger     core.Logger

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
	logger core.Logger,
) *Server {
	// Set gin mode to release by default
	gin.SetMode(gin.ReleaseMode)

	return &Server{
		config:         config,
		router:         gin.New(),
		logger:         logger.With("component", "api-server"),
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

	s.logger.Infow("Starting API server", "addr", addr)

	// Start server
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Stop stops the API server gracefully
func (s *Server) Stop() error {
	s.logger.Infow("Stopping API server")

	ctx, cancel := context.WithTimeout(context.Background(), s.config.GracefulStop)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	s.logger.Infow("API server stopped")

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

	// Idempotency middleware (for POST operations)
	// Create Redis-backed idempotency store
	if s.cache != nil && s.cache.Client != nil {
		redisStore := idempotent.NewRedisStore(s.cache.Client, "agent-manager")
		idempotentHandler := idempotent.NewHandler(redisStore, 24*time.Hour, 5*time.Minute)

		s.router.Use(middleware.Idempotent(middleware.IdempotentConfig{
			Handler: idempotentHandler,
			// Use default path blacklist which includes:
			// - POST /api/v1/commands
			// - POST /api/v1/events
			// - POST /api/v1/agents
			// - POST /api/v1/clusters
			PathBlacklist: middleware.DefaultPathBlacklist(),
		}))

		s.logger.Info("Idempotency middleware enabled for POST operations")
	} else {
		s.logger.Warn("Redis not available, idempotency middleware disabled")
	}
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
			commands.GET("/:id/events", s.handleGetCommandEvents)
			commands.GET("", s.handleListPendingCommands)
		}

		// Operation tracking for correlating external commands with events
		operations := v1.Group("/operations")
		{
			operations.POST("", s.handleRecordOperation)
			operations.GET("/:id/events", s.handleGetOperationEvents)
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
		Status:          "healthy",
		Version:         "1.0.0",
		Uptime:          time.Since(s.startTime),
		ActiveAgents:    onlineAgents,
		TotalClusters:   len(totalClusters),
		EventsProcessed: s.eventProcessor.GetStatistics()["events_processed"].(int64),
		CommandsIssued:  s.dispatcher.GetStatistics()["commands_issued"].(int64),
		Timestamp:       time.Now(),
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

	// Apply limit if specified
	limit := len(agents)
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit < len(agents) {
			limit = parsedLimit
			agents = agents[:limit]
		}
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

	// Apply limit if specified
	limit := len(clusters)
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit < len(clusters) {
			limit = parsedLimit
			clusters = clusters[:limit]
		}
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
		"cluster_id":     clusterID,
		"agent_status":   agent.Status,
		"last_heartbeat": agent.LastHeartbeat,
		"healthy":        agent.Status == types.AgentStatusOnline,
	}

	c.JSON(http.StatusOK, health)
}

// Event handlers

func (s *Server) handleListEvents(c *gin.Context) {
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	filter := storage.EventFilter{
		ClusterID: c.Query("cluster_id"),
		Severity:  c.Query("severity"),
		Namespace: c.Query("namespace"),
		Limit:     limit,
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

	// 启动后台任务，自动关联命令产生的事件
	// 在接下来30秒内，将匹配的事件的 command_id 设置为该命令的 ID
	go s.correlateEventsToCommand(cmd.ID, cmd.ClusterID, cmd.Namespace, cmd.IssuedBy)

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

// handleGetCommandEvents 获取与命令关联的所有事件
// GET /api/v1/commands/:id/events
// 路径参数: id - 命令ID (command_id)
func (s *Server) handleGetCommandEvents(c *gin.Context) {
	commandID := c.Param("id") // 从URL路径获取命令ID

	// 查询所有关联的事件
	var events []*types.Event
	if err := s.store.DB.WithContext(c.Request.Context()).
		Where("command_id = ?", commandID).
		Order("timestamp DESC").
		Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回参数
	c.JSON(http.StatusOK, gin.H{
		"command_id": commandID,   // 命令ID
		"events":     events,      // 事件列表，包含 reason, namespace, timestamp, triggered_by 等字段
		"count":      len(events), // 事件总数
	})
}

func (s *Server) handleListPendingCommands(c *gin.Context) {
	// Get optional query parameters
	clusterID := c.Query("cluster_id")
	status := c.Query("status")

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Query database for commands with filters
	var commands []*types.Command
	query := s.store.DB.WithContext(c.Request.Context())

	if clusterID != "" {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Order by creation time descending and apply limit
	if err := query.Order("created_at DESC").Limit(limit).Find(&commands).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"commands": commands,
		"count":    len(commands),
	})
}

// Operation tracking handlers

// handleRecordOperation 记录操作信息，用于关联后续触发的 Kubernetes 事件
// POST /api/v1/operations
func (s *Server) handleRecordOperation(c *gin.Context) {
	// 请求参数
	var req struct {
		Command     string            `json:"command" binding:"required"` // 执行的命令，如 "kubectl scale deployment redis --replicas=0"
		ClusterID   string            `json:"cluster_id"`                 // 集群ID
		Namespace   string            `json:"namespace"`                  // 命名空间，用于过滤事件
		User        string            `json:"user"`                       // 执行操作的用户名
		Description string            `json:"description"`                // 操作描述
		Metadata    map[string]string `json:"metadata"`                   // 额外的元数据
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成关联ID，用于跟踪相关事件
	// 格式: op-{纳秒时间戳}-{集群ID}
	// 例如: op-1759461933026009000-k8s-663078ee
	// 这个ID会在后续30秒内自动写入到相关事件的 correlation_id 字段
	correlationID := fmt.Sprintf("op-%d-%s", time.Now().UnixNano(), req.ClusterID)

	// 存储操作数据
	operationData := map[string]interface{}{
		"correlation_id": correlationID,   // 关联ID
		"command":        req.Command,     // 执行的命令
		"cluster_id":     req.ClusterID,   // 集群ID
		"namespace":      req.Namespace,   // 命名空间
		"user":           req.User,        // 用户名
		"description":    req.Description, // 操作描述
		"timestamp":      time.Now(),      // 记录时间
		"metadata":       req.Metadata,    // 元数据
	}

	// 启动后台任务，关联接下来30秒内的事件
	go s.correlateEventsToOperation(correlationID, req.ClusterID, req.Namespace, req.User)

	// 返回参数
	c.JSON(http.StatusOK, gin.H{
		"correlation_id": correlationID,                                                  // 关联ID，用于后续查询事件
		"message":        "Operation recorded. Events will be correlated automatically.", // 提示消息
		"operation":      operationData,                                                  // 操作详情
	})
}

// handleGetOperationEvents 获取与操作关联的所有事件
// GET /api/v1/operations/:id/events
// 路径参数: id - 关联ID (correlation_id)
func (s *Server) handleGetOperationEvents(c *gin.Context) {
	correlationID := c.Param("id") // 从URL路径获取关联ID

	// 查询所有关联的事件
	var events []*types.Event
	if err := s.store.DB.WithContext(c.Request.Context()).
		Where("correlation_id = ?", correlationID).
		Order("timestamp DESC").
		Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回参数
	c.JSON(http.StatusOK, gin.H{
		"correlation_id": correlationID, // 关联ID
		"events":         events,        // 事件列表，包含 reason, namespace, timestamp, triggered_by 等字段
		"count":          len(events),   // 事件总数
	})
}

// correlateEventsToOperation 后台任务：在接下来30秒内监听事件并自动关联
// 参数:
//   - correlationID: 关联ID
//   - clusterID: 集群ID
//   - namespace: 命名空间（可选，用于过滤）
//   - user: 执行操作的用户名
func (s *Server) correlateEventsToOperation(correlationID, clusterID, namespace, user string) {
	startTime := time.Now()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(30 * time.Second)

	for {
		select {
		case <-timeout:
			s.logger.Debugw("Operation correlation timeout", "correlation_id", correlationID)
			return
		case <-ticker.C:
			// Find recent events (last 5 seconds) that match cluster and namespace
			var recentEvents []*types.Event
			query := s.store.DB.
				Where("cluster_id = ?", clusterID).
				Where("timestamp >= ?", startTime).
				Where("correlation_id IS NULL OR correlation_id = ''")

			if namespace != "" {
				query = query.Where("namespace = ?", namespace)
			}

			if err := query.Find(&recentEvents).Error; err != nil {
				s.logger.Errorw("Failed to query events for correlation",
					"error", err,
					"correlation_id", correlationID)
				continue
			}

			// 更新事件的关联信息
			// 将找到的事件的 correlation_id 字段设置为操作的 correlation_id
			// 这样就建立了"操作 -> 事件"的关联关系
			for _, event := range recentEvents {
				event.CorrelationID = correlationID // 设置事件的关联ID为操作ID
				event.TriggeredBy = user            // 记录触发者
				if err := s.store.DB.Save(event).Error; err != nil {
					s.logger.Errorw("Failed to update event correlation",
						"error", err,
						"event_id", event.ID)
				} else {
					s.logger.Debugw("Event correlated to operation",
						"event_id", event.ID,
						"correlation_id", correlationID,
						"reason", event.Reason)
				}
			}
		}
	}
}

// correlateEventsToCommand 后台任务：在接下来30秒内监听事件并自动关联到命令
// 参数:
//   - commandID: 命令ID
//   - clusterID: 集群ID
//   - namespace: 命名空间（可选，用于过滤）
//   - issuedBy: 发送命令的用户
func (s *Server) correlateEventsToCommand(commandID, clusterID, namespace, issuedBy string) {
	startTime := time.Now()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(30 * time.Second)

	for {
		select {
		case <-timeout:
			s.logger.Debugw("Command correlation timeout", "command_id", commandID)
			return
		case <-ticker.C:
			// 查找最近的事件并关联到命令
			var recentEvents []*types.Event
			query := s.store.DB.
				Where("cluster_id = ?", clusterID).
				Where("timestamp >= ?", startTime).
				Where("command_id IS NULL OR command_id = ''")

			if namespace != "" {
				query = query.Where("namespace = ?", namespace)
			}

			if err := query.Find(&recentEvents).Error; err != nil {
				s.logger.Errorw("Failed to query events for command correlation",
					"error", err,
					"command_id", commandID)
				continue
			}

			// 更新事件的命令关联信息
			// 将找到的事件的 command_id 字段设置为该命令的 ID
			// 这样就建立了"命令 -> 事件"的关联关系
			for _, event := range recentEvents {
				event.CommandID = commandID  // 设置事件的命令ID
				event.TriggeredBy = issuedBy // 记录触发者
				if err := s.store.DB.Save(event).Error; err != nil {
					s.logger.Errorw("Failed to update event command correlation",
						"error", err,
						"event_id", event.ID)
				} else {
					s.logger.Debugw("Event correlated to command",
						"event_id", event.ID,
						"command_id", commandID,
						"reason", event.Reason)
				}
			}
		}
	}
}

// Middlewares

func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		s.logger.Infow("HTTP request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", duration,
			"client_ip", c.ClientIP())
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
