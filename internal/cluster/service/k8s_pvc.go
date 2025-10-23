package service

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sPVCService PersistentVolumeClaim 管理服务
type K8sPVCService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sPVCService 创建新的 PVC 服务
func NewK8sPVCService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sPVCService {
	return &K8sPVCService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// PVCInfo PersistentVolumeClaim 信息
type PVCInfo struct {
	Name             string            `json:"name"`
	Namespace        string            `json:"namespace"`
	Status           string            `json:"status"`
	Volume           string            `json:"volume,omitempty"`
	Capacity         string            `json:"capacity,omitempty"`
	AccessModes      []string          `json:"accessModes,omitempty"`
	StorageClassName string            `json:"storageClassName,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
	CreatedAt        string            `json:"createdAt"`
}

// ListPVCs 获取 PVC 列表
func (s *K8sPVCService) ListPVCs(ctx context.Context, clusterID, namespace string, offset, limit int) ([]PVCInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	pvcs, err := client.Clientset().CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list pvcs: %w", err))
	}

	total := int64(len(pvcs.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(pvcs.Items) {
		start = len(pvcs.Items)
	}
	if end > len(pvcs.Items) {
		end = len(pvcs.Items)
	}

	pagedPVCs := pvcs.Items[start:end]

	// 转换为 PVCInfo
	result := make([]PVCInfo, 0, len(pagedPVCs))
	for _, pvc := range pagedPVCs {
		result = append(result, convertToPVCInfo(&pvc))
	}

	return result, total, nil
}

// GetPVC 获取单个 PVC 详情
func (s *K8sPVCService) GetPVC(ctx context.Context, clusterID, namespace, name string) (*corev1.PersistentVolumeClaim, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	pvc, err := client.Clientset().CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get pvc: %w", err))
	}

	return pvc, nil
}

// DeletePVC 删除 PVC
func (s *K8sPVCService) DeletePVC(ctx context.Context, clusterID, namespace, name string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete pvc: %w", err))
	}

	return nil
}

// convertToPVCInfo 转换 PVC 为 PVCInfo
func convertToPVCInfo(pvc *corev1.PersistentVolumeClaim) PVCInfo {
	accessModes := make([]string, 0, len(pvc.Spec.AccessModes))
	for _, mode := range pvc.Spec.AccessModes {
		accessModes = append(accessModes, string(mode))
	}

	capacity := ""
	if pvc.Status.Capacity != nil {
		if storage, ok := pvc.Status.Capacity[corev1.ResourceStorage]; ok {
			capacity = storage.String()
		}
	}

	storageClassName := ""
	if pvc.Spec.StorageClassName != nil {
		storageClassName = *pvc.Spec.StorageClassName
	}

	return PVCInfo{
		Name:             pvc.Name,
		Namespace:        pvc.Namespace,
		Status:           string(pvc.Status.Phase),
		Volume:           pvc.Spec.VolumeName,
		Capacity:         capacity,
		AccessModes:      accessModes,
		StorageClassName: storageClassName,
		Labels:           pvc.Labels,
		CreatedAt:        pvc.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}
