package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/kart-io/k8s-agent/internal/reasoning/config"
	"github.com/kart-io/k8s-agent/internal/reasoning/llm"
	"github.com/kart-io/k8s-agent/internal/reasoning/types"
)

// RootCauseAnalyzer analyzes root causes using rules and LLM.
type RootCauseAnalyzer struct {
	config     *config.Config
	llmClients []llm.Client
	patterns   map[types.RootCauseType][]*regexp.Regexp
}

// NewRootCauseAnalyzer creates a new root cause analyzer.
func NewRootCauseAnalyzer(cfg *config.Config, llmClients []llm.Client) *RootCauseAnalyzer {
	analyzer := &RootCauseAnalyzer{
		config:     cfg,
		llmClients: llmClients,
		patterns:   make(map[types.RootCauseType][]*regexp.Regexp),
	}
	analyzer.loadPatterns()
	return analyzer
}

// loadPatterns loads regex patterns for log analysis.
func (a *RootCauseAnalyzer) loadPatterns() {
	patternDefs := map[types.RootCauseType][]string{
		types.OOMKiller: {
			`out of memory`,
			`OOM killer`,
			`fatal error: runtime: out of memory`,
			`java\.lang\.OutOfMemoryError`,
			`Cannot allocate memory`,
		},
		types.CPUThrottling: {
			`cpu throttling`,
			`CPU.*throttl`,
			`container.*throttled`,
		},
		types.NetworkError: {
			`connection refused`,
			`dial tcp.*timeout`,
			`no route to host`,
			`network.*unreachable`,
			`connection.*reset`,
			`dns.*resolution.*failed`,
		},
		types.ConfigError: {
			`invalid configuration`,
			`config.*error`,
			`misconfigured`,
			`configuration.*failed`,
			`invalid.*env.*var`,
		},
		types.ImagePullError: {
			`failed to pull image`,
			`image pull.*failed`,
			`manifest.*not found`,
			`unauthorized.*image`,
		},
		types.DiskPressure: {
			`disk.*full`,
			`no space left`,
			`disk pressure`,
			`volume.*full`,
		},
		types.VolumeError: {
			`failed to mount`,
			`mount.*failed`,
			`volume.*not found`,
			`pvc.*not bound`,
		},
	}

	for rootCause, patterns := range patternDefs {
		for _, pattern := range patterns {
			re, err := regexp.Compile(`(?i)` + pattern) // Case-insensitive
			if err != nil {
				continue
			}
			a.patterns[rootCause] = append(a.patterns[rootCause], re)
		}
	}
}

// Analyze analyzes the context and determines the root cause.
func (a *RootCauseAnalyzer) Analyze(ctx context.Context, req *types.AnalysisRequest) (*types.AnalysisResult, error) {
	result := &types.AnalysisResult{
		RequestID: req.RequestID,
		Status:    "completed",
	}

	// Run multiple analysis methods
	analyses := []analysisResult{}

	// 1. Event-based analysis
	if req.Context.Event != nil {
		if eventAnalysis := a.analyzeEvent(req.Context.Event); eventAnalysis != nil {
			analyses = append(analyses, *eventAnalysis)
		}
	}

	// 2. Log-based analysis
	if req.Context.Logs != "" {
		if logAnalysis := a.analyzeLogs(req.Context.Logs); logAnalysis != nil {
			analyses = append(analyses, *logAnalysis)
		}
	}

	// 3. Metrics-based analysis
	if req.Context.Metrics.Memory != nil || req.Context.Metrics.CPU != nil {
		if metricsAnalysis := a.analyzeMetrics(&req.Context.Metrics); metricsAnalysis != nil {
			analyses = append(analyses, *metricsAnalysis)
		}
	}

	// Select best analysis
	var bestAnalysis *analysisResult
	if len(analyses) > 0 {
		bestAnalysis = &analyses[0]
		for i := range analyses {
			if analyses[i].Confidence > bestAnalysis.Confidence {
				bestAnalysis = &analyses[i]
			}
		}
	}

	// 4. LLM-based analysis (if enabled and needed)
	useLLM := req.Options.UseLLM || (a.config.Analysis.UseLLMFallback && (bestAnalysis == nil || bestAnalysis.Confidence < a.config.Analysis.MinConfidence))

	if useLLM && len(a.llmClients) > 0 {
		llmAnalysis, err := a.analyzeLLM(ctx, req)
		if err == nil && llmAnalysis != nil {
			// If LLM has higher confidence, use it
			if bestAnalysis == nil || llmAnalysis.Confidence > bestAnalysis.Confidence {
				bestAnalysis = llmAnalysis
			}
		}
	}

	// Build result
	if bestAnalysis == nil {
		result.Result = &types.DetailedResult{
			Confidence: 0.0,
			Evidence:   []string{"Insufficient data for analysis"},
		}
		return result, nil
	}

	result.Result = &types.DetailedResult{
		RootCause: &types.RootCause{
			Type:        bestAnalysis.RootCauseType,
			Description: bestAnalysis.Description,
			Confidence:  bestAnalysis.Confidence,
			Evidence:    bestAnalysis.Evidence,
		},
		Confidence: bestAnalysis.Confidence,
		Evidence:   bestAnalysis.Evidence,
	}

	if bestAnalysis.LLMAnalysis != "" {
		result.Result.LLMAnalysis = bestAnalysis.LLMAnalysis
	}

	return result, nil
}

