package service

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/common/logger"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sJobService Job 管理服务
type K8sJobService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sJobService 创建新的 Job 服务
func NewK8sJobService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sJobService {
	return &K8sJobService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// JobInfo Job 信息
type JobInfo struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Completions       *int32            `json:"completions"`
	Parallelism       *int32            `json:"parallelism"`
	Active            int32             `json:"active"`
	Succeeded         int32             `json:"succeeded"`
	Failed            int32             `json:"failed"`
	CompletionTime    string            `json:"completionTime,omitempty"`
	StartTime         string            `json:"startTime,omitempty"`
	Selector          map[string]string `json:"selector"`
	Labels            map[string]string `json:"labels"`
	BackoffLimit      *int32            `json:"backoffLimit"`
	TTLSecondsAfterFinished *int32      `json:"ttlSecondsAfterFinished,omitempty"`
	CreatedAt         string            `json:"createdAt"`
}

// ListJobs 获取 Job 列表
func (s *K8sJobService) ListJobs(ctx context.Context, clusterID, namespace string, offset, limit int) ([]JobInfo, int64, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, 0, err
	}

	jobs, err := client.Clientset().BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, errors.NewK8sAPIError(fmt.Errorf("failed to list jobs: %w", err))
	}

	total := int64(len(jobs.Items))

	// 应用分页
	start := offset
	end := offset + limit
	if start > len(jobs.Items) {
		start = len(jobs.Items)
	}
	if end > len(jobs.Items) {
		end = len(jobs.Items)
	}

	result := make([]JobInfo, 0)
	for i := start; i < end; i++ {
		job := jobs.Items[i]
		result = append(result, s.convertJobInfo(&job))
	}

	return result, total, nil
}

// GetJob 获取 Job 详情
func (s *K8sJobService) GetJob(ctx context.Context, clusterID, namespace, jobName string) (*JobInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	job, err := client.Clientset().BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to get job: %w", err))
	}

	jobInfo := s.convertJobInfo(job)
	return &jobInfo, nil
}

// CreateJob 创建 Job
func (s *K8sJobService) CreateJob(ctx context.Context, clusterID, namespace string, job *batchv1.Job) (*JobInfo, error) {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	createdJob, err := client.Clientset().BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.NewK8sAPIError(fmt.Errorf("failed to create job: %w", err))
	}

	logger.Infow("Job created",
		"cluster_id", clusterID,
		"namespace", namespace,
		"job", createdJob.Name,
	)

	jobInfo := s.convertJobInfo(createdJob)
	return &jobInfo, nil
}

// DeleteJob 删除 Job
func (s *K8sJobService) DeleteJob(ctx context.Context, clusterID, namespace, jobName string) error {
	client, err := s.clusterService.getClient(ctx, clusterID)
	if err != nil {
		return err
	}

	// Delete with propagation policy to also delete pods
	propagationPolicy := metav1.DeletePropagationBackground
	err = client.Clientset().BatchV1().Jobs(namespace).Delete(ctx, jobName, metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	if err != nil {
		return errors.NewK8sAPIError(fmt.Errorf("failed to delete job: %w", err))
	}

	logger.Infow("Job deleted",
		"cluster_id", clusterID,
		"namespace", namespace,
		"job", jobName,
	)

	return nil
}

// convertJobInfo 转换 Job 信息
func (s *K8sJobService) convertJobInfo(job *batchv1.Job) JobInfo {
	jobInfo := JobInfo{
		Name:              job.Name,
		Namespace:         job.Namespace,
		Completions:       job.Spec.Completions,
		Parallelism:       job.Spec.Parallelism,
		Active:            job.Status.Active,
		Succeeded:         job.Status.Succeeded,
		Failed:            job.Status.Failed,
		Labels:            job.Labels,
		BackoffLimit:      job.Spec.BackoffLimit,
		TTLSecondsAfterFinished: job.Spec.TTLSecondsAfterFinished,
		CreatedAt:         job.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}

	if job.Spec.Selector != nil {
		jobInfo.Selector = job.Spec.Selector.MatchLabels
	}

	if job.Status.CompletionTime != nil {
		jobInfo.CompletionTime = job.Status.CompletionTime.Format("2006-01-02T15:04:05Z")
	}

	if job.Status.StartTime != nil {
		jobInfo.StartTime = job.Status.StartTime.Format("2006-01-02T15:04:05Z")
	}

	return jobInfo
}
