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
	fmt.Println("=== MCP 工具链编排示例 ===")
	fmt.Println()

	// 创建工具箱
	tb := toolbox.NewStandardToolBox()

	// 注册工具
	if err := tools.RegisterBuiltinTools(tb); err != nil {
		log.Fatal(err)
	}

	// 示例 1: 简单的工具链 - 写入文件然后读取
	fmt.Println("示例 1: 写入 -> 读取工具链")
	executeSimpleChain(tb)

	// 示例 2: 数据处理链 - HTTP请求 -> JSON解析
	fmt.Println("\n示例 2: HTTP请求 -> JSON解析工具链")
	executeDataProcessingChain(tb)

	// 示例 3: 文件处理链 - 搜索 -> 读取 -> 处理
	fmt.Println("\n示例 3: 文件搜索 -> 批量读取工具链")
	executeFileProcessingChain(tb)

	// 示例 4: 条件分支链
	fmt.Println("\n示例 4: 条件分支工具链")
	executeConditionalChain(tb)

	// 显示最终统计
	fmt.Println("\n=== 工具链执行统计 ===")
	displayStatistics(tb)
}

// executeSimpleChain 执行简单工具链：写入 -> 读取
func executeSimpleChain(tb *toolbox.StandardToolBox) {
	ctx := context.Background()

	// Step 1: 写入文件
	writeCall := &core.ToolCall{
		ID:       "chain1-step1",
		ToolName: "write_file",
		Input: map[string]interface{}{
			"path": "/tmp/chain_test.txt",
			"content": `{
				"name": "Tool Chain Example",
				"version": "1.0",
				"status": "running"
			}`,
		},
		Timestamp: time.Now(),
	}

	fmt.Println("  Step 1: 写入文件...")
	writeResult, err := tb.Execute(ctx, writeCall)
	if err != nil {
		fmt.Printf("    ❌ 失败: %v\n", err)
		return
	}
	fmt.Printf("    ✓ 成功 (耗时: %v)\n", writeResult.Result.Duration)

	// Step 2: 读取文件
	readCall := &core.ToolCall{
		ID:       "chain1-step2",
		ToolName: "read_file",
		Input: map[string]interface{}{
			"path": "/tmp/chain_test.txt",
		},
		Timestamp: time.Now(),
	}

	fmt.Println("  Step 2: 读取文件...")
	readResult, err := tb.Execute(ctx, readCall)
	if err != nil {
		fmt.Printf("    ❌ 失败: %v\n", err)
		return
	}

	data := readResult.Result.Data.(map[string]interface{})
	fmt.Printf("    ✓ 成功 (耗时: %v)\n", readResult.Result.Duration)
	fmt.Printf("    文件大小: %v 字节\n", data["size"])

	// Step 3: 解析 JSON
	jsonCall := &core.ToolCall{
		ID:       "chain1-step3",
		ToolName: "json_parse",
		Input: map[string]interface{}{
			"json": data["content"].(string),
		},
		Timestamp: time.Now(),
	}

	fmt.Println("  Step 3: 解析 JSON...")
	jsonResult, err := tb.Execute(ctx, jsonCall)
	if err != nil {
		fmt.Printf("    ❌ 失败: %v\n", err)
		return
	}

	parsed := jsonResult.Result.Data.(map[string]interface{})["parsed"]
	fmt.Printf("    ✓ 成功 (耗时: %v)\n", jsonResult.Result.Duration)
	prettyJSON, _ := json.MarshalIndent(parsed, "    ", "  ")
	fmt.Printf("    解析结果:\n    %s\n", string(prettyJSON))

	fmt.Printf("\n  工具链总耗时: %v\n",
		writeResult.Result.Duration+readResult.Result.Duration+jsonResult.Result.Duration)
}

