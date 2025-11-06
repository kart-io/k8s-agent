package audit

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/kart-io/k8s-agent/internal/auth/types"
)

// MySQLRepository implements Repository using MySQL via GORM.
type MySQLRepository struct {
	db *gorm.DB
}

// NewMySQLRepository creates a new MySQL-based audit repository.
func NewMySQLRepository(db *gorm.DB) *MySQLRepository {
	return &MySQLRepository{
		db: db,
	}
}

// CreateEvent inserts a new audit event with hash chain validation.
func (r *MySQLRepository) CreateEvent(ctx context.Context, event *types.ForcedLogoutEvent) error {
	// Use a transaction to ensure atomicity
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Insert the event
		if err := tx.Create(event).Error; err != nil {
			return fmt.Errorf("create audit event: %w", err)
		}
		return nil
	})
}

// GetEvent retrieves a single audit event by event_id.
func (r *MySQLRepository) GetEvent(ctx context.Context, eventID string) (*types.ForcedLogoutEvent, error) {
	var event types.ForcedLogoutEvent
	if err := r.db.WithContext(ctx).Where("event_id = ?", eventID).First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("audit event not found: %s", eventID)
		}
		return nil, fmt.Errorf("get audit event: %w", err)
	}
	return &event, nil
}

// ListEvents retrieves filtered audit events with pagination.
func (r *MySQLRepository) ListEvents(ctx context.Context, filter AuditFilter) (*AuditEventListResponse, error) {
	// Build query with filters
	query := r.db.WithContext(ctx).Model(&types.ForcedLogoutEvent{})

	// Apply filters
	if filter.TargetUserID != "" {
		query = query.Where("target_user_id = ?", filter.TargetUserID)
	}
	if filter.ActorID != "" {
		query = query.Where("actor_id = ?", filter.ActorID)
	}
	if filter.ActorType != "" {
		query = query.Where("actor_type = ?", filter.ActorType)
	}
	if filter.LogoutType != "" {
		query = query.Where("logout_type = ?", filter.LogoutType)
	}
	if filter.FromDate != nil {
		query = query.Where("timestamp >= ?", filter.FromDate)
	}
	if filter.ToDate != nil {
		query = query.Where("timestamp <= ?", filter.ToDate)
	}

	// Get total count before pagination
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count audit events: %w", err)
	}

	// Apply pagination defaults
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 1000 {
		filter.Limit = 1000 // Maximum limit for safety
	}

	// Retrieve events with pagination (most recent first)
	var events []types.ForcedLogoutEvent
	if err := query.
		Order("timestamp DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&events).Error; err != nil {
		return nil, fmt.Errorf("list audit events: %w", err)
	}

	return &AuditEventListResponse{
		Events: events,
		Total:  int(total),
		Pagination: &Pagination{
			Limit:   filter.Limit,
			Offset:  filter.Offset,
			Total:   int(total),
			HasMore: filter.Offset+len(events) < int(total),
		},
	}, nil
}

// ExportEvents returns audit events in specified format (JSON/CSV).
func (r *MySQLRepository) ExportEvents(ctx context.Context, filter AuditFilter, format ExportFormat) ([]byte, error) {
	// Remove pagination limits for export (but cap at reasonable max)
	filter.Limit = 10000
	filter.Offset = 0

	// Retrieve events
	result, err := r.ListEvents(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("export audit events: %w", err)
	}

	switch format {
	case ExportFormatJSON:
		return r.exportJSON(result.Events)
	case ExportFormatCSV:
		return r.exportCSV(result.Events)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// exportJSON converts events to JSON format.
func (r *MySQLRepository) exportJSON(events []types.ForcedLogoutEvent) ([]byte, error) {
	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal events to JSON: %w", err)
	}
	return data, nil
}

// exportCSV converts events to CSV format.
func (r *MySQLRepository) exportCSV(events []types.ForcedLogoutEvent) ([]byte, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Write CSV headers
	headers := []string{
		"EventID", "Timestamp", "ActorType", "ActorID", "ActorUsername",
		"TargetUserID", "TargetUsername", "SessionJTI", "SessionCount",
		"Reason", "LogoutType", "TriggeredBy", "CurrentHash",
	}
	if err := writer.Write(headers); err != nil {
		return nil, fmt.Errorf("write CSV headers: %w", err)
	}

	// Write event rows
	for _, event := range events {
		row := []string{
			event.EventID,
			event.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			event.ActorType,
			ptrToString(event.ActorID),
			ptrToString(event.ActorUsername),
			event.TargetUserID,
			event.TargetUsername,
			ptrToString(event.SessionJTI),
			fmt.Sprintf("%d", event.SessionCount),
			ptrToString(event.Reason),
			event.LogoutType,
			event.TriggeredBy,
			event.CurrentHash,
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("flush CSV writer: %w", err)
	}

	return []byte(buf.String()), nil
}

// GetLastHash retrieves the hash of the most recent audit event.
func (r *MySQLRepository) GetLastHash(ctx context.Context) (string, error) {
	var event types.ForcedLogoutEvent
	if err := r.db.WithContext(ctx).
		Select("current_hash").
		Order("timestamp DESC, id DESC").
		First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No events exist, return genesis marker
			return "genesis", nil
		}
		return "", fmt.Errorf("get last hash: %w", err)
	}
	return event.CurrentHash, nil
}

// ValidateHashChain verifies the integrity of the entire hash chain.
func (r *MySQLRepository) ValidateHashChain(ctx context.Context) error {
	// Retrieve all events ordered by timestamp
	var events []types.ForcedLogoutEvent
	if err := r.db.WithContext(ctx).
		Order("timestamp ASC, id ASC").
		Find(&events).Error; err != nil {
		return fmt.Errorf("retrieve events for validation: %w", err)
	}

	if len(events) == 0 {
		// Empty chain is valid
		return nil
	}

	// Validate first event has genesis as previous hash
	if ptrToString(events[0].PreviousHash) != "genesis" {
		return fmt.Errorf("first event previous_hash must be 'genesis', got: %s",
			ptrToString(events[0].PreviousHash))
	}

	// Validate each subsequent event's previous_hash matches the previous event's current_hash
	for i := 1; i < len(events); i++ {
		expectedPrevHash := events[i-1].CurrentHash
		actualPrevHash := ptrToString(events[i].PreviousHash)

		if actualPrevHash != expectedPrevHash {
			return fmt.Errorf("hash chain broken at event %d (event_id: %s): "+
				"expected previous_hash=%s, got=%s",
				i, events[i].EventID, expectedPrevHash, actualPrevHash)
		}
	}

	return nil
}

// ptrToString safely converts a string pointer to a string.
func ptrToString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
