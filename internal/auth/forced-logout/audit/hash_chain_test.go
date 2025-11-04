package audit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kart-io/k8s-agent/internal/auth/types"
)

func TestNewHashChain(t *testing.T) {
	hc := NewHashChain()
	assert.NotNil(t, hc)
}

func TestComputeHash(t *testing.T) {
	hc := NewHashChain()

	genesis := stringPtr("genesis")
	actorID := stringPtr("admin-123")
	reason := stringPtr("security_breach")

	event := &types.ForcedLogoutEvent{
		EventID:      "event-001",
		Timestamp:    time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC),
		ActorID:      actorID,
		TargetUserID: "user-456",
		SessionCount: 3,
		Reason:       reason,
		PreviousHash: genesis,
	}

	hash1 := hc.ComputeHash(event)
	assert.NotEmpty(t, hash1)
	assert.Len(t, hash1, 64) // SHA-256 produces 64 hex characters

	// Same input should produce same hash (deterministic)
	hash2 := hc.ComputeHash(event)
	assert.Equal(t, hash1, hash2)
}

func TestComputeHash_DifferentInputs(t *testing.T) {
	hc := NewHashChain()

	genesis := stringPtr("genesis")
	actorID := stringPtr("admin-123")
	reason := stringPtr("security_breach")

	event1 := &types.ForcedLogoutEvent{
		EventID:      "event-001",
		Timestamp:    time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC),
		ActorID:      actorID,
		TargetUserID: "user-456",
		SessionCount: 3,
		Reason:       reason,
		PreviousHash: genesis,
	}

	// Different event ID
	event2 := &types.ForcedLogoutEvent{
		EventID:      "event-002", // Changed
		Timestamp:    time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC),
		ActorID:      actorID,
		TargetUserID: "user-456",
		SessionCount: 3,
		Reason:       reason,
		PreviousHash: genesis,
	}

	hash1 := hc.ComputeHash(event1)
	hash2 := hc.ComputeHash(event2)
	assert.NotEqual(t, hash1, hash2, "Different event IDs should produce different hashes")
}

func TestComputeHash_TimestampPrecision(t *testing.T) {
	hc := NewHashChain()

	genesis := stringPtr("genesis")
	actorID := stringPtr("admin-123")

	// Events differ by 1 nanosecond
	event1 := &types.ForcedLogoutEvent{
		EventID:      "event-001",
		Timestamp:    time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC),
		ActorID:      actorID,
		TargetUserID: "user-456",
		SessionCount: 1,
		PreviousHash: genesis,
	}

	event2 := &types.ForcedLogoutEvent{
		EventID:      "event-001",
		Timestamp:    time.Date(2025, 10, 10, 12, 0, 0, 1, time.UTC), // 1ns difference
		ActorID:      actorID,
		TargetUserID: "user-456",
		SessionCount: 1,
		PreviousHash: genesis,
	}

	hash1 := hc.ComputeHash(event1)
	hash2 := hc.ComputeHash(event2)
	assert.NotEqual(t, hash1, hash2, "Timestamp changes should affect hash")
}

func TestValidateHash_ValidEvent(t *testing.T) {
	hc := NewHashChain()

	genesis := stringPtr("genesis")
	event := &types.ForcedLogoutEvent{
		EventID:      "event-001",
		Timestamp:    time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC),
		ActorID:      stringPtr("admin-123"),
		TargetUserID: "user-456",
		SessionCount: 1,
		PreviousHash: genesis,
	}

	// Compute and set the hash
	event.CurrentHash = hc.ComputeHash(event)

	assert.True(t, hc.ValidateHash(event))
}

func TestValidateHash_TamperedEvent(t *testing.T) {
	hc := NewHashChain()

	genesis := stringPtr("genesis")
	event := &types.ForcedLogoutEvent{
		EventID:      "event-001",
		Timestamp:    time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC),
		ActorID:      stringPtr("admin-123"),
		TargetUserID: "user-456",
		SessionCount: 1,
		PreviousHash: genesis,
	}

	// Compute and set the hash
	event.CurrentHash = hc.ComputeHash(event)

	// Tamper with the event
	event.SessionCount = 999

	// Validation should fail
	assert.False(t, hc.ValidateHash(event))
}

func TestValidateChain_EmptyChain(t *testing.T) {
	hc := NewHashChain()

	err := hc.ValidateChain([]types.ForcedLogoutEvent{})
	assert.NoError(t, err)
}

