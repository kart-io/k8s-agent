package audit

import (
	"context"
	"time"

	"github.com/kart-io/k8s-agent/auth-service/pkg/types"
)

// AuditFilter represents filtering criteria for audit event queries
type AuditFilter struct {
	TargetUserID string
	ActorID      string
	ActorType    string // admin, system
	LogoutType   string // single, all, bulk
	FromDate     *time.Time
	ToDate       *time.Time
	Limit        int
	Offset       int
}

// AuditEventListResponse represents a paginated list of audit events
type AuditEventListResponse struct {
	Events     []types.ForcedLogoutEvent `json:"events"`
	Total      int                       `json:"total"`
	Pagination *Pagination               `json:"pagination"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	Total   int  `json:"total"`
	HasMore bool `json:"has_more"`
}

// ExportFormat represents the format for exporting audit logs
type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatCSV  ExportFormat = "csv"
)

// Repository defines the interface for audit event storage operations
type Repository interface {
	// CreateEvent inserts a new audit event with hash chain validation
	CreateEvent(ctx context.Context, event *types.ForcedLogoutEvent) error

	// GetEvent retrieves a single audit event by event_id
	GetEvent(ctx context.Context, eventID string) (*types.ForcedLogoutEvent, error)

	// ListEvents retrieves filtered audit events with pagination
	ListEvents(ctx context.Context, filter AuditFilter) (*AuditEventListResponse, error)

	// ExportEvents returns audit events in specified format (JSON/CSV)
	ExportEvents(ctx context.Context, filter AuditFilter, format ExportFormat) ([]byte, error)

	// GetLastHash retrieves the hash of the most recent audit event
	// Returns empty string if no events exist
	GetLastHash(ctx context.Context) (string, error)

	// ValidateHashChain verifies the integrity of the entire hash chain
	// Returns nil if valid, error describing the failure point if invalid
	ValidateHashChain(ctx context.Context) error
}