// executeDataProcessingChain 执行数据处理链：HTTP -> JSON
func executeDataProcessingChain(tb *toolbox.StandardToolBox) {
	ctx := context.Background()

	// Step 1: HTTP 请求
	httpCall := &core.ToolCall{
		ID:       "chain2-step1",
		ToolName: "http_request",
		Input: map[string]interface{}{
			"url":    "https://api.github.com/repos/golang/go",
			"method": "GET",
			"headers": map[string]interface{}{
				"User-Agent": "MCP-Toolbox-Example",
			},
		},
		Timestamp: time.Now(),
	}

	fmt.Println("  Step 1: 发送 HTTP 请求到 GitHub API...")
	httpResult, err := tb.Execute(ctx, httpCall)
	if err != nil {
		fmt.Printf("    ❌ 失败: %v\n", err)
		return
	}

	httpData := httpResult.Result.Data.(map[string]interface{})
	fmt.Printf("    ✓ 成功 (状态码: %v, 耗时: %v)\n",
		httpData["status_code"], httpResult.Result.Duration)

	// Step 2: 解析响应 JSON
	body := httpData["body"].(string)
	jsonCall := &core.ToolCall{
		ID:       "chain2-step2",
		ToolName: "json_parse",
		Input: map[string]interface{}{
			"json": body,
		},
		Timestamp: time.Now(),
	}

	fmt.Println("  Step 2: 解析响应 JSON...")
	jsonResult, err := tb.Execute(ctx, jsonCall)
	if err != nil {
		fmt.Printf("    ❌ 失败: %v\n", err)
		return
	}

	parsed := jsonResult.Result.Data.(map[string]interface{})["parsed"].(map[string]interface{})
	fmt.Printf("    ✓ 成功 (耗时: %v)\n", jsonResult.Result.Duration)
	fmt.Printf("    仓库名称: %v\n", parsed["name"])
	fmt.Printf("    Stars: %v\n", parsed["stargazers_count"])
	fmt.Printf("    描述: %v\n", parsed["description"])

	fmt.Printf("\n  工具链总耗时: %v\n",
		httpResult.Result.Duration+jsonResult.Result.Duration)
}

// executeFileProcessingChain 执行文件处理链：搜索 -> 批量读取
func executeFileProcessingChain(tb *toolbox.StandardToolBox) {
	ctx := context.Background()

	// Step 1: 搜索文件
	searchCall := &core.ToolCall{
		ID:       "chain3-step1",
		ToolName: "search_files",
		Input: map[string]interface{}{
			"path":        "/tmp",
			"pattern":     "*.txt",
			"max_results": 5,
		},
		Timestamp: time.Now(),
	}

	fmt.Println("  Step 1: 搜索 .txt 文件...")
	searchResult, err := tb.Execute(ctx, searchCall)
	if err != nil {
		fmt.Printf("    ❌ 失败: %v\n", err)
		return
	}

	searchData := searchResult.Result.Data.(map[string]interface{})
	matches := searchData["matches"].([]map[string]interface{})
	fmt.Printf("    ✓ 找到 %d 个文件 (耗时: %v)\n",
		len(matches), searchResult.Result.Duration)

	// Step 2: 批量读取前 3 个文件
	if len(matches) > 3 {
		matches = matches[:3]
	}

	fmt.Println("  Step 2: 批量读取前 3 个文件...")
	readCalls := make([]*core.ToolCall, 0, len(matches))

	for i, match := range matches {
		call := &core.ToolCall{
			ID:       fmt.Sprintf("chain3-step2-%d", i),
			ToolName: "read_file",
			Input: map[string]interface{}{
				"path": match["path"].(string),
			},
			Timestamp: time.Now(),
		}
		readCalls = append(readCalls, call)
	}

	startTime := time.Now()
	results, err := tb.ExecuteBatch(ctx, readCalls)
	batchDuration := time.Since(startTime)

	if err != nil {
		fmt.Printf("    ❌ 批量读取失败: %v\n", err)
		return
	}

	fmt.Printf("    ✓ 成功读取 %d 个文件 (耗时: %v)\n", len(results), batchDuration)
	for i, result := range results {
		if result.Result.Success {
			data := result.Result.Data.(map[string]interface{})
			fmt.Printf("      [%d] %s (%v 字节)\n",
				i+1, matches[i]["name"], data["size"])
		}
	}

	fmt.Printf("\n  工具链总耗时: %v\n",
		searchResult.Result.Duration+batchDuration)
}

