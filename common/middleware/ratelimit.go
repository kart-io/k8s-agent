package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/common/response"
)

// RateLimiter 限流器接口
type RateLimiter interface {
	Allow(key string) bool
}

// TokenBucketLimiter 令牌桶限流器
type TokenBucketLimiter struct {
	rate     float64           // 每秒生成的令牌数
	capacity int               // 令牌桶容量
	buckets  map[string]*bucket
	mu       sync.RWMutex
}

type bucket struct {
	tokens    float64
	lastTime  time.Time
	mu        sync.Mutex
}

// NewTokenBucketLimiter 创建令牌桶限流器
func NewTokenBucketLimiter(rate float64, capacity int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		rate:     rate,
		capacity: capacity,
		buckets:  make(map[string]*bucket),
	}
}

// Allow 检查是否允许请求
func (l *TokenBucketLimiter) Allow(key string) bool {
	l.mu.RLock()
	b, exists := l.buckets[key]
	l.mu.RUnlock()

	if !exists {
		l.mu.Lock()
		b = &bucket{
			tokens:   float64(l.capacity),
			lastTime: time.Now(),
		}
		l.buckets[key] = b
		l.mu.Unlock()
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastTime).Seconds()
	b.tokens = min(b.tokens+elapsed*l.rate, float64(l.capacity))
	b.lastTime = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}

	return false
}

// RateLimit 限流中间件
func RateLimit(limiter RateLimiter, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		if !limiter.Allow(key) {
			response.Error(c, http.StatusTooManyRequests, 429, "Too many requests", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RateLimitByIP 基于 IP 的限流中间件
func RateLimitByIP(rate float64, capacity int) gin.HandlerFunc {
	limiter := NewTokenBucketLimiter(rate, capacity)
	return RateLimit(limiter, func(c *gin.Context) string {
		return c.ClientIP()
	})
}

// RateLimitByUser 基于用户 ID 的限流中间件
func RateLimitByUser(rate float64, capacity int) gin.HandlerFunc {
	limiter := NewTokenBucketLimiter(rate, capacity)
	return RateLimit(limiter, func(c *gin.Context) string {
		userID, exists := c.Get("userID")
		if !exists {
			return c.ClientIP() // 如果没有用户 ID，回退到 IP
		}
		return userID.(string)
	})
}

// min 返回较小值
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
