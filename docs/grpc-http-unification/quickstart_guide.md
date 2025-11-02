# 快速启动指南 - gRPC-HTTP统一Handler测试

## 概述

本指南帮助您快速启动和测试新实现的gRPC-HTTP统一handler功能。

---

## 准备工作

### 1. 构建服务

```bash
# 从项目根目录执行
cd /Users/costalong/code/go/src/github.com/kart/k8s-agent

# 构建所有服务
make build

# 或单独构建
make go.build.orchestrator
make go.build.reasoning
```

### 2. 启动依赖服务

```bash
# 启动PostgreSQL (Orchestrator需要)
docker run -d \
  --name k8s-agent-postgres \
  -e POSTGRES_USER=aetherius \
  -e POSTGRES_PASSWORD=aetherius \
  -e POSTGRES_DB=aetherius_orchestrator \
  -p 5432:5432 \
  postgres:14

# 启动Redis (可选，用于缓存)
docker run -d \
  --name k8s-agent-redis \
  -p 6379:6379 \
  redis:7-alpine

# 启动NATS (可选，用于消息)
docker run -d \
  --name k8s-agent-nats \
  -p 4222:4222 \
  nats:2.10-alpine
```

---

## 启动服务

### Orchestrator Service

```bash
# 方式1: 直接运行
_output/bin/orchestrator --config configs/orchestrator.yaml

# 方式2: 使用make
make run-orchestrator

# 服务端口:
# - HTTP: :8092
# - gRPC: :9092
# - Health: :8091
```

### Reasoning Service

```bash
# 方式1: 直接运行
_output/bin/reasoning --config configs/reasoning.yaml

# 方式2: 使用make
make run-reasoning

# 服务端口:
# - HTTP: :8082
# - gRPC: :9093
# - Health: :8081
```

---

## HTTP API 测试

### 1. 健康检查

```bash
# Orchestrator
curl http://localhost:8091/healthz

# Reasoning
curl http://localhost:8081/healthz
```

### 2. Orchestrator API测试

#### 创建工作流

```bash
curl -X POST http://localhost:8092/v1/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-workflow",
    "description": "测试工作流",
    "steps": [
      {
        "id": "step1",
        "name": "第一步",
        "type": "COMMAND",
        "config": {
          "command": "echo",
          "args": ["Hello World"]
        }
      }
    ]
  }' | jq
```

**期望响应**:
```json
{
  "workflow_id": "wf-xxx",
  "name": "test-workflow",
  "status": "CREATED",
  "created_at": "2025-11-01T12:00:00Z"
}
```

#### 获取工作流

```bash
# 替换为实际的workflow_id
curl http://localhost:8092/v1/workflows/wf-xxx | jq
```

#### 列出所有工作流

```bash
curl "http://localhost:8092/v1/workflows?page=1&page_size=10" | jq
```

#### 执行工作流

```bash
curl -X POST http://localhost:8092/v1/workflows/wf-xxx/executions \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "evt-test-001",
    "params": {
      "test": "value"
    }
  }' | jq
```

#### 查询执行状态

```bash
# 替换为实际的execution_id
curl http://localhost:8092/v1/executions/exec-xxx | jq
```

### 3. Reasoning API测试

#### 根因分析

```bash
curl -X POST http://localhost:8082/v1/analysis/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "evt-test-001",
    "context": {
      "events": [
        {
          "type": "CrashLoopBackOff",
          "message": "容器启动失败",
          "timestamp": "2025-11-01T12:00:00Z"
        }
      ],
      "logs": [
        "Error: Cannot find module",
        "at Function.Module._resolveFilename"
      ],
      "metrics": {
        "memory": {
          "usage_percent": 85.5,
          "usage_bytes": 1800000000,
          "limit_bytes": 2147483648
        },
        "cpu": {
          "usage_percent": 45.2
        }
      }
    },
    "options": {
      "use_knowledge_graph": true,
      "include_historical_cases": true,
      "model": "openai"
    }
  }' | jq
```

**期望响应**:
```json
{
  "analysis_id": "analysis-xxx",
  "root_cause": {
    "type": "CONFIGURATION_ERROR",
    "description": "缺少依赖模块",
    "confidence": 0.92,
    "evidences": [...]
  },
  "recommendations": [...],
  "similar_cases": [...],
  "confidence": 0.92,
  "duration_ms": 3500
}
```

#### 保存案例

```bash
curl -X POST http://localhost:8082/v1/cases \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "evt-test-001",
    "analysis_id": "analysis-xxx",
    "result": "CORRECT",
    "solution": "安装缺失的依赖模块",
    "feedback": "分析准确"
  }' | jq
```

---

## gRPC API 测试

### 使用grpcurl

