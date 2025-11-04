package service

import (
	"context"
	"fmt"
	"io"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	"github.com/kart-io/logger"
)

// K8sPodService Pod 管理服务
type K8sPodService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sPodService 创建新的 Pod 服务
func NewK8sPodService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sPodService {
	return &K8sPodService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// PodInfo Pod 信息
type PodInfo struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Status     string            `json:"status"`
	Phase      string            `json:"phase"`
	NodeName   string            `json:"nodeName"`
	PodIP      string            `json:"podIP"`
	Labels     map[string]string `json:"labels"`
	Containers []ContainerInfo   `json:"containers"`
	CreatedAt  string            `json:"createdAt"`
}

// ContainerInfo 容器信息
type ContainerInfo struct {
	Name         string `json:"name"`
	Image        string `json:"image"`
	Ready        bool   `json:"ready"`
	RestartCount int32  `json:"restartCount"`
	State        string `json:"state"`
}

// ListPods 获取 Pod 列表
func (s *K8sPodService) ListPods(ctx context.Context, clusterID, namespace string, offset, limit int) ([]PodInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	pods, err := client.Clientset().CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list pods: %w", err))
	}

	total := int64(len(pods.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(pods.Items) {
		start = len(pods.Items)
	}
	if end > len(pods.Items) {
		end = len(pods.Items)
	}

	result := make([]PodInfo, 0)
	for i := start; i < end; i++ {
		pod := pods.Items[i]
		result = append(result, s.convertPodInfo(&pod))
	}

	return result, total, nil
}

// GetPod 获取 Pod 详情
func (s *K8sPodService) GetPod(ctx context.Context, clusterID, namespace, podName string) (*PodInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	pod, err := client.Clientset().CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get pod: %w", err))
	}

	podInfo := s.convertPodInfo(pod)
	return &podInfo, nil
}

// DeletePod 删除 Pod
func (s *K8sPodService) DeletePod(ctx context.Context, clusterID, namespace, podName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	err = client.Clientset().CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete pod: %w", err))
	}

	logger.Infow("Pod deleted",
		"cluster_id", clusterID,
		"namespace", namespace,
		"pod", podName,
	)

	return nil
}

// GetPodLogs 获取 Pod 日志
func (s *K8sPodService) GetPodLogs(ctx context.Context, clusterID, namespace, podName, container, tailLines string, follow bool) (string, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return "", err
	}

	// 解析 tailLines
	tailLinesInt, err := strconv.ParseInt(tailLines, 10, 64)
	if err != nil {
		tailLinesInt = 100
	}

	podLogOpts := &corev1.PodLogOptions{
		Container: container,
		TailLines: &tailLinesInt,
		Follow:    follow,
	}

	req := client.Clientset().CoreV1().Pods(namespace).GetLogs(podName, podLogOpts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", errors.NewK8sAPIError(fmt.Errorf("failed to get pod logs: %w", err))
	}
	defer stream.Close()

	buf, err := io.ReadAll(stream)
	if err != nil {
		return "", errors.NewK8sAPIError(fmt.Errorf("failed to read logs: %w", err))
	}

	return string(buf), nil
}

// convertPodInfo 转换 Pod 信息
func (s *K8sPodService) convertPodInfo(pod *corev1.Pod) PodInfo {
	containers := make([]ContainerInfo, 0, len(pod.Status.ContainerStatuses))
	for _, cs := range pod.Status.ContainerStatuses {
		state := "unknown"
		if cs.State.Running != nil {
			state = "running"
		} else if cs.State.Waiting != nil {
			state = "waiting"
		} else if cs.State.Terminated != nil {
			state = "terminated"
		}

		containers = append(containers, ContainerInfo{
			Name:         cs.Name,
			Image:        cs.Image,
			Ready:        cs.Ready,
			RestartCount: cs.RestartCount,
			State:        state,
		})
	}

	return PodInfo{
		Name:       pod.Name,
		Namespace:  pod.Namespace,
		Status:     string(pod.Status.Phase),
		Phase:      string(pod.Status.Phase),
		NodeName:   pod.Spec.NodeName,
		PodIP:      pod.Status.PodIP,
		Labels:     pod.Labels,
		Containers: containers,
		CreatedAt:  pod.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}
