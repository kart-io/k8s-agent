// examples/grpc-client/main.go
// gRPC 客户端示例 - Agent Service 调用

package main

import (
	"context"
	"flag"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	agentv1 "github.com/kart-io/k8s-agent/pkg/api/agent/v1"
)

var (
	serverAddr = flag.String("addr", "localhost:50051", "gRPC server address")
	agentName  = flag.String("name", "test-agent", "agent name")
	clusterID  = flag.String("cluster", "cluster-1", "cluster ID")
)

func main() {
	flag.Parse()

	// 连接到 gRPC 服务器
	conn, err := grpc.NewClient(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close connection: %v", err)
		}
	}()

	// 创建客户端
	client := agentv1.NewAgentServiceClient(conn)

	// 演示所有 API 调用
	demonstrateAgentAPIs(client)
}

func demonstrateAgentAPIs(client agentv1.AgentServiceClient) {
	ctx := context.Background()

	log.Println("=== Agent Service API Demo ===")

	// 1. 注册 Agent
	log.Println("\n1. Registering agent...")
	registerResp, err := client.RegisterAgent(ctx, &agentv1.RegisterAgentRequest{
		Name:        *agentName,
		ClusterId:   *clusterID,
		ClusterName: "production-cluster",
		Version:     "v1.0.0",
		Metadata: map[string]string{
			"region": "us-west-1",
			"zone":   "us-west-1a",
		},
	})
	if err != nil {
		log.Fatalf("RegisterAgent failed: %v", err)
	}
	log.Printf("✓ Agent registered: id=%s, token=%s", registerResp.Agent.Id, registerResp.Token)

	agentID := registerResp.Agent.Id

	// 2. 发送心跳
	log.Println("\n2. Sending heartbeat...")
	heartbeatResp, err := client.Heartbeat(ctx, &agentv1.HeartbeatRequest{
		AgentId: agentID,
		Status:  agentv1.Agent_STATUS_UNSPECIFIED, // 使用生成的枚举
		Metrics: map[string]string{
			"cpu":    "45%",
			"memory": "60%",
		},
	})
	if err != nil {
		log.Fatalf("Heartbeat failed: %v", err)
	}
	log.Printf("✓ Heartbeat sent: next_interval=%d seconds", heartbeatResp.NextHeartbeatInterval)

	// 3. 获取 Agent 信息
	log.Println("\n3. Getting agent info...")
	getResp, err := client.GetAgent(ctx, &agentv1.GetAgentRequest{
		AgentId: agentID,
	})
	if err != nil {
		log.Fatalf("GetAgent failed: %v", err)
	}
	log.Printf("✓ Agent info: name=%s, cluster=%s, version=%s, status=%v",
		getResp.Agent.Name,
		getResp.Agent.ClusterName,
		getResp.Agent.Version,
		getResp.Agent.Status)

	// 4. 列出所有 Agent
	log.Println("\n4. Listing all agents...")
	listResp, err := client.ListAgents(ctx, &agentv1.ListAgentsRequest{
		// Pagination: &paginationv1.PaginationRequest{
		// 	Page:     1,
		// 	PageSize: 10,
		// },
	})
	if err != nil {
		log.Fatalf("ListAgents failed: %v", err)
	}
	log.Printf("✓ Found %d agents:", len(listResp.Agents))
	for i, agent := range listResp.Agents {
		log.Printf("  %d. %s (cluster: %s, status: %v)",
			i+1, agent.Name, agent.ClusterName, agent.Status)
	}

	// 5. 等待一会儿，模拟运行中
	log.Println("\n5. Running for 3 seconds...")
	time.Sleep(3 * time.Second)

	// 6. 注销 Agent
	log.Println("\n6. Unregistering agent...")
	unregisterResp, err := client.UnregisterAgent(ctx, &agentv1.UnregisterAgentRequest{
		AgentId: agentID,
	})
	if err != nil {
		log.Fatalf("UnregisterAgent failed: %v", err)
	}
	log.Printf("✓ Agent unregistered: success=%v", unregisterResp.Success)

	log.Println("\n=== Demo Complete ===")
}
