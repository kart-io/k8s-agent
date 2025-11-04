package k8s_tool

import (
	"context"
	"fmt"
	"log"
	"time"
)

// K8sTool K8s 工具实现.
type K8sTool struct {
	config *ToolConfig
	// 在实际实现中，这里会包含 client-go 的 clientset
	// clientset kubernetes.Interface
}

// NewK8sTool 创建新的 K8s 工具.
func NewK8sTool(config *ToolConfig) (*K8sTool, error) {
	if config == nil {
		config = DefaultToolConfig()
	}

	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	tool := &K8sTool{
		config: config,
	}

	// 在实际实现中，这里会初始化 K8s client
	// clientset, err := kubernetes.NewForConfig(...)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	// }
	// tool.clientset = clientset

	log.Printf("K8sTool initialized with kubeconfig: %s", config.KubeconfigPath)

	return tool, nil
}

// Name 返回工具名称.
func (t *K8sTool) Name() string {
	return "k8s_tool"
}

// Description 返回工具描述.
func (t *K8sTool) Description() string {
	return "Kubernetes cluster interaction tool for querying pods, deployments, services, logs, events, and metrics"
}

// Execute 执行工具操作.
func (t *K8sTool) Execute(ctx context.Context, input *ToolInput) (*ToolOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}

	// 应用超时
	if t.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, t.config.Timeout)
		defer cancel()
	}

	start := time.Now()

	// 验证输入
	if err := validateInput(input); err != nil {
		return t.createErrorOutput(input.Action, err), nil
	}

	// 应用默认命名空间
	if input.Namespace == "" {
		input.Namespace = t.config.DefaultNamespace
	}

	log.Printf("Executing K8s tool action: %s on %s/%s in namespace: %s",
		input.Action, input.ResourceType, input.ResourceName, input.Namespace)

	// 根据操作类型执行
	var data interface{}
	var err error

	switch input.Action {
	case "get":
		data, err = t.getResource(ctx, input)
	case "describe":
		data, err = t.describeResource(ctx, input)
	case "logs":
		data, err = t.getLogs(ctx, input)
	case "events":
		data, err = t.getEvents(ctx, input)
	case "list":
		data, err = t.listResources(ctx, input)
	case "metrics":
		data, err = t.getMetrics(ctx, input)
	case "top":
		data, err = t.getTop(ctx, input)
	default:
		return t.createErrorOutput(input.Action,
			fmt.Errorf("unsupported action: %s", input.Action)), nil
	}

	latency := time.Since(start)

	if err != nil {
		log.Printf("K8s tool action failed: %v (latency: %v)", err, latency)
		output := t.createErrorOutput(input.Action, err)
		output.Latency = latency
		return output, nil
	}

	log.Printf("K8s tool action completed successfully (latency: %v)", latency)

	return &ToolOutput{
		Success:   true,
		Data:      data,
		ToolName:  t.Name(),
		Action:    input.Action,
		Latency:   latency,
		Timestamp: time.Now(),
	}, nil
}

// getResource 获取单个资源.
func (t *K8sTool) getResource(ctx context.Context, input *ToolInput) (interface{}, error) {
	// 这是一个模拟实现
	// 在实际实现中，会调用 client-go:
	// switch input.ResourceType {
	// case "pod":
	//     pod, err := t.clientset.CoreV1().Pods(input.Namespace).Get(ctx, input.ResourceName, metav1.GetOptions{})
	//     return convertPodInfo(pod), err
	// ...
	// }

	switch input.ResourceType {
	case "pod":
		return t.getMockPodInfo(input.Namespace, input.ResourceName), nil
	case "deployment":
		return t.getMockDeploymentInfo(input.Namespace, input.ResourceName), nil
	case "service":
		return t.getMockServiceInfo(input.Namespace, input.ResourceName), nil
	case "node":
		return t.getMockNodeInfo(input.ResourceName), nil
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", input.ResourceType)
	}
}

// describeResource 描述资源详情.
func (t *K8sTool) describeResource(ctx context.Context, input *ToolInput) (interface{}, error) {
	// describe 操作返回更详细的信息
	return t.getResource(ctx, input)
}

// getLogs 获取 Pod 日志.
func (t *K8sTool) getLogs(ctx context.Context, input *ToolInput) (interface{}, error) {
	if input.ResourceType != "pod" {
		return nil, fmt.Errorf("logs only supported for pods")
	}

	// 解析日志选项
	opts := t.parseLogsOptions(input.Parameters)

	// 模拟实现
	// 在实际实现中:
	// req := t.clientset.CoreV1().Pods(input.Namespace).GetLogs(input.ResourceName, &corev1.PodLogOptions{
	//     Container:    opts.Container,
	//     Follow:       opts.Follow,
	//     Previous:     opts.Previous,
	//     TailLines:    &opts.TailLines,
	//     SinceSeconds: &opts.SinceSeconds,
	//     Timestamps:   opts.Timestamps,
	// })
	// return req.Stream(ctx)

	return map[string]interface{}{
		"pod":       input.ResourceName,
		"namespace": input.Namespace,
		"container": opts.Container,
		"logs":      t.getMockLogs(input.ResourceName, opts),
	}, nil
}

