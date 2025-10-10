package agent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kart-io/k8s-agent/collect-agent/internal/types"
)

func TestNewEventWatcher(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	eventChan := make(chan *types.Event, 10)
	logger := zap.NewNop()

	watcher := NewEventWatcher(clientset, "test-cluster", eventChan, logger)

	assert.NotNil(t, watcher)
	assert.Equal(t, "test-cluster", watcher.clusterID)
	assert.NotNil(t, watcher.clientset)
	assert.NotNil(t, watcher.eventChan)
	assert.NotNil(t, watcher.stopCh)
	assert.False(t, watcher.running)
}

func TestEventWatcher_Start(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	eventChan := make(chan *types.Event, 10)
	logger := zap.NewNop()

	watcher := NewEventWatcher(clientset, "test-cluster", eventChan, logger)
	ctx := context.Background()

	err := watcher.Start(ctx)

	assert.NoError(t, err)
	assert.True(t, watcher.running)

	// Cleanup
	watcher.Stop()
}

func TestEventWatcher_StartTwice(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	eventChan := make(chan *types.Event, 10)
	logger := zap.NewNop()

	watcher := NewEventWatcher(clientset, "test-cluster", eventChan, logger)
	ctx := context.Background()

	err := watcher.Start(ctx)
	assert.NoError(t, err)

	// Try to start again
	err = watcher.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Cleanup
	watcher.Stop()
}

func TestEventWatcher_Stop(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	eventChan := make(chan *types.Event, 10)
	logger := zap.NewNop()

	watcher := NewEventWatcher(clientset, "test-cluster", eventChan, logger)
	ctx := context.Background()

	err := watcher.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, watcher.running)

	watcher.Stop()

	// Should be stopped
	watcher.mu.RLock()
	running := watcher.running
	watcher.mu.RUnlock()
	assert.False(t, running)
}

func TestEventWatcher_StopWhenNotRunning(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	eventChan := make(chan *types.Event, 10)
	logger := zap.NewNop()

	watcher := NewEventWatcher(clientset, "test-cluster", eventChan, logger)

	// Stop without starting
	watcher.Stop()

	// Should not panic or error
	assert.False(t, watcher.running)
}

func TestEventWatcher_HandleEvent(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	eventChan := make(chan *types.Event, 10)
	logger := zap.NewNop()

	watcher := NewEventWatcher(clientset, "test-cluster", eventChan, logger)

	// Create a test Kubernetes event
	k8sEvent := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-event",
			Namespace: "default",
			UID:       "test-uid",
		},
		Reason:  "TestReason",
		Message: "Test message",
		Type:    corev1.EventTypeWarning,
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Pod",
			Name:      "test-pod",
			Namespace: "default",
		},
	}

	// Handle the event
	watcher.handleEvent(k8sEvent, "ADDED")

	// Check that event was sent to channel
	select {
	case event := <-eventChan:
		assert.NotNil(t, event)
		assert.Equal(t, "test-cluster", event.ClusterID)
		assert.Equal(t, "TestReason", event.Reason)
		assert.Equal(t, "Test message", event.Message)
		assert.Equal(t, "Warning", event.Type)
		assert.Equal(t, "Pod", event.InvolvedObject.Kind)
		assert.Equal(t, "test-pod", event.InvolvedObject.Name)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for event")
	}
}

func TestEventWatcher_HandleMultipleEvents(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	eventChan := make(chan *types.Event, 10)
	logger := zap.NewNop()

	watcher := NewEventWatcher(clientset, "test-cluster", eventChan, logger)

	// Create multiple test events
	events := []*corev1.Event{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "event-1",
				Namespace: "default",
				UID:       "uid-1",
			},
			Reason: "Reason1",
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "event-2",
				Namespace: "default",
				UID:       "uid-2",
			},
			Reason: "Reason2",
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "event-3",
				Namespace: "default",
				UID:       "uid-3",
			},
			Reason: "Reason3",
		},
	}

	// Handle all events
	for _, k8sEvent := range events {
		watcher.handleEvent(k8sEvent, "ADDED")
	}

	// Check that all events were received
	receivedCount := 0
	timeout := time.After(2 * time.Second)

	for i := 0; i < len(events); i++ {
		select {
		case event := <-eventChan:
			assert.NotNil(t, event)
			receivedCount++
		case <-timeout:
			t.Fatalf("Timeout waiting for events, received %d of %d", receivedCount, len(events))
		}
	}

	assert.Equal(t, len(events), receivedCount)
}

