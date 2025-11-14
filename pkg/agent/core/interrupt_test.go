package core

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInterruptManager(t *testing.T) {
	checkpointer := NewInMemorySaver()
	manager := NewInterruptManager(checkpointer)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.interrupts)
	assert.NotNil(t, manager.responses)
	assert.NotNil(t, manager.channels)
	assert.Equal(t, checkpointer, manager.checkpointer)
}

func TestInterruptManager_CreateAndRespond(t *testing.T) {
	manager := NewInterruptManager(nil)
	ctx := context.Background()

	// Create interrupt in goroutine
	interrupt := &Interrupt{
		Type:     InterruptTypeApproval,
		Priority: InterruptPriorityHigh,
		Message:  "Please approve this action",
		Context: map[string]interface{}{
			"action": "delete_database",
		},
	}

	// Start waiting for response
	responseChan := make(chan *InterruptResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		response, err := manager.CreateInterrupt(ctx, interrupt)
		if err != nil {
			errorChan <- err
			return
		}
		responseChan <- response
	}()

	// Wait a bit to ensure interrupt is created
	time.Sleep(50 * time.Millisecond)

	// Verify interrupt was created
	retrieved, err := manager.GetInterrupt(interrupt.ID)
	require.NoError(t, err)
	assert.Equal(t, interrupt.Type, retrieved.Type)
	assert.Equal(t, interrupt.Message, retrieved.Message)

	// Respond to interrupt
	response := &InterruptResponse{
		Approved:    true,
		Reason:      "Action approved by admin",
		RespondedBy: "admin@example.com",
	}

	err = manager.RespondToInterrupt(interrupt.ID, response)
	require.NoError(t, err)

	// Verify we received the response
	select {
	case receivedResponse := <-responseChan:
		assert.True(t, receivedResponse.Approved)
		assert.Equal(t, "Action approved by admin", receivedResponse.Reason)
	case err := <-errorChan:
		t.Fatalf("Unexpected error: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for response")
	}
}

func TestInterruptManager_RespondNotApproved(t *testing.T) {
	manager := NewInterruptManager(nil)
	ctx := context.Background()

	interrupt := &Interrupt{
		Type:     InterruptTypeApproval,
		Priority: InterruptPriorityMedium,
		Message:  "Dangerous operation",
	}

	responseChan := make(chan *InterruptResponse, 1)

	go func() {
		response, _ := manager.CreateInterrupt(ctx, interrupt)
		responseChan <- response
	}()

	time.Sleep(50 * time.Millisecond)

	// Respond with denial
	response := &InterruptResponse{
		Approved:    false,
		Reason:      "Too risky",
		RespondedBy: "reviewer@example.com",
	}

	err := manager.RespondToInterrupt(interrupt.ID, response)
	require.NoError(t, err)

	receivedResponse := <-responseChan
	assert.False(t, receivedResponse.Approved)
	assert.Equal(t, "Too risky", receivedResponse.Reason)
}

func TestInterruptManager_ListPending(t *testing.T) {
	manager := NewInterruptManager(nil)

	// Create multiple interrupts without responding
	interrupts := []*Interrupt{
		{Type: InterruptTypeApproval, Message: "Approval 1"},
		{Type: InterruptTypeInput, Message: "Input 1"},
		{Type: InterruptTypeReview, Message: "Review 1"},
	}

	for _, interrupt := range interrupts {
		manager.mu.Lock()
		if interrupt.ID == "" {
			interrupt.ID = generateInterruptID()
		}
		manager.interrupts[interrupt.ID] = interrupt
		manager.channels[interrupt.ID] = make(chan *InterruptResponse, 1)
		manager.mu.Unlock()
	}

	// List pending interrupts
	pending := manager.ListPendingInterrupts()
	assert.Len(t, pending, 3)

	// Respond to one
	err := manager.RespondToInterrupt(interrupts[0].ID, &InterruptResponse{
		Approved: true,
	})
	require.NoError(t, err)

	// List again
	pending = manager.ListPendingInterrupts()
	assert.Len(t, pending, 2)
}

