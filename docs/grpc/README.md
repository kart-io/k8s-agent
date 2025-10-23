# Agent Manager gRPC 服务

本文档描述如何使用 agent-manager 的 gRPC 服务。

## 概述

agent-manager 现在同时提供 HTTP REST API 和 gRPC 服务:

- **HTTP API**: 端口 `8080` (默认)
- **gRPC API**: 端口 `9090` (默认)

## gRPC 服务列表

### 1. AgentService

Agent 生命周期管理服务。

**可用方法:**

- `RegisterAgent` - 注册新的 Agent
- `Heartbeat` - 发送心跳信号
- `GetAgent` - 获取 Agent 信息
- `ListAgents` - 列出所有 Agents (支持分页和过滤)
- `UnregisterAgent` - 注销 Agent

### 2. CommandService

命令执行和管理服务。

**可用方法:**

- `ExecuteCommand` - 执行命令
- `GetCommandStatus` - 获取命令状态
- `CancelCommand` - 取消命令执行

## 配置

在 `configs/agent-manager.yaml` 中配置 gRPC:

```yaml
grpc:
  # 启用 gRPC 服务器
  enable: true
  # 监听地址
  host: "0.0.0.0"
  # 监听端口
  port: 9090
  # 最大接收消息大小（字节）
  max_recv_msg_size: 10485760  # 10MB
  # 最大发送消息大小（字节）
  max_send_msg_size: 10485760  # 10MB
  # KeepAlive 时间间隔
  keep_alive_time: 30s
  # KeepAlive 超时
  keep_alive_timeout: 10s
  # 启用 gRPC 反射服务（用于开发调试）
  enable_reflection: true
  # 启用 gRPC 健康检查服务
  enable_health_check: true
```

## 使用示例

### 使用 grpcurl 测试

安装 [grpcurl](https://github.com/fullstorydev/grpcurl):

```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

列出所有服务:

```bash
grpcurl -plaintext localhost:9090 list
```

列出 AgentService 的方法:

```bash
grpcurl -plaintext localhost:9090 list agent.v1.AgentService
```

注册 Agent:

```bash
grpcurl -plaintext -d '{
  "name": "test-agent",
  "cluster_id": "cluster-001",
  "cluster_name": "test-cluster",
  "version": "v1.0.0",
  "metadata": {
    "region": "us-west-2",
    "env": "dev"
  }
}' localhost:9090 agent.v1.AgentService/RegisterAgent
```

列出所有 Agents:

```bash
grpcurl -plaintext -d '{
  "pagination": {
    "page": 1,
    "page_size": 10
  }
}' localhost:9090 agent.v1.AgentService/ListAgents
```

### 使用 Go 客户端

参见 `examples/grpc-client/main.go` 的完整示例。

运行示例:

```bash
# 启动 agent-manager 服务
cd cmd/agent-manager
go run main.go -c ../../configs/agent-manager.yaml

# 在另一个终端运行客户端示例
cd examples/grpc-client
go run main.go -addr localhost:9090
```

### 使用其他语言的客户端

从 `pkg/api/agent/v1/*.proto` 文件生成客户端代码:

**Python:**

```bash
python -m grpc_tools.protoc \
  -I./pkg/api \
  --python_out=. \
  --grpc_python_out=. \
  pkg/api/agent/v1/*.proto
```

**Java:**

```bash
protoc -I./pkg/api \
  --java_out=./java/src \
  --grpc-java_out=./java/src \
  pkg/api/agent/v1/*.proto
```

**Node.js:**

```bash
grpc_tools_node_protoc \
  -I./pkg/api \
  --js_out=import_style=commonjs,binary:./js \
  --grpc_out=grpc_js:./js \
  pkg/api/agent/v1/*.proto
```

## API 文档

完整的 API 定义参见:

- Agent Service: `pkg/api/agent/v1/agent.proto`
- Command Service: `pkg/api/agent/v1/command.proto`

生成的 Go 代码在:

- `pkg/api/agent/v1/*.pb.go` - Protobuf 消息
- `pkg/api/agent/v1/*_grpc.pb.go` - gRPC 服务

## 开发

### 修改 Proto 定义

1. 编辑 `.proto` 文件:
   ```bash
   vim pkg/api/agent/v1/agent.proto
   ```

2. 重新生成代码:
   ```bash
   make gen-proto
   ```

3. 重新编译:
   ```bash
   make build
   ```

### 实现细节

gRPC 服务实现位于:

- `internal/agent-manager/grpc/agent_service.go` - AgentService 实现
- `internal/agent-manager/grpc/command_service.go` - CommandService 实现
- `internal/agent-manager/grpc/server.go` - gRPC 服务器

### 测试

运行测试:

```bash
# 测试 gRPC 服务
go test ./internal/agent-manager/grpc/...

# 测试整个 agent-manager
go test ./internal/agent-manager/...
```

## 性能考虑

- gRPC 使用 HTTP/2,支持多路复用和流式传输
- 默认消息大小限制为 10MB,可通过配置调整
- KeepAlive 机制确保长连接稳定性
- 支持并发请求处理

## 安全性

当前实现为开发环境配置,生产环境需要:

1. **启用 TLS**:
   ```go
   creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
   grpcServer := grpc.NewServer(grpc.Creds(creds))
   ```

2. **添加认证**:
   - JWT Token 验证
   - mTLS 客户端证书

3. **限流**:
   - 添加 rate limiting 中间件
   - 使用 grpc.UnaryInterceptor

## 故障排查

### gRPC 服务无法启动

检查端口是否被占用:

```bash
lsof -i :9090
```

### 连接超时

检查防火墙设置和网络连接:

```bash
telnet localhost 9090
```

### 查看详细日志

启用 gRPC 调试日志:

```bash
export GRPC_GO_LOG_VERBOSITY_LEVEL=99
export GRPC_GO_LOG_SEVERITY_LEVEL=info
```

## 参考资料

- [gRPC Go 文档](https://grpc.io/docs/languages/go/)
- [Protocol Buffers 文档](https://protobuf.dev/)
- [grpcurl 工具](https://github.com/fullstorydev/grpcurl)
