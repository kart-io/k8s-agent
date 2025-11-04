package service

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	"github.com/kart-io/logger"
)

// K8sConfigMapService ConfigMap 管理服务.
type K8sConfigMapService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sConfigMapService 创建新的 ConfigMap 服务.
func NewK8sConfigMapService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sConfigMapService {
	return &K8sConfigMapService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// ConfigMapInfo ConfigMap 信息.
type ConfigMapInfo struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Data        map[string]string `json:"data"`
	BinaryData  map[string][]byte `json:"binaryData,omitempty"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   string            `json:"createdAt"`
}

// CreateConfigMapRequest 创建 ConfigMap 请求.
type CreateConfigMapRequest struct {
	Name        string            `json:"name" binding:"required"`
	Namespace   string            `json:"namespace" binding:"required"`
	Data        map[string]string `json:"data"`
	BinaryData  map[string][]byte `json:"binaryData,omitempty"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// ListConfigMaps 获取 ConfigMap 列表.
func (s *K8sConfigMapService) ListConfigMaps(ctx context.Context, clusterID, namespace string, offset, limit int) ([]ConfigMapInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	configmaps, err := client.Clientset().CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list configmaps: %w", err))
	}

	total := int64(len(configmaps.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(configmaps.Items) {
		start = len(configmaps.Items)
	}
	if end > len(configmaps.Items) {
		end = len(configmaps.Items)
	}

	result := make([]ConfigMapInfo, 0)
	for i := start; i < end; i++ {
		cm := configmaps.Items[i]
		result = append(result, s.convertConfigMapInfo(&cm))
	}

	return result, total, nil
}

// GetConfigMap 获取 ConfigMap 详情.
func (s *K8sConfigMapService) GetConfigMap(ctx context.Context, clusterID, namespace, configmapName string) (*ConfigMapInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	configmap, err := client.Clientset().CoreV1().ConfigMaps(namespace).Get(ctx, configmapName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get configmap: %w", err))
	}

	cmInfo := s.convertConfigMapInfo(configmap)
	return &cmInfo, nil
}

// CreateConfigMap 创建 ConfigMap.
func (s *K8sConfigMapService) CreateConfigMap(ctx context.Context, clusterID string, req *CreateConfigMapRequest) (*ConfigMapInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	// 构建 ConfigMap 对象
	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   req.Namespace,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
		Data:       req.Data,
		BinaryData: req.BinaryData,
	}

	created, err := client.Clientset().CoreV1().ConfigMaps(req.Namespace).Create(ctx, configmap, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to create configmap: %w", err))
	}

	logger.Infow("ConfigMap created",
		"cluster_id", clusterID,
		"namespace", req.Namespace,
		"configmap", req.Name,
	)

	cmInfo := s.convertConfigMapInfo(created)
	return &cmInfo, nil
}

// UpdateConfigMap 更新 ConfigMap.
func (s *K8sConfigMapService) UpdateConfigMap(ctx context.Context, clusterID, namespace, configmapName string, req *CreateConfigMapRequest) (*ConfigMapInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	// 获取现有 ConfigMap
	existing, err := client.Clientset().CoreV1().ConfigMaps(namespace).Get(ctx, configmapName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get configmap: %w", err))
	}

	// 更新字段
	existing.Labels = req.Labels
	existing.Annotations = req.Annotations
	existing.Data = req.Data
	existing.BinaryData = req.BinaryData

	updated, err := client.Clientset().CoreV1().ConfigMaps(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to update configmap: %w", err))
	}

	logger.Infow("ConfigMap updated",
		"cluster_id", clusterID,
		"namespace", namespace,
		"configmap", configmapName,
	)

	cmInfo := s.convertConfigMapInfo(updated)
	return &cmInfo, nil
}

// DeleteConfigMap 删除 ConfigMap.
func (s *K8sConfigMapService) DeleteConfigMap(ctx context.Context, clusterID, namespace, configmapName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().CoreV1().ConfigMaps(namespace).Delete(ctx, configmapName, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete configmap: %w", err))
	}

	logger.Infow("ConfigMap deleted",
		"cluster_id", clusterID,
		"namespace", namespace,
		"configmap", configmapName,
	)

	return nil
}

// convertConfigMapInfo 转换 ConfigMap 信息.
func (s *K8sConfigMapService) convertConfigMapInfo(configmap *corev1.ConfigMap) ConfigMapInfo {
	return ConfigMapInfo{
		Name:        configmap.Name,
		Namespace:   configmap.Namespace,
		Data:        configmap.Data,
		BinaryData:  configmap.BinaryData,
		Labels:      configmap.Labels,
		Annotations: configmap.Annotations,
		CreatedAt:   configmap.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}
