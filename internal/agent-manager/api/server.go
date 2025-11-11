// Package api provides the HTTP API server implementation for the agent manager service.
package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/internal/agent-manager/agent"
	"github.com/kart-io/k8s-agent/internal/agent-manager/command"
	"github.com/kart-io/k8s-agent/internal/agent-manager/event"
	"github.com/kart-io/k8s-agent/internal/agent-manager/storage"
	"github.com/kart-io/k8s-agent/pkg/types"
	"github.com/kart-io/logger/core"
)

// Server represents the API server logic and handlers.
// It no longer manages the HTTP server lifecycle.
type Server struct {
	logger core.Logger

	// Components
	registry       *agent.Registry
	eventProcessor *event.Processor
	dispatcher     *command.Dispatcher
	store          *storage.MySQLStore
	cache          *storage.RedisStore

	// State
	startTime time.Time
}

// NewServer creates a new API server handler container.
func NewServer(
	_ types.ServerConfig, // Currently unused, will be used for configuration
	registry *agent.Registry,
	eventProcessor *event.Processor,
	dispatcher *command.Dispatcher,
	store *storage.MySQLStore,
	cache *storage.RedisStore,
	logger core.Logger,
) *Server {
	// Set gin mode to release by default
	gin.SetMode(gin.ReleaseMode)

	return &Server{
		logger:         logger.With("component", "api-server"),
		registry:       registry,
		eventProcessor: eventProcessor,
		dispatcher:     dispatcher,
		store:          store,
		cache:          cache,
		startTime:      time.Now(),
	}
}

// Health handlers

// HandleLiveness is the handler for liveness probe.
func (s *Server) HandleLiveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

