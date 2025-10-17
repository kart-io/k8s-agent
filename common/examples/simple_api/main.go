package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/common/logger"
	"github.com/kart-io/k8s-agent/common/middleware"
	"github.com/kart-io/k8s-agent/common/pagination"
	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/k8s-agent/common/validator"
)

// Cluster 示例数据结构
type Cluster struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

func main() {
	// 1. 初始化日志（使用 kart-io/logger）
	logConfig := &logger.Config{
		Engine:       "zap",
		Level:        "info",
		Format:       "json",
		OutputPaths:  []string{"stdout"},
		EnableCaller: true,
		Development:  false,
		InitialFields: map[string]interface{}{
			"service": "simple_api_example",
			"version": "v1.0.0",
		},
	}
	if err := logger.Init(logConfig); err != nil {
		panic(err)
	}
	defer logger.Sync()

	// 2. 创建 Gin 引擎
	r := gin.New()

	// 3. 注册中间件
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimitByIP(10, 20)) // 每秒 10 个请求
	r.Use(middleware.Timeout(30 * time.Second))

	// 4. 注册路由
	api := r.Group("/api/v1")
	{
		api.GET("/clusters", listClusters)
		api.GET("/clusters/:id", getCluster)
		api.POST("/clusters", createCluster)
		api.DELETE("/clusters/:id", deleteCluster)
	}

	// 5. 启动服务
	logger.Infow("Starting server", "port", "8080")
	if err := r.Run(":8080"); err != nil {
		logger.Fatalw("Failed to start server", "error", err)
	}
}

// listClusters 获取集群列表（带分页）
func listClusters(c *gin.Context) {
	// 解析分页参数
	params := pagination.Parse(c)

	logger.Infow("Listing clusters",
		"page", params.Page,
		"page_size", params.GetPageSize(),
	)

	// 模拟查询数据库
	clusters := []Cluster{
		{
			ID:          "cluster-1",
			Name:        "生产集群",
			Description: "Production cluster",
			Status:      "healthy",
			CreatedAt:   time.Now().Add(-24 * time.Hour),
		},
		{
			ID:          "cluster-2",
			Name:        "测试集群",
			Description: "Test cluster",
			Status:      "healthy",
			CreatedAt:   time.Now().Add(-12 * time.Hour),
		},
	}

	total := int64(len(clusters))

	// 返回分页结果
	resp := pagination.NewResponse(clusters, total, params)
	response.Success(c, resp)
}

// getCluster 获取单个集群
func getCluster(c *gin.Context) {
	clusterID := c.Param("id")

	// 验证参数
	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	logger.Infow("Getting cluster", "cluster_id", clusterID)

	// 模拟查询数据库
	if clusterID != "cluster-1" && clusterID != "cluster-2" {
		err := errors.ErrClusterNotFound
		response.NotFound(c, "Cluster not found", err)
		return
	}

	cluster := Cluster{
		ID:          clusterID,
		Name:        "生产集群",
		Description: "Production cluster",
		Status:      "healthy",
		CreatedAt:   time.Now().Add(-24 * time.Hour),
	}

	response.Success(c, cluster)
}

// CreateClusterRequest 创建集群请求
type CreateClusterRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// createCluster 创建集群
func createCluster(c *gin.Context) {
	var req CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	// 验证集群名称
	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid cluster name", err)
		return
	}

	logger.Infow("Creating cluster",
		"name", req.Name,
		"description", req.Description,
	)

	// 模拟创建集群
	cluster := Cluster{
		ID:          "cluster-new",
		Name:        req.Name,
		Description: req.Description,
		Status:      "creating",
		CreatedAt:   time.Now(),
	}

	response.SuccessWithMessage(c, "Cluster created successfully", cluster)
}

// deleteCluster 删除集群
func deleteCluster(c *gin.Context) {
	clusterID := c.Param("id")

	// 验证参数
	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	logger.Infow("Deleting cluster", "cluster_id", clusterID)

	// 模拟查询数据库
	if clusterID != "cluster-1" && clusterID != "cluster-2" {
		err := errors.ErrClusterNotFound
		response.NotFound(c, "Cluster not found", err)
		return
	}

	// 模拟删除集群
	response.SuccessWithMessage(c, "Cluster deleted successfully", gin.H{
		"cluster_id": clusterID,
	})
}
