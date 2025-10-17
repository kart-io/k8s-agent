package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// Recovery 恢复中间件（处理 panic）
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "Internal server error",
			"error":   "An unexpected error occurred",
		})
	})
}

// Timeout 超时中间件
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 创建带超时的上下文
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// 替换请求上下文
		c.Request = c.Request.WithContext(timeoutCtx)

		// 创建完成通道
		finish := make(chan struct{})

		go func() {
			c.Next()
			close(finish)
		}()

		select {
		case <-finish:
			// 请求正常完成
		case <-timeoutCtx.Done():
			// 请求超时
			c.JSON(504, gin.H{
				"code":    504,
				"message": "Request timeout",
				"error":   "The request took too long to process",
			})
			c.Abort()
		}
	}
}
