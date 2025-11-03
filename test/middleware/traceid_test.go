package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kart-io/k8s-agent/pkg/contextutil"
	"github.com/kart-io/k8s-agent/common/middleware"
)

func TestTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Generates new trace ID when not provided", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.TraceID())

		var capturedTraceID string
		router.GET("/test", func(c *gin.Context) {
			capturedTraceID = contextutil.GetTraceID(c.Request.Context())
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, capturedTraceID, "Should generate trace ID")
		assert.Equal(t, capturedTraceID, w.Header().Get("X-Trace-ID"),
			"Response should contain trace ID in header")
	})

	t.Run("Reuses trace ID from request header", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.TraceID())

		providedTraceID := "test-trace-123"
		var capturedTraceID string

		router.GET("/test", func(c *gin.Context) {
			capturedTraceID = contextutil.GetTraceID(c.Request.Context())
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Trace-ID", providedTraceID)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, providedTraceID, capturedTraceID,
			"Should reuse provided trace ID")
		assert.Equal(t, providedTraceID, w.Header().Get("X-Trace-ID"),
			"Response should echo provided trace ID")
	})

	t.Run("Trace ID propagates through context", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.TraceID())

		var firstTraceID, secondTraceID string

		router.GET("/test", func(c *gin.Context) {
			firstTraceID = contextutil.GetTraceID(c.Request.Context())

			// Simulate passing context to another function
			ctx := c.Request.Context()
			secondTraceID = contextutil.GetTraceID(ctx)

			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, firstTraceID)
		assert.Equal(t, firstTraceID, secondTraceID,
			"Trace ID should propagate through context")
	})

	t.Run("Each request gets unique trace ID", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.TraceID())

		var traceID1, traceID2 string

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"trace_id": contextutil.GetTraceID(c.Request.Context()),
			})
		})

		// First request
		req1 := httptest.NewRequest("GET", "/test", nil)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		traceID1 = w1.Header().Get("X-Trace-ID")

		// Second request
		req2 := httptest.NewRequest("GET", "/test", nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		traceID2 = w2.Header().Get("X-Trace-ID")

		assert.NotEmpty(t, traceID1)
		assert.NotEmpty(t, traceID2)
		assert.NotEqual(t, traceID1, traceID2,
			"Each request should get unique trace ID")
	})
}

func TestTraceIDWithConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Custom header name", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.TraceIDWithConfig(middleware.TraceIDConfig{
			HeaderName: "X-Request-ID",
		}))

		var capturedTraceID string
		router.GET("/test", func(c *gin.Context) {
			capturedTraceID = contextutil.GetTraceID(c.Request.Context())
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, capturedTraceID)
		assert.Equal(t, capturedTraceID, w.Header().Get("X-Request-ID"),
			"Should use custom header name")
		assert.Empty(t, w.Header().Get("X-Trace-ID"),
			"Should not use default header")
	})

	t.Run("Custom generator function", func(t *testing.T) {
		customID := "custom-trace-id-12345"
		router := gin.New()
		router.Use(middleware.TraceIDWithConfig(middleware.TraceIDConfig{
			Generator: func() string {
				return customID
			},
		}))

		var capturedTraceID string
		router.GET("/test", func(c *gin.Context) {
			capturedTraceID = contextutil.GetTraceID(c.Request.Context())
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, customID, capturedTraceID,
			"Should use custom generator")
		assert.Equal(t, customID, w.Header().Get("X-Trace-ID"))
	})

	t.Run("Skip paths configuration", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.TraceIDWithConfig(middleware.TraceIDConfig{
			SkipPaths: []string{"/health", "/metrics"},
		}))

		router.GET("/health", func(c *gin.Context) {
			traceID := contextutil.GetTraceID(c.Request.Context())
			c.JSON(http.StatusOK, gin.H{"trace_id": traceID})
		})

		router.GET("/api/test", func(c *gin.Context) {
			traceID := contextutil.GetTraceID(c.Request.Context())
			c.JSON(http.StatusOK, gin.H{"trace_id": traceID})
		})

		// Health endpoint should skip trace ID
		req1 := httptest.NewRequest("GET", "/health", nil)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)

		assert.Equal(t, http.StatusOK, w1.Code)
		assert.Empty(t, w1.Header().Get("X-Trace-ID"),
			"Health endpoint should skip trace ID")

		// Regular endpoint should generate trace ID
		req2 := httptest.NewRequest("GET", "/api/test", nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)
		assert.NotEmpty(t, w2.Header().Get("X-Trace-ID"),
			"Regular endpoint should generate trace ID")
	})

	t.Run("Reuse trace ID with custom header", func(t *testing.T) {
		providedID := "provided-trace-789"
		router := gin.New()
		router.Use(middleware.TraceIDWithConfig(middleware.TraceIDConfig{
			HeaderName: "X-Custom-Trace",
		}))

		var capturedTraceID string
		router.GET("/test", func(c *gin.Context) {
			capturedTraceID = contextutil.GetTraceID(c.Request.Context())
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Custom-Trace", providedID)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, providedID, capturedTraceID,
			"Should reuse trace ID from custom header")
		assert.Equal(t, providedID, w.Header().Get("X-Custom-Trace"))
	})
}

func TestTraceIDIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Integration with multiple middleware", func(t *testing.T) {
		router := gin.New()

		// Add TraceID middleware first
		router.Use(middleware.TraceID())

		// Add a custom middleware that uses trace ID
		router.Use(func(c *gin.Context) {
			traceID := contextutil.GetTraceID(c.Request.Context())
			require.NotEmpty(t, traceID, "Trace ID should be available in middleware")
			c.Next()
		})

		router.GET("/test", func(c *gin.Context) {
			traceID := contextutil.GetTraceID(c.Request.Context())
			c.JSON(http.StatusOK, gin.H{"trace_id": traceID})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Trace ID survives through error handling", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.TraceID())
		router.Use(gin.Recovery())

		var capturedTraceID string

		router.GET("/test", func(c *gin.Context) {
			capturedTraceID = contextutil.GetTraceID(c.Request.Context())
			c.JSON(http.StatusOK, gin.H{"trace_id": capturedTraceID})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, capturedTraceID)
		assert.Equal(t, capturedTraceID, w.Header().Get("X-Trace-ID"))
	})
}
