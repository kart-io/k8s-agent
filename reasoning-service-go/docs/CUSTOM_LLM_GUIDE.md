# 自定义 LLM 服务使用指南

本指南介绍如何将 Reasoning Service 连接到自定义的 LLM 服务。

---

## 🎯 适用场景

### 1. 自建 LLM 服务
- 使用 vLLM、FastChat、Text Generation WebUI 等框架部署的模型
- 私有化部署的商业模型
- 内网环境的 LLM 服务

### 2. OpenAI 兼容 API
- 任何实现了 OpenAI Chat Completions API 格式的服务
- 支持 `/v1/chat/completions` 端点
- 标准的请求/响应格式

### 3. 企业私有云
- 企业内部部署的 AI 服务
- 符合数据安全和合规要求
- 不希望数据离开内网环境

---

## 📋 前置要求

### API 格式要求

您的自定义 LLM 服务必须支持 OpenAI 风格的 Chat Completions API：

**请求格式**:
```json
POST /v1/chat/completions
Content-Type: application/json
Authorization: Bearer YOUR_API_KEY

{
  "model": "your-model-name",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Hello!"}
  ],
  "temperature": 0.7,
  "max_tokens": 4096
}
```

**响应格式**:
```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "your-model-name",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "Hello! How can I help you?"
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 10,
    "total_tokens": 20
  }
}
```

---

## 🚀 快速开始

### 方式一：配置文件

编辑 `configs/config.yaml`：

```yaml
llm:
  enabled: true
  providers:
    - name: "custom"
      api_key: "your-api-key-if-needed"  # 如果服务需要认证
      base_url: "http://your-llm-service:8000/v1"
      model: "your-model-name"
      max_tokens: 4096
      temperature: 0.7
      timeout: 60
      priority: 1  # 设置优先级
```

### 方式二：环境变量

```bash
# 必需：服务地址
export CUSTOM_LLM_BASE_URL="http://your-llm-service:8000/v1"

# 必需：模型名称
export CUSTOM_LLM_MODEL="your-model-name"

# 可选：API 密钥（如果服务需要认证）
export CUSTOM_LLM_API_KEY="your-api-key"

# 启动服务
go run cmd/server/main.go
```

### 方式三：混合配置

在配置文件中定义基本配置，通过环境变量覆盖敏感信息：

**configs/config.yaml**:
```yaml
llm:
  enabled: true
  providers:
    - name: "custom"
      api_key: ""  # 通过环境变量设置
      base_url: "http://localhost:8000/v1"  # 可被环境变量覆盖
      model: "llama-2-70b"  # 可被环境变量覆盖
      max_tokens: 4096
      temperature: 0.7
      timeout: 60
      priority: 1
```

**环境变量**:
```bash
export CUSTOM_LLM_API_KEY="secret-key"
export CUSTOM_LLM_BASE_URL="http://production-llm:8000/v1"  # 覆盖配置
```

---

## 🔧 常见 LLM 框架配置示例

### 1. vLLM

**部署 vLLM**:
```bash
# 安装 vLLM
pip install vllm

# 启动 OpenAI 兼容服务器
python -m vllm.entrypoints.openai.api_server \
  --model meta-llama/Llama-2-70b-chat-hf \
  --host 0.0.0.0 \
  --port 8000
```

**Reasoning Service 配置**:
```yaml
llm:
  providers:
    - name: "custom"
      api_key: ""  # vLLM 默认不需要认证
      base_url: "http://localhost:8000/v1"
      model: "meta-llama/Llama-2-70b-chat-hf"
      max_tokens: 4096
      temperature: 0.7
      timeout: 60
      priority: 1
```

### 2. FastChat

**部署 FastChat**:
```bash
# 安装 FastChat
pip install "fschat[model_worker,webui]"

# 启动控制器
python -m fastchat.serve.controller

# 启动模型工作器
python -m fastchat.serve.model_worker --model-path lmsys/vicuna-7b-v1.5

# 启动 OpenAI 兼容 API 服务器
python -m fastchat.serve.openai_api_server --host 0.0.0.0 --port 8000
```

**Reasoning Service 配置**:
```yaml
llm:
  providers:
    - name: "custom"
      api_key: ""
      base_url: "http://localhost:8000/v1"
      model: "vicuna-7b-v1.5"
      max_tokens: 2048
      temperature: 0.7
      timeout: 60
      priority: 1
```

### 3. Text Generation WebUI

**部署 oobabooga/text-generation-webui**:
```bash
# 克隆仓库
git clone https://github.com/oobabooga/text-generation-webui
cd text-generation-webui

# 安装依赖
./start_linux.sh  # 或 start_macos.sh, start_windows.bat

# 启动，开启 OpenAI API 扩展
python server.py --api --extensions openai
```

**Reasoning Service 配置**:
```yaml
llm:
  providers:
    - name: "custom"
      api_key: ""
      base_url: "http://localhost:5000/v1"
      model: "your-loaded-model"
      max_tokens: 4096
      temperature: 0.7
      timeout: 60
      priority: 1
```

