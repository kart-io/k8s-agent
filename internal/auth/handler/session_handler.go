package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/internal/auth/forced-logout/session"
)

// SessionHandler handles session-related HTTP requests
type SessionHandler struct {
	sessionService *session.Service
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(sessionService *session.Service) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

// ListUserSessions handles GET /api/v1/sessions/users/:userId
// Lists all active sessions for a specific user
func (h *SessionHandler) ListUserSessions(c *gin.Context) {
	// Extract userId from path parameter
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "user_id parameter is required",
		})
		return
	}

	// Validate UUID format (basic check)
	if len(userID) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid user_id format. Expected UUID.",
		})
		return
	}

	// Parse pagination parameters
	limit, err := h.parseIntQuery(c, "limit", 50)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid limit parameter",
			"details": err.Error(),
		})
		return
	}

	offset, err := h.parseIntQuery(c, "offset", 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid offset parameter",
			"details": err.Error(),
		})
		return
	}

	// Validate pagination limits
	if limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "limit must be between 1 and 100",
		})
		return
	}

	if offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "offset must be non-negative",
		})
		return
	}

	// Retrieve user sessions
	sessionList, err := h.sessionService.GetUserSessions(c.Request.Context(), userID, limit, offset)
	if err != nil {
		// Check if it's a "user not found" vs internal error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to retrieve user sessions",
			"details": err.Error(),
		})
		return
	}

	// Handle case where user has no active sessions
	if len(sessionList.Sessions) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"user_id":  userID,
			"username": sessionList.Username,
			"total":    0,
			"sessions": []interface{}{},
			"pagination": gin.H{
				"limit":    limit,
				"offset":   offset,
				"total":    0,
				"has_more": false,
			},
		})
		return
	}

	// Return successful response
	c.JSON(http.StatusOK, sessionList)
}

// parseIntQuery parses an integer query parameter with default value
func (h *SessionHandler) parseIntQuery(c *gin.Context, key string, defaultValue int) (int, error) {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, err
	}

	return value, nil
}
