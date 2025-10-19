package service

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cluster-service/internal/storage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K8sStorageClassService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

func NewK8sStorageClassService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sStorageClassService {
	return &K8sStorageClassService{
		storage:        storage,
		clusterService: clusterService,
	}
}

type StorageClassInfo struct {
	Name              string            `json:"name"`
	Provisioner       string            `json:"provisioner"`
	ReclaimPolicy     string            `json:"reclaimPolicy,omitempty"`
	VolumeBindingMode string            `json:"volumeBindingMode,omitempty"`
	AllowExpansion    bool              `json:"allowVolumeExpansion"`
	Parameters        map[string]string `json:"parameters,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	CreatedAt         string            `json:"createdAt"`
}

func (s *K8sStorageClassService) ListStorageClasses(ctx context.Context, clusterID string, offset, limit int) ([]StorageClassInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get client: %w", err)
	}

	storageClassList, err := client.Clientset().StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list storageclasses: %w", err)
	}

	total := int64(len(storageClassList.Items))

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(storageClassList.Items) {
		start = len(storageClassList.Items)
	}
	if end > len(storageClassList.Items) {
		end = len(storageClassList.Items)
	}

	storageClasses := make([]StorageClassInfo, 0, end-start)
	for i := start; i < end; i++ {
		sc := storageClassList.Items[i]

		reclaimPolicy := ""
		if sc.ReclaimPolicy != nil {
			reclaimPolicy = string(*sc.ReclaimPolicy)
		}

		volumeBindingMode := ""
		if sc.VolumeBindingMode != nil {
			volumeBindingMode = string(*sc.VolumeBindingMode)
		}

		allowExpansion := false
		if sc.AllowVolumeExpansion != nil {
			allowExpansion = *sc.AllowVolumeExpansion
		}

		scInfo := StorageClassInfo{
			Name:              sc.Name,
			Provisioner:       sc.Provisioner,
			ReclaimPolicy:     reclaimPolicy,
			VolumeBindingMode: volumeBindingMode,
			AllowExpansion:    allowExpansion,
			Parameters:        sc.Parameters,
			Labels:            sc.Labels,
			CreatedAt:         sc.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
		}
		storageClasses = append(storageClasses, scInfo)
	}

	return storageClasses, total, nil
}

func (s *K8sStorageClassService) GetStorageClass(ctx context.Context, clusterID, name string) (*StorageClassInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	sc, err := client.Clientset().StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get storageclass: %w", err)
	}

	reclaimPolicy := ""
	if sc.ReclaimPolicy != nil {
		reclaimPolicy = string(*sc.ReclaimPolicy)
	}

	volumeBindingMode := ""
	if sc.VolumeBindingMode != nil {
		volumeBindingMode = string(*sc.VolumeBindingMode)
	}

	allowExpansion := false
	if sc.AllowVolumeExpansion != nil {
		allowExpansion = *sc.AllowVolumeExpansion
	}

	scInfo := &StorageClassInfo{
		Name:              sc.Name,
		Provisioner:       sc.Provisioner,
		ReclaimPolicy:     reclaimPolicy,
		VolumeBindingMode: volumeBindingMode,
		AllowExpansion:    allowExpansion,
		Parameters:        sc.Parameters,
		Labels:            sc.Labels,
		CreatedAt:         sc.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}

	return scInfo, nil
}

func (s *K8sStorageClassService) DeleteStorageClass(ctx context.Context, clusterID, name string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	err = client.Clientset().StorageV1().StorageClasses().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete storageclass: %w", err)
	}

	return nil
}
