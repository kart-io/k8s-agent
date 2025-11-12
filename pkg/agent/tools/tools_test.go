package tools

import (
	"context"
	"testing"
	"time"

	agentcore "github.com/kart-io/k8s-agent/pkg/agent/core"
)

// TestBaseTool 测试基础工具
func TestBaseTool(t *testing.T) {
	tool := NewBaseTool(
		"test_tool",
		"A test tool",
		`{"type": "object"}`,
		func(ctx context.Context, input *ToolInput) (*ToolOutput, error) {
			return &ToolOutput{
				Result:  "test result",
				Success: true,
			}, nil
		},
	)

	if tool.Name() != "test_tool" {
		t.Errorf("Expected name 'test_tool', got '%s'", tool.Name())
	}

	if tool.Description() != "A test tool" {
		t.Errorf("Expected description 'A test tool', got '%s'", tool.Description())
	}

	ctx := context.Background()
	input := &ToolInput{
		Args:    map[string]interface{}{},
		Context: ctx,
	}

	output, err := tool.Invoke(ctx, input)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	if output.Result != "test result" {
		t.Errorf("Expected result 'test result', got '%v'", output.Result)
	}
}

// TestFunctionTool 测试函数工具
func TestFunctionTool(t *testing.T) {
	tool := NewFunctionTool(
		"adder",
		"Adds two numbers",
		`{"type": "object", "properties": {"a": {"type": "number"}, "b": {"type": "number"}}}`,
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			a := args["a"].(float64)
			b := args["b"].(float64)
			return a + b, nil
		},
	)

	ctx := context.Background()
	input := &ToolInput{
		Args: map[string]interface{}{
			"a": 5.0,
			"b": 3.0,
		},
		Context: ctx,
	}

	output, err := tool.Invoke(ctx, input)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	if result, ok := output.Result.(float64); !ok || result != 8.0 {
		t.Errorf("Expected result 8.0, got %v", output.Result)
	}
}

// TestCalculatorTool 测试计算器工具
func TestCalculatorTool(t *testing.T) {
	tool := NewCalculatorTool()
	ctx := context.Background()

	tests := []struct {
		expression string
		expected   float64
	}{
		{"2 + 3", 5.0},
		{"10 - 5", 5.0},
		{"4 * 3", 12.0},
		{"15 / 3", 5.0},
		{"2 + 3 * 4", 14.0},
		{"(2 + 3) * 4", 20.0},
		{"2^3", 8.0},
	}

	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			input := &ToolInput{
				Args: map[string]interface{}{
					"expression": tt.expression,
				},
				Context: ctx,
			}

			output, err := tool.Invoke(ctx, input)
			if err != nil {
				t.Fatalf("Invoke failed: %v", err)
			}

			if !output.Success {
				t.Errorf("Expected success=true, got error: %s", output.Error)
			}

			if result, ok := output.Result.(float64); !ok || result != tt.expected {
				t.Errorf("Expected result %v, got %v", tt.expected, output.Result)
			}
		})
	}
}

// TestSearchTool 测试搜索工具
func TestSearchTool(t *testing.T) {
	engine := NewMockSearchEngine()
	tool := NewSearchTool(engine)

	ctx := context.Background()
	input := &ToolInput{
		Args: map[string]interface{}{
			"query":       "golang",
			"max_results": 2.0,
		},
		Context: ctx,
	}

	output, err := tool.Invoke(ctx, input)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if !output.Success {
		t.Errorf("Expected success=true, got error: %s", output.Error)
	}

	results, ok := output.Result.([]SearchResult)
	if !ok {
		t.Fatalf("Expected result to be []SearchResult")
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

// TestShellTool 测试 Shell 工具
func TestShellTool(t *testing.T) {
	tool := NewShellTool([]string{"echo", "pwd"}, 5*time.Second)

	ctx := context.Background()

	t.Run("AllowedCommand", func(t *testing.T) {
		input := &ToolInput{
			Args: map[string]interface{}{
				"command": "echo",
				"args":    []interface{}{"hello", "world"},
			},
			Context: ctx,
		}

		output, err := tool.Invoke(ctx, input)
		if err != nil {
			t.Fatalf("Invoke failed: %v", err)
		}

		if !output.Success {
			t.Errorf("Expected success=true, got error: %s", output.Error)
		}

		result, ok := output.Result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected result to be map[string]interface{}")
		}

		if result["exit_code"] != 0 {
			t.Errorf("Expected exit_code 0, got %v", result["exit_code"])
		}
	})

	t.Run("DisallowedCommand", func(t *testing.T) {
		input := &ToolInput{
			Args: map[string]interface{}{
				"command": "rm",
				"args":    []interface{}{"-rf", "/"},
			},
			Context: ctx,
		}

		output, err := tool.Invoke(ctx, input)
		if err == nil {
			t.Error("Expected error for disallowed command")
		}

		if output.Success {
			t.Error("Expected success=false for disallowed command")
		}
	})
}

// TestAPITool 测试 API 工具
func TestAPITool(t *testing.T) {
	tool := NewAPITool("https://jsonplaceholder.typicode.com", 10*time.Second, nil)

	ctx := context.Background()

	t.Run("GET Request", func(t *testing.T) {
		input := &ToolInput{
			Args: map[string]interface{}{
				"method": "GET",
				"url":    "/posts/1",
			},
			Context: ctx,
		}

		output, err := tool.Invoke(ctx, input)
		if err != nil {
			t.Fatalf("Invoke failed: %v", err)
		}

		if !output.Success {
			t.Errorf("Expected success=true, got error: %s", output.Error)
		}

		result, ok := output.Result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected result to be map[string]interface{}")
		}

		if result["status_code"] != 200 {
			t.Errorf("Expected status_code 200, got %v", result["status_code"])
		}
	})
}

