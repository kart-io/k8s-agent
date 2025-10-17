# Kubernetes Agent - API 实现计划

## 文档版本

- **版本**: v1.0.0
- **创建日期**: 2025-10-17
- **基础路径**: `/api/k8s`

## 当前实现状态

### 已完成的资源 (✅)

1. **集群管理** (1.x) - 100%
   - GET /clusters (列表)
   - GET /clusters/:id (详情)
   - POST /clusters (创建)
   - PUT /clusters/:id (更新)
   - DELETE /clusters/:id (删除)
   - GET /clusters/:id/metrics (监控指标)

2. **命名空间管理** (2.x) - 100%
   - GET /clusters/:clusterId/namespaces (列表)
   - GET /clusters/:clusterId/namespaces/:name (详情)
   - POST /clusters/:clusterId/namespaces (创建)
   - DELETE /clusters/:clusterId/namespaces/:name (删除)

3. **节点管理** (3.x) - 100%
   - GET /clusters/:clusterId/nodes (列表)
   - GET /clusters/:clusterId/nodes/:name (详情)
   - POST /clusters/:clusterId/nodes/:name/cordon (封锁)
   - POST /clusters/:clusterId/nodes/:name/uncordon (解除封锁)
   - POST /clusters/:clusterId/nodes/:name/drain (驱逐)

4. **Pod 管理** (4.1.x) - 66%
   - ✅ GET /clusters/:clusterId/namespaces/:namespace/pods (列表)
   - ✅ GET /clusters/:clusterId/namespaces/:namespace/pods/:name (详情)
   - ❌ POST /clusters/:clusterId/namespaces/:namespace/pods (创建)
   - ❌ PUT /clusters/:clusterId/namespaces/:namespace/pods/:name (更新)
   - ✅ DELETE /clusters/:clusterId/namespaces/:namespace/pods/:name (删除)
   - ✅ GET /clusters/:clusterId/namespaces/:namespace/pods/:name/logs (日志)
   - ❌ POST /clusters/:clusterId/namespaces/:namespace/pods/:name/exec (执行命令)

5. **Deployment 管理** (4.2.x) - 100%
   - GET /clusters/:clusterId/namespaces/:namespace/deployments (列表)
   - GET /clusters/:clusterId/namespaces/:namespace/deployments/:name (详情)
   - POST /clusters/:clusterId/namespaces/:namespace/deployments/:name/scale (扩缩容)
   - POST /clusters/:clusterId/namespaces/:namespace/deployments/:name/restart (重启)

6. **StatefulSet 管理** (4.3.x) - 100%
   - GET /clusters/:clusterId/namespaces/:namespace/statefulsets (列表)
   - GET /clusters/:clusterId/namespaces/:namespace/statefulsets/:name (详情)
   - PUT /clusters/:clusterId/namespaces/:namespace/statefulsets/:name/scale (扩缩容)
   - POST /clusters/:clusterId/namespaces/:namespace/statefulsets/:name/restart (重启)
   - DELETE /clusters/:clusterId/namespaces/:namespace/statefulsets/:name (删除)

7. **DaemonSet 管理** (4.4.x) - 50% (Service 层已实现,Handler 层缺失)
   - ❌ GET /clusters/:clusterId/namespaces/:namespace/daemonsets (列表)
   - ❌ GET /clusters/:clusterId/namespaces/:namespace/daemonsets/:name (详情)
   - ❌ POST /clusters/:clusterId/namespaces/:namespace/daemonsets (创建)
   - ❌ DELETE /clusters/:clusterId/namespaces/:namespace/daemonsets/:name (删除)

8. **Service 管理** (5.1.x) - 100%
   - GET /clusters/:clusterId/namespaces/:namespace/services (列表)
   - GET /clusters/:clusterId/namespaces/:namespace/services/:name (详情)
   - POST /clusters/:clusterId/namespaces/:namespace/services (创建)
   - PUT /clusters/:clusterId/namespaces/:namespace/services/:name (更新)
   - DELETE /clusters/:clusterId/namespaces/:namespace/services/:name (删除)

9. **ConfigMap 管理** (6.1.x) - 50% (Service 层已实现,Handler 层缺失)
   - ❌ GET /clusters/:clusterId/namespaces/:namespace/configmaps (列表)
   - ❌ GET /clusters/:clusterId/namespaces/:namespace/configmaps/:name (详情)
   - ❌ POST /clusters/:clusterId/namespaces/:namespace/configmaps (创建)
   - ❌ PUT /clusters/:clusterId/namespaces/:namespace/configmaps/:name (更新)
   - ❌ DELETE /clusters/:clusterId/namespaces/:namespace/configmaps/:name (删除)

