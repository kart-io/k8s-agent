package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kart-io/k8s-agent/internal/reasoning/agents/k8s_tool"
	"github.com/kart-io/k8s-agent/internal/reasoning/agents/reasoning"
	"github.com/kart-io/k8s-agent/internal/reasoning/analyzer"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/description"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/root_cause"
	"github.com/kart-io/k8s-agent/internal/reasoning/config"
	"github.com/kart-io/k8s-agent/internal/reasoning/memory"
	"github.com/kart-io/k8s-agent/internal/reasoning/orchestrator"
	"github.com/kart-io/k8s-agent/internal/reasoning/recommender"
	"github.com/kart-io/k8s-agent/internal/reasoning/llm"
	"github.com/kart-io/k8s-agent/internal/reasoning/llm/proxy"
	"github.com/kart-io/k8s-agent/internal/reasoning/types"
)

// Server represents the HTTP API server
type Server struct {
	config       *config.Config
	analyzer     *analyzer.RootCauseAnalyzer
	recommender  *recommender.Engine
	llmClients   []llm.Client
	orchestrator *orchestrator.Orchestrator // 新增 Orchestrator
}

// NewServer creates a new API server with all required components including Orchestrator
func NewServer(cfg *config.Config, llmClients []llm.Client) *Server {
	// Initialize LLM Proxy
	llmProxy, err := proxy.NewProxyAdapter(&cfg.LLM)
	if err != nil {
		log.Printf("Warning: Failed to initialize LLM Proxy: %v", err)
		llmProxy = nil
	}

	// Initialize Root Cause Chain
	rootCauseChain, err := root_cause.NewRootCauseChain(llmProxy, nil)
	if err != nil {
		log.Printf("Warning: Failed to initialize Root Cause Chain: %v", err)
		rootCauseChain = nil
	}

	// Initialize Description Chain
	descriptionChain, err := description.NewDescriptionChain(llmProxy, nil)
	if err != nil {
		log.Printf("Warning: Failed to initialize Description Chain: %v", err)
		descriptionChain = nil
	}

	// Initialize K8s Tool
	k8sTool, err := k8s_tool.NewK8sTool(nil)
	if err != nil {
		log.Printf("Warning: Failed to initialize K8s Tool: %v", err)
		k8sTool = nil
	}

	// Initialize Reasoning Agent
	reasoningAgent, err := reasoning.NewReasoningAgent(rootCauseChain, descriptionChain, k8sTool, nil)
	if err != nil {
		log.Printf("Warning: Failed to initialize Reasoning Agent: %v", err)
		reasoningAgent = nil
	}

	// Initialize Memory Manager (always create one, even if vector store is disabled)
	// Orchestrator requires Memory Manager if EnableMemory is true in config
	memoryManager, err := memory.NewManager(nil) // Use default config
	if err != nil {
		log.Printf("Warning: Failed to initialize Memory Manager: %v", err)
		// If memory manager fails, disable memory in orchestrator config
		orchestratorConfig := orchestrator.DefaultOrchestratorConfig()
		orchestratorConfig.EnableMemory = false
		orch, err := orchestrator.NewOrchestrator(
			reasoningAgent,
			rootCauseChain,
			descriptionChain,
			k8sTool,
			nil, // No memory manager
			orchestratorConfig,
		)
		if err != nil {
			log.Printf("Error: Failed to initialize Orchestrator: %v", err)
			orch = nil
		} else {
			log.Printf("Successfully initialized Orchestrator (memory disabled)")
		}

		return &Server{
			config:       cfg,
			analyzer:     analyzer.NewRootCauseAnalyzer(cfg, llmClients),
			recommender:  recommender.NewEngine(cfg, llmClients),
			llmClients:   llmClients,
			orchestrator: orch,
		}
	}

	// Initialize Orchestrator with Memory Manager
	orchestratorConfig := orchestrator.DefaultOrchestratorConfig()
	orch, err := orchestrator.NewOrchestrator(
		reasoningAgent,
		rootCauseChain,
		descriptionChain,
		k8sTool,
		memoryManager,
		orchestratorConfig,
	)
	if err != nil {
		log.Printf("Error: Failed to initialize Orchestrator: %v", err)
		orch = nil
	} else {
		log.Printf("Successfully initialized Orchestrator")
	}

	return &Server{
		config:       cfg,
		analyzer:     analyzer.NewRootCauseAnalyzer(cfg, llmClients),
		recommender:  recommender.NewEngine(cfg, llmClients),
		llmClients:   llmClients,
		orchestrator: orch,
	}
}

