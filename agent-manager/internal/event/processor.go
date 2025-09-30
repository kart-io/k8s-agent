package event

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/agent-manager/internal/storage"
	"github.com/kart-io/k8s-agent/agent-manager/pkg/types"
)

// Processor handles event processing and routing
type Processor struct {
	store  *storage.PostgresStore
	cache  *storage.RedisStore
	nats   *nats.Conn
	logger *zap.Logger

	// Processing pipeline
	filters    []EventFilter
	enrichers  []EventEnricher
	aggregator *Aggregator
	publisher  *InternalPublisher

	// Metrics
	mu              sync.RWMutex
	eventsProcessed int64
	eventsFiltered  int64
	eventsFailed    int64
}

// EventFilter filters events
type EventFilter interface {
	ShouldProcess(event *types.Event) bool
}

// EventEnricher enriches events with additional context
type EventEnricher interface {
	Enrich(ctx context.Context, event *types.Event) error
}

// NewProcessor creates a new event processor
func NewProcessor(
	store *storage.PostgresStore,
	cache *storage.RedisStore,
	natsConn *nats.Conn,
	logger *zap.Logger,
) *Processor {
	p := &Processor{
		store:  store,
		cache:  cache,
		nats:   natsConn,
		logger: logger.With(zap.String("component", "event-processor")),
	}

	// Initialize components
	p.aggregator = NewAggregator(logger)
	p.publisher = NewInternalPublisher(natsConn, logger)

	// Setup default filters and enrichers
	p.filters = []EventFilter{
		&SeverityFilter{MinSeverity: "medium"},
		&DuplicateFilter{cache: cache, ttl: 5 * time.Minute},
	}

	p.enrichers = []EventEnricher{
		&ClusterEnricher{store: store},
	}

	return p
}

// ProcessEvent processes an incoming event
func (p *Processor) ProcessEvent(ctx context.Context, event *types.Event) error {
	// Apply filters
	for _, filter := range p.filters {
		if !filter.ShouldProcess(event) {
			p.mu.Lock()
			p.eventsFiltered++
			p.mu.Unlock()

			p.logger.Debug("Event filtered",
				zap.String("event_id", event.ID),
				zap.String("reason", "filter"))
			return nil
		}
	}

	// Enrich event
	for _, enricher := range p.enrichers {
		if err := enricher.Enrich(ctx, event); err != nil {
			p.logger.Warn("Failed to enrich event",
				zap.String("event_id", event.ID),
				zap.Error(err))
		}
	}

	// Set processing timestamp
	event.ProcessedAt = time.Now()

	// Save to database
	if err := p.store.SaveEvent(ctx, event); err != nil {
		p.mu.Lock()
		p.eventsFailed++
		p.mu.Unlock()
		return fmt.Errorf("failed to save event: %w", err)
	}

	// Update counters in Redis
	p.cache.IncrementEventCounter(ctx, event.ClusterID, event.Severity)

	// Check if event is critical
	if p.isCriticalEvent(event) {
		if err := p.handleCriticalEvent(ctx, event); err != nil {
			p.logger.Error("Failed to handle critical event",
				zap.String("event_id", event.ID),
				zap.Error(err))
		}
	}

	// Aggregate related events
	p.aggregator.Add(event)

	p.mu.Lock()
	p.eventsProcessed++
	p.mu.Unlock()

	p.logger.Info("Event processed",
		zap.String("event_id", event.ID),
		zap.String("cluster_id", event.ClusterID),
		zap.String("severity", event.Severity),
		zap.String("reason", event.Reason))

	return nil
}

// isCriticalEvent checks if event requires immediate attention
func (p *Processor) isCriticalEvent(event *types.Event) bool {
	criticalReasons := map[string]bool{
		"CrashLoopBackOff":      true,
		"OOMKilling":            true,
		"FailedScheduling":      true,
		"NodeNotReady":          true,
		"VolumeBindingFailed":   true,
		"ImagePullBackOff":      true,
		"DeadlineExceeded":      true,
	}

	return event.Severity == "critical" || criticalReasons[event.Reason]
}

// handleCriticalEvent handles critical events
func (p *Processor) handleCriticalEvent(ctx context.Context, event *types.Event) error {
	// Publish to internal event bus
	internalEvent := types.InternalEvent{
		Type:      string(types.InternalEventTypeCritical),
		ClusterID: event.ClusterID,
		Severity:  "critical",
		Payload: map[string]interface{}{
			"event_id":  event.ID,
			"reason":    event.Reason,
			"message":   event.Message,
			"namespace": event.Namespace,
			"labels":    event.Labels,
		},
		Timestamp: time.Now(),
	}

	return p.publisher.Publish(ctx, internalEvent)
}

