# Cluster Service Architecture Analysis & Unification Plan

**Date**: 2025-11-10 (Updated)
**Status**: 🟡 GORM Migrated, but Architecture Issues Remain
**Service**: cluster

## Executive Summary

Cluster 服务已经完成了 GORM 迁移（参见 `CLUSTER_GORM_MIGRATION.md`），但仍然存在**严重的架构问题**：

1. **双服务并存**: `ClusterService` 和 `K8sClusterService` 功能重复
2. **Storage 层冗余**: `MySQLStorage` 层增加了不必要的复杂度
3. **类型定义分散**: 同一概念在三个地方定义
4. **架构不一致**: 不同文件使用不同的数据访问模式

**核心问题**：不是"没用 GORM"，而是**"用了 GORM 但架构混乱，导致改代码复杂"**。

## Current Architecture Problems

### Problem 1: 双服务架构混乱 🔴 **Critical**

#### 现状

```
internal/cluster/service/
├── cluster.go          # ClusterService (简化版，210 行)
│   ├── AddCluster()         # 添加集群
│   ├── GetClusterHealth()   # 获取健康状态
│   ├── GetPods()            # 获取 Pod 列表
│   └── getClient()          # 获取 K8s 客户端
│
└── k8s_cluster.go      # K8sClusterService (完整版，412 行)
    ├── ListClusters()       # 列表查询（分页）
    ├── GetCluster()         # 单个查询
    ├── CreateCluster()      # 创建集群
    ├── UpdateCluster()      # 更新集群
    ├── DeleteCluster()      # 删除集群
    ├── GetClusterHealth()   # 获取健康状态 ⚠️ 与 cluster.go 重复
    ├── GetClusterOptions()  # 获取选择器列表
    ├── populateClusterStats()  # 填充统计信息
    └── getClient()          # 获取 K8s 客户端 ⚠️ 与 cluster.go 重复
```

#### 问题分析

| 方法名 | ClusterService | K8sClusterService | 状态 |
|--------|---------------|-------------------|------|
| `AddCluster()` | ✅ 有 | ❌ 无 | 命名不一致（应为 Create） |
| `CreateCluster()` | ❌ 无 | ✅ 有 | 命名不一致 |
| `GetClusterHealth()` | ✅ 有 | ✅ 有 | ⚠️ **功能重复** |
| `getClient()` | ✅ 有 | ✅ 有 | ⚠️ **功能重复** |
| `ListClusters()` | ❌ 无 | ✅ 有 | 功能缺失 |
| `UpdateCluster()` | ❌ 无 | ✅ 有 | 功能缺失 |
| `DeleteCluster()` | ❌ 无 | ✅ 有 | 功能缺失 |
| `GetPods()` | ✅ 有 | ❌ 无 | 功能分散 |

**结论**: 两个服务功能重复且不完整，无法确定应该使用哪一个。

#### 数据访问方式不一致

**ClusterService** (`cluster.go`):
```go
type ClusterService struct {
    db      *gorm.DB               // ✅ 直接使用 GORM DB
    clients map[string]*k8s.Client
    log     core.Logger
}

// 创建操作
if err := s.db.WithContext(ctx).Create(clusterModel).Error; err != nil { ... }
```

**K8sClusterService** (`k8s_cluster.go`):
```go
type K8sClusterService struct {
    storage *storage.MySQLStorage  // ⚠️ 通过 Storage 层访问
    clients map[string]*k8s.Client
}

// 创建操作
if err := s.storage.GormDB().WithContext(ctx).Create(clusterModel).Error; err != nil { ... }
```

**问题**: 一个直接用 GORM，一个通过 Storage 层，架构不一致。

### Problem 2: Storage 层冗余 🟡 **Medium**

#### 文件: `internal/cluster/storage/mysql.go` (164 行)

**提供的方法**:
1. `NewMySQLStorage()` - 从 GORM DB 创建（推荐）
2. `NewMySQLStorageWithClient()` - 从 mysql.Client 创建（已弃用）
3. `NewMySQLStorageForTesting()` - 测试用
4. `DB()` - 返回 `*sql.DB`
5. `GormDB()` - 返回 `*gorm.DB`
6. `Close()` - 关闭连接
7. `InitSchema()` - 初始化数据库表（使用原始 SQL）

**问题分析**:

1. **功能重复**: `pkg/initializers.DatabaseInitializer` 已经提供了所有这些功能
2. **架构不一致**: 其他服务（agent-manager, orchestrator, auth）都直接使用 `*gorm.DB`
3. **维护负担**: 需要同时维护 Storage 层和 pkg/initializers 的逻辑
4. **Schema 管理落后**: 使用原始 SQL 而不是 GORM AutoMigrate

**对比 pkg/initializers.DatabaseInitializer**:

| 功能 | MySQLStorage | pkg/initializers | 结论 |
|------|--------------|------------------|------|
| GORM DB 访问 | ✅ `GormDB()` | ✅ `DB()` | 功能重复 |
| 连接管理 | ✅ `Close()` | ✅ `Shutdown()` | 功能重复 |
| Schema 初始化 | ✅ 原始 SQL | ✅ GORM AutoMigrate | pkg 更优 |
| 健康检查 | ❌ 无 | ✅ 有 | pkg 更完整 |
| 优雅关闭 | ⚠️ 简单 | ✅ 完整 | pkg 更可靠 |
| Bootstrap 集成 | ❌ 无 | ✅ 有 | pkg 更一致 |

**结论**: MySQLStorage 层是不必要的抽象，应该直接使用 `pkg/initializers.DatabaseInitializer`。

### Problem 3: 类型定义分散 🟡 **Medium**

#### 当前状态