func TestInterruptManager_CancelInterrupt(t *testing.T) {
	manager := NewInterruptManager(nil)
	ctx := context.Background()

	interrupt := &Interrupt{
		Type:     InterruptTypeApproval,
		Priority: InterruptPriorityLow,
		Message:  "Cancellable interrupt",
	}

	responseChan := make(chan *InterruptResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		response, err := manager.CreateInterrupt(ctx, interrupt)
		if err != nil {
			errorChan <- err
			return
		}
		responseChan <- response
	}()

	time.Sleep(50 * time.Millisecond)

	// Cancel the interrupt
	err := manager.CancelInterrupt(interrupt.ID)
	require.NoError(t, err)

	// Verify the goroutine receives a nil response due to channel closure
	select {
	case response := <-responseChan:
		// Channel was closed, so we get nil
		assert.Nil(t, response)
	case err := <-errorChan:
		// Or we get an error
		assert.NotNil(t, err)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for cancellation")
	}
}

func TestInterruptManager_WithCheckpointer(t *testing.T) {
	checkpointer := NewInMemorySaver()
	manager := NewInterruptManager(checkpointer)
	ctx := context.Background()

	state := NewAgentState()
	state.Set("important_data", "critical_value")

	interrupt := &Interrupt{
		Type:     InterruptTypeApproval,
		Priority: InterruptPriorityHigh,
		Message:  "State should be saved",
		State:    state,
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		manager.RespondToInterrupt(interrupt.ID, &InterruptResponse{
			Approved: true,
		})
	}()

	_, err := manager.CreateInterrupt(ctx, interrupt)
	require.NoError(t, err)

	// Verify state was saved
	savedState, err := checkpointer.Load(ctx, "interrupt_"+interrupt.ID)
	require.NoError(t, err)
	assert.NotNil(t, savedState)

	value, ok := savedState.Get("important_data")
	assert.True(t, ok)
	assert.Equal(t, "critical_value", value)
}

func TestInterruptManager_Hooks(t *testing.T) {
	manager := NewInterruptManager(nil)
	ctx := context.Background()

	createdCalled := false
	resolvedCalled := false

	manager.OnInterruptCreated(func(i *Interrupt) {
		createdCalled = true
		assert.Equal(t, "Hook test", i.Message)
	})

	manager.OnInterruptResolved(func(i *Interrupt, r *InterruptResponse) {
		resolvedCalled = true
		assert.True(t, r.Approved)
	})

	interrupt := &Interrupt{
		Type:    InterruptTypeApproval,
		Message: "Hook test",
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		manager.RespondToInterrupt(interrupt.ID, &InterruptResponse{
			Approved: true,
		})
	}()

	_, err := manager.CreateInterrupt(ctx, interrupt)
	require.NoError(t, err)

	assert.True(t, createdCalled, "onCreate hook should be called")
	assert.True(t, resolvedCalled, "onResolved hook should be called")
}

func TestInterruptableExecutor_AddRule(t *testing.T) {
	manager := NewInterruptManager(nil)
	executor := NewInterruptableExecutor(manager, nil)

	rule := InterruptRule{
		Name: "test_rule",
		Condition: func(ctx context.Context, state State) bool {
			return true
		},
		CreateInterrupt: func(ctx context.Context, state State) *Interrupt {
			return &Interrupt{
				Type:    InterruptTypeApproval,
				Message: "Test interrupt",
			}
		},
	}

	executor.AddInterruptRule(rule)
	assert.Len(t, executor.interruptRules, 1)
}

func TestInterruptableExecutor_CheckInterrupts(t *testing.T) {
	manager := NewInterruptManager(nil)
	executor := NewInterruptableExecutor(manager, nil)
	ctx := context.Background()
	state := NewAgentState()

	// Add a rule that always triggers
	executor.AddInterruptRule(InterruptRule{
		Name: "always_trigger",
		Condition: func(ctx context.Context, state State) bool {
			return true
		},
		CreateInterrupt: func(ctx context.Context, state State) *Interrupt {
			return &Interrupt{
				Type:     InterruptTypeApproval,
				Priority: InterruptPriorityHigh,
				Message:  "Always interrupt",
			}
		},
	})

	// Add a rule that never triggers
	executor.AddInterruptRule(InterruptRule{
		Name: "never_trigger",
		Condition: func(ctx context.Context, state State) bool {
			return false
		},
		CreateInterrupt: func(ctx context.Context, state State) *Interrupt {
			return &Interrupt{
				Type:    InterruptTypeInput,
				Message: "Never interrupt",
			}
		},
	})

	interrupts, err := executor.CheckInterrupts(ctx, state)
	require.NoError(t, err)
	assert.Len(t, interrupts, 1)
	assert.Equal(t, "Always interrupt", interrupts[0].Message)
}