10. **Secret 管理** (6.2.x) - 50% (Service 层已实现,Handler 层缺失)
    - ❌ GET /clusters/:clusterId/namespaces/:namespace/secrets (列表)
    - ❌ GET /clusters/:clusterId/namespaces/:namespace/secrets/:name (详情)
    - ❌ POST /clusters/:clusterId/namespaces/:namespace/secrets (创建)
    - ❌ PUT /clusters/:clusterId/namespaces/:namespace/secrets/:name (更新)
    - ❌ DELETE /clusters/:clusterId/namespaces/:namespace/secrets/:name (删除)

### 待实现的资源 (❌)

#### 优先级 P0 (核心功能,立即实现)

1. **ReplicaSet 管理** (4.5.x) - 0%
2. **Job 管理** (4.6.x) - 0%
3. **CronJob 管理** (4.7.x) - 0%
4. **Ingress 管理** (5.2.x) - 0%
5. **PersistentVolume 管理** (6.3.x) - 0%
6. **PersistentVolumeClaim 管理** (6.4.x) - 0%
7. **StorageClass 管理** (6.5.x) - 0%

#### 优先级 P1 (重要功能,第二批实现)

8. **NetworkPolicy 管理** (5.3.x) - 0%
9. **Endpoints 管理** (5.4.x) - 0%
10. **EndpointSlice 管理** (5.5.x) - 0%
11. **ServiceAccount 管理** (7.1.x) - 0%
12. **Role 管理** (7.2.x) - 0%
13. **RoleBinding 管理** (7.3.x) - 0%
14. **ClusterRole 管理** (7.4.x) - 0%
15. **ClusterRoleBinding 管理** (7.5.x) - 0%

#### 优先级 P2 (增强功能,第三批实现)

16. **ResourceQuota 管理** (8.1.x) - 0%
17. **LimitRange 管理** (8.2.x) - 0%
18. **HorizontalPodAutoscaler 管理** (9.1.x) - 0%
19. **PriorityClass 管理** (10.1.x) - 0%
20. **Events 管理** (11.x) - 0%

## 实现计划

### 第一阶段:补充已有 Service 的 Handler 层 (2-3 天)

**目标**: 将已有 Service 层的功能暴露为 REST API

#### 任务列表

1. **DaemonSet Handler** (0.5 天)
   - 实现 ListDaemonSets
   - 实现 GetDaemonSet
   - 实现 CreateDaemonSet
   - 实现 DeleteDaemonSet

2. **ConfigMap Handler** (0.5 天)
   - 实现 ListConfigMaps
   - 实现 GetConfigMap
   - 实现 CreateConfigMap
   - 实现 UpdateConfigMap
   - 实现 DeleteConfigMap

3. **Secret Handler** (0.5 天)
   - 实现 ListSecrets
   - 实现 GetSecret
   - 实现 CreateSecret
   - 实现 UpdateSecret
   - 实现 DeleteSecret

4. **Pod 补充功能** (1 天)
   - 实现 CreatePod
   - 实现 UpdatePod
   - 实现 ExecPod (WebSocket 支持)

#### 交付物

- DaemonSet、ConfigMap、Secret 的完整 REST API
- Pod 的创建、更新、命令执行功能
- 路由配置更新
- API 测试脚本

### 第二阶段:实现 P0 优先级资源 (5-7 天)

**目标**: 实现核心工作负载和存储管理功能

#### 工作负载管理 (3 天)

1. **ReplicaSet** (1 天)
   - Service: k8s_replicaset.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get, Create, Update, Delete, Scale

2. **Job** (1 天)
   - Service: k8s_job.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get, Create, Delete

3. **CronJob** (1 天)
   - Service: k8s_cronjob.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get, Create, Update, Delete, Suspend/Resume

#### 网络管理 (1 天)

4. **Ingress** (1 天)
   - Service: k8s_ingress.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get, Create, Update, Delete

#### 存储管理 (3 天)

5. **PersistentVolume** (1 天)
   - Service: k8s_persistentvolume.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get, Create, Update, Delete

6. **PersistentVolumeClaim** (1 天)
   - Service: k8s_persistentvolumeclaim.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get, Create, Update, Delete

7. **StorageClass** (1 天)
   - Service: k8s_storageclass.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get, Create, Update, Delete

#### 交付物

- 完整的工作负载管理 API (ReplicaSet, Job, CronJob)
- 完整的 Ingress 管理 API
- 完整的存储管理 API (PV, PVC, StorageClass)
- 集成测试
- API 文档更新

### 第三阶段:实现 P1 优先级资源 (7-9 天)

**目标**: 实现网络策略和 RBAC 权限管理

#### 网络管理 (2 天)

1. **NetworkPolicy** (1 天)
   - Service: k8s_networkpolicy.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get, Create, Update, Delete

