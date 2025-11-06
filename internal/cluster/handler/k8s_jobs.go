package handler

import (
	"github.com/gin-gonic/gin"
	batchv1 "k8s.io/api/batch/v1"

	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/logger"
)

func (h *K8sAPIHandler) ListJobs(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing jobs",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	jobs, total, err := h.jobService.ListJobs(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list jobs",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list jobs", err)
		return
	}

	resp := pagination.NewResponse(jobs, total, params)
	response.Success(c, resp)
}

func (h *K8sAPIHandler) GetJob(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	jobName := c.Query("name")

	logger.Infow("Getting job details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"job", jobName,
	)

	job, err := h.jobService.GetJob(c.Request.Context(), clusterID, namespace, jobName)
	if err != nil {
		logger.Errorw("Failed to get job",
			"cluster_id", clusterID,
			"namespace", namespace,
			"job", jobName,
			"error", err.Error(),
		)
		response.NotFound(c, "Job not found", err)
		return
	}

	response.Success(c, job)
}

func (h *K8sAPIHandler) CreateJob(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

	var job batchv1.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		response.BadRequest(c, "Invalid job data", err)
		return
	}

	logger.Infow("Creating job",
		"cluster_id", clusterID,
		"namespace", namespace,
		"job", job.Name,
	)

	createdJob, err := h.jobService.CreateJob(c.Request.Context(), clusterID, namespace, &job)
	if err != nil {
		logger.Errorw("Failed to create job",
			"cluster_id", clusterID,
			"namespace", namespace,
			"job", job.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to create job", err)
		return
	}

	response.SuccessWithMessage(c, "Job created successfully", createdJob)
}

func (h *K8sAPIHandler) DeleteJob(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	jobName := c.Query("name")

	logger.Infow("Deleting job",
		"cluster_id", clusterID,
		"namespace", namespace,
		"job", jobName,
	)

	if err := h.jobService.DeleteJob(c.Request.Context(), clusterID, namespace, jobName); err != nil {
		logger.Errorw("Failed to delete job",
			"cluster_id", clusterID,
			"namespace", namespace,
			"job", jobName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete job", err)
		return
	}

	response.SuccessWithMessage(c, "Job deleted successfully", gin.H{
		"job": jobName,
	})
}

func (h *K8sAPIHandler) ListCronJobs(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	params := pagination.Parse(c)

	logger.Infow("Listing cronjobs",
		"cluster_id", clusterID,
		"namespace", namespace,
	)

	cronjobs, total, err := h.cronjobService.ListCronJobs(
		c.Request.Context(),
		clusterID,
		namespace,
		params.GetOffset(),
		params.GetLimit(),
	)
	if err != nil {
		logger.Errorw("Failed to list cronjobs",
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to list cronjobs", err)
		return
	}

	resp := pagination.NewResponse(cronjobs, total, params)
	response.Success(c, resp)
}

func (h *K8sAPIHandler) GetCronJob(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	cronjobName := c.Query("name")

	logger.Infow("Getting cronjob details",
		"cluster_id", clusterID,
		"namespace", namespace,
		"cronjob", cronjobName,
	)

	cronjob, err := h.cronjobService.GetCronJob(c.Request.Context(), clusterID, namespace, cronjobName)
	if err != nil {
		logger.Errorw("Failed to get cronjob",
			"cluster_id", clusterID,
			"namespace", namespace,
			"cronjob", cronjobName,
			"error", err.Error(),
		)
		response.NotFound(c, "CronJob not found", err)
		return
	}

	response.Success(c, cronjob)
}

func (h *K8sAPIHandler) CreateCronJob(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

	var cronjob batchv1.CronJob
	if err := c.ShouldBindJSON(&cronjob); err != nil {
		response.BadRequest(c, "Invalid cronjob data", err)
		return
	}

	logger.Infow("Creating cronjob",
		"cluster_id", clusterID,
		"namespace", namespace,
		"cronjob", cronjob.Name,
	)

	createdCronJob, err := h.cronjobService.CreateCronJob(c.Request.Context(), clusterID, namespace, &cronjob)
	if err != nil {
		logger.Errorw("Failed to create cronjob",
			"cluster_id", clusterID,
			"namespace", namespace,
			"cronjob", cronjob.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to create cronjob", err)
		return
	}

	response.SuccessWithMessage(c, "CronJob created successfully", createdCronJob)
}

func (h *K8sAPIHandler) UpdateCronJob(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")

	var cronjob batchv1.CronJob
	if err := c.ShouldBindJSON(&cronjob); err != nil {
		response.BadRequest(c, "Invalid cronjob data", err)
		return
	}

	logger.Infow("Updating cronjob",
		"cluster_id", clusterID,
		"namespace", namespace,
		"cronjob", cronjob.Name,
	)

	updatedCronJob, err := h.cronjobService.UpdateCronJob(c.Request.Context(), clusterID, namespace, &cronjob)
	if err != nil {
		logger.Errorw("Failed to update cronjob",
			"cluster_id", clusterID,
			"namespace", namespace,
			"cronjob", cronjob.Name,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to update cronjob", err)
		return
	}

	response.SuccessWithMessage(c, "CronJob updated successfully", updatedCronJob)
}

func (h *K8sAPIHandler) DeleteCronJob(c *gin.Context) {
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	cronjobName := c.Query("name")

	logger.Infow("Deleting cronjob",
		"cluster_id", clusterID,
		"namespace", namespace,
		"cronjob", cronjobName,
	)

	if err := h.cronjobService.DeleteCronJob(c.Request.Context(), clusterID, namespace, cronjobName); err != nil {
		logger.Errorw("Failed to delete cronjob",
			"cluster_id", clusterID,
			"namespace", namespace,
			"cronjob", cronjobName,
			"error", err.Error(),
		)
		response.InternalError(c, "Failed to delete cronjob", err)
		return
	}

	response.SuccessWithMessage(c, "CronJob deleted successfully", gin.H{
		"cronjob": cronjobName,
	})
}
