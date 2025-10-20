package description

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reasoning-service-go/pkg/llm/proxy"
	"sort"
	"strings"
	"time"
)

// DescriptionChain 故障描述 Chain 实现
type DescriptionChain struct {
	llmProxy *proxy.ProxyAdapter
	config   *ChainConfig
}

// NewDescriptionChain 创建新的故障描述 Chain
func NewDescriptionChain(llmProxy *proxy.ProxyAdapter, config *ChainConfig) (*DescriptionChain, error) {
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

	return &DescriptionChain{
		llmProxy: llmProxy,
		config:   config,
	}, nil
}

// Generate 生成故障描述
func (c *DescriptionChain) Generate(ctx context.Context, input *DescriptionInput) (*DescriptionOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}

	// 应用默认值
	if input.Language == "" {
		input.Language = c.config.DefaultLanguage
	}
	if input.DetailLevel == "" {
		input.DetailLevel = c.config.DefaultDetailLevel
	}

	// 验证输入
	if err := validateInput(input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// 应用超时
	if c.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()
	}

	start := time.Now()

	// 1. 构建 Prompt
	prompt, err := c.buildPrompt(input)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt: %w", err)
	}

	log.Printf("Description generation prompt built for resource: %s/%s, language: %s",
		input.Namespace, input.ResourceName, input.Language)

	// 2. 调用 LLM
	req := &proxy.CompletionRequest{
		Messages: []proxy.Message{
			{
				Role:    "system",
				Content: c.getSystemPrompt(input.Language),
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
	output.Language = input.Language
	output.GeneratedAt = start
	output.Provider = response.Provider
	output.Model = response.Model
	output.TokensUsed = response.TokensUsed
	output.Latency = time.Since(start)

	log.Printf("Description generated in %v, severity: %s",
		output.Latency, output.Severity)

	return output, nil
}

// buildPrompt 构建描述生成 Prompt
func (c *DescriptionChain) buildPrompt(input *DescriptionInput) (string, error) {
	var sb strings.Builder

	// 根据语言选择标题
	if input.Language == "zh" {
		sb.WriteString("# Kubernetes 故障描述生成\n\n")
	} else {
		sb.WriteString("# Kubernetes Failure Description Generation\n\n")
	}

	// 基本信息
	if input.Language == "zh" {
		sb.WriteString("## 故障信息\n")
		sb.WriteString(fmt.Sprintf("- **故障类型**: %s\n", input.FailureType))
		sb.WriteString(fmt.Sprintf("- **资源类型**: %s\n", input.ResourceType))
		sb.WriteString(fmt.Sprintf("- **资源名称**: %s\n", input.ResourceName))
		sb.WriteString(fmt.Sprintf("- **命名空间**: %s\n", input.Namespace))
		sb.WriteString(fmt.Sprintf("- **集群 ID**: %s\n", input.ClusterID))
		sb.WriteString(fmt.Sprintf("- **发生时间**: %s\n", input.Timestamp.Format(time.RFC3339)))
	} else {
		sb.WriteString("## Failure Information\n")
		sb.WriteString(fmt.Sprintf("- **Failure Type**: %s\n", input.FailureType))
		sb.WriteString(fmt.Sprintf("- **Resource Type**: %s\n", input.ResourceType))
		sb.WriteString(fmt.Sprintf("- **Resource Name**: %s\n", input.ResourceName))
		sb.WriteString(fmt.Sprintf("- **Namespace**: %s\n", input.Namespace))
		sb.WriteString(fmt.Sprintf("- **Cluster ID**: %s\n", input.ClusterID))
		sb.WriteString(fmt.Sprintf("- **Timestamp**: %s\n", input.Timestamp.Format(time.RFC3339)))
	}

	if input.ErrorMessage != "" {
		if input.Language == "zh" {
			sb.WriteString(fmt.Sprintf("- **错误信息**: %s\n", input.ErrorMessage))
		} else {
			sb.WriteString(fmt.Sprintf("- **Error Message**: %s\n", input.ErrorMessage))
		}
	}

	if input.Impact != "" {
		if input.Language == "zh" {
			sb.WriteString(fmt.Sprintf("- **影响范围**: %s\n", input.Impact))
		} else {
			sb.WriteString(fmt.Sprintf("- **Impact**: %s\n", input.Impact))
		}
	}

	// 症状
	if len(input.Symptoms) > 0 {
		if input.Language == "zh" {
			sb.WriteString("\n## 症状\n")
		} else {
			sb.WriteString("\n## Symptoms\n")
		}
		for i, symptom := range input.Symptoms {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, symptom))
		}
	}

	// Pod 事件
	if len(input.PodEvents) > 0 {
		if input.Language == "zh" {
			sb.WriteString("\n## Pod 事件\n")
		} else {
			sb.WriteString("\n## Pod Events\n")
		}
		for _, event := range input.PodEvents {
			sb.WriteString(fmt.Sprintf("- [%s] %s: %s (at %s)\n",
				event.Type, event.Reason, event.Message, event.Timestamp.Format(time.RFC3339)))
		}
	}

	// Pod 日志
	if input.PodLogs != "" {
		if input.Language == "zh" {
			sb.WriteString("\n## Pod 日志\n")
		} else {
			sb.WriteString("\n## Pod Logs\n")
		}
		sb.WriteString("```\n")
		logs := input.PodLogs
		if len(logs) > 3000 {
			logs = logs[:3000] + "\n... (truncated)"
		}
		sb.WriteString(logs)
		sb.WriteString("\n```\n")
	}

	// 资源状态
	if len(input.ResourceStatus) > 0 {
		if input.Language == "zh" {
			sb.WriteString("\n## 资源状态\n")
		} else {
			sb.WriteString("\n## Resource Status\n")
		}
		for key, value := range input.ResourceStatus {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", key, value))
		}
	}

	// 指标数据
	if len(input.Metrics) > 0 {
		if input.Language == "zh" {
			sb.WriteString("\n## 指标数据\n")
		} else {
			sb.WriteString("\n## Metrics\n")
		}
		for key, value := range input.Metrics {
			sb.WriteString(fmt.Sprintf("- **%s**: %.2f\n", key, value))
		}
	}

	// 根因分析结果
	if input.RootCause != nil {
		if input.Language == "zh" {
			sb.WriteString("\n## 根因分析结果\n")
			sb.WriteString(fmt.Sprintf("- **根本原因**: %s\n", input.RootCause.RootCause))
			sb.WriteString(fmt.Sprintf("- **置信度**: %.2f\n", input.RootCause.Confidence))
			sb.WriteString(fmt.Sprintf("- **类别**: %s\n", input.RootCause.Category))
			if input.RootCause.Reasoning != "" {
				sb.WriteString(fmt.Sprintf("- **推理过程**: %s\n", input.RootCause.Reasoning))
			}
		} else {
			sb.WriteString("\n## Root Cause Analysis\n")
			sb.WriteString(fmt.Sprintf("- **Root Cause**: %s\n", input.RootCause.RootCause))
			sb.WriteString(fmt.Sprintf("- **Confidence**: %.2f\n", input.RootCause.Confidence))
			sb.WriteString(fmt.Sprintf("- **Category**: %s\n", input.RootCause.Category))
			if input.RootCause.Reasoning != "" {
				sb.WriteString(fmt.Sprintf("- **Reasoning**: %s\n", input.RootCause.Reasoning))
			}
		}
	}

	// 生成要求
	sb.WriteString("\n")
	if input.Language == "zh" {
		sb.WriteString("## 生成要求\n")
		sb.WriteString("请基于以上信息生成一份清晰、专业的故障描述,包含:\n")
		sb.WriteString("1. **标题**: 简洁的故障标题\n")
		sb.WriteString("2. **摘要**: 简短的故障摘要 (1-2 句话)\n")
		sb.WriteString("3. **描述**: 详细的故障描述\n")
		sb.WriteString("4. **严重程度**: critical, high, medium, low\n")
		sb.WriteString("5. **受影响组件**: 受影响的组件列表\n")
		sb.WriteString("6. **用户影响**: 对用户的影响\n")
		sb.WriteString("7. **业务影响**: 对业务的影响\n")
		if input.IncludeTimeline {
			sb.WriteString("8. **时间线**: 故障发生的时间线\n")
		}

		detailLevelDesc := map[string]string{
			"brief":    "简要",
			"normal":   "正常",
			"detailed": "详细",
		}
		sb.WriteString(fmt.Sprintf("\n详细程度: %s\n", detailLevelDesc[input.DetailLevel]))
		sb.WriteString("语言: 中文\n")
	} else {
		sb.WriteString("## Generation Requirements\n")
		sb.WriteString("Please generate a clear, professional failure description including:\n")
		sb.WriteString("1. **Title**: Concise failure title\n")
		sb.WriteString("2. **Summary**: Brief summary (1-2 sentences)\n")
		sb.WriteString("3. **Description**: Detailed failure description\n")
		sb.WriteString("4. **Severity**: critical, high, medium, low\n")
		sb.WriteString("5. **Affected Components**: List of affected components\n")
		sb.WriteString("6. **User Impact**: Impact on users\n")
		sb.WriteString("7. **Business Impact**: Impact on business\n")
		if input.IncludeTimeline {
			sb.WriteString("8. **Timeline**: Timeline of events\n")
		}

		sb.WriteString(fmt.Sprintf("\nDetail Level: %s\n", input.DetailLevel))
		sb.WriteString("Language: English\n")
	}

	sb.WriteString("\nProvide the response in JSON format following this structure:\n")
	sb.WriteString("```json\n")
	sb.WriteString(c.getJSONSchema(input.Language, input.IncludeTimeline))
	sb.WriteString("\n```\n")

	return sb.String(), nil
}

