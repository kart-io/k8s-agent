package recommender

import (
	"context"
	"encoding/json"
	"strings"

	"reasoning-service-go/internal/config"
	"reasoning-service-go/pkg/llm"
	"reasoning-service-go/pkg/types"
)

// Engine generates recommendations based on root cause
type Engine struct {
	config     *config.Config
	llmClients []llm.Client
	rules      map[types.RootCauseType][]types.Recommendation
}

// NewEngine creates a new recommendation engine
func NewEngine(cfg *config.Config, llmClients []llm.Client) *Engine {
	engine := &Engine{
		config:     cfg,
		llmClients: llmClients,
		rules:      make(map[types.RootCauseType][]types.Recommendation),
	}
	engine.loadRules()
	return engine
}

// loadRules loads recommendation rules
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

	// Add more rules for other root cause types...
}

// GenerateRecommendations generates recommendations
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