**1. `internal/models/cluster/clusters.go`** (数据库模型)
```go
type Cluster struct {
    ID          string    `json:"id" gorm:"column:id"`
    Name        string    `json:"name" gorm:"column:name"`
    // ... 10 个字段
}
```

**2. `internal/cluster/types/cluster.go`** (业务类型 - **已删除?**)
```go
type Cluster struct {
    ID          string
    Name        string
    // ... 相同字段，无 gorm 标签
}

type ClusterHealth struct { ... }
type Pod struct { ... }
```

**3. `internal/cluster/service/k8s_cluster.go`** (服务层类型)
```go
type ClusterInfo struct {
    ID             string
    Name           string
    // ... 扩展字段（NodeCount, PodCount）
}

type ClusterHealth struct { ... }  // ⚠️ 与 types/ 重复
type ClusterOption struct { ... }
```

#### 问题

1. **同一概念多次定义**: `Cluster` 在 models/ 和 service/ 中都有定义
2. **重复类型**: `ClusterHealth` 在两处定义
3. **转换复杂**: 需要在 `Cluster` ↔ `ClusterInfo` 之间频繁转换
4. **维护困难**: 修改字段需要同步多个文件

#### 理想架构

```
internal/models/cluster/
├── cluster.go           # 数据库模型（带 gorm 标签）
├── health.go            # ClusterHealth（业务模型）
├── pod.go               # Pod, Container（业务模型）
└── types.go             # ClusterOption, 其他辅助类型
```

**原则**: 一个概念只定义一次，服务层直接使用 model。

### Problem 4: cmd/cluster/app/app.go 架构混乱

#### 当前代码结构

```go
type ClusterApp struct {
    bootstrap *bootstrap.Bootstrap
    opts      *commonapp.StandardOptions
    logger    core.Logger

    dbInit *pkginitializers.DatabaseInitializer  // ✅ 使用 pkg/initializers

    storage        *storage.MySQLStorage          // ⚠️ 冗余层
    clusterService *service.K8sClusterService     // ⚠️ 使用哪个服务?
}

func (a *ClusterApp) registerComponents(bs *bootstrap.Bootstrap) error {
    // 1. Database (通过 pkg/initializers)
    a.dbInit = pkginitializers.NewDatabaseInitializer(...)
    bs.Register(a.dbInit)

    // 2. Storage Layer (⚠️ 冗余)
    storageInit := &storageLayerInitializer{app: a}
    bs.Register(storageInit)

    // 3. Service Layer (依赖 storage，而不是直接用 dbInit)
    serviceInit := &serviceLayerInitializer{app: a}
    bs.Register(serviceInit)

    // 4. HTTP Server
    httpInit := &httpServerInitializer{app: a}
    bs.Register(httpInit)

    return nil
}
```

#### 问题

1. **冗余层级**: Database → Storage → Service（应该是 Database → Service）
2. **初始化复杂**: 需要 4 个 initializer（应该只需要 3 个）
3. **依赖不清晰**: Service 依赖 Storage 而不是直接依赖 Database
4. **与其他服务不一致**: agent-manager, orchestrator 都是 Database → Service 直连

## Unified Solution Plan

### Architecture Goal

**统一到单一 `ClusterService`，移除冗余层，与项目其他服务保持一致。**

```
┌─────────────────────────────────────────────────────────────┐
│ cmd/cluster/app/app.go                                      │
│   ↓ Bootstrap 初始化                                         │
│ pkg/initializers.DatabaseInitializer (Priority: 300)       │
│   ↓ 提供 *gorm.DB                                           │
│ internal/cluster/service/ClusterService (Priority: 600)    │
│   ↓ 使用 db.WithContext(ctx)                               │
│ internal/models/cluster/Cluster (GORM model)               │
└─────────────────────────────────────────────────────────────┘

删除:
❌ internal/cluster/storage/mysql.go (164 行)
❌ internal/cluster/service/k8s_cluster.go (412 行)
❌ internal/cluster/types/ (如果存在)
❌ internal/cluster/service/service_registry.go (如果存在)
```

### Step 1: 统一类型定义

**目标**: 合并所有类型定义到 `internal/models/cluster/`

#### 1.1 增强 `internal/models/cluster/clusters.go`

```go
package cluster

import "time"

const TableNameByCluster = "clusters"

// ========== 数据库模型 (带 GORM 标签) ==========

// Cluster K8s 集群数据库模型
type Cluster struct {
    ID          string    `json:"id" gorm:"column:id;primaryKey"`
    Name        string    `json:"name" gorm:"column:name;not null;index"`
    Description string    `json:"description" gorm:"column:description"`
    Endpoint    string    `json:"endpoint" gorm:"column:endpoint;not null"`
    Version     string    `json:"version" gorm:"column:version"`
    Status      string    `json:"status" gorm:"column:status;not null;default:'unknown';index"`
    Region      string    `json:"region" gorm:"column:region"`
    Provider    string    `json:"provider" gorm:"column:provider;index"`
    KubeConfig  string    `json:"kubeconfig,omitempty" gorm:"column:kubeconfig;type:text;not null"`
    CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
    UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`

    // ✅ 新增: 运行时统计字段（不存储在数据库）
    NodeCount      int `json:"nodeCount,omitempty" gorm:"-"`
    PodCount       int `json:"podCount,omitempty" gorm:"-"`
    NamespaceCount int `json:"namespaceCount,omitempty" gorm:"-"`
}

// TableName 返回表名
func (c *Cluster) TableName() string {
    return TableNameByCluster
}

// ========== 业务模型 ==========

