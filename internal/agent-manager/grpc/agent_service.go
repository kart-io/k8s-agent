package grpc

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/kart-io/k8s-agent/internal/agent-manager/agent"
	agentv1 "github.com/kart-io/k8s-agent/pkg/api/agent/v1"
	paginationv1 "github.com/kart-io/k8s-agent/pkg/api/common/pagination/v1"
	"github.com/kart-io/k8s-agent/pkg/types"
	"github.com/kart-io/logger/core"
)

// AgentServiceServer implements the AgentService gRPC service.
type AgentServiceServer struct {
	agentv1.UnimplementedAgentServiceServer
	registry *agent.Registry
	logger   core.Logger
}

// NewAgentServiceServer creates a new AgentServiceServer instance.
func NewAgentServiceServer(registry *agent.Registry, logger core.Logger) *AgentServiceServer {
	return &AgentServiceServer{
		registry: registry,
		logger:   logger,
	}
}

// RegisterAgent registers a new agent.
func (s *AgentServiceServer) RegisterAgent(ctx context.Context, req *agentv1.RegisterAgentRequest) (*agentv1.RegisterAgentResponse, error) {
	s.logger.Infow("Registering agent",
		"name", req.Name,
		"cluster_id", req.ClusterId,
		"cluster_name", req.ClusterName,
		"version", req.Version,
	)

	// Validate request
	if req.ClusterId == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster_id is required")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	// Create agent
	agentID := uuid.New().String()
	now := time.Now()

	agentData := &types.Agent{
		ID:            agentID,
		ClusterID:     req.ClusterId,
		ClusterName:   req.ClusterName,
		Version:       req.Version,
		Status:        types.AgentStatusOnline,
		LastHeartbeat: now,
		RegisteredAt:  now,
		UpdatedAt:     now,
		Metadata:      convertMetadata(req.Metadata),
	}

	// Register agent using RegisterAgent method
	if err := s.registry.RegisterAgent(ctx, agentData); err != nil {
		s.logger.Errorw("Failed to register agent",
			"error", err,
			"agent_id", agentID,
		)
		return nil, status.Errorf(codes.Internal, "failed to register agent: %v", err)
	}

	// Generate token (simplified - in production, use proper JWT)
	token := generateAgentToken(agentID)

	// Convert to protobuf
	pbAgent := convertAgentToProto(agentData)

	s.logger.Infow("Agent registered successfully",
		"agent_id", agentID,
		"cluster_id", req.ClusterId,
	)

	return &agentv1.RegisterAgentResponse{
		Agent: pbAgent,
		Token: token,
	}, nil
}

// Heartbeat processes agent heartbeat.
func (s *AgentServiceServer) Heartbeat(ctx context.Context, req *agentv1.HeartbeatRequest) (*agentv1.HeartbeatResponse, error) {
	if req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "agent_id is required")
	}

	// Update heartbeat - Registry.UpdateHeartbeat only takes agentID
	if err := s.registry.UpdateHeartbeat(ctx, req.AgentId); err != nil {
		s.logger.Warnw("Failed to update heartbeat",
			"error", err,
			"agent_id", req.AgentId,
		)
		return nil, status.Errorf(codes.Internal, "failed to update heartbeat: %v", err)
	}

	s.logger.Debugw("Heartbeat received",
		"agent_id", req.AgentId,
		"status", req.Status,
	)

	return &agentv1.HeartbeatResponse{
		Success:               true,
		NextHeartbeatInterval: 30, // 30 seconds
	}, nil
}

// GetAgent retrieves agent information.
func (s *AgentServiceServer) GetAgent(ctx context.Context, req *agentv1.GetAgentRequest) (*agentv1.GetAgentResponse, error) {
	if req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "agent_id is required")
	}

	// Get agent from registry
	agentData, err := s.registry.GetAgent(ctx, req.AgentId)
	if err != nil {
		s.logger.Warnw("Failed to get agent",
			"error", err,
			"agent_id", req.AgentId,
		)
		return nil, status.Errorf(codes.NotFound, "agent not found: %v", err)
	}

	// Convert to protobuf
	pbAgent := convertAgentToProto(agentData)

	return &agentv1.GetAgentResponse{
		Agent: pbAgent,
	}, nil
}

