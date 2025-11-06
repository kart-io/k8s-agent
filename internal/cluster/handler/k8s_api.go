package handler

import (
	"github.com/kart-io/k8s-agent/internal/cluster/service"
)

// K8sAPIHandler 处理所有 Kubernetes API 请求
// 基于 /api/k8s 路径的完整 K8s 管理接口.
type K8sAPIHandler struct {
	clusterService            *service.K8sClusterService
	namespaceService          *service.K8sNamespaceService
	podService                *service.K8sPodService
	deploymentService         *service.K8sDeploymentService
	nodeService               *service.K8sNodeService
	serviceService            *service.K8sServiceService
	statefulsetService        *service.K8sStatefulSetService
	daemonsetService          *service.K8sDaemonSetService
	configmapService          *service.K8sConfigMapService
	secretService             *service.K8sSecretService
	endpointService           *service.K8sEndpointService
	pvcService                *service.K8sPVCService
	pvService                 *service.K8sPVService
	endpointsliceService      *service.K8sEndpointSliceService
	hpaService                *service.K8sHPAService
	eventService              *service.K8sEventService
	rolebindingService        *service.K8sRoleBindingService
	clusterroleService        *service.K8sClusterRoleService
	priorityclassService      *service.K8sPriorityClassService
	roleService               *service.K8sRoleService
	storageclassService       *service.K8sStorageClassService
	jobService                *service.K8sJobService
	cronjobService            *service.K8sCronJobService
	ingressService            *service.K8sIngressService
	networkpolicyService      *service.K8sNetworkPolicyService
	replicasetService         *service.K8sReplicaSetService
	limitrangeService         *service.K8sLimitRangeService
	serviceaccountService     *service.K8sServiceAccountService
	clusterrolebindingService *service.K8sClusterRoleBindingService
	resourcequotaService      *service.K8sResourceQuotaService
}

// NewK8sAPIHandler 创建新的 K8s API 处理器.
// 使用服务注册表代替30个单独的参数，大幅简化接口.
func NewK8sAPIHandler(registry *service.K8sServiceRegistry) *K8sAPIHandler {
	return &K8sAPIHandler{
		clusterService:            registry.ClusterService,
		namespaceService:          registry.NamespaceService,
		podService:                registry.PodService,
		deploymentService:         registry.DeploymentService,
		nodeService:               registry.NodeService,
		serviceService:            registry.ServiceService,
		statefulsetService:        registry.StatefulSetService,
		daemonsetService:          registry.DaemonSetService,
		configmapService:          registry.ConfigMapService,
		secretService:             registry.SecretService,
		endpointService:           registry.EndpointService,
		pvcService:                registry.PVCService,
		pvService:                 registry.PVService,
		endpointsliceService:      registry.EndpointSliceService,
		hpaService:                registry.HPAService,
		eventService:              registry.EventService,
		rolebindingService:        registry.RoleBindingService,
		clusterroleService:        registry.ClusterRoleService,
		priorityclassService:      registry.PriorityClassService,
		roleService:               registry.RoleService,
		storageclassService:       registry.StorageClassService,
		jobService:                registry.JobService,
		cronjobService:            registry.CronJobService,
		ingressService:            registry.IngressService,
		networkpolicyService:      registry.NetworkPolicyService,
		replicasetService:         registry.ReplicaSetService,
		limitrangeService:         registry.LimitRangeService,
		serviceaccountService:     registry.ServiceAccountService,
		clusterrolebindingService: registry.ClusterRoleBindingService,
		resourcequotaService:      registry.ResourceQuotaService,
	}
}