// parseResponse 解析 LLM 响应
func (c *DescriptionChain) parseResponse(content string) (*DescriptionOutput, error) {
	// 提取 JSON 内容
	jsonContent := extractJSON(content)

	var output DescriptionOutput
	if err := json.Unmarshal([]byte(jsonContent), &output); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// 验证必填字段
	if output.Title == "" {
		return nil, fmt.Errorf("title is empty")
	}
	if output.Description == "" {
		return nil, fmt.Errorf("description is empty")
	}

	// 验证严重程度
	if !isValidSeverity(output.Severity) {
		output.Severity = "medium" // 默认值
	}

	// 排序时间线
	if len(output.Timeline) > 0 {
		sort.Slice(output.Timeline, func(i, j int) bool {
			return output.Timeline[i].Timestamp.Before(output.Timeline[j].Timestamp)
		})
	}

	return &output, nil
}

// getSystemPrompt 获取系统提示 (支持多语言)
func (c *DescriptionChain) getSystemPrompt(language string) string {
	if language == "zh" {
		return `你是一名专业的 Kubernetes 故障分析专家,擅长生成清晰、准确的故障描述。

你的任务是基于提供的故障信息,生成一份专业的故障描述,包括:
- 清晰的标题和摘要
- 详细的故障描述
- 准确的严重程度评估
- 受影响组件分析
- 用户和业务影响评估
- 时间线(如果需要)

请始终使用 JSON 格式输出,确保描述准确、专业、易于理解。`
	}

	return `You are an expert Kubernetes troubleshooting specialist specialized in generating clear, accurate failure descriptions.

Your task is to generate professional failure descriptions based on provided information, including:
- Clear title and summary
- Detailed failure description
- Accurate severity assessment
- Affected components analysis
- User and business impact assessment
- Timeline (if requested)

Always respond in valid JSON format. Ensure descriptions are accurate, professional, and easy to understand.`
}

