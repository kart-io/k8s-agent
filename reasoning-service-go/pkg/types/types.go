package types

import "time"

// RootCauseType represents the type of root cause
type RootCauseType string

const (
	// Pod & Container Level
	OOMKiller              RootCauseType = "OOMKiller"              // #25 ContainerOOMKilled
	CPUThrottling          RootCauseType = "CPUThrottling"          // #26 ContainerCPUThrottlingHigh
	ReadinessProbeFailure  RootCauseType = "ReadinessProbeFailure"  // #22 KubePodNotReady
	LivenessProbeFailure   RootCauseType = "LivenessProbeFailure"   // #21 KubePodCrashLooping (partial)
	ImagePullError         RootCauseType = "ImagePullError"         // #24 ImagePullBackOff
	ConfigError            RootCauseType = "ConfigError"            // #21 CrashLoopBackOff (config issues)
	PodStuckPending        RootCauseType = "PodStuckPending"        // #23 KubePodStuckInPending

	// Node Level
	NodeNotReady           RootCauseType = "NodeNotReady"           // #1 KubeNodeNotReady
	KubeletDown            RootCauseType = "KubeletDown"            // #2 KubeletDown
	NodeHighCPU            RootCauseType = "NodeHighCPU"            // #3 NodeHighCpuUsage
	NodeHighMemory         RootCauseType = "NodeHighMemory"         // #4 NodeHighMemoryUsage
	NodeDiskSpaceLow       RootCauseType = "NodeDiskSpaceLow"       // #5 NodeFilesystemAlmostOutOfSpace
	DiskPressure           RootCauseType = "DiskPressure"           // #6 NodeDiskPressure
	MemoryPressure         RootCauseType = "MemoryPressure"         // #7 NodeMemoryPressure
	PIDPressure            RootCauseType = "PIDPressure"            // #8 NodePIDPressure
	NetworkUnavailable     RootCauseType = "NetworkUnavailable"     // #9 NodeNetworkUnavailable
	CertificateExpiring    RootCauseType = "CertificateExpiring"    // #10-11 Certificate expiration

	// Control Plane Level
	APIServerDown          RootCauseType = "APIServerDown"          // #12 KubeAPIDown
	APIServerHighLatency   RootCauseType = "APIServerHighLatency"   // #13 KubeAPILatencyHigh
	APIServerHighError     RootCauseType = "APIServerHighError"     // #14 KubeAPIErrorRateHigh
	ControllerManagerDown  RootCauseType = "ControllerManagerDown"  // #15 KubeControllerManagerDown
	SchedulerDown          RootCauseType = "SchedulerDown"          // #16 KubeSchedulerDown
	EtcdInsufficientPeers  RootCauseType = "EtcdInsufficientPeers"  // #17 EtcdInsufficientMembers
	EtcdHighLatency        RootCauseType = "EtcdHighLatency"        // #18 EtcdHighFsyncLatency
	EtcdLeaderFlapping     RootCauseType = "EtcdLeaderFlapping"     // #19 EtcdLeaderChanges
	CoreDNSDown            RootCauseType = "CoreDNSDown"            // #20 CoreDNSDown

	// Workload Level
	DeploymentReplicasMismatch RootCauseType = "DeploymentReplicasMismatch" // #27 KubeDeploymentReplicasMismatch
	DeploymentStuck            RootCauseType = "DeploymentStuck"            // #28 KubeDeploymentStuck
	StatefulSetReplicasMismatch RootCauseType = "StatefulSetReplicasMismatch" // #29 KubeStatefulSetReplicasMismatch
	StatefulSetNotReady        RootCauseType = "StatefulSetNotReady"        // #30 KubeStatefulSetNotReady
	DaemonSetRolloutStuck      RootCauseType = "DaemonSetRolloutStuck"      // #31 KubeDaemonSetRolloutStuck
	JobFailed                  RootCauseType = "JobFailed"                  // #32 KubeJobFailed
	CronJobNotCompleting       RootCauseType = "CronJobNotCompleting"       // #33 CronJobNotCompleting
	HPAMaxedOut                RootCauseType = "HPAMaxedOut"                // #34 KubeHpaMaxedOut
	PDBAtRisk                  RootCauseType = "PDBAtRisk"                  // #35 PodDisruptionBudgetAtRisk

	// Storage & Network Level
	PVFillingUp            RootCauseType = "PVFillingUp"            // #36 KubePersistentVolumeFillingUp
	PVCPending             RootCauseType = "PVCPending"             // #37 PersistentVolumeClaimPending
	PVFailed               RootCauseType = "PVFailed"               // #38 PersistentVolumeFailed
	VolumeError            RootCauseType = "VolumeError"            // General volume issues
	IngressControllerDown  RootCauseType = "IngressControllerDown"  // #39 IngressControllerDown
	IngressCertExpiring    RootCauseType = "IngressCertExpiring"    // #40 IngressCertificateExpiration
	NetworkPolicyDenials   RootCauseType = "NetworkPolicyDenials"   // #41 HighNetworkPolicyDenials
	NetworkError           RootCauseType = "NetworkError"           // General network issues

	// Quotas & Monitoring
	ResourceQuotaExceeded  RootCauseType = "ResourceQuotaExceeded"  // #42 ResourceQuotaExceeded
	ResourceLimit          RootCauseType = "ResourceLimit"          // General resource limits
	PrometheusTargetDown   RootCauseType = "PrometheusTargetDown"   // #43 PrometheusTargetSyncFailure
	PrometheusRuleFailure  RootCauseType = "PrometheusRuleFailure"  // #44 PrometheusRuleFailures
	AlertmanagerConfigBad  RootCauseType = "AlertmanagerConfigBad"  // #45 AlertmanagerConfigInconsistent

	Unknown                RootCauseType = "Unknown"
)

