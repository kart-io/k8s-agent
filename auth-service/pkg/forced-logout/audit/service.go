package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kart-io/k8s-agent/auth-service/pkg/types"
)

// Service provides business logic for audit event management
type Service struct {
	repo      Repository
	hashChain *HashChain
}

// NewService creates a new audit service
func NewService(repo Repository) *Service {
	return &Service{
		repo:      repo,
		hashChain: NewHashChain(),
	}
}

// RecordEventParams contains parameters for recording an audit event
type RecordEventParams struct {
	ActorType       string                // admin, system
	ActorID         *string               // Optional: admin user ID
	ActorUsername   *string               // Optional: admin username
	ActorIPAddress  *string               // Optional: admin IP address
	TargetUserID    string                // Required: target user ID
	TargetUsername  string                // Required: target username
	SessionJTI      *string               // Optional: single session JTI (for single logout)
	SessionCount    int                   // Required: number of sessions terminated
	SessionMetadata types.SessionMetadata // Session details
	Reason          *string               // Optional: reason for forced logout
	LogoutType      string                // Required: single, all, bulk
	TriggeredBy     string                // Required: manual, policy, security_incident
	CorrelationID   *string               // Optional: correlation ID for tracking
}

// RecordEvent creates a new audit event with hash chain integrity
func (s *Service) RecordEvent(ctx context.Context, params RecordEventParams) (string, error) {
	// Validate required fields
	if err := s.validateParams(params); err != nil {
		return "", fmt.Errorf("validate params: %w", err)
	}

	// Get the previous hash for chain continuity
	previousHash, err := s.repo.GetLastHash(ctx)
	if err != nil {
		return "", fmt.Errorf("get last hash: %w", err)
	}

	// Generate unique event ID
	eventID := uuid.New().String()
	timestamp := time.Now()

	// Create event (without hash yet)
	event := &types.ForcedLogoutEvent{
		EventID:         eventID,
		Timestamp:       timestamp,
		ActorType:       params.ActorType,
		ActorID:         params.ActorID,
		ActorUsername:   params.ActorUsername,
		ActorIPAddress:  params.ActorIPAddress,
		TargetUserID:    params.TargetUserID,
		TargetUsername:  params.TargetUsername,
		SessionJTI:      params.SessionJTI,
		SessionCount:    params.SessionCount,
		SessionMetadata: params.SessionMetadata,
		Reason:          params.Reason,
		LogoutType:      params.LogoutType,
		TriggeredBy:     params.TriggeredBy,
		PreviousHash:    &previousHash,
		CorrelationID:   params.CorrelationID,
	}

	// Compute current hash
	event.CurrentHash = s.hashChain.ComputeHash(event)

	// Store event in database
	if err := s.repo.CreateEvent(ctx, event); err != nil {
		return "", fmt.Errorf("create audit event: %w", err)
	}

	return eventID, nil
}

// GetAuditTrail retrieves filtered audit events
func (s *Service) GetAuditTrail(ctx context.Context, filter AuditFilter) (*AuditEventListResponse, error) {
	return s.repo.ListEvents(ctx, filter)
}

// ExportAuditLogs formats audit data for export
func (s *Service) ExportAuditLogs(ctx context.Context, filter AuditFilter, format ExportFormat) ([]byte, error) {
	return s.repo.ExportEvents(ctx, filter, format)
}

// ValidateIntegrity checks the integrity of the entire audit hash chain
func (s *Service) ValidateIntegrity(ctx context.Context) error {
	return s.repo.ValidateHashChain(ctx)
}

// GetEvent retrieves a single audit event by ID
func (s *Service) GetEvent(ctx context.Context, eventID string) (*types.ForcedLogoutEvent, error) {
	return s.repo.GetEvent(ctx, eventID)
}

// DetectTampering checks if any event in the chain has been tampered with
func (s *Service) DetectTampering(ctx context.Context) (*TamperDetection, error) {
	// Retrieve all events
	filter := AuditFilter{
		Limit:  10000, // Large limit to get all events
		Offset: 0,
	}

	result, err := s.repo.ListEvents(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list events for tampering detection: %w", err)
	}

	// Sort events by timestamp (ListEvents returns DESC, we need ASC for chain validation)
	// Reverse the slice
	events := result.Events
	for i, j := 0, len(events)-1; i < j; i, j = i+1, j-1 {
		events[i], events[j] = events[j], events[i]
	}

	// Detect tampering
	return s.hashChain.DetectTampering(events), nil
}

// RecordEventAsync creates an audit event asynchronously (for performance)
// This is useful when audit logging shouldn't block the main operation
// Note: Errors are logged but not returned
func (s *Service) RecordEventAsync(ctx context.Context, params RecordEventParams, errHandler func(error)) {
	go func() {
		// Use a background context with timeout to prevent long-running goroutines
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if _, err := s.RecordEvent(bgCtx, params); err != nil {
			if errHandler != nil {
				errHandler(fmt.Errorf("async audit event recording failed: %w", err))
			}
		}
	}()
}

// validateParams validates the required fields in RecordEventParams
func (s *Service) validateParams(params RecordEventParams) error {
	if params.ActorType == "" {
		return fmt.Errorf("actor_type is required")
	}
	if params.ActorType != "admin" && params.ActorType != "system" {
		return fmt.Errorf("actor_type must be 'admin' or 'system', got: %s", params.ActorType)
	}

	if params.TargetUserID == "" {
		return fmt.Errorf("target_user_id is required")
	}
	if params.TargetUsername == "" {
		return fmt.Errorf("target_username is required")
	}

	if params.SessionCount < 0 {
		return fmt.Errorf("session_count must be non-negative")
	}

	if params.LogoutType == "" {
		return fmt.Errorf("logout_type is required")
	}
	if params.LogoutType != "single" && params.LogoutType != "all" && params.LogoutType != "bulk" {
		return fmt.Errorf("logout_type must be 'single', 'all', or 'bulk', got: %s", params.LogoutType)
	}

	if params.TriggeredBy == "" {
		return fmt.Errorf("triggered_by is required")
	}
	if params.TriggeredBy != "manual" && params.TriggeredBy != "policy" && params.TriggeredBy != "security_incident" {
		return fmt.Errorf("triggered_by must be 'manual', 'policy', or 'security_incident', got: %s", params.TriggeredBy)
	}

	// Validate logout_type and session_jti consistency
	if params.LogoutType == "single" && params.SessionJTI == nil {
		return fmt.Errorf("session_jti is required when logout_type is 'single'")
	}

	return nil
}
