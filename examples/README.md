# Examples - Proto API 使用示例

本目录包含了如何使用生成的 Protocol Buffer API 的完整示例。

## 目录结构

```
examples/
├── grpc-server/       # gRPC 服务器示例
├── grpc-client/       # gRPC 客户端示例
└── http-gateway/      # HTTP Gateway 示例（gRPC + HTTP）
```

## 示例说明

### 1. gRPC Server（纯 gRPC 服务器）

**位置**: `examples/grpc-server/`

**功能**:
- 实现完整的 Agent Service gRPC API
- 包括 RegisterAgent, Heartbeat, GetAgent, ListAgents, UnregisterAgent
- 使用内存存储（生产环境应使用数据库）

**运行**:
```bash
# 启动服务器
go run examples/grpc-server/main.go

# 输出:
# gRPC server listening on port 50051
# Try: grpcurl -plaintext localhost:50051 list
```

**测试**:
```bash
# 使用 grpcurl 测试（需要安装 grpcurl）
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext localhost:50051 agent.v1.AgentService/ListAgents
```

### 2. gRPC Client（gRPC 客户端）

**位置**: `examples/grpc-client/`

**功能**:
- 演示如何使用 gRPC 客户端调用 Agent Service
- 完整的 API 调用流程示例
- 包含所有 5 个 RPC 方法调用

**运行**:
```bash
# 确保 gRPC 服务器正在运行
go run examples/grpc-server/main.go &

# 运行客户端
go run examples/grpc-client/main.go

# 使用自定义参数
go run examples/grpc-client/main.go \
  -addr localhost:50051 \
  -name my-agent \
  -cluster prod-cluster
```

**输出示例**:
```
=== Agent Service API Demo ===

1. Registering agent...
✓ Agent registered: id=agent-1698765432, token=token-agent-1698765432

2. Sending heartbeat...
✓ Heartbeat sent: next_interval=30 seconds

3. Getting agent info...
✓ Agent info: name=test-agent, cluster=production-cluster, version=v1.0.0

4. Listing all agents...
✓ Found 1 agents:
  1. test-agent (cluster: production-cluster)

5. Running for 3 seconds...

6. Unregistering agent...
✓ Agent unregistered: success=true

=== Demo Complete ===
```

### 3. HTTP Gateway（gRPC + HTTP/JSON）

**位置**: `examples/http-gateway/`

**功能**:
- 同时提供 gRPC 和 HTTP/JSON API
- 使用 gRPC-Gateway 自动转换
- 包含 CORS 支持
- 提供 Swagger 文档
- 健康检查端点
- API 文档页面

**运行**:
```bash
go run examples/http-gateway/main.go

# 输出:
# gRPC server listening on :50051
# HTTP Gateway listening on :8080
#   Health check: http://localhost:8080/health
#   Documentation: http://localhost:8080/docs
#   Swagger: http://localhost:8080/swagger/api.swagger.json
```

**测试 HTTP API**:

```bash
# 1. 健康检查
curl http://localhost:8080/health

# 2. 注册 Agent
curl -X POST http://localhost:8080/agent.v1.AgentService/RegisterAgent \
  -H "Content-Type: application/json" \
  -d '{
    "name": "http-agent",
    "cluster_id": "cluster-1",
    "cluster_name": "production",
    "version": "v1.0.0",
    "metadata": {
      "region": "us-west-1"
    }
  }'

# 3. 列出所有 Agent
curl -X POST http://localhost:8080/agent.v1.AgentService/ListAgents \
  -H "Content-Type: application/json" \
  -d '{}'

# 4. 获取特定 Agent
curl -X POST http://localhost:8080/agent.v1.AgentService/GetAgent \
  -H "Content-Type: application/json" \
  -d '{"agent_id": "agent-xxx"}'

# 5. 心跳
curl -X POST http://localhost:8080/agent.v1.AgentService/Heartbeat \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent-xxx",
    "status": 0,
    "metrics": {
      "cpu": "45%",
      "memory": "60%"
    }
  }'

# 6. 注销 Agent
curl -X POST http://localhost:8080/agent.v1.AgentService/UnregisterAgent \
  -H "Content-Type: application/json" \
  -d '{"agent_id": "agent-xxx"}'
```