// AnalysisRequest represents the request for root cause analysis
type AnalysisRequest struct {
	RequestID    string                 `json:"request_id"`
	AnalysisType string                 `json:"analysis_type"` // "root_cause", "predict", etc.
	Context      AnalysisContext        `json:"context"`
	Options      AnalysisOptions        `json:"options,omitempty"`
}

// AnalysisContext contains the data for analysis
type AnalysisContext struct {
	Event       map[string]interface{} `json:"event,omitempty"`
	Logs        string                 `json:"logs,omitempty"`
	Metrics     MetricsData            `json:"metrics,omitempty"`
	ClusterID   string                 `json:"cluster_id,omitempty"`
	Namespace   string                 `json:"namespace,omitempty"`
	ResourceName string                `json:"resource_name,omitempty"`
}

// MetricsData contains resource metrics
type MetricsData struct {
	Memory  *MemoryMetrics  `json:"memory,omitempty"`
	CPU     *CPUMetrics     `json:"cpu,omitempty"`
	Disk    *DiskMetrics    `json:"disk,omitempty"`
	Network *NetworkMetrics `json:"network,omitempty"`
	History []MetricSnapshot `json:"history,omitempty"`
}

// MemoryMetrics represents memory usage metrics
type MemoryMetrics struct {
	UsagePercent float64 `json:"usage_percent"`
	UsageBytes   int64   `json:"usage_bytes,omitempty"`
	LimitBytes   int64   `json:"limit_bytes,omitempty"`
}

// CPUMetrics represents CPU usage metrics
type CPUMetrics struct {
	UsagePercent      float64 `json:"usage_percent"`
	ThrottlingPercent float64 `json:"throttling_percent,omitempty"`
	LimitCores        float64 `json:"limit_cores,omitempty"`
}

// DiskMetrics represents disk usage metrics
type DiskMetrics struct {
	UsagePercent float64 `json:"usage_percent"`
	UsageBytes   int64   `json:"usage_bytes,omitempty"`
	TotalBytes   int64   `json:"total_bytes,omitempty"`
}

// NetworkMetrics represents network metrics
type NetworkMetrics struct {
	ErrorRate      float64 `json:"error_rate,omitempty"`
	Latency        float64 `json:"latency,omitempty"`
	PacketLoss     float64 `json:"packet_loss,omitempty"`
}

// MetricSnapshot represents a point-in-time metric snapshot
type MetricSnapshot struct {
	Timestamp    time.Time       `json:"timestamp"`
	Memory       *MemoryMetrics  `json:"memory,omitempty"`
	CPU          *CPUMetrics     `json:"cpu,omitempty"`
	Disk         *DiskMetrics    `json:"disk,omitempty"`
	Network      *NetworkMetrics `json:"network,omitempty"`
	RestartCount int             `json:"restart_count,omitempty"`
}

