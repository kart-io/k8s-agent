package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/response"
	"github.com/kart-io/version"
)

// VersionHandler 版本信息处理器.
type VersionHandler struct{}

// NewVersionHandler 创建版本信息处理器.
func NewVersionHandler() *VersionHandler {
	return &VersionHandler{}
}

// GetVersion 获取版本信息
// GET /version.
func (h *VersionHandler) GetVersion(c *gin.Context) {
	info := version.Get()

	// 返回版本信息
	response.Success(c, gin.H{
		"service_name":   info.ServiceName,
		"git_version":    info.GitVersion,
		"git_commit":     info.GitCommit,
		"git_branch":     info.GitBranch,
		"git_tree_state": info.GitTreeState,
		"build_date":     info.BuildDate,
		"go_version":     info.GoVersion,
		"compiler":       info.Compiler,
		"platform":       info.Platform,
	})
}

// GetVersionSimple 获取简化版本信息
// GET /version/simple.
func (h *VersionHandler) GetVersionSimple(c *gin.Context) {
	info := version.Get()

	// 返回简化版本信息
	response.Success(c, gin.H{
		"service": info.ServiceName,
		"version": info.GitVersion,
	})
}

// GetVersionText 获取文本格式版本信息
// GET /version/text.
func (h *VersionHandler) GetVersionText(c *gin.Context) {
	info := version.Get()

	// 返回文本格式
	c.String(200, info.Text())
}

// GetVersionJSON 获取 JSON 格式版本信息
// GET /version/json.
func (h *VersionHandler) GetVersionJSON(c *gin.Context) {
	info := version.Get()

	// 返回 JSON 格式
	c.String(200, info.ToJSON())
}
