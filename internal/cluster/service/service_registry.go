// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package service

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	"github.com/kart-io/logger/core"
)

// K8sServiceRegistry manages all K8s-related services.
// This eliminates the need to manually initialize 30+ services.
type K8sServiceRegistry struct {
	// Core services - using new unified ClusterService
	ClusterService *ClusterService

	// Workload services
	NamespaceService   *K8sNamespaceService
	PodService         *K8sPodService
	DeploymentService  *K8sDeploymentService
	StatefulSetService *K8sStatefulSetService
	DaemonSetService   *K8sDaemonSetService
	ReplicaSetService  *K8sReplicaSetService
	JobService         *K8sJobService
	CronJobService     *K8sCronJobService

	// Network services
	ServiceService       *K8sServiceService
	EndpointService      *K8sEndpointService
	EndpointSliceService *K8sEndpointSliceService
	IngressService       *K8sIngressService
	NetworkPolicyService *K8sNetworkPolicyService

	// Config & Storage services
	ConfigMapService    *K8sConfigMapService
	SecretService       *K8sSecretService
	PVCService          *K8sPVCService
	PVService           *K8sPVService
	StorageClassService *K8sStorageClassService

	// Cluster services
	NodeService *K8sNodeService

	// Security & RBAC services
	ServiceAccountService     *K8sServiceAccountService
	RoleService               *K8sRoleService
	RoleBindingService        *K8sRoleBindingService
	ClusterRoleService        *K8sClusterRoleService
	ClusterRoleBindingService *K8sClusterRoleBindingService

	// Resource Management services
	HPAService           *K8sHPAService
	EventService         *K8sEventService
	LimitRangeService    *K8sLimitRangeService
	ResourceQuotaService *K8sResourceQuotaService
	PriorityClassService *K8sPriorityClassService
}

// NewK8sServiceRegistry creates and initializes all K8s services.
// This replaces 30+ lines of manual service initialization with a single function call.
// Deprecated: Use NewK8sServiceRegistryWithDB instead (direct GORM DB access).
func NewK8sServiceRegistry(storage *storage.MySQLStorage, logger core.Logger) *K8sServiceRegistry {
	// Create core cluster service using the new unified service
	// We need to extract the GORM DB from storage for the new service
	// This is a transitional compatibility layer
	clusterService := NewK8sClusterService(storage.GormDB(), logger)

	// Initialize all other services with cluster service dependency
	return &K8sServiceRegistry{
		ClusterService: clusterService,

		// Workload services
		NamespaceService:   NewK8sNamespaceService(storage, clusterService),
		PodService:         NewK8sPodService(storage, clusterService),
		DeploymentService:  NewK8sDeploymentService(storage, clusterService),
		StatefulSetService: NewK8sStatefulSetService(storage, clusterService),
		DaemonSetService:   NewK8sDaemonSetService(storage, clusterService),
		ReplicaSetService:  NewK8sReplicaSetService(storage, clusterService),
		JobService:         NewK8sJobService(storage, clusterService),
		CronJobService:     NewK8sCronJobService(storage, clusterService),

		// Network services
		ServiceService:       NewK8sServiceService(storage, clusterService),
		EndpointService:      NewK8sEndpointService(storage, clusterService),
		EndpointSliceService: NewK8sEndpointSliceService(storage, clusterService),
		IngressService:       NewK8sIngressService(storage, clusterService),
		NetworkPolicyService: NewK8sNetworkPolicyService(storage, clusterService),

		// Config & Storage services
		ConfigMapService:    NewK8sConfigMapService(storage, clusterService),
		SecretService:       NewK8sSecretService(storage, clusterService),
		PVCService:          NewK8sPVCService(storage, clusterService),
		PVService:           NewK8sPVService(storage, clusterService),
		StorageClassService: NewK8sStorageClassService(storage, clusterService),

		// Cluster services
		NodeService: NewK8sNodeService(storage, clusterService),

		// Security & RBAC services
		ServiceAccountService:     NewK8sServiceAccountService(storage, clusterService),
		RoleService:               NewK8sRoleService(storage, clusterService),
		RoleBindingService:        NewK8sRoleBindingService(storage, clusterService),
		ClusterRoleService:        NewK8sClusterRoleService(storage, clusterService),
		ClusterRoleBindingService: NewK8sClusterRoleBindingService(storage, clusterService),

		// Resource Management services
		HPAService:           NewK8sHPAService(storage, clusterService),
		EventService:         NewK8sEventService(storage, clusterService),
		LimitRangeService:    NewK8sLimitRangeService(storage, clusterService),
		ResourceQuotaService: NewK8sResourceQuotaService(storage, clusterService),
		PriorityClassService: NewK8sPriorityClassService(storage, clusterService),
	}
}