// AnalysisOptions contains options for the analysis
type AnalysisOptions struct {
	MinConfidence        float64 `json:"min_confidence,omitempty"`
	IncludeSimilarCases  bool    `json:"include_similar_cases,omitempty"`
	MaxRecommendations   int     `json:"max_recommendations,omitempty"`
	UseLLM               bool    `json:"use_llm,omitempty"`
	LLMProvider          string  `json:"llm_provider,omitempty"` // "openai", "gemini", "deepseek"
}

// RootCause represents the identified root cause
type RootCause struct {
	Type        RootCauseType `json:"type"`
	Description string        `json:"description"`
	Confidence  float64       `json:"confidence"`
	Evidence    []string      `json:"evidence"`
}

// Recommendation represents a recommended action
type Recommendation struct {
	Action            string   `json:"action"`
	Description       string   `json:"description"`
	Command           string   `json:"command,omitempty"`            // kubectl command to execute
	YAML              string   `json:"yaml,omitempty"`               // YAML configuration example
	Confidence        float64  `json:"confidence"`
	Risk              string   `json:"risk"` // "low", "medium", "high"
	Impact            string   `json:"impact"`
	Steps             []string `json:"steps"`
	RollbackSteps     []string `json:"rollback_steps,omitempty"`
	EstimatedDuration string   `json:"estimated_duration,omitempty"`
}

// AnalysisResult represents the result of the analysis
type AnalysisResult struct {
	RequestID       string           `json:"request_id"`
	Status          string           `json:"status"` // "completed", "failed", "in_progress"
	Result          *DetailedResult  `json:"result,omitempty"`
	Error           string           `json:"error,omitempty"`
	ProcessingTime  float64          `json:"processing_time"`
}

// DetailedResult contains the detailed analysis result
type DetailedResult struct {
	RootCause        *RootCause       `json:"root_cause"`
	Recommendations  []Recommendation `json:"recommendations"`
	Confidence       float64          `json:"confidence"`
	Evidence         []string         `json:"evidence"`
	SimilarCases     []CaseStudy      `json:"similar_cases,omitempty"`
	LLMAnalysis      string           `json:"llm_analysis,omitempty"`
}

// CaseStudy represents a historical case
type CaseStudy struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Symptoms    []string      `json:"symptoms"`
	RootCause   RootCauseType `json:"root_cause"`
	Solution    string        `json:"solution"`
	Outcome     string        `json:"outcome,omitempty"`
	ClusterID   string        `json:"cluster_id,omitempty"`
	Similarity  float64       `json:"similarity,omitempty"`
	CreatedAt   time.Time     `json:"created_at,omitempty"`
}

// PredictionRequest represents a failure prediction request
type PredictionRequest struct {
	ClusterID    string      `json:"cluster_id"`
	ResourceType string      `json:"resource_type"` // "pod", "node", etc.
	ResourceName string      `json:"resource_name"`
	Metrics      MetricsData `json:"metrics"`
	TimeWindow   string      `json:"time_window,omitempty"` // "1h", "24h", "7d"
}

// PredictionResult represents the prediction result
type PredictionResult struct {
	FailureProbability   float64         `json:"failure_probability"`
	PredictedFailureTime *time.Time      `json:"predicted_failure_time,omitempty"`
	FailureTypes         []RootCauseType `json:"failure_types"`
	Confidence           float64         `json:"confidence"`
	ContributingFactors  []string        `json:"contributing_factors"`
}

// FeedbackRequest represents user feedback
type FeedbackRequest struct {
	FeedbackID       string        `json:"feedback_id"`
	RequestID        string        `json:"request_id"`
	FeedbackType     string        `json:"feedback_type"` // "diagnosis_accuracy", "recommendation_helpful", etc.
	Rating           int           `json:"rating"`        // 1-5
	WasHelpful       bool          `json:"was_helpful"`
	ActualRootCause  RootCauseType `json:"actual_root_cause,omitempty"`
	Comments         string        `json:"comments,omitempty"`
	SubmittedBy      string        `json:"submitted_by,omitempty"`
	SubmittedAt      time.Time     `json:"submitted_at,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status     string            `json:"status"`
	Service    string            `json:"service"`
	Components map[string]bool   `json:"components"`
	Timestamp  time.Time         `json:"timestamp"`
}
