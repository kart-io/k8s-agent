package api

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/k8s-agent/internal/reasoning/orchestrator"
	"github.com/kart-io/k8s-agent/internal/reasoning/types"
)

// handleHealth handles health check endpoint.
func (s *Server) handleHealth(c *gin.Context) {
	health := HealthResponse{
		Status:  "healthy",
		Service: "reasoning-service-go",
		Components: map[string]bool{
			"analyzer":     true,
			"recommender":  true,
			"llm":          len(s.llmClients) > 0,
			"orchestrator": s.orchestrator != nil,
		},
		Timestamp: time.Now(),
	}

	response.Success(c, health)
}

// handleRootCauseAnalysis handles root cause analysis requests.
func (s *Server) handleRootCauseAnalysis(c *gin.Context) {
	var req types.AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err)
		return
	}

	// Set defaults
	if req.Options.MaxRecommendations == 0 {
		req.Options.MaxRecommendations = s.config.Analysis.MaxRecommendations
	}
	if req.Options.MinConfidence == 0 {
		req.Options.MinConfidence = s.config.Analysis.MinConfidence
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(c.Request.Context(), s.config.GetRequestTimeout())
	defer cancel()

	// Analyze root cause
	result := s.performAnalysis(ctx, &req)

	// Generate recommendations
	if result.Result != nil && result.Result.RootCause != nil {
		s.generateRecommendations(ctx, result, &req)
	}

	result.ProcessingTime = time.Since(start).Seconds()

	response.Success(c, result)
}

// performAnalysis performs the root cause analysis.
func (s *Server) performAnalysis(ctx context.Context, req *types.AnalysisRequest) *types.AnalysisResult {
	result, err := s.analyzer.Analyze(ctx, req)
	if err != nil {
		return &types.AnalysisResult{
			RequestID: req.RequestID,
			Status:    "failed",
			Error:     err.Error(),
		}
	}
	return result
}

// generateRecommendations generates recommendations for the analysis result.
func (s *Server) generateRecommendations(ctx context.Context, result *types.AnalysisResult, req *types.AnalysisRequest) {
	if err := s.recommender.GenerateRecommendations(ctx, result, &req.Context); err != nil {
		s.logRecommendationError(req.RequestID, err)
	}
}

// logRecommendationError logs recommendation generation errors.
func (s *Server) logRecommendationError(requestID string, err error) {
	if s.log != nil {
		s.log.Errorw("Failed to generate recommendations",
			"request_id", requestID,
			"error", err,
		)
	} else {
		fmt.Fprintf(os.Stderr, "Failed to generate recommendations for request %s: %v\n", requestID, err)
	}
}

// handleK8sEventAnalysis handles K8s event analysis requests.
func (s *Server) handleK8sEventAnalysis(c *gin.Context) {
	var req K8sEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err)
		return
	}

	// Validate event data
	if req.Event == nil {
		response.BadRequest(c, "Event data is required", nil)
		return
	}

	// Check if orchestrator is available
	if s.orchestrator == nil {
		response.ServiceUnavailable(c, "Orchestrator not initialized", nil)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), s.config.GetRequestTimeout())
	defer cancel()

	// 转换为 Orchestrator 请求
	orchReq := convertK8sEventToOrchestratorRequest(&req)

	// 调用 Orchestrator
	orchResp, err := s.orchestrator.Analyze(ctx, orchReq)
	if err != nil {
		response.InternalError(c, "Analysis failed", err)
		return
	}

	// 转换 Orchestrator 响应为 K8s Event 响应格式
	analysisResp := convertOrchestratorToK8sEventResponse(orchResp)

	response.Success(c, analysisResp)
}

// handleOrchestratorAnalysis handles analysis requests using the new Orchestrator.
func (s *Server) handleOrchestratorAnalysis(c *gin.Context) {
	// Check if orchestrator is enabled
	if s.orchestrator == nil {
		response.ServiceUnavailable(c, "Orchestrator not enabled", nil)
		return
	}

	var req orchestrator.AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err)
		return
	}

	// Set defaults
	setOrchestratorRequestDefaults(&req)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	// Call orchestrator
	result, err := s.orchestrator.Analyze(ctx, &req)
	if err != nil {
		response.InternalError(c, "Analysis failed", err)
		return
	}

	response.Success(c, result)
}

// setOrchestratorRequestDefaults sets default values for orchestrator request.
func setOrchestratorRequestDefaults(req *orchestrator.AnalysisRequest) {
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}
	if req.Language == "" {
		req.Language = "en"
	}
	if req.DetailLevel == "" {
		req.DetailLevel = "normal"
	}
}
