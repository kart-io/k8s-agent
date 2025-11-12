package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kart-io/k8s-agent/pkg/agent/mcp/core"
	"github.com/kart-io/k8s-agent/pkg/agent/mcp/toolbox"
	"github.com/kart-io/k8s-agent/pkg/agent/mcp/tools"
)

func main() {
	fmt.Println("=== MCP 工具箱基础示例 ===")
	fmt.Println()

	// 1. 创建工具箱
	tb := toolbox.NewStandardToolBox()
	fmt.Println("✓ 创建工具箱")

	// 2. 注册工具
	if err := registerTools(tb); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 注册了 %d 个工具\n\n", tb.Statistics().TotalTools)

	// 3. 列出所有工具
	fmt.Println("=== 可用工具列表 ===")
	listTools(tb)

	// 4. 执行工具示例
	fmt.Println("\n=== 工具执行示例 ===")
	fmt.Println()

	// 示例 1: 读取文件
	demonstrateReadFile(tb)

	// 示例 2: 列出目录
	demonstrateListDirectory(tb)

	// 示例 3: JSON 解析
	demonstrateJSONParse(tb)

	// 示例 4: HTTP 请求
	demonstrateHTTPRequest(tb)

	// 5. 显示统计信息
	fmt.Println("\n=== 工具箱统计 ===")
	displayStatistics(tb)

	// 6. 查看调用历史
	fmt.Println("\n=== 调用历史 ===")
	displayHistory(tb)
}

// registerTools 注册所有工具
func registerTools(tb *toolbox.StandardToolBox) error {
	tools := []core.Tool{
		tools.NewReadFileTool(),
		tools.NewWriteFileTool(),
		tools.NewListDirectoryTool(),
		tools.NewSearchFilesTool(),
		tools.NewHTTPRequestTool(),
		tools.NewJSONParseTool(),
		tools.NewShellExecuteTool(),
	}

	for _, tool := range tools {
		if err := tb.Register(tool); err != nil {
			return fmt.Errorf("failed to register tool %s: %w", tool.Name(), err)
		}
	}

	return nil
}

// listTools 列出所有工具
func listTools(tb *toolbox.StandardToolBox) {
	metadata := tb.ListMetadata()

	for _, meta := range metadata {
		fmt.Printf("工具: %s\n", meta.Name)
		fmt.Printf("  描述: %s\n", meta.Description)
		fmt.Printf("  分类: %s\n", meta.Category)
		fmt.Printf("  需要认证: %v\n", meta.RequiresAuth)
		fmt.Printf("  危险操作: %v\n", meta.IsDangerous)
		fmt.Println()
	}
}

// demonstrateReadFile 演示读取文件
func demonstrateReadFile(tb *toolbox.StandardToolBox) {
	fmt.Println("1. 读取文件示例")

	// 创建测试文件
	testFile := "/tmp/mcp_test.txt"
	testContent := "Hello MCP Toolbox!"

	// 先写入文件
	writeCall := &core.ToolCall{
		ID:       "write-1",
		ToolName: "write_file",
		Input: map[string]interface{}{
			"path":    testFile,
			"content": testContent,
		},
		Timestamp: time.Now(),
	}

	writeResult, err := tb.Execute(context.Background(), writeCall)
	if err != nil {
		fmt.Printf("❌ 写入文件失败: %v\n\n", err)
		return
	}
	fmt.Printf("✓ 写入文件成功: %s (耗时: %v)\n", testFile, writeResult.Result.Duration)

	// 读取文件
	readCall := &core.ToolCall{
		ID:       "read-1",
		ToolName: "read_file",
		Input: map[string]interface{}{
			"path": testFile,
		},
		Timestamp: time.Now(),
	}

	readResult, err := tb.Execute(context.Background(), readCall)
	if err != nil {
		fmt.Printf("❌ 读取文件失败: %v\n\n", err)
		return
	}

	if readResult.Result.Success {
		data := readResult.Result.Data.(map[string]interface{})
		fmt.Printf("✓ 读取文件成功\n")
		fmt.Printf("  内容: %s\n", data["content"])
		fmt.Printf("  大小: %v 字节\n", data["size"])
		fmt.Printf("  耗时: %v\n", readResult.Result.Duration)
	}
	fmt.Println()
}

