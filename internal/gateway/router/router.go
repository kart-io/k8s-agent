package router

import (
	"github.com/gin-gonic/gin"
	"github.com/kart-io/logger/core"
	"github.com/spf13/viper"

	"github.com/kart-io/k8s-agent/internal/gateway/handler"
	"github.com/kart-io/k8s-agent/internal/gateway/middleware"
	"github.com/kart-io/k8s-agent/internal/gateway/proxy"
	"github.com/kart-io/k8s-agent/internal/gateway/types"
)

// Setup 设置路由
func Setup(logger core.Logger) *gin.Engine {
	// 设置 Gin 模式
	mode := viper.GetString("server.mode")
	gin.SetMode(mode)

	router := gin.New()

	// 全局中间件
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit())

	// 创建代理处理器
	proxyHandler := proxy.NewProxy(logger)

	// 健康检查处理器
	healthHandler := handler.NewHealthHandler(proxyHandler, logger)

	// 网关自身的健康检查
	router.GET("/health", healthHandler.GatewayHealth)

	// API 路由组
	api := router.Group("/api/v1")
	{
		// 认证服务路由 (不需要认证)
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", proxyHandler.HandleRequest("auth"))
			authGroup.POST("/logout", proxyHandler.HandleRequest("auth"))
			authGroup.GET("/me", proxyHandler.HandleRequest("auth"))
			authGroup.GET("/menus", proxyHandler.HandleRequest("auth"))
			authGroup.POST("/check", proxyHandler.HandleRequest("auth"))
		}

		// 需要认证的路由
		authenticated := api.Group("")
		authenticated.Use(middleware.JWTAuth())
		{
			// Agent 管理路由
			agentGroup := authenticated.Group("/agents")
			{
				agentGroup.GET("", proxyHandler.HandleRequest("agent_manager"))
				agentGroup.GET("/:id", proxyHandler.HandleRequest("agent_manager"))
				agentGroup.POST("", proxyHandler.HandleRequest("agent_manager"))
				agentGroup.PUT("/:id", proxyHandler.HandleRequest("agent_manager"))
				agentGroup.DELETE("/:id", proxyHandler.HandleRequest("agent_manager"))
			}

			// 集群管理路由
			clusterGroup := authenticated.Group("/clusters")
			{
				clusterGroup.GET("", proxyHandler.HandleRequest("agent_manager"))
				clusterGroup.GET("/:id", proxyHandler.HandleRequest("agent_manager"))
				clusterGroup.POST("", proxyHandler.HandleRequest("agent_manager"))
				clusterGroup.PUT("/:id", proxyHandler.HandleRequest("agent_manager"))
				clusterGroup.DELETE("/:id", proxyHandler.HandleRequest("agent_manager"))
			}

			// 事件管理路由
			eventGroup := authenticated.Group("/events")
			{
				eventGroup.GET("", proxyHandler.HandleRequest("agent_manager"))
				eventGroup.GET("/:id", proxyHandler.HandleRequest("agent_manager"))
			}

			// 命令管理路由
			commandGroup := authenticated.Group("/commands")
			{
				commandGroup.GET("", proxyHandler.HandleRequest("agent_manager"))
				commandGroup.GET("/:id", proxyHandler.HandleRequest("agent_manager"))
				commandGroup.POST("", proxyHandler.HandleRequest("agent_manager"))
			}

			// 工作流路由
			workflowGroup := authenticated.Group("/workflows")
			{
				workflowGroup.GET("", proxyHandler.HandleRequest("orchestrator"))
				workflowGroup.GET("/:id", proxyHandler.HandleRequest("orchestrator"))
				workflowGroup.POST("", proxyHandler.HandleRequest("orchestrator"))
				workflowGroup.PUT("/:id", proxyHandler.HandleRequest("orchestrator"))
				workflowGroup.DELETE("/:id", proxyHandler.HandleRequest("orchestrator"))
			}

			// 策略路由
			strategyGroup := authenticated.Group("/strategies")
			{
				strategyGroup.GET("", proxyHandler.HandleRequest("orchestrator"))
				strategyGroup.GET("/:id", proxyHandler.HandleRequest("orchestrator"))
				strategyGroup.POST("", proxyHandler.HandleRequest("orchestrator"))
				strategyGroup.PUT("/:id", proxyHandler.HandleRequest("orchestrator"))
				strategyGroup.DELETE("/:id", proxyHandler.HandleRequest("orchestrator"))
			}

			// 推理服务路由
			reasoningGroup := authenticated.Group("/reasoning")
			{
				reasoningGroup.POST("/analyze", proxyHandler.HandleRequest("reasoning"))
				reasoningGroup.POST("/suggest", proxyHandler.HandleRequest("reasoning"))
			}

			// 用户管理路由
			userGroup := authenticated.Group("/users")
			{
				userGroup.GET("", proxyHandler.HandleRequest("auth"))
				userGroup.GET("/:id", proxyHandler.HandleRequest("auth"))
				userGroup.POST("", proxyHandler.HandleRequest("auth"))
				userGroup.PUT("/:id", proxyHandler.HandleRequest("auth"))
				userGroup.DELETE("/:id", proxyHandler.HandleRequest("auth"))
			}

			// 角色管理路由
			roleGroup := authenticated.Group("/roles")
			{
				roleGroup.GET("", proxyHandler.HandleRequest("auth"))
				roleGroup.GET("/:id", proxyHandler.HandleRequest("auth"))
				roleGroup.POST("", proxyHandler.HandleRequest("auth"))
				roleGroup.PUT("/:id", proxyHandler.HandleRequest("auth"))
				roleGroup.DELETE("/:id", proxyHandler.HandleRequest("auth"))
				roleGroup.POST("/:id/permissions", proxyHandler.HandleRequest("auth"))
			}

			// 权限管理路由
			permissionGroup := authenticated.Group("/permissions")
			{
				permissionGroup.GET("", proxyHandler.HandleRequest("auth"))
				permissionGroup.GET("/tree", proxyHandler.HandleRequest("auth"))
				permissionGroup.GET("/:id", proxyHandler.HandleRequest("auth"))
				permissionGroup.POST("", proxyHandler.HandleRequest("auth"))
				permissionGroup.PUT("/:id", proxyHandler.HandleRequest("auth"))
				permissionGroup.DELETE("/:id", proxyHandler.HandleRequest("auth"))
			}
		}

		// 服务健康检查
		api.GET("/health/services", healthHandler.ServicesHealth)
		api.GET("/health/services/:service", healthHandler.ServiceHealth)
	}

	// 监控指标
	if viper.GetBool("metrics.enabled") {
		router.GET(viper.GetString("metrics.path"), handler.MetricsHandler)
	}

	// 加载自定义路由配置
	loadCustomRoutes(router, proxyHandler)

	return router
}

// loadCustomRoutes 加载自定义路由配置
func loadCustomRoutes(router *gin.Engine, proxyHandler *proxy.Proxy) {
	var routes []types.RouteConfig
	if err := viper.UnmarshalKey("routes", &routes); err != nil {
		return
	}

	for _, route := range routes {
		// 根据是否需要认证添加中间件
		if route.AuthRequired {
			router.Any(route.Path, middleware.JWTAuth(), proxyHandler.HandleRequest(route.Service))
		} else {
			router.Any(route.Path, proxyHandler.HandleRequest(route.Service))
		}
	}
}
