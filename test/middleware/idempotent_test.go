package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kart-io/k8s-agent/pkg/idempotent"
	"github.com/kart-io/k8s-agent/common/middleware"
)

func TestIdempotent(t *testing.T) {
	// Setup Gin in test mode
	gin.SetMode(gin.TestMode)

	// Create an in-memory store for testing
	store := idempotent.NewMemoryStore()
	handler := idempotent.NewHandler(store, time.Hour, 5*time.Minute)

	// Setup router with idempotency middleware
	router := gin.New()
	router.Use(middleware.Idempotent(middleware.IdempotentConfig{
		Handler:       handler,
		PathBlacklist: middleware.DefaultPathBlacklist(),
	}))

	// Test endpoint
	type TestRequest struct {
		Name string `json:"name"`
	}
	type TestResponse struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Created string `json:"created"`
	}

	router.POST("/api/v1/workflows", func(c *gin.Context) {
		var req TestRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get idempotent key
		key := middleware.GetIdempotentKey(c)
		require.NotEmpty(t, key, "Idempotent key should be available in context")

		// Execute operation idempotently
		response, err := handler.Execute(c.Request.Context(), key, func(ctx context.Context) ([]byte, error) {
			// Simulate workflow creation
			resp := TestResponse{
				ID:      "wf-12345",
				Name:    req.Name,
				Created: time.Now().Format(time.RFC3339),
			}
			return json.Marshal(resp)
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Data(http.StatusOK, "application/json", response)
	})

	t.Run("Missing idempotent key", func(t *testing.T) {
		reqBody := TestRequest{Name: "test-workflow"}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/workflows", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Missing X-Idempotent-Key")
	})

	t.Run("First request succeeds", func(t *testing.T) {
		reqBody := TestRequest{Name: "test-workflow"}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/workflows", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Idempotent-Key", "test-key-001")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp TestResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "wf-12345", resp.ID)
		assert.Equal(t, "test-workflow", resp.Name)
	})

	t.Run("Duplicate request returns cached response", func(t *testing.T) {
		reqBody := TestRequest{Name: "test-workflow"}
		bodyBytes, _ := json.Marshal(reqBody)

		// First request
		req1 := httptest.NewRequest("POST", "/api/v1/workflows", bytes.NewReader(bodyBytes))
		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("X-Idempotent-Key", "test-key-002")
		w1 := httptest.NewRecorder()

		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		var resp1 TestResponse
		json.Unmarshal(w1.Body.Bytes(), &resp1)
		firstCreatedTime := resp1.Created

		// Wait a bit
		time.Sleep(10 * time.Millisecond)

		// Second request with same key (should return cached response)
		req2 := httptest.NewRequest("POST", "/api/v1/workflows", bytes.NewReader(bodyBytes))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("X-Idempotent-Key", "test-key-002")
		w2 := httptest.NewRecorder()

		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)

		var resp2 TestResponse
		json.Unmarshal(w2.Body.Bytes(), &resp2)

		// Should return the same response (including timestamp)
		assert.Equal(t, resp1.ID, resp2.ID)
		assert.Equal(t, resp1.Created, resp2.Created)
		assert.Equal(t, firstCreatedTime, resp2.Created)
		assert.Equal(t, "true", w2.Header().Get("X-Idempotent-Replayed"))
	})

	t.Run("Different keys create different resources", func(t *testing.T) {
		reqBody := TestRequest{Name: "test-workflow"}
		bodyBytes, _ := json.Marshal(reqBody)

		// First request
		req1 := httptest.NewRequest("POST", "/api/v1/workflows", bytes.NewReader(bodyBytes))
		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("X-Idempotent-Key", "test-key-003")
		w1 := httptest.NewRecorder()

		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Second request with different key (should create new resource)
		req2 := httptest.NewRequest("POST", "/api/v1/workflows", bytes.NewReader(bodyBytes))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("X-Idempotent-Key", "test-key-004")
		w2 := httptest.NewRecorder()

		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)

		// Both should succeed (different operations)
		assert.NotEqual(t, "true", w2.Header().Get("X-Idempotent-Replayed"))
	})

	t.Run("Path not in blacklist bypasses idempotency", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.Idempotent(middleware.IdempotentConfig{
			Handler:       handler,
			PathBlacklist: middleware.DefaultPathBlacklist(),
		}))

		router.GET("/api/v1/workflows", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "list workflows"})
		})

		req := httptest.NewRequest("GET", "/api/v1/workflows", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// GET requests don't require idempotency key
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestGetIdempotentKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/test", func(c *gin.Context) {
		c.Set("idempotent_key", "test-key-123")
		key := middleware.GetIdempotentKey(c)
		c.JSON(http.StatusOK, gin.H{"key": key})
	})

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test-key-123")
}

func TestDefaultPathBlacklist(t *testing.T) {
	blacklist := middleware.DefaultPathBlacklist()

	// Test orchestrator paths
	assert.True(t, blacklist["POST /api/v1/workflows"])
	assert.True(t, blacklist["POST /api/v1/strategies"])
	assert.True(t, blacklist["POST /api/v1/workflows/:id/execute"])

	// Test agent-manager paths
	assert.True(t, blacklist["POST /api/v1/commands"])
	assert.True(t, blacklist["POST /api/v1/events"])
	assert.True(t, blacklist["POST /api/v1/agents"])

	// Test reasoning paths
	assert.True(t, blacklist["POST /api/v1/analyze/root-cause"])
	assert.True(t, blacklist["POST /api/v1/recommendations"])

	// GET requests should not be in blacklist
	assert.False(t, blacklist["GET /api/v1/workflows"])
}