type analysisResult struct {
	RootCauseType types.RootCauseType
	Description   string
	Confidence    float64
	Evidence      []string
	LLMAnalysis   string
}

// analyzeEvent analyzes Kubernetes event.
func (a *RootCauseAnalyzer) analyzeEvent(event map[string]interface{}) *analysisResult {
	reason, _ := event["reason"].(string)
	if reason == "" {
		return nil
	}

	// Direct mapping from event reason
	reasonMap := map[string]types.RootCauseType{
		"OOMKilling":         types.OOMKiller,
		"OOMKilled":          types.OOMKiller,
		"FailedScheduling":   types.ResourceLimit,
		"ImagePullBackOff":   types.ImagePullError,
		"ErrImagePull":       types.ImagePullError,
		"FailedMount":        types.VolumeError,
		"FailedAttachVolume": types.VolumeError,
		"BackOff":            types.ConfigError,
		"CrashLoopBackOff":   types.ConfigError,
	}

	rootCause, found := reasonMap[reason]
	if !found {
		return nil
	}

	message, _ := event["message"].(string)
	evidence := []string{
		fmt.Sprintf("Event reason: %s", reason),
	}
	if message != "" {
		evidence = append(evidence, fmt.Sprintf("Event message: %s", message))
	}

	return &analysisResult{
		RootCauseType: rootCause,
		Description:   a.getDescription(rootCause),
		Confidence:    0.9, // High confidence for direct event mapping
		Evidence:      evidence,
	}
}

// analyzeLogs analyzes logs using pattern matching.
func (a *RootCauseAnalyzer) analyzeLogs(logs string) *analysisResult {
	if logs == "" {
		return nil
	}

	// Limit log size
	if len(logs) > a.config.Performance.MaxContextSize {
		logs = logs[:a.config.Performance.MaxContextSize]
	}

	var bestMatch *analysisResult
	bestMatchCount := 0

	for rootCause, patterns := range a.patterns {
		matchCount := 0
		var matchedPatterns []string

		for _, pattern := range patterns {
			if matches := pattern.FindAllString(logs, -1); len(matches) > 0 {
				matchCount += len(matches)
				matchedPatterns = append(matchedPatterns, fmt.Sprintf("%s (%d occurrences)", pattern.String(), len(matches)))
			}
		}

		if matchCount > bestMatchCount {
			bestMatchCount = matchCount
			confidence := float64(matchCount) / 10.0 // Scale to 0-1
			if confidence > 0.95 {
				confidence = 0.95
			}

			bestMatch = &analysisResult{
				RootCauseType: rootCause,
				Description:   a.getDescription(rootCause),
				Confidence:    confidence,
				Evidence: []string{
					fmt.Sprintf("Found %d matching patterns in logs", matchCount),
					fmt.Sprintf("Matched patterns: %s", strings.Join(matchedPatterns, ", ")),
				},
			}
		}
	}

	return bestMatch
}