// ListAgents lists all agents with optional filters.
func (s *AgentServiceServer) ListAgents(ctx context.Context, req *agentv1.ListAgentsRequest) (*agentv1.ListAgentsResponse, error) {
	// Build status filter
	var statusFilter *types.AgentStatus
	if req.Status != agentv1.Agent_STATUS_UNSPECIFIED {
		agentStatus := convertProtoStatus(req.Status)
		statusFilter = &agentStatus
	}

	// List agents - Registry.ListAgents returns ([]*types.Agent, error)
	agents, err := s.registry.ListAgents(ctx, statusFilter)
	if err != nil {
		s.logger.Errorw("Failed to list agents",
			"error", err,
		)
		return nil, status.Errorf(codes.Internal, "failed to list agents: %v", err)
	}

	// Apply cluster filter if specified
	if req.ClusterId != "" {
		filteredAgents := make([]*types.Agent, 0)
		for _, a := range agents {
			if a.ClusterID == req.ClusterId {
				filteredAgents = append(filteredAgents, a)
			}
		}
		agents = filteredAgents
	}

	// Get pagination parameters
	page := int(1)
	pageSize := int(50)
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = int(req.Pagination.Page)
		}
		if req.Pagination.PageSize > 0 {
			pageSize = int(req.Pagination.PageSize)
		}
	}

	// Apply pagination
	total := int64(len(agents))
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(agents) {
		agents = []*types.Agent{}
	} else if end > len(agents) {
		agents = agents[start:]
	} else {
		agents = agents[start:end]
	}

	// Convert to protobuf
	pbAgents := make([]*agentv1.Agent, 0, len(agents))
	for _, a := range agents {
		pbAgents = append(pbAgents, convertAgentToProto(a))
	}

	// Build pagination metadata with safe conversion
	var totalPages int32
	if total > 0 {
		pages := (total + int64(pageSize) - 1) / int64(pageSize)
		if pages > math.MaxInt32 {
			totalPages = math.MaxInt32
		} else {
			// #nosec G115 -- overflow checked above
			totalPages = int32(pages)
		}
	} else {
		totalPages = 1
	}

	var pageInt32 int32
	if int64(page) > math.MaxInt32 {
		pageInt32 = math.MaxInt32
	} else {
		// #nosec G115 -- overflow checked above
		pageInt32 = int32(page)
	}

	hasNext := pageInt32 < totalPages
	hasPrev := page > 1

	var pageSizeInt32 int32
	if int64(pageSize) > math.MaxInt32 {
		pageSizeInt32 = math.MaxInt32
	} else {
		// #nosec G115 -- overflow checked above
		pageSizeInt32 = int32(pageSize)
	}

	pagination := &paginationv1.PaginationMetadata{
		Page:       pageInt32,
		PageSize:   pageSizeInt32,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}

	return &agentv1.ListAgentsResponse{
		Agents:     pbAgents,
		Pagination: pagination,
	}, nil
}

// UnregisterAgent unregisters an agent.
func (s *AgentServiceServer) UnregisterAgent(ctx context.Context, req *agentv1.UnregisterAgentRequest) (*agentv1.UnregisterAgentResponse, error) {
	if req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "agent_id is required")
	}

	// Unregister agent
	if err := s.registry.UnregisterAgent(ctx, req.AgentId); err != nil {
		s.logger.Errorw("Failed to unregister agent",
			"error", err,
			"agent_id", req.AgentId,
		)
		return nil, status.Errorf(codes.Internal, "failed to unregister agent: %v", err)
	}

	s.logger.Infow("Agent unregistered successfully",
		"agent_id", req.AgentId,
	)

	return &agentv1.UnregisterAgentResponse{
		Success: true,
	}, nil
}

// Helper functions

// convertAgentToProto converts types.Agent to protobuf Agent.
func convertAgentToProto(a *types.Agent) *agentv1.Agent {
	if a == nil {
		return nil
	}

	// Convert metadata to map[string]string
	metadata := make(map[string]string)
	for k, v := range a.Metadata {
		if str, ok := v.(string); ok {
			metadata[k] = str
		}
	}

	return &agentv1.Agent{
		Id:            a.ID,
		Name:          a.ID, // Using ID as name for now
		ClusterId:     a.ClusterID,
		ClusterName:   a.ClusterName,
		Version:       a.Version,
		Status:        convertStatusToProto(a.Status),
		LastHeartbeat: timestamppb.New(a.LastHeartbeat),
		RegisteredAt:  timestamppb.New(a.RegisteredAt),
		Metadata:      metadata,
	}
}

// convertStatusToProto converts types.AgentStatus to protobuf Status.
func convertStatusToProto(status types.AgentStatus) agentv1.Agent_Status {
	switch status {
	case types.AgentStatusOnline:
		return agentv1.Agent_ONLINE
	case types.AgentStatusOffline:
		return agentv1.Agent_OFFLINE
	case types.AgentStatusError:
		return agentv1.Agent_UNHEALTHY
	case types.AgentStatusRegistering:
		return agentv1.Agent_INITIALIZING
	default:
		return agentv1.Agent_STATUS_UNSPECIFIED
	}
}

// convertProtoStatus converts protobuf Status to types.AgentStatus.
func convertProtoStatus(status agentv1.Agent_Status) types.AgentStatus {
	switch status {
	case agentv1.Agent_ONLINE:
		return types.AgentStatusOnline
	case agentv1.Agent_OFFLINE:
		return types.AgentStatusOffline
	case agentv1.Agent_UNHEALTHY:
		return types.AgentStatusError
	case agentv1.Agent_INITIALIZING:
		return types.AgentStatusRegistering
	default:
		return types.AgentStatusOffline
	}
}

// convertMetadata converts map[string]string to map[string]interface{}.
func convertMetadata(metadata map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range metadata {
		result[k] = v
	}
	return result
}

// generateAgentToken generates a simple token for the agent
// In production, this should use proper JWT with signing.
func generateAgentToken(agentID string) string {
	// Simplified token generation - should use JWT in production
	return fmt.Sprintf("agent-token-%s-%d", agentID, time.Now().Unix())
}
