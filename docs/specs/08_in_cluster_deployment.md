# Kubernetes 集群内部署实现指南

## 文档信息

- **版本**: v1.6
- **最后更新**: 2025年9月27日
- **状态**: 正式版
- **所属系统**: Aetherius AI Agent
- **文档类型**: In-Cluster 部署实现指南

> **文档关联**:
> - **核心架构**: [02_architecture.md](./02_architecture.md) - 系统整体架构设计
> - **通用部署**: [04_deployment.md](./04_deployment.md) - 通用部署配置指南
> - **对比方案**: [09_agent_proxy_mode.md](./09_agent_proxy_mode.md) - Agent代理模式

## 目录

- [1. In-Cluster 部署概述](#1-in-cluster-部署概述)
- [2. 部署架构设计](#2-部署架构设计)
- [3. 核心实现方案](#3-核心实现方案)
- [4. 完整部署配置](#4-完整部署配置)
- [5. 多集群监控方案](#5-多集群监控方案)
- [6. 安全与权限](#6-安全与权限)

## 1. In-Cluster 部署概述

### 1.1 什么是 In-Cluster 部署

**In-Cluster** 部署指将 Aetherius 服务直接部署在 Kubernetes 集群内部,利用集群的 ServiceAccount 和 RBAC 机制进行认证和授权。

### 1.2 部署模式对比

> **模式选择指南**: 根据集群数量和管理需求选择合适的部署模式

| 部署模式 | 核心特点 | 主要优势 | 主要限制 | 最佳适用场景 | 推荐度 |
|---------|----------|----------|----------|-------------|--------|
| **In-Cluster**<br>(本文档) | 完整部署在单个集群内 | • 无需外部kubeconfig<br>• ServiceAccount自动认证<br>• 集群内网络高效<br>• 部署配置简单 | • 每个集群需独立部署<br>• 无法跨集群统一管理<br>• 资源占用相对较高 | **单集群环境**<br>• 独立的生产集群<br>• 开发/测试环境<br>• 网络隔离严格的环境 | ⭐⭐⭐⭐⭐ |
| **Out-of-Cluster** | 中央控制平面+远程kubeconfig | • 集中管理多个集群<br>• 统一的控制界面<br>• 配置和策略统一 | • 需要管理多个kubeconfig<br>• 网络连通性要求高<br>• 安全性配置复杂 | **传统多集群管理**<br>• 集群数量较少(2-5个)<br>• 网络连通性良好 | ⭐⭐⭐ |
| **Agent代理模式**<br>([详见09文档](./09_agent_proxy_mode.md)) | 中央控制平面+轻量级Agent | • 兼具两者优势<br>• Agent主动上报<br>• 网络要求较低<br>• 安全隔离性好 | • 架构相对复杂<br>• 需要维护Agent<br>• 消息总线依赖 | **大规模多集群**<br>• 多云/混合云环境<br>• 网络条件复杂<br>• 安全要求较高 | ⭐⭐⭐⭐⭐ |

### 1.3 In-Cluster 部署架构

```
┌────────────────────────────────────────────────────────────────┐
│                   Kubernetes Cluster                           │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │           aetherius-system namespace                     │  │
│  │                                                          │  │
│  │  ┌─────────────────────────────────────────────────┐    │  │
│  │  │         Control Plane (中央控制平面)             │    │  │
│  │  │  ┌──────────────┐  ┌──────────────────────────┐ │    │  │
│  │  │  │ API Gateway  │  │  Dashboard Web           │ │    │  │
│  │  │  └──────────────┘  └──────────────────────────┘ │    │  │
│  │  │  ┌──────────────┐  ┌──────────────────────────┐ │    │  │
│  │  │  │ Orchestrator │  │  Reasoning Service       │ │    │  │
│  │  │  └──────────────┘  └──────────────────────────┘ │    │  │
│  │  └─────────────────────────────────────────────────┘    │  │
│  │                                                          │  │
│  │  ┌─────────────────────────────────────────────────┐    │  │
│  │  │         Event Collection (事件收集层)            │    │  │
│  │  │  ┌──────────────┐  ┌──────────────────────────┐ │    │  │
│  │  │  │Event Gateway │  │  K8s Event Watcher       │ │    │  │
│  │  │  │(Webhook)     │  │  (In-Cluster Informer)   │ │    │  │
│  │  │  └──────────────┘  └──────────────────────────┘ │    │  │
│  │  └─────────────────────────────────────────────────┘    │  │
│  │                         │                                │  │
│  │                         │ Watch K8s Events               │  │
│  │                         ▼                                │  │
│  │  ┌──────────────────────────────────────────────────┐   │  │
│  │  │         Kubernetes API Server                    │   │  │
│  │  │  - /api/v1/events                                │   │  │
│  │  │  - /api/v1/pods                                  │   │  │
│  │  │  - /api/v1/namespaces/{ns}/events               │   │  │
│  │  └──────────────────────────────────────────────────┘   │  │
│  │                                                          │  │
│  │  ┌─────────────────────────────────────────────────┐    │  │
│  │  │         Data Layer (数据层)                      │    │  │
│  │  │  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │    │  │
│  │  │  │PostgreSQL│  │  Redis   │  │  Weaviate    │  │    │  │
│  │  │  └──────────┘  └──────────┘  └──────────────┘  │    │  │
│  │  └─────────────────────────────────────────────────┘    │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                │
│  认证方式: ServiceAccount Token (自动挂载)                      │
│  网络: ClusterIP Service (集群内部通信)                         │
└────────────────────────────────────────────────────────────────┘
```

## 2. 部署架构设计

### 2.1 单集群 In-Cluster 部署

适用于单个 Kubernetes 集群的监控场景。

```
                    ┌─────────────────────────────────┐
                    │   Kubernetes API Server         │
                    └────────────┬────────────────────┘
                                 │
                    ┌────────────▼────────────────────┐
                    │    ServiceAccount Token         │
                    │    (自动挂载到Pod)               │
                    └────────────┬────────────────────┘
                                 │
              ┌──────────────────┼──────────────────┐
              │                  │                  │
              ▼                  ▼                  ▼
    ┌──────────────────┐ ┌─────────────┐ ┌─────────────┐
    │  Event Watcher   │ │ Orchestrator│ │  Execution  │
    │  (Deployment)    │ │ (Deployment)│ │  Gateway    │
    │  Replicas: 1     │ │ Replicas: 3 │ │ Replicas: 2 │
    └──────────────────┘ └─────────────┘ └─────────────┘
            │                    │                │
            └────────────────────┼────────────────┘
                                 │
                    ┌────────────▼────────────────────┐
                    │      Message Bus (NATS)         │
                    └─────────────────────────────────┘
```

**特点**:
- ✅ 部署简单,无需外部配置
- ✅ 自动服务发现
- ✅ 安全性高 (ServiceAccount)
- ⚠️ 仅限当前集群

### 2.2 多集群Agent代理部署

适用于需要监控多个 Kubernetes 集群的场景。详细实现请参考 [Agent代理模式文档](./09_agent_proxy_mode.md)。

```
┌─────────────────────────────────────────────────────────────────┐
│                     Central Cluster (中心集群)                   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  Aetherius Control Plane (中央控制平面)                  │   │
│  │  - Orchestrator                                         │   │
│  │  - Reasoning Service                                    │   │
│  │  - Knowledge Service                                    │   │
│  │  - Report Service                                       │   │
│  │  - API Service                                          │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              ▲                                  │
└──────────────────────────────┼──────────────────────────────────┘
                               │
                     ┌─────────┴─────────┐
                     │  Message Bus      │
                     │  (NATS Cluster)   │
                     └─────────┬─────────┘
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
        ▼                      ▼                      ▼
┌───────────────┐      ┌───────────────┐      ┌───────────────┐
│  Cluster A    │      │  Cluster B    │      │  Cluster C    │
│  (prod-us-1)  │      │  (prod-eu-1)  │      │  (prod-asia)  │
│               │      │               │      │               │
│  ┌─────────┐  │      │  ┌─────────┐  │      │  ┌─────────┐  │
│  │ Event   │  │      │  │ Event   │  │      │  │ Event   │  │
│  │ Watcher │──┼──────┼─►│ Watcher │──┼──────┼─►│ Watcher │  │
│  │(Agent)  │  │      │  │(Agent)  │  │      │  │(Agent)  │  │
│  └─────────┘  │      │  └─────────┘  │      │  └─────────┘  │
│       │       │      │       │       │      │       │       │
│  ┌────▼────┐  │      │  ┌────▼────┐  │      │  ┌────▼────┐  │
│  │   K8s   │  │      │  │   K8s   │  │      │  │   K8s   │  │
│  │   API   │  │      │  │   API   │  │      │  │   API   │  │
│  └─────────┘  │      │  └─────────┘  │      │  └─────────┘  │
└───────────────┘      └───────────────┘      └───────────────┘
```

**特点**:
- ✅ 统一管理多个集群
- ✅ 边缘Agent轻量化
- ✅ 数据集中存储和分析
- ⚠️ 需要跨集群网络连通

### 2.3 推荐架构选择

> **提示**: 详细的部署模式对比请参考[第1.2节](#12-部署模式对比)

| 场景 | 推荐架构 | 说明 | 参考文档 |
|------|---------|------|----------|
| **单一生产集群** | 单集群In-Cluster | 简单可靠,全功能部署 | 本文档 |
| **2-10个集群** | Agent代理模式 | 轻量级Agent+中央控制平面 | [09_agent_proxy_mode.md](./09_agent_proxy_mode.md) |
| **10+个集群** | 分层Agent架构 | Hub-Spoke分层管理 | 待规划 |
| **多云/混合云** | Agent代理模式 | Agent主动连接,跨云友好 | [09_agent_proxy_mode.md](./09_agent_proxy_mode.md) |

## 3. 核心实现方案

### 3.1 In-Cluster 客户端创建

```go
package k8s

import (
    "fmt"
    "os"

    "go.uber.org/zap"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
)

// 全局日志实例(实际使用时应通过依赖注入)
var log *zap.Logger

// ClientFactory 客户端工厂
type ClientFactory struct {
    inCluster bool
}

// NewClientFactory 创建客户端工厂
func NewClientFactory() *ClientFactory {
    // 检测是否在集群内运行
    inCluster := isInCluster()
    return &ClientFactory{
        inCluster: inCluster,
    }
}

// CreateClient 创建Kubernetes客户端
func (f *ClientFactory) CreateClient() (kubernetes.Interface, error) {
    var config *rest.Config
    var err error

    if f.inCluster {
        // In-Cluster 配置 (推荐)
        config, err = rest.InClusterConfig()
        if err != nil {
            return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
        }
        log.Info("Using in-cluster configuration")
    } else {
        // Out-of-Cluster 配置 (开发环境)
        kubeconfig := os.Getenv("KUBECONFIG")
        if kubeconfig == "" {
            kubeconfig = os.Getenv("HOME") + "/.kube/config"
        }

        config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
        if err != nil {
            return nil, fmt.Errorf("failed to create config from kubeconfig: %w", err)
        }
        log.Info("Using out-of-cluster configuration", zap.String("kubeconfig", kubeconfig))
    }

    // 配置优化
    config.QPS = 50      // 每秒查询数
    config.Burst = 100   // 突发查询数

    // 创建客户端
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create clientset: %w", err)
    }

    return clientset, nil
}

// isInCluster 检测是否在集群内运行
func isInCluster() bool {
    // 检查是否存在 ServiceAccount token
    _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token")
    return err == nil
}
```

### 3.2 自动集群 ID 检测

```go
// ClusterIDDetector 集群ID检测器
// 注意: 本代码段与3.1节共享相同的import声明
type ClusterIDDetector struct {
    clientset kubernetes.Interface
}

// DetectClusterID 自动检测集群ID
func (d *ClusterIDDetector) DetectClusterID(ctx context.Context) (string, error) {
    // 方法1: 从环境变量获取 (优先级最高)
    if clusterID := os.Getenv("CLUSTER_ID"); clusterID != "" {
        log.Info("Cluster ID from environment variable", zap.String("cluster_id", clusterID))
        return clusterID, nil
    }

    // 方法2: 从 ConfigMap 获取
    clusterID, err := d.getClusterIDFromConfigMap(ctx)
    if err == nil && clusterID != "" {
        log.Info("Cluster ID from ConfigMap", zap.String("cluster_id", clusterID))
        return clusterID, nil
    }

    // 方法3: 从集群信息推断 (UID)
    clusterID, err = d.getClusterIDFromUID(ctx)
    if err == nil && clusterID != "" {
        log.Info("Cluster ID from kube-system namespace UID", zap.String("cluster_id", clusterID))
        return clusterID, nil
    }

    // 方法4: 从云provider metadata获取
    clusterID, err = d.getClusterIDFromCloudProvider(ctx)
    if err == nil && clusterID != "" {
        log.Info("Cluster ID from cloud provider", zap.String("cluster_id", clusterID))
        return clusterID, nil
    }

    return "", fmt.Errorf("failed to detect cluster ID")
}

// 从 ConfigMap 获取集群ID
func (d *ClusterIDDetector) getClusterIDFromConfigMap(ctx context.Context) (string, error) {
    cm, err := d.clientset.CoreV1().ConfigMaps("aetherius-system").
        Get(ctx, "aetherius-config", metav1.GetOptions{})
    if err != nil {
        return "", err
    }

    return cm.Data["cluster_id"], nil
}

// 从 kube-system namespace UID 生成集群ID
func (d *ClusterIDDetector) getClusterIDFromUID(ctx context.Context) (string, error) {
    ns, err := d.clientset.CoreV1().Namespaces().
        Get(ctx, "kube-system", metav1.GetOptions{})
    if err != nil {
        return "", err
    }

    // 使用 kube-system namespace 的 UID 作为集群唯一标识
    clusterID := fmt.Sprintf("cluster-%s", ns.UID[:8])
    return clusterID, nil
}

// 从云provider metadata获取
func (d *ClusterIDDetector) getClusterIDFromCloudProvider(ctx context.Context) (string, error) {
    // 尝试从Node标签获取
    nodes, err := d.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
    if err != nil || len(nodes.Items) == 0 {
        return "", fmt.Errorf("no nodes found")
    }

    node := nodes.Items[0]

    // AWS EKS
    if clusterName, ok := node.Labels["alpha.eksctl.io/cluster-name"]; ok {
        return clusterName, nil
    }

    // GKE
    if clusterName, ok := node.Labels["cloud.google.com/gke-cluster-name"]; ok {
        return clusterName, nil
    }

    // AKS
    if clusterName, ok := node.Labels["kubernetes.azure.com/cluster"]; ok {
        return clusterName, nil
    }

    return "", fmt.Errorf("cluster ID not found in cloud provider metadata")
}
```

### 3.3 In-Cluster Event Watcher 实现

```go
// InClusterEventWatcher In-Cluster事件监听器
// 注意: 需要额外导入以下包:
//   "context"
//   "sync/atomic"
//   "time"
//   corev1 "k8s.io/api/core/v1"
//   "k8s.io/client-go/informers"
//   "k8s.io/client-go/tools/cache"
//   metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
type InClusterEventWatcher struct {
    config      WatcherConfig
    clientset   kubernetes.Interface
    clusterID   string
    namespace   string
    eventChan   chan *corev1.Event
    stopCh      chan struct{}
    running     atomic.Bool
}

// NewInClusterEventWatcher 创建In-Cluster事件监听器
func NewInClusterEventWatcher(config WatcherConfig) (*InClusterEventWatcher, error) {
    // 1. 创建In-Cluster客户端
    factory := NewClientFactory()
    clientset, err := factory.CreateClient()
    if err != nil {
        return nil, fmt.Errorf("failed to create k8s client: %w", err)
    }

    // 2. 自动检测集群ID
    detector := &ClusterIDDetector{clientset: clientset}
    clusterID, err := detector.DetectClusterID(context.Background())
    if err != nil {
        return nil, fmt.Errorf("failed to detect cluster ID: %w", err)
    }

    // 3. 获取当前Pod所在的namespace
    namespace := getCurrentNamespace()

    watcher := &InClusterEventWatcher{
        config:    config,
        clientset: clientset,
        clusterID: clusterID,
        namespace: namespace,
        eventChan: make(chan *corev1.Event, 1000),
        stopCh:    make(chan struct{}),
    }

    return watcher, nil
}

// getCurrentNamespace 获取当前Pod所在的namespace
func getCurrentNamespace() string {
    // 从挂载的ServiceAccount信息读取
    nsBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
    if err != nil {
        log.Warn("Failed to read namespace from ServiceAccount, using default",
            zap.Error(err))
        return "aetherius-system"
    }
    return string(nsBytes)
}

// Start 启动监听
func (w *InClusterEventWatcher) Start(ctx context.Context) error {
    if w.running.Load() {
        return fmt.Errorf("watcher already running")
    }

    log.Info("Starting in-cluster event watcher",
        zap.String("cluster_id", w.clusterID),
        zap.String("namespace", w.namespace))

    // 创建Informer工厂
    informerFactory := informers.NewSharedInformerFactory(
        w.clientset,
        30*time.Minute, // resync period
    )

    // 获取Event Informer
    eventInformer := informerFactory.Core().V1().Events().Informer()

    // 注册事件处理器
    eventInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
        AddFunc: func(obj interface{}) {
            event := obj.(*corev1.Event)
            w.handleEvent(event)
        },
        UpdateFunc: func(oldObj, newObj interface{}) {
            newEvent := newObj.(*corev1.Event)
            oldEvent := oldObj.(*corev1.Event)
            // 只处理Count增加的更新
            if newEvent.Count > oldEvent.Count {
                w.handleEvent(newEvent)
            }
        },
    })

    // 启动Informer
    informerFactory.Start(w.stopCh)

    // 等待缓存同步
    log.Info("Waiting for cache sync...")
    if !cache.WaitForCacheSync(w.stopCh, eventInformer.HasSynced) {
        return fmt.Errorf("failed to sync cache")
    }

    log.Info("Cache synced successfully")
    w.running.Store(true)

    // 启动事件处理协程
    go w.processEvents(ctx)

    // 等待停止信号
    <-ctx.Done()
    return w.Stop()
}

// handleEvent 处理事件
func (w *InClusterEventWatcher) handleEvent(event *corev1.Event) {
    select {
    case w.eventChan <- event:
        // 事件发送成功
    default:
        // 缓冲区满,丢弃事件
        log.Warn("Event channel full, dropping event",
            zap.String("reason", event.Reason),
            zap.String("namespace", event.Namespace))
    }
}

// processEvents 处理事件队列
func (w *InClusterEventWatcher) processEvents(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case event := <-w.eventChan:
            if err := w.processEvent(event); err != nil {
                log.Error("Failed to process event",
                    zap.Error(err),
                    zap.String("reason", event.Reason))
            }
        }
    }
}

// Stop 停止监听
func (w *InClusterEventWatcher) Stop() error {
    if !w.running.Load() {
        return nil
    }

    log.Info("Stopping in-cluster event watcher")
    close(w.stopCh)
    w.running.Store(false)

    return nil
}
```

### 3.4 服务健康检查

```go
// HealthChecker In-Cluster健康检查
type HealthChecker struct {
    clientset kubernetes.Interface
    namespace string
}

// CheckHealth 执行健康检查
func (h *HealthChecker) CheckHealth(ctx context.Context) HealthStatus {
    status := HealthStatus{
        Healthy: true,
        Checks:  make(map[string]CheckResult),
    }

    // 1. 检查API Server连通性
    status.Checks["api_server"] = h.checkAPIServer(ctx)

    // 2. 检查自身Pod状态
    status.Checks["self_pod"] = h.checkSelfPod(ctx)

    // 3. 检查依赖服务
    status.Checks["dependencies"] = h.checkDependencies(ctx)

    // 判断整体健康状态
    for _, check := range status.Checks {
        if !check.Healthy {
            status.Healthy = false
            break
        }
    }

    return status
}

// checkAPIServer 检查API Server连通性
func (h *HealthChecker) checkAPIServer(ctx context.Context) CheckResult {
    start := time.Now()

    // 尝试获取API版本信息
    _, err := h.clientset.Discovery().ServerVersion()

    return CheckResult{
        Name:     "api_server",
        Healthy:  err == nil,
        Duration: time.Since(start),
        Error:    err,
    }
}

// checkSelfPod 检查自身Pod状态
func (h *HealthChecker) checkSelfPod(ctx context.Context) CheckResult {
    start := time.Now()

    // 获取当前Pod名称
    podName := os.Getenv("HOSTNAME")
    if podName == "" {
        return CheckResult{
            Name:    "self_pod",
            Healthy: false,
            Error:   fmt.Errorf("HOSTNAME not set"),
        }
    }

    // 获取Pod信息
    pod, err := h.clientset.CoreV1().Pods(h.namespace).
        Get(ctx, podName, metav1.GetOptions{})

    if err != nil {
        return CheckResult{
            Name:     "self_pod",
            Healthy:  false,
            Duration: time.Since(start),
            Error:    err,
        }
    }

    // 检查Pod是否Running
    healthy := pod.Status.Phase == corev1.PodRunning

    return CheckResult{
        Name:     "self_pod",
        Healthy:  healthy,
        Duration: time.Since(start),
        Message:  fmt.Sprintf("Pod phase: %s", pod.Status.Phase),
    }
}
```

## 4. 完整部署配置

### 4.1 Namespace 和 RBAC

```yaml
# 01-namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: aetherius-system
  labels:
    name: aetherius-system
    app.kubernetes.io/name: aetherius
    app.kubernetes.io/part-of: aetherius
---
# 02-serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: aetherius-event-watcher
  namespace: aetherius-system
  labels:
    app.kubernetes.io/name: aetherius
    app.kubernetes.io/component: event-watcher
---
# 03-clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: aetherius-event-watcher
  labels:
    app.kubernetes.io/name: aetherius
rules:
# 事件读取权限
- apiGroups: [""]
  resources: ["events"]
  verbs: ["get", "list", "watch"]

# Pod信息读取 (用于上下文增强)
- apiGroups: [""]
  resources: ["pods", "pods/log"]
  verbs: ["get", "list"]

# Namespace信息读取
- apiGroups: [""]
  resources: ["namespaces"]
  verbs: ["get", "list"]

# Node信息读取 (用于集群ID检测)
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["get", "list"]

# ConfigMap读取 (用于配置)
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list"]
---
# 04-clusterrolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: aetherius-event-watcher
  labels:
    app.kubernetes.io/name: aetherius
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: aetherius-event-watcher
subjects:
- kind: ServiceAccount
  name: aetherius-event-watcher
  namespace: aetherius-system
```

### 4.2 ConfigMap 配置

```yaml
# 05-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: aetherius-config
  namespace: aetherius-system
  labels:
    app.kubernetes.io/name: aetherius
data:
  # 集群配置
  cluster_id: "prod-us-west-2"  # 可选,不设置则自动检测

  # 事件监听配置
  watcher.yaml: |
    # 监听配置
    namespaces: []  # 空表示所有namespace
    resync_period: 30m

    # 过滤配置
    event_types:
      - "Warning"

    reason_filters:
      - "OOMKilled"
      - "CrashLoopBackOff"
      - "ImagePullBackOff"
      - "ErrImagePull"
      - "FailedScheduling"
      - "Unhealthy"
      - "BackOff"
      - "FailedMount"

    min_severity: "Warning"

    # 性能配置
    worker_count: 5
    buffer_size: 1000

  # 应用配置
  app.yaml: |
    log_level: "info"
    log_format: "json"

    # 消息总线
    message_bus:
      url: "nats://nats.aetherius-system.svc.cluster.local:4222"
      reconnect_delay: 2s
      max_reconnect_attempts: 10

    # 指标
    metrics:
      enabled: true
      port: 9090
      path: "/metrics"

    # 健康检查
    health:
      port: 8080
      liveness_path: "/health/live"
      readiness_path: "/health/ready"
```

### 4.3 Deployment 部署

```yaml
# 06-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aetherius-event-watcher
  namespace: aetherius-system
  labels:
    app.kubernetes.io/name: aetherius
    app.kubernetes.io/component: event-watcher
    app.kubernetes.io/version: "v1.0"
spec:
  # 每个集群只需一个实例即可
  replicas: 1

  strategy:
    type: Recreate  # 确保同时只有一个实例运行

  selector:
    matchLabels:
      app.kubernetes.io/name: aetherius
      app.kubernetes.io/component: event-watcher

  template:
    metadata:
      labels:
        app.kubernetes.io/name: aetherius
        app.kubernetes.io/component: event-watcher
        app.kubernetes.io/version: "v1.0"
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"

    spec:
      serviceAccountName: aetherius-event-watcher

      # 安全上下文
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000

      containers:
      - name: event-watcher
        image: aetherius/event-watcher:v1.0
        imagePullPolicy: IfNotPresent

        # 命令参数
        args:
        - --in-cluster=true  # 启用In-Cluster模式
        - --config=/etc/aetherius/watcher.yaml
        - --log-level=$(LOG_LEVEL)

        # 环境变量
        env:
        # 自动注入Pod信息
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName

        # 应用配置
        - name: LOG_LEVEL
          value: "info"
        - name: CLUSTER_ID
          valueFrom:
            configMapKeyRef:
              name: aetherius-config
              key: cluster_id
              optional: true  # 可选,不存在则自动检测

        # 端口
        ports:
        - name: http
          containerPort: 8080
          protocol: TCP
        - name: metrics
          containerPort: 9090
          protocol: TCP

        # 挂载配置
        volumeMounts:
        - name: config
          mountPath: /etc/aetherius
          readOnly: true

        # 资源限制
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "250m"

        # 健康检查
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3

        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3

        # 安全上下文
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL

      # 挂载卷
      volumes:
      - name: config
        configMap:
          name: aetherius-config

      # 节点亲和性 (可选)
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app.kubernetes.io/component
                  operator: In
                  values:
                  - event-watcher
              topologyKey: kubernetes.io/hostname

      # 容忍度
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
```

### 4.4 Service 和监控

```yaml
# 07-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: aetherius-event-watcher
  namespace: aetherius-system
  labels:
    app.kubernetes.io/name: aetherius
    app.kubernetes.io/component: event-watcher
spec:
  selector:
    app.kubernetes.io/name: aetherius
    app.kubernetes.io/component: event-watcher
  ports:
  - name: http
    port: 8080
    targetPort: 8080
    protocol: TCP
  - name: metrics
    port: 9090
    targetPort: 9090
    protocol: TCP
  type: ClusterIP
---
# 08-servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: aetherius-event-watcher
  namespace: aetherius-system
  labels:
    app.kubernetes.io/name: aetherius
    app.kubernetes.io/component: event-watcher
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: aetherius
      app.kubernetes.io/component: event-watcher
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
```

### 4.5 Helm Chart 打包

```yaml
# Chart.yaml
apiVersion: v2
name: aetherius
description: Aetherius AI Agent for Kubernetes diagnostics
type: application
version: 1.0.0
appVersion: "1.0.0"
keywords:
  - kubernetes
  - ai
  - diagnostics
  - monitoring
maintainers:
  - name: Aetherius Team
    email: team@aetherius.io
---
# values.yaml
# 集群配置
cluster:
  id: ""  # 留空则自动检测
  name: ""

# Event Watcher配置
eventWatcher:
  enabled: true
  replicas: 1

  image:
    repository: aetherius/event-watcher
    tag: v1.0
    pullPolicy: IfNotPresent

  resources:
    requests:
      memory: "128Mi"
      cpu: "100m"
    limits:
      memory: "256Mi"
      cpu: "250m"

  config:
    namespaces: []
    eventTypes:
      - "Warning"
    reasonFilters:
      - "OOMKilled"
      - "CrashLoopBackOff"
      - "ImagePullBackOff"
      - "FailedScheduling"
    minSeverity: "Warning"

# RBAC配置
rbac:
  create: true

serviceAccount:
  create: true
  name: aetherius-event-watcher

# 监控配置
monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
    interval: 30s
```

## 5. 多集群监控方案

### 5.1 DaemonSet 方式 (每节点一个)

适用于需要节点级别监控的场景。

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: aetherius-node-agent
  namespace: aetherius-system
spec:
  selector:
    matchLabels:
      app: aetherius-node-agent
  template:
    metadata:
      labels:
        app: aetherius-node-agent
    spec:
      serviceAccountName: aetherius-event-watcher
      hostNetwork: true  # 使用主机网络

      containers:
      - name: node-agent
        image: aetherius/node-agent:v1.0

        env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName

        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"

        securityContext:
          privileged: false
          readOnlyRootFilesystem: true
```

### 5.2 边缘 Agent 模式

```yaml
# edge-agent.yaml - 轻量级边缘Agent
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aetherius-edge-agent
  namespace: aetherius-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: aetherius-edge-agent
  template:
    metadata:
      labels:
        app: aetherius-edge-agent
    spec:
      serviceAccountName: aetherius-event-watcher

      containers:
      - name: edge-agent
        image: aetherius/edge-agent:v1.0

        args:
        - --mode=edge  # 边缘模式
        - --central-endpoint=$(CENTRAL_ENDPOINT)

        env:
        - name: CENTRAL_ENDPOINT
          value: "https://aetherius-central.example.com"
        - name: AGENT_TOKEN
          valueFrom:
            secretKeyRef:
              name: edge-agent-token
              key: token

        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
```

### 5.3 部署脚本

```bash
#!/bin/bash
# deploy-to-cluster.sh - 一键部署脚本

set -e

CLUSTER_ID=${1:-""}
KUBECONFIG=${2:-"$HOME/.kube/config"}

if [ -z "$CLUSTER_ID" ]; then
    echo "Usage: $0 <cluster-id> [kubeconfig]"
    exit 1
fi

echo "=== Deploying Aetherius to cluster: $CLUSTER_ID ==="

# 1. 创建namespace
kubectl --kubeconfig=$KUBECONFIG apply -f manifests/01-namespace.yaml

# 2. 创建RBAC
kubectl --kubeconfig=$KUBECONFIG apply -f manifests/02-serviceaccount.yaml
kubectl --kubeconfig=$KUBECONFIG apply -f manifests/03-clusterrole.yaml
kubectl --kubeconfig=$KUBECONFIG apply -f manifests/04-clusterrolebinding.yaml

# 3. 创建ConfigMap (注入cluster_id)
cat manifests/05-configmap.yaml | \
    sed "s/cluster_id: \"\"/cluster_id: \"$CLUSTER_ID\"/" | \
    kubectl --kubeconfig=$KUBECONFIG apply -f -

# 4. 部署应用
kubectl --kubeconfig=$KUBECONFIG apply -f manifests/06-deployment.yaml
kubectl --kubeconfig=$KUBECONFIG apply -f manifests/07-service.yaml
kubectl --kubeconfig=$KUBECONFIG apply -f manifests/08-servicemonitor.yaml

# 5. 等待就绪
echo "Waiting for deployment to be ready..."
kubectl --kubeconfig=$KUBECONFIG -n aetherius-system \
    rollout status deployment/aetherius-event-watcher --timeout=5m

# 6. 验证
echo "Verifying deployment..."
kubectl --kubeconfig=$KUBECONFIG -n aetherius-system \
    get pods -l app.kubernetes.io/component=event-watcher

echo "=== Deployment completed successfully ==="
echo "Check logs: kubectl -n aetherius-system logs -l app.kubernetes.io/component=event-watcher"
```

## 6. 安全与权限

### 6.1 最小权限原则

**安全承诺与权限说明**:

Aetherius系统遵循最小权限原则，仅请求AI诊断所需的**只读权限**。虽然某些权限在技术上可访问敏感数据，但系统在设计和实现上提供多层安全保证：

#### 系统安全保证

1. **严格只读**: 系统架构设计上不包含任何写操作（create、update、patch、delete），从根本上避免对集群的修改风险
2. **按需访问**: 仅在诊断任务执行时按需读取相关资源，而非持续收集所有数据
3. **完整审计**: 所有读取操作都记录在审计日志中，包括访问时间、资源类型、诊断上下文
4. **范围限制**: 通过resourceNames限制配置访问范围，避免读取应用敏感配置
5. **本地处理**: 敏感数据（如日志）仅在诊断分析时临时读取，不进行长期存储

#### 权限用途说明

```yaml
# 诊断只读权限 - 推荐配置
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: aetherius-diagnostic-readonly
rules:
# 事件监听 - 核心功能，监听集群事件触发诊断
- apiGroups: [""]
  resources: ["events"]
  verbs: ["get", "list", "watch"]

# Pod诊断信息 - 故障分析必需
# pods: 获取Pod状态、资源使用、重启次数等诊断信息
# pods/log: 读取日志用于错误分析（仅在诊断任务中按需读取）
- apiGroups: [""]
  resources: ["pods", "pods/status", "pods/log"]
  verbs: ["get", "list"]

# 集群上下文信息 - 提供诊断背景
# 用于了解集群拓扑、资源分布、服务关系等
- apiGroups: [""]
  resources: ["nodes", "namespaces", "services", "endpoints"]
  verbs: ["get", "list"]

# 系统配置读取 - 仅限Aetherius自身配置
# resourceNames限制确保只能读取系统自身配置，不涉及业务应用配置
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "list"]
  resourceNames: ["aetherius-*"]  # 严格限制只能访问系统自身配置

# 明确禁止所有写操作
# 不授予以下权限: create, update, patch, delete, deletecollection
# 系统架构设计上不支持任何修改集群状态的操作
```

#### 额外安全建议

对于高安全要求的生产环境，可以进一步限制权限：

```yaml
# 更严格的权限配置示例
rules:
# 限制Pod日志读取的namespace范围
- apiGroups: [""]
  resources: ["pods/log"]
  verbs: ["get"]
  namespaces: ["app-namespace-1", "app-namespace-2"]  # 仅允许特定namespace

# 或使用labelSelector进一步限制
# 在实际部署中，可通过admission webhook实现细粒度控制
```

### 6.2 网络策略

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: aetherius-event-watcher
  namespace: aetherius-system
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/component: event-watcher

  policyTypes:
  - Ingress
  - Egress

  # 入站规则
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: monitoring
    ports:
    - protocol: TCP
      port: 9090  # Prometheus metrics

  # 出站规则
  egress:
  # 允许访问K8s API Server
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: TCP
      port: 443

  # 允许访问NATS
  - to:
    - podSelector:
        matchLabels:
          app: nats
    ports:
    - protocol: TCP
      port: 4222

  # 允许DNS解析
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: UDP
      port: 53
```

### 6.3 Pod Security Policy

```yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: aetherius-restricted
spec:
  privileged: false
  allowPrivilegeEscalation: false

  requiredDropCapabilities:
  - ALL

  volumes:
  - configMap
  - emptyDir
  - projected
  - secret
  - downwardAPI

  runAsUser:
    rule: MustRunAsNonRoot

  seLinux:
    rule: RunAsAny

  fsGroup:
    rule: RunAsAny

  readOnlyRootFilesystem: true
```

## 附录

### A. 故障排查

#### 问题1: Pod无法访问API Server

```bash
# 检查ServiceAccount
kubectl describe sa aetherius-event-watcher -n aetherius-system

# 检查RBAC权限
kubectl auth can-i list events --as=system:serviceaccount:aetherius-system:aetherius-event-watcher

# 查看Pod日志
kubectl logs -n aetherius-system deployment/aetherius-event-watcher
```

#### 问题2: 集群ID检测失败

```bash
# 手动设置集群ID
kubectl set env deployment/aetherius-event-watcher \
  -n aetherius-system \
  CLUSTER_ID=my-cluster-id

# 或修改ConfigMap
kubectl edit cm aetherius-config -n aetherius-system
```

### B. 相关文档

- [K8s事件监听文档](./07_k8s_event_watcher.md)
- [微服务架构文档](./06_microservices.md)
- [部署配置文档](./04_deployment.md)

### C. Helm 部署命令

```bash
# 添加Helm仓库
helm repo add aetherius https://charts.aetherius.io
helm repo update

# 安装到集群
helm install aetherius aetherius/aetherius \
  --namespace aetherius-system \
  --create-namespace \
  --set cluster.id=prod-us-west-2

# 更新
helm upgrade aetherius aetherius/aetherius \
  --namespace aetherius-system \
  --reuse-values

# 卸载
helm uninstall aetherius -n aetherius-system
```