package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/kart/k8s-agent/collect-agent/internal/types"
)

// EventWatcher watches Kubernetes events and sends them to the event channel
type EventWatcher struct {
	clientset   kubernetes.Interface
	clusterID   string
	eventChan   chan<- *types.Event
	stopCh      chan struct{}
	wg          sync.WaitGroup
	running     bool
	mu          sync.RWMutex
	lastEventID string
	logger      *zap.Logger
}

// NewEventWatcher creates a new event watcher
func NewEventWatcher(clientset kubernetes.Interface, clusterID string, eventChan chan<- *types.Event, logger *zap.Logger) *EventWatcher {
	return &EventWatcher{
		clientset: clientset,
		clusterID: clusterID,
		eventChan: eventChan,
		stopCh:    make(chan struct{}),
		logger:    logger.With(zap.String("component", "event-watcher")),
	}
}

// Start begins watching for Kubernetes events
func (ew *EventWatcher) Start(ctx context.Context) error {
	ew.mu.Lock()
	if ew.running {
		ew.mu.Unlock()
		return fmt.Errorf("event watcher already running")
	}
	ew.running = true
	ew.mu.Unlock()

	ew.logger.Info("Starting event watcher", zap.String("cluster_id", ew.clusterID))

	watchlist := cache.NewListWatchFromClient(
		ew.clientset.CoreV1().RESTClient(),
		"events",
		metav1.NamespaceAll,
		fields.Everything(),
	)

	_, controller := cache.NewInformer(
		watchlist,
		&corev1.Event{},
		time.Second*10,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				ew.handleEvent(obj, "ADDED")
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				ew.handleEvent(newObj, "MODIFIED")
			},
			DeleteFunc: func(obj interface{}) {
				ew.handleEvent(obj, "DELETED")
			},
		},
	)

	ew.wg.Add(1)
	go func() {
		defer ew.wg.Done()
		controller.Run(ew.stopCh)
	}()

	ew.logger.Info("Event watcher started")
	return nil
}

// Stop stops the event watcher
func (ew *EventWatcher) Stop() {
	ew.mu.Lock()
	defer ew.mu.Unlock()

	if !ew.running {
		return
	}

	ew.logger.Info("Stopping event watcher")
	close(ew.stopCh)
	ew.wg.Wait()
	ew.running = false
	ew.logger.Info("Event watcher stopped")
}

// handleEvent processes a Kubernetes event and converts it to our Event type
func (ew *EventWatcher) handleEvent(obj interface{}, eventType string) {
	event, ok := obj.(*corev1.Event)
	if !ok {
		ew.logger.Warn("Failed to cast object to Event")
		return
	}

	// Filter events based on importance
	if !ew.shouldProcessEvent(event) {
		return
	}

	// Convert Kubernetes event to our Event type
	agentEvent := ew.convertEvent(event, eventType)

	// Avoid duplicate events
	if ew.isDuplicateEvent(agentEvent) {
		return
	}

	select {
	case ew.eventChan <- agentEvent:
		ew.logger.Debug("Event sent",
			zap.String("event_id", agentEvent.ID),
			zap.String("reason", agentEvent.Reason),
			zap.String("namespace", agentEvent.Namespace))
		ew.lastEventID = agentEvent.ID
	default:
		ew.logger.Warn("Event channel full, dropping event",
			zap.String("event_id", agentEvent.ID))
	}
}

// shouldProcessEvent determines if an event should be processed based on its importance
func (ew *EventWatcher) shouldProcessEvent(event *corev1.Event) bool {
	// Focus on critical events that indicate problems
	criticalReasons := map[string]bool{
		"Failed":                   true,
		"FailedMount":              true,
		"FailedSync":               true,
		"FailedCreatePodSandBox":   true,
		"FailedKillPod":            true,
		"FailedPodSandBoxStatus":   true,
		"FailedScheduling":         true,
		"InspectFailed":            true,
		"Killing":                  true,
		"NodeNotReady":             true,
		"NodeNotSchedulable":       true,
		"Preempting":               true,
		"Pulling":                  true,
		"Rebooted":                 true,
		"Starting":                 true,
		"Unhealthy":                true,
		"BackOff":                  true,
		"CrashLoopBackOff":         true,
		"ImagePullBackOff":         true,
		"InvalidImageName":         true,
		"ErrImagePull":             true,
		"ImageInspectError":        true,
		"ErrImageNeverPull":        true,
		"RegistryUnavailable":      true,
		"OOMKilling":               true,
		"ProbeWarning":             true,
		"ExceededGracePeriod":      true,
		"FailedPostStartHook":      true,
		"FailedPreStopHook":        true,
		"UnexpectedAdmissionError": true,
		"DNSConfigForming":         true,
		"ContainerGCFailed":        true,
		"ImageGCFailed":            true,
		"FailedNodeAllocatableEnforcement": true,
		"FailedAttachVolume":               true,
		"FailedDetachVolume":               true,
		"VolumeResizeFailed":               true,
		"FileSystemResizeFailed":           true,
		"FailedMapVolume":                  true,
		"FailedUnmapDevice":                true,
	}

	// Also include Warning and Error type events
	if event.Type == corev1.EventTypeWarning || event.Type == corev1.EventTypeNormal {
		return criticalReasons[event.Reason]
	}

	return false
}

