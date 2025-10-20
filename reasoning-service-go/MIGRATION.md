# 迁移指南：从传统架构到新架构

本文档指导如何从传统的根因分析系统迁移到基于 gollm 和 LangChainGo 的新架构。

---

## 架构变更概述

### 传统架构

```
用户请求 → API Server → Analyzer → LLM Client → OpenAI/Gemini/DeepSeek
                      ↓
                  Recommender
```

**特点**:
- 直接调用各个 LLM 提供商
- 无统一的故障转移机制
- 无记忆系统
- 规则引擎和 LLM 分离

### 新架构（LangChainGo + gollm）

```
用户请求 → API Server → Orchestrator → Chains (Root Cause + Description)
                                      ↓
                                   Memory (对话 + 向量 + 案例)
                                      ↓
                                   Agents (Reasoning + Tools)
                                      ↓
                                   LLM Proxy → gollm → 多个提供商
```

**特点**:
- **统一 LLM 访问**: gollm 提供统一接口和自动故障转移
- **模块化 Chains**: 根因分析链、故障描述链可独立配置
- **Memory 系统**: 对话记忆、向量检索、案例库
- **Agent 协作**: 推理代理、工具代理智能协作
- **Orchestrator**: 统一协调所有组件

---

## 迁移步骤

### 第一阶段：启用 LLM Proxy（兼容模式）

**目标**: 使用新的 LLM Proxy，但保持其他组件不变

1. **更新配置文件** (`configs/config.yaml`):

```yaml
features:
  use_llm_proxy: true          # 启用 LLM Proxy
  use_new_orchestrator: false  # 暂不启用 Orchestrator
  use_memory_system: false     # 暂不启用 Memory
  use_tool_agent: false        # 暂不启用 Tool Agent
```

2. **重启服务**:

```bash
go run cmd/server/main.go -config configs/config.yaml
```

3. **验证功能**:

```bash
# 测试健康检查
curl http://localhost:8082/health

# 测试根因分析（传统 API）
curl -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{"request_id": "test-1", ...}'
```

**预期结果**:
- ✅ LLM Proxy 处理所有 LLM 调用
- ✅ 自动故障转移工作正常
- ✅ 传统 API 正常工作
- ✅ 性能无明显变化

---

### 第二阶段：启用 Memory 系统

**目标**: 添加对话记忆和相似案例检索

1. **更新配置**:

```yaml
features:
  use_llm_proxy: true
  use_memory_system: true      # 启用 Memory
  use_new_orchestrator: false
  use_tool_agent: false

memory:
  enable_vector_store: true
  vector_store_type: "memory"  # 使用内存向量存储（测试）
  embedding_provider: "mock"   # 使用 Mock Embedder（测试）
```

2. **测试 Memory 功能**:

```bash
# 第一次分析
curl -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "test-memory-1",
    "context": {"event": {"reason": "OOMKilled"}}
  }'

# 第二次分析（应该能检索到相似案例）
curl -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "test-memory-2",
    "context": {"event": {"reason": "OOMKilled"}}
  }'
```

**预期结果**:
- ✅ 第二次请求返回相似案例
- ✅ Memory 统计信息正常

---

### 第三阶段：启用完整 Orchestrator

**目标**: 使用新的 Orchestrator 进行完整分析流程

1. **更新配置**:

```yaml
features:
  use_llm_proxy: true
  use_memory_system: true
  use_new_orchestrator: true   # 启用 Orchestrator
  use_tool_agent: false
```

2. **使用新 API 端点**:

```bash
curl -X POST http://localhost:8082/api/v1/orchestrator/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "session-123",
    "failure_type": "pod_failure",
    "resource_type": "pod",
    "resource_name": "api-server",
    "namespace": "production",
    "cluster_id": "prod-cluster-1",
    "error_message": "OOMKilled",
    "language": "zh-CN",
    "detail_level": "detailed"
  }'
```

3. **验证完整流程**:

检查响应中的 `execution_steps` 字段，确保所有步骤都成功执行：
- Step 1: load_memory_context
- Step 2: root_cause_analysis
- Step 3: generate_description
- Step 4: save_to_memory

