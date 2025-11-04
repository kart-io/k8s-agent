package service

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	"github.com/kart-io/logger"
)

// K8sCronJobService CronJob 管理服务.
type K8sCronJobService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sCronJobService 创建新的 CronJob 服务.
func NewK8sCronJobService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sCronJobService {
	return &K8sCronJobService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// CronJobInfo CronJob 信息.
type CronJobInfo struct {
	Name                       string            `json:"name"`
	Namespace                  string            `json:"namespace"`
	Schedule                   string            `json:"schedule"`
	Suspend                    *bool             `json:"suspend"`
	Active                     int               `json:"active"`
	LastScheduleTime           string            `json:"lastScheduleTime,omitempty"`
	ConcurrencyPolicy          string            `json:"concurrencyPolicy"`
	SuccessfulJobsHistoryLimit *int32            `json:"successfulJobsHistoryLimit,omitempty"`
	FailedJobsHistoryLimit     *int32            `json:"failedJobsHistoryLimit,omitempty"`
	Selector                   map[string]string `json:"selector"`
	Labels                     map[string]string `json:"labels"`
	CreatedAt                  string            `json:"createdAt"`
}

// ListCronJobs 获取 CronJob 列表.
func (s *K8sCronJobService) ListCronJobs(ctx context.Context, clusterID, namespace string, offset, limit int) ([]CronJobInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	cronjobs, err := client.Clientset().BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list cronjobs: %w", err))
	}

	total := int64(len(cronjobs.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(cronjobs.Items) {
		start = len(cronjobs.Items)
	}
	if end > len(cronjobs.Items) {
		end = len(cronjobs.Items)
	}

	result := make([]CronJobInfo, 0)
	for i := start; i < end; i++ {
		cronjob := cronjobs.Items[i]
		result = append(result, s.convertCronJobInfo(&cronjob))
	}

	return result, total, nil
}

// GetCronJob 获取 CronJob 详情.
func (s *K8sCronJobService) GetCronJob(ctx context.Context, clusterID, namespace, cronjobName string) (*CronJobInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	cronjob, err := client.Clientset().BatchV1().CronJobs(namespace).Get(ctx, cronjobName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get cronjob: %w", err))
	}

	cronjobInfo := s.convertCronJobInfo(cronjob)
	return &cronjobInfo, nil
}

// CreateCronJob 创建 CronJob.
func (s *K8sCronJobService) CreateCronJob(ctx context.Context, clusterID, namespace string, cronjob *batchv1.CronJob) (*CronJobInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	createdCronJob, err := client.Clientset().BatchV1().CronJobs(namespace).Create(ctx, cronjob, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to create cronjob: %w", err))
	}

	logger.Infow("CronJob created",
		"cluster_id", clusterID,
		"namespace", namespace,
		"cronjob", createdCronJob.Name,
	)

	cronjobInfo := s.convertCronJobInfo(createdCronJob)
	return &cronjobInfo, nil
}

// UpdateCronJob 更新 CronJob.
func (s *K8sCronJobService) UpdateCronJob(ctx context.Context, clusterID, namespace string, cronjob *batchv1.CronJob) (*CronJobInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	updatedCronJob, err := client.Clientset().BatchV1().CronJobs(namespace).Update(ctx, cronjob, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to update cronjob: %w", err))
	}

	logger.Infow("CronJob updated",
		"cluster_id", clusterID,
		"namespace", namespace,
		"cronjob", updatedCronJob.Name,
	)

	cronjobInfo := s.convertCronJobInfo(updatedCronJob)
	return &cronjobInfo, nil
}

// DeleteCronJob 删除 CronJob.
func (s *K8sCronJobService) DeleteCronJob(ctx context.Context, clusterID, namespace, cronjobName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	// Delete with propagation policy to also delete jobs
	propagationPolicy := metav1.DeletePropagationBackground
	err = client.Clientset().BatchV1().CronJobs(namespace).Delete(ctx, cronjobName, metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete cronjob: %w", err))
	}

	logger.Infow("CronJob deleted",
		"cluster_id", clusterID,
		"namespace", namespace,
		"cronjob", cronjobName,
	)

	return nil
}

// convertCronJobInfo 转换 CronJob 信息.
func (s *K8sCronJobService) convertCronJobInfo(cronjob *batchv1.CronJob) CronJobInfo {
	cronjobInfo := CronJobInfo{
		Name:                       cronjob.Name,
		Namespace:                  cronjob.Namespace,
		Schedule:                   cronjob.Spec.Schedule,
		Suspend:                    cronjob.Spec.Suspend,
		Active:                     len(cronjob.Status.Active),
		ConcurrencyPolicy:          string(cronjob.Spec.ConcurrencyPolicy),
		SuccessfulJobsHistoryLimit: cronjob.Spec.SuccessfulJobsHistoryLimit,
		FailedJobsHistoryLimit:     cronjob.Spec.FailedJobsHistoryLimit,
		Labels:                     cronjob.Labels,
		CreatedAt:                  cronjob.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}

	if cronjob.Status.LastScheduleTime != nil {
		cronjobInfo.LastScheduleTime = cronjob.Status.LastScheduleTime.Format("2006-01-02T15:04:05Z")
	}

	return cronjobInfo
}