// NewServerWithOrchestrator creates a new API server with Orchestrator
func NewServerWithOrchestrator(cfg *config.Config, llmClients []llm.Client, orch *orchestrator.Orchestrator) *Server {
	return &Server{
		config:       cfg,
		analyzer:     analyzer.NewRootCauseAnalyzer(cfg, llmClients),
		recommender:  recommender.NewEngine(cfg, llmClients),
		llmClients:   llmClients,
		orchestrator: orch,
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

	// New orchestrator endpoint
	mux.HandleFunc("/api/v1/orchestrator/analyze", s.handleOrchestratorAnalysis)

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
			"analyzer":     true,
			"recommender":  true,
			"llm":          len(s.llmClients) > 0,
			"orchestrator": s.orchestrator != nil,
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

// convertK8sEventToOrchestratorRequest 将 K8s Event 请求转换为 Orchestrator 请求
func (s *Server) convertK8sEventToOrchestratorRequest(req *K8sEventRequest) *orchestrator.AnalysisRequest {
	orchReq := &orchestrator.AnalysisRequest{
		ClusterID:   req.ClusterID,
		Timestamp:   time.Now(),
		Language:    "zh-CN", // K8s Event API 默认中文
		DetailLevel: "normal",
	}

	// 从 event 中提取信息
	if req.Event != nil {
		// 提取 reason (作为 failure_type 和 error_message)
		if reason, ok := req.Event["reason"].(string); ok {
			orchReq.FailureType = reason
			orchReq.ErrorMessage = reason
		}

		// 提取 message
		if message, ok := req.Event["message"].(string); ok {
			if orchReq.ErrorMessage == "" {
				orchReq.ErrorMessage = message
			}
		}

		// 提取 involvedObject 信息
		if involvedObj, ok := req.Event["involvedObject"].(map[string]interface{}); ok {
			if namespace, ok := involvedObj["namespace"].(string); ok {
				orchReq.Namespace = namespace
			}
			if name, ok := involvedObj["name"].(string); ok {
				orchReq.ResourceName = name
			}
			if kind, ok := involvedObj["kind"].(string); ok {
				orchReq.ResourceType = strings.ToLower(kind)
			}
		}

		// 构建事件信息
		eventInfo := k8s_tool.EventInfo{
			LastTimestamp: time.Now(),
		}

		if eventType, ok := req.Event["type"].(string); ok {
			eventInfo.Type = eventType
		}
		if reason, ok := req.Event["reason"].(string); ok {
			eventInfo.Reason = reason
		}
		if message, ok := req.Event["message"].(string); ok {
			eventInfo.Message = message
		}
		if source, ok := req.Event["source"].(map[string]interface{}); ok {
			if component, ok := source["component"].(string); ok {
				eventInfo.Source = component
			}
		}

		orchReq.Events = []k8s_tool.EventInfo{eventInfo}
	}

	// 设置默认值
	if orchReq.FailureType == "" {
		orchReq.FailureType = "unknown"
	}
	if orchReq.ResourceType == "" {
		orchReq.ResourceType = "pod"
	}
	if orchReq.ResourceName == "" {
		orchReq.ResourceName = "unknown"
	}
	if orchReq.Namespace == "" {
		orchReq.Namespace = "default"
	}

	return orchReq
}

// convertOrchestratorToK8sEventResponse 将 Orchestrator 响应转换为 K8s Event 响应
func (s *Server) convertOrchestratorToK8sEventResponse(orchResp *orchestrator.AnalysisResponse) K8sEventAnalysisResponse {
	response := K8sEventAnalysisResponse{
		Confidence:      0.0,
		Recommendations: []string{},
	}

	// 转换根因分析
	if orchResp.RootCause != nil {
		response.RootCause = orchResp.RootCause.RootCause
		response.Confidence = orchResp.RootCause.Confidence

		// 从根因推荐中提取建议描述
		if len(orchResp.RootCause.Recommendations) > 0 {
			for _, rec := range orchResp.RootCause.Recommendations {
				response.Recommendations = append(response.Recommendations, rec.Description)
			}
		}
	}

	// 生成 HTML 格式的分析内容
	response.Analysis = s.formatOrchestratorAnalysis(orchResp)

	return response
}

// formatOrchestratorAnalysis 将 Orchestrator 响应格式化为 HTML（兼容旧格式）
func (s *Server) formatOrchestratorAnalysis(orchResp *orchestrator.AnalysisResponse) string {
	var html string

	// 诊断结果表格
	if orchResp.RootCause != nil {
		html += `<div class="diagnosis-section">` + "\n"
		html += `<h3>🔍 诊断结果</h3>` + "\n"
		html += `<table class="diagnosis-table">` + "\n"
		html += `<tbody>` + "\n"

		// 问题类型
		html += `<tr>` + "\n"
		html += `<td class="label-cell">问题类型</td>` + "\n"
		html += fmt.Sprintf(`<td class="value-cell"><span class="type-badge">%s</span></td>`, orchResp.RootCause.Category) + "\n"
		html += `</tr>` + "\n"

		// 置信度
		html += `<tr>` + "\n"
		html += `<td class="label-cell">置信度</td>` + "\n"
		html += fmt.Sprintf(`<td class="value-cell"><span class="confidence-badge">%.0f%%</span></td>`, orchResp.RootCause.Confidence*100) + "\n"
		html += `</tr>` + "\n"

		// 问题描述
		html += `<tr>` + "\n"
		html += `<td class="label-cell">问题描述</td>` + "\n"
		html += fmt.Sprintf(`<td class="value-cell"><div class="problem-desc">%s</div></td>`, orchResp.RootCause.RootCause) + "\n"
		html += `</tr>` + "\n"

		// 推理过程
		if orchResp.RootCause.Reasoning != "" {
			html += `<tr>` + "\n"
			html += `<td class="label-cell">分析推理</td>` + "\n"
			html += fmt.Sprintf(`<td class="value-cell"><div class="reasoning">%s</div></td>`, orchResp.RootCause.Reasoning) + "\n"
			html += `</tr>` + "\n"
		}

		html += `</tbody>` + "\n"
		html += `</table>` + "\n"
		html += `</div>` + "\n"
	}

	// 故障描述（如果有）
	if orchResp.Description != nil {
		html += `<div class="description-section">` + "\n"
		html += `<h3>📝 故障描述</h3>` + "\n"
		html += `<table class="description-table">` + "\n"
		html += `<tbody>` + "\n"

		html += `<tr>` + "\n"
		html += `<td class="label-cell">标题</td>` + "\n"
		html += fmt.Sprintf(`<td class="value-cell">%s</td>`, orchResp.Description.Title) + "\n"
		html += `</tr>` + "\n"

		html += `<tr>` + "\n"
		html += `<td class="label-cell">摘要</td>` + "\n"
		html += fmt.Sprintf(`<td class="value-cell">%s</td>`, orchResp.Description.Summary) + "\n"
		html += `</tr>` + "\n"

		// 影响组件
		if len(orchResp.Description.AffectedComponents) > 0 {
			html += `<tr>` + "\n"
			html += `<td class="label-cell">影响组件</td>` + "\n"
			html += `<td class="value-cell">` + strings.Join(orchResp.Description.AffectedComponents, ", ") + `</td>` + "\n"
			html += `</tr>` + "\n"
		}

		html += `</tbody>` + "\n"
		html += `</table>` + "\n"
		html += `</div>` + "\n"
	}

	// 建议解决方案
	if orchResp.RootCause != nil && len(orchResp.RootCause.Recommendations) > 0 {
		html += `<div class="solutions-section">` + "\n"
		html += `<h3>💡 建议解决方案</h3>` + "\n"
		html += `<table class="solutions-table">` + "\n"
		html += `<tbody>` + "\n"

		for i, rec := range orchResp.RootCause.Recommendations {
			html += `<tr>` + "\n"
			html += fmt.Sprintf(`<td class="number-cell">%d</td>`, i+1) + "\n"
			html += `<td class="solution-cell">` + "\n"
			html += fmt.Sprintf(`<div class="solution-title">%s</div>`, rec.Action) + "\n"
			html += fmt.Sprintf(`<div class="solution-desc">%s</div>`, rec.Description) + "\n"

			// 命令
			if len(rec.Commands) > 0 {
				html += `<div class="solution-command">` + "\n"
				html += `<div class="command-label">🔧 执行命令:</div>` + "\n"
				html += `<pre class="command-code">` + strings.Join(rec.Commands, "\n") + `</pre>` + "\n"
				html += `</div>` + "\n"
			}

			// 影响和风险级别
			if rec.Impact != "" || rec.RiskLevel != "" {
				html += `<div class="solution-meta">` + "\n"
				if rec.Impact != "" {
					html += fmt.Sprintf(`<span class="meta-tag impact-tag">影响: %s</span>`, rec.Impact) + "\n"
				}
				if rec.RiskLevel != "" {
					html += fmt.Sprintf(`<span class="meta-tag risk-tag risk-%s">风险: %s</span>`, rec.RiskLevel, rec.RiskLevel) + "\n"
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

	// 相似案例（新增，旧 API 没有）
	if len(orchResp.SimilarCases) > 0 {
		html += `<div class="similar-cases-section">` + "\n"
		html += `<h3>📚 相似案例</h3>` + "\n"
		html += `<table class="similar-cases-table">` + "\n"
		html += `<tbody>` + "\n"

		for i, sc := range orchResp.SimilarCases {
			html += `<tr>` + "\n"
			html += fmt.Sprintf(`<td class="number-cell">%d</td>`, i+1) + "\n"
			html += `<td class="case-cell">` + "\n"
			html += fmt.Sprintf(`<div class="case-desc">%s</div>`, sc.Description) + "\n"
			html += fmt.Sprintf(`<div class="case-solution">解决方案: %s</div>`, sc.Solution) + "\n"
			html += fmt.Sprintf(`<div class="case-similarity">相似度: %.0f%%</div>`, sc.Similarity*100) + "\n"
			html += `</td>` + "\n"
			html += `</tr>` + "\n"
		}

		html += `</tbody>` + "\n"
		html += `</table>` + "\n"
		html += `</div>` + "\n"
	}

	if html == "" {
		return "<p>无法分析该事件。</p>"
	}

	return html
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
	RootCause       string   `json:"rootCause"` // 使用驼峰命名与前端一致
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

	// Check if orchestrator is available
	if s.orchestrator == nil {
		http.Error(w, "Orchestrator not initialized", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.config.GetRequestTimeout())
	defer cancel()

	// 转换为 Orchestrator 请求
	orchReq := s.convertK8sEventToOrchestratorRequest(&req)

	// 调用 Orchestrator
	orchResp, err := s.orchestrator.Analyze(ctx, orchReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Analysis failed: %v", err), http.StatusInternalServerError)
		return
	}

	// 转换 Orchestrator 响应为 K8s Event 响应格式
	response := s.convertOrchestratorToK8sEventResponse(orchResp)

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

// handleOrchestratorAnalysis handles analysis requests using the new Orchestrator
func (s *Server) handleOrchestratorAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if orchestrator is enabled
	if s.orchestrator == nil {
		http.Error(w, "Orchestrator not enabled", http.StatusServiceUnavailable)
		return
	}

	var req orchestrator.AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}
	if req.Language == "" {
		req.Language = "en"
	}
	if req.DetailLevel == "" {
		req.DetailLevel = "normal"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// Call orchestrator
	result, err := s.orchestrator.Analyze(ctx, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Analysis failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(result)
}