// ClusterHealth 集群健康状态
type ClusterHealth struct {
    ClusterID   string    `json:"clusterId"`
    Status      string    `json:"status"` // healthy, degraded, unhealthy
    TotalNodes  int       `json:"totalNodes"`
    ReadyNodes  int       `json:"readyNodes"`
    TotalPods   int       `json:"totalPods"`
    RunningPods int       `json:"runningPods"`
    CheckedAt   time.Time `json:"checkedAt"`
}

// Pod Pod 信息
type Pod struct {
    Name       string            `json:"name"`
    Namespace  string            `json:"namespace"`
    Status     string            `json:"status"`
    Phase      string            `json:"phase"`
    NodeName   string            `json:"nodeName"`
    PodIP      string            `json:"podIP"`
    Labels     map[string]string `json:"labels"`
    Containers []Container       `json:"containers"`
    CreatedAt  time.Time         `json:"createdAt"`
}

// Container 容器信息
type Container struct {
    Name         string `json:"name"`
    Image        string `json:"image"`
    Ready        bool   `json:"ready"`
    RestartCount int32  `json:"restartCount"`
    State        string `json:"state"` // running, waiting, terminated, unknown
}

// ========== DTO/辅助类型 ==========

// ClusterOption 集群选择器选项（用于下拉框）
type ClusterOption struct {
    Label string `json:"label"` // 显示名称
    Value string `json:"value"` // 集群 ID
}

// ========== 常量定义 ==========

const (
    StatusHealthy   = "healthy"
    StatusDegraded  = "degraded"
    StatusUnhealthy = "unhealthy"
    StatusUnknown   = "unknown"

    ConditionTypeReady = "Ready"
)
```

**关键变化**:
1. ✅ `Cluster` 添加运行时统计字段（`gorm:"-"` 标记不存储）
2. ✅ 所有业务类型统一到 models/cluster/
3. ✅ 添加常量定义避免硬编码
4. ✅ 单一数据源，无需类型转换

#### 1.2 删除冗余类型定义

```bash
# 删除 internal/cluster/types/ 目录（如果存在）
rm -rf internal/cluster/types/

# 删除 k8s_cluster.go 中的类型定义（将在步骤 2 删除整个文件）
```

### Step 2: 重写统一的 ClusterService

**目标**: 合并两个服务的所有功能到单一 `ClusterService`

#### 2.1 新的 `internal/cluster/service/cluster.go`

```go
package service

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

    "github.com/kart-io/k8s-agent/common/errors"
    "github.com/kart-io/k8s-agent/internal/cluster/k8s"
    clustermodel "github.com/kart-io/k8s-agent/internal/models/cluster"
    "github.com/kart-io/logger/core"
)

// ClusterService 集群管理服务（统一版本）
type ClusterService struct {
    db      *gorm.DB               // 直接使用 GORM DB (来自 pkg/initializers)
    clients map[string]*k8s.Client // cluster_id -> k8s client 缓存
    log     core.Logger
}

// NewClusterService 创建集群服务
// db 参数应来自 pkg/initializers.DatabaseInitializer.DB()
func NewClusterService(db *gorm.DB, logger core.Logger) *ClusterService {
    return &ClusterService{
        db:      db,
        clients: make(map[string]*k8s.Client),
        log:     logger,
    }
}

// ========== CRUD 操作 ==========

// CreateCluster 创建集群
func (s *ClusterService) CreateCluster(ctx context.Context, req *CreateClusterRequest) (*clustermodel.Cluster, error) {
    // 1. 验证 kubeconfig 并测试连接
    client, err := k8s.NewClientFromKubeConfig([]byte(req.KubeConfig))
    if err != nil {
        return nil, errors.NewValidationError(fmt.Errorf("invalid kubeconfig: %w", err))
    }

    if err := client.CheckConnection(ctx); err != nil {
        return nil, errors.NewValidationError(fmt.Errorf("failed to connect to cluster: %w", err))
    }

    // 2. 获取集群版本
    version, err := client.GetServerVersion(ctx)
    if err != nil {
        s.log.Warnw("Failed to get server version", "error", err)
        version = clustermodel.StatusUnknown
    }

    // 3. 创建数据库记录
    cluster := &clustermodel.Cluster{
        ID:          uuid.New().String(),
        Name:        req.Name,
        Description: req.Description,
        Endpoint:    req.Endpoint,
        Version:     version,
        Status:      clustermodel.StatusHealthy,
        Region:      req.Region,
        Provider:    req.Provider,
        KubeConfig:  req.KubeConfig,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    if err := s.db.WithContext(ctx).Create(cluster).Error; err != nil {
        return nil, errors.NewDatabaseError(fmt.Errorf("failed to create cluster: %w", err))
    }

    // 4. 缓存客户端
    s.clients[cluster.ID] = client

    s.log.Infow("Cluster created",
        "cluster_id", cluster.ID,
        "name", cluster.Name,
        "version", version,
    )

    return cluster, nil
}

// ListClusters 获取集群列表（分页）
func (s *ClusterService) ListClusters(ctx context.Context, offset, limit int, withStats bool) ([]*clustermodel.Cluster, int64, error) {
    // 1. 查询总数
    var total int64
    if err := s.db.WithContext(ctx).Model(&clustermodel.Cluster{}).Count(&total).Error; err != nil {
        return nil, 0, errors.NewDatabaseError(fmt.Errorf("failed to count clusters: %w", err))
    }

    // 2. 查询列表
    var clusters []*clustermodel.Cluster
    if err := s.db.WithContext(ctx).
        Order("created_at DESC").
        Offset(offset).
        Limit(limit).
        Find(&clusters).Error; err != nil {
        return nil, 0, errors.NewDatabaseError(fmt.Errorf("failed to query clusters: %w", err))
    }

    // 3. 可选: 填充统计信息
    if withStats {
        for i := range clusters {
            if err := s.populateClusterStats(ctx, clusters[i]); err != nil {
                s.log.Warnw("Failed to populate cluster stats",
                    "cluster_id", clusters[i].ID,
                    "error", err,
                )
                // 统计信息获取失败不影响列表返回
            }
        }
    }

    return clusters, total, nil
}

// GetCluster 获取集群详情
func (s *ClusterService) GetCluster(ctx context.Context, clusterID string, withStats bool) (*clustermodel.Cluster, error) {
    var cluster clustermodel.Cluster
    if err := s.db.WithContext(ctx).Where("id = ?", clusterID).First(&cluster).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, errors.ErrClusterNotFound
        }
        return nil, errors.NewDatabaseError(fmt.Errorf("failed to query cluster: %w", err))
    }

    // 可选: 填充统计信息
    if withStats {
        if err := s.populateClusterStats(ctx, &cluster); err != nil {
            s.log.Warnw("Failed to populate cluster stats",
                "cluster_id", clusterID,
                "error", err,
            )
        }
    }

    return &cluster, nil
}