// convertEvent converts a Kubernetes event to our Event type
func (ew *EventWatcher) convertEvent(k8sEvent *corev1.Event, eventType string) *types.Event {
	severity := ew.determineSeverity(k8sEvent)

	return &types.Event{
		ID:         string(uuid.NewUUID()),
		ClusterID:  ew.clusterID,
		Type:       "k8s_event",
		Source:     "kubernetes",
		Namespace:  k8sEvent.Namespace,
		Severity:   severity,
		Reason:     k8sEvent.Reason,
		Message:    k8sEvent.Message,
		Timestamp:  k8sEvent.FirstTimestamp.Time,
		ReportedAt: time.Now(),
		Labels: map[string]string{
			"kind":             k8sEvent.InvolvedObject.Kind,
			"name":             k8sEvent.InvolvedObject.Name,
			"uid":              string(k8sEvent.InvolvedObject.UID),
			"event_type":       eventType,
			"k8s_event_type":   k8sEvent.Type,
			"reporting_component": k8sEvent.Source.Component,
			"reporting_instance":  k8sEvent.Source.Host,
		},
		RawData: map[string]interface{}{
			"count":          k8sEvent.Count,
			"first_timestamp": k8sEvent.FirstTimestamp,
			"last_timestamp":  k8sEvent.LastTimestamp,
			"involved_object": map[string]interface{}{
				"kind":       k8sEvent.InvolvedObject.Kind,
				"namespace":  k8sEvent.InvolvedObject.Namespace,
				"name":       k8sEvent.InvolvedObject.Name,
				"uid":        k8sEvent.InvolvedObject.UID,
				"api_version": k8sEvent.InvolvedObject.APIVersion,
			},
		},
	}
}

// determineSeverity determines the severity level based on the event
func (ew *EventWatcher) determineSeverity(event *corev1.Event) string {
	// Critical events that indicate immediate attention needed
	criticalPatterns := []string{
		"CrashLoopBackOff",
		"OOMKilling",
		"Failed",
		"NodeNotReady",
		"FailedScheduling",
	}

	// High priority events
	highPatterns := []string{
		"BackOff",
		"ImagePullBackOff",
		"ErrImagePull",
		"Unhealthy",
		"Killing",
		"Preempting",
	}

	// Medium priority events
	mediumPatterns := []string{
		"Pulling",
		"Starting",
		"Rebooted",
		"ProbeWarning",
	}

	reason := event.Reason
	eventType := event.Type

	// Check for critical patterns
	for _, pattern := range criticalPatterns {
		if strings.Contains(reason, pattern) {
			return "critical"
		}
	}

	// Warning events are generally high priority
	if eventType == corev1.EventTypeWarning {
		// Check for high patterns in warning events
		for _, pattern := range highPatterns {
			if strings.Contains(reason, pattern) {
				return "high"
			}
		}
		// Other warnings are medium priority
		return "medium"
	}

	// Check for medium patterns in normal events
	for _, pattern := range mediumPatterns {
		if strings.Contains(reason, pattern) {
			return "medium"
		}
	}

	// Default to low priority
	return "low"
}

// isDuplicateEvent checks if this event has already been processed recently
func (ew *EventWatcher) isDuplicateEvent(event *types.Event) bool {
	// Simple duplicate check - in production, you might want a more sophisticated approach
	// using a time-based cache or hash of event content
	return event.ID == ew.lastEventID
}