// TestToolWithCallbacks 测试工具回调
func TestToolWithCallbacks(t *testing.T) {
	var callbackExecuted bool

	callback := &testCallback{
		onToolStart: func(ctx context.Context, toolName string, input interface{}) error {
			callbackExecuted = true
			if toolName != "test_tool" {
				t.Errorf("Expected toolName 'test_tool', got '%s'", toolName)
			}
			return nil
		},
	}

	tool := NewBaseTool(
		"test_tool",
		"A test tool",
		`{}`,
		func(ctx context.Context, input *ToolInput) (*ToolOutput, error) {
			return &ToolOutput{Success: true}, nil
		},
	)

	toolWithCallback := tool.WithCallbacks(callback).(Tool)

	ctx := context.Background()
	input := &ToolInput{
		Args:    map[string]interface{}{},
		Context: ctx,
	}

	_, err := toolWithCallback.Invoke(ctx, input)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if !callbackExecuted {
		t.Error("Callback was not executed")
	}
}

// TestToolkitBasic 测试基础工具集
func TestToolkitBasic(t *testing.T) {
	tool1 := NewCalculatorTool()
	tool2 := NewSearchTool(NewMockSearchEngine())

	toolkit := NewBaseToolkit(tool1, tool2)

	tools := toolkit.GetTools()
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}

	names := toolkit.GetToolNames()
	if len(names) != 2 {
		t.Errorf("Expected 2 tool names, got %d", len(names))
	}

	tool, err := toolkit.GetToolByName("calculator")
	if err != nil {
		t.Fatalf("GetToolByName failed: %v", err)
	}

	if tool.Name() != "calculator" {
		t.Errorf("Expected tool name 'calculator', got '%s'", tool.Name())
	}

	_, err = toolkit.GetToolByName("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent tool")
	}
}

// TestToolkitBuilder 测试工具集构建器
func TestToolkitBuilder(t *testing.T) {
	toolkit := NewToolkitBuilder().
		WithCalculator().
		WithSearch(nil).
		Build()

	tools := toolkit.GetTools()
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}

	names := toolkit.GetToolNames()
	expectedNames := map[string]bool{
		"calculator": true,
		"search":     true,
	}

	for _, name := range names {
		if !expectedNames[name] {
			t.Errorf("Unexpected tool name: %s", name)
		}
	}
}

// TestToolRegistry 测试工具注册表
func TestToolRegistry(t *testing.T) {
	registry := NewToolRegistry()

	tool := NewCalculatorTool()
	err := registry.Register(tool)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	retrieved, err := registry.Get("calculator")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Name() != "calculator" {
		t.Errorf("Expected tool name 'calculator', got '%s'", retrieved.Name())
	}

	// Test duplicate registration
	err = registry.Register(tool)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}

	// Test unregister
	err = registry.Unregister("calculator")
	if err != nil {
		t.Fatalf("Unregister failed: %v", err)
	}

	_, err = registry.Get("calculator")
	if err == nil {
		t.Error("Expected error after unregistration")
	}
}

// testCallback 测试回调实现
type testCallback struct {
	agentcore.BaseCallback
	onToolStart func(ctx context.Context, toolName string, input interface{}) error
	onToolEnd   func(ctx context.Context, toolName string, output interface{}) error
	onToolError func(ctx context.Context, toolName string, err error) error
}

func (t *testCallback) OnToolStart(ctx context.Context, toolName string, input interface{}) error {
	if t.onToolStart != nil {
		return t.onToolStart(ctx, toolName, input)
	}
	return nil
}

func (t *testCallback) OnToolEnd(ctx context.Context, toolName string, output interface{}) error {
	if t.onToolEnd != nil {
		return t.onToolEnd(ctx, toolName, output)
	}
	return nil
}

func (t *testCallback) OnToolError(ctx context.Context, toolName string, err error) error {
	if t.onToolError != nil {
		return t.onToolError(ctx, toolName, err)
	}
	return nil
}

// BenchmarkCalculatorTool 性能测试
func BenchmarkCalculatorTool(b *testing.B) {
	tool := NewCalculatorTool()
	ctx := context.Background()
	input := &ToolInput{
		Args: map[string]interface{}{
			"expression": "2 + 3 * 4",
		},
		Context: ctx,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tool.Invoke(ctx, input)
	}
}

// BenchmarkFunctionTool 性能测试
func BenchmarkFunctionTool(b *testing.B) {
	tool := NewFunctionTool(
		"adder",
		"Adds numbers",
		`{}`,
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			a := args["a"].(float64)
			b := args["b"].(float64)
			return a + b, nil
		},
	)

	ctx := context.Background()
	input := &ToolInput{
		Args: map[string]interface{}{
			"a": 5.0,
			"b": 3.0,
		},
		Context: ctx,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tool.Invoke(ctx, input)
	}
}
