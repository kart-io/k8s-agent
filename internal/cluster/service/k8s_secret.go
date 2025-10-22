package service

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/common/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sSecretService Secret 管理服务
type K8sSecretService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sSecretService 创建新的 Secret 服务
func NewK8sSecretService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sSecretService {
	return &K8sSecretService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// SecretInfo Secret 信息
type SecretInfo struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Type        string            `json:"type"`
	Data        map[string][]byte `json:"data,omitempty"`        // 实际数据（Base64 编码）
	StringData  map[string]string `json:"stringData,omitempty"`  // 字符串数据（用于显示）
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   string            `json:"createdAt"`
}

// CreateSecretRequest 创建 Secret 请求
type CreateSecretRequest struct {
	Name        string            `json:"name" binding:"required"`
	Namespace   string            `json:"namespace" binding:"required"`
	Type        string            `json:"type"`                  // Opaque, kubernetes.io/tls, etc.
	Data        map[string][]byte `json:"data,omitempty"`        // Base64 编码的数据
	StringData  map[string]string `json:"stringData,omitempty"`  // 明文数据（会自动编码）
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// ListSecrets 获取 Secret 列表
func (s *K8sSecretService) ListSecrets(ctx context.Context, clusterID, namespace string, offset, limit int) ([]SecretInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	secrets, err := client.Clientset().CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list secrets: %w", err))
	}

	total := int64(len(secrets.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(secrets.Items) {
		start = len(secrets.Items)
	}
	if end > len(secrets.Items) {
		end = len(secrets.Items)
	}

	result := make([]SecretInfo, 0)
	for i := start; i < end; i++ {
		secret := secrets.Items[i]
		result = append(result, s.convertSecretInfo(&secret, false)) // 不返回敏感数据
	}

	return result, total, nil
}

// GetSecret 获取 Secret 详情
func (s *K8sSecretService) GetSecret(ctx context.Context, clusterID, namespace, secretName string, includeData bool) (*SecretInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	secret, err := client.Clientset().CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get secret: %w", err))
	}

	secretInfo := s.convertSecretInfo(secret, includeData)
	return &secretInfo, nil
}

// CreateSecret 创建 Secret
func (s *K8sSecretService) CreateSecret(ctx context.Context, clusterID string, req *CreateSecretRequest) (*SecretInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	// 构建 Secret 对象
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   req.Namespace,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
		Type:       corev1.SecretType(req.Type),
		Data:       req.Data,
		StringData: req.StringData,
	}

	// 如果没有指定类型，默认为 Opaque
	if secret.Type == "" {
		secret.Type = corev1.SecretTypeOpaque
	}

	created, err := client.Clientset().CoreV1().Secrets(req.Namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to create secret: %w", err))
	}

	logger.Infow("Secret created",
		"cluster_id", clusterID,
		"namespace", req.Namespace,
		"secret", req.Name,
		"type", secret.Type,
	)

	secretInfo := s.convertSecretInfo(created, false) // 创建后不返回敏感数据
	return &secretInfo, nil
}

// UpdateSecret 更新 Secret
func (s *K8sSecretService) UpdateSecret(ctx context.Context, clusterID, namespace, secretName string, req *CreateSecretRequest) (*SecretInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	// 获取现有 Secret
	existing, err := client.Clientset().CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get secret: %w", err))
	}

	// 更新字段
	existing.Labels = req.Labels
	existing.Annotations = req.Annotations
	existing.Data = req.Data
	existing.StringData = req.StringData

	if req.Type != "" {
		existing.Type = corev1.SecretType(req.Type)
	}

	updated, err := client.Clientset().CoreV1().Secrets(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to update secret: %w", err))
	}

	logger.Infow("Secret updated",
		"cluster_id", clusterID,
		"namespace", namespace,
		"secret", secretName,
	)

	secretInfo := s.convertSecretInfo(updated, false) // 更新后不返回敏感数据
	return &secretInfo, nil
}

// DeleteSecret 删除 Secret
func (s *K8sSecretService) DeleteSecret(ctx context.Context, clusterID, namespace, secretName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().CoreV1().Secrets(namespace).Delete(ctx, secretName, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete secret: %w", err))
	}

	logger.Infow("Secret deleted",
		"cluster_id", clusterID,
		"namespace", namespace,
		"secret", secretName,
	)

	return nil
}

// convertSecretInfo 转换 Secret 信息
func (s *K8sSecretService) convertSecretInfo(secret *corev1.Secret, includeData bool) SecretInfo {
	info := SecretInfo{
		Name:        secret.Name,
		Namespace:   secret.Namespace,
		Type:        string(secret.Type),
		Labels:      secret.Labels,
		Annotations: secret.Annotations,
		CreatedAt:   secret.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}

	// 根据 includeData 参数决定是否包含敏感数据
	if includeData {
		info.Data = secret.Data
		// 将 Data 转换为 StringData 用于显示（仅用于纯文本数据）
		info.StringData = make(map[string]string)
		for k, v := range secret.Data {
			info.StringData[k] = string(v)
		}
	}

	return info
}
