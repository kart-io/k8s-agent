package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	commonoptions "github.com/kart-io/k8s-agent/common/options"
	commonserver "github.com/kart-io/k8s-agent/common/server"
	httpserver "github.com/kart-io/k8s-agent/common/server/http"
	"github.com/kart-io/k8s-agent/internal/reasoning/agents/k8s_tool"
	"github.com/kart-io/k8s-agent/internal/reasoning/agents/reasoning"
	"github.com/kart-io/k8s-agent/internal/reasoning/analyzer"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/description"
	"github.com/kart-io/k8s-agent/internal/reasoning/chains/root_cause"
	"github.com/kart-io/k8s-agent/internal/reasoning/config"
	"github.com/kart-io/k8s-agent/internal/reasoning/llm"
	"github.com/kart-io/k8s-agent/internal/reasoning/llm/proxy"
	"github.com/kart-io/k8s-agent/internal/reasoning/memory"
	"github.com/kart-io/k8s-agent/internal/reasoning/orchestrator"
	"github.com/kart-io/k8s-agent/internal/reasoning/recommender"
	"github.com/kart-io/k8s-agent/internal/reasoning/types"
	"github.com/kart-io/logger/core"
)

// Server represents the HTTP API server using common/server framework.
type Server struct {
	config       *config.Config
	analyzer     *analyzer.RootCauseAnalyzer
	recommender  *recommender.Engine
	llmClients   []llm.Client
	orchestrator *orchestrator.Orchestrator
	ginServer    commonserver.Server // 使用 common/server 的 Server 接口
	log          core.Logger
}

// NewServer creates a new API server with all required components including Orchestrator.
func NewServer(cfg *config.Config, llmClients []llm.Client, logger core.Logger) *Server {
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

		// 创建 common/server 的配置
		serverOpts := &commonoptions.ServerOptions{
			Host:         cfg.Server.Host,
			Port:         cfg.Server.Port,
			Mode:         "release",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		}

		// 创建 Gin 服务器配置
		ginConfig := httpserver.NewGinServerConfig(serverOpts)

		// 创建 Gin 服务器
		ginServer := httpserver.NewGinServerFromFullConfig(logger, ginConfig)

		s := &Server{
			config:       cfg,
			analyzer:     analyzer.NewRootCauseAnalyzer(cfg, llmClients),
			recommender:  recommender.NewEngine(cfg, llmClients),
			llmClients:   llmClients,
			orchestrator: orch,
			ginServer:    ginServer,
			log:          logger,
		}

		// 设置路由
		s.setupRoutes(ginServer.GetEngine())

		return s
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

	// 创建 common/server 的配置
	serverOpts := &commonoptions.ServerOptions{
		Host:         cfg.Server.Host,
		Port:         cfg.Server.Port,
		Mode:         "release",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 创建 Gin 服务器配置
	ginConfig := httpserver.NewGinServerConfig(serverOpts)

	// 创建 Gin 服务器
	ginServer := httpserver.NewGinServerFromFullConfig(logger, ginConfig)

	s := &Server{
		config:       cfg,
		analyzer:     analyzer.NewRootCauseAnalyzer(cfg, llmClients),
		recommender:  recommender.NewEngine(cfg, llmClients),
		llmClients:   llmClients,
		orchestrator: orch,
		ginServer:    ginServer,
		log:          logger,
	}

	// 设置路由
	s.setupRoutes(ginServer.GetEngine())

	return s
}

// NewServerWithOrchestrator creates a new API server with Orchestrator.
func NewServerWithOrchestrator(cfg *config.Config, llmClients []llm.Client, orch *orchestrator.Orchestrator, logger core.Logger) *Server {
	// 创建 common/server 的配置
	serverOpts := &commonoptions.ServerOptions{
		Host:         cfg.Server.Host,
		Port:         cfg.Server.Port,
		Mode:         "release",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 创建 Gin 服务器配置
	ginConfig := httpserver.NewGinServerConfig(serverOpts)

	// 创建 Gin 服务器
	ginServer := httpserver.NewGinServerFromFullConfig(logger, ginConfig)

	s := &Server{
		config:       cfg,
		analyzer:     analyzer.NewRootCauseAnalyzer(cfg, llmClients),
		recommender:  recommender.NewEngine(cfg, llmClients),
		llmClients:   llmClients,
		orchestrator: orch,
		ginServer:    ginServer,
		log:          logger,
	}

	// 设置路由
	s.setupRoutes(ginServer.GetEngine())

	return s
}

// setupRoutes 设置所有路由
func (s *Server) setupRoutes(engine *gin.Engine) {
	// 设置中间件
	engine.Use(s.corsMiddleware())
	engine.Use(s.loggingMiddleware())

	// Health check
	engine.GET("/health", s.handleHealth)

	// API v1
	v1 := engine.Group("/api/v1")
	{
		// Analysis endpoints
		analyze := v1.Group("/analyze")
		{
			analyze.POST("/root-cause", s.handleRootCauseAnalysis)
			analyze.POST("/k8s-event", s.handleK8sEventAnalysis)
		}

		// Orchestrator endpoint
		orchestratorGroup := v1.Group("/orchestrator")
		{
			orchestratorGroup.POST("/analyze", s.handleOrchestratorAnalysis)
		}
	}

	if s.log != nil {
		s.log.Infow("Reasoning service routes configured",
			"port", s.config.Server.Port,
			"orchestrator_enabled", s.orchestrator != nil,
			"llm_count", len(s.llmClients),
		)
	}
}

// Start starts the HTTP server using common/server framework.
// Note: This method is for backward compatibility
func (s *Server) Start() error {
	return s.Run(context.Background())
}

// Run runs the server using common/server framework
func (s *Server) Run(ctx context.Context) error {
	if s.log != nil {
		s.log.Infow("Starting reasoning service with common/server framework",
			"host", s.config.Server.Host,
			"port", s.config.Server.Port,
		)
	}

	// 使用 common/server 的 Serve 方法来管理生命周期
	return commonserver.Serve(ctx, s.ginServer, s.log)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.log != nil {
		s.log.Info("Reasoning service shutting down")
	}

	// common/server 的 Serve 方法会自动处理优雅关闭
	return nil
}

// GetServer returns the underlying common/server.Server instance
func (s *Server) GetServer() commonserver.Server {
	return s.ginServer
}

// GetEngine returns the Gin Engine instance (for testing or other needs)
func (s *Server) GetEngine() *gin.Engine {
	if ginSrv, ok := s.ginServer.(*httpserver.GinServer); ok {
		return ginSrv.GetEngine()
	}
	return nil
}

// corsMiddleware returns a Gin CORS middleware
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// loggingMiddleware returns a Gin logging middleware
func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		if s.log != nil {
			s.log.Infow("Request processed",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"remote_addr", c.ClientIP(),
				"duration", duration,
				"status", c.Writer.Status(),
			)
		} else {
			fmt.Printf("%s %s %s %v %d\n",
				c.Request.Method,
				c.Request.URL.Path,
				c.ClientIP(),
				duration,
				c.Writer.Status(),
			)
		}
	}
}