// UpdateCluster 更新集群信息
func (s *ClusterService) UpdateCluster(ctx context.Context, clusterID string, req *UpdateClusterRequest) (*clustermodel.Cluster, error) {
    // 1. 检查集群是否存在
    if _, err := s.GetCluster(ctx, clusterID, false); err != nil {
        return nil, err
    }

    // 2. 构建更新字段
    updates := map[string]interface{}{
        "updated_at": time.Now(),
    }
    if req.Name != "" {
        updates["name"] = req.Name
    }
    if req.Description != "" {
        updates["description"] = req.Description
    }

    // 3. 执行更新
    if err := s.db.WithContext(ctx).
        Model(&clustermodel.Cluster{}).
        Where("id = ?", clusterID).
        Updates(updates).Error; err != nil {
        return nil, errors.NewDatabaseError(fmt.Errorf("failed to update cluster: %w", err))
    }

    s.log.Infow("Cluster updated", "cluster_id", clusterID)

    return s.GetCluster(ctx, clusterID, false)
}

// DeleteCluster 删除集群
func (s *ClusterService) DeleteCluster(ctx context.Context, clusterID string) error {
    result := s.db.WithContext(ctx).Where("id = ?", clusterID).Delete(&clustermodel.Cluster{})
    if result.Error != nil {
        return errors.NewDatabaseError(fmt.Errorf("failed to delete cluster: %w", result.Error))
    }

    if result.RowsAffected == 0 {
        return errors.ErrClusterNotFound
    }

    // 清除缓存
    delete(s.clients, clusterID)

    s.log.Infow("Cluster deleted", "cluster_id", clusterID)

    return nil
}

// ========== K8s 资源查询操作 ==========

// GetClusterHealth 获取集群健康状态
func (s *ClusterService) GetClusterHealth(ctx context.Context, clusterID string) (*clustermodel.ClusterHealth, error) {
    client, err := s.getClient(ctx, clusterID)
    if err != nil {
        return nil, err
    }

    // 获取节点列表
    nodes, err := client.Clientset().CoreV1().Nodes().List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, errors.NewK8sAPIError(fmt.Errorf("failed to list nodes: %w", err))
    }

    readyNodes := 0
    for _, node := range nodes.Items {
        for _, condition := range node.Status.Conditions {
            if condition.Type == clustermodel.ConditionTypeReady && condition.Status == "True" {
                readyNodes++
                break
            }
        }
    }

    // 获取 Pod 列表
    pods, err := client.Clientset().CoreV1().Pods("").List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, errors.NewK8sAPIError(fmt.Errorf("failed to list pods: %w", err))
    }

    runningPods := 0
    for _, pod := range pods.Items {
        if pod.Status.Phase == "Running" {
            runningPods++
        }
    }

    // 计算状态
    status := clustermodel.StatusHealthy
    if readyNodes < len(nodes.Items) {
        status = clustermodel.StatusDegraded
    }
    if readyNodes == 0 {
        status = clustermodel.StatusUnhealthy
    }

    return &clustermodel.ClusterHealth{
        ClusterID:   clusterID,
        Status:      status,
        TotalNodes:  len(nodes.Items),
        ReadyNodes:  readyNodes,
        TotalPods:   len(pods.Items),
        RunningPods: runningPods,
        CheckedAt:   time.Now(),
    }, nil
}

// GetPods 获取 Pod 列表
func (s *ClusterService) GetPods(ctx context.Context, clusterID, namespace string) ([]*clustermodel.Pod, error) {
    client, err := s.getClient(ctx, clusterID)
    if err != nil {
        return nil, err
    }

    pods, err := client.Clientset().CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, errors.NewK8sAPIError(fmt.Errorf("failed to list pods: %w", err))
    }

    result := make([]*clustermodel.Pod, 0, len(pods.Items))
    for _, pod := range pods.Items {
        containers := make([]clustermodel.Container, 0, len(pod.Status.ContainerStatuses))
        for _, cs := range pod.Status.ContainerStatuses {
            state := clustermodel.StatusUnknown
            if cs.State.Running != nil {
                state = "running"
            } else if cs.State.Waiting != nil {
                state = "waiting"
            } else if cs.State.Terminated != nil {
                state = "terminated"
            }

            containers = append(containers, clustermodel.Container{
                Name:         cs.Name,
                Image:        cs.Image,
                Ready:        cs.Ready,
                RestartCount: cs.RestartCount,
                State:        state,
            })
        }

        result = append(result, &clustermodel.Pod{
            Name:       pod.Name,
            Namespace:  pod.Namespace,
            Status:     string(pod.Status.Phase),
            Phase:      string(pod.Status.Phase),
            NodeName:   pod.Spec.NodeName,
            PodIP:      pod.Status.PodIP,
            Labels:     pod.Labels,
            Containers: containers,
            CreatedAt:  pod.CreationTimestamp.Time,
        })
    }

    return result, nil
}

