# Aetherius Reasoning Service (Go Implementation)

**AI-powered root cause analysis service with LLM support**

完整的 Go 语言实现，基于 **gollm** 和 **LangChainGo** 框架，集成多个 LLM 提供商，提供智能的 Kubernetes 故障诊断和根因分析。

---

## ✨ 主要特性

### 🤖 统一的 LLM 访问层 (gollm)
- **多提供商支持**: OpenAI (GPT-4), Google Gemini, DeepSeek, **SiliconFlow**, **Kimi (月之暗面)**, **Ollama (本地部署)**, **自定义 LLM 服务**
- **统一接口**: 通过 `go-llm-proxy` 统一访问所有 LLM 提供商
- **自动故障转移**: 按优先级自动切换提供商，保证服务可用性
- **智能后备**: 规则引擎 + LLM 混合分析，自动降级
- **成本优化**: 支持多个模型配置，智能选择最优提供商
- **私有部署**: 支持 Ollama 本地模型和自建 LLM 服务，数据不出网
- **国内优化**: SiliconFlow 和 Kimi 国内访问速度快
- **使用指标**: 实时统计调用次数、成功率、延迟和成本

### 🔗 LangChainGo 架构
- **Chain**: 模块化的分析链（根因分析链、故障描述链）
- **Agent**: 智能代理（推理代理、工具代理）
- **Memory**: 三层记忆系统
  - **对话记忆**: 会话上下文管理
  - **向量存储**: 基于语义的相似案例检索
  - **案例记忆**: 历史故障案例库
- **Tools**: K8s 工具集成（事件查询、日志获取、指标采集）

### 🔍 Orchestrator 协调器
- **统一入口**: 协调所有分析组件（Chains、Agents、Memory、Tools）
- **执行追踪**: 详细记录每个步骤的执行状态和耗时
- **智能流程**:
  1. **加载上下文**: 从 Memory 加载对话历史和相似案例
  2. **根因分析**: 调用根因分析链，利用历史经验
  3. **生成描述**: 生成人类可读的故障描述
  4. **保存记忆**: 将分析结果保存到 Memory 供未来使用
- **灵活配置**: 可独立启用/禁用各个功能模块
- **超时控制**: 各阶段独立超时配置，防止长时间阻塞

### 🔍 根因分析
- **多模态分析**: 综合事件、日志、指标进行分析
- **规则引擎**: 基于 Kubernetes 最佳实践的模式匹配
- **LLM 增强**: 深度分析和智能推理

### 💡 智能推荐
- **规则库**: 预定义的修复动作和步骤
- **LLM 增强**: 动态生成针对性建议
- **风险评估**: 包含风险等级、回滚步骤

### 🎯 支持的根因类型
- OOMKiller - 内存溢出
- CPUThrottling - CPU 限流
- DiskPressure - 磁盘压力
- NetworkError - 网络错误
- ConfigError - 配置错误
- ImagePullError - 镜像拉取失败
- VolumeError - 存储卷错误
- ResourceLimit - 资源限制

---

## 🚀 快速开始

### 前置要求
- Go 1.21+
- 至少一个 LLM 提供商:
  - **云端**: OpenAI / Gemini / DeepSeek API Key
  - **本地**: Ollama (推荐，免费且私有)

### 安装依赖

```bash
go mod download
```

### 配置

编辑 `configs/config.yaml` 或设置环境变量：

```bash
# OpenAI
export OPENAI_API_KEY="sk-..."

# Google Gemini
export GEMINI_API_KEY="..."
# 或
export GOOGLE_API_KEY="..."

# DeepSeek
export DEEPSEEK_API_KEY="..."

# SiliconFlow (国内大模型平台)
export SILICONFLOW_API_KEY="..."

# Kimi (月之暗面 Moonshot AI)
export KIMI_API_KEY="..."

# Ollama (本地部署，无需 API Key)
# 1. 安装 Ollama: curl -fsSL https://ollama.com/install.sh | sh
# 2. 下载模型: ollama pull llama3.1
# 3. Ollama 会自动在 localhost:11434 运行

# 自定义 LLM 服务 (任何 OpenAI 兼容的 API)
export CUSTOM_LLM_BASE_URL="http://your-llm-service:8000/v1"
export CUSTOM_LLM_MODEL="your-model-name"
export CUSTOM_LLM_API_KEY="..."  # 如果需要认证
```

