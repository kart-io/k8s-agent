# Ollama 集成指南

Ollama 是一个本地运行大模型的工具，无需 API Key，完全私有化部署，非常适合企业内网环境。

---

## 🚀 快速开始

### 步骤 1: 安装 Ollama

#### macOS / Linux
```bash
curl -fsSL https://ollama.com/install.sh | sh
```

#### Windows
下载安装包: https://ollama.com/download/windows

#### Docker
```bash
docker run -d -v ollama:/root/.ollama -p 11434:11434 --name ollama ollama/ollama
```

### 步骤 2: 下载模型

```bash
# Llama 3.1 (推荐 - 8B 参数)
ollama pull llama3.1

# Llama 3.2 (最新版本)
ollama pull llama3.2

# Qwen 2.5 (中文优秀)
ollama pull qwen2.5

# Mistral (小巧高效)
ollama pull mistral

# DeepSeek Coder (代码专用)
ollama pull deepseek-coder
```

### 步骤 3: 验证 Ollama 运行

```bash
# 检查服务状态
curl http://localhost:11434/api/tags

# 测试模型
ollama run llama3.1 "What is Kubernetes?"
```

### 步骤 4: 配置 Reasoning Service

编辑 `configs/config.yaml`:

```yaml
llm:
  enabled: true
  providers:
    - name: "ollama"
      base_url: "http://localhost:11434/v1"
      model: "llama3.1"  # 或其他已下载的模型
      max_tokens: 4096
      temperature: 0.7
      timeout: 60
      priority: 1  # 设为最高优先级
```

### 步骤 5: 启动服务

```bash
cd reasoning-service-go
go run cmd/server/main.go
```

**输出**:
```
Initializing LLM providers:
  [OK] ollama (model: llama3.1, priority: 1)

Starting server...
```

---

## 🎯 使用示例

### 示例 1: 分析 OOM 问题

```bash
curl -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "ollama-test-001",
    "context": {
      "event": {
        "reason": "OOMKilled",
        "message": "Container killed due to OOM"
      },
      "logs": "fatal error: runtime: out of memory\ngoroutine 1: exceeded memory limit",
      "metrics": {
        "memory": {
          "usage_percent": 99.2
        }
      }
    },
    "options": {
      "use_llm": true,
      "llm_provider": "ollama"
    }
  }'
```

### 示例 2: 分析配置错误

```bash
curl -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "ollama-test-002",
    "context": {
      "event": {
        "reason": "CrashLoopBackOff"
      },
      "logs": "Error: Cannot find module '\''./config'\''\nRequire stack:\n- /app/index.js"
    },
    "options": {
      "use_llm": true
    }
  }'
```

---

## 📊 推荐模型对比

| 模型 | 参数量 | 内存占用 | 速度 | 中文能力 | 推荐场景 |
|------|--------|---------|------|---------|---------|
| **llama3.1** | 8B | ~5GB | 快 | 良好 | 通用分析，推荐首选 |
| **llama3.2** | 3B | ~2GB | 很快 | 良好 | 资源受限环境 |
| **qwen2.5** | 7B | ~4.5GB | 快 | **优秀** | 中文为主的环境 |
| **mistral** | 7B | ~4GB | 快 | 中等 | 快速推理 |
| **deepseek-coder** | 7B | ~4.5GB | 中 | 良好 | 代码分析 |

---

## ⚙️ 高级配置

### 1. GPU 加速

Ollama 自动检测并使用 GPU（NVIDIA、AMD、Apple Silicon）。

**检查 GPU 使用**:
```bash
# macOS/Linux
ollama ps

# 查看资源使用
watch -n 1 'ps aux | grep ollama'
```

### 2. 多模型配置

可以配置多个 Ollama 模型作为不同优先级：

```yaml
llm:
  providers:
    - name: "ollama"
      model: "llama3.1"
      priority: 1

    - name: "ollama"
      model: "qwen2.5"
      priority: 2

    # 云端 API 作为备份
    - name: "openai"
      model: "gpt-4o-mini"
      priority: 3
```

### 3. 自定义模型参数

**创建自定义 Modelfile**:

```dockerfile
# Modelfile
FROM llama3.1

# 针对 Kubernetes 优化的系统提示词
SYSTEM """
You are an expert Kubernetes troubleshooting assistant specializing in:
- Pod failures and crash analysis
- Resource management (CPU, Memory, Disk)
- Network connectivity issues
- Configuration errors
- Container image problems

Always provide actionable, step-by-step solutions.
"""

# 参数调优
PARAMETER temperature 0.5
PARAMETER top_p 0.9
PARAMETER num_ctx 4096
```

**创建自定义模型**:
```bash
ollama create k8s-expert -f Modelfile
```

**使用自定义模型**:
```yaml
llm:
  providers:
    - name: "ollama"
      model: "k8s-expert"
```

