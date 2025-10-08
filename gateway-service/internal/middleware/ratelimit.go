package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

var (
	limiters = make(map[string]*rate.Limiter)
	rdb      *redis.Client
)

// InitRateLimiter 初始化限流器
func InitRateLimiter(redisClient *redis.Client) {
	rdb = redisClient
}

// RateLimit 限流中间件
func RateLimit() gin.HandlerFunc {
	enabled := viper.GetBool("rate_limit.enabled")
	if !enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	rps := viper.GetInt("rate_limit.requests_per_second")
	burst := viper.GetInt("rate_limit.burst")

	return func(c *gin.Context) {
		// 获取客户端标识（IP 或 用户ID）
		clientID := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			clientID = fmt.Sprintf("%v", userID)
		}

		// 使用 Redis 实现分布式限流
		if rdb != nil {
			allowed, err := checkRedisRateLimit(clientID, rps)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "限流检查失败",
				})
				c.Abort()
				return
			}

			if !allowed {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "请求过于频繁，请稍后再试",
				})
				c.Abort()
				return
			}
		} else {
			// 本地限流
			limiter := getLimiter(clientID, rps, burst)
			if !limiter.Allow() {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "请求过于频繁，请稍后再试",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// getLimiter 获取或创建限流器
func getLimiter(clientID string, rps, burst int) *rate.Limiter {
	if limiter, exists := limiters[clientID]; exists {
		return limiter
	}

	limiter := rate.NewLimiter(rate.Limit(rps), burst)
	limiters[clientID] = limiter
	return limiter
}

// checkRedisRateLimit 使用 Redis 检查限流
func checkRedisRateLimit(clientID string, rps int) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("rate_limit:%s", clientID)

	// 使用 Redis 的 INCR 和 EXPIRE 实现简单限流
	count, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		// 第一次请求，设置过期时间为 1 秒
		rdb.Expire(ctx, key, time.Second)
	}

	// 检查是否超过限制
	return count <= int64(rps), nil
}
