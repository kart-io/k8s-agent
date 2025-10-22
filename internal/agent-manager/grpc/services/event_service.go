package services

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kart-io/logger/core"

	eventv1 "github.com/kart-io/k8s-agent/api/proto/gen/agentmanager/event/v1"
	"github.com/kart-io/k8s-agent/pkg/types"
)

// EventServiceServer 事件服务实现
type EventServiceServer struct {
	eventv1.UnimplementedEventServiceServer
	logger    core.Logger
	processor EventProcessor
	store     EventStore
}

// EventProcessor 定义事件处理器接口
type EventProcessor interface {
	ProcessEvent(ctx context.Context, event *types.Event) error
}

// EventStore 定义事件存储接口
type EventStore interface {
	GetEvent(ctx context.Context, eventID string) (*types.Event, error)
	ListEvents(ctx context.Context, filters interface{}) ([]*types.Event, error)
	SearchEvents(ctx context.Context, query interface{}) ([]*types.Event, error)
}

// NewEventServiceServer 创建事件服务
func NewEventServiceServer(logger core.Logger, processor EventProcessor, store EventStore) *EventServiceServer {
	return &EventServiceServer{
		logger:    logger.With("service", "event"),
		processor: processor,
		store:     store,
	}
}

// ListEvents 列出事件
func (s *EventServiceServer) ListEvents(ctx context.Context, req *eventv1.ListEventsRequest) (*eventv1.ListEventsResponse, error) {
	s.logger.Infow("ListEvents called",
		"cluster_id", req.ClusterId,
		"namespace", req.Namespace,
	)

	// TODO: 从存储层获取事件列表
	return &eventv1.ListEventsResponse{
		Events:        []*eventv1.Event{},
		NextPageToken: "",
		TotalCount:    0,
	}, nil
}

// GetEvent 获取事件详情
func (s *EventServiceServer) GetEvent(ctx context.Context, req *eventv1.GetEventRequest) (*eventv1.Event, error) {
	s.logger.Infow("GetEvent called", "event_id", req.EventId)

	if req.EventId == "" {
		return nil, status.Error(codes.InvalidArgument, "event_id is required")
	}

	// TODO: 从存储层获取事件信息
	return nil, status.Error(codes.NotFound, "event not found")
}

// SearchEvents 搜索事件
func (s *EventServiceServer) SearchEvents(ctx context.Context, req *eventv1.SearchEventsRequest) (*eventv1.SearchEventsResponse, error) {
	s.logger.Infow("SearchEvents called",
		"cluster_id", req.ClusterId,
		"namespace", req.Namespace,
		"keywords", req.Keywords,
	)

	// TODO: 实现事件搜索逻辑
	return &eventv1.SearchEventsResponse{
		Events:        []*eventv1.Event{},
		NextPageToken: "",
		TotalCount:    0,
	}, nil
}

// SubscribeEvents 订阅事件流
func (s *EventServiceServer) SubscribeEvents(req *eventv1.SubscribeEventsRequest, stream eventv1.EventService_SubscribeEventsServer) error {
	s.logger.Infow("SubscribeEvents called",
		"cluster_id", req.ClusterId,
		"namespace", req.Namespace,
	)

	// TODO: 实现事件订阅逻辑
	// 从 NATS 订阅事件并推送到流
	return status.Error(codes.Unimplemented, "not implemented yet")
}

// PublishEvent 发布事件
func (s *EventServiceServer) PublishEvent(ctx context.Context, req *eventv1.PublishEventRequest) (*eventv1.Event, error) {
	s.logger.Infow("PublishEvent called",
		"cluster_id", req.Event.ClusterId,
		"namespace", req.Event.Namespace,
	)

	// 验证输入
	if req.Event == nil {
		return nil, status.Error(codes.InvalidArgument, "event is required")
	}
	if req.Event.ClusterId == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster_id is required")
	}

	// 转换为内部类型
	event := convertProtoToTypesEvent(req.Event)

	// 处理事件
	if err := s.processor.ProcessEvent(ctx, event); err != nil {
		s.logger.Errorw("Failed to process event", "error", err)
		return nil, status.Error(codes.Internal, "failed to process event")
	}

	s.logger.Infow("Event published successfully", "event_id", event.ID)
	return convertTypesToProtoEvent(event), nil
}

// Type conversion functions

// convertTypesToProtoEvent converts types.Event to eventv1.Event
func convertTypesToProtoEvent(event *types.Event) *eventv1.Event {
	if event == nil {
		return nil
	}

	protoEvent := &eventv1.Event{
		EventId:   event.ID,
		ClusterId: event.ClusterID,
		Namespace: event.Namespace,
		EventType: event.Type,
		Reason:    event.Reason,
		Message:   event.Message,
		Severity:  convertSeverityStringToProto(event.Severity),
	}

	// Note: Proto Event doesn't have FirstTimestamp/LastTimestamp or Count
	// These are tracked differently in the proto definition

	return protoEvent
}

// convertProtoToTypesEvent converts eventv1.Event to types.Event
func convertProtoToTypesEvent(protoEvent *eventv1.Event) *types.Event {
	if protoEvent == nil {
		return nil
	}

	now := time.Now()
	event := &types.Event{
		ID:          protoEvent.EventId,
		ClusterID:   protoEvent.ClusterId,
		Namespace:   protoEvent.Namespace,
		Type:        protoEvent.EventType,
		Reason:      protoEvent.Reason,
		Message:     protoEvent.Message,
		Severity:    convertProtoToSeverityString(protoEvent.Severity),
		ProcessedAt: now,
	}

	return event
}

// convertSeverityStringToProto converts string severity to eventv1.EventSeverity
func convertSeverityStringToProto(severity string) eventv1.EventSeverity {
	switch severity {
	case "info":
		return eventv1.EventSeverity_EVENT_SEVERITY_INFO
	case "warning":
		return eventv1.EventSeverity_EVENT_SEVERITY_WARNING
	case "error":
		return eventv1.EventSeverity_EVENT_SEVERITY_ERROR
	case "critical":
		return eventv1.EventSeverity_EVENT_SEVERITY_CRITICAL
	default:
		return eventv1.EventSeverity_EVENT_SEVERITY_UNSPECIFIED
	}
}

// convertProtoToSeverityString converts eventv1.EventSeverity to string
func convertProtoToSeverityString(severity eventv1.EventSeverity) string {
	switch severity {
	case eventv1.EventSeverity_EVENT_SEVERITY_INFO:
		return "info"
	case eventv1.EventSeverity_EVENT_SEVERITY_WARNING:
		return "warning"
	case eventv1.EventSeverity_EVENT_SEVERITY_ERROR:
		return "error"
	case eventv1.EventSeverity_EVENT_SEVERITY_CRITICAL:
		return "critical"
	default:
		return "info"
	}
}
