// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package initializers

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/cmd/cluster/app/options"
	"github.com/kart-io/k8s-agent/internal/cluster/handler"
	"github.com/kart-io/k8s-agent/internal/cluster/service"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// HTTPServerInitializer wraps the common HTTP server initializer.
// 使用标准化的 common/server 框架替代手动服务器管理
type HTTPServerInitializer struct {
	*pkginitializers.HTTPServerInitializer
	opts           *options.ServerOptions
	logger         core.Logger
	dbInit         *DatabaseInitializer
	handler        *handler.ClusterHandler
	k8sAPIHandler  *handler.K8sAPIHandler
	versionHandler *handler.VersionHandler
}

// NewHTTPServerInitializer creates a new HTTP server initializer using common/server framework.
func NewHTTPServerInitializer(
	opts *options.ServerOptions,
	logger core.Logger,
	dbInit *DatabaseInitializer,
) *HTTPServerInitializer {
	h := &HTTPServerInitializer{
		opts:   opts,
		logger: logger,
		dbInit: dbInit,
	}

	// Create standard HTTP server config
	serverConfig := &pkginitializers.HTTPServerConfig{
		Name:       "cluster-http-server",
		Priority:   bootstrap.PriorityHTTP,
		Config:     opts.Server,
		RouteSetup: h.setupRoutes, // Method defined below
	}

	// Create standard HTTP server initializer
	h.HTTPServerInitializer = pkginitializers.NewHTTPServerInitializer(serverConfig, logger)
	return h
}

// Initialize overrides to initialize services and handlers before server setup.
func (h *HTTPServerInitializer) Initialize(ctx context.Context) error {
	h.logger.Infow("Initializing cluster services and handlers")

	// Get database storage instance
	storage := h.dbInit.GetStorage()
	if storage == nil {
		return fmt.Errorf("database storage not initialized")
	}

	// Initialize all services and handlers
	if err := h.initializeServices(storage); err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}

	// Call the embedded initializer to set up the server
	return h.HTTPServerInitializer.Initialize(ctx)
}

// initializeServices initializes all service layers and handlers.
func (h *HTTPServerInitializer) initializeServices(storage interface{}) error {
	// Use actual storage instance
	mysqlStorage := h.dbInit.GetStorage()

	// Initialize base services
	clusterService := service.NewClusterService(mysqlStorage, h.logger)
	h.handler = handler.NewClusterHandler(clusterService, h.logger)
	h.versionHandler = handler.NewVersionHandler()

	// Initialize all K8s services using the registry (replaces 30+ lines of repetitive code)
	k8sServiceRegistry := service.NewK8sServiceRegistry(mysqlStorage)

	// Create K8s API handler with service registry (single parameter instead of 30!)
	h.k8sAPIHandler = handler.NewK8sAPIHandler(k8sServiceRegistry)

	h.logger.Infow("K8s API endpoints initialized",
		"count", k8sServiceRegistry.Count(),
		"base_path", "/api/k8s",
	)

	return nil
}

