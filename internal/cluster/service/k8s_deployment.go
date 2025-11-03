package service

import (
	"context"
	"fmt"
	"time"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/logger"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sDeploymentService Deployment 管理服务
type K8sDeploymentService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sDeploymentService 创建新的 Deployment 服务
func NewK8sDeploymentService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sDeploymentService {
	return &K8sDeploymentService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// DeploymentInfo Deployment 信息
type DeploymentInfo struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Replicas          int32             `json:"replicas"`
	AvailableReplicas int32             `json:"availableReplicas"`
	ReadyReplicas     int32             `json:"readyReplicas"`
	UpdatedReplicas   int32             `json:"updatedReplicas"`
	Labels            map[string]string `json:"labels"`
	Selector          map[string]string `json:"selector"`
	Strategy          string            `json:"strategy"`
	CreatedAt         string            `json:"createdAt"`
}

// ListDeployments 获取 Deployment 列表
func (s *K8sDeploymentService) ListDeployments(ctx context.Context, clusterID, namespace string, offset, limit int) ([]DeploymentInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	deployments, err := client.Clientset().AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list deployments: %w", err))
	}

	total := int64(len(deployments.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(deployments.Items) {
		start = len(deployments.Items)
	}
	if end > len(deployments.Items) {
		end = len(deployments.Items)
	}

	result := make([]DeploymentInfo, 0)
	for i := start; i < end; i++ {
		deploy := deployments.Items[i]
		result = append(result, s.convertDeploymentInfo(&deploy))
	}

	return result, total, nil
}

// GetDeployment 获取 Deployment 详情
func (s *K8sDeploymentService) GetDeployment(ctx context.Context, clusterID, namespace, deploymentName string) (*DeploymentInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	deployment, err := client.Clientset().AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get deployment: %w", err))
	}

	deploymentInfo := s.convertDeploymentInfo(deployment)
	return &deploymentInfo, nil
}

// ScaleDeployment 扩缩容 Deployment
func (s *K8sDeploymentService) ScaleDeployment(ctx context.Context, clusterID, namespace, deploymentName string, replicas int32) (*DeploymentInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	// 获取当前的 Deployment
	deployment, err := client.Clientset().AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get deployment: %w", err))
	}

	// 更新副本数
	deployment.Spec.Replicas = &replicas

	// 更新 Deployment
	updated, err := client.Clientset().AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to scale deployment: %w", err))
	}

	logger.Infow("Deployment scaled",
		"cluster_id", clusterID,
		"namespace", namespace,
		"deployment", deploymentName,
		"replicas", replicas,
	)

	deploymentInfo := s.convertDeploymentInfo(updated)
	return &deploymentInfo, nil
}

// RestartDeployment 重启 Deployment
func (s *K8sDeploymentService) RestartDeployment(ctx context.Context, clusterID, namespace, deploymentName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	// 获取当前的 Deployment
	deployment, err := client.Clientset().AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to get deployment: %w", err))
	}

	// 添加重启注解（通过修改 Pod 模板的注解来触发重启）
	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	// 更新 Deployment
	_, err = client.Clientset().AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to restart deployment: %w", err))
	}

	logger.Infow("Deployment restarted",
		"cluster_id", clusterID,
		"namespace", namespace,
		"deployment", deploymentName,
	)

	return nil
}

// convertDeploymentInfo 转换 Deployment 信息
func (s *K8sDeploymentService) convertDeploymentInfo(deployment *appsv1.Deployment) DeploymentInfo {
	var replicas int32
	if deployment.Spec.Replicas != nil {
		replicas = *deployment.Spec.Replicas
	}

	strategy := "RollingUpdate"
	if deployment.Spec.Strategy.Type != "" {
		strategy = string(deployment.Spec.Strategy.Type)
	}

	return DeploymentInfo{
		Name:              deployment.Name,
		Namespace:         deployment.Namespace,
		Replicas:          replicas,
		AvailableReplicas: deployment.Status.AvailableReplicas,
		ReadyReplicas:     deployment.Status.ReadyReplicas,
		UpdatedReplicas:   deployment.Status.UpdatedReplicas,
		Labels:            deployment.Labels,
		Selector:          deployment.Spec.Selector.MatchLabels,
		Strategy:          strategy,
		CreatedAt:         deployment.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}
