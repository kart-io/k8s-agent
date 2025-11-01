package agent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/kart-io/logger/core"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kart-io/k8s-agent/internal/collect-agent/types"
)

func TestNewMetricsCollector(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	metricsChan := make(chan *types.Metrics, 10)
	logger := core.NewNoOpLogger(nil)

	collector := NewMetricsCollector(clientset, "test-cluster", metricsChan, logger)

	assert.NotNil(t, collector)
	assert.Equal(t, "test-cluster", collector.clusterID)
	assert.NotNil(t, collector.clientset)
	assert.NotNil(t, collector.metricsChan)
	assert.NotNil(t, collector.stopCh)
	assert.False(t, collector.running)
}

func TestMetricsCollector_Start(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	metricsChan := make(chan *types.Metrics, 10)
	logger := core.NewNoOpLogger(nil)

	collector := NewMetricsCollector(clientset, "test-cluster", metricsChan, logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start collector in background
	go collector.Start(ctx, 100*time.Millisecond)

	// Wait a bit for it to start
	time.Sleep(50 * time.Millisecond)

	collector.mu.RLock()
	running := collector.running
	collector.mu.RUnlock()

	assert.True(t, running)

	// Cleanup
	collector.Stop()
}

func TestMetricsCollector_StartTwice(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	metricsChan := make(chan *types.Metrics, 10)
	logger := core.NewNoOpLogger(nil)

	collector := NewMetricsCollector(clientset, "test-cluster", metricsChan, logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start collector
	go collector.Start(ctx, 100*time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	collector.mu.RLock()
	running := collector.running
	collector.mu.RUnlock()
	assert.True(t, running)

	// Try to start again - should not start a second time
	go collector.Start(ctx, 100*time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	// Should still be running (only once)
	collector.mu.RLock()
	running = collector.running
	collector.mu.RUnlock()
	assert.True(t, running)

	// Cleanup
	collector.Stop()
}

func TestMetricsCollector_Stop(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	metricsChan := make(chan *types.Metrics, 10)
	logger := core.NewNoOpLogger(nil)

	collector := NewMetricsCollector(clientset, "test-cluster", metricsChan, logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start collector
	go collector.Start(ctx, 100*time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	collector.mu.RLock()
	running := collector.running
	collector.mu.RUnlock()
	assert.True(t, running)

	// Stop collector
	collector.Stop()

	// Wait for stop to complete
	time.Sleep(50 * time.Millisecond)

	collector.mu.RLock()
	running = collector.running
	collector.mu.RUnlock()
	assert.False(t, running)
}

func TestMetricsCollector_StopWhenNotRunning(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	metricsChan := make(chan *types.Metrics, 10)
	logger := core.NewNoOpLogger(nil)

	collector := NewMetricsCollector(clientset, "test-cluster", metricsChan, logger)

	// Stop without starting
	collector.Stop()

	// Should not panic
	assert.False(t, collector.running)
}

func TestMetricsCollector_CollectMetrics(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	metricsChan := make(chan *types.Metrics, 10)
	logger := core.NewNoOpLogger(nil)

	collector := NewMetricsCollector(clientset, "test-cluster", metricsChan, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start collector with short interval
	go collector.Start(ctx, 100*time.Millisecond)

	// Wait for at least one metrics collection
	select {
	case metrics := <-metricsChan:
		assert.NotNil(t, metrics)
		assert.Equal(t, "test-cluster", metrics.ClusterID)
		assert.NotNil(t, metrics.Data)
		assert.NotZero(t, metrics.Timestamp)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for metrics")
	}

	// Cleanup
	collector.Stop()
}

func TestMetricsCollector_MultipleCollections(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	metricsChan := make(chan *types.Metrics, 10)
	logger := core.NewNoOpLogger(nil)

	collector := NewMetricsCollector(clientset, "test-cluster", metricsChan, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start collector with very short interval
	go collector.Start(ctx, 50*time.Millisecond)

	// Collect multiple metrics
	receivedCount := 0
	timeout := time.After(500 * time.Millisecond)

	for receivedCount < 3 {
		select {
		case metrics := <-metricsChan:
			assert.NotNil(t, metrics)
			assert.Equal(t, "test-cluster", metrics.ClusterID)
			receivedCount++
		case <-timeout:
			// We should have received at least some metrics
			assert.GreaterOrEqual(t, receivedCount, 1)
			collector.Stop()
			return
		}
	}

	assert.GreaterOrEqual(t, receivedCount, 3)
	collector.Stop()
}

func TestMetricsCollector_ContextCancellation(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	metricsChan := make(chan *types.Metrics, 10)
	logger := core.NewNoOpLogger(nil)

	collector := NewMetricsCollector(clientset, "test-cluster", metricsChan, logger)
	ctx, cancel := context.WithCancel(context.Background())

	// Start collector
	go collector.Start(ctx, 100*time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	collector.mu.RLock()
	running := collector.running
	collector.mu.RUnlock()
	assert.True(t, running)

	// Cancel context
	cancel()

	// Wait for collector to stop
	time.Sleep(150 * time.Millisecond)

	collector.mu.RLock()
	running = collector.running
	collector.mu.RUnlock()
	assert.False(t, running)
}

func TestMetricsCollector_IsRunning(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	metricsChan := make(chan *types.Metrics, 10)
	logger := core.NewNoOpLogger(nil)

	collector := NewMetricsCollector(clientset, "test-cluster", metricsChan, logger)

	// Initially not running
	collector.mu.RLock()
	running := collector.running
	collector.mu.RUnlock()
	assert.False(t, running)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start collector
	go collector.Start(ctx, 100*time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	// Should be running
	collector.mu.RLock()
	running = collector.running
	collector.mu.RUnlock()
	assert.True(t, running)

	// Stop collector
	collector.Stop()
	time.Sleep(50 * time.Millisecond)

	// Should not be running
	collector.mu.RLock()
	running = collector.running
	collector.mu.RUnlock()
	assert.False(t, running)
}

func TestMetricsCollector_MetricsDataStructure(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	metricsChan := make(chan *types.Metrics, 10)
	logger := core.NewNoOpLogger(nil)

	collector := NewMetricsCollector(clientset, "test-cluster", metricsChan, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Start collector
	go collector.Start(ctx, 100*time.Millisecond)

	// Wait for metrics
	select {
	case metrics := <-metricsChan:
		assert.NotNil(t, metrics)
		assert.Equal(t, "test-cluster", metrics.ClusterID)
		assert.NotNil(t, metrics.Data)
		assert.IsType(t, map[string]interface{}{}, metrics.Data)
		assert.NotZero(t, metrics.Timestamp)
		// Timestamp should be recent
		assert.WithinDuration(t, time.Now(), metrics.Timestamp, 5*time.Second)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for metrics")
	}

	// Cleanup
	collector.Stop()
}

// Benchmark tests
func BenchmarkMetricsCollector_CollectMetrics(b *testing.B) {
	clientset := fake.NewSimpleClientset()
	metricsChan := make(chan *types.Metrics, 1000)
	logger := core.NewNoOpLogger(nil)

	collector := NewMetricsCollector(clientset, "test-cluster", metricsChan, logger)

	// Drain channel in background
	go func() {
		for range metricsChan {
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.collectAndSendMetrics()
	}
}

func BenchmarkMetricsCollector_StartStop(b *testing.B) {
	clientset := fake.NewSimpleClientset()
	metricsChan := make(chan *types.Metrics, 1000)
	logger := core.NewNoOpLogger(nil)

	// Drain channel in background
	go func() {
		for range metricsChan {
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector := NewMetricsCollector(clientset, "test-cluster", metricsChan, logger)
		ctx, cancel := context.WithCancel(context.Background())

		go collector.Start(ctx, 1*time.Second)
		time.Sleep(10 * time.Millisecond)
		collector.Stop()
		cancel()
	}
}
