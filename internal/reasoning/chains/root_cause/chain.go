package root_cause

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kart-io/k8s-agent/internal/reasoning/llm/proxy"
)

// RootCauseChain 根因分析 Chain 实现.
type RootCauseChain struct {
	llmProxy *proxy.ProxyAdapter
	config   *ChainConfig
}

// NewRootCauseChain 创建新的根因分析 Chain.
func NewRootCauseChain(llmProxy *proxy.ProxyAdapter, config *ChainConfig) (*RootCauseChain, error) {
	if llmProxy == nil {
		return nil, fmt.Errorf("llmProxy is nil")
	}

	if config == nil {
		config = DefaultChainConfig()
	}

	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &RootCauseChain{
		llmProxy: llmProxy,
		config:   config,
	}, nil
}

// Analyze 执行根因分析.
func (c *RootCauseChain) Analyze(ctx context.Context, input *AnalysisInput) (*AnalysisOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}

	// 应用超时
	if c.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()
	}

	start := time.Now()

	// 1. 构建 Prompt
	prompt := c.buildPrompt(input)

	log.Printf("Root cause analysis prompt built for resource: %s/%s", input.Namespace, input.ResourceName)

	// 2. 调用 LLM
	req := &proxy.CompletionRequest{
		Messages: []proxy.Message{
			{
				Role:    "system",
				Content: c.config.SystemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: c.config.Temperature,
		MaxTokens:   c.config.MaxTokens,
	}

	response, err := c.llmProxy.Complete(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	log.Printf("LLM response received from provider: %s, model: %s, tokens: %d",
		response.Provider, response.Model, response.TokensUsed)

	// 3. 解析响应
	output, err := c.parseResponse(response.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 4. 填充元数据
	output.AnalysisTime = start
	output.Provider = response.Provider
	output.Model = response.Model
	output.TokensUsed = response.TokensUsed
	output.Latency = time.Since(start)

	// 5. 验证置信度
	if output.Confidence < c.config.MinConfidence {
		log.Printf("Warning: confidence %.2f below threshold %.2f",
			output.Confidence, c.config.MinConfidence)
	}

	log.Printf("Root cause analysis completed in %v, confidence: %.2f",
		output.Latency, output.Confidence)

	return output, nil
}

// buildPrompt 构建分析 Prompt.
func (c *RootCauseChain) buildPrompt(input *AnalysisInput) string {
	var sb strings.Builder

	// 基本信息
	sb.WriteString("# Kubernetes Failure Root Cause Analysis\n\n")
	sb.WriteString("## Failure Information\n")
	sb.WriteString(fmt.Sprintf("- **Failure Type**: %s\n", input.FailureType))
	sb.WriteString(fmt.Sprintf("- **Resource Type**: %s\n", input.ResourceType))
	sb.WriteString(fmt.Sprintf("- **Resource Name**: %s\n", input.ResourceName))
	sb.WriteString(fmt.Sprintf("- **Namespace**: %s\n", input.Namespace))
	sb.WriteString(fmt.Sprintf("- **Cluster ID**: %s\n", input.ClusterID))
	sb.WriteString(fmt.Sprintf("- **Timestamp**: %s\n", input.Timestamp.Format(time.RFC3339)))

	if input.ErrorMessage != "" {
		sb.WriteString(fmt.Sprintf("- **Error Message**: %s\n", input.ErrorMessage))
	}

	// 症状
	if len(input.Symptoms) > 0 {
		sb.WriteString("\n## Symptoms\n")
		for i, symptom := range input.Symptoms {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, symptom))
		}
	}

	// Pod 事件
	if len(input.PodEvents) > 0 {
		sb.WriteString("\n## Pod Events\n")
		for _, event := range input.PodEvents {
			sb.WriteString(fmt.Sprintf("- [%s] %s: %s (at %s)\n",
				event.Type, event.Reason, event.Message, event.Timestamp.Format(time.RFC3339)))
		}
	}

	// Pod 日志
	if input.PodLogs != "" {
		sb.WriteString("\n## Pod Logs\n")
		sb.WriteString("```\n")
		// 限制日志长度
		logs := input.PodLogs
		if len(logs) > 5000 {
			logs = logs[:5000] + "\n... (truncated)"
		}
		sb.WriteString(logs)
		sb.WriteString("\n```\n")
	}

	// 资源状态
	if len(input.ResourceStatus) > 0 {
		sb.WriteString("\n## Resource Status\n")
		for key, value := range input.ResourceStatus {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", key, value))
		}
	}

	// 指标数据
	if len(input.Metrics) > 0 {
		sb.WriteString("\n## Metrics\n")
		for key, value := range input.Metrics {
			sb.WriteString(fmt.Sprintf("- **%s**: %.2f\n", key, value))
		}
	}

	// 相似案例
	if c.config.IncludeSimilar && len(input.SimilarCases) > 0 {
		sb.WriteString("\n## Similar Cases\n")
		maxCases := c.config.MaxSimilarCases
		if maxCases <= 0 || maxCases > len(input.SimilarCases) {
			maxCases = len(input.SimilarCases)
		}
		for i := 0; i < maxCases; i++ {
			sc := input.SimilarCases[i]
			sb.WriteString(fmt.Sprintf("\n### Case %d (Similarity: %.2f)\n", i+1, sc.Similarity))
			sb.WriteString(fmt.Sprintf("- **Description**: %s\n", sc.Description))
			sb.WriteString(fmt.Sprintf("- **Root Cause**: %s\n", sc.RootCause))
			if sc.Resolution != "" {
				sb.WriteString(fmt.Sprintf("- **Resolution**: %s\n", sc.Resolution))
			}
		}
	}

	// 分析要求
	sb.WriteString("\n## Analysis Requirements\n")
	sb.WriteString("Please analyze the above information and provide:\n")
	sb.WriteString("1. **Root Cause**: The fundamental reason for the failure\n")
	sb.WriteString("2. **Confidence**: Your confidence level (0.0-1.0)\n")
	sb.WriteString("3. **Reasoning**: Step-by-step reasoning process\n")
	sb.WriteString("4. **Category**: Category of the root cause (e.g., resource, configuration, network, etc.)\n")
	sb.WriteString("5. **Contributing Factors**: Other factors that contributed to the issue\n")
	sb.WriteString("6. **Recommendations**: Actionable recommendations to fix and prevent recurrence\n")
	sb.WriteString("\nProvide the response in JSON format following this structure:\n")
	sb.WriteString("```json\n")
	sb.WriteString(getJSONSchema())
	sb.WriteString("\n```\n")

	return sb.String()
}

// parseResponse 解析 LLM 响应.
func (c *RootCauseChain) parseResponse(content string) (*AnalysisOutput, error) {
	// 提取 JSON 内容 (可能被 markdown 代码块包围)
	jsonContent := extractJSON(content)

	var output AnalysisOutput
	if err := json.Unmarshal([]byte(jsonContent), &output); err != nil {
		// 如果 JSON 解析失败，尝试从文本中提取关键信息
		log.Printf("JSON parsing failed, attempting text extraction: %v", err)
		return c.parseTextResponse(content)
	}

	// 验证必填字段
	if output.RootCause == "" {
		return nil, fmt.Errorf("root_cause is empty")
	}

	// 规范化置信度
	if output.Confidence < 0 {
		output.Confidence = 0
	} else if output.Confidence > 1 {
		output.Confidence = 1
	}

	return &output, nil
}

// parseTextResponse 从文本响应中提取信息 (fallback).
func (c *RootCauseChain) parseTextResponse(content string) (*AnalysisOutput, error) {
	// 简单的文本解析作为 fallback
	output := &AnalysisOutput{
		RootCause:  "Unable to parse structured response",
		Confidence: 0.5,
		Reasoning:  content,
		Category:   "unknown",
	}

	// 尝试从文本中提取根因
	if idx := strings.Index(strings.ToLower(content), "root cause"); idx != -1 {
		// 提取根因相关的句子
		lines := strings.Split(content[idx:], "\n")
		if len(lines) > 0 {
			output.RootCause = strings.TrimSpace(lines[0])
		}
	}

	return output, nil
}

// extractJSON 从文本中提取 JSON 内容.
func extractJSON(content string) string {
	// 移除 markdown 代码块标记
	content = strings.TrimSpace(content)

	// 查找 JSON 代码块
	if idx := strings.Index(content, "```json"); idx != -1 {
		content = content[idx+7:]
		if endIdx := strings.Index(content, "```"); endIdx != -1 {
			content = content[:endIdx]
		}
	} else if idx := strings.Index(content, "```"); idx != -1 {
		content = content[idx+3:]
		if endIdx := strings.Index(content, "```"); endIdx != -1 {
			content = content[:endIdx]
		}
	}

	// 查找 JSON 对象边界
	startIdx := strings.Index(content, "{")
	endIdx := strings.LastIndex(content, "}")

	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		content = content[startIdx : endIdx+1]
	}

	return strings.TrimSpace(content)
}

