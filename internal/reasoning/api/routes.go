package api

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/middleware"
)

// setupRoutes 设置所有路由.
func (s *Server) setupRoutes(engine *gin.Engine) {
	// 设置中间件
	s.setupMiddleware(engine)

	// Health check
	engine.GET("/health", s.handleHealth)

	// API v1
	v1 := engine.Group("/api/v1")
	{
		// Analysis endpoints
		s.setupAnalysisRoutes(v1)

		// Orchestrator endpoint
		s.setupOrchestratorRoutes(v1)
	}

	s.logRouteSetup()
}

// setupMiddleware 设置中间件.
func (s *Server) setupMiddleware(engine *gin.Engine) {
	// Use common CORS middleware
	engine.Use(middleware.CORS())

	// Use common request logging middleware
	engine.Use(middleware.RequestLoggerWithLogger(s.log))
}

// setupAnalysisRoutes 设置分析相关路由.
func (s *Server) setupAnalysisRoutes(v1 *gin.RouterGroup) {
	analyze := v1.Group("/analyze")
	{
		analyze.POST("/root-cause", s.handleRootCauseAnalysis)
		analyze.POST("/k8s-event", s.handleK8sEventAnalysis)
	}
}

// setupOrchestratorRoutes 设置 Orchestrator 相关路由.
func (s *Server) setupOrchestratorRoutes(v1 *gin.RouterGroup) {
	orchestratorGroup := v1.Group("/orchestrator")
	{
		orchestratorGroup.POST("/analyze", s.handleOrchestratorAnalysis)
	}
}

// logRouteSetup 记录路由设置完成.
func (s *Server) logRouteSetup() {
	if s.log != nil {
		s.log.Infow("Reasoning service routes configured",
			"port", s.config.Server.Port,
			"orchestrator_enabled", s.orchestrator != nil,
			"llm_count", len(s.llmClients),
		)
	}
}