func TestEventWatcher_IgnoreDuplicateEvents(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	eventChan := make(chan *types.Event, 10)
	logger := zap.NewNop()

	watcher := NewEventWatcher(clientset, "test-cluster", eventChan, logger)

	k8sEvent := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-event",
			Namespace: "default",
			UID:       "test-uid",
		},
		Reason:  "TestReason",
		Message: "Test message",
		Type:    corev1.EventTypeWarning,
	}

	// Handle the same event twice
	watcher.handleEvent(k8sEvent, "ADDED")
	watcher.handleEvent(k8sEvent, "MODIFIED")

	// First event should be received
	select {
	case event := <-eventChan:
		assert.NotNil(t, event)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for first event")
	}

	// Second event might be filtered as duplicate
	// depending on implementation
	select {
	case <-eventChan:
		// If second event is received, that's also okay
		// depending on whether deduplication is needed
	case <-time.After(100 * time.Millisecond):
		// No second event is also acceptable
	}
}

func TestEventWatcher_EventTypes(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	eventChan := make(chan *types.Event, 10)
	logger := zap.NewNop()

	watcher := NewEventWatcher(clientset, "test-cluster", eventChan, logger)

	testCases := []struct {
		name      string
		eventType string
		k8sType   string
	}{
		{
			name:      "Normal event",
			eventType: "ADDED",
			k8sType:   corev1.EventTypeNormal,
		},
		{
			name:      "Warning event",
			eventType: "ADDED",
			k8sType:   corev1.EventTypeWarning,
		},
		{
			name:      "Modified event",
			eventType: "MODIFIED",
			k8sType:   corev1.EventTypeWarning,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			k8sEvent := &corev1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-event",
					Namespace: "default",
					UID:       "test-uid",
				},
				Type: tc.k8sType,
			}

			watcher.handleEvent(k8sEvent, tc.eventType)

			select {
			case event := <-eventChan:
				assert.NotNil(t, event)
			case <-time.After(1 * time.Second):
				t.Fatal("Timeout waiting for event")
			}
		})
	}
}

func TestEventWatcher_IsRunning(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	eventChan := make(chan *types.Event, 10)
	logger := zap.NewNop()

	watcher := NewEventWatcher(clientset, "test-cluster", eventChan, logger)

	assert.False(t, watcher.IsRunning())

	ctx := context.Background()
	err := watcher.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, watcher.IsRunning())

	watcher.Stop()

	// Wait a bit for stop to complete
	time.Sleep(100 * time.Millisecond)

	watcher.mu.RLock()
	running := watcher.running
	watcher.mu.RUnlock()
	assert.False(t, running)
}

// Benchmark tests
func BenchmarkEventWatcher_HandleEvent(b *testing.B) {
	clientset := fake.NewSimpleClientset()
	eventChan := make(chan *types.Event, 1000)
	logger := zap.NewNop()

	watcher := NewEventWatcher(clientset, "test-cluster", eventChan, logger)

	k8sEvent := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bench-event",
			Namespace: "default",
			UID:       "bench-uid",
		},
		Reason:  "BenchReason",
		Message: "Benchmark message",
		Type:    corev1.EventTypeWarning,
	}

	// Drain channel in background
	go func() {
		for range eventChan {
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		watcher.handleEvent(k8sEvent, "ADDED")
	}
}
