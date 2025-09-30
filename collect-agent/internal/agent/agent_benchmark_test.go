package agent

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kart/k8s-agent/collect-agent/internal/types"
)

// BenchmarkEventProcessing benchmarks event processing throughput
func BenchmarkEventProcessing(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	clientset := fake.NewSimpleClientset()

	eventChan := make(chan *types.Event, 1000)
	watcher := NewEventWatcher(clientset, "test-cluster", eventChan, logger)

	// Create sample K8s event
	k8sEvent := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-event",
			Namespace: "default",
		},
		Type:   corev1.EventTypeWarning,
		Reason: "CrashLoopBackOff",
		Message: "Back-off restarting failed container",
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Pod",
			Name:      "test-pod",
			Namespace: "default",
		},
		FirstTimestamp: metav1.Time{Time: time.Now()},
		LastTimestamp:  metav1.Time{Time: time.Now()},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		watcher.handleEvent(k8sEvent, "ADDED")
	}
}

// BenchmarkMetricsCollection benchmarks metrics collection performance
func BenchmarkMetricsCollection(b *testing.B) {
	logger, _ := zap.NewDevelopment()

	// Create fake nodes
	nodes := make([]corev1.Node, 10)
	for i := 0; i < 10; i++ {
		nodes[i] = corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-" + string(rune(i)),
			},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
				},
			},
		}
	}

	clientset := fake.NewSimpleClientset()
	metricsChan := make(chan *types.Metrics, 100)
	collector := NewMetricsCollector(clientset, "test-cluster", metricsChan, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.collectAndSendMetrics()
	}
}

// BenchmarkCommandValidation benchmarks command validation performance
func BenchmarkCommandValidation(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	clientset := fake.NewSimpleClientset()
	executor := NewCommandExecutor(clientset, "test-cluster", logger)

	cmd := types.Command{
		ID:      "test-cmd",
		Type:    "diagnostic",
		Tool:    "kubectl",
		Action:  "get",
		Args:    []string{"pods", "-n", "default"},
		Timeout: 30 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.validateCommand(cmd)
	}
}

// BenchmarkChannelThroughput benchmarks channel communication throughput
func BenchmarkChannelThroughput(b *testing.B) {
	eventChan := make(chan *types.Event, 1000)

	// Producer
	go func() {
		for i := 0; i < b.N; i++ {
			eventChan <- &types.Event{
				ID:        "event-" + string(rune(i)),
				ClusterID: "test-cluster",
				Type:      "k8s_event",
				Severity:  "high",
			}
		}
		close(eventChan)
	}()

	// Consumer
	b.ResetTimer()
	count := 0
	for range eventChan {
		count++
	}
}

// BenchmarkAgentStatus benchmarks status retrieval
func BenchmarkAgentStatus(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	config := &types.AgentConfig{
		ClusterID:         "test-cluster",
		CentralEndpoint:   "nats://localhost:4222",
		ReconnectDelay:    5 * time.Second,
		HeartbeatInterval: 30 * time.Second,
		MetricsInterval:   60 * time.Second,
		BufferSize:        1000,
		MaxRetries:        10,
		LogLevel:          "info",
		EnableMetrics:     true,
		EnableEvents:      true,
	}

	// Create fake agent (simplified for benchmark)
	agent := &Agent{
		config:      config,
		clusterID:   config.ClusterID,
		eventChan:   make(chan *types.Event, config.BufferSize),
		metricsChan: make(chan *types.Metrics, 100),
		commandChan: make(chan *types.Command, 100),
		resultChan:  make(chan *types.CommandResult, 100),
		stopCh:      make(chan struct{}),
		running:     true,
		startTime:   time.Now(),
		logger:      logger,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agent.GetStatus()
	}
}

// BenchmarkConcurrentEventProcessing benchmarks concurrent event processing
func BenchmarkConcurrentEventProcessing(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	clientset := fake.NewSimpleClientset()
	eventChan := make(chan *types.Event, 10000)

	watcher := NewEventWatcher(clientset, "test-cluster", eventChan, logger)

	k8sEvent := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-event",
			Namespace: "default",
		},
		Type:   corev1.EventTypeWarning,
		Reason: "CrashLoopBackOff",
		Message: "Back-off restarting failed container",
		InvolvedObject: corev1.ObjectReference{
			Kind: "Pod",
			Name: "test-pod",
		},
		FirstTimestamp: metav1.Time{Time: time.Now()},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			watcher.handleEvent(k8sEvent, "ADDED")
		}
	})
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		event := &types.Event{
			ID:        "event-id",
			ClusterID: "test-cluster",
			Type:      "k8s_event",
			Source:    "kubernetes",
			Namespace: "default",
			Severity:  "high",
			Reason:    "CrashLoopBackOff",
			Message:   "Container failed",
			Timestamp: time.Now(),
			Labels: map[string]string{
				"kind": "Pod",
				"name": "test-pod",
			},
			RawData: map[string]interface{}{
				"count": 5,
			},
		}
		_ = event
	}
}