# SiliconFlow 集成指南

## 问题分析

### 原始错误

```text
ERROR: Failed to create internal LLM [error unknown provider: siliconflow]
Failed to create gollm client for siliconflow: failed to create gollm client: failed to create internal LLM: unknown provider: siliconflow
```

### 根本原因

gollm 库 (v0.1.9) 不直接支持 `siliconflow` 作为独立的 provider。gollm 支持的 provider 包括:

- `openai`
- `anthropic`
- `groq`
- `ollama`
- `mistral`
- `openrouter`

SiliconFlow 提供的是 **OpenAI 兼容的 API**,因此应该使用 `openai` provider 并配置自定义的 BaseURL。

## 解决方案

### 1. 代码修复

修改了 `pkg/llm/proxy/adapter.go` 文件,实现了 provider 名称映射机制:

#### 主要修改

1. **添加 provider 映射函数**

```go
func mapToGollmProvider(providerName string) string {
    switch providerName {
    case "openai":
        return "openai"
    case "anthropic", "claude":
        return "anthropic"
    case "groq":
        return "groq"
    case "ollama":
        return "ollama"
    case "mistral":
        return "mistral"
    case "openrouter":
        return "openrouter"
    // OpenAI 兼容的提供商
    case "siliconflow", "deepseek", "kimi", "custom":
        return "openai"
    case "gemini":
        return "openai"
    default:
        log.Printf("Warning: Unknown provider %s, defaulting to openai", providerName)
        return "openai"
    }
}
```

2. **支持自定义 BaseURL**

通过环境变量 `OPENAI_API_BASE` 设置自定义端点:

```go
if cfg.BaseURL != "" && providerName == "openai" && cfg.Name != "openai" {
    os.Setenv("OPENAI_API_BASE", cfg.BaseURL)
    log.Printf("Provider %s (mapped to %s) using custom base URL: %s (via OPENAI_API_BASE)",
        cfg.Name, providerName, cfg.BaseURL)
}
```

### 2. 配置说明

#### 环境变量配置

```bash
# 设置 SiliconFlow API Key
export SILICONFLOW_API_KEY="sk-your-api-key-here"
```

#### 配置文件

配置文件 `configs/config-dev.yaml` 中的 SiliconFlow 配置:

```yaml
llm:
  enabled: true
  providers:
    - name: "siliconflow"
      api_key: ""  # 从环境变量 SILICONFLOW_API_KEY 读取
      base_url: "https://api.siliconflow.cn/v1"
      model: "Qwen/Qwen2.5-7B-Instruct"
      max_tokens: 2048
      temperature: 0.7
      timeout: 15
      priority: 1  # 优先级最高
```

## 使用方法

### 方法一: 直接运行

```bash
# 设置环境变量并运行
SILICONFLOW_API_KEY="sk-your-api-key" go run cmd/server/main.go -c configs/config-dev.yaml
```

### 方法二: 使用 Makefile

```bash
# 在 shell 中导出环境变量
export SILICONFLOW_API_KEY="sk-your-api-key"

# 运行开发服务器
make run-dev
```

### 方法三: 使用 .env 文件

1. 创建 `.env` 文件:

```bash
echo "SILICONFLOW_API_KEY=sk-your-api-key" > .env
```

2. 加载环境变量并运行:

```bash
source .env
make run-dev
```

## 验证集成

### 使用测试脚本

项目提供了完整的测试脚本 `test-siliconflow.sh`:

```bash
# 设置 API Key
export SILICONFLOW_API_KEY="sk-your-api-key"

# 运行测试脚本
./test-siliconflow.sh
```

测试脚本会执行以下检查:

1. ✅ 环境变量检查
2. ✅ API Key 格式验证
3. ✅ 网络连接测试
4. ✅ API 认证测试
5. ✅ 服务启动验证

### 手动验证

#### 1. 检查环境变量

```bash
echo $SILICONFLOW_API_KEY
```

#### 2. 测试 API 连接

```bash
curl -X GET "https://api.siliconflow.cn/v1/models" \
  -H "Authorization: Bearer $SILICONFLOW_API_KEY" \
  -H "Content-Type: application/json"
```

#### 3. 启动服务并查看日志

```bash
SILICONFLOW_API_KEY="$SILICONFLOW_API_KEY" go run cmd/server/main.go -c configs/config-dev.yaml
```

期望看到的日志输出:

```text
[OK] siliconflow (model: Qwen/Qwen2.5-7B-Instruct, priority: 1)
Provider siliconflow (mapped to openai) using custom base URL: https://api.siliconflow.cn/v1
Initialized LLM Proxy Adapter with 1 providers:
  1. siliconflow (priority=1, model=Qwen/Qwen2.5-7B-Instruct, status=ready)
```

#### 4. 测试健康检查端点

```bash
curl http://localhost:8083/health
```

## 支持的其他 OpenAI 兼容提供商

使用相同的方案,以下提供商也可以正常工作:

### DeepSeek

```bash
export DEEPSEEK_API_KEY="your-key"
```

### Kimi (Moonshot)

```bash
export KIMI_API_KEY="your-key"
# 或
export MOONSHOT_API_KEY="your-key"
```

### 自定义 OpenAI 兼容服务

```yaml
- name: "custom"
  api_key: ""  # CUSTOM_LLM_API_KEY
  base_url: "http://your-service:8000/v1"
  model: "your-model"
  priority: 10
```

