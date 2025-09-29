package agent

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/kart/k8s-agent/collect-agent/internal/types"
)

// MetricsCollector collects cluster metrics and sends them to the metrics channel
type MetricsCollector struct {
	clientset       kubernetes.Interface
	metricsClientset *metricsv1beta1.Clientset
	clusterID       string
	metricsChan     chan<- *types.Metrics
	stopCh          chan struct{}
	running         bool
	mu              sync.RWMutex
	logger          *zap.Logger
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(clientset kubernetes.Interface, clusterID string, metricsChan chan<- *types.Metrics, logger *zap.Logger) *MetricsCollector {
	// Try to create metrics clientset, but don't fail if metrics server is not available
	var metricsClientset *metricsv1beta1.Clientset
	// Note: In production, you would get this from the same config as the main clientset
	// For now, we'll collect basic metrics without the metrics server

	return &MetricsCollector{
		clientset:       clientset,
		metricsClientset: metricsClientset,
		clusterID:       clusterID,
		metricsChan:     metricsChan,
		stopCh:          make(chan struct{}),
		logger:          logger.With(zap.String("component", "metrics-collector")),
	}
}

// Start begins collecting metrics at the specified interval
func (mc *MetricsCollector) Start(ctx context.Context, interval time.Duration) {
	mc.mu.Lock()
	if mc.running {
		mc.mu.Unlock()
		return
	}
	mc.running = true
	mc.mu.Unlock()

	mc.logger.Info("Starting metrics collector",
		zap.String("cluster_id", mc.clusterID),
		zap.Duration("interval", interval))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Collect initial metrics
	mc.collectAndSendMetrics()

	for {
		select {
		case <-ctx.Done():
			mc.Stop()
			return
		case <-mc.stopCh:
			return
		case <-ticker.C:
			mc.collectAndSendMetrics()
		}
	}
}

// Stop stops the metrics collector
func (mc *MetricsCollector) Stop() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if !mc.running {
		return
	}

	mc.logger.Info("Stopping metrics collector")
	close(mc.stopCh)
	mc.running = false
	mc.logger.Info("Metrics collector stopped")
}

// collectAndSendMetrics collects various cluster metrics and sends them
func (mc *MetricsCollector) collectAndSendMetrics() {
	metrics := &types.Metrics{
		ClusterID: mc.clusterID,
		Timestamp: time.Now(),
		Data:      make(map[string]interface{}),
	}

	// Collect cluster-level metrics
	mc.collectClusterMetrics(metrics)

	// Collect node metrics
	mc.collectNodeMetrics(metrics)

	// Collect pod metrics
	mc.collectPodMetrics(metrics)

	// Collect namespace metrics
	mc.collectNamespaceMetrics(metrics)

	// Send metrics
	select {
	case mc.metricsChan <- metrics:
		mc.logger.Debug("Metrics sent", zap.String("cluster_id", mc.clusterID))
	default:
		mc.logger.Warn("Metrics channel full, dropping metrics")
	}
}

// collectClusterMetrics collects cluster-level metrics
func (mc *MetricsCollector) collectClusterMetrics(metrics *types.Metrics) {
	ctx := context.Background()

	// Get cluster version
	version, err := mc.clientset.Discovery().ServerVersion()
	if err != nil {
		mc.logger.Warn("Failed to get server version", zap.Error(err))
	} else {
		metrics.Data["cluster"] = map[string]interface{}{
			"version":    version.String(),
			"git_version": version.GitVersion,
			"platform":  version.Platform,
		}
	}

	// Count total nodes
	nodes, err := mc.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		mc.logger.Warn("Failed to list nodes", zap.Error(err))
		return
	}

	nodeStats := mc.analyzeNodes(nodes.Items)
	metrics.Data["nodes"] = nodeStats
}

// collectNodeMetrics collects metrics for each node
func (mc *MetricsCollector) collectNodeMetrics(metrics *types.Metrics) {
	ctx := context.Background()

	nodes, err := mc.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		mc.logger.Warn("Failed to list nodes for metrics", zap.Error(err))
		return
	}

	nodeMetrics := make(map[string]interface{})
	for _, node := range nodes.Items {
		nodeMetrics[node.Name] = mc.getNodeMetrics(&node)
	}

	metrics.Data["node_details"] = nodeMetrics
}

// collectPodMetrics collects pod-level metrics
func (mc *MetricsCollector) collectPodMetrics(metrics *types.Metrics) {
	ctx := context.Background()

	pods, err := mc.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		mc.logger.Warn("Failed to list pods", zap.Error(err))
		return
	}

	podStats := mc.analyzePods(pods.Items)
	metrics.Data["pods"] = podStats
}

// collectNamespaceMetrics collects namespace-level metrics
func (mc *MetricsCollector) collectNamespaceMetrics(metrics *types.Metrics) {
	ctx := context.Background()

	namespaces, err := mc.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		mc.logger.Warn("Failed to list namespaces", zap.Error(err))
		return
	}

	nsStats := make(map[string]interface{})
	for _, ns := range namespaces.Items {
		nsStats[ns.Name] = mc.getNamespaceMetrics(&ns)
	}

	metrics.Data["namespaces"] = nsStats
}