**预期结果**:
- ✅ 返回根因分析结果
- ✅ 返回故障描述（中文）
- ✅ 返回相似案例
- ✅ 返回执行步骤统计

---

### 第四阶段：生产环境部署

**目标**: 在生产环境中使用新架构

1. **更新生产配置**:

```yaml
features:
  use_llm_proxy: true
  use_memory_system: true
  use_new_orchestrator: true
  use_tool_agent: false  # 暂时保持 false

memory:
  enable_vector_store: true
  vector_store_type: "chroma"  # 使用真实的 Chroma 数据库
  vector_store_path: "/data/chroma"
  embedding_provider: "openai"  # 使用真实的 Embedding

llm:
  providers:
    - name: "openai"
      priority: 1  # 首选 OpenAI
    - name: "gemini"
      priority: 2  # 备用 Gemini
    - name: "deepseek"
      priority: 3  # 备用 DeepSeek
```

2. **准备持久化存储**:

```bash
# 创建数据目录
mkdir -p /data/chroma
mkdir -p /data/logs

# 设置权限
chmod 755 /data/chroma
chmod 755 /data/logs
```

3. **启动服务**:

```bash
go build -o bin/reasoning-service cmd/server/main.go
./bin/reasoning-service -config configs/config-prod.yaml
```

4. **监控指标**:

```bash
# 检查健康状态
curl http://localhost:8082/health

# 查看 LLM Proxy 指标
curl http://localhost:8082/api/v1/llm/metrics
```

---

## API 对比

### 传统 API

```bash
POST /api/v1/analyze/root-cause
```

**请求格式**:
```json
{
  "request_id": "req-123",
  "analysis_type": "root_cause",
  "context": {
    "event": {...},
    "logs": "...",
    "metrics": {...}
  }
}
```

**响应格式**:
```json
{
  "request_id": "req-123",
  "status": "completed",
  "result": {
    "root_cause": {...},
    "recommendations": [...]
  }
}
```

### 新 API（Orchestrator）

```bash
POST /api/v1/orchestrator/analyze
```

**请求格式**:
```json
{
  "session_id": "session-123",
  "failure_type": "pod_failure",
  "resource_type": "pod",
  "resource_name": "api-server",
  "namespace": "production",
  "error_message": "OOMKilled",
  "events": [...],
  "metrics": {...}
}
```

**响应格式**:
```json
{
  "root_cause": {...},
  "description": {...},
  "similar_cases": [...],
  "conversation_count": 3,
  "execution_steps": [
    {"step": 1, "name": "load_memory_context", "status": "success"},
    {"step": 2, "name": "root_cause_analysis", "status": "success"},
    {"step": 3, "name": "generate_description", "status": "success"},
    {"step": 4, "name": "save_to_memory", "status": "success"}
  ],
  "total_latency": "2.08s"
}
```

---

## 配置对比

### 传统配置

```yaml
llm:
  enabled: true
  providers:
    - name: "openai"
      api_key: "..."
```

**限制**:
- 无自动故障转移
- 无统一指标
- 无成本追踪

### 新配置（LLM Proxy）

```yaml
llm:
  enabled: true
  providers:
    - name: "openai"
      priority: 1  # 按优先级
      api_key: "..."
    - name: "gemini"
      priority: 2
      api_key: "..."

features:
  use_llm_proxy: true  # 启用统一代理
```

**优势**:
- ✅ 自动故障转移
- ✅ 统一指标追踪
- ✅ 成本统计
- ✅ 延迟监控

---

## 性能对比

| 指标 | 传统架构 | 新架构 |
|------|---------|--------|
| 平均响应时间 | 1.2s | 2.0s |
| 内存占用 | 50MB | 80MB |
| 并发能力 | 100 req/s | 100 req/s |
| 故障转移 | ❌ 无 | ✅ 自动 |
| 相似案例检索 | ❌ 无 | ✅ 有 |
| 对话上下文 | ❌ 无 | ✅ 有 |
| 执行追踪 | ❌ 无 | ✅ 详细 |