// handleHealth handles health check endpoint
func (s *Server) handleHealth(c *gin.Context) {
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

	c.JSON(http.StatusOK, health)
}

// handleRootCauseAnalysis handles root cause analysis requests
func (s *Server) handleRootCauseAnalysis(c *gin.Context) {
	var req types.AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request: %v", err),
		})
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
		if err := s.recommender.GenerateRecommendations(ctx, result, &req.Context); err != nil {
			// Log error
			if s.log != nil {
				s.log.Errorw("Failed to generate recommendations",
					"request_id", req.RequestID,
					"error", err,
				)
			} else {
				fmt.Fprintf(os.Stderr, "Failed to generate recommendations for request %s: %v\n", req.RequestID, err)
			}
			// Continue without recommendations
		}
	}

	result.ProcessingTime = time.Since(start).Seconds()

	c.JSON(http.StatusOK, result)
}

// K8sEventRequest represents a simplified request for K8s event analysis.
type K8sEventRequest struct {
	ClusterID string                 `json:"cluster_id,omitempty"`
	Event     map[string]interface{} `json:"event"`
	UseLLM    bool                   `json:"use_llm,omitempty"`
}

// convertK8sEventToOrchestratorRequest 将 K8s Event 请求转换为 Orchestrator 请求.
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

// convertOrchestratorToK8sEventResponse 将 Orchestrator 响应转换为 K8s Event 响应.
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

// formatOrchestratorAnalysis 将 Orchestrator 响应格式化为 HTML（兼容旧格式）.
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

// APIResponse 统一的 API 响应格式（与 common/response 保持一致）.
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// K8sEventAnalysisResponse represents the response for K8s event analysis.
type K8sEventAnalysisResponse struct {
	Analysis        string   `json:"analysis"`
	RootCause       string   `json:"rootCause"` // 使用驼峰命名与前端一致
	Confidence      float64  `json:"confidence"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// handleK8sEventAnalysis handles K8s event analysis requests
func (s *Server) handleK8sEventAnalysis(c *gin.Context) {
	var req K8sEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	// Validate event data
	if req.Event == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Event data is required",
		})
		return
	}

	// Check if orchestrator is available
	if s.orchestrator == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Orchestrator not initialized",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), s.config.GetRequestTimeout())
	defer cancel()

	// 转换为 Orchestrator 请求
	orchReq := s.convertK8sEventToOrchestratorRequest(&req)

	// 调用 Orchestrator
	orchResp, err := s.orchestrator.Analyze(ctx, orchReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Analysis failed: %v", err),
		})
		return
	}

	// 转换 Orchestrator 响应为 K8s Event 响应格式
	response := s.convertOrchestratorToK8sEventResponse(orchResp)

	// 使用标准 APIResponse 格式包装响应
	apiResp := APIResponse{
		Code:    0,
		Message: "success",
		Data:    response,
	}

	c.JSON(http.StatusOK, apiResp)
}

// handleOrchestratorAnalysis handles analysis requests using the new Orchestrator.
func (s *Server) handleOrchestratorAnalysis(c *gin.Context) {
	// Check if orchestrator is enabled
	if s.orchestrator == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Orchestrator not enabled",
		})
		return
	}

	var req orchestrator.AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request: %v", err),
		})
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

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	// Call orchestrator
	result, err := s.orchestrator.Analyze(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Analysis failed: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