func TestInterruptableExecutor_ExecuteWithInterrupts(t *testing.T) {
	manager := NewInterruptManager(nil)
	executor := NewInterruptableExecutor(manager, nil)
	ctx := context.Background()
	state := NewAgentState()
	state.Set("action", "delete")

	// Add rule that triggers on delete action
	executor.AddInterruptRule(InterruptRule{
		Name: "delete_approval",
		Condition: func(ctx context.Context, state State) bool {
			action, ok := state.Get("action")
			return ok && action == "delete"
		},
		CreateInterrupt: func(ctx context.Context, state State) *Interrupt {
			return &Interrupt{
				Type:     InterruptTypeApproval,
				Priority: InterruptPriorityCritical,
				Message:  "Approve delete action",
			}
		},
	})

	// Approve in background
	go func() {
		time.Sleep(50 * time.Millisecond)
		pending := manager.ListPendingInterrupts()
		if len(pending) > 0 {
			manager.RespondToInterrupt(pending[0].ID, &InterruptResponse{
				Approved: true,
				Reason:   "Approved by test",
			})
		}
	}()

	executed := false
	err := executor.ExecuteWithInterrupts(ctx, state, func(ctx context.Context, state State) error {
		executed = true
		return nil
	})

	require.NoError(t, err)
	assert.True(t, executed, "Function should be executed after approval")
}

func TestInterruptableExecutor_ExecuteRejected(t *testing.T) {
	manager := NewInterruptManager(nil)
	executor := NewInterruptableExecutor(manager, nil)
	ctx := context.Background()
	state := NewAgentState()
	state.Set("risky", true)

	executor.AddInterruptRule(InterruptRule{
		Name: "risky_check",
		Condition: func(ctx context.Context, state State) bool {
			risky, ok := state.Get("risky")
			return ok && risky == true
		},
		CreateInterrupt: func(ctx context.Context, state State) *Interrupt {
			return &Interrupt{
				Type:     InterruptTypeReview,
				Priority: InterruptPriorityHigh,
				Message:  "Risky operation needs review",
			}
		},
	})

	// Reject in background
	go func() {
		time.Sleep(50 * time.Millisecond)
		pending := manager.ListPendingInterrupts()
		if len(pending) > 0 {
			manager.RespondToInterrupt(pending[0].ID, &InterruptResponse{
				Approved:    false,
				Reason:      "Too risky",
				RespondedBy: "safety_officer",
			})
		}
	}()

	executed := false
	err := executor.ExecuteWithInterrupts(ctx, state, func(ctx context.Context, state State) error {
		executed = true
		return nil
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not approved")
	assert.False(t, executed, "Function should not be executed if not approved")
}

func TestGetTimeoutForInterrupt(t *testing.T) {
	tests := []struct {
		name     string
		priority InterruptPriority
		expected time.Duration
	}{
		{"critical", InterruptPriorityCritical, 5 * time.Minute},
		{"high", InterruptPriorityHigh, 15 * time.Minute},
		{"medium", InterruptPriorityMedium, 1 * time.Hour},
		{"low", InterruptPriorityLow, 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interrupt := &Interrupt{
				Priority: tt.priority,
			}
			timeout := getTimeoutForInterrupt(interrupt)
			assert.Equal(t, tt.expected, timeout)
		})
	}
}

func TestGenerateInterruptID(t *testing.T) {
	id1 := generateInterruptID()
	id2 := generateInterruptID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2, "IDs should be unique")
	assert.Contains(t, id1, "interrupt_")
}