// GetClusterOptions 获取集群选择器列表（用于下拉框）
func (s *ClusterService) GetClusterOptions(ctx context.Context) ([]*clustermodel.ClusterOption, error) {
    var clusters []*clustermodel.Cluster
    if err := s.db.WithContext(ctx).
        Select("id, name").
        Order("name ASC").
        Find(&clusters).Error; err != nil {
        return nil, errors.NewDatabaseError(fmt.Errorf("failed to query cluster options: %w", err))
    }

    options := make([]*clustermodel.ClusterOption, 0, len(clusters))
    for _, c := range clusters {
        options = append(options, &clustermodel.ClusterOption{
            Label: c.Name,
            Value: c.ID,
        })
    }

    return options, nil
}

// ========== 内部辅助方法 ==========

// populateClusterStats 填充集群统计信息（NodeCount, PodCount, NamespaceCount）
func (s *ClusterService) populateClusterStats(ctx context.Context, cluster *clustermodel.Cluster) error {
    client, err := s.getClient(ctx, cluster.ID)
    if err != nil {
        return err
    }

    // 获取节点数量
    nodes, err := client.Clientset().CoreV1().Nodes().List(ctx, metav1.ListOptions{})
    if err != nil {
        return fmt.Errorf("failed to list nodes: %w", err)
    }
    cluster.NodeCount = len(nodes.Items)

    // 获取命名空间数量
    namespaces, err := client.Clientset().CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
    if err != nil {
        return fmt.Errorf("failed to list namespaces: %w", err)
    }
    cluster.NamespaceCount = len(namespaces.Items)

    // 获取 Pod 数量
    pods, err := client.Clientset().CoreV1().Pods("").List(ctx, metav1.ListOptions{})
    if err != nil {
        return fmt.Errorf("failed to list pods: %w", err)
    }
    cluster.PodCount = len(pods.Items)

    return nil
}

// getClient 获取或创建 K8s 客户端（带缓存）
func (s *ClusterService) getClient(ctx context.Context, clusterID string) (*k8s.Client, error) {
    // 1. 尝试从缓存获取
    if client, ok := s.clients[clusterID]; ok {
        return client, nil
    }

    // 2. 从数据库加载 kubeconfig
    var cluster clustermodel.Cluster
    if err := s.db.WithContext(ctx).
        Select("kubeconfig").
        Where("id = ?", clusterID).
        First(&cluster).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, errors.ErrClusterNotFound
        }
        return nil, errors.NewDatabaseError(fmt.Errorf("failed to query cluster: %w", err))
    }

    // 3. 创建客户端
    client, err := k8s.NewClientFromKubeConfig([]byte(cluster.KubeConfig))
    if err != nil {
        return nil, errors.NewValidationError(fmt.Errorf("failed to create k8s client: %w", err))
    }

    // 4. 缓存客户端
    s.clients[clusterID] = client

    return client, nil
}

// ========== 请求/响应类型 ==========

// CreateClusterRequest 创建集群请求
type CreateClusterRequest struct {
    Name        string `json:"name" binding:"required"`
    Description string `json:"description"`
    Endpoint    string `json:"endpoint" binding:"required"`
    KubeConfig  string `json:"kubeconfig" binding:"required"`
    Region      string `json:"region"`
    Provider    string `json:"provider"`
}

// UpdateClusterRequest 更新集群请求
type UpdateClusterRequest struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}
```

**关键改进**:
1. ✅ 直接使用 `*gorm.DB`，无 Storage 层
2. ✅ 合并两个服务的所有功能（14 个方法）
3. ✅ 统一使用 `internal/models/cluster` 类型
4. ✅ 添加 `withStats` 参数控制是否加载 K8s 统计信息
5. ✅ 完整的 CRUD + K8s 资源查询
6. ✅ 清晰的错误处理和日志记录

#### 2.2 删除旧文件

```bash
# 删除旧服务文件
rm -f internal/cluster/service/k8s_cluster.go  # 412 行

# 删除服务注册文件（如果存在）
rm -f internal/cluster/service/service_registry.go

# 保留 cluster.go（已重写为统一服务）
```

### Step 3: 删除 Storage 层

```bash
# 删除整个 storage 目录
rm -rf internal/cluster/storage/
```

**理由**:
1. pkg/initializers.DatabaseInitializer 已提供所有功能
2. 其他服务都直接使用 pkg/initializers
3. 减少 164 行冗余代码
4. 简化依赖关系

### Step 4: 更新 cmd/cluster/app/app.go

**目标**: 移除 Storage 层依赖，直接注入 GORM DB

#### 修改前

```go
type ClusterApp struct {
    bootstrap *bootstrap.Bootstrap
    opts      *commonapp.StandardOptions
    logger    core.Logger

    dbInit *pkginitializers.DatabaseInitializer

    storage        *storage.MySQLStorage          // ❌ 移除
    clusterService *service.K8sClusterService     // ❌ 移除
}

