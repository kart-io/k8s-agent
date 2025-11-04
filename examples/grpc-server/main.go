// examples/grpc-server/main.go
// gRPC 服务器示例 - Agent Service 实现

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	agentv1 "github.com/kart-io/k8s-agent/pkg/api/agent/v1"
)

// AgentServer 实现 Agent Service.
type AgentServer struct {
	agentv1.UnimplementedAgentServiceServer
	agents map[string]*agentv1.Agent // 简单的内存存储
}

// NewAgentServer 创建新的 Agent 服务器.
func NewAgentServer() *AgentServer {
	return &AgentServer{
		agents: make(map[string]*agentv1.Agent),
	}
}

// RegisterAgent 实现注册 Agent RPC.
func (s *AgentServer) RegisterAgent(ctx context.Context, req *agentv1.RegisterAgentRequest) (*agentv1.RegisterAgentResponse, error) {
	log.Printf("Registering agent: name=%s, cluster=%s", req.Name, req.ClusterName)

	// 生成 Agent ID
	agentID := fmt.Sprintf("agent-%d", time.Now().Unix())

	// 创建 Agent
	agent := &agentv1.Agent{
		Id:            agentID,
		Name:          req.Name,
		ClusterId:     req.ClusterId,
		ClusterName:   req.ClusterName,
		Version:       req.Version,
		Status:        agentv1.Agent_STATUS_UNSPECIFIED, // 使用生成的枚举
		LastHeartbeat: timestamppb.Now(),
		RegisteredAt:  timestamppb.Now(),
		Metadata:      req.Metadata,
	}

	// 保存到内存
	s.agents[agentID] = agent

	// 生成认证 Token
	token := fmt.Sprintf("token-%s", agentID)

	log.Printf("Agent registered successfully: id=%s", agentID)

	return &agentv1.RegisterAgentResponse{
		Agent: agent,
		Token: token,
	}, nil
}

// Heartbeat 实现心跳 RPC.
func (s *AgentServer) Heartbeat(ctx context.Context, req *agentv1.HeartbeatRequest) (*agentv1.HeartbeatResponse, error) {
	log.Printf("Heartbeat from agent: id=%s, status=%v", req.AgentId, req.Status)

	// 检查 Agent 是否存在
	agent, exists := s.agents[req.AgentId]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "agent not found: %s", req.AgentId)
	}

	// 更新心跳时间和状态
	agent.LastHeartbeat = timestamppb.Now()
	agent.Status = req.Status

	return &agentv1.HeartbeatResponse{
		Success:               true,
		NextHeartbeatInterval: 30, // 30 秒
	}, nil
}

// GetAgent 实现获取 Agent RPC.
func (s *AgentServer) GetAgent(ctx context.Context, req *agentv1.GetAgentRequest) (*agentv1.GetAgentResponse, error) {
	log.Printf("Getting agent: id=%s", req.AgentId)

	agent, exists := s.agents[req.AgentId]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "agent not found: %s", req.AgentId)
	}

	return &agentv1.GetAgentResponse{
		Agent: agent,
	}, nil
}

// ListAgents 实现列出 Agent RPC.
func (s *AgentServer) ListAgents(ctx context.Context, req *agentv1.ListAgentsRequest) (*agentv1.ListAgentsResponse, error) {
	log.Printf("Listing agents: cluster=%s", req.ClusterId)

	// 收集所有 Agent
	var agents []*agentv1.Agent
	for _, agent := range s.agents {
		// 如果指定了 cluster_id，进行过滤
		if req.ClusterId != "" && agent.ClusterId != req.ClusterId {
			continue
		}
		// 如果指定了 status，进行过滤
		if req.Status != agentv1.Agent_STATUS_UNSPECIFIED && agent.Status != req.Status {
			continue
		}
		agents = append(agents, agent)
	}

	// TODO: 实现真正的分页逻辑
	return &agentv1.ListAgentsResponse{
		Agents: agents,
		// Pagination: ..., // 实际应用中需要实现分页
	}, nil
}

// UnregisterAgent 实现注销 Agent RPC.
func (s *AgentServer) UnregisterAgent(ctx context.Context, req *agentv1.UnregisterAgentRequest) (*agentv1.UnregisterAgentResponse, error) {
	log.Printf("Unregistering agent: id=%s", req.AgentId)

	if _, exists := s.agents[req.AgentId]; !exists {
		return nil, status.Errorf(codes.NotFound, "agent not found: %s", req.AgentId)
	}

	delete(s.agents, req.AgentId)

	log.Printf("Agent unregistered successfully: id=%s", req.AgentId)

	return &agentv1.UnregisterAgentResponse{
		Success: true,
	}, nil
}

func main() {
	// 监听端口
	port := 50051
	lc := net.ListenConfig{}
	lis, err := lc.Listen(context.Background(), "tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// 创建 gRPC 服务器
	grpcServer := grpc.NewServer()

	// 注册服务
	agentServer := NewAgentServer()
	agentv1.RegisterAgentServiceServer(grpcServer, agentServer)

	log.Printf("gRPC server listening on port %d", port)
	log.Printf("Try: grpcurl -plaintext localhost:%d list", port)

	// 启动服务器
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
