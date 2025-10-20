package testutil

import (
	"context"
	"fmt"
	"strings"

	"github.com/teilomillet/gollm/llm"
)

// MockLLM 实现 gollm.LLM 接口用于测试
type MockLLM struct {
	// 预定义的响应映射 (prompt -> response)
	Responses map[string]string
	// 调用历史
	CallHistory []MockCall
	// 是否返回错误
	ShouldError bool
	ErrorMsg    string
}

// MockCall 记录每次调用的信息
type MockCall struct {
	Prompt   *llm.Prompt
	Response string
	Error    error
}

// NewMockLLM 创建新的 Mock LLM
func NewMockLLM() *MockLLM {
	return &MockLLM{
		Responses:   make(map[string]string),
		CallHistory: make([]MockCall, 0),
	}
}

// WithResponse 添加预定义响应
func (m *MockLLM) WithResponse(prompt, response string) *MockLLM {
	m.Responses[prompt] = response
	return m
}

// WithError 设置返回错误
func (m *MockLLM) WithError(errMsg string) *MockLLM {
	m.ShouldError = true
	m.ErrorMsg = errMsg
	return m
}

// Generate 实现 gollm.LLM 接口 - 生成响应
func (m *MockLLM) Generate(ctx context.Context, prompt *llm.Prompt, opts ...llm.GenerateOption) (string, error) {
	if m.ShouldError {
		err := fmt.Errorf("%s", m.ErrorMsg)
		m.CallHistory = append(m.CallHistory, MockCall{
			Prompt: prompt,
			Error:  err,
		})
		return "", err
	}

	// 获取 prompt 文本
	promptText := prompt.String()

	// 查找匹配的响应
	var response string
	found := false

	// 精确匹配
	if resp, ok := m.Responses[promptText]; ok {
		response = resp
		found = true
	} else {
		// 模糊匹配 (包含关键词)
		for key, resp := range m.Responses {
			if strings.Contains(promptText, key) {
				response = resp
				found = true
				break
			}
		}
	}

	// 如果没有匹配的响应，返回默认响应
	if !found {
		response = "This is a mock LLM response for: " + promptText
	}

	m.CallHistory = append(m.CallHistory, MockCall{
		Prompt:   prompt,
		Response: response,
	})

	return response, nil
}

// GetCallCount 获取调用次数
func (m *MockLLM) GetCallCount() int {
	return len(m.CallHistory)
}

// GetLastCall 获取最后一次调用
func (m *MockLLM) GetLastCall() *MockCall {
	if len(m.CallHistory) == 0 {
		return nil
	}
	return &m.CallHistory[len(m.CallHistory)-1]
}

// Reset 重置 Mock 状态
func (m *MockLLM) Reset() {
	m.CallHistory = make([]MockCall, 0)
	m.ShouldError = false
	m.ErrorMsg = ""
}

// CreateMockRootCauseResponse 创建根因分析的 Mock 响应
func CreateMockRootCauseResponse() string {
	return `{
  "root_cause": "Pod OOMKilled due to memory limit exceeded",
  "confidence": 0.95,
  "reasoning": "Analysis of pod events shows container memory usage exceeded 512Mi limit",
  "recommendations": [
    "Increase memory limit to 1Gi",
    "Add memory resource requests",
    "Review application memory leaks"
  ]
}`
}

// CreateMockDescriptionResponse 创建故障描述的 Mock 响应
func CreateMockDescriptionResponse() string {
	return `{
  "description": "The application pod experienced an Out Of Memory (OOM) error and was terminated by Kubernetes",
  "severity": "high",
  "affected_components": ["api-server", "database-connection-pool"],
  "timeline": "Error occurred at 2024-01-15 10:30:00 UTC"
}`
}

// CreateMockRecommendationResponse 创建建议的 Mock 响应
func CreateMockRecommendationResponse() string {
	return `{
  "recommendations": [
    {
      "action": "Increase memory limit",
      "priority": "high",
      "command": "kubectl set resources deployment/api-server --limits=memory=1Gi"
    },
    {
      "action": "Add horizontal pod autoscaling",
      "priority": "medium",
      "command": "kubectl autoscale deployment api-server --min=3 --max=10 --cpu-percent=70"
    }
  ]
}`
}

// NewMockPrompt 创建用于测试的 gollm.Prompt
func NewMockPrompt(text string) *llm.Prompt {
	return llm.NewPrompt(text)
}
