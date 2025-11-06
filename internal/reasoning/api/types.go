package api

import "time"

// K8sEventRequest represents a simplified request for K8s event analysis.
type K8sEventRequest struct {
	ClusterID string                 `json:"cluster_id,omitempty"`
	Event     map[string]interface{} `json:"event"`
	UseLLM    bool                   `json:"use_llm,omitempty"`
}

// K8sEventAnalysisResponse represents the response for K8s event analysis.
type K8sEventAnalysisResponse struct {
	Analysis        string   `json:"analysis"`
	RootCause       string   `json:"rootCause"` // 使用驼峰命名与前端一致
	Confidence      float64  `json:"confidence"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// HealthResponse represents health check response (shared with types package).
type HealthResponse struct {
	Status     string          `json:"status"`
	Service    string          `json:"service"`
	Components map[string]bool `json:"components"`
	Timestamp  time.Time       `json:"timestamp"`
}
