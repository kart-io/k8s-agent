package services

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/kart-io/logger/core"

	agentv1 "github.com/kart-io/k8s-agent/api/proto/gen/agentmanager/agent/v1"
	"github.com/kart-io/k8s-agent/pkg/types"
)

// AgentServiceServer Agent 服务实现
type AgentServiceServer struct {
	agentv1.UnimplementedAgentServiceServer
	logger   core.Logger
	registry AgentRegistry
	store    AgentStore
}

// AgentRegistry 定义 Agent 注册表接口
type AgentRegistry interface {
	RegisterAgent(ctx context.Context, agent *types.Agent) error
	UnregisterAgent(ctx context.Context, agentID string) error
	UpdateHeartbeat(ctx context.Context, agentID string) error
	GetAgent(ctx context.Context, agentID string) (*types.Agent, error)
	ListAgents(ctx context.Context, status *types.AgentStatus) ([]*types.Agent, error)
}

// AgentStore 定义 Agent 存储接口
type AgentStore interface {
	GetAgentMetrics(ctx context.Context, agentID string, startTime, endTime *time.Time) ([]*types.Metrics, error)
}

// NewAgentServiceServer 创建 Agent 服务
func NewAgentServiceServer(logger core.Logger, registry AgentRegistry, store AgentStore) *AgentServiceServer {
	return &AgentServiceServer{
		logger:   logger.With("service", "agent"),
		registry: registry,
		store:    store,
	}
}

// ListAgents 列出所有 Agents
func (s *AgentServiceServer) ListAgents(ctx context.Context, req *agentv1.ListAgentsRequest) (*agentv1.ListAgentsResponse, error) {
	s.logger.Infow("ListAgents called",
		"cluster_id", req.ClusterId,
		"status", req.Status.String(),
	)

	// 转换状态过滤器
	var statusFilter *types.AgentStatus
	if req.Status != agentv1.AgentStatus_AGENT_STATUS_UNSPECIFIED {
		status := convertProtoStatusToTypes(req.Status)
		statusFilter = &status
	}

	// 从注册表获取 agents
	agents, err := s.registry.ListAgents(ctx, statusFilter)
	if err != nil {
		s.logger.Errorw("Failed to list agents", "error", err)
		return nil, status.Error(codes.Internal, "failed to list agents")
	}

	// 如果指定了 cluster_id，过滤结果
	if req.ClusterId != "" {
		filtered := make([]*types.Agent, 0)
		for _, agent := range agents {
			if agent.ClusterID == req.ClusterId {
				filtered = append(filtered, agent)
			}
		}
		agents = filtered
	}

	// 转换为 proto 类型
	protoAgents := make([]*agentv1.Agent, 0, len(agents))
	for _, agent := range agents {
		protoAgent := convertTypesToProtoAgent(agent)
		protoAgents = append(protoAgents, protoAgent)
	}

	return &agentv1.ListAgentsResponse{
		Agents:        protoAgents,
		NextPageToken: "", // TODO: 实现分页
		TotalCount:    int32(len(protoAgents)),
	}, nil
}

// GetAgent 获取 Agent 详情
func (s *AgentServiceServer) GetAgent(ctx context.Context, req *agentv1.GetAgentRequest) (*agentv1.Agent, error) {
	s.logger.Infow("GetAgent called", "agent_id", req.AgentId)

	if req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "agent_id is required")
	}

	// 从注册表获取 agent
	agent, err := s.registry.GetAgent(ctx, req.AgentId)
	if err != nil {
		s.logger.Errorw("Failed to get agent", "agent_id", req.AgentId, "error", err)
		return nil, status.Error(codes.NotFound, "agent not found")
	}

	return convertTypesToProtoAgent(agent), nil
}

// RegisterAgent Agent 注册
func (s *AgentServiceServer) RegisterAgent(ctx context.Context, req *agentv1.RegisterAgentRequest) (*agentv1.Agent, error) {
	s.logger.Infow("RegisterAgent called",
		"agent_id", req.AgentId,
		"cluster_id", req.ClusterId,
	)

	if req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "agent_id is required")
	}
	if req.ClusterId == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster_id is required")
	}

	// 创建 agent 对象
	now := time.Now()
	agent := &types.Agent{
		ID:            req.AgentId,
		ClusterID:     req.ClusterId,
		Version:       req.Version,
		Status:        types.AgentStatusOnline,
		LastHeartbeat: now,
		RegisteredAt:  now,
		UpdatedAt:     now,
		Metadata:      make(map[string]interface{}),
		Capabilities:  []string{},
	}

	// 注册到注册表
	if err := s.registry.RegisterAgent(ctx, agent); err != nil {
		s.logger.Errorw("Failed to register agent", "agent_id", req.AgentId, "error", err)
		return nil, status.Error(codes.Internal, "failed to register agent")
	}

	s.logger.Infow("Agent registered successfully", "agent_id", req.AgentId)
	return convertTypesToProtoAgent(agent), nil
}

