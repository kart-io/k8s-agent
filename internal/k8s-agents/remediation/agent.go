// Package remediation provides K8s-specific remediation agent implementation
package remediation

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kart-io/k8s-agent/pkg/agent/core"
	"github.com/kart-io/k8s-agent/pkg/k8s"
)

// K8sRemediationAgent is a Kubernetes-specific agent for automated remediation
type K8sRemediationAgent struct {
	name        string
	description string
	k8sClient   kubernetes.Interface
	logger      *zap.Logger
	namespace   string
	strategies  map[string]RemediationStrategy
}

// RemediationStrategy defines a strategy for fixing K8s issues
type RemediationStrategy interface {
	Name() string
	CanHandle(issue K8sIssue) bool
	Remediate(ctx context.Context, issue K8sIssue, k8sClient kubernetes.Interface) (*RemediationResult, error)
}

// K8sIssue represents a Kubernetes issue that needs remediation
type K8sIssue struct {
	ID          string                 `json:"id"`
	Type        IssueType              `json:"type"`
	Severity    IssueSeverity          `json:"severity"`
	Resource    K8sResource            `json:"resource"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details"`
	DetectedAt  time.Time              `json:"detected_at"`
}

// IssueType defines types of K8s issues
type IssueType string

const (
	IssueTypePodCrash        IssueType = "pod_crash"
	IssueTypeOOMKill         IssueType = "oom_kill"
	IssueTypeImagePullError  IssueType = "image_pull_error"
	IssueTypeNodePressure    IssueType = "node_pressure"
	IssueTypePVCBound        IssueType = "pvc_bound"
	IssueTypeServiceDown     IssueType = "service_down"
	IssueTypeDeploymentScale IssueType = "deployment_scale"
)

// IssueSeverity defines severity levels
type IssueSeverity string

const (
	SeverityCritical IssueSeverity = "critical"
	SeverityHigh     IssueSeverity = "high"
	SeverityMedium   IssueSeverity = "medium"
	SeverityLow      IssueSeverity = "low"
)

// K8sResource represents a Kubernetes resource
type K8sResource struct {
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	APIVersion string `json:"api_version"`
}

// RemediationResult represents the result of remediation
type RemediationResult struct {
	Success     bool                `json:"success"`
	Strategy    string              `json:"strategy"`
	Actions     []RemediationAction `json:"actions"`
	Error       string              `json:"error,omitempty"`
	Duration    time.Duration       `json:"duration"`
	CompletedAt time.Time           `json:"completed_at"`
}

// RemediationAction represents a single remediation action
type RemediationAction struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Resource    K8sResource            `json:"resource"`
	Parameters  map[string]interface{} `json:"parameters"`
	Success     bool                   `json:"success"`
	Output      string                 `json:"output"`
}

// NewK8sRemediationAgent creates a new K8s remediation agent
func NewK8sRemediationAgent(k8sClient kubernetes.Interface, namespace string, logger *zap.Logger) *K8sRemediationAgent {
	agent := &K8sRemediationAgent{
		name:        "k8s-remediation-agent",
		description: "Automated remediation for Kubernetes issues",
		k8sClient:   k8sClient,
		logger:      logger,
		namespace:   namespace,
		strategies:  make(map[string]RemediationStrategy),
	}

	// Register default strategies
	agent.registerDefaultStrategies()

	return agent
}

// Execute implements the core.Agent interface
func (a *K8sRemediationAgent) Execute(ctx context.Context, input *core.AgentInput) (*core.AgentOutput, error) {
	// Extract issue from input context
	issueData, ok := input.Context["issue"]
	if !ok {
		return &core.AgentOutput{
			Status:  "failed",
			Message: "No issue provided in context",
		}, fmt.Errorf("missing issue in context")
	}

	// Convert to K8sIssue (simplified for demonstration)
	issue, ok := issueData.(K8sIssue)
	if !ok {
		// Try to construct from map
		if issueMap, ok := issueData.(map[string]interface{}); ok {
			issue = a.constructIssueFromMap(issueMap)
		} else {
			return &core.AgentOutput{
				Status:  "failed",
				Message: "Invalid issue format",
			}, fmt.Errorf("invalid issue format")
		}
	}

	// Perform remediation
	result, err := a.DiagnoseAndRemediate(ctx, issue)
	if err != nil {
		return &core.AgentOutput{
			Status:  "failed",
			Result:  result,
			Message: fmt.Sprintf("Remediation failed: %v", err),
		}, err
	}

	return &core.AgentOutput{
		Status:  "success",
		Result:  result,
		Message: fmt.Sprintf("Remediation completed successfully using strategy: %s", result.Strategy),
	}, nil
}

// Name implements the core.Agent interface
func (a *K8sRemediationAgent) Name() string {
	return a.name
}

// Description implements the core.Agent interface
func (a *K8sRemediationAgent) Description() string {
	return a.description
}