func TestValidateChain_SingleEvent(t *testing.T) {
	hc := NewHashChain()

	genesis := stringPtr("genesis")
	event := types.ForcedLogoutEvent{
		EventID:      "event-001",
		Timestamp:    time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC),
		ActorID:      stringPtr("admin-123"),
		TargetUserID: "user-456",
		SessionCount: 1,
		PreviousHash: genesis,
	}
	event.CurrentHash = hc.ComputeHash(&event)

	err := hc.ValidateChain([]types.ForcedLogoutEvent{event})
	assert.NoError(t, err)
}

func TestValidateChain_MultipleEvents(t *testing.T) {
	hc := NewHashChain()

	genesis := stringPtr("genesis")

	// Event 1 (genesis)
	event1 := types.ForcedLogoutEvent{
		EventID:      "event-001",
		Timestamp:    time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC),
		ActorID:      stringPtr("admin-123"),
		TargetUserID: "user-456",
		SessionCount: 1,
		PreviousHash: genesis,
	}
	event1.CurrentHash = hc.ComputeHash(&event1)

	// Event 2 (links to event1)
	prevHash1 := stringPtr(event1.CurrentHash)
	event2 := types.ForcedLogoutEvent{
		EventID:      "event-002",
		Timestamp:    time.Date(2025, 10, 10, 12, 1, 0, 0, time.UTC),
		ActorID:      stringPtr("admin-123"),
		TargetUserID: "user-789",
		SessionCount: 2,
		PreviousHash: prevHash1,
	}
	event2.CurrentHash = hc.ComputeHash(&event2)

	// Event 3 (links to event2)
	prevHash2 := stringPtr(event2.CurrentHash)
	event3 := types.ForcedLogoutEvent{
		EventID:      "event-003",
		Timestamp:    time.Date(2025, 10, 10, 12, 2, 0, 0, time.UTC),
		ActorID:      stringPtr("admin-456"),
		TargetUserID: "user-111",
		SessionCount: 5,
		PreviousHash: prevHash2,
	}
	event3.CurrentHash = hc.ComputeHash(&event3)

	events := []types.ForcedLogoutEvent{event1, event2, event3}
	err := hc.ValidateChain(events)
	assert.NoError(t, err)
}

func TestValidateChain_InvalidGenesis(t *testing.T) {
	hc := NewHashChain()

	wrongGenesis := stringPtr("wrong-genesis")
	event := types.ForcedLogoutEvent{
		EventID:      "event-001",
		Timestamp:    time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC),
		ActorID:      stringPtr("admin-123"),
		TargetUserID: "user-456",
		SessionCount: 1,
		PreviousHash: wrongGenesis,
	}
	event.CurrentHash = hc.ComputeHash(&event)

	err := hc.ValidateChain([]types.ForcedLogoutEvent{event})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must have previous_hash='genesis'")
}

func TestValidateChain_BrokenChain(t *testing.T) {
	hc := NewHashChain()

	genesis := stringPtr("genesis")

	event1 := types.ForcedLogoutEvent{
		EventID:      "event-001",
		Timestamp:    time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC),
		ActorID:      stringPtr("admin-123"),
		TargetUserID: "user-456",
		SessionCount: 1,
		PreviousHash: genesis,
	}
	event1.CurrentHash = hc.ComputeHash(&event1)

	// Event 2 with WRONG previous hash
	wrongHash := stringPtr("wrong-hash-value")
	event2 := types.ForcedLogoutEvent{
		EventID:      "event-002",
		Timestamp:    time.Date(2025, 10, 10, 12, 1, 0, 0, time.UTC),
		ActorID:      stringPtr("admin-123"),
		TargetUserID: "user-789",
		SessionCount: 2,
		PreviousHash: wrongHash, // Should be event1.CurrentHash
	}
	event2.CurrentHash = hc.ComputeHash(&event2)

	events := []types.ForcedLogoutEvent{event1, event2}
	err := hc.ValidateChain(events)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chain broken")
}

func TestValidateChain_TamperedHash(t *testing.T) {
	hc := NewHashChain()

	genesis := stringPtr("genesis")

	event1 := types.ForcedLogoutEvent{
		EventID:      "event-001",
		Timestamp:    time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC),
		ActorID:      stringPtr("admin-123"),
		TargetUserID: "user-456",
		SessionCount: 1,
		PreviousHash: genesis,
	}
	event1.CurrentHash = hc.ComputeHash(&event1)

	// Tamper with event1
	event1.SessionCount = 999 // Hash no longer matches

	err := hc.ValidateChain([]types.ForcedLogoutEvent{event1})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "has invalid current_hash")
}

