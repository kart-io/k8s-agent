package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"reasoning-service-go/internal/analyzer"
	"reasoning-service-go/internal/config"
	"reasoning-service-go/internal/recommender"
	"reasoning-service-go/pkg/llm"
	"reasoning-service-go/pkg/types"
)

// Server represents the HTTP API server
type Server struct {
	config      *config.Config
	analyzer    *analyzer.RootCauseAnalyzer
	recommender *recommender.Engine
	llmClients  []llm.Client
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, llmClients []llm.Client) *Server {
	return &Server{
		config:      cfg,
		analyzer:    analyzer.NewRootCauseAnalyzer(cfg, llmClients),
		recommender: recommender.NewEngine(cfg, llmClients),
		llmClients:  llmClients,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", s.handleHealth)

	// Analysis endpoints
	mux.HandleFunc("/api/v1/analyze/root-cause", s.handleRootCauseAnalysis)
	mux.HandleFunc("/api/v1/analyze/k8s-event", s.handleK8sEventAnalysis)

	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	fmt.Printf("Starting Reasoning Service on %s\n", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      s.corsMiddleware(s.loggingMiddleware(mux)),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server.ListenAndServe()
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		fmt.Printf("%s %s %s %v\n", r.Method, r.URL.Path, r.RemoteAddr, duration)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := types.HealthResponse{
		Status:  "healthy",
		Service: "reasoning-service-go",
		Components: map[string]bool{
			"analyzer":    true,
			"recommender": true,
			"llm":         len(s.llmClients) > 0,
		},
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (s *Server) handleRootCauseAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req types.AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
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
	ctx, cancel := context.WithTimeout(r.Context(), s.config.GetRequestTimeout())
	defer cancel()

	// Analyze root cause
	result, err := s.analyzer.Analyze(ctx, &req)
	if err != nil {
		result = &types.AnalysisResult{
			RequestID: req.RequestID,
			Status:    "failed",
			Error:     err.Error(),
		}
	}

	// Generate recommendations
	if result.Result != nil && result.Result.RootCause != nil {
		s.recommender.GenerateRecommendations(ctx, result, &req.Context)
	}

	result.ProcessingTime = time.Since(start).Seconds()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// K8sEventRequest represents a simplified request for K8s event analysis
type K8sEventRequest struct {
	ClusterID string                 `json:"cluster_id,omitempty"`
	Event     map[string]interface{} `json:"event"`
	UseLLM    bool                   `json:"use_llm,omitempty"`
}

// APIResponse 统一的 API 响应格式（与 common/response 保持一致）
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// K8sEventAnalysisResponse represents the response for K8s event analysis
type K8sEventAnalysisResponse struct {
	Analysis        string   `json:"analysis"`
	RootCause       string   `json:"rootCause"`        // 使用驼峰命名与前端一致
	Confidence      float64  `json:"confidence"`
	Recommendations []string `json:"recommendations,omitempty"`
}

func (s *Server) handleK8sEventAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req K8sEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Validate event data
	if req.Event == nil {
		http.Error(w, "Event data is required", http.StatusBadRequest)
		return
	}

	// Build standard analysis request
	analysisReq := &types.AnalysisRequest{
		RequestID:    fmt.Sprintf("k8s-event-%d", time.Now().UnixNano()),
		AnalysisType: "root_cause",
		Context: types.AnalysisContext{
			Event:     req.Event,
			ClusterID: req.ClusterID,
		},
		Options: types.AnalysisOptions{
			UseLLM:             req.UseLLM,
			MinConfidence:      s.config.Analysis.MinConfidence,
			MaxRecommendations: s.config.Analysis.MaxRecommendations,
		},
	}

	// Extract namespace from event if available
	if involvedObj, ok := req.Event["involvedObject"].(map[string]interface{}); ok {
		if namespace, ok := involvedObj["namespace"].(string); ok {
			analysisReq.Context.Namespace = namespace
		}
		if name, ok := involvedObj["name"].(string); ok {
			analysisReq.Context.ResourceName = name
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.config.GetRequestTimeout())
	defer cancel()

	// Analyze
	result, err := s.analyzer.Analyze(ctx, analysisReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Analysis failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate recommendations
	if result.Result != nil && result.Result.RootCause != nil {
		s.recommender.GenerateRecommendations(ctx, result, &analysisReq.Context)
	}

	// Build formatted response
	response := K8sEventAnalysisResponse{
		Analysis: s.formatAnalysis(result),
		Confidence: 0.0,
		Recommendations: []string{},
	}

	if result.Result != nil {
		response.Confidence = result.Result.Confidence

		if result.Result.RootCause != nil {
			response.RootCause = string(result.Result.RootCause.Type)
		}

		if result.Result.Recommendations != nil {
			for _, rec := range result.Result.Recommendations {
				response.Recommendations = append(response.Recommendations, rec.Description)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// 使用标准 APIResponse 格式包装响应
	apiResp := APIResponse{
		Code:    0,
		Message: "success",
		Data:    response,
	}

	// 创建 JSON encoder 并禁用 HTML 转义
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(apiResp)
}

// formatAnalysis formats the analysis result as HTML for frontend display
func (s *Server) formatAnalysis(result *types.AnalysisResult) string {
	if result == nil || result.Result == nil {
		return "<p>无法分析该事件。</p>"
	}

	var html string

	// 诊断结果表格
	if result.Result.RootCause != nil {
		html += `<div class="diagnosis-section">` + "\n"
		html += `<h3>🔍 诊断结果</h3>` + "\n"
		html += `<table class="diagnosis-table">` + "\n"
		html += `<tbody>` + "\n"

		// 问题类型
		html += `<tr>` + "\n"
		html += `<td class="label-cell">问题类型</td>` + "\n"
		html += fmt.Sprintf(`<td class="value-cell"><span class="type-badge">%s</span></td>`, result.Result.RootCause.Type) + "\n"
		html += `</tr>` + "\n"

		// 置信度
		html += `<tr>` + "\n"
		html += `<td class="label-cell">置信度</td>` + "\n"
		html += fmt.Sprintf(`<td class="value-cell"><span class="confidence-badge">%.0f%%</span></td>`, result.Result.Confidence*100) + "\n"
		html += `</tr>` + "\n"

		// 问题描述 - 完整显示，不截断
		if result.Result.RootCause.Description != "" {
			html += `<tr>` + "\n"
			html += `<td class="label-cell">问题描述</td>` + "\n"
			html += fmt.Sprintf(`<td class="value-cell"><div class="problem-desc">%s</div></td>`, result.Result.RootCause.Description) + "\n"
			html += `</tr>` + "\n"
		}

		html += `</tbody>` + "\n"
		html += `</table>` + "\n"
		html += `</div>` + "\n"
	}

	// 证据表格
	if len(result.Result.Evidence) > 0 {
		html += `<div class="evidence-section">` + "\n"
		html += `<h3>📋 证据</h3>` + "\n"
		html += `<table class="evidence-table">` + "\n"
		html += `<tbody>` + "\n"
		for i, evidence := range result.Result.Evidence {
			html += `<tr>` + "\n"
			html += fmt.Sprintf(`<td class="number-cell">%d</td>`, i+1) + "\n"
			html += fmt.Sprintf(`<td class="evidence-cell">%s</td>`, evidence) + "\n"
			html += `</tr>` + "\n"
		}
		html += `</tbody>` + "\n"
		html += `</table>` + "\n"
		html += `</div>` + "\n"
	}

	// 解决方案表格
	if len(result.Result.Recommendations) > 0 {
		html += `<div class="solutions-section">` + "\n"
		html += `<h3>💡 建议解决方案</h3>` + "\n"
		html += `<table class="solutions-table">` + "\n"
		html += `<tbody>` + "\n"
		for i, rec := range result.Result.Recommendations {
			html += `<tr>` + "\n"
			html += fmt.Sprintf(`<td class="number-cell">%d</td>`, i+1) + "\n"
			html += `<td class="solution-cell">` + "\n"
			html += fmt.Sprintf(`<div class="solution-title">%s</div>`, rec.Action) + "\n"
			html += fmt.Sprintf(`<div class="solution-desc">%s</div>`, rec.Description) + "\n"

			// 命令
			if rec.Command != "" {
				html += `<div class="solution-command">` + "\n"
				html += `<div class="command-label">🔧 执行命令:</div>` + "\n"
				html += fmt.Sprintf(`<pre class="command-code">%s</pre>`, rec.Command) + "\n"
				html += `</div>` + "\n"
			}

			// YAML 配置
			if rec.YAML != "" {
				html += `<div class="solution-yaml">` + "\n"
				html += `<div class="yaml-label">📝 YAML 配置:</div>` + "\n"
				html += fmt.Sprintf(`<pre class="yaml-code">%s</pre>`, rec.YAML) + "\n"
				html += `</div>` + "\n"
			}

			// 影响和风险级别
			if rec.Impact != "" || rec.Risk != "" {
				html += `<div class="solution-meta">` + "\n"
				if rec.Impact != "" {
					html += fmt.Sprintf(`<span class="meta-tag impact-tag">影响: %s</span>`, rec.Impact) + "\n"
				}
				if rec.Risk != "" {
					html += fmt.Sprintf(`<span class="meta-tag risk-tag risk-%s">风险: %s</span>`, rec.Risk, rec.Risk) + "\n"
				}
				html += `</div>` + "\n"
			}

			html += `</td>` + "\n"
			html += `</tr>` + "\n"
		}
		html += `</tbody>` + "\n"
		html += `</table>` + "\n"
		html += `</div>` + "\n"
	}

	return html
}

