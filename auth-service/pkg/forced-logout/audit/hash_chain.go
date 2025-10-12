package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/kart-io/k8s-agent/auth-service/pkg/types"
)

// HashChain provides cryptographic hash chain operations for tamper detection
type HashChain struct{}

// NewHashChain creates a new hash chain instance
func NewHashChain() *HashChain {
	return &HashChain{}
}

// ComputeHash generates a SHA-256 hash for an audit event
// Hash includes: event_id, timestamp, actor_id, target_user_id, session_count, reason, previous_hash
// This ensures any modification to these fields will break the chain
func (hc *HashChain) ComputeHash(event *types.ForcedLogoutEvent) string {
	// Build the data string to hash
	data := fmt.Sprintf("%s|%s|%s|%s|%d|%s|%s",
		event.EventID,
		event.Timestamp.UTC().Format(time.RFC3339Nano),
		ptrToStringOrEmpty(event.ActorID),
		event.TargetUserID,
		event.SessionCount,
		ptrToStringOrEmpty(event.Reason),
		ptrToStringOrEmpty(event.PreviousHash),
	)

	// Compute SHA-256 hash
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ValidateHash verifies that an event's current_hash matches its computed hash
func (hc *HashChain) ValidateHash(event *types.ForcedLogoutEvent) bool {
	expectedHash := hc.ComputeHash(event)
	return event.CurrentHash == expectedHash
}

// ValidateChain verifies the integrity of a chain of events
// Returns nil if the chain is valid, error describing the failure point if invalid
func (hc *HashChain) ValidateChain(events []types.ForcedLogoutEvent) error {
	if len(events) == 0 {
		// Empty chain is valid
		return nil
	}

	// Validate first event
	if ptrToStringOrEmpty(events[0].PreviousHash) != "genesis" {
		return fmt.Errorf("first event (event_id: %s) must have previous_hash='genesis', got: '%s'",
			events[0].EventID,
			ptrToStringOrEmpty(events[0].PreviousHash))
	}

	// Validate each event's hash computation
	for i := range events {
		if !hc.ValidateHash(&events[i]) {
			return fmt.Errorf("event %d (event_id: %s) has invalid current_hash: "+
				"expected=%s, got=%s",
				i,
				events[i].EventID,
				hc.ComputeHash(&events[i]),
				events[i].CurrentHash)
		}
	}

	// Validate chain linkage
	for i := 1; i < len(events); i++ {
		expectedPrevHash := events[i-1].CurrentHash
		actualPrevHash := ptrToStringOrEmpty(events[i].PreviousHash)

		if actualPrevHash != expectedPrevHash {
			return fmt.Errorf("chain broken between event %d and %d: "+
				"event %d (event_id: %s) previous_hash=%s does not match "+
				"event %d (event_id: %s) current_hash=%s",
				i-1, i,
				i, events[i].EventID, actualPrevHash,
				i-1, events[i-1].EventID, expectedPrevHash)
		}
	}

	return nil
}

// DetectTampering checks if any event in the chain has been tampered with
// Returns the first tampered event's index and details, or nil if chain is valid
func (hc *HashChain) DetectTampering(events []types.ForcedLogoutEvent) *TamperDetection {
	if len(events) == 0 {
		return nil
	}

	// Check first event genesis
	if ptrToStringOrEmpty(events[0].PreviousHash) != "genesis" {
		return &TamperDetection{
			EventIndex:  0,
			EventID:     events[0].EventID,
			TamperType:  TamperTypeInvalidGenesis,
			Description: fmt.Sprintf("First event previous_hash should be 'genesis', got: '%s'", ptrToStringOrEmpty(events[0].PreviousHash)),
		}
	}

	// Check each event's hash integrity
	for i := range events {
		if !hc.ValidateHash(&events[i]) {
			return &TamperDetection{
				EventIndex: i,
				EventID:    events[i].EventID,
				TamperType: TamperTypeInvalidHash,
				Description: fmt.Sprintf("Event hash mismatch: expected=%s, got=%s",
					hc.ComputeHash(&events[i]),
					events[i].CurrentHash),
			}
		}
	}

	// Check chain linkage
	for i := 1; i < len(events); i++ {
		expectedPrevHash := events[i-1].CurrentHash
		actualPrevHash := ptrToStringOrEmpty(events[i].PreviousHash)

		if actualPrevHash != expectedPrevHash {
			return &TamperDetection{
				EventIndex: i,
				EventID:    events[i].EventID,
				TamperType: TamperTypeBrokenChain,
				Description: fmt.Sprintf("Chain broken: previous_hash=%s does not match previous event's current_hash=%s",
					actualPrevHash,
					expectedPrevHash),
			}
		}
	}

	// Chain is valid
	return nil
}

// TamperType represents the type of tampering detected
type TamperType string

const (
	TamperTypeInvalidGenesis TamperType = "invalid_genesis"
	TamperTypeInvalidHash    TamperType = "invalid_hash"
	TamperTypeBrokenChain    TamperType = "broken_chain"
)

// TamperDetection represents a detected tampering in the audit chain
type TamperDetection struct {
	EventIndex  int        `json:"event_index"`
	EventID     string     `json:"event_id"`
	TamperType  TamperType `json:"tamper_type"`
	Description string     `json:"description"`
}

// ptrToStringOrEmpty safely converts a string pointer to a string, returning empty string if nil
func ptrToStringOrEmpty(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