// getJSONSchema 获取 JSON 响应模板
func (c *DescriptionChain) getJSONSchema(language string, includeTimeline bool) string {
	if language == "zh" {
		schema := `{
  "title": "故障标题",
  "summary": "简短摘要",
  "description": "详细描述",
  "severity": "critical|high|medium|low",
  "affected_components": ["组件1", "组件2"],
  "user_impact": "对用户的影响",
  "business_impact": "对业务的影响"`

		if includeTimeline {
			schema += `,
  "timeline": [
    {
      "timestamp": "2024-01-01T00:00:00Z",
      "event": "事件描述",
      "severity": "high",
      "component": "组件名称"
    }
  ]`
		}

		schema += `,
  "technical_details": {
    "key1": "value1",
    "key2": "value2"
  }
}`
		return schema
	}

	// English schema
	schema := `{
  "title": "Failure title",
  "summary": "Brief summary",
  "description": "Detailed description",
  "severity": "critical|high|medium|low",
  "affected_components": ["component1", "component2"],
  "user_impact": "Impact on users",
  "business_impact": "Impact on business"`

	if includeTimeline {
		schema += `,
  "timeline": [
    {
      "timestamp": "2024-01-01T00:00:00Z",
      "event": "Event description",
      "severity": "high",
      "component": "Component name"
    }
  ]`
	}

	schema += `,
  "technical_details": {
    "key1": "value1",
    "key2": "value2"
  }
}`
	return schema
}

// extractJSON 从文本中提取 JSON 内容
func extractJSON(content string) string {
	content = strings.TrimSpace(content)

	// 移除 markdown 代码块
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

// validateConfig 验证配置
func validateConfig(config *ChainConfig) error {
	if config.Temperature < 0 || config.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}

	if config.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive")
	}

	if config.DefaultLanguage != "" {
		if _, ok := SupportedLanguages[config.DefaultLanguage]; !ok {
			return fmt.Errorf("unsupported language: %s", config.DefaultLanguage)
		}
	}

	if config.DefaultDetailLevel != "" {
		valid := false
		for _, level := range SupportedDetailLevels {
			if config.DefaultDetailLevel == level {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("unsupported detail level: %s", config.DefaultDetailLevel)
		}
	}

	return nil
}

// validateInput 验证输入
func validateInput(input *DescriptionInput) error {
	if input.Language != "" {
		if _, ok := SupportedLanguages[input.Language]; !ok {
			return fmt.Errorf("unsupported language: %s", input.Language)
		}
	}

	if input.DetailLevel != "" {
		valid := false
		for _, level := range SupportedDetailLevels {
			if input.DetailLevel == level {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("unsupported detail level: %s", input.DetailLevel)
		}
	}

	return nil
}

// isValidSeverity 验证严重程度
func isValidSeverity(severity string) bool {
	for _, s := range SupportedSeverities {
		if severity == s {
			return true
		}
	}
	return false
}

// DefaultChainConfig 返回默认配置
func DefaultChainConfig() *ChainConfig {
	return &ChainConfig{
		Temperature:        0.5, // 较低温度以确保一致性
		MaxTokens:          2048,
		DefaultLanguage:    "en",
		DefaultDetailLevel: "normal",
		IncludeTimeline:    true,
		SystemPrompt:       "", // 使用动态生成的 system prompt
		Timeout:            30 * time.Second,
	}
}