// HandleReadiness is the handler for readiness probe.
func (s *Server) HandleReadiness(c *gin.Context) {
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

// HandleStatus provides detailed status of the service.
func (s *Server) HandleStatus(c *gin.Context) {
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

// HandleListAgents lists registered agents.
func (s *Server) HandleListAgents(c *gin.Context) {
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
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit < len(agents) {
			agents = agents[:parsedLimit]
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": agents,
		"count":  len(agents),
	})
}

// HandleGetAgent retrieves a specific agent by ID.
func (s *Server) HandleGetAgent(c *gin.Context) {
	agentID := c.Param("id")

	agent, err := s.registry.GetAgent(c.Request.Context(), agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	c.JSON(http.StatusOK, agent)
}

// HandleDeleteAgent deletes a specific agent by ID.
func (s *Server) HandleDeleteAgent(c *gin.Context) {
	agentID := c.Param("id")

	if err := s.registry.UnregisterAgent(c.Request.Context(), agentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "agent deleted"})
}

// Cluster handlers

// HandleListClusters lists all clusters.
func (s *Server) HandleListClusters(c *gin.Context) {
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

// HandleGetCluster retrieves a specific cluster by ID.
func (s *Server) HandleGetCluster(c *gin.Context) {
	clusterID := c.Param("id")

	cluster, err := s.store.GetCluster(c.Request.Context(), clusterID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}

	c.JSON(http.StatusOK, cluster)
}

// HandleCreateCluster creates a new cluster.
func (s *Server) HandleCreateCluster(c *gin.Context) {
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

// HandleUpdateCluster updates an existing cluster.
func (s *Server) HandleUpdateCluster(c *gin.Context) {
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

// HandleDeleteCluster deletes a cluster by ID.
func (s *Server) HandleDeleteCluster(c *gin.Context) {
	clusterID := c.Param("id")

	if err := s.store.DeleteCluster(c.Request.Context(), clusterID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cluster deleted"})
}

// HandleClusterHealth checks the health of a cluster's agent.
func (s *Server) HandleClusterHealth(c *gin.Context) {
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

// HandleListEvents lists events with optional filters.
func (s *Server) HandleListEvents(c *gin.Context) {
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

// HandleGetEvent retrieves a specific event by ID.
func (s *Server) HandleGetEvent(c *gin.Context) {
	eventID := c.Param("id")

	event, err := s.store.GetEvent(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// HandleSearchEvents searches for events using a filter from the request body.
func (s *Server) HandleSearchEvents(c *gin.Context) {
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

// HandleSendCommand dispatches a new command.
func (s *Server) HandleSendCommand(c *gin.Context) {
	var cmd types.Command
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.dispatcher.DispatchCommand(c.Request.Context(), &cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Start a background task to automatically correlate events to this command.
	go s.correlateEventsToCommand(cmd.ID, cmd.ClusterID, cmd.Namespace, cmd.IssuedBy)

	c.JSON(http.StatusCreated, cmd)
}

// HandleGetCommand retrieves a specific command by ID.
func (s *Server) HandleGetCommand(c *gin.Context) {
	commandID := c.Param("id")

	cmd, err := s.dispatcher.GetCommand(c.Request.Context(), commandID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "command not found"})
		return
	}

	c.JSON(http.StatusOK, cmd)
}

// HandleGetCommandResult retrieves the result of a command.
func (s *Server) HandleGetCommandResult(c *gin.Context) {
	commandID := c.Param("id")

	result, err := s.dispatcher.GetCommandResult(c.Request.Context(), commandID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "result not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// HandleGetCommandEvents retrieves all events associated with a command.
func (s *Server) HandleGetCommandEvents(c *gin.Context) {
	commandID := c.Param("id") // Get command ID from URL path

	// Query for all associated events
	var events []*types.Event
	if err := s.store.DB.WithContext(c.Request.Context()).
		Where("command_id = ?", commandID).
		Order("timestamp DESC").
		Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"command_id": commandID,
		"events":     events,
		"count":      len(events),
	})
}

// HandleListPendingCommands lists commands with optional filters.
func (s *Server) HandleListPendingCommands(c *gin.Context) {
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

// HandleRecordOperation records an operation to correlate with subsequent Kubernetes events.
func (s *Server) HandleRecordOperation(c *gin.Context) {
	var req struct {
		Command     string            `json:"command" binding:"required"`
		ClusterID   string            `json:"cluster_id"`
		Namespace   string            `json:"namespace"`
		User        string            `json:"user"`
		Description string            `json:"description"`
		Metadata    map[string]string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	correlationID := fmt.Sprintf("op-%d-%s", time.Now().UnixNano(), req.ClusterID)

	operationData := map[string]interface{}{
		"correlation_id": correlationID,
		"command":        req.Command,
		"cluster_id":     req.ClusterID,
		"namespace":      req.Namespace,
		"user":           req.User,
		"description":    req.Description,
		"timestamp":      time.Now(),
		"metadata":       req.Metadata,
	}

	go s.correlateEventsToOperation(correlationID, req.ClusterID, req.Namespace, req.User)

	c.JSON(http.StatusOK, gin.H{
		"correlation_id": correlationID,
		"message":        "Operation recorded. Events will be correlated automatically.",
		"operation":      operationData,
	})
}

// HandleGetOperationEvents retrieves all events associated with an operation.
func (s *Server) HandleGetOperationEvents(c *gin.Context) {
	correlationID := c.Param("id")

	var events []*types.Event
	if err := s.store.DB.WithContext(c.Request.Context()).
		Where("correlation_id = ?", correlationID).
		Order("timestamp DESC").
		Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"correlation_id": correlationID,
		"events":         events,
		"count":          len(events),
	})
}

// correlateEventsToOperation is a background task to listen for and correlate events.
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

			for _, event := range recentEvents {
				event.CorrelationID = correlationID
				event.TriggeredBy = user
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

// correlateEventsToCommand is a background task to listen for and correlate events to a command.
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

			for _, event := range recentEvents {
				event.CommandID = commandID
				event.TriggeredBy = issuedBy
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

// Helper functions

func ptrAgentStatus(status types.AgentStatus) *types.AgentStatus {
	return &status
}