### 4. LocalAI

**部署 LocalAI**:
```bash
# Docker 启动
docker run -p 8080:8080 --name localai -ti localai/localai:latest

# 或使用 docker-compose
curl -O https://raw.githubusercontent.com/mudler/LocalAI/master/docker-compose.yaml
docker-compose up -d
```

**Reasoning Service 配置**:
```yaml
llm:
  providers:
    - name: "custom"
      api_key: ""
      base_url: "http://localhost:8080/v1"
      model: "gpt-3.5-turbo"  # LocalAI 的模型别名
      max_tokens: 4096
      temperature: 0.7
      timeout: 60
      priority: 1
```

### 5. LM Studio

**LM Studio 配置**:
1. 下载并启动 LM Studio
2. 在设置中开启 "Local Server"
3. 默认端口：1234

**Reasoning Service 配置**:
```yaml
llm:
  providers:
    - name: "custom"
      api_key: ""
      base_url: "http://localhost:1234/v1"
      model: "local-model"  # LM Studio 自动处理
      max_tokens: 4096
      temperature: 0.7
      timeout: 60
      priority: 1
```

---

## 🔐 认证方式

### 1. Bearer Token 认证（推荐）

大多数 OpenAI 兼容 API 使用 Bearer Token：

```yaml
llm:
  providers:
    - name: "custom"
      api_key: "your-secret-token"
      base_url: "http://your-service/v1"
      model: "your-model"
```

服务会自动添加 HTTP 头：
```
Authorization: Bearer your-secret-token
```

### 2. 无认证

如果您的服务在内网运行且不需要认证：

```yaml
llm:
  providers:
    - name: "custom"
      api_key: ""  # 留空
      base_url: "http://internal-llm:8000/v1"
      model: "your-model"
```

### 3. 自定义认证头

如果需要其他认证方式，可以修改 `pkg/llm/custom.go`：

```go
// 在 Complete 方法中添加自定义头
httpReq.Header.Set("X-Custom-Auth", "your-auth-value")
```

---

## 🧪 测试配置

### 1. 检查服务可用性

```bash
# 测试自定义 LLM 端点是否可访问
curl -X POST http://your-llm-service:8000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key" \
  -d '{
    "model": "your-model-name",
    "messages": [{"role": "user", "content": "Hello"}],
    "max_tokens": 10
  }'
```

### 2. 验证 Reasoning Service 集成

```bash
# 启动 Reasoning Service
go run cmd/server/main.go

# 检查健康状态
curl http://localhost:8082/health

# 测试根因分析（使用自定义 LLM）
curl -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "custom-llm-test-001",
    "context": {
      "event": {
        "reason": "OOMKilled",
        "message": "容器因内存不足被杀死"
      },
      "logs": "fatal error: out of memory"
    },
    "options": {
      "use_llm": true,
      "llm_provider": "custom"
    }
  }'
```

---

## 🎨 高级配置

### 多个自定义 LLM 服务

您可以配置多个自定义 LLM 服务作为备份：

```yaml
llm:
  enabled: true
  providers:
    # 主力自定义 LLM
    - name: "custom"
      api_key: ""
      base_url: "http://primary-llm:8000/v1"
      model: "llama-2-70b"
      max_tokens: 4096
      priority: 1

    # 备用 Ollama 本地服务
    - name: "ollama"
      api_key: ""
      base_url: "http://localhost:11434/v1"
      model: "llama3.1"
      max_tokens: 4096
      priority: 2

    # 云端备份
    - name: "openai"
      api_key: "${OPENAI_API_KEY}"
      model: "gpt-4o-mini"
      priority: 3
```

### 针对不同场景使用不同模型

为不同的分析类型配置不同的自定义 LLM：

```yaml
llm:
  providers:
    # 快速响应模型（一般分析）
    - name: "custom"
      base_url: "http://fast-llm:8000/v1"
      model: "llama-2-7b"
      priority: 1

    # 高质量模型（复杂分析）
    - name: "custom"
      base_url: "http://powerful-llm:8000/v1"
      model: "llama-2-70b"
      priority: 2
```

注意：当前配置中，多个同名 `custom` 提供商需要在代码中区分。建议使用不同的 `base_url` 或在请求时指定 provider。

### 性能优化

```yaml
llm:
  providers:
    - name: "custom"
      base_url: "http://your-llm:8000/v1"
      model: "your-model"
      max_tokens: 2048  # 减少最大令牌数以提高速度
      temperature: 0.3  # 降低温度以提高确定性
      timeout: 120      # 增加超时时间（大模型推理较慢）

performance:
  max_context_size: 5000  # 限制上下文大小
  request_timeout: "60s"
```

---

## 📊 性能对比

