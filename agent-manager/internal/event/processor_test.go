package event

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/agent-manager/pkg/types"
)

// MockEventFilter is a mock for EventFilter
type MockEventFilter struct {
	mock.Mock
}

func (m *MockEventFilter) ShouldProcess(event *types.Event) bool {
	args := m.Called(event)
	return args.Bool(0)
}

// MockEventEnricher is a mock for EventEnricher
type MockEventEnricher struct {
	mock.Mock
}

func (m *MockEventEnricher) Enrich(ctx context.Context, event *types.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestNewProcessor(t *testing.T) {
	logger := zap.NewNop()
	processor := NewProcessor(nil, nil, nil, logger)

	assert.NotNil(t, processor)
	assert.NotNil(t, processor.aggregator)
	assert.NotNil(t, processor.publisher)
	assert.Len(t, processor.filters, 2)   // Default filters
	assert.Len(t, processor.enrichers, 1) // Default enrichers
}

func TestProcessEvent_WithFilters(t *testing.T) {
	logger := zap.NewNop()
	processor := NewProcessor(nil, nil, nil, logger)
	ctx := context.Background()

	event := &types.Event{
		ID:        "event-001",
		Type:      "pod_failure",
		Severity:  "high",
		ClusterID: "cluster-001",
		Message:   "Pod crashed",
		Timestamp: time.Now(),
	}

	// Mock filter that accepts event
	mockFilter := new(MockEventFilter)
	mockFilter.On("ShouldProcess", event).Return(true)
	processor.filters = []EventFilter{mockFilter}

	// Mock enricher
	mockEnricher := new(MockEventEnricher)
	mockEnricher.On("Enrich", ctx, event).Return(nil)
	processor.enrichers = []EventEnricher{mockEnricher}

	err := processor.ProcessEvent(ctx, event)

	assert.NoError(t, err)
	mockFilter.AssertExpectations(t)
	mockEnricher.AssertExpectations(t)
}

func TestProcessEvent_FilteredOut(t *testing.T) {
	logger := zap.NewNop()
	processor := NewProcessor(nil, nil, nil, logger)
	ctx := context.Background()

	event := &types.Event{
		ID:        "event-002",
		Type:      "pod_warning",
		Severity:  "low",
		ClusterID: "cluster-001",
		Message:   "Low severity warning",
		Timestamp: time.Now(),
	}

	// Mock filter that rejects event
	mockFilter := new(MockEventFilter)
	mockFilter.On("ShouldProcess", event).Return(false)
	processor.filters = []EventFilter{mockFilter}

	err := processor.ProcessEvent(ctx, event)

	assert.NoError(t, err)
	mockFilter.AssertExpectations(t)
	// Verify metrics
	processor.mu.RLock()
	assert.Equal(t, int64(1), processor.eventsFiltered)
	processor.mu.RUnlock()
}

func TestSeverityFilter(t *testing.T) {
	tests := []struct {
		name        string
		minSeverity string
		event       *types.Event
		expected    bool
	}{
		{
			name:        "high severity passes medium filter",
			minSeverity: "medium",
			event: &types.Event{
				Severity: "high",
			},
			expected: true,
		},
		{
			name:        "low severity filtered by medium filter",
			minSeverity: "medium",
			event: &types.Event{
				Severity: "low",
			},
			expected: false,
		},
		{
			name:        "critical severity passes high filter",
			minSeverity: "high",
			event: &types.Event{
				Severity: "critical",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &SeverityFilter{MinSeverity: tt.minSeverity}
			result := filter.ShouldProcess(tt.event)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAggregator(t *testing.T) {
	logger := zap.NewNop()
	aggregator := NewAggregator(logger)

	event1 := &types.Event{
		ID:        "event-001",
		Type:      "pod_failure",
		Severity:  "high",
		ClusterID: "cluster-001",
		Message:   "Pod crashed",
		Timestamp: time.Now(),
	}

	event2 := &types.Event{
		ID:        "event-002",
		Type:      "pod_failure",
		Severity:  "high",
		ClusterID: "cluster-001",
		Message:   "Pod crashed",
		Timestamp: time.Now().Add(1 * time.Second),
	}

	// Add events to aggregator
	aggregator.AddEvent(event1)
	aggregator.AddEvent(event2)

	// Check if events are correlated
	assert.NotNil(t, aggregator)
	// More detailed testing would require exposing internal state
}

func TestGetMetrics(t *testing.T) {
	logger := zap.NewNop()
	processor := NewProcessor(nil, nil, nil, logger)

	// Simulate processing
	processor.mu.Lock()
	processor.eventsProcessed = 100
	processor.eventsFiltered = 20
	processor.eventsFailed = 5
	processor.mu.Unlock()

	metrics := processor.GetMetrics()

	assert.Equal(t, int64(100), metrics["events_processed"])
	assert.Equal(t, int64(20), metrics["events_filtered"])
	assert.Equal(t, int64(5), metrics["events_failed"])
}

// Integration-like test (can be moved to integration tests)
func TestProcessEvent_FullPipeline(t *testing.T) {
	logger := zap.NewNop()
	processor := NewProcessor(nil, nil, nil, logger)
	ctx := context.Background()

	// Clear default filters/enrichers for controlled test
	processor.filters = nil
	processor.enrichers = nil

	event := &types.Event{
		ID:        "event-integration-001",
		Type:      "node_failure",
		Severity:  "critical",
		ClusterID: "cluster-001",
		Message:   "Node is down",
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"node_name": "worker-01",
			"zone":      "us-west-2a",
		},
	}

	err := processor.ProcessEvent(ctx, event)

	assert.NoError(t, err)

	// Verify metrics
	processor.mu.RLock()
	assert.Equal(t, int64(1), processor.eventsProcessed)
	processor.mu.RUnlock()
}

// Benchmark tests
func BenchmarkProcessEvent(b *testing.B) {
	logger := zap.NewNop()
	processor := NewProcessor(nil, nil, nil, logger)
	ctx := context.Background()

	event := &types.Event{
		ID:        "benchmark-event",
		Type:      "pod_failure",
		Severity:  "high",
		ClusterID: "cluster-001",
		Message:   "Benchmark test",
		Timestamp: time.Now(),
	}

	// Clear filters/enrichers for pure processing benchmark
	processor.filters = nil
	processor.enrichers = nil

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processor.ProcessEvent(ctx, event)
	}
}

func BenchmarkSeverityFilter(b *testing.B) {
	filter := &SeverityFilter{MinSeverity: "medium"}

	event := &types.Event{
		Severity: "high",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filter.ShouldProcess(event)
	}
}