// setupRoutes configures all routes for the cluster service.
// This is called by the common HTTP server initializer during Initialize.
func (h *HTTPServerInitializer) setupRoutes(engine *gin.Engine) error {
	h.logger.Infow("Setting up cluster service routes")

	// Health check
	engine.GET("/health", h.handler.HealthCheck)

	// Version endpoints
	version := engine.Group("/version")
	{
		version.GET("", h.versionHandler.GetVersion)              // Full version info
		version.GET("/simple", h.versionHandler.GetVersionSimple) // Simplified version
		version.GET("/text", h.versionHandler.GetVersionText)     // Text format
		version.GET("/json", h.versionHandler.GetVersionJSON)     // JSON format
	}

	// K8s API routes - flattened routes with query parameters
	k8sAPI := engine.Group("/api/k8s")
	{
		// ===========================
		// Cluster Management API
		// ===========================
		k8sAPI.GET("/clusters", h.k8sAPIHandler.ListClusters)                 // List clusters
		k8sAPI.GET("/clusters/options", h.k8sAPIHandler.GetClusterOptions)    // Get cluster selector list
		k8sAPI.POST("/clusters", h.k8sAPIHandler.CreateCluster)               // Create cluster
		k8sAPI.GET("/cluster", h.k8sAPIHandler.GetCluster)                    // Get cluster details ?clusterId=xxx
		k8sAPI.PUT("/cluster", h.k8sAPIHandler.UpdateCluster)                 // Update cluster ?clusterId=xxx
		k8sAPI.DELETE("/cluster", h.k8sAPIHandler.DeleteCluster)              // Delete cluster ?clusterId=xxx
		k8sAPI.GET("/cluster/health", h.k8sAPIHandler.GetClusterHealthStatus) // Get cluster health ?clusterId=xxx

		// ===========================
		// Namespace Management API
		// ===========================
		k8sAPI.GET("/namespaces", h.k8sAPIHandler.ListNamespaces)    // List namespaces ?clusterId=xxx
		k8sAPI.POST("/namespaces", h.k8sAPIHandler.CreateNamespace)  // Create namespace ?clusterId=xxx
		k8sAPI.GET("/namespace", h.k8sAPIHandler.GetNamespace)       // Get namespace details ?clusterId=xxx&namespace=xxx
		k8sAPI.DELETE("/namespace", h.k8sAPIHandler.DeleteNamespace) // Delete namespace ?clusterId=xxx&namespace=xxx

		// ===========================
		// Workload Resources
		// ===========================
		// Pods
		k8sAPI.GET("/pods", h.k8sAPIHandler.ListPods)       // List pods ?clusterId=xxx&namespace=xxx
		k8sAPI.GET("/pod", h.k8sAPIHandler.GetPod)          // Get pod details ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.DELETE("/pod", h.k8sAPIHandler.DeletePod)    // Delete pod ?clusterId=xxx&namespace=xxx&name=xxx
		k8sAPI.GET("/pod/logs", h.k8sAPIHandler.GetPodLogs) // Get pod logs ?clusterId=xxx&namespace=xxx&name=xxx&container=xxx

		// Deployments
		k8sAPI.GET("/deployments", h.k8sAPIHandler.ListDeployments)          // List deployments
		k8sAPI.GET("/deployment", h.k8sAPIHandler.GetDeployment)             // Get deployment details
		k8sAPI.PUT("/deployment/scale", h.k8sAPIHandler.ScaleDeployment)     // Scale deployment
		k8sAPI.PUT("/deployment/restart", h.k8sAPIHandler.RestartDeployment) // Restart deployment

		// StatefulSets
		k8sAPI.GET("/statefulsets", h.k8sAPIHandler.ListStatefulSets)          // List statefulsets
		k8sAPI.GET("/statefulset", h.k8sAPIHandler.GetStatefulSet)             // Get statefulset details
		k8sAPI.DELETE("/statefulset", h.k8sAPIHandler.DeleteStatefulSet)       // Delete statefulset
		k8sAPI.PUT("/statefulset/scale", h.k8sAPIHandler.ScaleStatefulSet)     // Scale statefulset
		k8sAPI.PUT("/statefulset/restart", h.k8sAPIHandler.RestartStatefulSet) // Restart statefulset

		// DaemonSets
		k8sAPI.GET("/daemonsets", h.k8sAPIHandler.ListDaemonSets)          // List daemonsets
		k8sAPI.GET("/daemonset", h.k8sAPIHandler.GetDaemonSet)             // Get daemonset details
		k8sAPI.DELETE("/daemonset", h.k8sAPIHandler.DeleteDaemonSet)       // Delete daemonset
		k8sAPI.PUT("/daemonset/restart", h.k8sAPIHandler.RestartDaemonSet) // Restart daemonset

		// ReplicaSets
		k8sAPI.GET("/replicasets", h.k8sAPIHandler.ListReplicaSets)      // List replicasets
		k8sAPI.GET("/replicaset", h.k8sAPIHandler.GetReplicaSet)         // Get replicaset details
		k8sAPI.DELETE("/replicaset", h.k8sAPIHandler.DeleteReplicaSet)   // Delete replicaset
		k8sAPI.PUT("/replicaset/scale", h.k8sAPIHandler.ScaleReplicaSet) // Scale replicaset

		// Jobs
		k8sAPI.GET("/jobs", h.k8sAPIHandler.ListJobs)    // List jobs
		k8sAPI.GET("/job", h.k8sAPIHandler.GetJob)       // Get job details
		k8sAPI.POST("/jobs", h.k8sAPIHandler.CreateJob)  // Create job
		k8sAPI.DELETE("/job", h.k8sAPIHandler.DeleteJob) // Delete job

		// CronJobs
		k8sAPI.GET("/cronjobs", h.k8sAPIHandler.ListCronJobs)    // List cronjobs
		k8sAPI.GET("/cronjob", h.k8sAPIHandler.GetCronJob)       // Get cronjob details
		k8sAPI.POST("/cronjobs", h.k8sAPIHandler.CreateCronJob)  // Create cronjob
		k8sAPI.PUT("/cronjob", h.k8sAPIHandler.UpdateCronJob)    // Update cronjob
		k8sAPI.DELETE("/cronjob", h.k8sAPIHandler.DeleteCronJob) // Delete cronjob

		// ===========================
		// Service & Discovery
		// ===========================
		// Services
		k8sAPI.GET("/services", h.k8sAPIHandler.ListServices)    // List services
		k8sAPI.GET("/service", h.k8sAPIHandler.GetService)       // Get service details
		k8sAPI.POST("/services", h.k8sAPIHandler.CreateService)  // Create service
		k8sAPI.PUT("/service", h.k8sAPIHandler.UpdateService)    // Update service
		k8sAPI.DELETE("/service", h.k8sAPIHandler.DeleteService) // Delete service

		// Endpoints
		k8sAPI.GET("/endpoints", h.k8sAPIHandler.ListEndpoints) // List endpoints
		k8sAPI.GET("/endpoint", h.k8sAPIHandler.GetEndpoint)    // Get endpoint details

		// EndpointSlices
		k8sAPI.GET("/endpointslices", h.k8sAPIHandler.ListEndpointSlices) // List endpoint slices
		k8sAPI.GET("/endpointslice", h.k8sAPIHandler.GetEndpointSlice)    // Get endpoint slice details

		// Ingresses
		k8sAPI.GET("/ingresses", h.k8sAPIHandler.ListIngresses)  // List ingresses
		k8sAPI.GET("/ingress", h.k8sAPIHandler.GetIngress)       // Get ingress details
		k8sAPI.POST("/ingresses", h.k8sAPIHandler.CreateIngress) // Create ingress
		k8sAPI.PUT("/ingress", h.k8sAPIHandler.UpdateIngress)    // Update ingress
		k8sAPI.DELETE("/ingress", h.k8sAPIHandler.DeleteIngress) // Delete ingress

		// NetworkPolicies
		k8sAPI.GET("/networkpolicies", h.k8sAPIHandler.ListNetworkPolicies)  // List network policies
		k8sAPI.GET("/networkpolicy", h.k8sAPIHandler.GetNetworkPolicy)       // Get network policy details
		k8sAPI.POST("/networkpolicies", h.k8sAPIHandler.CreateNetworkPolicy) // Create network policy
		k8sAPI.PUT("/networkpolicy", h.k8sAPIHandler.UpdateNetworkPolicy)    // Update network policy
		k8sAPI.DELETE("/networkpolicy", h.k8sAPIHandler.DeleteNetworkPolicy) // Delete network policy

		// ===========================
		// Configuration & Storage
		// ===========================
		// ConfigMaps
		k8sAPI.GET("/configmaps", h.k8sAPIHandler.ListConfigMaps)    // List configmaps
		k8sAPI.GET("/configmap", h.k8sAPIHandler.GetConfigMap)       // Get configmap details
		k8sAPI.POST("/configmaps", h.k8sAPIHandler.CreateConfigMap)  // Create configmap
		k8sAPI.PUT("/configmap", h.k8sAPIHandler.UpdateConfigMap)    // Update configmap
		k8sAPI.DELETE("/configmap", h.k8sAPIHandler.DeleteConfigMap) // Delete configmap

		// Secrets
		k8sAPI.GET("/secrets", h.k8sAPIHandler.ListSecrets)    // List secrets
		k8sAPI.GET("/secret", h.k8sAPIHandler.GetSecret)       // Get secret details
		k8sAPI.POST("/secrets", h.k8sAPIHandler.CreateSecret)  // Create secret
		k8sAPI.PUT("/secret", h.k8sAPIHandler.UpdateSecret)    // Update secret
		k8sAPI.DELETE("/secret", h.k8sAPIHandler.DeleteSecret) // Delete secret

		// PersistentVolumeClaims
		k8sAPI.GET("/pvcs", h.k8sAPIHandler.ListPVCs)    // List PVCs
		k8sAPI.GET("/pvc", h.k8sAPIHandler.GetPVC)       // Get PVC details
		k8sAPI.DELETE("/pvc", h.k8sAPIHandler.DeletePVC) // Delete PVC

		// PersistentVolumes
		k8sAPI.GET("/pvs", h.k8sAPIHandler.ListPVs) // List PVs
		k8sAPI.GET("/pv", h.k8sAPIHandler.GetPV)    // Get PV details

		// StorageClasses
		k8sAPI.GET("/storageclasses", h.k8sAPIHandler.ListStorageClasses) // List storage classes
		k8sAPI.GET("/storageclass", h.k8sAPIHandler.GetStorageClass)      // Get storage class details

		// ===========================
		// Cluster Resources
		// ===========================
		// Nodes
		k8sAPI.GET("/nodes", h.k8sAPIHandler.ListNodes)            // List nodes
		k8sAPI.GET("/node", h.k8sAPIHandler.GetNode)               // Get node details
		k8sAPI.PUT("/node/cordon", h.k8sAPIHandler.CordonNode)     // Cordon node
		k8sAPI.PUT("/node/uncordon", h.k8sAPIHandler.UncordonNode) // Uncordon node
		k8sAPI.PUT("/node/drain", h.k8sAPIHandler.DrainNode)       // Drain node

		// ===========================
		// Security & RBAC
		// ===========================
		// ServiceAccounts
		k8sAPI.GET("/serviceaccounts", h.k8sAPIHandler.ListServiceAccounts)    // List service accounts
		k8sAPI.GET("/serviceaccount", h.k8sAPIHandler.GetServiceAccount)       // Get service account details
		k8sAPI.DELETE("/serviceaccount", h.k8sAPIHandler.DeleteServiceAccount) // Delete service account

		// Roles
		k8sAPI.GET("/roles", h.k8sAPIHandler.ListRoles) // List roles
		k8sAPI.GET("/role", h.k8sAPIHandler.GetRole)    // Get role details

		// RoleBindings
		k8sAPI.GET("/rolebindings", h.k8sAPIHandler.ListRoleBindings) // List role bindings
		k8sAPI.GET("/rolebinding", h.k8sAPIHandler.GetRoleBinding)    // Get role binding details

		// ClusterRoles
		k8sAPI.GET("/clusterroles", h.k8sAPIHandler.ListClusterRoles) // List cluster roles
		k8sAPI.GET("/clusterrole", h.k8sAPIHandler.GetClusterRole)    // Get cluster role details

		// ClusterRoleBindings
		k8sAPI.GET("/clusterrolebindings", h.k8sAPIHandler.ListClusterRoleBindings) // List cluster role bindings
		k8sAPI.GET("/clusterrolebinding", h.k8sAPIHandler.GetClusterRoleBinding)    // Get cluster role binding details

		// ===========================
		// Monitoring & Scaling
		// ===========================
		// HorizontalPodAutoscalers
		k8sAPI.GET("/hpas", h.k8sAPIHandler.ListHPAs)    // List HPAs
		k8sAPI.GET("/hpa", h.k8sAPIHandler.GetHPA)       // Get HPA details
		k8sAPI.DELETE("/hpa", h.k8sAPIHandler.DeleteHPA) // Delete HPA

		// Events
		k8sAPI.GET("/events", h.k8sAPIHandler.ListEvents) // List events

		// ===========================
		// Resource Management
		// ===========================
		// LimitRanges
		k8sAPI.GET("/limitranges", h.k8sAPIHandler.ListLimitRanges) // List limit ranges
		k8sAPI.GET("/limitrange", h.k8sAPIHandler.GetLimitRange)    // Get limit range details

		// ResourceQuotas
		k8sAPI.GET("/resourcequotas", h.k8sAPIHandler.ListResourceQuotas) // List resource quotas
		k8sAPI.GET("/resourcequota", h.k8sAPIHandler.GetResourceQuota)    // Get resource quota details

		// PriorityClasses
		k8sAPI.GET("/priorityclasses", h.k8sAPIHandler.ListPriorityClasses) // List priority classes
		k8sAPI.GET("/priorityclass", h.k8sAPIHandler.GetPriorityClass)      // Get priority class details

		// ===========================
		// Utility APIs
		// ===========================
		// Note: WebSocket endpoints like exec and port-forward are not implemented yet
	}

	h.logger.Infow("Cluster Service Routes Registered",
		"health", "GET /health",
		"version", "GET /version/*",
		"k8s_api", "/api/k8s/*",
		"total_endpoints", "150+",
	)

	return nil
}
