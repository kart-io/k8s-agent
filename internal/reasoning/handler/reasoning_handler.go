// Package handler provides unified gRPC and HTTP handlers for the reasoning service.
// This package follows the OneX architecture pattern where a single handler implements
// both server interfaces, ensuring consistent business logic across protocols.
package handler

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/kart-io/k8s-agent/internal/reasoning/analyzer"
	"github.com/kart-io/k8s-agent/internal/reasoning/types"
	reasoningv1 "github.com/kart-io/k8s-agent/pkg/api/reasoning/v1"
	"github.com/kart-io/logger/core"
)

const (
	// Default analysis parameters.
	defaultMinConfidence      = 0.6
	defaultMaxRecommendations = 5
	defaultPriority           = 5
)

// ReasoningHandler implements both ReasoningServiceServer (gRPC) and ReasoningServiceHTTPServer (HTTP).
// This follows the OneX architecture pattern where a single handler implements both interfaces.
type ReasoningHandler struct {
	reasoningv1.UnimplementedReasoningServiceServer

	analyzer *analyzer.RootCauseAnalyzer
	logger   core.Logger
}

// NewReasoningHandler creates a new unified handler for both gRPC and HTTP.
func NewReasoningHandler(
	analyzer *analyzer.RootCauseAnalyzer,
	logger core.Logger,
) *ReasoningHandler {
	return &ReasoningHandler{
		analyzer: analyzer,
		logger:   logger.With("component", "reasoning-handler"),
	}
}

// RootCauseAnalysis performs root cause analysis.
// This method implements both:
//   - reasoningv1.ReasoningServiceServer.RootCauseAnalysis (gRPC)
//   - reasoningv1.ReasoningServiceHTTPServer.RootCauseAnalysis (HTTP)
func (h *ReasoningHandler) RootCauseAnalysis(
	ctx context.Context,
	req *reasoningv1.RootCauseAnalysisRequest,
) (*reasoningv1.RootCauseAnalysisResponse, error) {
	h.logger.Infow("Root cause analysis request",
		"event_id", req.EventId,
	)

	startTime := time.Now()

	// Convert proto request to internal request
	analysisReq := &types.AnalysisRequest{
		RequestID:    req.EventId,
		AnalysisType: "root_cause",
		Context:      convertProtoContext(req.Context),
		Options:      convertProtoOptions(req.Options),
	}

	// Perform analysis
	result, err := h.analyzer.Analyze(ctx, analysisReq)
	if err != nil {
		h.logger.Errorw("Root cause analysis failed",
			"event_id", req.EventId,
			"error", err,
		)
		return nil, status.Errorf(codes.Internal, "analysis failed: %v", err)
	}

	duration := time.Since(startTime)

	h.logger.Infow("Root cause analysis completed",
		"event_id", req.EventId,
		"root_cause", getResultRootCause(result),
		"confidence", getResultConfidence(result),
		"duration_ms", duration.Milliseconds(),
	)

	// Convert result to proto response
	response, err := convertAnalysisResultToProto(result, duration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert result: %v", err)
	}

	return response, nil
}

// SaveCase saves a historical case for learning
// This method implements both:
// - reasoningv1.ReasoningServiceServer.SaveCase (gRPC)
// - reasoningv1.ReasoningServiceHTTPServer.SaveCase (HTTP)
func (h *ReasoningHandler) SaveCase(
	ctx context.Context,
	req *reasoningv1.SaveCaseRequest,
) (*reasoningv1.SaveCaseResponse, error) {
	h.logger.Infow("Save case request",
		"event_id", req.EventId,
		"analysis_id", req.AnalysisId,
		"result", req.Result.String(),
	)

	// TODO: Implement case saving to knowledge base/database
	// For now, just acknowledge the request
	caseID := req.EventId + "-" + req.AnalysisId

	h.logger.Infow("Case saved",
		"case_id", caseID,
		"solution", req.Solution,
		"result", req.Result.String(),
	)

	return &reasoningv1.SaveCaseResponse{
		CaseId:  caseID,
		Success: true,
	}, nil
}

// Helper functions for type conversion