// getEvents 获取资源事件.
func (t *K8sTool) getEvents(ctx context.Context, input *ToolInput) (interface{}, error) {
	// 模拟实现
	// 在实际实现中:
	// events, err := t.clientset.CoreV1().Events(input.Namespace).List(ctx, metav1.ListOptions{
	//     FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=%s",
	//         input.ResourceName, getKind(input.ResourceType)),
	// })

	return t.getMockEvents(input.Namespace, input.ResourceName), nil
}

// listResources 列出资源.
func (t *K8sTool) listResources(ctx context.Context, input *ToolInput) (interface{}, error) {
	// 模拟实现
	// 在实际实现中:
	// pods, err := t.clientset.CoreV1().Pods(input.Namespace).List(ctx, metav1.ListOptions{
	//     LabelSelector: input.LabelSelector,
	//     FieldSelector: input.FieldSelector,
	//     Limit:         int64(input.Limit),
	// })

	switch input.ResourceType {
	case "pod":
		return t.getMockPodList(input.Namespace), nil
	case "deployment":
		return t.getMockDeploymentList(input.Namespace), nil
	default:
		return nil, fmt.Errorf("list not implemented for resource type: %s", input.ResourceType)
	}
}

// getMetrics 获取资源指标.
func (t *K8sTool) getMetrics(ctx context.Context, input *ToolInput) (interface{}, error) {
	// 模拟实现
	// 在实际实现中，会使用 metrics-server API

	return t.getMockMetrics(input.Namespace, input.ResourceName, input.ResourceType), nil
}

// getTop 获取资源使用情况 (类似 kubectl top).
func (t *K8sTool) getTop(ctx context.Context, input *ToolInput) (interface{}, error) {
	// 模拟实现
	return t.getMetrics(ctx, input)
}

// createErrorOutput 创建错误输出.
func (t *K8sTool) createErrorOutput(action string, err error) *ToolOutput {
	return &ToolOutput{
		Success:   false,
		Data:      nil,
		ErrorMsg:  err.Error(),
		ToolName:  t.Name(),
		Action:    action,
		Timestamp: time.Now(),
	}
}

// parseLogsOptions 解析日志选项.
func (t *K8sTool) parseLogsOptions(params map[string]string) *LogsOptions {
	opts := &LogsOptions{
		TailLines: t.config.MaxLogLines,
	}

	if container, ok := params["container"]; ok {
		opts.Container = container
	}
	if previous, ok := params["previous"]; ok && previous == "true" {
		opts.Previous = true
	}
	if timestamps, ok := params["timestamps"]; ok && timestamps == "true" {
		opts.Timestamps = true
	}

	return opts
}

// validateConfig 验证配置.
func validateConfig(config *ToolConfig) error {
	if config.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	if config.MaxLogLines < 0 {
		return fmt.Errorf("max_log_lines cannot be negative")
	}

	return nil
}

// validateInput 验证输入.
func validateInput(input *ToolInput) error {
	if input.Action == "" {
		return fmt.Errorf("action is required")
	}

	// 验证操作是否支持
	validAction := false
	for _, action := range SupportedActions {
		if input.Action == action {
			validAction = true
			break
		}
	}
	if !validAction {
		return fmt.Errorf("unsupported action: %s", input.Action)
	}

	// 某些操作需要资源类型
	if input.Action != "list" && input.ResourceType == "" {
		return fmt.Errorf("resource_type is required for action: %s", input.Action)
	}

	// 验证资源类型
	if input.ResourceType != "" {
		validType := false
		for _, rt := range SupportedResourceTypes {
			if input.ResourceType == rt {
				validType = true
				break
			}
		}
		if !validType {
			return fmt.Errorf("unsupported resource type: %s", input.ResourceType)
		}
	}

	// 某些操作需要资源名称
	needsResourceName := []string{"get", "describe", "logs", "events"}
	for _, action := range needsResourceName {
		if input.Action == action && input.ResourceName == "" {
			return fmt.Errorf("resource_name is required for action: %s", input.Action)
		}
	}

	return nil
}

// DefaultToolConfig 返回默认配置.
func DefaultToolConfig() *ToolConfig {
	return &ToolConfig{
		KubeconfigPath:   "", // 使用默认路径
		DefaultNamespace: "default",
		Timeout:          30 * time.Second,
		MaxLogLines:      100,
		EnableCache:      false,
		CacheTTL:         5 * time.Minute,
	}
}
