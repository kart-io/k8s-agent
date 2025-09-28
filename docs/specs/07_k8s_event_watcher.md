# Kubernetes 事件监听实现指南

## 文档信息

- **版本**: v1.6
- **最后更新**: 2025-09-28
- **状态**: 正式版
- **所属系统**: Aetherius AI Agent
- **文档类型**: 技术实现指南

## 目录

- [1. Kubernetes 事件概述](#1-kubernetes-事件概述)
- [2. 事件监听架构](#2-事件监听架构)
- [3. 实现方案](#3-实现方案)
- [4. 完整代码示例](#4-完整代码示例)
- [5. 部署配置](#5-部署配置)
- [6. 最佳实践](#6-最佳实践)

## 1. Kubernetes 事件概述

**文档说明**: 本文档描述K8s事件监听的技术实现,适用于:
- 单集群In-Cluster部署(直接集成到Event Gateway)
- Agent代理模式(集成到Agent组件中)

### 1.1 什么是 Kubernetes 事件

Kubernetes 事件 (Events) 是集群中发生的重要状态变化的记录,包括:

- **Pod 生命周期事件**: Created, Started, Failed, Killing, Killed
- **调度事件**: Scheduled, FailedScheduling
- **资源事件**: OOMKilled, FailedMount, Unhealthy
- **控制器事件**: ScalingReplicaSet, SuccessfulCreate
- **节点事件**: NodeNotReady, NodeReady

### 1.2 事件数据结构

```go
// Kubernetes 事件对象结构
type Event struct {
    // 元数据
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata"`

    // 涉及的对象
    InvolvedObject corev1.ObjectReference `json:"involvedObject"`

    // 事件信息
    Reason  string      `json:"reason"`
    Message string      `json:"message"`
    Source  EventSource `json:"source"`
    Type    string      `json:"type"` // Normal, Warning

    // 时间信息
    FirstTimestamp      metav1.Time `json:"firstTimestamp"`
    LastTimestamp       metav1.Time `json:"lastTimestamp"`
    Count               int32       `json:"count"`
    EventTime           metav1.MicroTime `json:"eventTime"`
}
```

### 1.3 为什么需要监听事件

相比 Alertmanager 告警,Kubernetes 事件提供:

- ✅ **更快的响应速度**: 事件几乎实时产生
- ✅ **更细粒度的信息**: 包含详细的上下文
- ✅ **补充告警源**: 与 Prometheus 告警互补
- ✅ **原生 K8s 集成**: 无需额外配置

## 2. 事件监听架构

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────┐
│              Kubernetes API Server                      │
│                                                         │
│  /api/v1/events                                         │
│  /api/v1/watch/events                                   │
│  /api/v1/namespaces/{ns}/events                        │
└────────────────┬────────────────────────────────────────┘
                 │
                 │ Watch API (HTTP Long Polling / WebSocket)
                 │
                 ▼
┌─────────────────────────────────────────────────────────┐
│         Event Watcher Service (Aetherius)               │
│                                                         │
│  ┌───────────────────────────────────────────────┐     │
│  │  Multi-Cluster Event Watcher Manager          │     │
│  │  ┌─────────────┐  ┌─────────────┐            │     │
│  │  │ Cluster A   │  │ Cluster B   │   ...      │     │
│  │  │ Watcher     │  │ Watcher     │            │     │
│  │  └─────────────┘  └─────────────┘            │     │
│  └───────────────────────────────────────────────┘     │
│                       │                                │
│  ┌────────────────────▼──────────────────────────┐     │
│  │       Event Filter & Processor                │     │
│  │  - 事件过滤 (Severity, Type)                   │     │
│  │  - 去重处理 (Fingerprint)                     │     │
│  │  - 事件增强 (Enrichment)                      │     │
│  │  - 格式标准化                                  │     │
│  └────────────────────┬──────────────────────────┘     │
└─────────────────────────┼──────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────┐
│              Message Bus (NATS/Kafka)                   │
│                 event.k8s.received                      │
└────────────────┬────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────┐
│              Orchestrator Service                       │
│           (创建诊断任务并处理)                            │
└─────────────────────────────────────────────────────────┘
```

### 2.2 监听策略

#### 方案对比

| 方案 | 实现方式 | 优点 | 缺点 | 适用场景 |
|------|---------|------|------|---------|
| **Watch API** | client-go Informer | 实时、资源高效 | 需要处理重连 | 生产环境推荐 ⭐⭐⭐⭐⭐ |
| **List + Watch** | 定期List + Watch | 可靠、支持断点续传 | 初始List开销大 | 大规模集群 ⭐⭐⭐⭐ |
| **定期轮询** | 定期List Events | 简单、易实现 | 延迟高、资源浪费 | 开发测试 ⭐⭐ |

**推荐方案**: **Watch API + Informer 机制**

## 3. 实现方案

### 3.1 核心组件设计

```go
// 事件监听服务
type K8sEventWatcher struct {
    // 配置
    config WatcherConfig

    // K8s 客户端
    clientset kubernetes.Interface

    // Informer 工厂
    informerFactory informers.SharedInformerFactory

    // 事件处理器
    eventHandler EventHandler

    // 过滤器
    filters []EventFilter

    // 状态管理
    running atomic.Bool
    stopCh  chan struct{}

    // 指标
    metrics *WatcherMetrics
}

// 配置
type WatcherConfig struct {
    // 集群信息
    ClusterID     string        `yaml:"cluster_id"`
    KubeConfig    string        `yaml:"kubeconfig"`

    // 监听配置
    Namespaces    []string      `yaml:"namespaces"` // 空表示所有namespace
    ResyncPeriod  time.Duration `yaml:"resync_period" default:"30m"`

    // 过滤配置
    EventTypes    []string      `yaml:"event_types"` // Warning, Normal
    ReasonFilters []string      `yaml:"reason_filters"`
    MinSeverity   string        `yaml:"min_severity" default:"Warning"`

    // 性能配置
    WorkerCount   int           `yaml:"worker_count" default:"5"`
    BufferSize    int           `yaml:"buffer_size" default:"1000"`
}

// 事件处理器接口
type EventHandler interface {
    OnAdd(event *corev1.Event) error
    OnUpdate(oldEvent, newEvent *corev1.Event) error
    OnDelete(event *corev1.Event) error
}

// 事件过滤器接口
type EventFilter interface {
    ShouldProcess(event *corev1.Event) bool
}
```

### 3.2 Watch API 实现

```go
package watcher

import (
    "context"
    "fmt"
    "time"

    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/informers"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/cache"
    "k8s.io/client-go/tools/clientcmd"
)

// 创建事件监听器
func NewK8sEventWatcher(config WatcherConfig, handler EventHandler) (*K8sEventWatcher, error) {
    // 1. 创建 K8s 客户端
    kubeConfig, err := clientcmd.BuildConfigFromFlags("", config.KubeConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
    }

    clientset, err := kubernetes.NewForConfig(kubeConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create clientset: %w", err)
    }

    // 2. 创建 Informer 工厂
    var informerFactory informers.SharedInformerFactory
    if len(config.Namespaces) == 0 {
        // 监听所有 namespace
        informerFactory = informers.NewSharedInformerFactory(clientset, config.ResyncPeriod)
    } else {
        // 监听特定 namespace (需要为每个namespace创建informer)
        informerFactory = informers.NewSharedInformerFactory(clientset, config.ResyncPeriod)
    }

    watcher := &K8sEventWatcher{
        config:          config,
        clientset:       clientset,
        informerFactory: informerFactory,
        eventHandler:    handler,
        stopCh:          make(chan struct{}),
        metrics:         NewWatcherMetrics(config.ClusterID),
    }

    // 3. 添加默认过滤器
    watcher.addDefaultFilters()

    return watcher, nil
}

// 启动监听
func (w *K8sEventWatcher) Start(ctx context.Context) error {
    if w.running.Load() {
        return fmt.Errorf("watcher already running")
    }

    log.Info("Starting K8s event watcher",
        zap.String("cluster_id", w.config.ClusterID),
        zap.Strings("namespaces", w.config.Namespaces))

    // 1. 获取 Event Informer
    eventInformer := w.informerFactory.Core().V1().Events().Informer()

    // 2. 注册事件处理函数
    eventInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
        AddFunc: func(obj interface{}) {
            event := obj.(*corev1.Event)
            w.handleEventAdd(event)
        },
        UpdateFunc: func(oldObj, newObj interface{}) {
            oldEvent := oldObj.(*corev1.Event)
            newEvent := newObj.(*corev1.Event)
            w.handleEventUpdate(oldEvent, newEvent)
        },
        DeleteFunc: func(obj interface{}) {
            event := obj.(*corev1.Event)
            w.handleEventDelete(event)
        },
    })

    // 3. 启动 Informer
    w.informerFactory.Start(w.stopCh)

    // 4. 等待缓存同步
    log.Info("Waiting for cache sync...")
    if !cache.WaitForCacheSync(w.stopCh, eventInformer.HasSynced) {
        return fmt.Errorf("failed to sync cache")
    }

    log.Info("Cache synced successfully")
    w.running.Store(true)

    // 5. 等待停止信号
    <-ctx.Done()
    return w.Stop()
}

// 停止监听
func (w *K8sEventWatcher) Stop() error {
    if !w.running.Load() {
        return nil
    }

    log.Info("Stopping K8s event watcher", zap.String("cluster_id", w.config.ClusterID))
    close(w.stopCh)
    w.running.Store(false)

    return nil
}

// 处理事件新增
func (w *K8sEventWatcher) handleEventAdd(event *corev1.Event) {
    w.metrics.EventsReceived.Inc()

    // 1. 过滤检查
    if !w.shouldProcess(event) {
        w.metrics.EventsFiltered.Inc()
        return
    }

    // 2. 调用处理器
    if err := w.eventHandler.OnAdd(event); err != nil {
        log.Error("Failed to handle event add",
            zap.Error(err),
            zap.String("event_name", event.Name),
            zap.String("reason", event.Reason))
        w.metrics.EventsError.Inc()
        return
    }

    w.metrics.EventsProcessed.Inc()
}

// 处理事件更新
func (w *K8sEventWatcher) handleEventUpdate(oldEvent, newEvent *corev1.Event) {
    w.metrics.EventsReceived.Inc()

    // 只处理重要的更新 (Count增加表示事件重复发生)
    if newEvent.Count > oldEvent.Count {
        if !w.shouldProcess(newEvent) {
            w.metrics.EventsFiltered.Inc()
            return
        }

        if err := w.eventHandler.OnUpdate(oldEvent, newEvent); err != nil {
            log.Error("Failed to handle event update", zap.Error(err))
            w.metrics.EventsError.Inc()
            return
        }

        w.metrics.EventsProcessed.Inc()
    }
}

// 处理事件删除
func (w *K8sEventWatcher) handleEventDelete(event *corev1.Event) {
    // 通常不需要处理删除事件 (事件会自动过期)
    log.Debug("Event deleted",
        zap.String("event_name", event.Name),
        zap.String("reason", event.Reason))
}

// 过滤检查
func (w *K8sEventWatcher) shouldProcess(event *corev1.Event) bool {
    for _, filter := range w.filters {
        if !filter.ShouldProcess(event) {
            return false
        }
    }
    return true
}

// 添加默认过滤器
func (w *K8sEventWatcher) addDefaultFilters() {
    // 1. 事件类型过滤器 (只处理 Warning)
    if len(w.config.EventTypes) > 0 {
        w.filters = append(w.filters, &EventTypeFilter{
            AllowedTypes: w.config.EventTypes,
        })
    }

    // 2. Reason 过滤器
    if len(w.config.ReasonFilters) > 0 {
        w.filters = append(w.filters, &ReasonFilter{
            AllowedReasons: w.config.ReasonFilters,
        })
    }

    // 3. 时间过滤器 (忽略太旧的事件)
    w.filters = append(w.filters, &TimeFilter{
        MaxAge: 5 * time.Minute,
    })
}
```

### 3.3 事件过滤器实现

```go
// 事件类型过滤器
type EventTypeFilter struct {
    AllowedTypes []string // "Warning", "Normal"
}

func (f *EventTypeFilter) ShouldProcess(event *corev1.Event) bool {
    if len(f.AllowedTypes) == 0 {
        return true
    }

    for _, t := range f.AllowedTypes {
        if event.Type == t {
            return true
        }
    }
    return false
}

// Reason 过滤器
type ReasonFilter struct {
    AllowedReasons []string
}

func (f *ReasonFilter) ShouldProcess(event *corev1.Event) bool {
    if len(f.AllowedReasons) == 0 {
        return true
    }

    for _, reason := range f.AllowedReasons {
        if event.Reason == reason {
            return true
        }
    }
    return false
}

// 时间过滤器 (忽略太旧的事件)
type TimeFilter struct {
    MaxAge time.Duration
}

func (f *TimeFilter) ShouldProcess(event *corev1.Event) bool {
    // 使用 LastTimestamp 作为最后发生时间
    if event.LastTimestamp.IsZero() {
        return false
    }

    age := time.Since(event.LastTimestamp.Time)
    return age <= f.MaxAge
}

// 命名空间过滤器
type NamespaceFilter struct {
    AllowedNamespaces []string
    DeniedNamespaces  []string
}

func (f *NamespaceFilter) ShouldProcess(event *corev1.Event) bool {
    ns := event.InvolvedObject.Namespace

    // 黑名单检查
    for _, denied := range f.DeniedNamespaces {
        if ns == denied {
            return false
        }
    }

    // 白名单检查
    if len(f.AllowedNamespaces) > 0 {
        for _, allowed := range f.AllowedNamespaces {
            if ns == allowed {
                return true
            }
        }
        return false
    }

    return true
}

// 严重程度过滤器 (基于Reason映射)
type SeverityFilter struct {
    MinSeverity Severity
}

type Severity int

const (
    SeverityLow Severity = iota
    SeverityMedium
    SeverityHigh
    SeverityCritical
)

var reasonSeverityMap = map[string]Severity{
    "OOMKilled":          SeverityCritical,
    "CrashLoopBackOff":   SeverityCritical,
    "ImagePullBackOff":   SeverityHigh,
    "ErrImagePull":       SeverityHigh,
    "FailedScheduling":   SeverityHigh,
    "Unhealthy":          SeverityMedium,
    "BackOff":            SeverityMedium,
    "FailedMount":        SeverityMedium,
}

func (f *SeverityFilter) ShouldProcess(event *corev1.Event) bool {
    severity, exists := reasonSeverityMap[event.Reason]
    if !exists {
        severity = SeverityLow
    }

    return severity >= f.MinSeverity
}
```

### 3.4 事件处理器实现

```go
// 事件处理器实现
type AetheriusEventHandler struct {
    clusterID    string
    messageBus   MessageBus
    deduplicator *EventDeduplicator
}

func NewAetheriusEventHandler(clusterID string, messageBus MessageBus) *AetheriusEventHandler {
    return &AetheriusEventHandler{
        clusterID:    clusterID,
        messageBus:   messageBus,
        deduplicator: NewEventDeduplicator(5 * time.Minute),
    }
}

// 处理新增事件
func (h *AetheriusEventHandler) OnAdd(event *corev1.Event) error {
    // 1. 去重检查
    fingerprint := h.calculateFingerprint(event)
    if h.deduplicator.IsDuplicate(fingerprint) {
        log.Debug("Duplicate event, skipping",
            zap.String("fingerprint", fingerprint),
            zap.String("reason", event.Reason))
        return nil
    }

    // 2. 转换为标准事件格式
    standardEvent := h.convertToStandardEvent(event)

    // 3. 发布到消息总线
    if err := h.messageBus.Publish("event.k8s.received", standardEvent); err != nil {
        return fmt.Errorf("failed to publish event: %w", err)
    }

    log.Info("K8s event processed",
        zap.String("cluster_id", h.clusterID),
        zap.String("namespace", event.Namespace),
        zap.String("reason", event.Reason),
        zap.String("involved_object", fmt.Sprintf("%s/%s",
            event.InvolvedObject.Kind, event.InvolvedObject.Name)))

    return nil
}

// 处理更新事件
func (h *AetheriusEventHandler) OnUpdate(oldEvent, newEvent *corev1.Event) error {
    // 事件Count增加表示重复发生,增加严重性
    if newEvent.Count > oldEvent.Count {
        log.Warn("Event repeated",
            zap.String("reason", newEvent.Reason),
            zap.Int32("old_count", oldEvent.Count),
            zap.Int32("new_count", newEvent.Count))

        // 如果重复次数超过阈值,提高优先级
        if newEvent.Count >= 5 {
            standardEvent := h.convertToStandardEvent(newEvent)
            standardEvent.Severity = "high" // 提升严重程度
            standardEvent.Labels["repeated_count"] = fmt.Sprintf("%d", newEvent.Count)

            return h.messageBus.Publish("event.k8s.received", standardEvent)
        }
    }

    return nil
}

// 处理删除事件
func (h *AetheriusEventHandler) OnDelete(event *corev1.Event) error {
    // 通常不需要处理
    return nil
}

// 计算事件指纹 (用于去重)
func (h *AetheriusEventHandler) calculateFingerprint(event *corev1.Event) string {
    return fmt.Sprintf("%s/%s/%s/%s/%s",
        h.clusterID,
        event.Namespace,
        event.InvolvedObject.Kind,
        event.InvolvedObject.Name,
        event.Reason)
}

// 转换为标准事件格式
func (h *AetheriusEventHandler) convertToStandardEvent(event *corev1.Event) *StandardEvent {
    return &StandardEvent{
        ID:        uuid.New().String(),
        Type:      "k8s_event",
        Source:    "kubernetes",
        Timestamp: time.Now(),
        ClusterID: h.clusterID,
        Namespace: event.Namespace,
        Severity:  h.mapSeverity(event),
        Labels: map[string]string{
            "event_type":       event.Type,
            "reason":           event.Reason,
            "involved_kind":    event.InvolvedObject.Kind,
            "involved_name":    event.InvolvedObject.Name,
            "count":            fmt.Sprintf("%d", event.Count),
        },
        Annotations: map[string]string{
            "message":          event.Message,
            "first_timestamp":  event.FirstTimestamp.Format(time.RFC3339),
            "last_timestamp":   event.LastTimestamp.Format(time.RFC3339),
        },
        Fingerprint: h.calculateFingerprint(event),
        RawData: map[string]interface{}{
            "k8s_event": event,
        },
    }
}

// 映射严重程度
func (h *AetheriusEventHandler) mapSeverity(event *corev1.Event) string {
    if event.Type == "Warning" {
        // 根据Reason判断严重程度
        switch event.Reason {
        case "OOMKilled", "CrashLoopBackOff":
            return "critical"
        case "ImagePullBackOff", "FailedScheduling":
            return "high"
        case "Unhealthy", "BackOff":
            return "medium"
        default:
            return "low"
        }
    }
    return "info"
}
```

### 3.5 事件去重器

```go
// 事件去重器
type EventDeduplicator struct {
    cache *cache.Cache
    ttl   time.Duration
    mu    sync.RWMutex
}

func NewEventDeduplicator(ttl time.Duration) *EventDeduplicator {
    c := cache.New(ttl, ttl*2)
    return &EventDeduplicator{
        cache: c,
        ttl:   ttl,
    }
}

// 检查是否重复
func (d *EventDeduplicator) IsDuplicate(fingerprint string) bool {
    d.mu.RLock()
    defer d.mu.RUnlock()

    _, found := d.cache.Get(fingerprint)
    if found {
        return true
    }

    // 标记为已处理
    d.cache.Set(fingerprint, true, d.ttl)
    return false
}

// 清理缓存
func (d *EventDeduplicator) Cleanup() {
    d.cache.Flush()
}
```

## 4. 完整代码示例

### 4.1 主程序入口

```go
package main

import (
    "context"
    "flag"
    "os"
    "os/signal"
    "syscall"

    "github.com/aetherius/k8s-agent/pkg/watcher"
    "go.uber.org/zap"
)

func main() {
    // 解析命令行参数
    var (
        kubeconfig = flag.String("kubeconfig", "", "Path to kubeconfig file")
        clusterID  = flag.String("cluster-id", "", "Cluster ID")
        configFile = flag.String("config", "config.yaml", "Config file path")
    )
    flag.Parse()

    // 初始化日志
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    zap.ReplaceGlobals(logger)

    // 加载配置
    config, err := loadConfig(*configFile)
    if err != nil {
        log.Fatal("Failed to load config", zap.Error(err))
    }

    if *kubeconfig != "" {
        config.KubeConfig = *kubeconfig
    }
    if *clusterID != "" {
        config.ClusterID = *clusterID
    }

    // 创建消息总线客户端
    messageBus, err := newMessageBus(config.MessageBusURL)
    if err != nil {
        log.Fatal("Failed to create message bus", zap.Error(err))
    }
    defer messageBus.Close()

    // 创建事件处理器
    eventHandler := watcher.NewAetheriusEventHandler(config.ClusterID, messageBus)

    // 创建事件监听器
    eventWatcher, err := watcher.NewK8sEventWatcher(config, eventHandler)
    if err != nil {
        log.Fatal("Failed to create event watcher", zap.Error(err))
    }

    // 启动监听
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go func() {
        if err := eventWatcher.Start(ctx); err != nil {
            log.Error("Event watcher error", zap.Error(err))
        }
    }()

    log.Info("K8s Event Watcher started",
        zap.String("cluster_id", config.ClusterID),
        zap.Strings("namespaces", config.Namespaces))

    // 等待退出信号
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh

    log.Info("Shutting down...")
    cancel()
}
```

### 4.2 多集群监听

```go
// 多集群事件监听管理器
type MultiClusterWatcherManager struct {
    watchers map[string]*K8sEventWatcher
    mu       sync.RWMutex
}

func NewMultiClusterWatcherManager() *MultiClusterWatcherManager {
    return &MultiClusterWatcherManager{
        watchers: make(map[string]*K8sEventWatcher),
    }
}

// 添加集群监听
func (m *MultiClusterWatcherManager) AddCluster(ctx context.Context, config WatcherConfig) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    if _, exists := m.watchers[config.ClusterID]; exists {
        return fmt.Errorf("cluster %s already exists", config.ClusterID)
    }

    // 创建事件处理器
    handler := NewAetheriusEventHandler(config.ClusterID, m.messageBus)

    // 创建监听器
    watcher, err := NewK8sEventWatcher(config, handler)
    if err != nil {
        return fmt.Errorf("failed to create watcher for cluster %s: %w", config.ClusterID, err)
    }

    // 启动监听
    go func() {
        if err := watcher.Start(ctx); err != nil {
            log.Error("Watcher failed",
                zap.String("cluster_id", config.ClusterID),
                zap.Error(err))
        }
    }()

    m.watchers[config.ClusterID] = watcher
    log.Info("Cluster watcher added", zap.String("cluster_id", config.ClusterID))

    return nil
}

// 移除集群监听
func (m *MultiClusterWatcherManager) RemoveCluster(clusterID string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    watcher, exists := m.watchers[clusterID]
    if !exists {
        return fmt.Errorf("cluster %s not found", clusterID)
    }

    if err := watcher.Stop(); err != nil {
        return fmt.Errorf("failed to stop watcher: %w", err)
    }

    delete(m.watchers, clusterID)
    log.Info("Cluster watcher removed", zap.String("cluster_id", clusterID))

    return nil
}

// 获取所有集群状态
func (m *MultiClusterWatcherManager) GetStatus() map[string]WatcherStatus {
    m.mu.RLock()
    defer m.mu.RUnlock()

    status := make(map[string]WatcherStatus)
    for clusterID, watcher := range m.watchers {
        status[clusterID] = WatcherStatus{
            ClusterID: clusterID,
            Running:   watcher.running.Load(),
            Metrics:   watcher.metrics.Snapshot(),
        }
    }

    return status
}
```

## 5. 部署配置

### 5.1 配置文件

```yaml
# watcher-config.yaml
cluster_id: "prod-us-west-2"
kubeconfig: "/path/to/kubeconfig"

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
  - "FailedScheduling"
  - "Unhealthy"
  - "BackOff"
min_severity: "Warning"

# 性能配置
worker_count: 5
buffer_size: 1000

# 消息总线
message_bus_url: "nats://localhost:4222"
```

### 5.2 Kubernetes 部署

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8s-event-watcher
  namespace: aetherius
spec:
  replicas: 1  # 每个集群一个实例
  selector:
    matchLabels:
      app: k8s-event-watcher
  template:
    metadata:
      labels:
        app: k8s-event-watcher
    spec:
      serviceAccountName: aetherius-event-watcher
      containers:
      - name: watcher
        image: aetherius/k8s-event-watcher:v1.0
        args:
        - --cluster-id=$(CLUSTER_ID)
        - --config=/etc/aetherius/watcher-config.yaml
        env:
        - name: CLUSTER_ID
          value: "prod-us-west-2"
        volumeMounts:
        - name: config
          mountPath: /etc/aetherius
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "250m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: watcher-config
---
# RBAC 配置
apiVersion: v1
kind: ServiceAccount
metadata:
  name: aetherius-event-watcher
  namespace: aetherius
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: aetherius-event-watcher
rules:
- apiGroups: [""]
  resources: ["events"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: aetherius-event-watcher
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: aetherius-event-watcher
subjects:
- kind: ServiceAccount
  name: aetherius-event-watcher
  namespace: aetherius
```

### 5.3 监控配置

```yaml
# ServiceMonitor for Prometheus
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: k8s-event-watcher
  namespace: aetherius
spec:
  selector:
    matchLabels:
      app: k8s-event-watcher
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
```

## 6. 最佳实践

### 6.1 性能优化

#### 1. 使用 Informer 机制

```go
// ✅ 推荐: 使用 SharedInformerFactory
informerFactory := informers.NewSharedInformerFactory(clientset, resyncPeriod)
eventInformer := informerFactory.Core().V1().Events().Informer()

// ❌ 不推荐: 直接 Watch API (资源浪费)
watch, err := clientset.CoreV1().Events("").Watch(ctx, metav1.ListOptions{})
```

#### 2. 合理设置 ResyncPeriod

```go
// 生产环境推荐: 30分钟
resyncPeriod := 30 * time.Minute

// 开发环境可以更短
resyncPeriod := 5 * time.Minute
```

#### 3. 实施事件去重

```go
// 使用 TTL 缓存去重
deduplicator := NewEventDeduplicator(5 * time.Minute)
if deduplicator.IsDuplicate(fingerprint) {
    return // 跳过重复事件
}
```

### 6.2 可靠性保障

#### 1. 优雅重启

```go
func (w *K8sEventWatcher) Stop() error {
    // 1. 停止接收新事件
    close(w.stopCh)

    // 2. 等待处理中的事件完成
    w.wg.Wait()

    // 3. 清理资源
    w.cleanup()

    return nil
}
```

#### 2. 断线重连

```go
// Informer 机制自动处理重连
// 只需确保正确处理错误
eventInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
    AddFunc: func(obj interface{}) {
        if err := handler.OnAdd(obj); err != nil {
            log.Error("Handler failed", zap.Error(err))
            // 记录到死信队列,稍后重试
            deadLetterQueue.Add(obj)
        }
    },
})
```

#### 3. 状态持久化

```go
// 保存最后处理的ResourceVersion
func (w *K8sEventWatcher) saveCheckpoint(resourceVersion string) {
    checkpoint := Checkpoint{
        ClusterID:       w.config.ClusterID,
        ResourceVersion: resourceVersion,
        Timestamp:       time.Now(),
    }
    w.storage.Save(checkpoint)
}

// 从断点恢复
func (w *K8sEventWatcher) resumeFromCheckpoint() {
    checkpoint := w.storage.Load(w.config.ClusterID)
    if checkpoint != nil {
        listOptions.ResourceVersion = checkpoint.ResourceVersion
    }
}
```

### 6.3 监控告警

```go
// Prometheus 指标
type WatcherMetrics struct {
    EventsReceived  prometheus.Counter
    EventsProcessed prometheus.Counter
    EventsFiltered  prometheus.Counter
    EventsError     prometheus.Counter
    ProcessingTime  prometheus.Histogram
    CacheSize       prometheus.Gauge
}

func NewWatcherMetrics(clusterID string) *WatcherMetrics {
    return &WatcherMetrics{
        EventsReceived: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "k8s_events_received_total",
            Help: "Total number of K8s events received",
            ConstLabels: prometheus.Labels{"cluster_id": clusterID},
        }),
        // ... 其他指标
    }
}
```

### 6.4 安全考虑

#### 1. 最小权限原则

```yaml
# 只授予必要的权限
rules:
- apiGroups: [""]
  resources: ["events"]
  verbs: ["get", "list", "watch"]  # 只读权限
```

#### 2. 敏感信息过滤

```go
// 过滤敏感信息
func (h *AetheriusEventHandler) sanitizeEvent(event *corev1.Event) {
    // 移除可能包含敏感信息的字段
    event.Message = sanitizeMessage(event.Message)

    // 不要记录 Secret 相关的事件
    if event.InvolvedObject.Kind == "Secret" {
        return
    }
}
```

## 附录

### A. 常见事件 Reason 列表

| Reason | 类型 | 严重程度 | 说明 |
|--------|------|----------|------|
| **OOMKilled** | Warning | Critical | Pod 内存溢出被杀死 |
| **CrashLoopBackOff** | Warning | Critical | 容器持续崩溃重启 |
| **ImagePullBackOff** | Warning | High | 镜像拉取失败 |
| **FailedScheduling** | Warning | High | Pod 调度失败 |
| **Unhealthy** | Warning | Medium | 健康检查失败 |
| **BackOff** | Warning | Medium | 容器启动失败回退 |
| **FailedMount** | Warning | Medium | 卷挂载失败 |
| **Killing** | Normal | Info | 正在终止 Pod |
| **Created** | Normal | Info | 创建成功 |
| **Started** | Normal | Info | 启动成功 |

### B. 相关文档

- [client-go 官方文档](https://github.com/kubernetes/client-go)
- [Informer 机制详解](https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes)
- [微服务架构文档](./06_microservices.md)
- [Event Gateway 设计](./06_microservices.md#21-event-gateway-service-事件网关服务)

### C. 故障排查

#### 问题 1: Informer 缓存不同步

```bash
# 检查日志
kubectl logs -f deployment/k8s-event-watcher -n aetherius | grep "cache sync"

# 解决方案: 增加超时时间或检查网络连接
```

#### 问题 2: 事件丢失

```bash
# 检查过滤器配置
# 检查事件处理错误率
kubectl exec -it deployment/k8s-event-watcher -n aetherius -- \
  curl localhost:9090/metrics | grep k8s_events_error
```

#### 问题 3: 内存占用过高

```bash
# 检查缓存大小
# 调整 ResyncPeriod 和去重TTL
# 增加资源限制
```