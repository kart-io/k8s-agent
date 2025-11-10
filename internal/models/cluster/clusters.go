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