| LLM 框架 | 部署难度 | 推理速度 | 资源占用 | 适用场景 |
|---------|---------|---------|---------|---------|
| **vLLM** | 中等 | ⚡⚡⚡ | GPU 高 | 生产环境，高并发 |
| **FastChat** | 中等 | ⚡⚡ | GPU 高 | 多模型服务 |
| **Text Generation WebUI** | 简单 | ⚡ | GPU 中 | 开发测试，单用户 |
| **LocalAI** | 简单 | ⚡⚡ | CPU/GPU 中 | CPU 推理，多后端 |
| **LM Studio** | 非常简单 | ⚡⚡ | CPU/GPU 低 | 桌面开发，快速测试 |
| **Ollama** | 非常简单 | ⚡⚡⚡ | CPU/GPU 低 | 本地部署首选 |

---

## ❓ 常见问题

### Q1: 如何判断自定义 LLM 服务是否兼容？

**A**: 使用 curl 测试：

```bash
curl -X POST http://your-service/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "test",
    "messages": [{"role": "user", "content": "ping"}]
  }'
```

如果返回类似 OpenAI 的响应格式，则兼容。

### Q2: 连接失败怎么办？

**A**: 检查以下几点：
1. 服务是否启动：`curl http://your-service:8000/health`
2. 网络是否可达：`telnet your-service 8000`
3. 端口是否正确：检查 `base_url` 配置
4. 防火墙规则是否允许连接
5. 查看 Reasoning Service 日志获取详细错误

### Q3: 认证失败怎么办？

**A**:
1. 确认 API Key 是否正确设置
2. 检查自定义服务是否需要认证
3. 尝试使用 curl 直接测试认证
4. 查看服务日志确认认证方式

### Q4: 推理速度太慢？

**A**: 优化建议：
1. 减少 `max_tokens` 设置
2. 降低 `temperature`（减少采样时间）
3. 使用更小的模型
4. 启用 GPU 加速
5. 使用 vLLM 等优化框架
6. 增加 `timeout` 设置以避免超时

### Q5: 如何在 Docker 中使用自定义 LLM？

**A**: 使用 Docker 网络连接：

```yaml
# docker-compose.yaml
version: '3'
services:
  llm-service:
    image: vllm/vllm-openai:latest
    ports:
      - "8000:8000"
    environment:
      - MODEL_NAME=meta-llama/Llama-2-7b-chat-hf

  reasoning-service:
    build: .
    ports:
      - "8082:8082"
    environment:
      - CUSTOM_LLM_BASE_URL=http://llm-service:8000/v1
      - CUSTOM_LLM_MODEL=meta-llama/Llama-2-7b-chat-hf
    depends_on:
      - llm-service
```

### Q6: 支持流式响应吗？

**A**: 当前版本不支持流式响应（`stream: false`）。流式响应支持计划在未来版本中添加。

---

## 🔄 与其他提供商混合使用

### 混合部署示例

结合自定义 LLM 和云服务，实现成本优化和高可用：

```yaml
llm:
  enabled: true
  providers:
    # 首选：本地自定义 LLM（免费，快速）
    - name: "custom"
      base_url: "http://localhost:8000/v1"
      model: "llama-2-70b"
      priority: 1

    # 备选：Ollama（免费备份）
    - name: "ollama"
      base_url: "http://localhost:11434/v1"
      model: "llama3.1"
      priority: 2

    # 兜底：云端 LLM（付费，高质量）
    - name: "openai"
      api_key: "${OPENAI_API_KEY}"
      model: "gpt-4o-mini"
      priority: 3

analysis:
  use_llm_fallback: true  # 启用自动降级
```

**工作流程**:
1. 优先使用本地自定义 LLM（快速、免费）
2. 如果本地服务不可用，降级到 Ollama
3. 如果本地服务都失败，使用云端 OpenAI（确保服务可用性）

---

## 📝 总结

### 推荐配置

**企业私有云**:
```yaml
providers:
  - name: "custom"
    base_url: "http://internal-llm.company.com/v1"
    model: "enterprise-model"
    priority: 1
```

**开发测试**:
```yaml
providers:
  - name: "custom"
    base_url: "http://localhost:1234/v1"  # LM Studio
    model: "local-model"
    priority: 1
```

**生产环境**:
```yaml
providers:
  - name: "custom"
    base_url: "http://vllm-cluster:8000/v1"
    model: "llama-2-70b"
    priority: 1
  - name: "openai"
    api_key: "${OPENAI_API_KEY}"
    priority: 2
```

---

## 🔗 参考资源

### LLM 框架文档
- **vLLM**: https://docs.vllm.ai/
- **FastChat**: https://github.com/lm-sys/FastChat
- **Text Generation WebUI**: https://github.com/oobabooga/text-generation-webui
- **LocalAI**: https://localai.io/
- **LM Studio**: https://lmstudio.ai/

### OpenAI API 规范
- **Chat Completions API**: https://platform.openai.com/docs/api-reference/chat

---

**开始使用自定义 LLM，构建完全可控的 AI 诊断服务！** 🚀