### 运行

```bash
# 使用默认配置
go run cmd/server/main.go

# 指定配置文件
go run cmd/server/main.go -config configs/config.yaml
```

服务启动在 `http://localhost:8082`

---

## 📡 API 使用

### 健康检查

```bash
curl http://localhost:8082/health
```

**响应**:
```json
{
  "status": "healthy",
  "service": "reasoning-service-go",
  "components": {
    "analyzer": true,
    "recommender": true,
    "llm": true,
    "orchestrator": true
  },
  "timestamp": "2025-10-03T14:00:00Z"
}
```

### 根因分析（传统 API）

```bash
curl -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "req-123",
    "analysis_type": "root_cause",
    "context": {
      "event": {
        "reason": "OOMKilled",
        "message": "Container killed due to OOM"
      },
      "logs": "fatal error: runtime: out of memory\ngoroutine stack exceeds limit",
      "metrics": {
        "memory": {
          "usage_percent": 98.5
        },
        "cpu": {
          "usage_percent": 75.0
        }
      }
    },
    "options": {
      "use_llm": true,
      "llm_provider": "openai",
      "min_confidence": 0.7,
      "max_recommendations": 5
    }
  }'
```

### Orchestrator 分析（新 API）

使用新的 Orchestrator 进行完整的分析流程：

```bash
curl -X POST http://localhost:8082/api/v1/orchestrator/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "session-123",
    "failure_type": "pod_failure",
    "resource_type": "pod",
    "resource_name": "api-server-7d9f8c",
    "namespace": "production",
    "cluster_id": "prod-cluster-1",
    "error_message": "OOMKilled",
    "timestamp": "2025-10-03T14:30:00Z",
    "language": "zh-CN",
    "detail_level": "detailed",
    "events": [
      {
        "type": "Warning",
        "reason": "OOMKilled",
        "message": "Container exceeded memory limit",
        "last_timestamp": "2025-10-03T14:30:00Z",
        "source": "kubelet"
      }
    ],
    "metrics": {
      "cpu": {
        "utilization": 0.75,
        "cores_used": 1.5,
        "limit": 2.0
      },
      "memory": {
        "utilization": 0.985,
        "bytes_used": 1020054732,
        "limit": 1073741824
      }
    }
  }'
```

**Orchestrator 响应**:
```json
{
  "root_cause": {
    "root_cause": "Pod OOMKilled due to memory limit exceeded",
    "confidence": 0.95,
    "category": "resource_exhaustion",
    "reasoning": "Container memory usage reached 98.5% before being killed...",
    "recommendations": [
      "Increase memory limit to 2Gi",
      "Add memory resource requests",
      "Review application memory leaks"
    ]
  },
  "description": {
    "title": "内存溢出导致容器被终止",
    "summary": "生产环境中的 API 服务器容器因内存使用超限被 Kubernetes 强制终止",
    "affected_components": ["api-server", "database-connection-pool"],
    "severity": "high",
    "timeline": ["14:29:30 - 内存使用达到 95%", "14:30:00 - OOMKilled 事件触发"]
  },
  "similar_cases": [
    {
      "description": "Previous OOM issue in same pod",
      "root_cause": "Memory leak in cache layer",
      "similarity": 0.87,
      "solution": "Fixed by clearing cache periodically"
    }
  ],
  "conversation_count": 3,
  "execution_steps": [
    {
      "step": 1,
      "name": "load_memory_context",
      "description": "Load history and similar cases from memory",
      "status": "success",
      "duration": "50ms"
    },
    {
      "step": 2,
      "name": "root_cause_analysis",
      "description": "Analyze root cause using LLM and rules",
      "status": "success",
      "duration": "1.2s"
    },
    {
      "step": 3,
      "name": "generate_description",
      "description": "Generate human-readable failure description",
      "status": "success",
      "duration": "800ms"
    },
    {
      "step": 4,
      "name": "save_to_memory",
      "description": "Save analysis result to memory",
      "status": "success",
      "duration": "30ms"
    }
  ],
  "total_latency": "2.08s",
  "timestamp": "2025-10-03T14:30:02Z"
}
```