// GetStatistics returns processor statistics
func (p *Processor) GetStatistics() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"events_processed": p.eventsProcessed,
		"events_filtered":  p.eventsFiltered,
		"events_failed":    p.eventsFailed,
		"aggregator_stats": p.aggregator.GetStatistics(),
	}
}

// SeverityFilter filters events by severity
type SeverityFilter struct {
	MinSeverity string
}

func (f *SeverityFilter) ShouldProcess(event *types.Event) bool {
	severityLevels := map[string]int{
		"low":      1,
		"medium":   2,
		"high":     3,
		"critical": 4,
	}

	minLevel := severityLevels[f.MinSeverity]
	eventLevel := severityLevels[event.Severity]

	return eventLevel >= minLevel
}

// DuplicateFilter filters duplicate events
type DuplicateFilter struct {
	cache *storage.RedisStore
	ttl   time.Duration
}

func (f *DuplicateFilter) ShouldProcess(event *types.Event) bool {
	ctx := context.Background()
	key := fmt.Sprintf("event:seen:%s:%s:%s", event.ClusterID, event.Reason, event.Labels["name"])

	// Try to set key (returns false if already exists)
	existed, err := f.cache.AcquireLock(ctx, key, f.ttl)
	if err != nil {
		// On error, allow processing
		return true
	}

	// If key was newly set, process the event
	return existed
}

// ClusterEnricher enriches events with cluster information
type ClusterEnricher struct {
	store *storage.PostgresStore
}

func (e *ClusterEnricher) Enrich(ctx context.Context, event *types.Event) error {
	cluster, err := e.store.GetCluster(ctx, event.ClusterID)
	if err != nil {
		// Cluster info not found, not critical
		return nil
	}

	// Add cluster metadata to event
	if event.RawData == nil {
		event.RawData = make(map[string]interface{})
	}

	event.RawData["cluster_name"] = cluster.Name
	event.RawData["cluster_environment"] = cluster.Environment
	event.RawData["cluster_region"] = cluster.Region

	return nil
}

// Aggregator aggregates related events
type Aggregator struct {
	logger *zap.Logger
	mu     sync.RWMutex
	groups map[string]*EventGroup
}

// EventGroup represents a group of related events
type EventGroup struct {
	Key        string
	Events     []*types.Event
	FirstSeen  time.Time
	LastSeen   time.Time
	Count      int
}

func NewAggregator(logger *zap.Logger) *Aggregator {
	return &Aggregator{
		logger: logger.With(zap.String("component", "event-aggregator")),
		groups: make(map[string]*EventGroup),
	}
}

// Add adds an event to aggregation
func (a *Aggregator) Add(event *types.Event) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Create grouping key
	key := fmt.Sprintf("%s:%s:%s",
		event.ClusterID,
		event.Namespace,
		event.Labels["name"])

	group, exists := a.groups[key]
	if !exists {
		group = &EventGroup{
			Key:       key,
			Events:    []*types.Event{},
			FirstSeen: event.Timestamp,
		}
		a.groups[key] = group
	}

	group.Events = append(group.Events, event)
	group.LastSeen = event.Timestamp
	group.Count++

	// Keep only recent events (last 10)
	if len(group.Events) > 10 {
		group.Events = group.Events[len(group.Events)-10:]
	}
}

// GetStatistics returns aggregator statistics
func (a *Aggregator) GetStatistics() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return map[string]interface{}{
		"active_groups": len(a.groups),
	}
}

// InternalPublisher publishes events to internal event bus
type InternalPublisher struct {
	conn   *nats.Conn
	logger *zap.Logger
}

func NewInternalPublisher(conn *nats.Conn, logger *zap.Logger) *InternalPublisher {
	return &InternalPublisher{
		conn:   conn,
		logger: logger.With(zap.String("component", "internal-publisher")),
	}
}

// Publish publishes an internal event
func (p *InternalPublisher) Publish(ctx context.Context, event types.InternalEvent) error {
	subject := fmt.Sprintf("internal.event.%s", event.Type)

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal internal event: %w", err)
	}

	if err := p.conn.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish internal event: %w", err)
	}

	p.logger.Info("Internal event published",
		zap.String("type", event.Type),
		zap.String("cluster_id", event.ClusterID),
		zap.String("subject", subject))

	return nil
}