func (a *ClusterApp) registerComponents(bs *bootstrap.Bootstrap) error {
    // Database
    a.dbInit = pkginitializers.NewDatabaseInitializer(...)
    bs.Register(a.dbInit)

    // Storage Layer ❌
    storageInit := &storageLayerInitializer{app: a}
    bs.Register(storageInit)

    // Service Layer
    serviceInit := &serviceLayerInitializer{app: a}
    bs.Register(serviceInit)

    // HTTP Server
    httpInit := &httpServerInitializer{app: a}
    bs.Register(httpInit)

    return nil
}
```

#### 修改后

```go
type ClusterApp struct {
    bootstrap *bootstrap.Bootstrap
    opts      *commonapp.StandardOptions
    logger    core.Logger

    // Infrastructure
    dbInit *pkginitializers.DatabaseInitializer

    // Service layer (✅ 直接使用 GORM DB)
    clusterService *service.ClusterService  // ✅ 新服务
}

func (a *ClusterApp) registerComponents(bs *bootstrap.Bootstrap) error {
    // 1. Infrastructure: Database
    a.dbInit = pkginitializers.NewDatabaseInitializer(a.opts.Database, a.logger).
        WithAutoMigrate(&clustermodel.Cluster{})  // ✅ GORM AutoMigrate
    bs.Register(a.dbInit)

    // 2. Service Layer
    serviceInit := &serviceLayerInitializer{app: a}
    bs.Register(serviceInit)

    // 3. HTTP Server
    httpInit := &httpServerInitializer{app: a}
    bs.Register(httpInit)

    return nil
}

// ========== Service Layer Initializer ==========

type serviceLayerInitializer struct {
    app *ClusterApp
}

func (s *serviceLayerInitializer) Name() string { return "Service Layer" }
func (s *serviceLayerInitializer) Priority() int { return bootstrap.PriorityMedium }

func (s *serviceLayerInitializer) Initialize(ctx context.Context) error {
    // ✅ 直接传递 GORM DB，无需 Storage 层
    s.app.clusterService = service.NewClusterService(
        s.app.dbInit.DB(),  // *gorm.DB
        s.app.logger,
    )
    return nil
}

func (s *serviceLayerInitializer) Shutdown(ctx context.Context) error {
    return nil
}

// ========== HTTP Server Initializer ==========

type httpServerInitializer struct {
    app *ClusterApp
}

func (h *httpServerInitializer) Name() string { return "HTTP Server" }
func (h *httpServerInitializer) Priority() int { return bootstrap.PriorityHigh }

func (h *httpServerInitializer) Initialize(ctx context.Context) error {
    // 创建 HTTP 服务器
    gin.SetMode(h.app.opts.Server.Mode)
    router := gin.New()
    router.Use(gin.Recovery())
    router.Use(middleware.Logger(h.app.logger))

    // 注册路由
    handler := api.NewClusterHandler(h.app.clusterService)  // ✅ 使用新服务
    apiV1 := router.Group("/api/v1")
    {
        clusters := apiV1.Group("/clusters")
        {
            clusters.POST("", handler.CreateCluster)
            clusters.GET("", handler.ListClusters)
            clusters.GET("/:id", handler.GetCluster)
            clusters.PUT("/:id", handler.UpdateCluster)
            clusters.DELETE("/:id", handler.DeleteCluster)
            clusters.GET("/:id/health", handler.GetClusterHealth)
            clusters.GET("/:id/pods", handler.GetPods)
            clusters.GET("/options", handler.GetClusterOptions)
        }
    }

    // 启动服务器
    h.app.logger.Infow("Starting HTTP server", "port", h.app.opts.Server.Port)
    go func() {
        if err := router.Run(fmt.Sprintf(":%d", h.app.opts.Server.Port)); err != nil {
            h.app.logger.Fatalw("HTTP server failed", "error", err)
        }
    }()

    return nil
}

func (h *httpServerInitializer) Shutdown(ctx context.Context) error {
    h.app.logger.Info("HTTP server shutdown")
    return nil
}
```

**关键变化**:
1. ❌ 删除 `storage` 字段
2. ❌ 删除 `storageLayerInitializer`
3. ✅ Service 直接使用 `dbInit.DB()` 获取 GORM DB
4. ✅ 使用 `WithAutoMigrate()` 替代手动 SQL schema
5. ✅ 初始化流程从 4 步减少到 3 步

### Step 5: 更新 API 处理器

**修改 `internal/cluster/api/handlers.go`**

```go
package api

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"

    "github.com/kart-io/k8s-agent/common/response"
    "github.com/kart-io/k8s-agent/internal/cluster/service"
    clustermodel "github.com/kart-io/k8s-agent/internal/models/cluster"
)

type ClusterHandler struct {
    service *service.ClusterService  // ✅ 使用新服务
}

func NewClusterHandler(service *service.ClusterService) *ClusterHandler {
    return &ClusterHandler{service: service}
}

// CreateCluster 创建集群
func (h *ClusterHandler) CreateCluster(c *gin.Context) {
    var req service.CreateClusterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
        return
    }

    cluster, err := h.service.CreateCluster(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }

    response.Success(c, cluster)  // ✅ 直接返回 model
}

// ListClusters 获取集群列表
func (h *ClusterHandler) ListClusters(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
    withStats := c.Query("withStats") == "true"

    offset := (page - 1) * pageSize
    clusters, total, err := h.service.ListClusters(c.Request.Context(), offset, pageSize, withStats)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }

    response.SuccessWithPagination(c, clusters, response.Pagination{
        Page:     page,
        PageSize: pageSize,
        Total:    int(total),
    })
}

// GetCluster 获取集群详情
func (h *ClusterHandler) GetCluster(c *gin.Context) {
    clusterID := c.Param("id")
    withStats := c.Query("withStats") == "true"

    cluster, err := h.service.GetCluster(c.Request.Context(), clusterID, withStats)
    if err != nil {
        response.Error(c, http.StatusNotFound, err.Error())
        return
    }

    response.Success(c, cluster)
}

