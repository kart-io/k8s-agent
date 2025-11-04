package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// CORS 跨域中间件.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取配置
		allowOrigins := viper.GetStringSlice("cors.allow_origins")
		allowMethods := viper.GetStringSlice("cors.allow_methods")
		allowHeaders := viper.GetStringSlice("cors.allow_headers")
		exposeHeaders := viper.GetStringSlice("cors.expose_headers")
		allowCredentials := viper.GetBool("cors.allow_credentials")

		// 设置 CORS 响应头
		origin := c.Request.Header.Get("Origin")
		if origin != "" && isAllowedOrigin(origin, allowOrigins) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(allowMethods, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(allowHeaders, ", "))
		c.Writer.Header().Set("Access-Control-Expose-Headers", strings.Join(exposeHeaders, ", "))

		if allowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		c.Writer.Header().Set("Access-Control-Max-Age", "43200")

		// 处理 OPTIONS 预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isAllowedOrigin 检查源是否被允许.
func isAllowedOrigin(origin string, allowOrigins []string) bool {
	for _, allowed := range allowOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}