// analyzeMetrics analyzes metrics to detect issues.
func (a *RootCauseAnalyzer) analyzeMetrics(metrics *types.MetricsData) *analysisResult {
	var evidence []string
	var rootCause types.RootCauseType
	confidence := 0.0

	// Memory analysis
	if metrics.Memory != nil {
		if metrics.Memory.UsagePercent >= 95 {
			rootCause = types.OOMKiller
			confidence = 0.85
			evidence = append(evidence, fmt.Sprintf("Memory usage at %.1f%% (critical)", metrics.Memory.UsagePercent))
		} else if metrics.Memory.UsagePercent >= 85 {
			rootCause = types.OOMKiller
			confidence = 0.6
			evidence = append(evidence, fmt.Sprintf("Memory usage at %.1f%% (warning)", metrics.Memory.UsagePercent))
		}
	}

	// CPU analysis
	if metrics.CPU != nil {
		if metrics.CPU.ThrottlingPercent >= 80 {
			if confidence < 0.85 {
				rootCause = types.CPUThrottling
				confidence = 0.85
				evidence = []string{fmt.Sprintf("CPU throttling at %.1f%% (critical)", metrics.CPU.ThrottlingPercent)}
			}
		} else if metrics.CPU.ThrottlingPercent >= 60 {
			if confidence < 0.6 {
				rootCause = types.CPUThrottling
				confidence = 0.6
				evidence = []string{fmt.Sprintf("CPU throttling at %.1f%% (warning)", metrics.CPU.ThrottlingPercent)}
			}
		}
	}

	// Disk analysis
	if metrics.Disk != nil {
		if metrics.Disk.UsagePercent >= 95 {
			if confidence < 0.85 {
				rootCause = types.DiskPressure
				confidence = 0.85
				evidence = []string{fmt.Sprintf("Disk usage at %.1f%% (critical)", metrics.Disk.UsagePercent)}
			}
		}
	}

	// Network analysis
	if metrics.Network != nil {
		if metrics.Network.ErrorRate >= 0.1 { // 10% error rate
			if confidence < 0.75 {
				rootCause = types.NetworkError
				confidence = 0.75
				evidence = []string{fmt.Sprintf("Network error rate at %.1f%%", metrics.Network.ErrorRate*100)}
			}
		}
	}

	if confidence == 0 {
		return nil
	}

	return &analysisResult{
		RootCauseType: rootCause,
		Description:   a.getDescription(rootCause),
		Confidence:    confidence,
		Evidence:      evidence,
	}
}

// analyzeLLM uses LLM to analyze the context.
func (a *RootCauseAnalyzer) analyzeLLM(ctx context.Context, req *types.AnalysisRequest) (*analysisResult, error) {
	// Try LLM clients in priority order
	var lastErr error
	for _, client := range a.llmClients {
		if !client.IsAvailable() {
			continue
		}

		metricsJSON, _ := json.Marshal(req.Context.Metrics)
		response, err := client.AnalyzeRootCause(ctx, req.Context.Event, req.Context.Logs, string(metricsJSON))
		if err != nil {
			fmt.Printf("[DEBUG] LLM %s error: %v\n", client.Provider(), err)
			lastErr = err
			continue
		}

		fmt.Printf("[DEBUG] LLM %s raw response (first 500 chars): %s\n", client.Provider(), func() string {
			if len(response) > 500 {
				return response[:500] + "..."
			}
			return response
		}())

		// Parse LLM response
		var llmResult struct {
			RootCauseType string   `json:"root_cause_type"`
			Confidence    float64  `json:"confidence"`
			Evidence      []string `json:"evidence"`
			Explanation   string   `json:"explanation"`
		}

		// Try to extract JSON from response
		response = strings.TrimSpace(response)
		if strings.HasPrefix(response, "```json") {
			response = strings.TrimPrefix(response, "```json")
			response = strings.TrimSuffix(response, "```")
			response = strings.TrimSpace(response)
		} else if strings.HasPrefix(response, "```") {
			response = strings.TrimPrefix(response, "```")
			response = strings.TrimSuffix(response, "```")
			response = strings.TrimSpace(response)
		}

		fmt.Printf("[DEBUG] LLM response after cleanup (first 500 chars): %s\n", func() string {
			if len(response) > 500 {
				return response[:500] + "..."
			}
			return response
		}())

		if err := json.Unmarshal([]byte(response), &llmResult); err != nil {
			fmt.Printf("[DEBUG] JSON parse error: %v\n", err)
			// If we can't parse, return the raw response
			return &analysisResult{
				RootCauseType: types.Unknown,
				Description:   "LLM analysis (unstructured)",
				Confidence:    0.5,
				Evidence:      []string{"LLM provided analysis but in non-standard format"},
				LLMAnalysis:   response,
			}, nil
		}

		fmt.Printf("[DEBUG] Parsed LLM result - Type: %s, Confidence: %.2f, Evidence count: %d\n",
			llmResult.RootCauseType, llmResult.Confidence, len(llmResult.Evidence))

		// Convert string root cause type to typed constant
		rootCause := types.RootCauseType(llmResult.RootCauseType)

		return &analysisResult{
			RootCauseType: rootCause,
			Description:   llmResult.Explanation,
			Confidence:    llmResult.Confidence,
			Evidence:      llmResult.Evidence,
			LLMAnalysis:   response,
		}, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all LLM clients failed: %w", lastErr)
	}

	return nil, fmt.Errorf("no LLM clients available")
}

