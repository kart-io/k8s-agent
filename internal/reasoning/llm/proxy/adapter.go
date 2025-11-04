package proxy

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/teilomillet/gollm"

	"github.com/kart-io/k8s-agent/internal/reasoning/config"
	llmclient "github.com/kart-io/k8s-agent/internal/reasoning/llm"
)

// ProxyAdapter 封装 gollm 提供统一的 LLM 访问接口
type ProxyAdapter struct {
	providers []*ProviderClient // 按优先级排序的提供商列表
	config    *config.LLMConfig // LLM 配置
	metrics   *UsageMetrics     // 使用指标
	mu        sync.RWMutex      // 保护 metrics 的并发访问
}

// ProviderClient 单个提供商客户端
type ProviderClient struct {
	name      string                    // 提供商名称
	priority  int                       // 优先级 (数值越小优先级越高)
	client    interface{}               // gollm 客户端或原生 LLM 客户端
	useGollm  bool                      // 是否使用 gollm
	config    *config.LLMProviderConfig // 提供商配置
	healthy   bool                      // 健康状态
	lastErr   string                    // 最后一次错误信息
	lastCheck time.Time                 // 最后一次检查时间
}

// NewProxyAdapter 创建新的代理适配器
// 从配置初始化提供商列表并按 priority 排序
func NewProxyAdapter(cfg *config.LLMConfig) (*ProxyAdapter, error) {
	if cfg == nil {
		return nil, fmt.Errorf("LLM config is nil")
	}

	if !cfg.Enabled {
		return nil, fmt.Errorf("LLM is disabled in config")
	}

	if len(cfg.Providers) == 0 {
		return nil, fmt.Errorf("no LLM providers configured")
	}

	adapter := &ProxyAdapter{
		config: cfg,
		metrics: &UsageMetrics{
			ProviderStats: make(map[string]ProviderMetrics),
		},
	}

	// 初始化提供商客户端
	for _, providerCfg := range cfg.Providers {
		// 跳过没有 API key 的提供商(ollama 除外)
		if providerCfg.APIKey == "" && providerCfg.Name != "ollama" {
			log.Printf("Skipping provider %s: no API key configured", providerCfg.Name)
			continue
		}

		// 判断是否应该使用 gollm
		useGollm := shouldUseGollm(providerCfg.Name)

		var clientInterface interface{}
		var err error

		if useGollm {
			// 使用 gollm 创建客户端
			clientInterface, err = createGollmClient(&providerCfg)
			if err != nil {
				log.Printf("Failed to create gollm client for %s: %v", providerCfg.Name, err)
				continue
			}
		} else {
			// 使用项目原生 LLM 客户端
			clientInterface, err = createNativeLLMClient(&providerCfg)
			if err != nil {
				log.Printf("Failed to create native LLM client for %s: %v", providerCfg.Name, err)
				continue
			}
		}

		providerClient := &ProviderClient{
			name:      providerCfg.Name,
			priority:  providerCfg.Priority,
			client:    clientInterface,
			useGollm:  useGollm,
			config:    &providerCfg,
			healthy:   true,
			lastCheck: time.Now(),
		}

		adapter.providers = append(adapter.providers, providerClient)
	}

	if len(adapter.providers) == 0 {
		return nil, fmt.Errorf("no valid LLM providers available (all missing API keys)")
	}

	// 按 priority 排序提供商 (数值越小优先级越高)
	sort.Slice(adapter.providers, func(i, j int) bool {
		return adapter.providers[i].priority < adapter.providers[j].priority
	})

	// 输出配置的提供商列表
	log.Printf("Initialized LLM Proxy Adapter with %d providers:", len(adapter.providers))
	for i, p := range adapter.providers {
		status := "configured"
		if p.config.APIKey != "" {
			status = "ready"
		}
		log.Printf("  %d. %s (priority=%d, model=%s, status=%s)",
			i+1, p.name, p.priority, p.config.Model, status)
	}

	return adapter, nil
}

