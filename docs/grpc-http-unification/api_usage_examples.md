# gRPC-HTTP统一Handler API使用示例

## 概述

k8s-agent项目的Orchestrator和Reasoning服务现在支持**双协议访问**：
- **gRPC**: 高性能二进制协议，适合服务间通信
- **HTTP/JSON**: RESTful API，适合Web客户端和调试

**核心特性**: 两种协议使用同一个service实现，零代码重复！

---

## 1. Orchestrator Service

### 服务端口
- **HTTP**: `localhost:8092`
- **gRPC**: `localhost:9092`

### HTTP API示例

#### 1.1 创建工作流
```bash
curl -X POST http://localhost:8092/v1/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "diagnose-pod-crash",
    "description": "Diagnose pod crash loop",
    "steps": [
      {
        "id": "collect_logs",
        "name": "Collect Pod Logs",
        "type": "COMMAND",
        "config": {
          "command": "kubectl logs",
          "args": ["--tail=100", "${pod_name}"],
          "namespace": "${namespace}"
        }
      },
      {
        "id": "analyze",
        "name": "AI Analysis",
        "type": "AI",
        "config": {
          "service": "reasoning",
          "method": "RootCauseAnalysis"
        }
      }
    ]
  }'
```

**响应示例**:
```json
{
  "workflow_id": "wf-a1b2c3d4",
  "name": "diagnose-pod-crash",
  "status": "CREATED",
  "created_at": "2025-11-01T12:00:00Z"
}
```

#### 1.2 获取工作流详情
```bash
curl http://localhost:8092/v1/workflows/wf-a1b2c3d4
```

#### 1.3 列出所有工作流
```bash
curl "http://localhost:8092/v1/workflows?page=1&page_size=20"
```

#### 1.4 执行工作流
```bash
curl -X POST http://localhost:8092/v1/workflows/wf-a1b2c3d4/executions \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "evt-crash-12345",
    "params": {
      "pod_name": "my-app-7d8f9b5c-xyz",
      "namespace": "production"
    }
  }'
```

**响应示例**:
```json
{
  "execution_id": "exec-e5f6g7h8",
  "workflow_id": "wf-a1b2c3d4",
  "status": "RUNNING",
  "started_at": "2025-11-01T12:05:00Z"
}
```

#### 1.5 查询执行状态
```bash
curl http://localhost:8092/v1/executions/exec-e5f6g7h8
```

**响应示例**:
```json
{
  "execution_id": "exec-e5f6g7h8",
  "workflow_id": "wf-a1b2c3d4",
  "status": "COMPLETED",
  "started_at": "2025-11-01T12:05:00Z",
  "completed_at": "2025-11-01T12:05:30Z",
  "steps": [
    {
      "id": "collect_logs",
      "status": "COMPLETED",
      "output": "Error: OOMKilled..."
    },
    {
      "id": "analyze",
      "status": "COMPLETED",
      "output": {
        "root_cause": "RESOURCE_EXHAUSTION",
        "confidence": 0.95
      }
    }
  ]
}
```

### gRPC API示例

```go
package main

import (
    "context"
    "fmt"
    "log"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    orchestratorv1 "github.com/kart-io/k8s-agent/pkg/api/orchestrator/v1"
)

func main() {
    // 连接到gRPC服务
    conn, err := grpc.Dial("localhost:9092",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatalf("连接失败: %v", err)
    }
    defer conn.Close()

    // 创建客户端
    client := orchestratorv1.NewWorkflowServiceClient(conn)
    ctx := context.Background()

    // 1. 创建工作流
    createResp, err := client.CreateWorkflow(ctx, &orchestratorv1.CreateWorkflowRequest{
        Name:        "diagnose-pod-crash",
        Description: "Diagnose pod crash loop",
        Steps: []*orchestratorv1.WorkflowStep{
            {
                Id:   "collect_logs",
                Name: "Collect Pod Logs",
                Type: orchestratorv1.WorkflowStep_COMMAND,
            },
        },
    })
    if err != nil {
        log.Fatalf("创建工作流失败: %v", err)
    }
    fmt.Printf("工作流已创建: %s\n", createResp.WorkflowId)

    // 2. 获取工作流
    workflow, err := client.GetWorkflow(ctx, &orchestratorv1.GetWorkflowRequest{
        WorkflowId: createResp.WorkflowId,
    })
    if err != nil {
        log.Fatalf("获取工作流失败: %v", err)
    }
    fmt.Printf("工作流名称: %s\n", workflow.Name)

    // 3. 执行工作流
    execResp, err := client.ExecuteWorkflow(ctx, &orchestratorv1.ExecuteWorkflowRequest{
        WorkflowId: createResp.WorkflowId,
        EventId:    "evt-crash-12345",
    })
    if err != nil {
        log.Fatalf("执行工作流失败: %v", err)
    }
    fmt.Printf("执行ID: %s\n", execResp.ExecutionId)
}
```