2. **Endpoints** (0.5 天)
   - Service: k8s_endpoints.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get, Create, Update, Delete

3. **EndpointSlice** (0.5 天)
   - Service: k8s_endpointslice.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get, Create, Update, Delete

#### RBAC 权限管理 (5 天)

4. **ServiceAccount** (1 天)
   - Service: k8s_serviceaccount.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get, Create, Update, Delete

5. **Role** (1 天)
   - Service: k8s_role.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get, Create, Update, Delete

6. **RoleBinding** (1 天)
   - Service: k8s_rolebinding.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get, Create, Update, Delete

7. **ClusterRole** (1 天)
   - Service: k8s_clusterrole.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get, Create, Update, Delete

8. **ClusterRoleBinding** (1 天)
   - Service: k8s_clusterrolebinding.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get, Create, Update, Delete

#### 交付物

- 完整的网络策略 API
- 完整的 RBAC 权限管理 API
- 权限验证中间件
- 集成测试
- API 文档更新

### 第四阶段:实现 P2 优先级资源 (4-5 天)

**目标**: 实现资源配额、自动扩缩容和事件管理

#### 资源配额管理 (2 天)

1. **ResourceQuota** (1 天)
   - Service: k8s_resourcequota.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get, Create, Update, Delete

2. **LimitRange** (1 天)
   - Service: k8s_limitrange.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get, Create, Update, Delete

#### 自动扩缩容 (1 天)

3. **HorizontalPodAutoscaler** (1 天)
   - Service: k8s_hpa.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get, Create, Update, Delete

#### 调度与优先级 (1 天)

4. **PriorityClass** (1 天)
   - Service: k8s_priorityclass.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get, Create, Update, Delete

#### 事件管理 (1 天)

5. **Events** (1 天)
   - Service: k8s_events.go
   - Handler: 添加到 k8s_api.go
   - 功能: List, Get

#### 交付物

- 完整的资源配额管理 API
- 完整的 HPA 管理 API
- 完整的 PriorityClass 管理 API
- 完整的事件查询 API
- 集成测试
- API 文档更新

## 技术架构

### 文件组织结构

```
cluster-service/
├── internal/
│   ├── handler/
│   │   ├── cluster.go           # 原有集群管理 Handler
│   │   └── k8s_api.go           # K8s API Handler (包含所有资源)
│   └── service/
│       ├── cluster.go           # 原有集群服务
│       ├── k8s_cluster.go       # K8s 集群管理服务
│       ├── k8s_namespace.go     # 命名空间服务
│       ├── k8s_node.go          # 节点服务
│       ├── k8s_pod.go           # Pod 服务
│       ├── k8s_deployment.go    # Deployment 服务
│       ├── k8s_statefulset.go   # StatefulSet 服务
│       ├── k8s_daemonset.go     # DaemonSet 服务 (已存在)
│       ├── k8s_service.go       # Service 服务
│       ├── k8s_configmap.go     # ConfigMap 服务 (已存在)
│       ├── k8s_secret.go        # Secret 服务 (已存在)
│       ├── k8s_replicaset.go    # ReplicaSet 服务 (待创建)
│       ├── k8s_job.go           # Job 服务 (待创建)
│       ├── k8s_cronjob.go       # CronJob 服务 (待创建)
│       ├── k8s_ingress.go       # Ingress 服务 (待创建)
│       ├── k8s_networkpolicy.go # NetworkPolicy 服务 (待创建)
│       ├── k8s_endpoints.go     # Endpoints 服务 (待创建)
│       ├── k8s_endpointslice.go # EndpointSlice 服务 (待创建)
│       ├── k8s_pv.go            # PersistentVolume 服务 (待创建)
│       ├── k8s_pvc.go           # PersistentVolumeClaim 服务 (待创建)
│       ├── k8s_storageclass.go  # StorageClass 服务 (待创建)
│       ├── k8s_sa.go            # ServiceAccount 服务 (待创建)
│       ├── k8s_role.go          # Role 服务 (待创建)
│       ├── k8s_rolebinding.go   # RoleBinding 服务 (待创建)
│       ├── k8s_clusterrole.go   # ClusterRole 服务 (待创建)
│       ├── k8s_crb.go           # ClusterRoleBinding 服务 (待创建)
│       ├── k8s_resourcequota.go # ResourceQuota 服务 (待创建)
│       ├── k8s_limitrange.go    # LimitRange 服务 (待创建)
│       ├── k8s_hpa.go           # HPA 服务 (待创建)
│       ├── k8s_priorityclass.go # PriorityClass 服务 (待创建)
│       └── k8s_events.go        # Events 服务 (待创建)
└── cmd/
    └── server/
        └── main.go              # 服务启动入口
```

### 代码规范

#### Service 层规范

每个 Service 文件应包含:

```go
package service

import (
    "context"
    "k8s.io/client-go/kubernetes"
)

type K8sXxxService struct {
    clientManager *K8sClientManager
}

func NewK8sXxxService(clientManager *K8sClientManager) *K8sXxxService {
    return &K8sXxxService{
        clientManager: clientManager,
    }
}

// List 获取资源列表
func (s *K8sXxxService) ListXxx(ctx context.Context, clusterID, namespace string, offset, limit int) ([]interface{}, int64, error)

// Get 获取资源详情
func (s *K8sXxxService) GetXxx(ctx context.Context, clusterID, namespace, name string) (interface{}, error)

// Create 创建资源
func (s *K8sXxxService) CreateXxx(ctx context.Context, clusterID string, req *CreateXxxRequest) (interface{}, error)

// Update 更新资源
func (s *K8sXxxService) UpdateXxx(ctx context.Context, clusterID, namespace, name string, req *UpdateXxxRequest) (interface{}, error)

// Delete 删除资源
func (s *K8sXxxService) DeleteXxx(ctx context.Context, clusterID, namespace, name string) error
```

#### Handler 层规范

在 `k8s_api.go` 中添加:

```go
// ListXxx GET /api/k8s/clusters/:clusterId/namespaces/:namespace/xxx
func (h *K8sAPIHandler) ListXxx(c *gin.Context) {
    clusterID := c.Param("clusterId")
    namespace := c.Param("namespace")
    params := pagination.Parse(c)

    // 实现逻辑
}

// GetXxx GET /api/k8s/clusters/:clusterId/namespaces/:namespace/xxx/:name
func (h *K8sAPIHandler) GetXxx(c *gin.Context) {
    // 实现逻辑
}

// CreateXxx POST /api/k8s/clusters/:clusterId/namespaces/:namespace/xxx
func (h *K8sAPIHandler) CreateXxx(c *gin.Context) {
    // 实现逻辑
}

// UpdateXxx PUT /api/k8s/clusters/:clusterId/namespaces/:namespace/xxx/:name
func (h *K8sAPIHandler) UpdateXxx(c *gin.Context) {
    // 实现逻辑
}

// DeleteXxx DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/xxx/:name
func (h *K8sAPIHandler) DeleteXxx(c *gin.Context) {
    // 实现逻辑
}
```

### 测试策略

#### 单元测试

每个 Service 需要对应的测试文件:

```
k8s_xxx.go → k8s_xxx_test.go
```

测试覆盖率目标: > 80%

#### 集成测试

在 `test-api.sh` 中添加测试用例,测试实际的 API 调用

#### E2E 测试

使用真实的 Kubernetes 集群进行端到端测试

## 依赖与风险

### 技术依赖

- client-go: v0.28.x
- Gin: v1.9.x
- Go: 1.21+

### 风险评估

1. **API 兼容性**: 不同 Kubernetes 版本的 API 可能有差异
   - 缓解措施: 使用 client-go 的稳定版本,做好版本兼容性测试

2. **性能问题**: 大规模集群的资源列表可能很大
   - 缓解措施: 实现分页、字段选择器、标签选择器过滤

3. **权限问题**: 某些操作需要特定的 RBAC 权限
   - 缓解措施: 在文档中明确说明所需权限,提供错误提示

4. **并发安全**: 多个请求同时操作同一资源
   - 缓解措施: 使用乐观锁 (ResourceVersion),处理冲突重试

## 测试与验证

### 测试环境

- 开发环境: Minikube / Kind
- 测试环境: 真实 K8s 集群 (v1.28+)
- 生产环境: 待定

### 验收标准

- [ ] 所有 API 接口实现并通过测试
- [ ] API 文档完整且准确
- [ ] 单元测试覆盖率 > 80%
- [ ] 集成测试全部通过
- [ ] 性能测试达标 (列表 < 1s, 详情 < 500ms)
- [ ] 错误处理完善,错误信息清晰

## 里程碑

- **M1 (第 1 周)**: 完成第一阶段,补充 Handler 层
- **M2 (第 2-3 周)**: 完成第二阶段,实现 P0 资源
- **M3 (第 4-5 周)**: 完成第三阶段,实现 P1 资源
- **M4 (第 6 周)**: 完成第四阶段,实现 P2 资源
- **M5 (第 7 周)**: 完成测试、文档和优化

## 总结

当前项目已经实现了约 40% 的 API 功能,主要覆盖了核心的集群管理、节点管理、工作负载管理 (Deployment, StatefulSet) 和服务管理。

接下来需要按照上述计划,分阶段实现剩余的 60% 功能,预计需要 6-7 周时间完成全部 API 的实现和测试。

建议优先实现第一和第二阶段的功能,这样可以覆盖大部分常用的 Kubernetes 资源管理场景。
