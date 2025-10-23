package service

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sEventService Event 管理服务
type K8sEventService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sEventService 创建新的 Event 服务
func NewK8sEventService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sEventService {
	return &K8sEventService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// ListEvents 获取 Event 列表
func (s *K8sEventService) ListEvents(ctx context.Context, clusterID, namespace string, eventType string, eventReason string, offset, limit int) ([]corev1.Event, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	// 如果 namespace 为 "-"，表示查询所有命名空间，使用空字符串
	queryNamespace := namespace
	if namespace == "-" {
		queryNamespace = ""
	}

	events, err := client.Clientset().CoreV1().Events(queryNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list events: %w", err))
	}

	// 过滤事件类型和原因
	filteredEvents := events.Items
	if eventType != "" || eventReason != "" {
		filtered := make([]corev1.Event, 0)
		for _, event := range events.Items {
			// 类型过滤
			if eventType != "" && event.Type != eventType {
				continue
			}
			// 原因过滤
			if eventReason != "" && event.Reason != eventReason {
				continue
			}
			filtered = append(filtered, event)
		}
		filteredEvents = filtered
	}

	total := int64(len(filteredEvents))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(filteredEvents) {
		start = len(filteredEvents)
	}
	if end > len(filteredEvents) {
		end = len(filteredEvents)
	}

	pagedEvents := filteredEvents[start:end]

	return pagedEvents, total, nil
}

// GetEvent 获取单个 Event 详情
func (s *K8sEventService) GetEvent(ctx context.Context, clusterID, namespace, name string) (*corev1.Event, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	event, err := client.Clientset().CoreV1().Events(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get event: %w", err))
	}

	return event, nil
}