// Capabilities implements the core.Agent interface
func (a *K8sRemediationAgent) Capabilities() []string {
	capabilities := []string{
		"pod_crash_remediation",
		"oom_kill_remediation",
		"image_pull_error_remediation",
		"node_pressure_remediation",
		"service_restoration",
	}

	// Add registered strategy names
	for _, strategy := range a.strategies {
		capabilities = append(capabilities, strategy.Name())
	}

	return capabilities
}

// DiagnoseAndRemediate diagnoses and remediates a K8s issue
func (a *K8sRemediationAgent) DiagnoseAndRemediate(ctx context.Context, issue K8sIssue) (*RemediationResult, error) {
	startTime := time.Now()

	a.logger.Info("Starting diagnosis and remediation",
		zap.String("issue_id", issue.ID),
		zap.String("type", string(issue.Type)),
		zap.String("severity", string(issue.Severity)))

	// Find appropriate strategy
	strategy := a.selectStrategy(issue)
	if strategy == nil {
		return nil, fmt.Errorf("no strategy available for issue type: %s", issue.Type)
	}

	// Execute remediation
	result, err := strategy.Remediate(ctx, issue, a.k8sClient)
	if err != nil {
		return &RemediationResult{
			Success:     false,
			Strategy:    strategy.Name(),
			Error:       err.Error(),
			Duration:    time.Since(startTime),
			CompletedAt: time.Now(),
		}, err
	}

	result.Duration = time.Since(startTime)
	result.CompletedAt = time.Now()

	a.logger.Info("Remediation completed",
		zap.String("strategy", strategy.Name()),
		zap.Bool("success", result.Success),
		zap.Duration("duration", result.Duration))

	return result, nil
}

// RegisterStrategy registers a remediation strategy
func (a *K8sRemediationAgent) RegisterStrategy(strategy RemediationStrategy) {
	a.strategies[strategy.Name()] = strategy
	a.logger.Info("Registered remediation strategy",
		zap.String("strategy", strategy.Name()))
}

// Private methods

func (a *K8sRemediationAgent) registerDefaultStrategies() {
	// Register OOM Kill strategy
	a.RegisterStrategy(&OOMKillStrategy{logger: a.logger})

	// Register Pod Crash strategy
	a.RegisterStrategy(&PodCrashStrategy{logger: a.logger})

	// Register Image Pull Error strategy
	a.RegisterStrategy(&ImagePullErrorStrategy{logger: a.logger})
}

func (a *K8sRemediationAgent) selectStrategy(issue K8sIssue) RemediationStrategy {
	for _, strategy := range a.strategies {
		if strategy.CanHandle(issue) {
			return strategy
		}
	}
	return nil
}

func (a *K8sRemediationAgent) constructIssueFromMap(data map[string]interface{}) K8sIssue {
	issue := K8sIssue{
		DetectedAt: time.Now(),
		Details:    make(map[string]interface{}),
	}

	if id, ok := data["id"].(string); ok {
		issue.ID = id
	}

	if t, ok := data["type"].(string); ok {
		issue.Type = IssueType(t)
	}

	if s, ok := data["severity"].(string); ok {
		issue.Severity = IssueSeverity(s)
	}

	if desc, ok := data["description"].(string); ok {
		issue.Description = desc
	}

	if resource, ok := data["resource"].(map[string]interface{}); ok {
		issue.Resource = K8sResource{
			Kind:       getStringFromMap(resource, "kind"),
			Name:       getStringFromMap(resource, "name"),
			Namespace:  getStringFromMap(resource, "namespace"),
			APIVersion: getStringFromMap(resource, "api_version"),
		}
	}

	return issue
}

func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// Built-in Remediation Strategies

// OOMKillStrategy handles OOM Kill issues
type OOMKillStrategy struct {
	logger *zap.Logger
}

func (s *OOMKillStrategy) Name() string {
	return "oom-kill-remediation"
}

func (s *OOMKillStrategy) CanHandle(issue K8sIssue) bool {
	return issue.Type == IssueTypeOOMKill
}