// executeConditionalChain 执行条件分支链
func executeConditionalChain(tb *toolbox.StandardToolBox) {
	ctx := context.Background()

	// 创建测试数据
	testData := map[string]interface{}{
		"temperature": 25.5,
		"humidity":    60,
		"status":      "normal",
	}

	// Step 1: 写入测试数据
	dataJSON, _ := json.Marshal(testData)
	writeCall := &core.ToolCall{
		ID:       "chain4-step1",
		ToolName: "write_file",
		Input: map[string]interface{}{
			"path":    "/tmp/sensor_data.json",
			"content": string(dataJSON),
		},
		Timestamp: time.Now(),
	}

	fmt.Println("  Step 1: 写入传感器数据...")
	_, err := tb.Execute(ctx, writeCall)
	if err != nil {
		fmt.Printf("    ❌ 失败: %v\n", err)
		return
	}
	fmt.Println("    ✓ 成功")

	// Step 2: 读取数据
	readCall := &core.ToolCall{
		ID:       "chain4-step2",
		ToolName: "read_file",
		Input: map[string]interface{}{
			"path": "/tmp/sensor_data.json",
		},
		Timestamp: time.Now(),
	}

	fmt.Println("  Step 2: 读取传感器数据...")
	readResult, err := tb.Execute(ctx, readCall)
	if err != nil {
		fmt.Printf("    ❌ 失败: %v\n", err)
		return
	}
	fmt.Println("    ✓ 成功")

	// Step 3: 解析并检查条件
	data := readResult.Result.Data.(map[string]interface{})
	jsonCall := &core.ToolCall{
		ID:       "chain4-step3",
		ToolName: "json_parse",
		Input: map[string]interface{}{
			"json": data["content"].(string),
		},
		Timestamp: time.Now(),
	}

	fmt.Println("  Step 3: 解析数据并检查条件...")
	jsonResult, err := tb.Execute(ctx, jsonCall)
	if err != nil {
		fmt.Printf("    ❌ 失败: %v\n", err)
		return
	}

	parsed := jsonResult.Result.Data.(map[string]interface{})["parsed"].(map[string]interface{})
	temperature := parsed["temperature"].(float64)
	humidity := parsed["humidity"].(float64)

	fmt.Printf("    ✓ 解析成功\n")
	fmt.Printf("    温度: %.1f°C\n", temperature)
	fmt.Printf("    湿度: %.0f%%\n", humidity)

	// Step 4: 条件判断和响应
	fmt.Println("  Step 4: 条件判断...")
	var action string
	var actionData map[string]interface{}

	if temperature > 30 {
		action = "高温警告"
		actionData = map[string]interface{}{
			"level":   "warning",
			"message": "温度过高，需要降温",
		}
	} else if temperature < 10 {
		action = "低温警告"
		actionData = map[string]interface{}{
			"level":   "warning",
			"message": "温度过低，需要加热",
		}
	} else {
		action = "正常"
		actionData = map[string]interface{}{
			"level":   "info",
			"message": "温度正常",
		}
	}

	fmt.Printf("    判断结果: %s\n", action)
	fmt.Printf("    级别: %s\n", actionData["level"])
	fmt.Printf("    消息: %s\n", actionData["message"])

	// Step 5: 记录响应
	responseJSON, _ := json.MarshalIndent(actionData, "", "  ")
	responseCall := &core.ToolCall{
		ID:       "chain4-step5",
		ToolName: "write_file",
		Input: map[string]interface{}{
			"path":    "/tmp/sensor_response.json",
			"content": string(responseJSON),
		},
		Timestamp: time.Now(),
	}

	fmt.Println("  Step 5: 记录响应...")
	_, err = tb.Execute(ctx, responseCall)
	if err != nil {
		fmt.Printf("    ❌ 失败: %v\n", err)
		return
	}
	fmt.Println("    ✓ 响应已记录到文件")
}

// displayStatistics 显示统计信息
func displayStatistics(tb *toolbox.StandardToolBox) {
	stats := tb.Statistics()

	fmt.Printf("工具总数: %d\n", stats.TotalTools)
	fmt.Printf("总调用次数: %d\n", stats.TotalCalls)
	fmt.Printf("成功次数: %d (%.1f%%)\n",
		stats.SuccessfulCalls,
		float64(stats.SuccessfulCalls)/float64(stats.TotalCalls)*100)
	fmt.Printf("失败次数: %d\n", stats.FailedCalls)
	fmt.Printf("平均延迟: %.2f ms\n", stats.AverageLatency)

	fmt.Println("\n工具使用排名:")
	for tool, count := range stats.ToolUsage {
		fmt.Printf("  %s: %d 次\n", tool, count)
	}
}