---

## 2. Reasoning Service

### 服务端口
- **HTTP**: `localhost:8082`
- **gRPC**: `localhost:9093`

### HTTP API示例

#### 2.1 根因分析
```bash
curl -X POST http://localhost:8082/v1/analysis/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "evt-crash-12345",
    "context": {
      "events": [
        {
          "type": "CrashLoopBackOff",
          "message": "Back-off restarting failed container",
          "timestamp": "2025-11-01T12:00:00Z"
        }
      ],
      "logs": [
        "fatal error: runtime: out of memory",
        "runtime stack:",
        "runtime.throw(0x1c8a5e0, 0x16)"
      ],
      "metrics": {
        "memory": {
          "usage_percent": 98.5,
          "usage_bytes": 2097152000,
          "limit_bytes": 2147483648
        },
        "cpu": {
          "usage_percent": 45.2
        }
      },
      "resources": {
        "pod": "my-app-7d8f9b5c-xyz",
        "namespace": "production",
        "container": "app"
      }
    },
    "options": {
      "use_knowledge_graph": true,
      "include_historical_cases": true,
      "model": "openai"
    }
  }'
```

**响应示例**:
```json
{
  "analysis_id": "analysis-i9j0k1l2",
  "root_cause": {
    "type": "RESOURCE_EXHAUSTION",
    "description": "容器内存使用超过限制，导致OOMKilled",
    "confidence": 0.95,
    "evidences": [
      {
        "type": "LOG",
        "content": "fatal error: runtime: out of memory",
        "source": "container-logs",
        "timestamp": "2025-11-01T12:00:00Z"
      },
      {
        "type": "METRIC",
        "content": "Memory usage: 98.5% (2GB/2GB)",
        "source": "prometheus",
        "timestamp": "2025-11-01T12:00:00Z"
      }
    ]
  },
  "recommendations": [
    {
      "id": "rec-1",
      "type": "AUTOMATED",
      "title": "增加内存限制",
      "description": "建议将容器内存限制从2GB增加到4GB",
      "steps": [
        "编辑Deployment YAML",
        "修改resources.limits.memory为4Gi",
        "应用配置: kubectl apply -f deployment.yaml"
      ],
      "commands": [
        "kubectl set resources deployment my-app --limits=memory=4Gi -n production"
      ],
      "priority": 8,
      "expected_result": "容器不再OOMKilled，稳定运行",
      "risk_assessment": "低风险 - 仅增加资源配额"
    }
  ],
  "similar_cases": [
    {
      "id": "case-m3n4o5p6",
      "title": "Production Memory OOM - 2025-10-15",
      "description": "类似的内存耗尽问题",
      "solution": "增加内存限制到4GB",
      "similarity": 0.89,
      "occurred_at": "2025-10-15T08:30:00Z"
    }
  ],
  "confidence": 0.95,
  "duration_ms": 3542
}
```

#### 2.2 保存案例
```bash
curl -X POST http://localhost:8082/v1/cases \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "evt-crash-12345",
    "analysis_id": "analysis-i9j0k1l2",
    "result": "CORRECT",
    "solution": "Increased memory limit to 4GB, issue resolved",
    "feedback": "Analysis was accurate, recommendation worked perfectly"
  }'
```

**响应示例**:
```json
{
  "case_id": "case-q7r8s9t0",
  "success": true
}
```

### gRPC API示例

