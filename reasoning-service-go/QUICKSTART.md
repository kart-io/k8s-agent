# 快速启动指南

## 📦 安装和配置

### 步骤 1: 克隆或进入项目目录

```bash
cd /Users/costalong/code/go/src/github.com/kart/k8s-agent/reasoning-service-go
```

### 步骤 2: 设置 LLM 提供商

**选项 A: 使用 Ollama（推荐 - 免费且私有）**

```bash
# 1. 安装 Ollama
curl -fsSL https://ollama.com/install.sh | sh

# 2. 下载模型
ollama pull llama3.1

# 3. 验证 Ollama 运行
curl http://localhost:11434/api/tags

# 完成！无需 API Key，配置文件已包含 Ollama 设置
```

**选项 B: 使用云端 LLM（需要 API Key）**

```bash
# 方式 1: 导出环境变量
export OPENAI_API_KEY="sk-your-key-here"
# 或
export GEMINI_API_KEY="your-gemini-key"
# 或
export DEEPSEEK_API_KEY="your-deepseek-key"

# 方式 2: 创建 .env 文件（不推荐提交到 Git）
cp .env.example .env
# 编辑 .env 文件，填入你的 API Key
```

**推荐**: 先使用 Ollama 本地体验，无需任何费用和网络依赖！

### 步骤 3: 安装依赖

```bash
go mod download
```

### 步骤 4: 运行服务

```bash
# 直接运行
go run cmd/server/main.go

# 或使用 Make
make run
```

**输出示例（使用 Ollama）**:
```
Aetherius Reasoning Service (Go)
=================================
Config: configs/config.yaml
Server: 0.0.0.0:8082

Initializing LLM providers:
  [SKIP] openai: No API key
  [SKIP] gemini: No API key
  [SKIP] deepseek: No API key
  [OK] ollama (model: llama3.1, priority: 4)

Starting server...
Starting Reasoning Service on 0.0.0.0:8082
```

**输出示例（使用 OpenAI）**:
```
Initializing LLM providers:
  [OK] openai (model: gpt-4o-mini, priority: 1)
  [SKIP] gemini: No API key
  [SKIP] deepseek: No API key
  [OK] ollama (model: llama3.1, priority: 4)
```

---

## 🧪 测试服务

### 测试 1: 健康检查

```bash
curl http://localhost:8082/health
```

**预期响应**:
```json
{
  "status": "healthy",
  "service": "reasoning-service-go",
  "components": {
    "analyzer": true,
    "recommender": true,
    "llm": true
  }
}
```

### 测试 2: 不使用 LLM 的规则引擎分析

```bash
curl -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "test-001",
    "context": {
      "event": {
        "reason": "OOMKilled"
      },
      "logs": "fatal error: out of memory",
      "metrics": {
        "memory": {
          "usage_percent": 98
        }
      }
    },
    "options": {
      "use_llm": false
    }
  }'
```

### 测试 3: 使用 LLM 的智能分析

```bash
curl -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "test-002",
    "context": {
      "event": {
        "reason": "CrashLoopBackOff"
      },
      "logs": "Error: Cannot find module ./config"
    },
    "options": {
      "use_llm": true
    }
  }'
```

### 测试 4: 运行完整测试脚本

```bash
chmod +x examples/test_request.sh
./examples/test_request.sh
```

---

## 🏗️ 构建和部署

### 构建二进制文件

```bash
make build
# 或
go build -o bin/reasoning-service cmd/server/main.go
```

### 运行二进制文件

```bash
./bin/reasoning-service -config configs/config.yaml
```

### Docker 部署

```bash
# 构建镜像
make docker-build

# 运行容器（设置 API Key）
docker run -d \
  -p 8082:8082 \
  -e OPENAI_API_KEY="sk-your-key" \
  --name reasoning-service \
  reasoning-service-go:latest

# 查看日志
docker logs -f reasoning-service

# 停止容器
docker stop reasoning-service
docker rm reasoning-service
```

---

## 🎯 使用场景示例

### 场景 1: OOM 分析

**问题**: Pod 因内存不足被 Kill

```bash
curl -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "oom-001",
    "context": {
      "event": {
        "reason": "OOMKilled",
        "message": "Container killed due to OOM"
      },
      "logs": "java.lang.OutOfMemoryError: Java heap space",
      "metrics": {
        "memory": {
          "usage_percent": 99,
          "usage_bytes": 2097000000,
          "limit_bytes": 2097152000
        }
      }
    },
    "options": {
      "use_llm": true,
      "max_recommendations": 3
    }
  }'
```

**结果**:
- 根因: OOMKiller
- 推荐: 增加内存限制、检查内存泄漏

