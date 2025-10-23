// examples/http-gateway/main.go
// HTTP Gateway 示例 - 同时提供 gRPC 和 HTTP/JSON API

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	agentv1 "github.com/kart-io/k8s-agent/pkg/api/agent/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	grpcPort = 50051
	httpPort = 8080
)

// AgentServer 实现 Agent Service
type AgentServer struct {
	agentv1.UnimplementedAgentServiceServer
	agents map[string]*agentv1.Agent
}

func NewAgentServer() *AgentServer {
	return &AgentServer{
		agents: make(map[string]*agentv1.Agent),
	}
}

// 实现所有 RPC 方法（简化版）
func (s *AgentServer) RegisterAgent(ctx context.Context, req *agentv1.RegisterAgentRequest) (*agentv1.RegisterAgentResponse, error) {
	agentID := fmt.Sprintf("agent-%d", time.Now().Unix())
	agent := &agentv1.Agent{
		Id:            agentID,
		Name:          req.Name,
		ClusterId:     req.ClusterId,
		ClusterName:   req.ClusterName,
		Version:       req.Version,
		Status:        agentv1.Agent_STATUS_UNSPECIFIED,
		LastHeartbeat: timestamppb.Now(),
		RegisteredAt:  timestamppb.Now(),
		Metadata:      req.Metadata,
	}
	s.agents[agentID] = agent
	return &agentv1.RegisterAgentResponse{
		Agent: agent,
		Token: fmt.Sprintf("token-%s", agentID),
	}, nil
}

func (s *AgentServer) ListAgents(ctx context.Context, req *agentv1.ListAgentsRequest) (*agentv1.ListAgentsResponse, error) {
	var agents []*agentv1.Agent
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}
	return &agentv1.ListAgentsResponse{Agents: agents}, nil
}

func (s *AgentServer) GetAgent(ctx context.Context, req *agentv1.GetAgentRequest) (*agentv1.GetAgentResponse, error) {
	agent := s.agents[req.AgentId]
	return &agentv1.GetAgentResponse{Agent: agent}, nil
}

func (s *AgentServer) Heartbeat(ctx context.Context, req *agentv1.HeartbeatRequest) (*agentv1.HeartbeatResponse, error) {
	return &agentv1.HeartbeatResponse{Success: true, NextHeartbeatInterval: 30}, nil
}

func (s *AgentServer) UnregisterAgent(ctx context.Context, req *agentv1.UnregisterAgentRequest) (*agentv1.UnregisterAgentResponse, error) {
	delete(s.agents, req.AgentId)
	return &agentv1.UnregisterAgentResponse{Success: true}, nil
}

// startGRPCServer 启动 gRPC 服务器
func startGRPCServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	agentServer := NewAgentServer()
	agentv1.RegisterAgentServiceServer(grpcServer, agentServer)

	log.Printf("gRPC server listening on :%d", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}

// startHTTPGateway 启动 HTTP Gateway
func startHTTPGateway() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 创建 gRPC-Gateway mux
	mux := runtime.NewServeMux()

	// 连接到 gRPC 服务器
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := agentv1.RegisterAgentServiceHandlerFromEndpoint(
		ctx,
		mux,
		fmt.Sprintf("localhost:%d", grpcPort),
		opts,
	)
	if err != nil {
		log.Fatalf("Failed to register gateway: %v", err)
	}

	// 添加自定义路由
	httpMux := http.NewServeMux()
	httpMux.Handle("/", mux)

	// 添加健康检查端点
	httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","service":"agent-gateway"}`)
	})

	// 添加 API 文档端点
	httpMux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Agent API Documentation</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        code { background: #f4f4f4; padding: 2px 6px; border-radius: 3px; }
        pre { background: #f4f4f4; padding: 15px; border-radius: 5px; overflow-x: auto; }
    </style>
</head>
<body>
    <h1>Agent Service API Documentation</h1>

    <h2>Endpoints</h2>

    <h3>1. Register Agent</h3>
    <p><code>POST /agent.v1.AgentService/RegisterAgent</code></p>
    <pre>curl -X POST http://localhost:%d/agent.v1.AgentService/RegisterAgent \
  -H "Content-Type: application/json" \
  -d '{
    "name": "agent-1",
    "cluster_id": "cluster-1",
    "cluster_name": "production",
    "version": "v1.0.0"
  }'</pre>

    <h3>2. List Agents</h3>
    <p><code>POST /agent.v1.AgentService/ListAgents</code></p>
    <pre>curl -X POST http://localhost:%d/agent.v1.AgentService/ListAgents \
  -H "Content-Type: application/json" \
  -d '{}'</pre>

    <h3>3. Get Agent</h3>
    <p><code>POST /agent.v1.AgentService/GetAgent</code></p>
    <pre>curl -X POST http://localhost:%d/agent.v1.AgentService/GetAgent \
  -H "Content-Type: application/json" \
  -d '{"agent_id": "agent-xxx"}'</pre>

    <h2>Health Check</h2>
    <p><code>GET /health</code></p>
    <pre>curl http://localhost:%d/health</pre>

    <h2>More Information</h2>
    <ul>
        <li>Swagger Doc: <a href="/swagger/api.swagger.json">/swagger/api.swagger.json</a></li>
        <li>gRPC Port: %d</li>
        <li>HTTP Port: %d</li>
    </ul>
</body>
</html>
`, httpPort, httpPort, httpPort, httpPort, grpcPort, httpPort)
	})

	// 提供 Swagger 文档
	httpMux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("pkg/api/docs/swagger"))))

	// 启动 HTTP 服务器
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: corsMiddleware(httpMux),
	}

	log.Printf("HTTP Gateway listening on :%d", httpPort)
	log.Printf("  Health check: http://localhost:%d/health", httpPort)
	log.Printf("  Documentation: http://localhost:%d/docs", httpPort)
	log.Printf("  Swagger: http://localhost:%d/swagger/api.swagger.json", httpPort)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to serve HTTP: %v", err)
	}
}

// corsMiddleware 添加 CORS 支持
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	log.Println("Starting Agent Service with gRPC and HTTP Gateway...")

	// 启动 gRPC 服务器（goroutine）
	go startGRPCServer()

	// 等待 gRPC 服务器启动
	time.Sleep(1 * time.Second)

	// 启动 HTTP Gateway（主线程）
	startHTTPGateway()
}