func convertProtoContext(protoCtx *reasoningv1.AnalysisContext) types.AnalysisContext {
	ctx := types.AnalysisContext{}

	if len(protoCtx.Events) > 0 {
		// Convert events to map[string]interface{}
		ctx.Event = make(map[string]interface{})
		ctx.Event["events"] = protoCtx.Events
	}

	if len(protoCtx.Logs) > 0 {
		ctx.Logs = protoCtx.Logs[0] // Take first log entry
	}

	if protoCtx.Metrics != nil {
		ctx.Metrics = convertProtoMetrics(protoCtx.Metrics)
	}

	if protoCtx.Resources != nil {
		// Convert resources struct to map
		ctx.Event["resources"] = protoCtx.Resources.AsMap()
	}

	return ctx
}

func convertProtoMetrics(protoMetrics *structpb.Struct) types.MetricsData {
	metricsMap := protoMetrics.AsMap()
	metrics := types.MetricsData{}

	// Extract memory metrics
	if memMap, ok := metricsMap["memory"].(map[string]interface{}); ok {
		metrics.Memory = &types.MemoryMetrics{
			UsagePercent: getFloat64(memMap, "usage_percent"),
			UsageBytes:   getInt64(memMap, "usage_bytes"),
			LimitBytes:   getInt64(memMap, "limit_bytes"),
		}
	}

	// Extract CPU metrics
	if cpuMap, ok := metricsMap["cpu"].(map[string]interface{}); ok {
		metrics.CPU = &types.CPUMetrics{
			UsagePercent:      getFloat64(cpuMap, "usage_percent"),
			ThrottlingPercent: getFloat64(cpuMap, "throttling_percent"),
			LimitCores:        getFloat64(cpuMap, "limit_cores"),
		}
	}

	// Extract disk metrics
	if diskMap, ok := metricsMap["disk"].(map[string]interface{}); ok {
		metrics.Disk = &types.DiskMetrics{
			UsagePercent: getFloat64(diskMap, "usage_percent"),
			UsageBytes:   getInt64(diskMap, "usage_bytes"),
			TotalBytes:   getInt64(diskMap, "total_bytes"),
		}
	}

	// Extract network metrics
	if netMap, ok := metricsMap["network"].(map[string]interface{}); ok {
		metrics.Network = &types.NetworkMetrics{
			ErrorRate:  getFloat64(netMap, "error_rate"),
			Latency:    getFloat64(netMap, "latency"),
			PacketLoss: getFloat64(netMap, "packet_loss"),
		}
	}

	return metrics
}

func convertProtoOptions(protoOpts *reasoningv1.AnalysisOptions) types.AnalysisOptions {
	if protoOpts == nil {
		return types.AnalysisOptions{
			UseLLM:             true,
			MinConfidence:      defaultMinConfidence,
			LLMProvider:        "openai",
			MaxRecommendations: defaultMaxRecommendations,
		}
	}

	return types.AnalysisOptions{
		MinConfidence:       defaultMinConfidence,
		IncludeSimilarCases: protoOpts.IncludeHistoricalCases,
		MaxRecommendations:  defaultMaxRecommendations,
		UseLLM:              protoOpts.UseKnowledgeGraph, // Map knowledge graph to LLM
		LLMProvider:         protoOpts.Model,
	}
}