// NewK8sServiceRegistryWithDB creates and initializes all K8s services with direct GORM DB access.
// This is the recommended way to create the registry, eliminating the storage layer wrapper.
// The DB should come from pkg/initializers.DatabaseInitializer.DB()
func NewK8sServiceRegistryWithDB(db *gorm.DB, logger core.Logger) *K8sServiceRegistry {
	// For now, we need to create storage from GORM DB for other K8s services
	// that still expect storage. This is a transitional state.
	// TODO: Update all K8s service constructors to accept *gorm.DB directly
	storage, err := storage.NewMySQLStorageWithDB(db, logger)
	if err != nil {
		// In production, handle this error properly
		panic(fmt.Sprintf("failed to create storage: %v", err))
	}

	// Create core cluster service with new unified ClusterService (no storage wrapper)
	clusterService := NewClusterService(db, logger)

	// Initialize all other services with cluster service dependency
	// Other services still use storage temporarily
	return &K8sServiceRegistry{
		ClusterService: clusterService,

		// Workload services - these still use storage temporarily
		NamespaceService:   NewK8sNamespaceService(storage, clusterService),
		PodService:         NewK8sPodService(storage, clusterService),
		DeploymentService:  NewK8sDeploymentService(storage, clusterService),
		StatefulSetService: NewK8sStatefulSetService(storage, clusterService),
		DaemonSetService:   NewK8sDaemonSetService(storage, clusterService),
		ReplicaSetService:  NewK8sReplicaSetService(storage, clusterService),
		JobService:         NewK8sJobService(storage, clusterService),
		CronJobService:     NewK8sCronJobService(storage, clusterService),

		// Network services
		ServiceService:       NewK8sServiceService(storage, clusterService),
		EndpointService:      NewK8sEndpointService(storage, clusterService),
		EndpointSliceService: NewK8sEndpointSliceService(storage, clusterService),
		IngressService:       NewK8sIngressService(storage, clusterService),
		NetworkPolicyService: NewK8sNetworkPolicyService(storage, clusterService),

		// Config & Storage services
		ConfigMapService:    NewK8sConfigMapService(storage, clusterService),
		SecretService:       NewK8sSecretService(storage, clusterService),
		PVCService:          NewK8sPVCService(storage, clusterService),
		PVService:           NewK8sPVService(storage, clusterService),
		StorageClassService: NewK8sStorageClassService(storage, clusterService),

		// Cluster services
		NodeService: NewK8sNodeService(storage, clusterService),

		// Security & RBAC services
		ServiceAccountService:     NewK8sServiceAccountService(storage, clusterService),
		RoleService:               NewK8sRoleService(storage, clusterService),
		RoleBindingService:        NewK8sRoleBindingService(storage, clusterService),
		ClusterRoleService:        NewK8sClusterRoleService(storage, clusterService),
		ClusterRoleBindingService: NewK8sClusterRoleBindingService(storage, clusterService),

		// Resource Management services
		HPAService:           NewK8sHPAService(storage, clusterService),
		EventService:         NewK8sEventService(storage, clusterService),
		LimitRangeService:    NewK8sLimitRangeService(storage, clusterService),
		ResourceQuotaService: NewK8sResourceQuotaService(storage, clusterService),
		PriorityClassService: NewK8sPriorityClassService(storage, clusterService),
	}
}

// Count returns the total number of registered services.
func (r *K8sServiceRegistry) Count() int {
	return 30 // Total K8s services
}