### 4. 远程 Ollama 服务器

如果 Ollama 运行在其他机器：

```yaml
llm:
  providers:
    - name: "ollama"
      base_url: "http://192.168.1.100:11434/v1"
      model: "llama3.1"
```

---

## 🐳 Docker Compose 部署

**docker-compose.yml**:

```yaml
version: '3.8'

services:
  ollama:
    image: ollama/ollama:latest
    container_name: ollama
    ports:
      - "11434:11434"
    volumes:
      - ollama_data:/root/.ollama
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]

  reasoning-service:
    image: reasoning-service-go:latest
    container_name: reasoning-service
    ports:
      - "8082:8082"
    environment:
      - OLLAMA_BASE_URL=http://ollama:11434/v1
    depends_on:
      - ollama
    volumes:
      - ./configs:/app/configs

volumes:
  ollama_data:
```

**启动**:
```bash
docker-compose up -d

# 下载模型到 Ollama 容器
docker exec ollama ollama pull llama3.1

# 查看日志
docker-compose logs -f reasoning-service
```

---

## 🔧 性能优化

### 1. 并发限制

Ollama 默认同时处理 1 个请求，大模型推理需要时间。

**调整并发**:
```bash
# 设置环境变量
export OLLAMA_NUM_PARALLEL=2
ollama serve
```

### 2. 上下文长度

```yaml
llm:
  providers:
    - name: "ollama"
      model: "llama3.1"
      max_tokens: 2048  # 减少输出长度加快速度
```

### 3. 缓存优化

```bash
# 预加载模型到内存
ollama run llama3.1 ""

# 保持模型在内存中
export OLLAMA_KEEP_ALIVE=24h
```

---

## 📈 监控和调试

### 查看 Ollama 日志

```bash
# macOS
tail -f ~/Library/Logs/Ollama/server.log

# Linux
journalctl -u ollama -f

# Docker
docker logs -f ollama
```

### 性能监控

```bash
# 查看内存使用
ps aux | grep ollama

# 查看 GPU 使用（NVIDIA）
nvidia-smi -l 1

# 查看 API 状态
curl http://localhost:11434/api/tags
```

### 测试推理速度

```bash
time curl -X POST http://localhost:11434/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3.1",
    "messages": [{"role": "user", "content": "What is Kubernetes?"}],
    "stream": false
  }'
```

---

## 🆚 Ollama vs 云端 API

| 特性 | Ollama | OpenAI/Gemini/DeepSeek |
|------|--------|------------------------|
| **成本** | 免费（硬件成本） | 按 Token 付费 |
| **隐私** | 完全本地，数据不出网 | 数据发送到云端 |
| **速度** | 取决于本地硬件 | 通常较快（云端优化） |
| **依赖** | 无需互联网 | 需要稳定网络 |
| **模型选择** | 开源模型 | 商业模型（更强大） |
| **部署** | 需要本地资源 | 即开即用 |

**推荐策略**: **混合部署**

```yaml
# 本地优先，云端备份
llm:
  providers:
    - name: "ollama"
      model: "llama3.1"
      priority: 1  # 首选

    - name: "openai"
      model: "gpt-4o-mini"
      priority: 2  # Ollama 失败时使用
```

---

## ❓ 常见问题

### Q1: Ollama 启动失败

**检查**:
```bash
# 检查端口占用
lsof -i :11434

# 重启服务
ollama serve
```

### Q2: 模型推理很慢

**优化**:
1. 使用 GPU: 确保 NVIDIA/AMD 驱动正确
2. 使用更小的模型: `llama3.2` (3B) 而不是 `llama3.1` (8B)
3. 减少 `max_tokens`

### Q3: 内存不足

**解决**:
```bash
# 使用量化模型（更小）
ollama pull llama3.1:7b-q4_0  # 4-bit 量化
```

### Q4: Ollama 和 Reasoning Service 连接失败

**检查**:
```bash
# 测试 Ollama API
curl http://localhost:11434/v1/models

# 检查防火墙
sudo ufw allow 11434
```

---

## 📚 参考资源

- **Ollama 官网**: https://ollama.com
- **模型库**: https://ollama.com/library
- **GitHub**: https://github.com/ollama/ollama
- **文档**: https://github.com/ollama/ollama/tree/main/docs

---

## 🎉 总结

使用 Ollama 的优势：

✅ **完全免费** - 无 API 调用费用
✅ **数据隐私** - 敏感信息不离开本地
✅ **离线可用** - 无需互联网连接
✅ **可定制** - 可以微调和定制模型
✅ **快速部署** - 一行命令安装启动

**推荐用于**:
- 企业内网环境
- 敏感数据分析
- 成本敏感场景
- 离线/边缘部署

---

**开始使用 Ollama，享受本地 AI 的强大能力！** 🚀