```go
package main

import (
    "context"
    "fmt"
    "log"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    reasoningv1 "github.com/kart-io/k8s-agent/pkg/api/reasoning/v1"
)

func main() {
    // 连接到gRPC服务
    conn, err := grpc.Dial("localhost:9093",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatalf("连接失败: %v", err)
    }
    defer conn.Close()

    // 创建客户端
    client := reasoningv1.NewReasoningServiceClient(conn)
    ctx := context.Background()

    // 1. 根因分析
    analysisResp, err := client.RootCauseAnalysis(ctx, &reasoningv1.RootCauseAnalysisRequest{
        EventId: "evt-crash-12345",
        Context: &reasoningv1.AnalysisContext{
            Events: []*reasoningv1.Event{
                {
                    Type:    "CrashLoopBackOff",
                    Message: "Back-off restarting failed container",
                },
            },
            Logs: []string{
                "fatal error: runtime: out of memory",
            },
        },
        Options: &reasoningv1.AnalysisOptions{
            UseKnowledgeGraph:      true,
            IncludeHistoricalCases: true,
            Model:                  "openai",
        },
    })
    if err != nil {
        log.Fatalf("根因分析失败: %v", err)
    }

    fmt.Printf("分析ID: %s\n", analysisResp.AnalysisId)
    fmt.Printf("根因类型: %s\n", analysisResp.RootCause.Type)
    fmt.Printf("置信度: %.2f\n", analysisResp.Confidence)
    fmt.Printf("建议数量: %d\n", len(analysisResp.Recommendations))

    // 2. 保存案例
    caseResp, err := client.SaveCase(ctx, &reasoningv1.SaveCaseRequest{
        EventId:    "evt-crash-12345",
        AnalysisId: analysisResp.AnalysisId,
        Result:     reasoningv1.CaseResult_CORRECT,
        Solution:   "Increased memory limit to 4GB",
    })
    if err != nil {
        log.Fatalf("保存案例失败: %v", err)
    }

    fmt.Printf("案例已保存: %s\n", caseResp.CaseId)
}
```

---

## 3. 协议对比

### 性能特性

| 特性 | HTTP/JSON | gRPC |
|------|-----------|------|
| 序列化 | JSON (文本) | Protocol Buffers (二进制) |
| 传输效率 | 较低 | 高 (约3-10倍) |
| 可读性 | 高 (易调试) | 低 (需要工具) |
| 浏览器支持 | 原生支持 | 需要gRPC-Web |
| 适用场景 | Web客户端、调试 | 服务间通信 |

### 使用建议

**使用HTTP/JSON时**:
- ✅ Web前端调用
- ✅ 命令行调试 (curl)
- ✅ 第三方集成
- ✅ 快速原型开发

**使用gRPC时**:
- ✅ 微服务间通信
- ✅ 高性能要求
- ✅ 流式传输
- ✅ 强类型保证

---

## 4. 测试工具

### HTTP测试

**使用curl**:
```bash
# 简单GET请求
curl http://localhost:8092/v1/workflows/wf-12345

# POST请求
curl -X POST http://localhost:8082/v1/analysis/root-cause \
  -H "Content-Type: application/json" \
  -d @request.json
```

**使用httpie** (更友好的输出):
```bash
# 安装: brew install httpie

# GET请求
http GET localhost:8092/v1/workflows/wf-12345

# POST请求
http POST localhost:8082/v1/analysis/root-cause < request.json
```

### gRPC测试

**使用grpcurl**:
```bash
# 安装: brew install grpcurl

# 列出服务
grpcurl -plaintext localhost:9092 list

# 调用方法
grpcurl -plaintext -d '{
  "workflow_id": "wf-12345"
}' localhost:9092 orchestrator.v1.WorkflowService/GetWorkflow
```

---

## 5. 常见问题

### Q1: HTTP和gRPC返回的数据一样吗？
**A**: 是的！两种协议使用完全相同的service实现，返回的数据结构和内容完全一致，只是序列化格式不同（JSON vs Protobuf）。

### Q2: 性能差异有多大？
**A**: gRPC通常比HTTP/JSON快3-10倍，主要因为：
- Protocol Buffers二进制序列化更高效
- HTTP/2多路复用
- 更小的数据包大小

### Q3: 可以混用HTTP和gRPC吗？
**A**: 可以！你可以在不同场景使用不同协议：
- 前端调用用HTTP
- 后端服务间用gRPC
- 开发调试用HTTP
- 生产环境用gRPC

### Q4: 如何选择端口？
**A**: 服务端口配置：
- Orchestrator: HTTP 8092, gRPC 9092
- Reasoning: HTTP 8082, gRPC 9093
可以在配置文件中修改

---

## 6. 下一步

### 集成示例
- [ ] 编写完整的客户端SDK
- [ ] 添加认证和授权示例
- [ ] 编写集成测试用例

### 文档
- [ ] 生成OpenAPI/Swagger文档
- [ ] 添加更多使用场景示例
- [ ] 性能基准测试报告

---

**生成时间**: 2025-11-01
**作者**: Claude Code (gRPC-HTTP统一Handler实现)