**访问文档**:
- API 文档: http://localhost:8080/docs
- Swagger JSON: http://localhost:8080/swagger/api.swagger.json

## 快速开始

### 场景 1: 快速测试 gRPC API

```bash
# 终端 1: 启动服务器
go run examples/grpc-server/main.go

# 终端 2: 运行客户端
go run examples/grpc-client/main.go
```

### 场景 2: HTTP/JSON API 开发

```bash
# 启动 HTTP Gateway
go run examples/http-gateway/main.go

# 使用 curl 或 Postman 测试 HTTP API
curl http://localhost:8080/health
curl -X POST http://localhost:8080/agent.v1.AgentService/ListAgents \
  -H "Content-Type: application/json" \
  -d '{}'
```

### 场景 3: 同时使用 gRPC 和 HTTP

```bash
# 启动 HTTP Gateway（同时提供 gRPC 和 HTTP）
go run examples/http-gateway/main.go

# 使用 gRPC 客户端
go run examples/grpc-client/main.go -addr localhost:50051

# 或使用 HTTP API
curl http://localhost:8080/health
```

## 依赖项

示例代码依赖以下包：

```bash
# gRPC
go get google.golang.org/grpc
go get google.golang.org/protobuf

# gRPC Gateway
go get github.com/grpc-ecosystem/grpc-gateway/v2/runtime

# 生成的 API
# (已包含在 pkg/api/ 目录中)
```

## 工具推荐

### grpcurl - gRPC 命令行工具

安装:
```bash
# macOS
brew install grpcurl

# Linux
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

使用:
```bash
# 列出所有服务
grpcurl -plaintext localhost:50051 list

# 列出服务的方法
grpcurl -plaintext localhost:50051 list agent.v1.AgentService

# 调用方法
grpcurl -plaintext -d '{"name":"test"}' \
  localhost:50051 \
  agent.v1.AgentService/RegisterAgent
```

### Postman - HTTP API 测试

1. 导入 Swagger 文档: http://localhost:8080/swagger/api.swagger.json
2. 创建请求集合
3. 测试所有 HTTP 端点

## 集成到现有服务

### 在 Agent Manager 中使用

```go
// internal/agent-manager/grpc/server.go
import agentv1 "github.com/kart-io/k8s-agent/pkg/api/agent/v1"

type AgentManagerServer struct {
    agentv1.UnimplementedAgentServiceServer
    // 你的依赖...
}

func (s *AgentManagerServer) RegisterAgent(ctx context.Context, req *agentv1.RegisterAgentRequest) (*agentv1.RegisterAgentResponse, error) {
    // 你的实现...
}
```

### 添加 HTTP Gateway

```go
// internal/agent-manager/api/server.go
import (
    agentv1 "github.com/kart-io/k8s-agent/pkg/api/agent/v1"
    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func SetupGateway(grpcAddr string) *http.Server {
    mux := runtime.NewServeMux()
    opts := []grpc.DialOption{grpc.WithInsecure()}

    agentv1.RegisterAgentServiceHandlerFromEndpoint(
        context.Background(),
        mux,
        grpcAddr,
        opts,
    )

    return &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }
}
```

## 故障排查

### 问题 1: "Failed to connect"

**原因**: gRPC 服务器未启动
**解决**: 确保先启动 gRPC 服务器

### 问题 2: "Port already in use"

**原因**: 端口被占用
**解决**:
```bash
# 查找占用端口的进程
lsof -i :50051
lsof -i :8080

# 杀死进程
kill -9 <PID>
```

### 问题 3: 代码生成问题

**原因**: Proto 代码未生成或过时
**解决**:
```bash
# 重新生成代码
make proto.clean
make proto.generate
```

## 下一步

1. 查看 `pkg/api/` 目录了解生成的代码结构
2. 阅读 `api/proto/README.md` 了解 API 定义
3. 参考 `docs/PROTO_IMPLEMENTATION.md` 了解实施细节
4. 在实际服务中实现 gRPC handlers

## 参考资料

- [gRPC Go 快速开始](https://grpc.io/docs/languages/go/quickstart/)
- [gRPC-Gateway 文档](https://grpc-ecosystem.github.io/grpc-gateway/)
- [Protocol Buffers Go 教程](https://protobuf.dev/getting-started/gotutorial/)