func convertAnalysisResultToProto(
	result *types.AnalysisResult,
	duration time.Duration,
) (*reasoningv1.RootCauseAnalysisResponse, error) {
	// Check if we have a detailed result
	if result.Result == nil {
		return nil, fmt.Errorf("analysis result is empty")
	}

	// Generate analysis ID
	analysisID := result.RequestID + "-" + time.Now().Format("20060102150405")

	// Convert root cause
	var protoRootCause *reasoningv1.RootCause
	if result.Result.RootCause != nil {
		protoRootCause = convertRootCauseToProto(result.Result.RootCause)
	} else {
		// Provide default root cause
		protoRootCause = &reasoningv1.RootCause{
			Type:        reasoningv1.RootCause_UNKNOWN,
			Description: "Unable to determine root cause",
			Confidence:  0,
		}
	}

	// Convert recommendations
	protoRecommendations := make([]*reasoningv1.Recommendation, 0, len(result.Result.Recommendations))
	for i, rec := range result.Result.Recommendations {
		// Extract command list from Command field
		commands := []string{}
		if rec.Command != "" {
			commands = append(commands, rec.Command)
		}

		protoRec := &reasoningv1.Recommendation{
			Id:             fmt.Sprintf("%s-rec-%d", analysisID, i),
			Type:           convertRecommendationType(rec.Risk),
			Title:          rec.Action,
			Description:    rec.Description,
			Steps:          rec.Steps,
			Commands:       commands,
			Priority:       defaultPriority, // Default priority
			ExpectedResult: rec.Impact,
			RiskAssessment: rec.Risk,
		}
		protoRecommendations = append(protoRecommendations, protoRec)
	}

	// Convert similar cases (if any)
	protoSimilarCases := make([]*reasoningv1.HistoricalCase, 0, len(result.Result.SimilarCases))
	for _, caseStudy := range result.Result.SimilarCases {
		protoCase := &reasoningv1.HistoricalCase{
			Id:          caseStudy.ID,
			Title:       caseStudy.Title,
			Description: caseStudy.Description,
			Solution:    caseStudy.Solution,
			Similarity:  float32(caseStudy.Similarity),
		}
		if !caseStudy.CreatedAt.IsZero() {
			protoCase.OccurredAt = timestamppb.New(caseStudy.CreatedAt)
		}
		protoSimilarCases = append(protoSimilarCases, protoCase)
	}

	return &reasoningv1.RootCauseAnalysisResponse{
		AnalysisId:      analysisID,
		RootCause:       protoRootCause,
		Recommendations: protoRecommendations,
		SimilarCases:    protoSimilarCases,
		Confidence:      float32(result.Result.Confidence),
		DurationMs:      duration.Milliseconds(),
	}, nil
}

func convertRootCauseToProto(rootCause *types.RootCause) *reasoningv1.RootCause {
	// Convert evidence strings to Evidence objects
	protoEvidences := make([]*reasoningv1.Evidence, 0, len(rootCause.Evidence))
	for _, evidenceStr := range rootCause.Evidence {
		protoEvidence := &reasoningv1.Evidence{
			Type:      reasoningv1.Evidence_LOG, // Default to LOG type
			Content:   evidenceStr,
			Source:    "analyzer",
			Timestamp: timestamppb.Now(),
		}
		protoEvidences = append(protoEvidences, protoEvidence)
	}

	return &reasoningv1.RootCause{
		Type:        convertRootCauseType(rootCause.Type),
		Description: rootCause.Description,
		Impact:      "", // Not in current type definition
		Evidences:   protoEvidences,
		Confidence:  float32(rootCause.Confidence),
	}
}

func convertRootCauseType(rcType types.RootCauseType) reasoningv1.RootCause_Type {
	switch rcType {
	case types.OOMKiller, types.NodeHighMemory, types.MemoryPressure:
		return reasoningv1.RootCause_RESOURCE_EXHAUSTION
	case types.ConfigError:
		return reasoningv1.RootCause_CONFIGURATION_ERROR
	case types.NetworkError, types.NetworkUnavailable:
		return reasoningv1.RootCause_NETWORK_ISSUE
	case types.ImagePullError:
		return reasoningv1.RootCause_IMAGE_PULL_ERROR
	case types.LivenessProbeFailure, types.ReadinessProbeFailure:
		return reasoningv1.RootCause_CRASH_LOOP
	case types.VolumeError, types.PVCPending:
		return reasoningv1.RootCause_DEPENDENCY_FAILURE
	default:
		return reasoningv1.RootCause_UNKNOWN
	}
}

func convertRecommendationType(risk string) reasoningv1.Recommendation_Type {
	switch risk {
	case "low":
		return reasoningv1.Recommendation_AUTOMATED
	case "high", "critical":
		return reasoningv1.Recommendation_MANUAL
	default:
		return reasoningv1.Recommendation_MANUAL
	}
}

// Helper functions for getting result fields safely
func getResultRootCause(result *types.AnalysisResult) string {
	if result.Result != nil && result.Result.RootCause != nil {
		return string(result.Result.RootCause.Type)
	}
	return "unknown"
}

func getResultConfidence(result *types.AnalysisResult) float64 {
	if result.Result != nil {
		return result.Result.Confidence
	}
	return 0
}

// Helper functions for extracting values from maps
func getFloat64(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}

func getInt64(m map[string]interface{}, key string) int64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return 0
}