// UpdateAgentHeartbeat 更新 Agent 心跳
func (s *AgentServiceServer) UpdateAgentHeartbeat(ctx context.Context, req *agentv1.UpdateAgentHeartbeatRequest) (*emptypb.Empty, error) {
	s.logger.Debugw("UpdateAgentHeartbeat called", "agent_id", req.AgentId)

	if req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "agent_id is required")
	}

	// 更新心跳
	if err := s.registry.UpdateHeartbeat(ctx, req.AgentId); err != nil {
		s.logger.Errorw("Failed to update heartbeat", "agent_id", req.AgentId, "error", err)
		return nil, status.Error(codes.Internal, "failed to update heartbeat")
	}

	return &emptypb.Empty{}, nil
}

// DeregisterAgent Agent 注销
func (s *AgentServiceServer) DeregisterAgent(ctx context.Context, req *agentv1.DeregisterAgentRequest) (*emptypb.Empty, error) {
	s.logger.Infow("DeregisterAgent called", "agent_id", req.AgentId)

	if req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "agent_id is required")
	}

	// 从注册表注销
	if err := s.registry.UnregisterAgent(ctx, req.AgentId); err != nil {
		s.logger.Errorw("Failed to deregister agent", "agent_id", req.AgentId, "error", err)
		return nil, status.Error(codes.Internal, "failed to deregister agent")
	}

	s.logger.Infow("Agent deregistered successfully", "agent_id", req.AgentId)
	return &emptypb.Empty{}, nil
}

// GetAgentMetrics 获取 Agent 指标
func (s *AgentServiceServer) GetAgentMetrics(ctx context.Context, req *agentv1.GetAgentMetricsRequest) (*agentv1.AgentMetrics, error) {
	s.logger.Infow("GetAgentMetrics called", "agent_id", req.AgentId)

	if req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "agent_id is required")
	}

	// TODO: 从存储层获取指标数据
	return nil, status.Error(codes.NotFound, "metrics not found")
}

// Type conversion functions

// convertTypesToProtoAgent converts types.Agent to agentv1.Agent
func convertTypesToProtoAgent(agent *types.Agent) *agentv1.Agent {
	if agent == nil {
		return nil
	}

	protoAgent := &agentv1.Agent{
		AgentId:   agent.ID,
		ClusterId: agent.ClusterID,
		Version:   agent.Version,
		Status:    convertTypesStatusToProto(agent.Status),
	}

	if !agent.RegisteredAt.IsZero() {
		protoAgent.RegisteredAt = timestamppb.New(agent.RegisteredAt)
	}
	if !agent.LastHeartbeat.IsZero() {
		protoAgent.LastHeartbeat = timestamppb.New(agent.LastHeartbeat)
	}

	return protoAgent
}

// convertProtoStatusToTypes converts agentv1.AgentStatus to types.AgentStatus
func convertProtoStatusToTypes(status agentv1.AgentStatus) types.AgentStatus {
	switch status {
	case agentv1.AgentStatus_AGENT_STATUS_ONLINE:
		return types.AgentStatusOnline
	case agentv1.AgentStatus_AGENT_STATUS_OFFLINE:
		return types.AgentStatusOffline
	case agentv1.AgentStatus_AGENT_STATUS_ERROR:
		return types.AgentStatusError
	default:
		return types.AgentStatusOffline
	}
}

// convertTypesStatusToProto converts types.AgentStatus to agentv1.AgentStatus
func convertTypesStatusToProto(status types.AgentStatus) agentv1.AgentStatus {
	switch status {
	case types.AgentStatusOnline:
		return agentv1.AgentStatus_AGENT_STATUS_ONLINE
	case types.AgentStatusOffline:
		return agentv1.AgentStatus_AGENT_STATUS_OFFLINE
	case types.AgentStatusError:
		return agentv1.AgentStatus_AGENT_STATUS_ERROR
	case types.AgentStatusRegistering:
		return agentv1.AgentStatus_AGENT_STATUS_ONLINE // Map registering to online
	default:
		return agentv1.AgentStatus_AGENT_STATUS_UNSPECIFIED
	}
}