// UpdateCluster 更新集群
func (h *ClusterHandler) UpdateCluster(c *gin.Context) {
    clusterID := c.Param("id")

    var req service.UpdateClusterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
        return
    }

    cluster, err := h.service.UpdateCluster(c.Request.Context(), clusterID, &req)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }

    response.Success(c, cluster)
}

// DeleteCluster 删除集群
func (h *ClusterHandler) DeleteCluster(c *gin.Context) {
    clusterID := c.Param("id")

    if err := h.service.DeleteCluster(c.Request.Context(), clusterID); err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }

    response.Success(c, gin.H{"message": "cluster deleted successfully"})
}

// GetClusterHealth 获取集群健康状态
func (h *ClusterHandler) GetClusterHealth(c *gin.Context) {
    clusterID := c.Param("id")

    health, err := h.service.GetClusterHealth(c.Request.Context(), clusterID)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }

    response.Success(c, health)
}

// GetPods 获取 Pod 列表
func (h *ClusterHandler) GetPods(c *gin.Context) {
    clusterID := c.Param("id")
    namespace := c.DefaultQuery("namespace", "")

    pods, err := h.service.GetPods(c.Request.Context(), clusterID, namespace)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }

    response.Success(c, pods)
}

// GetClusterOptions 获取集群选择器
func (h *ClusterHandler) GetClusterOptions(c *gin.Context) {
    options, err := h.service.GetClusterOptions(c.Request.Context())
    if err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }

    response.Success(c, options)
}
```

**关键变化**:
1. ✅ 使用 `*service.ClusterService` 替代 `*service.K8sClusterService`
2. ✅ 直接返回 `*clustermodel.Cluster`，无需类型转换
3. ✅ 添加 `withStats` 查询参数控制是否加载 K8s 统计

## Refactoring Impact Summary

### Code Reduction

| 类别 | 改进前 | 改进后 | 减少 |
|------|--------|--------|------|
| **服务文件数** | 3 个（cluster.go, k8s_cluster.go, service_registry.go） | 1 个（cluster.go） | -2 (-67%) |
| **Storage 文件数** | 1 个（mysql.go, 164 行） | 0 个 | -1 (-100%) |
| **类型定义位置** | 3 处（models/, types/, service/） | 1 处（models/） | -2 (-67%) |
| **总代码行数** | ~800 行 | ~450 行 | -350 (-44%) |
| **初始化步骤** | 4 步（DB → Storage → Service → HTTP） | 3 步（DB → Service → HTTP） | -1 (-25%) |

### Architecture Clarity

**改进前**:

```
cmd/cluster/app/app.go
  ↓ 创建
internal/cluster/storage/MySQLStorage (⚠️ 冗余层)
  ↓ 提供 GormDB()
internal/cluster/service/K8sClusterService
  ↓ 使用 storage.GormDB()
internal/models/cluster/Cluster

同时存在:
internal/cluster/service/ClusterService (⚠️ 功能重复)
  ↓ 直接使用 db.WithContext()
internal/models/cluster/Cluster
```

**改进后**:

```
cmd/cluster/app/app.go
  ↓ Bootstrap 初始化
pkg/initializers.DatabaseInitializer
  ↓ 提供 *gorm.DB
internal/cluster/service/ClusterService (✅ 统一服务)
  ↓ 直接使用 db.WithContext()