// getJSONSchema 获取 JSON 响应模板.
func getJSONSchema() string {
	return `{
  "root_cause": "string",
  "confidence": 0.0,
  "reasoning": "string",
  "category": "string",
  "contributing_factors": [
    {
      "name": "string",
      "description": "string",
      "impact": "high|medium|low",
      "evidence": "string"
    }
  ],
  "recommendations": [
    {
      "action": "string",
      "priority": "high|medium|low",
      "description": "string",
      "commands": ["string"],
      "impact": "string",
      "risk_level": "high|medium|low"
    }
  ]
}`
}

// validateConfig 验证配置.
func validateConfig(config *ChainConfig) error {
	if config.Temperature < 0 || config.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}

	if config.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive")
	}

	if config.MinConfidence < 0 || config.MinConfidence > 1 {
		return fmt.Errorf("min_confidence must be between 0 and 1")
	}

	return nil
}

// DefaultChainConfig 返回默认配置.
func DefaultChainConfig() *ChainConfig {
	return &ChainConfig{
		Temperature:     0.7,
		MaxTokens:       4096,
		MinConfidence:   0.6,
		IncludeSimilar:  true,
		MaxSimilarCases: 3,
		SystemPrompt:    defaultSystemPrompt(),
		Timeout:         30 * time.Second,
	}
}

// defaultSystemPrompt 默认系统提示.
func defaultSystemPrompt() string {
	return `You are an expert Kubernetes troubleshooting assistant specialized in root cause analysis.

Your task is to analyze Kubernetes failures and identify the root cause based on:
- Pod events and logs
- Resource status and metrics
- Error messages and symptoms
- Historical similar cases

Provide detailed, actionable analysis with:
1. Clear identification of the root cause
2. Confidence level based on available evidence
3. Step-by-step reasoning
4. Contributing factors
5. Specific recommendations with kubectl commands

Always respond in valid JSON format following the provided schema.
Be precise, thorough, and prioritize reliability over speculation.`
}