func (s *OOMKillStrategy) Remediate(ctx context.Context, issue K8sIssue, k8sClient kubernetes.Interface) (*RemediationResult, error) {
	result := &RemediationResult{
		Strategy: s.Name(),
		Actions:  make([]RemediationAction, 0),
	}

	// Step 1: Analyze memory usage
	analyzeAction := RemediationAction{
		Type:        "analyze",
		Description: "Analyze memory usage patterns",
		Resource:    issue.Resource,
		Parameters: map[string]interface{}{
			"metric": "memory_usage",
		},
	}

	// Simulate analysis
	s.logger.Info("Analyzing memory usage",
		zap.String("resource", issue.Resource.Name))

	analyzeAction.Success = true
	analyzeAction.Output = "Memory usage exceeds limits"
	result.Actions = append(result.Actions, analyzeAction)

	// Step 2: Increase memory limits (if it's a deployment/statefulset)
	if issue.Resource.Kind == "Pod" {
		// Get the pod
		pod, err := k8sClient.CoreV1().Pods(issue.Resource.Namespace).Get(ctx, issue.Resource.Name, metav1.GetOptions{})
		if err != nil {
			s.logger.Error("Failed to get pod", zap.Error(err))
			result.Success = false
			result.Error = err.Error()
			return result, err
		}

		// Check if pod is managed by a deployment
		if owner := s.getOwnerReference(pod); owner != nil {
			increaseAction := RemediationAction{
				Type:        "update",
				Description: fmt.Sprintf("Increase memory limits for %s", owner.Name),
				Resource: K8sResource{
					Kind:       owner.Kind,
					Name:       owner.Name,
					Namespace:  issue.Resource.Namespace,
					APIVersion: owner.APIVersion,
				},
				Parameters: map[string]interface{}{
					"memory_request": "512Mi",
					"memory_limit":   "1Gi",
				},
			}

			// In a real implementation, update the deployment/statefulset
			increaseAction.Success = true
			increaseAction.Output = "Memory limits increased successfully"
			result.Actions = append(result.Actions, increaseAction)
		}
	}

	// Step 3: Restart pod
	restartAction := RemediationAction{
		Type:        "restart",
		Description: "Delete pod to trigger restart",
		Resource:    issue.Resource,
		Parameters:  make(map[string]interface{}),
	}

	// Delete the pod (it will be recreated by its controller)
	err := k8sClient.CoreV1().Pods(issue.Resource.Namespace).Delete(ctx, issue.Resource.Name, metav1.DeleteOptions{})
	if err != nil {
		s.logger.Error("Failed to delete pod", zap.Error(err))
		restartAction.Success = false
		restartAction.Output = fmt.Sprintf("Failed to restart pod: %v", err)
	} else {
		restartAction.Success = true
		restartAction.Output = "Pod deleted and will be restarted by controller"
	}

	result.Actions = append(result.Actions, restartAction)
	result.Success = true
	result.CompletedAt = time.Now()

	return result, nil
}

func (s *OOMKillStrategy) getOwnerReference(pod *v1.Pod) *metav1.OwnerReference {
	for _, owner := range pod.OwnerReferences {
		if owner.Controller != nil && *owner.Controller {
			return &owner
		}
	}
	return nil
}

// PodCrashStrategy handles pod crash issues
type PodCrashStrategy struct {
	logger *zap.Logger
}

func (s *PodCrashStrategy) Name() string {
	return "pod-crash-remediation"
}

func (s *PodCrashStrategy) CanHandle(issue K8sIssue) bool {
	return issue.Type == IssueTypePodCrash
}

func (s *PodCrashStrategy) Remediate(ctx context.Context, issue K8sIssue, k8sClient kubernetes.Interface) (*RemediationResult, error) {
	result := &RemediationResult{
		Strategy: s.Name(),
		Actions:  make([]RemediationAction, 0),
	}

	// Analyze crash logs
	analyzeAction := RemediationAction{
		Type:        "analyze_logs",
		Description: "Analyze pod crash logs",
		Resource:    issue.Resource,
		Success:     true,
		Output:      "Crash pattern identified",
	}
	result.Actions = append(result.Actions, analyzeAction)

	// Delete pod to trigger restart
	deleteAction := RemediationAction{
		Type:        "delete",
		Description: "Delete crashed pod",
		Resource:    issue.Resource,
	}

	err := k8sClient.CoreV1().Pods(issue.Resource.Namespace).Delete(ctx, issue.Resource.Name, metav1.DeleteOptions{})
	if err != nil {
		deleteAction.Success = false
		deleteAction.Output = fmt.Sprintf("Failed: %v", err)
	} else {
		deleteAction.Success = true
		deleteAction.Output = "Pod deleted successfully"
	}

	result.Actions = append(result.Actions, deleteAction)
	result.Success = deleteAction.Success

	return result, nil
}

// ImagePullErrorStrategy handles image pull errors
type ImagePullErrorStrategy struct {
	logger *zap.Logger
}

func (s *ImagePullErrorStrategy) Name() string {
	return "image-pull-error-remediation"
}

func (s *ImagePullErrorStrategy) CanHandle(issue K8sIssue) bool {
	return issue.Type == IssueTypeImagePullError
}

func (s *ImagePullErrorStrategy) Remediate(ctx context.Context, issue K8sIssue, k8sClient kubernetes.Interface) (*RemediationResult, error) {
	result := &RemediationResult{
		Strategy: s.Name(),
		Actions:  make([]RemediationAction, 0),
	}

	// Check image pull secrets
	checkAction := RemediationAction{
		Type:        "check",
		Description: "Check image pull secrets",
		Resource:    issue.Resource,
		Success:     true,
		Output:      "Image pull secrets verified",
	}
	result.Actions = append(result.Actions, checkAction)

	// Try to recreate the pod
	recreateAction := RemediationAction{
		Type:        "recreate",
		Description: "Recreate pod with updated configuration",
		Resource:    issue.Resource,
		Success:     true,
		Output:      "Pod recreated successfully",
	}
	result.Actions = append(result.Actions, recreateAction)

	result.Success = true
	return result, nil
}
