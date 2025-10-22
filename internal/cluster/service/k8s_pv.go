package service

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	"github.com/kart-io/k8s-agent/common/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sPVService PersistentVolume 管理服务
type K8sPVService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sPVService 创建新的 PV 服务
func NewK8sPVService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sPVService {
	return &K8sPVService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// PVInfo PersistentVolume 信息
type PVInfo struct {
	Name             string            `json:"name"`
	Status           string            `json:"status"`
	Claim            string            `json:"claim,omitempty"`
	Capacity         string            `json:"capacity,omitempty"`
	AccessModes      []string          `json:"accessModes,omitempty"`
	StorageClassName string            `json:"storageClassName,omitempty"`
	ReclaimPolicy    string            `json:"reclaimPolicy,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
	CreatedAt        string            `json:"createdAt"`
}

// ListPVs 获取 PV 列表
func (s *K8sPVService) ListPVs(ctx context.Context, clusterID string, offset, limit int) ([]PVInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	pvs, err := client.Clientset().CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list pvs: %w", err))
	}

	total := int64(len(pvs.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(pvs.Items) {
		start = len(pvs.Items)
	}
	if end > len(pvs.Items) {
		end = len(pvs.Items)
	}

	pagedPVs := pvs.Items[start:end]

	// 转换为 PVInfo
	result := make([]PVInfo, 0, len(pagedPVs))
	for _, pv := range pagedPVs {
		result = append(result, convertToPVInfo(&pv))
	}

	return result, total, nil
}

// GetPV 获取单个 PV 详情
func (s *K8sPVService) GetPV(ctx context.Context, clusterID, name string) (*corev1.PersistentVolume, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	pv, err := client.Clientset().CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get pv: %w", err))
	}

	return pv, nil
}

// DeletePV 删除 PV
func (s *K8sPVService) DeletePV(ctx context.Context, clusterID, name string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().CoreV1().PersistentVolumes().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete pv: %w", err))
	}

	return nil
}

// convertToPVInfo 转换 PV 为 PVInfo
func convertToPVInfo(pv *corev1.PersistentVolume) PVInfo {
	accessModes := make([]string, 0, len(pv.Spec.AccessModes))
	for _, mode := range pv.Spec.AccessModes {
		accessModes = append(accessModes, string(mode))
	}

	capacity := ""
	if pv.Spec.Capacity != nil {
		if storage, ok := pv.Spec.Capacity[corev1.ResourceStorage]; ok {
			capacity = storage.String()
		}
	}

	claim := ""
	if pv.Spec.ClaimRef != nil {
		claim = fmt.Sprintf("%s/%s", pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
	}

	return PVInfo{
		Name:             pv.Name,
		Status:           string(pv.Status.Phase),
		Claim:            claim,
		Capacity:         capacity,
		AccessModes:      accessModes,
		StorageClassName: pv.Spec.StorageClassName,
		ReclaimPolicy:    string(pv.Spec.PersistentVolumeReclaimPolicy),
		Labels:           pv.Labels,
		CreatedAt:        pv.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}
