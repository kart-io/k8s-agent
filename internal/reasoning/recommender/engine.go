package recommender

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/kart-io/k8s-agent/internal/reasoning/config"
	"github.com/kart-io/k8s-agent/internal/reasoning/llm"
	"github.com/kart-io/k8s-agent/internal/reasoning/types"
)

// Engine generates recommendations based on root cause.
type Engine struct {
	config     *config.Config
	llmClients []llm.Client
	rules      map[types.RootCauseType][]types.Recommendation
}

// NewEngine creates a new recommendation engine.
func NewEngine(cfg *config.Config, llmClients []llm.Client) *Engine {
	engine := &Engine{
		config:     cfg,
		llmClients: llmClients,
		rules:      make(map[types.RootCauseType][]types.Recommendation),
	}
	engine.loadRules()
	return engine
}

// loadRules loads recommendation rules.
func (e *Engine) loadRules() {
	e.rules[types.OOMKiller] = []types.Recommendation{
		{
			Action:      "increase_memory_limit",
			Description: "Increase container memory limits to prevent OOM kills",
			Confidence:  0.90,
			Risk:        "low",
			Impact:      "Prevents future OOM kills, may increase cluster resource usage",
			Steps: []string{
				"Analyze current memory usage patterns",
				"Calculate recommended memory limit (current + 50%)",
				"Update Deployment/StatefulSet memory limits",
				"kubectl apply -f updated-manifest.yaml",
				"Monitor for OOM recurrence",
			},
			RollbackSteps: []string{
				"Revert to previous memory limits",
				"kubectl rollout undo deployment/<name>",
			},
			EstimatedDuration: "5 minutes",
		},
		{
			Action:      "investigate_memory_leak",
			Description: "Investigate potential memory leak in application",
			Confidence:  0.70,
			Risk:        "none",
			Impact:      "Identify and fix memory leaks for long-term stability",
			Steps: []string{
				"Enable memory profiling",
				"Collect heap dumps",
				"Analyze memory allocation patterns",
				"Fix memory leaks in application code",
			},
			EstimatedDuration: "2-4 hours",
		},
	}

	e.rules[types.CPUThrottling] = []types.Recommendation{
		{
			Action:      "increase_cpu_limit",
			Description: "Increase CPU limits to reduce throttling",
			Confidence:  0.85,
			Risk:        "low",
			Impact:      "Improves application performance, increases resource usage",
			Steps: []string{
				"Analyze CPU usage patterns",
				"Increase CPU limits (e.g., from 1 to 2 cores)",
				"Update resource limits in manifest",
				"Apply changes and monitor throttling metrics",
			},
			RollbackSteps: []string{
				"Revert to previous CPU limits",
			},
			EstimatedDuration: "5 minutes",
		},
	}

	e.rules[types.ImagePullError] = []types.Recommendation{
		{
			Action:      "check_image_access",
			Description: "Verify image repository access and credentials",
			Confidence:  0.95,
			Risk:        "none",
			Impact:      "Resolves image pull failures",
			Steps: []string{
				"Verify image name and tag are correct",
				"Check image exists in repository",
				"Verify imagePullSecrets are configured",
				"Test docker login with credentials",
				"Update imagePullSecrets if needed",
			},
			EstimatedDuration: "10 minutes",
		},
	}

	e.rules[types.NetworkError] = []types.Recommendation{
		{
			Action:      "check_network_policies",
			Description: "Review network policies and connectivity",
			Confidence:  0.80,
			Risk:        "none",
			Impact:      "Identifies and resolves network connectivity issues",
			Steps: []string{
				"Check NetworkPolicies affecting the pod",
				"Verify service endpoints are healthy",
				"Test DNS resolution from pod",
				"Check firewall rules",
				"Review security groups/NSGs",
			},
			EstimatedDuration: "15-30 minutes",
		},
	}

	e.rules[types.ReadinessProbeFailure] = []types.Recommendation{
		{
			Action:      "adjust_readiness_probe_timing",
			Description: "增加就绪探针的初始延迟时间和超时时间，给应用更多启动时间",
			Command:     "kubectl edit deployment <deployment-name> -n <namespace>",
			YAML: `readinessProbe:
  httpGet:
    path: /ping
    port: 8080
  initialDelaySeconds: 30
  timeoutSeconds: 5
  periodSeconds: 10
  failureThreshold: 3`,
			Confidence: 0.90,
			Risk:       "low",
			Impact:     "Pod 启动时会等待更长时间才被标记为 Ready",
			Steps: []string{
				"Analyze current readiness probe configuration",
				"Increase initialDelaySeconds to allow more startup time",
				"Increase timeoutSeconds if application needs more response time",
				"Update Deployment/StatefulSet manifest",
				"kubectl apply -f updated-manifest.yaml",
				"Monitor pod readiness status",
			},
			RollbackSteps: []string{
				"Revert to previous readiness probe configuration",
				"kubectl rollout undo deployment/<name>",
			},
			EstimatedDuration: "5 minutes",
		},
		{
			Action:      "increase_pod_resources",
			Description: "提高 CPU 和内存限制，确保应用有足够资源响应健康检查",
			Command:     "kubectl set resources deployment <deployment-name> -n <namespace> --limits=cpu=500m,memory=512Mi",
			YAML: `resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi`,
			Confidence: 0.75,
			Risk:       "medium",
			Impact:     "需要集群有足够可用资源，可能触发 Pod 重新调度",
			Steps: []string{
				"Check current resource usage",
				"Increase CPU and memory limits",
				"Update resource limits in manifest",
				"Apply changes and monitor resource metrics",
			},
			RollbackSteps: []string{
				"Revert to previous resource limits",
			},
			EstimatedDuration: "5 minutes",
		},
		{
			Action:      "check_application_logs",
			Description: "检查应用日志，确认启动失败或响应超时的具体原因",
			Command:     "kubectl logs <pod-name> -n <namespace> --tail=100",
			Confidence:  0.85,
			Risk:        "low",
			Impact:      "仅查看日志，不会对集群产生影响",
			Steps: []string{
				"View recent pod logs",
				"Look for application startup errors",
				"Check for slow initialization or timeouts",
				"Identify and fix application issues",
			},
			EstimatedDuration: "10-15 minutes",
		},
	}

	e.rules[types.LivenessProbeFailure] = []types.Recommendation{
		{
			Action:      "adjust_liveness_probe_timing",
			Description: "增加存活探针的初始延迟时间和超时时间，避免误杀正常 Pod",
			Command:     "kubectl edit deployment <deployment-name> -n <namespace>",
			YAML: `livenessProbe:
  httpGet:
    path: /ping
    port: 8080
  initialDelaySeconds: 60
  timeoutSeconds: 5
  periodSeconds: 10
  failureThreshold: 3`,
			Confidence: 0.90,
			Risk:       "low",
			Impact:     "Pod 启动后会等待更长时间才开始存活检查",
			Steps: []string{
				"Analyze current liveness probe configuration",
				"Increase initialDelaySeconds to prevent early kills",
				"Increase failureThreshold to allow more tolerance",
				"Update Deployment/StatefulSet manifest",
				"kubectl apply -f updated-manifest.yaml",
				"Monitor pod status and restarts",
			},
			RollbackSteps: []string{
				"Revert to previous liveness probe configuration",
				"kubectl rollout undo deployment/<name>",
			},
			EstimatedDuration: "5 minutes",
		},
		{
			Action:      "investigate_application_hang",
			Description: "调查应用是否出现死锁或挂起，导致无法响应存活检查",
			Command:     "kubectl logs <pod-name> -n <namespace> --tail=100",
			Confidence:  0.80,
			Risk:        "low",
			Impact:      "仅查看日志和状态，不会对集群产生影响",
			Steps: []string{
				"View recent pod logs",
				"Check for application deadlocks or hangs",
				"Review thread dumps if available",
				"Identify and fix application issues",
			},
			EstimatedDuration: "15-30 minutes",
		},
		{
			Action:      "check_resource_constraints",
			Description: "检查是否因资源不足导致应用响应缓慢或挂起",
			Command:     "kubectl top pod <pod-name> -n <namespace>",
			Confidence:  0.75,
			Risk:        "low",
			Impact:      "仅查看资源使用情况，不会对集群产生影响",
			Steps: []string{
				"Check pod resource usage",
				"Compare with resource limits",
				"Identify resource bottlenecks",
				"Consider increasing resource limits if needed",
			},
			EstimatedDuration: "10 minutes",
		},
	}

	// Node Level Alerts - 节点层面告警
	e.rules[types.NodeNotReady] = []types.Recommendation{
		{
			Action:      "check_node_status",
			Description: "检查节点状态和 Kubelet 日志，诊断节点不可用的原因",
			Command:     "kubectl describe node <node-name>",
			Confidence:  0.90,
			Risk:        "low",
			Impact:      "仅查看状态，不影响集群",
			Steps: []string{
				"kubectl get nodes",
				"kubectl describe node <node-name>",
				"Check kubelet logs: journalctl -u kubelet -n 100",
				"Check node conditions and events",
			},
			EstimatedDuration: "10 minutes",
		},
		{
			Action:      "restart_kubelet",
			Description: "重启节点上的 Kubelet 服务",
			Command:     "systemctl restart kubelet",
			Confidence:  0.75,
			Risk:        "medium",
			Impact:      "可能导致节点上的 Pod 短暂中断",
			Steps: []string{
				"SSH to the node",
				"systemctl restart kubelet",
				"Check kubelet status: systemctl status kubelet",
			},
			EstimatedDuration: "5 minutes",
		},
	}

	e.rules[types.NodeHighMemory] = []types.Recommendation{
		{
			Action:      "identify_memory_consumers",
			Description: "识别节点上消耗内存最多的 Pod",
			Command:     "kubectl top pods --all-namespaces --sort-by=memory | head -20",
			Confidence:  0.85,
			Risk:        "low",
			Impact:      "仅查看资源使用情况",
			Steps: []string{
				"kubectl top nodes",
				"kubectl top pods --all-namespaces --sort-by=memory",
				"Check for memory leaks in top consumers",
				"Consider evicting or rescheduling heavy pods",
			},
			EstimatedDuration: "10 minutes",
		},
		{
			Action:      "add_node_capacity",
			Description: "增加集群节点或扩容现有节点",
			Command:     "kubectl scale deployment <autoscaler> --replicas=<new-count>",
			Confidence:  0.70,
			Risk:        "high",
			Impact:      "需要集群有足够资源或云服务商额度",
			Steps: []string{
				"Evaluate cluster capacity needs",
				"Add new nodes to cluster",
				"Redistribute workloads",
			},
			EstimatedDuration: "30-60 minutes",
		},
	}

	// Control Plane Alerts - 控制平面告警
	e.rules[types.APIServerDown] = []types.Recommendation{
		{
			Action:      "check_apiserver_status",
			Description: "检查 API Server Pod 状态和日志",
			Command:     "kubectl get pods -n kube-system | grep kube-apiserver",
			Confidence:  0.95,
			Risk:        "low",
			Impact:      "仅诊断问题，不影响集群",
			Steps: []string{
				"kubectl get pods -n kube-system -l component=kube-apiserver",
				"kubectl logs -n kube-system kube-apiserver-<node-name>",
				"Check API server process on control plane node",
				"Review /var/log/kube-apiserver.log",
			},
			EstimatedDuration: "10 minutes",
		},
		{
			Action:      "restart_apiserver",
			Description: "重启 API Server 服务（高风险操作）",
			Command:     "systemctl restart kube-apiserver",
			Confidence:  0.60,
			Risk:        "high",
			Impact:      "集群 API 将短暂不可用，影响所有操作",
			Steps: []string{
				"Backup etcd data first",
				"systemctl restart kube-apiserver",
				"Monitor API server recovery",
			},
			EstimatedDuration: "5-10 minutes",
		},
	}

	e.rules[types.EtcdHighLatency] = []types.Recommendation{
		{
			Action:      "check_etcd_metrics",
			Description: "检查 Etcd 性能指标和磁盘 IO",
			Command:     "kubectl exec -n kube-system etcd-<node> -- etcdctl endpoint status --cluster -w table",
			Confidence:  0.85,
			Risk:        "low",
			Impact:      "仅查看指标，不影响集群",
			Steps: []string{
				"Check etcd metrics",
				"Monitor disk IO latency: iostat -x 1",
				"Check for slow disk or network",
				"Review etcd logs for warnings",
			},
			EstimatedDuration: "15 minutes",
		},
		{
			Action:      "optimize_etcd_disk",
			Description: "优化 Etcd 存储，使用更快的磁盘（SSD）",
			Confidence:  0.75,
			Risk:        "high",
			Impact:      "需要迁移数据或更换存储设备",
			Steps: []string{
				"Backup etcd data",
				"Migrate to SSD storage",
				"Defragment etcd database: etcdctl defrag",
			},
			EstimatedDuration: "1-2 hours",
		},
	}

	// Workload Alerts - 工作负载告警
	e.rules[types.PodStuckPending] = []types.Recommendation{
		{
			Action:      "check_scheduling_issues",
			Description: "检查 Pod 无法调度的原因（资源不足、节点选择器、污点等）",
			Command:     "kubectl describe pod <pod-name> -n <namespace>",
			Confidence:  0.95,
			Risk:        "low",
			Impact:      "仅诊断问题",
			Steps: []string{
				"kubectl describe pod <pod-name>",
				"Check Events section for scheduling failures",
				"kubectl get nodes -o wide",
				"Check resource availability on nodes",
			},
			EstimatedDuration: "10 minutes",
		},
		{
			Action:      "adjust_resource_requests",
			Description: "降低 Pod 的资源请求，使其能够被调度",
			Command:     "kubectl edit deployment <deployment-name> -n <namespace>",
			YAML: `resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi`,
			Confidence: 0.80,
			Risk:       "medium",
			Impact:     "Pod 可用资源减少，可能影响性能",
			Steps: []string{
				"Edit deployment resource requests",
				"Reduce CPU/memory requests",
				"Monitor pod startup",
			},
			EstimatedDuration: "5 minutes",
		},
	}

	e.rules[types.DeploymentReplicasMismatch] = []types.Recommendation{
		{
			Action:      "check_deployment_events",
			Description: "检查 Deployment 事件，查看副本不匹配的原因",
			Command:     "kubectl describe deployment <deployment-name> -n <namespace>",
			Confidence:  0.90,
			Risk:        "low",
			Impact:      "仅诊断问题",
			Steps: []string{
				"kubectl get deployment <name> -o wide",
				"kubectl describe deployment <name>",
				"kubectl get replicasets -l app=<name>",
				"Check for image pull errors or resource issues",
			},
			EstimatedDuration: "10 minutes",
		},
		{
			Action:      "rollback_deployment",
			Description: "回滚到上一个稳定版本",
			Command:     "kubectl rollout undo deployment/<deployment-name> -n <namespace>",
			Confidence:  0.85,
			Risk:        "medium",
			Impact:      "回滚到之前的版本，可能丢失新功能",
			Steps: []string{
				"kubectl rollout history deployment/<name>",
				"kubectl rollout undo deployment/<name>",
				"Monitor rollout status",
			},
			EstimatedDuration: "5 minutes",
		},
	}

	e.rules[types.HPAMaxedOut] = []types.Recommendation{
		{
			Action:      "increase_hpa_max_replicas",
			Description: "增加 HPA 的最大副本数限制",
			Command:     "kubectl edit hpa <hpa-name> -n <namespace>",
			YAML: `spec:
  maxReplicas: 20  # 从 10 增加到 20
  minReplicas: 2`,
			Confidence: 0.90,
			Risk:       "medium",
			Impact:     "可能消耗更多集群资源",
			Steps: []string{
				"kubectl get hpa",
				"Edit HPA maxReplicas",
				"Monitor scaling behavior",
			},
			EstimatedDuration: "5 minutes",
		},
		{
			Action:      "optimize_resource_usage",
			Description: "优化应用资源使用，减少单个 Pod 的资源消耗",
			Confidence:  0.75,
			Risk:        "low",
			Impact:      "需要代码或配置优化",
			Steps: []string{
				"Profile application performance",
				"Optimize code for lower resource usage",
				"Add caching or connection pooling",
				"Review and fix inefficient queries",
			},
			EstimatedDuration: "1-2 days",
		},
	}

	// Storage Alerts - 存储告警
	e.rules[types.PVCPending] = []types.Recommendation{
		{
			Action:      "check_pvc_status",
			Description: "检查 PVC 为何无法绑定到 PV",
			Command:     "kubectl describe pvc <pvc-name> -n <namespace>",
			Confidence:  0.95,
			Risk:        "low",
			Impact:      "仅诊断问题",
			Steps: []string{
				"kubectl get pvc",
				"kubectl describe pvc <name>",
				"kubectl get pv",
				"Check StorageClass availability",
			},
			EstimatedDuration: "10 minutes",
		},
		{
			Action:      "create_matching_pv",
			Description: "手动创建匹配的 PV 或配置动态 provisioner",
			YAML: `apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv-example
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: standard
  hostPath:
    path: /data/pv-example`,
			Confidence: 0.80,
			Risk:       "medium",
			Impact:     "创建新的存储资源",
			Steps: []string{
				"Create PV with matching specs",
				"kubectl apply -f pv.yaml",
				"Verify PVC binding",
			},
			EstimatedDuration: "15 minutes",
		},
	}

	// Add more rules for other root cause types...
}