### 场景 2: 镜像拉取失败

**问题**: 镜像无法从仓库拉取

```bash
curl -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "img-001",
    "context": {
      "event": {
        "reason": "ImagePullBackOff",
        "message": "Failed to pull image: unauthorized"
      },
      "logs": "Error response from daemon: pull access denied"
    },
    "options": {
      "use_llm": false
    }
  }'
```

**结果**:
- 根因: ImagePullError
- 推荐: 检查镜像访问权限、验证 imagePullSecrets

### 场景 3: 配置错误（使用 LLM）

**问题**: 应用启动失败，原因不明

```bash
curl -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "config-001",
    "context": {
      "event": {
        "reason": "CrashLoopBackOff"
      },
      "logs": "Error: ENOENT: no such file or directory, open '\'''/app/data/config.json'\''"
    },
    "options": {
      "use_llm": true,
      "llm_provider": "openai"
    }
  }'
```

**结果**:
- 根因: ConfigError (由 LLM 分析得出)
- LLM 分析: 缺少配置文件或挂载路径错误
- 推荐: 检查 ConfigMap/Secret 挂载、验证文件路径

### 场景 4: 使用 Ollama 本地分析

**问题**: 希望使用本地模型，保护数据隐私

```bash
# 1. 确保 Ollama 运行并已下载模型
ollama list

# 2. 使用 Ollama 进行分析
curl -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "ollama-001",
    "context": {
      "event": {
        "reason": "OOMKilled"
      },
      "logs": "java.lang.OutOfMemoryError: GC overhead limit exceeded",
      "metrics": {
        "memory": {
          "usage_percent": 98
        }
      }
    },
    "options": {
      "use_llm": true,
      "llm_provider": "ollama"
    }
  }'
```

**优势**:
- ✅ 完全本地推理，数据不出网
- ✅ 无 API 调用费用
- ✅ 离线可用
- ✅ 可自定义模型

---

## ⚙️ 配置调整

### 调整分析参数

编辑 `configs/config.yaml`:

```yaml
analysis:
  min_confidence: 0.6  # 降低置信度阈值（更宽松）
  max_recommendations: 3  # 只返回前 3 个推荐
  use_llm_fallback: true  # 规则引擎置信度低时自动使用 LLM
```

### 配置多个 LLM 提供商（故障转移）

```yaml
llm:
  enabled: true
  providers:
    - name: "openai"
      priority: 1  # 优先使用
      model: "gpt-4o-mini"

    - name: "gemini"
      priority: 2  # OpenAI 失败时使用
      model: "gemini-1.5-flash"

    - name: "deepseek"
      priority: 3  # 最后备选
      model: "deepseek-chat"
```

---

## 🔍 查看和监控

### 查看日志

```bash
# 如果配置了文件日志
tail -f logs/reasoning-service.log

# Docker 容器日志
docker logs -f reasoning-service
```

### 监控请求

每个请求都会输出到控制台:
```
POST /api/v1/analyze/root-cause 127.0.0.1:52341 1.234s
```

---

## ❓ 常见问题

### Q: LLM 调用失败怎么办？

**A**: 检查以下几点:
1. API Key 是否正确: `echo $OPENAI_API_KEY`
2. 网络连接: `curl https://api.openai.com`
3. 查看错误日志

### Q: 如何只使用规则引擎，不用 LLM？

**A**: 在请求中设置:
```json
{
  "options": {
    "use_llm": false
  }
}
```

或在配置中关闭:
```yaml
llm:
  enabled: false
```

### Q: 如何选择特定的 LLM 提供商？

**A**: 在请求中指定:
```json
{
  "options": {
    "use_llm": true,
    "llm_provider": "gemini"  // 或 "openai", "deepseek"
  }
}
```

### Q: 服务响应很慢怎么办？

**A**:
1. LLM 调用通常需要 1-3 秒，这是正常的
2. 可以先用规则引擎（`use_llm: false`）获得快速结果
3. 配置超时时间: `performance.request_timeout: "10s"`

---

## 📚 下一步

- 查看完整 [README.md](README.md) 了解更多功能
- 阅读 [API 文档](#) 了解所有端点
- 集成到 orchestrator-service 进行自动化分析
- 添加自定义规则和推荐

---

## 💡 提示

1. **成本优化**: 默认使用 `gpt-4o-mini` 而不是 `gpt-4`，可节省 90% 成本
2. **混合策略**: 启用 `use_llm_fallback`，规则引擎先行，LLM 辅助
3. **多提供商**: 配置多个 LLM 提供商实现故障转移
4. **日志分析**: LLM 对复杂日志的分析效果最佳

---

**祝您使用愉快！** 🚀
