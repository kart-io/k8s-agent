package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart/k8s-agent/auth-service/pkg/forced-logout/audit"
)

// AuditHandler handles audit log queries and exports
type AuditHandler struct {
	auditService *audit.Service
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService *audit.Service) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// ListAuditEvents handles GET /api/v1/audit/forced-logout
// Queries forced logout audit events with filtering and pagination
func (h *AuditHandler) ListAuditEvents(c *gin.Context) {
	// Parse filter parameters from query
	filter := audit.AuditFilter{}

	// Target user filter
	if targetUserID := c.Query("target_user_id"); targetUserID != "" {
		filter.TargetUserID = targetUserID
	}

	// Actor filters
	if actorID := c.Query("actor_id"); actorID != "" {
		filter.ActorID = actorID
	}

	if actorType := c.Query("actor_type"); actorType != "" {
		// Validate enum value
		if actorType != "admin" && actorType != "system" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "actor_type must be 'admin' or 'system'",
			})
			return
		}
		filter.ActorType = actorType
	}

	// Logout type filter
	if logoutType := c.Query("logout_type"); logoutType != "" {
		// Validate enum value
		if logoutType != "single" && logoutType != "all" && logoutType != "bulk" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "logout_type must be 'single', 'all', or 'bulk'",
			})
			return
		}
		filter.LogoutType = logoutType
	}

	// Date range filters
	if fromDateStr := c.Query("from_date"); fromDateStr != "" {
		fromDate, err := time.Parse(time.RFC3339, fromDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "Invalid from_date format. Expected ISO 8601 (RFC3339)",
				"details": err.Error(),
			})
			return
		}
		filter.FromDate = &fromDate
	}

	if toDateStr := c.Query("to_date"); toDateStr != "" {
		toDate, err := time.Parse(time.RFC3339, toDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "Invalid to_date format. Expected ISO 8601 (RFC3339)",
				"details": err.Error(),
			})
			return
		}
		filter.ToDate = &toDate
	}

	// Pagination parameters
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
	if limit <= 0 || limit > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "limit must be between 1 and 1000",
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

	filter.Limit = limit
	filter.Offset = offset

	// Retrieve audit events
	result, err := h.auditService.GetAuditTrail(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to retrieve audit events",
			"details": err.Error(),
		})
		return
	}

	// Return successful response
	c.JSON(http.StatusOK, result)
}

// ExportAuditEvents handles GET /api/v1/audit/forced-logout/export
// Exports audit events in JSON or CSV format
func (h *AuditHandler) ExportAuditEvents(c *gin.Context) {
	// Parse export format
	format := c.Query("format")
	if format == "" {
		format = "json" // Default to JSON
	}

	var exportFormat audit.ExportFormat
	switch format {
	case "json":
		exportFormat = audit.ExportFormatJSON
	case "csv":
		exportFormat = audit.ExportFormatCSV
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid format. Must be 'json' or 'csv'",
		})
		return
	}

	// Parse filter parameters (same as ListAuditEvents)
	filter := audit.AuditFilter{}

	if targetUserID := c.Query("target_user_id"); targetUserID != "" {
		filter.TargetUserID = targetUserID
	}

	if actorID := c.Query("actor_id"); actorID != "" {
		filter.ActorID = actorID
	}

	if actorType := c.Query("actor_type"); actorType != "" {
		if actorType != "admin" && actorType != "system" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "actor_type must be 'admin' or 'system'",
			})
			return
		}
		filter.ActorType = actorType
	}

	if logoutType := c.Query("logout_type"); logoutType != "" {
		if logoutType != "single" && logoutType != "all" && logoutType != "bulk" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "logout_type must be 'single', 'all', or 'bulk'",
			})
			return
		}
		filter.LogoutType = logoutType
	}

	if fromDateStr := c.Query("from_date"); fromDateStr != "" {
		fromDate, err := time.Parse(time.RFC3339, fromDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "Invalid from_date format",
			})
			return
		}
		filter.FromDate = &fromDate
	}

	if toDateStr := c.Query("to_date"); toDateStr != "" {
		toDate, err := time.Parse(time.RFC3339, toDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "Invalid to_date format",
			})
			return
		}
		filter.ToDate = &toDate
	}

	// Export audit events
	data, err := h.auditService.ExportAuditLogs(c.Request.Context(), filter, exportFormat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to export audit events",
			"details": err.Error(),
		})
		return
	}

	// Set appropriate headers based on format
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("forced-logout-audit-%s.%s", timestamp, format)

	switch exportFormat {
	case audit.ExportFormatJSON:
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		c.Data(http.StatusOK, "application/json", data)
	case audit.ExportFormatCSV:
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		c.Data(http.StatusOK, "text/csv", data)
	}
}

// parseIntQuery parses an integer query parameter with default value
func (h *AuditHandler) parseIntQuery(c *gin.Context, key string, defaultValue int) (int, error) {
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