// GenerateRecommendations generates recommendations.
func (e *Engine) GenerateRecommendations(ctx context.Context, result *types.AnalysisResult, analysisCtx *types.AnalysisContext) error {
	if result.Result == nil || result.Result.RootCause == nil {
		return nil
	}

	// Get rule-based recommendations
	recommendations := e.getRuleBasedRecommendations(result.Result.RootCause.Type)

	// Optionally enhance with LLM
	if e.config.LLM.Enabled && len(e.llmClients) > 0 {
		llmRecs, err := e.getLLMRecommendations(ctx, result.Result.RootCause, analysisCtx)
		if err == nil && len(llmRecs) > 0 {
			// Merge LLM recommendations
			recommendations = append(recommendations, llmRecs...)
		}
	}

	// Limit to max recommendations
	if len(recommendations) > e.config.Analysis.MaxRecommendations {
		recommendations = recommendations[:e.config.Analysis.MaxRecommendations]
	}

	result.Result.Recommendations = recommendations
	return nil
}

func (e *Engine) getRuleBasedRecommendations(rootCause types.RootCauseType) []types.Recommendation {
	if recs, ok := e.rules[rootCause]; ok {
		return recs
	}
	return []types.Recommendation{}
}

func (e *Engine) getLLMRecommendations(ctx context.Context, rootCause *types.RootCause, analysisCtx *types.AnalysisContext) ([]types.Recommendation, error) {
	for _, client := range e.llmClients {
		if !client.IsAvailable() {
			continue
		}

		contextJSON, _ := json.Marshal(analysisCtx)
		response, err := client.GenerateRecommendations(ctx, string(rootCause.Type), string(contextJSON))
		if err != nil {
			continue
		}

		// Parse response
		response = strings.TrimSpace(response)
		if strings.HasPrefix(response, "```json") {
			response = strings.TrimPrefix(response, "```json")
			response = strings.TrimSuffix(response, "```")
			response = strings.TrimSpace(response)
		}

		var recommendations []types.Recommendation
		if err := json.Unmarshal([]byte(response), &recommendations); err != nil {
			continue
		}

		return recommendations, nil
	}

	return nil, nil
}
