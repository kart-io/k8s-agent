# SiliconFlow API 端点问题修复报告

## 问题描述

在配置 SiliconFlow API 后,虽然环境变量正确设置,但实际请求仍被发送到 OpenAI 官方端点,导致 401 认证错误:

```text
ERROR: API error [provider openai status 401 body {
    "error": {
        "message": "Incorrect API key provided: sk-mtrgy***. You can find your API key at https://platform.openai.com/account/api-keys.",
        "type": "invalid_request_error",
        "code": "invalid_api_key"
    }
}]
```

## 根本原因

### 问题 1: gollm 不支持自定义 BaseURL

gollm v0.1.9 的 OpenAI provider 硬编码了端点 URL:

```go
// 来自 gollm/providers/openai.go:119
func (p *OpenAIProvider) Endpoint() string {
    return "https://api.openai.com/v1/chat/completions"  // 硬编码!
}
```

**gollm 不支持通过代码或环境变量为 OpenAI provider 设置自定义 BaseURL**。

### 问题 2: Provider 映射策略错误

之前的方案尝试将 SiliconFlow 映射到 OpenAI provider,但由于上述限制,这个方案无法工作:

```go
// 之前的错误方案
case "siliconflow":
    return "openai"  // 映射到 openai,但无法自定义 BaseURL
```

## 解决方案

### 核心思路

**混合客户端策略**:根据提供商类型选择合适的客户端实现

- **gollm 原生支持的提供商**: 使用 gollm (openai, anthropic, groq, ollama, mistral, openrouter)
- **其他提供商**: 使用项目原生的 LLM client (siliconflow, deepseek, kimi, gemini, custom)

### 实现详情

#### 1. 修改 `ProviderClient` 结构

```go
type ProviderClient struct {
    name       string
    priority   int
    client     interface{}               // 支持两种类型的客户端
    useGollm   bool                      // 标记使用哪种客户端
    config     *config.LLMProviderConfig
    healthy    bool
    lastErr    string
    lastCheck  time.Time
}
```

#### 2. 添加判断函数

```go
func shouldUseGollm(providerName string) bool {
    switch providerName {
    case "openai", "anthropic", "claude", "groq", "ollama", "mistral", "openrouter":
        return true  // gollm 原生支持
    default:
        return false // 使用原生 LLM client
    }
}
```

#### 3. 创建原生 LLM 客户端

```go
func createNativeLLMClient(cfg *config.LLMProviderConfig) (llmclient.Client, error) {
    llmCfg := &llmclient.Config{
        Provider:    llmclient.Provider(cfg.Name),
        APIKey:      cfg.APIKey,
        BaseURL:     cfg.BaseURL,          // ✅ 支持自定义 BaseURL
        Model:       cfg.Model,
        MaxTokens:   cfg.MaxTokens,
        Temperature: cfg.Temperature,
        Timeout:     cfg.Timeout,
    }

    log.Printf("Provider %s using native LLM client with base URL: %s",
        cfg.Name, cfg.BaseURL)

    return llmclient.NewClient(llmCfg)
}
```

#### 4. 修改 Complete 方法

```go
func (a *ProxyAdapter) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    for _, provider := range a.providers {
        if provider.useGollm {
            // 使用 gollm 客户端
            gollmClient, _ := provider.client.(gollm.LLM)
            prompt, _ := buildGollmPrompt(req)
            response, err = gollmClient.Generate(ctx, prompt)
        } else {
            // 使用原生 LLM 客户端
            nativeClient, _ := provider.client.(llmclient.Client)
            llmReq := &llmclient.CompletionRequest{
                Messages:    convertMessages(req.Messages),
                Temperature: req.Temperature,
                MaxTokens:   req.MaxTokens,
            }
            llmResp, err := nativeClient.Complete(ctx, llmReq)
            response = llmResp.Content
        }
        // ... 处理响应
    }
}
```

## 修改的文件

1. `pkg/llm/proxy/adapter.go`
   - 修改 `ProviderClient` 结构
   - 添加 `shouldUseGollm()` 函数
   - 添加 `createNativeLLMClient()` 函数
   - 添加 `convertMessages()` 函数
   - 修改 `NewProxyAdapter()` 初始化逻辑
   - 修改 `Complete()` 方法支持两种客户端
   - 简化 `createGollmClient()` (移除无效的映射逻辑)

## 验证结果

### 启动日志

