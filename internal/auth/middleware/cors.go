package middleware

import (
	"github.com/gin-gonic/gin"

	commonmiddleware "github.com/kart-io/k8s-agent/common/middleware"
)

// CORS creates CORS middleware for frontend integration.
// 使用 common/middleware 的通用实现，避免代码重复
func CORS() gin.HandlerFunc {
	return commonmiddleware.CORS()
}