// Complete 标准的补全请求
// 调用 gollm 或原生客户端发送请求到优先级最高的可用提供商
func (a *ProxyAdapter) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("completion request is nil")
	}

	// 遍历按优先级排序的提供商
	for _, provider := range a.providers {
		if !provider.healthy {
			log.Printf("Skipping unhealthy provider: %s", provider.name)
			continue
		}

		start := time.Now()
		var response string
		var err error

		if provider.useGollm {
			// 使用 gollm 客户端
			gollmClient, ok := provider.client.(gollm.LLM)
			if !ok {
				log.Printf("Provider %s: invalid gollm client type", provider.name)
				continue
			}

			// 构建 gollm prompt
			prompt, promptErr := buildGollmPrompt(req)
			if promptErr != nil {
				return nil, fmt.Errorf("failed to build prompt: %w", promptErr)
			}

			response, err = gollmClient.Generate(ctx, prompt)
		} else {
			// 使用原生 LLM 客户端
			nativeClient, ok := provider.client.(llmclient.Client)
			if !ok {
				log.Printf("Provider %s: invalid native client type", provider.name)
				continue
			}

			// 转换请求格式
			llmReq := &llmclient.CompletionRequest{
				Messages:    convertMessages(req.Messages),
				Temperature: req.Temperature,
				MaxTokens:   req.MaxTokens,
			}

			llmResp, respErr := nativeClient.Complete(ctx, llmReq)
			if respErr != nil {
				err = respErr
			} else {
				response = llmResp.Content
			}
		}

		latency := time.Since(start)

		if err != nil {
			log.Printf("Provider %s failed: %v, trying next...", provider.name, err)
			a.recordFailure(provider.name, err)
			continue
		}

		// 成功,构建响应
		tokensUsed := estimateTokens(response)
		cost := calculateCost(provider.name, tokensUsed)

		a.recordSuccess(provider.name, latency, tokensUsed, cost)

		return &CompletionResponse{
			Content:      response,
			Provider:     provider.name,
			Model:        provider.config.Model,
			TokensUsed:   tokensUsed,
			Cost:         cost,
			Latency:      latency,
			FinishReason: "stop",
		}, nil
	}

	return nil, fmt.Errorf("all providers failed")
}

// GetMetrics 获取使用指标
func (a *ProxyAdapter) GetMetrics() *UsageMetrics {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 返回副本以避免并发修改
	metricsCopy := &UsageMetrics{
		TotalRequests:   a.metrics.TotalRequests,
		SuccessfulCalls: a.metrics.SuccessfulCalls,
		FailedCalls:     a.metrics.FailedCalls,
		TotalCost:       a.metrics.TotalCost,
		ProviderStats:   make(map[string]ProviderMetrics),
	}

	for k, v := range a.metrics.ProviderStats {
		metricsCopy.ProviderStats[k] = v
	}

	return metricsCopy
}

// GetProviderStatus 获取所有提供商状态
func (a *ProxyAdapter) GetProviderStatus() map[string]ProviderStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()

	status := make(map[string]ProviderStatus)
	for _, p := range a.providers {
		status[p.name] = ProviderStatus{
			Name:      p.name,
			Healthy:   p.healthy,
			LastError: p.lastErr,
			LastCheck: p.lastCheck,
		}
	}

	return status
}

// recordSuccess 记录成功调用 (内部方法)
func (a *ProxyAdapter) recordSuccess(providerName string, latency time.Duration, tokensUsed int, cost float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.metrics.TotalRequests++
	a.metrics.SuccessfulCalls++
	a.metrics.TotalCost += cost

	stats := a.metrics.ProviderStats[providerName]
	stats.Calls++
	stats.Successes++
	stats.TotalCost += cost

	// 更新平均延迟
	if stats.AvgLatency == 0 {
		stats.AvgLatency = latency
	} else {
		// 简单移动平均
		stats.AvgLatency = (stats.AvgLatency + latency) / 2
	}

	a.metrics.ProviderStats[providerName] = stats
}