// getDescription returns a description for a root cause type.
func (a *RootCauseAnalyzer) getDescription(rc types.RootCauseType) string {
	descriptions := map[types.RootCauseType]string{
		// Pod & Container Level
		types.OOMKiller:             "Container was killed due to out of memory (OOM)",
		types.CPUThrottling:         "Container is experiencing CPU throttling due to resource limits",
		types.ReadinessProbeFailure: "Readiness probe is failing, pod not ready to receive traffic",
		types.LivenessProbeFailure:  "Liveness probe is failing, pod may be restarted",
		types.ImagePullError:        "Failed to pull container image",
		types.ConfigError:           "Configuration error preventing container startup",
		types.PodStuckPending:       "Pod is stuck in Pending state, unable to be scheduled",

		// Node Level
		types.NodeNotReady:        "Node is not ready to accept workloads",
		types.KubeletDown:         "Kubelet agent is down on the node",
		types.NodeHighCPU:         "Node CPU usage is critically high",
		types.NodeHighMemory:      "Node memory usage is critically high",
		types.NodeDiskSpaceLow:    "Node disk space is running low",
		types.DiskPressure:        "Node is experiencing disk pressure",
		types.MemoryPressure:      "Node is experiencing memory pressure",
		types.PIDPressure:         "Node is running out of process IDs",
		types.NetworkUnavailable:  "Node network is unavailable or misconfigured",
		types.CertificateExpiring: "Certificate is expiring soon",

		// Control Plane Level
		types.APIServerDown:         "Kubernetes API Server is down or unreachable",
		types.APIServerHighLatency:  "API Server is experiencing high latency",
		types.APIServerHighError:    "API Server is returning high error rates",
		types.ControllerManagerDown: "Controller Manager component is down",
		types.SchedulerDown:         "Scheduler component is down",
		types.EtcdInsufficientPeers: "Etcd cluster has insufficient members",
		types.EtcdHighLatency:       "Etcd is experiencing high latency",
		types.EtcdLeaderFlapping:    "Etcd leader is changing frequently",
		types.CoreDNSDown:           "CoreDNS pods are down or unhealthy",

		// Workload Level
		types.DeploymentReplicasMismatch:  "Deployment replicas don't match desired count",
		types.DeploymentStuck:             "Deployment rollout is stuck",
		types.StatefulSetReplicasMismatch: "StatefulSet replicas don't match desired count",
		types.StatefulSetNotReady:         "StatefulSet is not ready",
		types.DaemonSetRolloutStuck:       "DaemonSet rollout is stuck on some nodes",
		types.JobFailed:                   "Batch job execution failed",
		types.CronJobNotCompleting:        "CronJob is not completing successfully",
		types.HPAMaxedOut:                 "HPA has reached maximum replica count",
		types.PDBAtRisk:                   "PodDisruptionBudget is at risk of violation",

		// Storage & Network Level
		types.PVFillingUp:           "Persistent Volume is filling up",
		types.PVCPending:            "PersistentVolumeClaim is stuck in Pending state",
		types.PVFailed:              "Persistent Volume has failed",
		types.VolumeError:           "Failed to mount or access storage volume",
		types.IngressControllerDown: "Ingress Controller is down",
		types.IngressCertExpiring:   "Ingress TLS certificate is expiring",
		types.NetworkPolicyDenials:  "Network policies are denying traffic",
		types.NetworkError:          "Network connectivity or DNS resolution issues",

		// Quotas & Monitoring
		types.ResourceQuotaExceeded: "Resource quota has been exceeded",
		types.ResourceLimit:         "Insufficient cluster resources for scheduling",
		types.PrometheusTargetDown:  "Prometheus target sync is failing",
		types.PrometheusRuleFailure: "Prometheus alert rule evaluation failed",
		types.AlertmanagerConfigBad: "Alertmanager configuration is inconsistent",
	}

	if desc, ok := descriptions[rc]; ok {
		return desc
	}
	return "Unknown root cause"
}