```bash
# 安装grpcurl (如果未安装)
brew install grpcurl

# 查看可用服务
grpcurl -plaintext localhost:9092 list
grpcurl -plaintext localhost:9093 list

# 查看服务方法
grpcurl -plaintext localhost:9092 list orchestrator.v1.WorkflowService
grpcurl -plaintext localhost:9093 list reasoning.v1.ReasoningService

# 调用CreateWorkflow
grpcurl -plaintext -d '{
  "name": "test-workflow-grpc",
  "description": "通过gRPC创建的工作流",
  "steps": [
    {
      "id": "step1",
      "name": "测试步骤",
      "type": "COMMAND"
    }
  ]
}' localhost:9092 orchestrator.v1.WorkflowService/CreateWorkflow

# 调用GetWorkflow
grpcurl -plaintext -d '{
  "workflow_id": "wf-xxx"
}' localhost:9092 orchestrator.v1.WorkflowService/GetWorkflow

# 调用RootCauseAnalysis
grpcurl -plaintext -d '{
  "event_id": "evt-grpc-001",
  "context": {
    "events": [
      {
        "type": "CrashLoopBackOff",
        "message": "容器崩溃"
      }
    ],
    "logs": ["Error: Out of memory"]
  }
}' localhost:9093 reasoning.v1.ReasoningService/RootCauseAnalysis
```

---

## 验证统一Handler

### 验证HTTP和gRPC返回相同结果

**测试方法**:
1. 通过HTTP创建工作流，记录workflow_id
2. 通过gRPC获取相同workflow_id的工作流
3. 对比两次返回的数据

```bash
# 1. HTTP创建
WORKFLOW_ID=$(curl -s -X POST http://localhost:8092/v1/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-unified",
    "description": "验证统一handler"
  }' | jq -r '.workflow_id')

echo "Created workflow: $WORKFLOW_ID"

# 2. HTTP获取
echo "=== HTTP Response ==="
curl -s http://localhost:8092/v1/workflows/$WORKFLOW_ID | jq

# 3. gRPC获取
echo "=== gRPC Response ==="
grpcurl -plaintext -d "{\"workflow_id\": \"$WORKFLOW_ID\"}" \
  localhost:9092 orchestrator.v1.WorkflowService/GetWorkflow
```

**期望结果**: HTTP和gRPC返回的数据应该完全一致（除了格式差异）

---

## 性能对比测试

### HTTP vs gRPC 延迟测试

```bash
# HTTP性能测试 (需要安装ab工具)
ab -n 1000 -c 10 -H "Content-Type: application/json" \
  -p test_request.json \
  http://localhost:8092/v1/workflows

# gRPC性能测试 (需要安装ghz工具)
# brew install ghz
ghz --insecure \
  --proto pkg/api/orchestrator/v1/workflow.proto \
  --call orchestrator.v1.WorkflowService/CreateWorkflow \
  -d '{"name": "perf-test", "description": "性能测试"}' \
  -n 1000 -c 10 \
  localhost:9092
```

---

## 故障排查

### 1. 服务无法启动

**检查端口占用**:
```bash
# 检查端口是否被占用
lsof -i :8092  # Orchestrator HTTP
lsof -i :9092  # Orchestrator gRPC
lsof -i :8082  # Reasoning HTTP
lsof -i :9093  # Reasoning gRPC
```

**检查依赖服务**:
```bash
# 检查PostgreSQL
docker ps | grep postgres

# 测试连接
psql -h localhost -U aetherius -d aetherius_orchestrator
```

### 2. HTTP 404错误

**可能原因**:
- HTTP服务未正确启动
- 路由未注册

**检查方法**:
```bash
# 查看日志
tail -f logs/orchestrator.log

# 检查HTTP server是否启动
curl -v http://localhost:8092/v1/workflows
```

### 3. gRPC连接失败

**检查方法**:
```bash
# 使用grpcurl测试连接
grpcurl -plaintext localhost:9092 list

# 检查reflection是否启用
grpcurl -plaintext localhost:9092 describe
```

### 4. 数据库连接错误

**检查配置**:
```bash
# 查看配置文件
cat configs/orchestrator.yaml

# 检查数据库连接
psql -h localhost -U aetherius -d aetherius_orchestrator -c "\dt"
```

---

## 日志查看

```bash
# Orchestrator日志
tail -f logs/orchestrator.log

# Reasoning日志
tail -f logs/reasoning.log

# 查看gRPC Gateway日志
grep "gRPC-Gateway" logs/*.log

# 查看HTTP请求日志
grep "HTTP" logs/*.log
```

---

## OpenAPI文档

生成的OpenAPI/Swagger文档位于:
```
pkg/api/docs/swagger/api.swagger.json
```

### 使用Swagger UI查看

```bash
# 方式1: 使用Docker
docker run -p 8080:8080 \
  -e SWAGGER_JSON=/api/api.swagger.json \
  -v $(pwd)/pkg/api/docs/swagger:/api \
  swaggerapi/swagger-ui

# 访问: http://localhost:8080

# 方式2: 在线查看
# 访问: https://editor.swagger.io/
# 将api.swagger.json内容粘贴进去
```

---

## 常见问题

### Q: HTTP和gRPC哪个更快？
A: gRPC通常快3-10倍，因为:
- Protocol Buffers二进制序列化
- HTTP/2多路复用
- 更小的数据包

### Q: 如何切换协议？
A: 客户端可以自由选择:
- Web前端 → HTTP
- 服务间调用 → gRPC
- 调试 → HTTP (更易读)

### Q: 数据一致性如何保证？
A: 两种协议使用**完全相同**的service实现，数据100%一致

---

## 下一步

1. **编写集成测试**: 覆盖所有API端点
2. **性能基准测试**: 记录HTTP vs gRPC性能差异
3. **生产部署**: 准备Kubernetes部署配置
4. **监控告警**: 添加Prometheus指标

---

**文档版本**: v1.0
**更新时间**: 2025-11-01
**状态**: ✅ 可用于测试