// recordFailure 记录失败调用 (内部方法)
func (a *ProxyAdapter) recordFailure(providerName string, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.metrics.TotalRequests++
	a.metrics.FailedCalls++

	stats := a.metrics.ProviderStats[providerName]
	stats.Calls++
	stats.Failures++
	a.metrics.ProviderStats[providerName] = stats

	// 更新提供商健康状态
	for _, p := range a.providers {
		if p.name == providerName {
			p.healthy = false
			p.lastErr = err.Error()
			p.lastCheck = time.Now()
			break
		}
	}
}

// shouldUseGollm 判断提供商是否应该使用 gollm
// gollm 原生支持: openai, anthropic, groq, ollama, mistral, openrouter
// 其他提供商使用项目原生 LLM 客户端
func shouldUseGollm(providerName string) bool {
	switch providerName {
	case "openai", "anthropic", "claude", "groq", "ollama", "mistral", "openrouter":
		return true
	default:
		return false
	}
}

// createNativeLLMClient 创建项目原生 LLM 客户端
func createNativeLLMClient(cfg *config.LLMProviderConfig) (llmclient.Client, error) {
	llmCfg := &llmclient.Config{
		Provider:    llmclient.Provider(cfg.Name),
		APIKey:      cfg.APIKey,
		BaseURL:     cfg.BaseURL,
		Model:       cfg.Model,
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
		Timeout:     cfg.Timeout,
	}

	log.Printf("Provider %s using native LLM client with base URL: %s", cfg.Name, cfg.BaseURL)

	return llmclient.NewClient(llmCfg)
}

// convertMessages 将 proxy 消息格式转换为 llmclient 格式
func convertMessages(proxyMessages []Message) []llmclient.Message {
	messages := make([]llmclient.Message, len(proxyMessages))
	for i, msg := range proxyMessages {
		messages[i] = llmclient.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return messages
}

// createGollmClient 创建 gollm 客户端
func createGollmClient(cfg *config.LLMProviderConfig) (gollm.LLM, error) {
	// 将超时秒转换为 time.Duration
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second // 默认 30 秒
	}

	// 创建 gollm 客户端
	var opts []gollm.ConfigOption
	opts = append(opts,
		gollm.SetProvider(cfg.Name),
		gollm.SetModel(cfg.Model),
		gollm.SetAPIKey(cfg.APIKey),
		gollm.SetMaxTokens(cfg.MaxTokens),
		gollm.SetTemperature(cfg.Temperature),
		gollm.SetTimeout(timeout),
		gollm.SetMaxRetries(3), // 默认重试 3 次
	)

	// Ollama 使用专门的端点设置
	if cfg.BaseURL != "" && cfg.Name == "ollama" {
		opts = append(opts, gollm.SetOllamaEndpoint(cfg.BaseURL))
		log.Printf("Provider ollama using endpoint: %s", cfg.BaseURL)
	}

	llm, err := gollm.NewLLM(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create internal LLM: %w", err)
	}

	log.Printf("Provider %s using gollm", cfg.Name)
	return llm, nil
}

// buildGollmPrompt 构建 gollm 提示
func buildGollmPrompt(req *CompletionRequest) (*gollm.Prompt, error) {
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("no messages in request")
	}

	// 合并所有消息为一个 prompt
	// TODO: 未来可以使用 gollm 的对话模式支持多轮对话
	var promptText string
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			promptText += msg.Content + "\n\n"
		} else if msg.Role == "user" {
			promptText += msg.Content
		}
	}

	prompt := gollm.NewPrompt(promptText)
	return prompt, nil
}

// estimateTokens 估算 token 数量
// 简单估算: 1 token ≈ 4 个字符
func estimateTokens(text string) int {
	return len(text) / 4
}

// calculateCost 计算调用成本
// 简化版本,实际成本应根据提供商和模型定价
func calculateCost(providerName string, tokensUsed int) float64 {
	// 简化的成本计算 (美元)
	// 实际应该根据 provider 和 model 查表
	costPerToken := 0.00002 // $0.02 per 1K tokens

	switch providerName {
	case "openai":
		costPerToken = 0.00002 // GPT-4
	case "gemini":
		costPerToken = 0.00001 // Gemini Pro
	case "deepseek":
		costPerToken = 0.000001 // DeepSeek
	default:
		costPerToken = 0.00001
	}

	return float64(tokensUsed) * costPerToken
}