### K8s Event 分析

**🎉 统一架构**: `/api/v1/analyze/k8s-event` 端点完全使用 Orchestrator 架构！

K8s Event 分析端点直接使用 Orchestrator 架构，提供：
- ✅ LLM 自动故障转移
- ✅ 相似案例检索
- ✅ 对话上下文记忆
- ✅ 详细的执行步骤追踪
- ✅ 多语言故障描述

**使用方法**:

```bash
curl -X POST http://localhost:8082/api/v1/analyze/k8s-event \
  -H "Content-Type: application/json" \
  -d '{
    "cluster_id": "prod-cluster-1",
    "event": {
      "reason": "OOMKilled",
      "message": "Container exceeded memory limit",
      "type": "Warning",
      "involvedObject": {
        "namespace": "production",
        "name": "api-server-pod",
        "kind": "Pod"
      },
      "source": {
        "component": "kubelet"
      }
    },
    "use_llm": true
  }'
```

**响应格式**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "analysis": "<div class=\"diagnosis-section\">...</div>",
    "rootCause": "OOMKiller",
    "confidence": 0.95,
    "recommendations": [
      "将 Pod 的内存限制从 512Mi 增加到 1Gi",
      "检查应用程序是否有内存泄漏"
    ]
  }
}
```

**必需配置**:

```yaml
features:
  use_new_orchestrator: true  # 必须启用
  use_llm_proxy: true          # 必须启用
  use_memory_system: true      # 推荐启用
```

**注意事项**:
- Orchestrator 必须正确初始化，否则返回 503 错误
- 需要至少配置一个 LLM 提供商
- 建议启用 Memory 系统以获得相似案例功能

---

## ⚙️ 配置说明

### 功能开关

```yaml
features:
  # 新架构功能开关
  use_new_orchestrator: true   # 启用新的 Orchestrator (推荐)
  use_llm_proxy: true           # 启用 LLM Proxy Adapter (推荐)
  use_memory_system: true       # 启用 Memory 系统
  use_tool_agent: false         # 启用 Tool Agent (实验性)

  # 传统功能开关
  enable_prediction: false
  enable_learning: false
  enable_knowledge_graph: false
  enable_anomaly_detection: false
  enable_case_similarity: true
```

### Memory 系统配置

```yaml
memory:
  enable_vector_store: true              # 启用向量存储
  vector_store_type: "chroma"            # 向量存储类型 (chroma/memory)
  vector_store_path: "./data/chroma"     # Chroma 数据路径
  embedding_model: "text-embedding-ada-002"  # Embedding 模型
  embedding_provider: "openai"           # Embedding 提供商 (openai/local)
