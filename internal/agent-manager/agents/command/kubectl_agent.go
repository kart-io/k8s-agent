package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kart-io/k8s-agent/internal/agent-manager/command"
	agentcore "github.com/kart-io/k8s-agent/pkg/agent/core"
	"github.com/kart-io/k8s-agent/pkg/types"
	"github.com/kart-io/logger/core"
)

// KubectlAgent kubectl 命令专用 Agent
// 提供更高级的 kubectl 命令抽象和结果解析
type KubectlAgent struct {
	*agentcore.BaseAgent
	commandAgent *CommandAgent
	logger       core.Logger
}

// NewKubectlAgent 创建 kubectl Agent
func NewKubectlAgent(dispatcher *command.Dispatcher, logger core.Logger) *KubectlAgent {
	return &KubectlAgent{
		BaseAgent: agentcore.NewBaseAgent(
			"kubectl-agent",
			"Specialized agent for kubectl command execution and result parsing",
			[]string{
				"kubectl_get",
				"kubectl_describe",
				"kubectl_logs",
				"kubectl_events",
				"result_parsing",
			},
		),
		commandAgent: NewCommandAgent(dispatcher, logger),
		logger:       logger.With("agent", "kubectl"),
	}
}

// Execute 执行 kubectl 命令
func (a *KubectlAgent) Execute(ctx context.Context, input *agentcore.AgentInput) (*agentcore.AgentOutput, error) {
	start := time.Now()

	// 解析 kubectl 参数
	clusterID, ok := input.Context["cluster_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid input: missing cluster_id")
	}

	action, ok := input.Context["action"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid input: missing action")
	}

	args, _ := input.Context["args"].([]string)
	namespace, _ := input.Context["namespace"].(string)

	// 构建 kubectl 命令
	cmd := &types.Command{
		ClusterID: clusterID,
		Type:      "diagnostic",
		Tool:      "kubectl",
		Action:    action,
		Args:      convertStringSlice(args),
		Timeout:   input.Options.Timeout,
	}

	// 如果指定了 namespace，添加到参数中
	if namespace != "" {
		cmd.Args = append([]interface{}{"-n", namespace}, cmd.Args...)
	}

	a.logger.Info("Executing kubectl command",
		"cluster_id", clusterID,
		"action", action,
		"namespace", namespace,
		"args", args)

	// 通过 CommandAgent 执行
	cmdInput := &agentcore.AgentInput{
		Task:        input.Task,
		Instruction: input.Instruction,
		Context: map[string]interface{}{
			"command": cmd,
		},
		Options:   input.Options,
		SessionID: input.SessionID,
		Timestamp: start,
	}

	output, err := a.commandAgent.Execute(ctx, cmdInput)
	if err != nil {
		return output, err
	}

	// 解析和增强结果
	if result, ok := output.Result.(map[string]interface{}); ok {
		parsed, parseErr := a.parseKubectlOutput(action, result)
		if parseErr != nil {
			a.logger.Warnw("Failed to parse kubectl output", "error", parseErr)
		} else {
			result["parsed"] = parsed
			output.Result = result
		}
	}

	return output, nil
}

// parseKubectlOutput 解析 kubectl 输出
func (a *KubectlAgent) parseKubectlOutput(action string, result map[string]interface{}) (interface{}, error) {
	rawOutput, ok := result["output"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid output format")
	}

	switch action {
	case "get":
		return a.parseGetOutput(rawOutput), nil
	case "describe":
		return a.parseDescribeOutput(rawOutput), nil
	case "logs":
		return a.parseLogsOutput(rawOutput), nil
	case "events":
		return a.parseEventsOutput(rawOutput), nil
	default:
		return rawOutput, nil
	}
}

// parseGetOutput 解析 kubectl get 输出
func (a *KubectlAgent) parseGetOutput(output string) map[string]interface{} {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return map[string]interface{}{
			"raw": output,
		}
	}

	// 解析表头和数据行
	headers := strings.Fields(lines[0])
	rows := make([]map[string]string, 0, len(lines)-1)

	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < len(headers) {
			continue
		}

		row := make(map[string]string)
		for i, header := range headers {
			if i < len(fields) {
				row[strings.ToLower(header)] = fields[i]
			}
		}
		rows = append(rows, row)
	}

	return map[string]interface{}{
		"headers": headers,
		"rows":    rows,
		"count":   len(rows),
	}
}

// parseDescribeOutput 解析 kubectl describe 输出
func (a *KubectlAgent) parseDescribeOutput(output string) map[string]interface{} {
	// 简单的键值对解析
	sections := make(map[string]interface{})
	lines := strings.Split(output, "\n")

	currentSection := "general"
	sectionContent := make([]string, 0)

	for _, line := range lines {
		if strings.HasSuffix(line, ":") && !strings.HasPrefix(line, " ") {
			// 新的 section
			if len(sectionContent) > 0 {
				sections[currentSection] = strings.Join(sectionContent, "\n")
				sectionContent = make([]string, 0)
			}
			currentSection = strings.TrimSuffix(strings.TrimSpace(line), ":")
		} else {
			sectionContent = append(sectionContent, line)
		}
	}

	if len(sectionContent) > 0 {
		sections[currentSection] = strings.Join(sectionContent, "\n")
	}

	return sections
}

// parseLogsOutput 解析 kubectl logs 输出
func (a *KubectlAgent) parseLogsOutput(output string) map[string]interface{} {
	lines := strings.Split(output, "\n")

	return map[string]interface{}{
		"total_lines": len(lines),
		"logs":        output,
		"lines":       lines,
	}
}

// parseEventsOutput 解析 kubectl events 输出
func (a *KubectlAgent) parseEventsOutput(output string) map[string]interface{} {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return map[string]interface{}{
			"raw": output,
		}
	}

	events := make([]map[string]string, 0)
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		event := map[string]string{
			"type":    fields[0],
			"reason":  fields[1],
			"age":     fields[2],
			"from":    fields[3],
			"message": strings.Join(fields[4:], " "),
		}
		events = append(events, event)
	}

	return map[string]interface{}{
		"events": events,
		"count":  len(events),
	}
}

// convertStringSlice 转换字符串切片为 interface{} 切片
func convertStringSlice(strs []string) []interface{} {
	result := make([]interface{}, len(strs))
	for i, s := range strs {
		result[i] = s
	}
	return result
}