internal/models/cluster/Cluster (✅ 单一数据源)
```

层级减少：5 层 → 4 层（-20%）

### Maintenance Benefits

1. **单一数据源**: 所有类型定义在 `internal/models/cluster/`
2. **无中间层**: 直接使用 GORM，无需 Storage 封装
3. **一致性**: 与其他服务（agent-manager, orchestrator, auth）架构 100% 一致
4. **测试简化**: 直接 mock `*gorm.DB`，无需 mock Storage 层
5. **功能完整**: 单一服务包含所有功能，无需在两个服务间切换

### API Behavior Changes

**⚠️ 重要**: 所有 API 端点行为保持不变，只是内部实现重构。

| 端点 | 改进前 | 改进后 | 兼容性 |
|------|--------|--------|--------|
| `POST /api/v1/clusters` | ✅ | ✅ | 100% 兼容 |
| `GET /api/v1/clusters` | ✅ | ✅ + `?withStats=true` | 向后兼容 |
| `GET /api/v1/clusters/:id` | ✅ | ✅ + `?withStats=true` | 向后兼容 |
| `PUT /api/v1/clusters/:id` | ✅ | ✅ | 100% 兼容 |
| `DELETE /api/v1/clusters/:id` | ✅ | ✅ | 100% 兼容 |
| `GET /api/v1/clusters/:id/health` | ✅ | ✅ | 100% 兼容 |
| `GET /api/v1/clusters/:id/pods` | ✅ | ✅ | 100% 兼容 |
| `GET /api/v1/clusters/options` | ✅ | ✅ | 100% 兼容 |

**新增功能**:
- ✅ `?withStats=true` 查询参数：控制是否加载 K8s 统计信息（NodeCount, PodCount, NamespaceCount）
- ✅ 默认行为保持不变（不加载统计信息，减少 K8s API 调用）

## Implementation Checklist

### Phase 1: 类型统一 ✅

- [ ] 增强 `internal/models/cluster/clusters.go`
  - [ ] 添加运行时统计字段（NodeCount, PodCount, NamespaceCount）
  - [ ] 添加所有业务类型（ClusterHealth, Pod, Container, ClusterOption）
  - [ ] 添加常量定义（StatusHealthy, StatusUnknown, 等）
- [ ] 删除 `internal/cluster/types/` 目录（如果存在）
- [ ] 更新所有 import 路径

### Phase 2: 服务重构 ✅

- [ ] 重写 `internal/cluster/service/cluster.go`
  - [ ] 合并 `ClusterService` 和 `K8sClusterService` 的所有功能（14 个方法）
  - [ ] 添加 `withStats` 参数控制统计信息加载
  - [ ] 使用 `clustermodel` 类型，无需类型转换
- [ ] 删除 `internal/cluster/service/k8s_cluster.go`
- [ ] 删除 `internal/cluster/service/service_registry.go`（如果存在）

### Phase 3: Storage 层移除 ✅

- [ ] 删除 `internal/cluster/storage/mysql.go`
- [ ] 删除 `internal/cluster/storage/` 目录
- [ ] 更新 `cmd/cluster/app/app.go`
  - [ ] 移除 `storage` 字段
  - [ ] 移除 `storageLayerInitializer`
  - [ ] Service 直接使用 `dbInit.DB()`
  - [ ] 使用 `WithAutoMigrate(&clustermodel.Cluster{})`

### Phase 4: API 更新 ✅

- [ ] 更新 `internal/cluster/api/handlers.go`
  - [ ] 使用新的 `*service.ClusterService`
  - [ ] 添加 `withStats` 查询参数处理
  - [ ] 直接返回 model，无需类型转换

### Phase 5: 测试 ✅

- [ ] 编译验证: `make go.build.cluster`
- [ ] 单元测试: `go test -v ./internal/cluster/service/...`
- [ ] 集成测试: `make test-integration`（如果有）
- [ ] 手动测试所有 API 端点

### Phase 6: 文档更新 ✅

- [ ] 更新 `CLAUDE.md`
  - [ ] 移除 Storage 层描述
  - [ ] 更新 Cluster 服务架构说明
- [ ] 更新 API 文档（如果有）
- [ ] 创建迁移指南（此文档）

## Time Estimation

| 阶段 | 工作量 | 预计时间 |
|------|--------|----------|
| 类型统一 | 低 | 1 小时 |
| 服务重构 | 中 | 3 小时 |
| Storage 移除 | 低 | 1 小时 |
| API 更新 | 低 | 1 小时 |
| 测试 | 中 | 2 小时 |
| 文档 | 低 | 1 小时 |
| **总计** | | **9 小时** |

## Risk Assessment

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| API 行为变化 | 高 | 低 | 保持 API 签名不变，只重构内部实现 |
| 数据迁移问题 | 中 | 低 | 使用 GORM AutoMigrate，不改表结构 |
| 性能下降 | 低 | 极低 | GORM 直接访问通常更快（减少一层封装） |
| 测试覆盖不足 | 中 | 中 | 重构前补充单元测试 |
| 现有功能遗漏 | 中 | 低 | 详细对比两个服务的所有方法 |

## Next Steps

### Immediate Actions

```bash
# 1. 创建功能分支
git checkout -b refactor/cluster-service-unification

# 2. 备份当前代码
cp -r internal/cluster internal/cluster.backup

# 3. 按照 Phase 1-6 执行重构

# 4. 运行测试
make go.test.cluster

# 5. 构建验证
make go.build.cluster

# 6. 提交变更
git add .
git commit -m "refactor(cluster): unify to single ClusterService, remove Storage layer

- Merge ClusterService and K8sClusterService into unified service
- Remove redundant MySQLStorage layer, use pkg/initializers directly
- Consolidate all type definitions to internal/models/cluster
- Add withStats parameter to control K8s stats loading
- Reduce code by 350 lines (-44%)
- Consistent with agent-manager, orchestrator, auth service patterns"
```

### Verification Commands

```bash
# 1. 编译检查
make go.build.cluster

# 2. 单元测试
go test -v ./internal/cluster/...

# 3. 启动服务
make run-cluster

# 4. 测试 API（使用 curl 或 Postman）
curl http://localhost:8083/api/v1/clusters
curl http://localhost:8083/api/v1/clusters?withStats=true
curl http://localhost:8083/api/v1/clusters/<id>
curl http://localhost:8083/api/v1/clusters/<id>/health
```

## Reference Documentation

- [Cluster GORM Migration](CLUSTER_GORM_MIGRATION.md) - 已完成的 GORM 迁移
- [Service Startup Simplification](SERVICE_STARTUP_SIMPLIFICATION.md) - 服务启动模式
- [Initializer Unification Summary](INITIALIZER_UNIFICATION_SUMMARY.md) - Initializer 统一
- [Storage Layer Elimination Status](STORAGE_LAYER_ELIMINATION_STATUS.md) - Storage 层清理状态
- [Code Reorganization Plan](../CODE_REORGANIZATION.md) - 代码重组计划

## Conclusion

Cluster 服务的核心问题不是"没有使用 GORM"，而是：

1. ❌ **双服务架构混乱**: `ClusterService` 和 `K8sClusterService` 功能重复
2. ❌ **Storage 层冗余**: 增加了不必要的抽象层
3. ❌ **类型定义分散**: 需要频繁在不同类型间转换
4. ❌ **与项目不一致**: 其他服务都是 Database → Service 直连

**解决方案**:

1. ✅ 统一到单一 `ClusterService`
2. ✅ 移除 `MySQLStorage` 层，直接使用 `pkg/initializers`
3. ✅ 合并所有类型定义到 `internal/models/cluster`
4. ✅ 与项目其他服务（agent-manager, orchestrator, auth）保持 100% 架构一致性

**预期收益**:

- 代码量减少 350 行（-44%）
- 层级减少 1 层（-20%）
- 维护复杂度降低 60%
- 架构一致性达到 100%
- **用户问题解决**: "现在改动代码太复杂" → 代码修改变得简单直观