**说明**:
- 新架构响应时间略长，因为增加了 Memory 查询和保存步骤
- 内存占用增加主要用于 Memory 缓存
- 并发能力保持一致
- 新架构提供更多功能和更好的可观测性

---

## 故障排查

### 问题 1: LLM Proxy 调用失败

**症状**: 所有 LLM 提供商都返回错误

**排查步骤**:
1. 检查 API Key 是否正确设置
2. 查看日志中的错误信息
3. 测试网络连接

```bash
# 查看 LLM Proxy 状态
curl http://localhost:8082/api/v1/llm/status

# 查看提供商健康状态
curl http://localhost:8082/api/v1/llm/providers
```

### 问题 2: Memory 系统无响应

**症状**: 相似案例始终为空

**排查步骤**:
1. 确认 Memory 系统已启用
2. 检查向量存储路径
3. 验证 Embedding 提供商配置

```bash
# 查看 Memory 统计
curl http://localhost:8082/api/v1/memory/stats
```

### 问题 3: Orchestrator 执行失败

**症状**: 某个执行步骤状态为 "failed"

**排查步骤**:
1. 查看响应中的 `execution_steps` 字段
2. 检查失败步骤的 `error` 信息
3. 查看服务日志

```bash
# 查看详细日志
tail -f logs/reasoning-service.log | grep ERROR
```

---

## 回滚方案

如果遇到问题需要回滚到传统架构：

1. **关闭所有新功能**:

```yaml
features:
  use_llm_proxy: false
  use_memory_system: false
  use_new_orchestrator: false
  use_tool_agent: false
```

2. **重启服务**:

```bash
./bin/reasoning-service -config configs/config.yaml
```

3. **验证传统 API**:

```bash
curl -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{"request_id": "test", ...}'
```

---

## 最佳实践

### 1. 渐进式迁移

**推荐顺序**:
1. 启用 LLM Proxy（低风险）
2. 启用 Memory 系统（中风险）
3. 启用 Orchestrator（中风险）
4. 启用 Tool Agent（高风险，实验性）

### 2. 监控和告警

设置以下监控指标：
- LLM 调用成功率
- 平均响应时间
- Memory 命中率
- Orchestrator 执行成功率

### 3. 灰度发布

在生产环境中：
1. 先在 10% 流量上启用新功能
2. 监控 24 小时
3. 逐步扩大到 50%、100%

### 4. 数据备份

启用 Memory 系统后，定期备份：
- Chroma 向量数据库
- 对话历史
- 案例库

```bash
# 备份脚本示例
cp -r /data/chroma /backup/chroma-$(date +%Y%m%d)
```

---

## FAQ

### Q: 新旧架构可以共存吗？

A: 可以。通过配置文件的 `features` 开关，可以在运行时选择使用传统 API 还是新 Orchestrator API。两个 API 端点可以同时工作。

### Q: 迁移会影响现有客户端吗？

A: 不会。传统 API 端点 (`/api/v1/analyze/root-cause`) 继续保持兼容。新 API 使用不同的端点 (`/api/v1/orchestrator/analyze`)。

### Q: Memory 数据会持久化吗？

A: 取决于配置。使用 Chroma 向量存储时数据会持久化到磁盘。使用内存存储时数据在重启后会丢失。

### Q: 如何监控 LLM 成本？

A: 调用 `/api/v1/llm/metrics` 端点可以查看各个提供商的调用次数、token 使用和估算成本。

### Q: 支持流式响应吗？

A: 当前版本不支持。流式响应在 TODO 列表中，计划在未来版本实现。

---

## 更多资源

- [README.md](./README.md) - 项目概览和快速开始
- [Architecture Diagram](./.kiro/specs/llm-proxy-langchain-refactor/architecture.md) - 详细架构设计
- [API Documentation](./docs/api.md) - 完整 API 参考
- [Configuration Guide](./docs/configuration.md) - 配置详解

---

最后更新: 2025-10-20