```

### LLM 提供商配置

```yaml
llm:
  enabled: true
  providers:
    - name: "openai"
      api_key: ""  # 通过 OPENAI_API_KEY 环境变量设置
      model: "gpt-4o-mini"  # 或 "gpt-4", "gpt-3.5-turbo"
      max_tokens: 4096
      temperature: 0.7
      priority: 1  # 数字越小优先级越高

    - name: "gemini"
      api_key: ""  # 通过 GEMINI_API_KEY 环境变量设置
      model: "gemini-1.5-flash"  # 或 "gemini-1.5-pro"
      max_tokens: 8192
      priority: 2

    - name: "deepseek"
      api_key: ""  # 通过 DEEPSEEK_API_KEY 环境变量设置
      model: "deepseek-chat"
      max_tokens: 4096
      priority: 3

    - name: "ollama"
      api_key: ""  # 本地 Ollama 无需 API Key
      base_url: "http://localhost:11434/v1"
      model: "llama3.1"  # 或 llama3.2, qwen2.5, mistral 等
      max_tokens: 4096
      timeout: 60
      priority: 4  # 可设为 1 作为首选（本地推理，免费）

    - name: "siliconflow"
      api_key: ""  # 通过 SILICONFLOW_API_KEY 环境变量设置
      model: "Qwen/Qwen2.5-7B-Instruct"
      priority: 5

    - name: "kimi"
      api_key: ""  # 通过 KIMI_API_KEY 环境变量设置
      model: "moonshot-v1-8k"  # 或 moonshot-v1-32k, moonshot-v1-128k
      priority: 6

    - name: "custom"
      api_key: ""  # 通过 CUSTOM_LLM_API_KEY 环境变量设置（如需认证）
      base_url: "http://your-llm-service:8000/v1"  # 自定义服务地址
      model: "your-model-name"  # 模型名称
      max_tokens: 4096
      timeout: 60
      priority: 7
```

### 分析设置

```yaml
analysis:
  min_confidence: 0.7  # 最低置信度阈值
  max_recommendations: 5  # 最多推荐数量
  use_llm_fallback: true  # 规则引擎置信度低时使用 LLM
```

---

## 🏗️ 项目结构

```
reasoning-service-go/
├── cmd/
│   └── server/
│       └── main.go              # 主程序入口
├── internal/
│   ├── api/
│   │   └── server.go            # HTTP API 服务器
│   ├── agents/                  # Agent 实现
│   │   ├── k8s_tool/            # K8s 工具 Agent
│   │   └── reasoning/           # 推理 Agent
│   ├── chains/                  # Chain 实现
│   │   ├── description/         # 故障描述 Chain
│   │   └── root_cause/          # 根因分析 Chain
│   ├── memory/                  # Memory 系统
│   │   ├── manager.go           # Memory 管理器
│   │   ├── conversation.go      # 对话记忆
│   │   ├── vectorstore.go       # 向量存储
│   │   └── embedder.go          # 嵌入向量生成
│   ├── orchestrator/            # Orchestrator 协调器
│   │   ├── orchestrator.go      # 核心协调逻辑
│   │   └── types.go             # 类型定义
│   ├── analyzer/
│   │   └── root_cause.go        # 传统根因分析器（向后兼容）
│   ├── recommender/
│   │   └── engine.go            # 推荐引擎（向后兼容）
│   └── config/
│       └── config.go            # 配置管理
├── pkg/
│   ├── llm/
│   │   ├── proxy/               # LLM Proxy 适配器
│   │   │   ├── adapter.go       # gollm 适配器
│   │   │   ├── types.go         # 类型定义
│   │   │   └── metrics.go       # 使用指标
│   │   ├── interface.go         # LLM 客户端接口（传统）
│   │   ├── openai.go            # OpenAI 客户端（传统）
│   │   ├── gemini.go            # Gemini 客户端（传统）
│   │   └── deepseek.go          # DeepSeek 客户端（传统）
│   └── types/
│       └── types.go             # 类型定义
├── tests/
│   └── integration/             # 集成测试
│       ├── orchestrator_test.go # Orchestrator 集成测试
│       └── testutil/            # 测试工具
├── configs/
│   └── config.yaml              # 配置文件
├── go.mod
├── Makefile
└── README.md
```

---

## 🔧 开发

### 构建

```bash
go build -o bin/reasoning-service cmd/server/main.go
```

### 运行

```bash
./bin/reasoning-service -config configs/config.yaml
```

### Docker

```bash
# 构建镜像
docker build -t reasoning-service-go:latest .