```bash
export CUSTOM_LLM_API_KEY="your-key"
export CUSTOM_LLM_BASE_URL="http://your-service:8000/v1"
export CUSTOM_LLM_MODEL="your-model"
```

## 常见问题

### Q1: 为什么日志显示 "no API key configured"?

**原因**: 环境变量未正确设置或未传递给 Go 程序。

**解决方案**:

```bash
# 确保环境变量已设置
echo $SILICONFLOW_API_KEY

# 直接在命令中设置
SILICONFLOW_API_KEY="your-key" make run-dev
```

### Q2: 如何验证 API Key 是否有效?

**使用测试脚本**:

```bash
./test-siliconflow.sh
```

**手动测试**:

```bash
curl -X GET "https://api.siliconflow.cn/v1/models" \
  -H "Authorization: Bearer $SILICONFLOW_API_KEY"
```

### Q3: 多个提供商使用相同的 openai provider 会冲突吗?

**回答**: 不会。每个提供商在创建 gollm 客户端时会设置各自的 `OPENAI_API_BASE` 环境变量,并且客户端实例是独立的。

但需要注意:

- 每个 OpenAI 兼容提供商会依次创建客户端
- 最后创建的提供商的 BaseURL 会保留在环境变量中
- 建议按优先级顺序配置提供商

### Q4: 如何查看详细的调试日志?

**设置日志级别为 DEBUG**:

在 `configs/config-dev.yaml` 中:

```yaml
logging:
  level: "DEBUG"
```

或使用环境变量:

```bash
export LOG_LEVEL="DEBUG"
```

### Q5: SiliconFlow 支持哪些模型?

**常用模型**:

- `Qwen/Qwen2.5-7B-Instruct` (默认)
- `Qwen/Qwen2.5-14B-Instruct`
- `Qwen/Qwen2.5-32B-Instruct`
- `deepseek-ai/DeepSeek-V2.5`
- `meta-llama/Llama-3.1-8B-Instruct`

**查询可用模型**:

```bash
curl -X GET "https://api.siliconflow.cn/v1/models" \
  -H "Authorization: Bearer $SILICONFLOW_API_KEY" | jq '.data[].id'
```

## 技术细节

### Provider 映射机制

项目实现了 provider 名称映射机制,将配置文件中的 provider 名称映射到 gollm 支持的 provider:

```text
配置文件中的名称    →    gollm provider    →    实际使用的 API
─────────────────────────────────────────────────────────────
siliconflow         →    openai            →    https://api.siliconflow.cn/v1
deepseek            →    openai            →    https://api.deepseek.com/v1
kimi                →    openai            →    https://api.moonshot.cn/v1
openai              →    openai            →    https://api.openai.com/v1
gemini              →    openai            →    (如果支持 OpenAI 兼容接口)
```

### BaseURL 设置原理

gollm 的 OpenAI provider 通过环境变量 `OPENAI_API_BASE` 读取自定义端点:

1. 代码设置环境变量: `os.Setenv("OPENAI_API_BASE", baseURL)`
2. gollm 创建客户端时读取该环境变量
3. 所有后续请求都发送到自定义端点

### 优先级和故障转移

项目支持配置多个 LLM 提供商,并按优先级(priority)排序:

- **Priority 越小,优先级越高**
- 如果高优先级提供商失败,自动尝试下一个
- 支持同时配置多个 OpenAI 兼容提供商

示例:

```yaml
providers:
  - name: "siliconflow"
    priority: 1  # 最高优先级
  - name: "deepseek"
    priority: 2
  - name: "openai"
    priority: 3
```

## 最佳实践

### 1. 使用环境变量管理 API Key

**推荐**: 将 API Key 存储在环境变量中,不要硬编码在配置文件中。

```bash
# ~/.bashrc 或 ~/.zshrc
export SILICONFLOW_API_KEY="your-key"
export DEEPSEEK_API_KEY="your-key"
export OPENAI_API_KEY="your-key"
```

### 2. 使用 .env 文件(开发环境)

```bash
# .env
SILICONFLOW_API_KEY=sk-xxx
DEEPSEEK_API_KEY=sk-xxx
OPENAI_API_KEY=sk-xxx
```

**加载 .env**:

```bash
source .env
make run-dev
```

### 3. 生产环境配置

**使用 Kubernetes Secrets**:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: llm-api-keys
type: Opaque
stringData:
  SILICONFLOW_API_KEY: "sk-xxx"
  DEEPSEEK_API_KEY: "sk-xxx"
```

**在 Deployment 中引用**:

```yaml
env:
  - name: SILICONFLOW_API_KEY
    valueFrom:
      secretKeyRef:
        name: llm-api-keys
        key: SILICONFLOW_API_KEY
```

### 4. 监控和告警

**记录 LLM 调用指标**:

- 请求成功率
- 响应延迟
- Token 使用量
- 错误率

**设置告警**:

- API Key 即将过期
- 请求失败率超过阈值
- 响应延迟异常

## 参考资料

- [gollm 文档](https://github.com/teilomillet/gollm)
- [SiliconFlow API 文档](https://siliconflow.cn/docs)
- [OpenAI API 兼容规范](https://platform.openai.com/docs/api-reference)

## 更新日志

- **2025-10-20**: 初始版本
  - 实现 provider 名称映射机制
  - 支持 SiliconFlow、DeepSeek、Kimi 等 OpenAI 兼容服务
  - 创建测试脚本和集成文档