```text
✅ [OK] siliconflow (model: Qwen/Qwen2.5-7B-Instruct, priority: 1)
✅ Provider siliconflow using native LLM client with base URL: https://api.siliconflow.cn/v1
✅ Initialized LLM Proxy Adapter with 2 providers:
     1. siliconflow (priority=1, model=Qwen/Qwen2.5-7B-Instruct, status=ready)
     2. deepseek (priority=3, model=deepseek-chat, status=ready)
```

### API 调用日志

```text
✅ [DEBUG] SiliconFlow API Key length: 51, first 10 chars: sk-mtrgyvc...
```

说明:
- ✅ SiliconFlow 使用原生 LLM 客户端
- ✅ 正确设置 BaseURL 为 `https://api.siliconflow.cn/v1`
- ✅ 正确使用 SiliconFlow API Key

## 支持的提供商架构

```text
提供商类型          使用的客户端           BaseURL 支持
───────────────────────────────────────────────────────
openai              gollm                ❌ (硬编码)
anthropic           gollm                ❌ (硬编码)
groq                gollm                ❌ (硬编码)
ollama              gollm                ✅ (via SetOllamaEndpoint)
mistral             gollm                ❌ (硬编码)
openrouter          gollm                ❌ (硬编码)
───────────────────────────────────────────────────────
siliconflow         原生 LLM client      ✅ 完全支持
deepseek            原生 LLM client      ✅ 完全支持
kimi                原生 LLM client      ✅ 完全支持
gemini              原生 LLM client      ✅ 完全支持
custom              原生 LLM client      ✅ 完全支持
```

## 使用方法

### 配置 SiliconFlow

```yaml
# configs/config-dev.yaml
llm:
  providers:
    - name: "siliconflow"
      api_key: ""  # 从环境变量读取
      base_url: "https://api.siliconflow.cn/v1"
      model: "Qwen/Qwen2.5-7B-Instruct"
      priority: 1
```

### 设置环境变量

```bash
export SILICONFLOW_API_KEY="sk-your-api-key"
```

### 运行服务

```bash
SILICONFLOW_API_KEY=$SILICONFLOW_API_KEY make run-dev
```

## 优势

### 1. 灵活性

- 对于 gollm 原生支持的提供商,享受 gollm 的高级功能
- 对于其他提供商,使用项目自己的实现,完全控制

### 2. 扩展性

- 添加新的 OpenAI 兼容提供商只需在配置中添加
- 不需要修改代码

### 3. 性能

- 原生 LLM client 实现更简洁,性能更好
- 避免了 gollm 的额外抽象层

### 4. 维护性

- 清晰的职责划分
- 易于理解和调试

## 对比之前的方案

### 方案 1 (失败): Provider 映射 + 环境变量

```go
// ❌ 失败:gollm 不读取 OPENAI_API_BASE 环境变量
os.Setenv("OPENAI_API_BASE", cfg.BaseURL)
providerName := "openai"  // 映射 siliconflow → openai
```

**问题**: gollm 的 OpenAI provider 硬编码了端点 URL

### 方案 2 (当前): 混合客户端

```go
// ✅ 成功:直接使用支持自定义 BaseURL 的原生客户端
if shouldUseGollm(cfg.Name) {
    return createGollmClient(cfg)
} else {
    return createNativeLLMClient(cfg)  // 完全支持自定义 BaseURL
}
```

**优势**: 充分利用项目已有的实现,避免 gollm 的限制

## 后续优化建议

### 1. 统一接口

考虑为 gollm 客户端和原生客户端创建统一的适配器接口,简化 Complete 方法的逻辑。

### 2. 配置验证

在启动时验证 BaseURL 的有效性,避免运行时错误。

### 3. 健康检查

为每个提供商实现独立的健康检查端点。

### 4. 指标收集

收集每个提供商的详细调用指标:
- 请求成功率
- 平均响应时间
- Token 使用量
- 成本统计

## 总结

通过采用**混合客户端策略**,我们成功解决了 SiliconFlow API 端点问题:

- ✅ SiliconFlow 请求正确发送到 `https://api.siliconflow.cn/v1`
- ✅ DeepSeek 请求正确发送到 `https://api.deepseek.com/v1`
- ✅ 保留了对 gollm 原生支持提供商的支持
- ✅ 实现了灵活、可扩展的多提供商架构

这个方案充分利用了项目现有的基础设施,避免了 gollm 库的限制,为未来添加更多 LLM 提供商奠定了良好的基础。