# 运行容器
docker run -d \
  -p 8082:8082 \
  -e OPENAI_API_KEY="sk-..." \
  -v $(pwd)/configs:/app/configs \
  reasoning-service-go:latest
```

---

## 🎯 与 Python 版本的对比

| 特性 | Python 版本 | Go 版本 |
|------|------------|---------|
| 规则引擎 | ✅ | ✅ |
| LLM 集成 | ❌ | ✅ (OpenAI, Gemini, DeepSeek, **SiliconFlow**, **Kimi**, **Ollama**, **自定义**) |
| 本地 LLM | ❌ | ✅ (**Ollama**, **自定义服务**) |
| 国内 LLM | ❌ | ✅ (**SiliconFlow**, **Kimi**) |
| 自建服务 | ❌ | ✅ (**vLLM, FastChat, LocalAI 等**) |
| 性能 | 中等 | 高（并发、低内存） |
| 部署 | FastAPI + Uvicorn | 单一二进制文件 |
| 依赖 | 多个 Python 包 | 仅 YAML 解析库 |
| 容器大小 | ~200MB | ~20MB |
| 启动时间 | 慢 | 快 |
| 私有部署 | 部分支持 | ✅ (Ollama 本地模型 + 自定义 LLM) |
| Neo4j 支持 | ✅ | 🔜 (计划中) |
| 异常检测 | ✅ (sklearn) | 🔜 (计划中) |

---

## 🚦 使用示例

### 示例 1: 分析 OOM 事件

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

func main() {
    request := map[string]interface{}{
        "request_id": "req-001",
        "context": map[string]interface{}{
            "event": map[string]string{
                "reason": "OOMKilled",
            },
            "logs": "fatal error: out of memory",
            "metrics": map[string]interface{}{
                "memory": map[string]float64{
                    "usage_percent": 98.0,
                },
            },
        },
        "options": map[string]interface{}{
            "use_llm": true,
        },
    }

    body, _ := json.Marshal(request)
    resp, _ := http.Post(
        "http://localhost:8082/api/v1/analyze/root-cause",
        "application/json",
        bytes.NewReader(body),
    )
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    fmt.Printf("Analysis Result: %+v\n", result)
}
```

---

## 📊 性能指标

- **平均响应时间**: 500ms - 2s (包含 LLM 调用)
- **并发能力**: 支持数百并发请求
- **内存占用**: ~50MB (空闲)
- **CPU 使用**: 低 (除 LLM 调用外)

---

## 🛠️ 故障排查

### LLM 调用失败

**症状**: 分析结果中没有 `llm_analysis` 字段

**检查**:
1. API Key 是否正确设置
2. 网络连接是否正常
3. 查看日志中的 LLM 错误信息

### 分析置信度低

**症状**: `confidence < 0.7`

**解决**:
1. 启用 LLM 后备: `use_llm_fallback: true`
2. 提供更详细的日志和指标数据
3. 检查事件 reason 是否被支持

---

## 📝 TODO

- [x] **Ollama 本地模型支持** ✅
- [x] **自定义 LLM 服务支持** ✅
- [x] **gollm 统一 LLM 访问层** ✅
- [x] **LangChainGo 架构重构** ✅
- [x] **Memory 系统（对话+向量+案例）** ✅
- [x] **Orchestrator 协调器** ✅
- [ ] 完善 Tool Agent 实现
- [ ] Chroma 向量数据库集成
- [ ] 添加故障预测功能
- [ ] 集成 Neo4j 知识图谱
- [ ] 实现学习系统（反馈收集）
- [ ] 添加异常检测算法
- [ ] 支持流式响应
- [ ] WebSocket 实时分析
- [ ] Prometheus 指标导出
- [ ] 更多 LLM 提供商（Claude, Anthropic）

---

## 📄 许可证

MIT License

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

## 📞 联系

- 项目地址: https://github.com/kart-io/k8s-agent
- 文档: [ARCHITECTURE.md](../../docs/architecture/SYSTEM_ARCHITECTURE.md)
