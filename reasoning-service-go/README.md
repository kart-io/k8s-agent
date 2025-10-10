# Aetherius Reasoning Service (Go Implementation)

**AI-powered root cause analysis service with LLM support**

完整的 Go 语言实现，集成 OpenAI、Google Gemini 和 DeepSeek 大模型，提供智能的 Kubernetes 故障诊断和根因分析。

---

## ✨ 主要特性

### 🤖 LLM 集成
- **多提供商支持**: OpenAI (GPT-4), Google Gemini, DeepSeek, **SiliconFlow**, **Kimi (月之暗面)**, **Ollama (本地部署)**, **自定义 LLM 服务**
- **自定义 LLM**: 支持任何 OpenAI 兼容的 API（vLLM、FastChat、LocalAI 等）
- **智能后备**: 规则引擎 + LLM 混合分析，自动降级
- **优先级配置**: 支持多个 LLM 提供商，按优先级尝试
- **成本优化**: 默认使用 GPT-4o-mini 等高性价比模型
- **私有部署**: 支持 Ollama 本地模型和自建 LLM 服务，数据不出网
- **国内优化**: SiliconFlow 和 Kimi 国内访问速度快

### 🔍 根因分析
- **多模态分析**: 综合事件、日志、指标进行分析
- **规则引擎**: 基于 Kubernetes 最佳实践的模式匹配
- **智能增强**: LLM 提供深度分析和解释

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
    "llm": true
  },
  "timestamp": "2025-10-03T14:00:00Z"
}
```

### 根因分析

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

**响应**:
```json
{
  "request_id": "req-123",
  "status": "completed",
  "result": {
    "root_cause": {
      "type": "OOMKiller",
      "description": "Container was killed due to out of memory (OOM)",
      "confidence": 0.95,
      "evidence": [
        "Event reason: OOMKilled",
        "Found 2 matching patterns in logs",
        "Memory usage at 98.5% (critical)"
      ]
    },
    "recommendations": [
      {
        "action": "increase_memory_limit",
        "description": "Increase container memory limits to prevent OOM kills",
        "confidence": 0.90,
        "risk": "low",
        "impact": "Prevents future OOM kills, may increase cluster resource usage",
        "steps": [
          "Analyze current memory usage patterns",
          "Calculate recommended memory limit (current + 50%)",
          "Update Deployment/StatefulSet memory limits",
          "kubectl apply -f updated-manifest.yaml",
          "Monitor for OOM recurrence"
        ],
        "rollback_steps": [
          "Revert to previous memory limits",
          "kubectl rollout undo deployment/<name>"
        ],
        "estimated_duration": "5 minutes"
      }
    ],
    "confidence": 0.95,
    "evidence": [...],
    "llm_analysis": "{\n  \"root_cause_type\": \"OOMKiller\",\n  \"confidence\": 0.95,\n  ...\n}"
  },
  "processing_time": 1.234
}
```

---

## ⚙️ 配置说明

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
│   ├── analyzer/
│   │   └── root_cause.go        # 根因分析器
│   ├── recommender/
│   │   └── engine.go            # 推荐引擎
│   └── config/
│       └── config.go            # 配置管理
├── pkg/
│   ├── llm/
│   │   ├── interface.go         # LLM 客户端接口
│   │   ├── openai.go            # OpenAI 客户端
│   │   ├── gemini.go            # Gemini 客户端
│   │   ├── deepseek.go          # DeepSeek 客户端
│   │   └── factory.go           # 客户端工厂
│   └── types/
│       └── types.go             # 类型定义
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