func TestDetectTampering_NoTampering(t *testing.T) {
	hc := NewHashChain()

	genesis := stringPtr("genesis")

	event1 := types.ForcedLogoutEvent{
		EventID:      "event-001",
		Timestamp:    time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC),
		ActorID:      stringPtr("admin-123"),
		TargetUserID: "user-456",
		SessionCount: 1,
		PreviousHash: genesis,
	}
	event1.CurrentHash = hc.ComputeHash(&event1)

	prevHash := stringPtr(event1.CurrentHash)
	event2 := types.ForcedLogoutEvent{
		EventID:      "event-002",
		Timestamp:    time.Date(2025, 10, 10, 12, 1, 0, 0, time.UTC),
		ActorID:      stringPtr("admin-123"),
		TargetUserID: "user-789",
		SessionCount: 2,
		PreviousHash: prevHash,
	}
	event2.CurrentHash = hc.ComputeHash(&event2)

	tamper := hc.DetectTampering([]types.ForcedLogoutEvent{event1, event2})
	assert.Nil(t, tamper)
}

func TestDetectTampering_InvalidGenesis(t *testing.T) {
	hc := NewHashChain()

	wrongGenesis := stringPtr("wrong-genesis")
	event := types.ForcedLogoutEvent{
		EventID:      "event-001",
		Timestamp:    time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC),
		ActorID:      stringPtr("admin-123"),
		TargetUserID: "user-456",
		SessionCount: 1,
		PreviousHash: wrongGenesis,
	}
	event.CurrentHash = hc.ComputeHash(&event)

	tamper := hc.DetectTampering([]types.ForcedLogoutEvent{event})
	assert.NotNil(t, tamper)
	assert.Equal(t, 0, tamper.EventIndex)
	assert.Equal(t, "event-001", tamper.EventID)
	assert.Equal(t, TamperTypeInvalidGenesis, tamper.TamperType)
	assert.Contains(t, tamper.Description, "genesis")
}

func TestDetectTampering_InvalidHash(t *testing.T) {
	hc := NewHashChain()

	genesis := stringPtr("genesis")
	event := types.ForcedLogoutEvent{
		EventID:      "event-001",
		Timestamp:    time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC),
		ActorID:      stringPtr("admin-123"),
		TargetUserID: "user-456",
		SessionCount: 1,
		PreviousHash: genesis,
		CurrentHash:  "fake-hash-value", // Wrong hash
	}

	tamper := hc.DetectTampering([]types.ForcedLogoutEvent{event})
	assert.NotNil(t, tamper)
	assert.Equal(t, 0, tamper.EventIndex)
	assert.Equal(t, "event-001", tamper.EventID)
	assert.Equal(t, TamperTypeInvalidHash, tamper.TamperType)
	assert.Contains(t, tamper.Description, "hash mismatch")
}

func TestDetectTampering_BrokenChain(t *testing.T) {
	hc := NewHashChain()

	genesis := stringPtr("genesis")

	event1 := types.ForcedLogoutEvent{
		EventID:      "event-001",
		Timestamp:    time.Date(2025, 10, 10, 12, 0, 0, 0, time.UTC),
		ActorID:      stringPtr("admin-123"),
		TargetUserID: "user-456",
		SessionCount: 1,
		PreviousHash: genesis,
	}
	event1.CurrentHash = hc.ComputeHash(&event1)

	wrongPrev := stringPtr("wrong-previous-hash")
	event2 := types.ForcedLogoutEvent{
		EventID:      "event-002",
		Timestamp:    time.Date(2025, 10, 10, 12, 1, 0, 0, time.UTC),
		ActorID:      stringPtr("admin-123"),
		TargetUserID: "user-789",
		SessionCount: 2,
		PreviousHash: wrongPrev, // Should link to event1.CurrentHash
	}
	event2.CurrentHash = hc.ComputeHash(&event2)

	tamper := hc.DetectTampering([]types.ForcedLogoutEvent{event1, event2})
	assert.NotNil(t, tamper)
	assert.Equal(t, 1, tamper.EventIndex)
	assert.Equal(t, "event-002", tamper.EventID)
	assert.Equal(t, TamperTypeBrokenChain, tamper.TamperType)
	assert.Contains(t, tamper.Description, "Chain broken")
}

func TestDetectTampering_EmptyChain(t *testing.T) {
	hc := NewHashChain()

	tamper := hc.DetectTampering([]types.ForcedLogoutEvent{})
	assert.Nil(t, tamper)
}

func TestPtrToStringOrEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{
			name:     "Nil pointer",
			input:    nil,
			expected: "",
		},
		{
			name:     "Non-nil pointer",
			input:    stringPtr("test-value"),
			expected: "test-value",
		},
		{
			name:     "Empty string pointer",
			input:    stringPtr(""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ptrToStringOrEmpty(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create string pointers.
func stringPtr(s string) *string {
	return &s
}
