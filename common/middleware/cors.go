package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// CORSWithConfig 自定义配置的 CORS 中间件
func CORSWithConfig(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查是否允许该来源
		if config.AllowAllOrigins {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else if contains(config.AllowOrigins, origin) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		if config.AllowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if len(config.AllowHeaders) > 0 {
			c.Writer.Header().Set("Access-Control-Allow-Headers", joinStrings(config.AllowHeaders))
		}

		if len(config.AllowMethods) > 0 {
			c.Writer.Header().Set("Access-Control-Allow-Methods", joinStrings(config.AllowMethods))
		}

		if len(config.ExposeHeaders) > 0 {
			c.Writer.Header().Set("Access-Control-Expose-Headers", joinStrings(config.ExposeHeaders))
		}

		if config.MaxAge > 0 {
			c.Writer.Header().Set("Access-Control-Max-Age", string(rune(config.MaxAge)))
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// CORSConfig CORS 配置
type CORSConfig struct {
	AllowAllOrigins  bool
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig 默认 CORS 配置
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           43200, // 12 hours
	}
}

// contains 检查字符串是否在切片中
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// joinStrings 连接字符串切片
func joinStrings(slice []string) string {
	result := ""
	for i, s := range slice {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}