// demonstrateListDirectory 演示列出目录
func demonstrateListDirectory(tb *toolbox.StandardToolBox) {
	fmt.Println("2. 列出目录示例")

	call := &core.ToolCall{
		ID:       "list-1",
		ToolName: "list_directory",
		Input: map[string]interface{}{
			"path":      "/tmp",
			"recursive": false,
		},
		Timestamp: time.Now(),
	}

	result, err := tb.Execute(context.Background(), call)
	if err != nil {
		fmt.Printf("❌ 列出目录失败: %v\n\n", err)
		return
	}

	if result.Result.Success {
		data := result.Result.Data.(map[string]interface{})
		files := data["files"].([]map[string]interface{})

		fmt.Printf("✓ 列出目录成功\n")
		fmt.Printf("  路径: %s\n", data["path"])
		fmt.Printf("  文件数: %v\n", data["count"])
		fmt.Printf("  前 5 个文件:\n")

		for i, file := range files {
			if i >= 5 {
				break
			}
			fmt.Printf("    - %s (%v 字节)\n", file["name"], file["size"])
		}
		fmt.Printf("  耗时: %v\n", result.Result.Duration)
	}
	fmt.Println()
}

// demonstrateJSONParse 演示 JSON 解析
func demonstrateJSONParse(tb *toolbox.StandardToolBox) {
	fmt.Println("3. JSON 解析示例")

	jsonData := `{
		"user": {
			"name": "Alice",
			"age": 30,
			"email": "alice@example.com"
		},
		"status": "active"
	}`

	call := &core.ToolCall{
		ID:       "json-1",
		ToolName: "json_parse",
		Input: map[string]interface{}{
			"json": jsonData,
		},
		Timestamp: time.Now(),
	}

	result, err := tb.Execute(context.Background(), call)
	if err != nil {
		fmt.Printf("❌ JSON 解析失败: %v\n\n", err)
		return
	}

	if result.Result.Success {
		data := result.Result.Data.(map[string]interface{})
		parsed := data["parsed"]

		fmt.Printf("✓ JSON 解析成功\n")
		prettyJSON, _ := json.MarshalIndent(parsed, "  ", "  ")
		fmt.Printf("  解析结果:\n  %s\n", string(prettyJSON))
		fmt.Printf("  耗时: %v\n", result.Result.Duration)
	}
	fmt.Println()
}

// demonstrateHTTPRequest 演示 HTTP 请求
func demonstrateHTTPRequest(tb *toolbox.StandardToolBox) {
	fmt.Println("4. HTTP 请求示例")

	call := &core.ToolCall{
		ID:       "http-1",
		ToolName: "http_request",
		Input: map[string]interface{}{
			"url":    "https://httpbin.org/get",
			"method": "GET",
		},
		Timestamp: time.Now(),
	}

	result, err := tb.Execute(context.Background(), call)
	if err != nil {
		fmt.Printf("❌ HTTP 请求失败: %v\n\n", err)
		return
	}

	if result.Result.Success {
		data := result.Result.Data.(map[string]interface{})

		fmt.Printf("✓ HTTP 请求成功\n")
		fmt.Printf("  URL: %s\n", call.Input["url"])
		fmt.Printf("  状态: %v\n", data["status"])
		fmt.Printf("  状态码: %v\n", data["status_code"])
		fmt.Printf("  响应大小: %v 字节\n", data["size"])
		fmt.Printf("  耗时: %v\n", result.Result.Duration)
	}
	fmt.Println()
}

// displayStatistics 显示统计信息
func displayStatistics(tb *toolbox.StandardToolBox) {
	stats := tb.Statistics()

	fmt.Printf("工具总数: %d\n", stats.TotalTools)
	fmt.Printf("总调用次数: %d\n", stats.TotalCalls)
	fmt.Printf("成功次数: %d\n", stats.SuccessfulCalls)
	fmt.Printf("失败次数: %d\n", stats.FailedCalls)
	fmt.Printf("平均延迟: %.2f ms\n", stats.AverageLatency)

	fmt.Println("\n工具使用统计:")
	for tool, count := range stats.ToolUsage {
		fmt.Printf("  %s: %d 次\n", tool, count)
	}

	fmt.Println("\n分类使用统计:")
	for category, count := range stats.CategoryUsage {
		fmt.Printf("  %s: %d 次\n", category, count)
	}
}

// displayHistory 显示调用历史
func displayHistory(tb *toolbox.StandardToolBox) {
	history := tb.GetCallHistory()

	fmt.Printf("总共 %d 条调用记录\n\n", len(history))

	for i, record := range history {
		fmt.Printf("%d. 工具: %s\n", i+1, record.Call.ToolName)
		fmt.Printf("   调用ID: %s\n", record.Call.ID)
		fmt.Printf("   成功: %v\n", record.Result.Success)
		fmt.Printf("   耗时: %v\n", record.Result.Duration)
		fmt.Printf("   时间: %v\n", record.ExecutedAt.Format("15:04:05"))
		fmt.Println()
	}
}
