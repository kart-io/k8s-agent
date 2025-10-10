package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	forcedlogout "github.com/kart/k8s-agent/auth-service/pkg/forced-logout"
	"github.com/kart/k8s-agent/auth-service/pkg/types"
)

// ForcedLogoutHandler handles forced logout operations
type ForcedLogoutHandler struct {
	forcedLogoutService *forcedlogout.Service
}

// NewForcedLogoutHandler creates a new forced logout handler
func NewForcedLogoutHandler(forcedLogoutService *forcedlogout.Service) *ForcedLogoutHandler {
	return &ForcedLogoutHandler{
		forcedLogoutService: forcedLogoutService,
	}
}

// ForceLogoutSession handles POST /api/v1/forced-logout/session/:jti
// Terminates a single session by JTI
func (h *ForcedLogoutHandler) ForceLogoutSession(c *gin.Context) {
	// Extract JTI from path parameter
	jti := c.Param("jti")
	if jti == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "jti parameter is required",
		})
		return
	}

	// Parse request body
	var req types.ForceLogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Set default triggered_by if not provided
	if req.TriggeredBy == "" {
		req.TriggeredBy = "manual"
	}

	// Extract actor information from context (set by auth middleware)
	actorID, _ := c.Get("actor_id")
	actorUsername, _ := c.Get("actor_username")
	actorIPAddress := c.ClientIP()

	// Extract user_id from query parameter (required for verification)
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "user_id query parameter is required",
		})
		return
	}

	// Call forced logout service
	result, err := h.forcedLogoutService.ForceLogoutSession(c.Request.Context(), forcedlogout.ForceLogoutSessionParams{
		JTI:            jti,
		UserID:         userID,
		ActorID:        actorID.(string),
		ActorUsername:  actorUsername.(string),
		ActorIPAddress: actorIPAddress,
		Reason:         req.Reason,
		TriggeredBy:    req.TriggeredBy,
		CorrelationID:  req.CorrelationID,
	})

	if err != nil {
		// Handle specific error cases
		if err.Error() == "session not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"message": "Session not found",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to force logout session",
			"details": err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, types.ForceLogoutResponse{
		EventID:      result.EventID,
		Success:      true,
		SessionCount: result.SessionCount,
		Timestamp:    result.Timestamp,
		Message:      "Session terminated successfully",
	})
}

// ForceLogoutUser handles POST /api/v1/forced-logout/user/:userId
// Terminates all sessions for a user
func (h *ForcedLogoutHandler) ForceLogoutUser(c *gin.Context) {
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

	// Parse request body
	var req types.ForceLogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Set default triggered_by if not provided
	if req.TriggeredBy == "" {
		req.TriggeredBy = "manual"
	}

	// Extract actor information from context
	actorID, _ := c.Get("actor_id")
	actorUsername, _ := c.Get("actor_username")
	actorIPAddress := c.ClientIP()

	// Call forced logout service
	result, err := h.forcedLogoutService.ForceLogoutUser(c.Request.Context(), forcedlogout.ForceLogoutUserParams{
		UserID:         userID,
		ActorID:        actorID.(string),
		ActorUsername:  actorUsername.(string),
		ActorIPAddress: actorIPAddress,
		Reason:         req.Reason,
		TriggeredBy:    req.TriggeredBy,
		CorrelationID:  req.CorrelationID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to force logout user sessions",
			"details": err.Error(),
		})
		return
	}

	// Handle case where user has no active sessions
	if result.SessionCount == 0 {
		c.JSON(http.StatusOK, types.ForceLogoutResponse{
			EventID:      "",
			Success:      true,
			SessionCount: 0,
			TargetUserID: userID,
			Timestamp:    result.Timestamp,
			Message:      "User has no active sessions",
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, types.ForceLogoutResponse{
		EventID:      result.EventID,
		Success:      true,
		SessionCount: result.SessionCount,
		TargetUserID: userID,
		Timestamp:    result.Timestamp,
		Message:      "All user sessions terminated successfully",
	})
}

// BulkForceLogout handles POST /api/v1/forced-logout/sessions
// Terminates multiple sessions in bulk
func (h *ForcedLogoutHandler) BulkForceLogout(c *gin.Context) {
	// Parse request body
	var req types.BulkForceLogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate array length
	if len(req.SessionJTIs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "At least one session JTI is required",
		})
		return
	}

	if len(req.SessionJTIs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Maximum 100 sessions allowed per bulk operation",
		})
		return
	}

	// Set default triggered_by if not provided
	if req.TriggeredBy == "" {
		req.TriggeredBy = "manual"
	}

	// Extract actor information from context
	actorID, _ := c.Get("actor_id")
	actorUsername, _ := c.Get("actor_username")
	actorIPAddress := c.ClientIP()

	// Call forced logout service
	result, err := h.forcedLogoutService.BulkForceLogout(c.Request.Context(), forcedlogout.BulkForceLogoutParams{
		JTIs:           req.SessionJTIs,
		ActorID:        actorID.(string),
		ActorUsername:  actorUsername.(string),
		ActorIPAddress: actorIPAddress,
		Reason:         req.Reason,
		TriggeredBy:    req.TriggeredBy,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to perform bulk forced logout",
			"details": err.Error(),
		})
		return
	}

	// Return response (200 OK even with partial failures)
	// Detailed results per session are in the response body
	c.JSON(http.StatusOK, types.BulkForceLogoutResponse{
		EventID:        result.EventID,
		TotalRequested: result.TotalCount,
		Successful:     result.SuccessCount,
		Failed:         result.FailedCount,
		Results:        result.Results,
	})
}