// analyzeNodes analyzes the list of nodes and returns statistics
func (mc *MetricsCollector) analyzeNodes(nodes []corev1.Node) map[string]interface{} {
	total := len(nodes)
	ready := 0
	notReady := 0
	schedulable := 0

	capacityStats := map[string]int64{
		"cpu_cores":       0,
		"memory_bytes":    0,
		"pods":           0,
		"ephemeral_storage": 0,
	}

	for _, node := range nodes {
		// Check node conditions
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				if condition.Status == corev1.ConditionTrue {
					ready++
				} else {
					notReady++
				}
				break
			}
		}

		// Check if node is schedulable
		if !node.Spec.Unschedulable {
			schedulable++
		}

		// Aggregate capacity
		if cpu := node.Status.Capacity.Cpu(); cpu != nil {
			capacityStats["cpu_cores"] += cpu.MilliValue()
		}
		if memory := node.Status.Capacity.Memory(); memory != nil {
			capacityStats["memory_bytes"] += memory.Value()
		}
		if pods := node.Status.Capacity.Pods(); pods != nil {
			capacityStats["pods"] += pods.Value()
		}
		if storage := node.Status.Capacity.StorageEphemeral(); storage != nil {
			capacityStats["ephemeral_storage"] += storage.Value()
		}
	}

	// Convert CPU from milli-cores to cores
	capacityStats["cpu_cores"] = capacityStats["cpu_cores"] / 1000

	return map[string]interface{}{
		"total":       total,
		"ready":       ready,
		"not_ready":   notReady,
		"schedulable": schedulable,
		"capacity":    capacityStats,
	}
}

// getNodeMetrics gets detailed metrics for a specific node
func (mc *MetricsCollector) getNodeMetrics(node *corev1.Node) map[string]interface{} {
	nodeMetrics := map[string]interface{}{
		"name":      node.Name,
		"ready":     false,
		"schedulable": !node.Spec.Unschedulable,
		"labels":    node.Labels,
		"annotations": node.Annotations,
	}

	// Check node conditions
	conditions := make(map[string]interface{})
	for _, condition := range node.Status.Conditions {
		conditions[string(condition.Type)] = map[string]interface{}{
			"status":  string(condition.Status),
			"reason":  condition.Reason,
			"message": condition.Message,
		}

		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			nodeMetrics["ready"] = true
		}
	}
	nodeMetrics["conditions"] = conditions

	// Node info
	nodeMetrics["node_info"] = map[string]interface{}{
		"architecture":    node.Status.NodeInfo.Architecture,
		"os_image":        node.Status.NodeInfo.OSImage,
		"kernel_version":  node.Status.NodeInfo.KernelVersion,
		"kubelet_version": node.Status.NodeInfo.KubeletVersion,
		"runtime_version": node.Status.NodeInfo.ContainerRuntimeVersion,
	}

	// Capacity and allocatable
	if node.Status.Capacity != nil {
		capacity := make(map[string]interface{})
		for k, v := range node.Status.Capacity {
			capacity[string(k)] = v.String()
		}
		nodeMetrics["capacity"] = capacity
	}

	if node.Status.Allocatable != nil {
		allocatable := make(map[string]interface{})
		for k, v := range node.Status.Allocatable {
			allocatable[string(k)] = v.String()
		}
		nodeMetrics["allocatable"] = allocatable
	}

	return nodeMetrics
}

// analyzePods analyzes the list of pods and returns statistics
func (mc *MetricsCollector) analyzePods(pods []corev1.Pod) map[string]interface{} {
	total := len(pods)
	phaseCount := make(map[string]int)
	namespaceCount := make(map[string]int)
	restartCount := 0

	for _, pod := range pods {
		// Count by phase
		phase := string(pod.Status.Phase)
		phaseCount[phase]++

		// Count by namespace
		namespaceCount[pod.Namespace]++

		// Count restarts
		for _, containerStatus := range pod.Status.ContainerStatuses {
			restartCount += int(containerStatus.RestartCount)
		}
	}

	return map[string]interface{}{
		"total":           total,
		"by_phase":        phaseCount,
		"by_namespace":    namespaceCount,
		"total_restarts":  restartCount,
	}
}

// getNamespaceMetrics gets detailed metrics for a specific namespace
func (mc *MetricsCollector) getNamespaceMetrics(namespace *corev1.Namespace) map[string]interface{} {
	ctx := context.Background()

	// Count pods in this namespace
	pods, err := mc.clientset.CoreV1().Pods(namespace.Name).List(ctx, metav1.ListOptions{})
	podCount := 0
	if err == nil {
		podCount = len(pods.Items)
	}

	// Count services
	services, err := mc.clientset.CoreV1().Services(namespace.Name).List(ctx, metav1.ListOptions{})
	serviceCount := 0
	if err == nil {
		serviceCount = len(services.Items)
	}

	// Count configmaps
	configMaps, err := mc.clientset.CoreV1().ConfigMaps(namespace.Name).List(ctx, metav1.ListOptions{})
	configMapCount := 0
	if err == nil {
		configMapCount = len(configMaps.Items)
	}

	// Count secrets
	secrets, err := mc.clientset.CoreV1().Secrets(namespace.Name).List(ctx, metav1.ListOptions{})
	secretCount := 0
	if err == nil {
		secretCount = len(secrets.Items)
	}

	return map[string]interface{}{
		"name":        namespace.Name,
		"status":      string(namespace.Status.Phase),
		"created":     namespace.CreationTimestamp.Time,
		"labels":      namespace.Labels,
		"annotations": namespace.Annotations,
		"resources": map[string]int{
			"pods":        podCount,
			"services":    serviceCount,
			"configmaps":  configMapCount,
			"secrets":     secretCount,
		},
	}
}