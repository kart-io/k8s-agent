package api

import (
	"context"
	"log"
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
	// Initialize orchestrator components
	orch := initializeOrchestrator(cfg, logger)

	// Create server instance
	return createServer(cfg, llmClients, orch, logger)
}

// NewServerWithOrchestrator creates a new API server with Orchestrator.
func NewServerWithOrchestrator(cfg *config.Config, llmClients []llm.Client, orch *orchestrator.Orchestrator, logger core.Logger) *Server {
	return createServer(cfg, llmClients, orch, logger)
}

// initializeOrchestrator initializes the orchestrator with all its dependencies.
func initializeOrchestrator(cfg *config.Config, logger core.Logger) *orchestrator.Orchestrator {
	// Initialize LLM Proxy
	llmProxy, err := proxy.NewProxyAdapter(&cfg.LLM)
	if err != nil {
		log.Printf("Warning: Failed to initialize LLM Proxy: %v", err)
		return nil
	}

	// Initialize Root Cause Chain
	rootCauseChain, err := root_cause.NewRootCauseChain(llmProxy, nil)
	if err != nil {
		log.Printf("Warning: Failed to initialize Root Cause Chain: %v", err)
		return nil
	}

	// Initialize Description Chain
	descriptionChain, err := description.NewDescriptionChain(llmProxy, nil)
	if err != nil {
		log.Printf("Warning: Failed to initialize Description Chain: %v", err)
		return nil
	}

	// Initialize K8s Tool
	k8sTool, err := k8s_tool.NewK8sTool(nil)
	if err != nil {
		log.Printf("Warning: Failed to initialize K8s Tool: %v", err)
		return nil
	}

	// Initialize Reasoning Agent
	reasoningAgent, err := reasoning.NewReasoningAgent(rootCauseChain, descriptionChain, k8sTool, nil)
	if err != nil {
		log.Printf("Warning: Failed to initialize Reasoning Agent: %v", err)
		return nil
	}

	// Initialize Memory Manager
	memoryManager, orch := initializeOrchestratorWithMemory(reasoningAgent, rootCauseChain, descriptionChain, k8sTool)

	if orch != nil {
		log.Printf("Successfully initialized Orchestrator (memory: %v)", memoryManager != nil)
	}

	return orch
}

// initializeOrchestratorWithMemory initializes orchestrator with or without memory manager.
func initializeOrchestratorWithMemory(
	reasoningAgent *reasoning.ReasoningAgent,
	rootCauseChain *root_cause.RootCauseChain,
	descriptionChain *description.DescriptionChain,
	k8sTool *k8s_tool.K8sTool,
) (memory.Manager, *orchestrator.Orchestrator) {
	// Try to initialize Memory Manager
	memoryManager, err := memory.NewManager(nil)
	if err != nil {
		log.Printf("Warning: Failed to initialize Memory Manager: %v", err)
		// Create orchestrator without memory
		return nil, createOrchestratorWithoutMemory(reasoningAgent, rootCauseChain, descriptionChain, k8sTool)
	}

	// Create orchestrator with memory
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
		// Fallback to orchestrator without memory
		return nil, createOrchestratorWithoutMemory(reasoningAgent, rootCauseChain, descriptionChain, k8sTool)
	}

	return memoryManager, orch
}

// createOrchestratorWithoutMemory creates an orchestrator without memory support.
func createOrchestratorWithoutMemory(
	reasoningAgent *reasoning.ReasoningAgent,
	rootCauseChain *root_cause.RootCauseChain,
	descriptionChain *description.DescriptionChain,
	k8sTool *k8s_tool.K8sTool,
) *orchestrator.Orchestrator {
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
		return nil
	}

	return orch
}

// createServer creates the server instance with all components.
func createServer(cfg *config.Config, llmClients []llm.Client, orch *orchestrator.Orchestrator, logger core.Logger) *Server {
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
	ginConfig := httpserver.NewGinServerOptions(serverOpts)

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
