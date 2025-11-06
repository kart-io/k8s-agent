package command

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/kart-io/k8s-agent/internal/agent-manager/agent"
	"github.com/kart-io/k8s-agent/internal/agent-manager/constants"
	"github.com/kart-io/k8s-agent/internal/agent-manager/nats"
	"github.com/kart-io/k8s-agent/internal/agent-manager/storage"
	"github.com/kart-io/k8s-agent/pkg/types"
	"github.com/kart-io/logger/core"
)

// Dispatcher handles command dispatch and tracking.
type Dispatcher struct {
	store    *storage.MySQLStore
	cache    *storage.RedisStore
	registry *agent.Registry
	nats     *nats.Server
	logger   core.Logger

	// Command tracking
	mu              sync.RWMutex
	pendingCommands map[string]*types.Command
	commandTimeouts map[string]*time.Timer

	// Cleanup
	stopCh chan struct{}
	wg     sync.WaitGroup

	// Metrics
	commandsIssued    int64
	commandsCompleted int64
	commandsFailed    int64
	commandsTimeout   int64
}

// NewDispatcher creates a new command dispatcher.
func NewDispatcher(
	store *storage.MySQLStore,
	cache *storage.RedisStore,
	registry *agent.Registry,
	natsServer *nats.Server,
	logger core.Logger,
) *Dispatcher {
	d := &Dispatcher{
		store:           store,
		cache:           cache,
		registry:        registry,
		nats:            natsServer,
		logger:          logger.With("component", "command-dispatcher"),
		pendingCommands: make(map[string]*types.Command),
		commandTimeouts: make(map[string]*time.Timer),
		stopCh:          make(chan struct{}),
	}

	// Start cleanup goroutine
	d.wg.Add(1)
	go d.cleanupExpiredTimers()

	return d
}

// Stop stops the dispatcher.
func (d *Dispatcher) Stop() error {
	close(d.stopCh)
	d.wg.Wait()

	// Cancel all pending timers
	d.mu.Lock()
	defer d.mu.Unlock()

	for id, timer := range d.commandTimeouts {
		timer.Stop()
		delete(d.commandTimeouts, id)
	}

	return nil
}

// cleanupExpiredTimers periodically cleans up stale timers.
func (d *Dispatcher) cleanupExpiredTimers() {
	defer d.wg.Done()
	defer func() {
		if rec := recover(); rec != nil {
			d.logger.Errorw("Panic in timer cleanup", "panic", rec)
		}
	}()

	ticker := time.NewTicker(constants.CommandTimeoutCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-d.stopCh:
			return
		case <-ticker.C:
			d.performTimerCleanup()
		}
	}
}

// performTimerCleanup removes stopped timers from map.
func (d *Dispatcher) performTimerCleanup() {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Create a copy to avoid modifying map during iteration
	toDelete := []string{}

	for id, timer := range d.commandTimeouts {
		// Check if timer has already fired (Stop returns false)
		if !timer.Stop() {
			toDelete = append(toDelete, id)
		}
	}

	// Clean up
	for _, id := range toDelete {
		delete(d.commandTimeouts, id)
		d.logger.Debugw("Cleaned up expired timer", "command_id", id)
	}

	if len(toDelete) > 0 {
		d.logger.Infow("Timer cleanup completed", "cleaned", len(toDelete), "remaining", len(d.commandTimeouts))
	}
}

// DispatchCommand dispatches a command to an agent.
func (d *Dispatcher) DispatchCommand(ctx context.Context, cmd *types.Command) error {
	// Validate command
	if err := d.validateCommand(cmd); err != nil {
		return fmt.Errorf("command validation failed: %w", err)
	}

	// Generate command ID if not present
	if cmd.ID == "" {
		cmd.ID = uuid.New().String()
	}

	// Set default timeout
	if cmd.Timeout == 0 {
		cmd.Timeout = constants.DefaultCommandTimeout
	}

	// Set metadata
	cmd.Status = types.CommandStatusPending
	cmd.CreatedAt = time.Now()
	cmd.UpdatedAt = time.Now()

	// Verify target agent is online
	targetAgent, err := d.registry.GetAgentByClusterID(ctx, cmd.ClusterID)
	if err != nil {
		return fmt.Errorf("target cluster not found: %w", err)
	}

	if targetAgent.Status != types.AgentStatusOnline {
		return fmt.Errorf("target agent is offline")
	}

	// Save command to database
	if err := d.store.SaveCommand(ctx, cmd); err != nil {
		return fmt.Errorf("failed to save command: %w", err)
	}

	// Track command
	d.mu.Lock()
	d.pendingCommands[cmd.ID] = cmd
	d.mu.Unlock()

	// Publish command via NATS
	if err := d.nats.PublishCommand(cmd.ClusterID, cmd); err != nil {
		// Update status to failed
		if updateErr := d.updateCommandStatus(ctx, cmd.ID, types.CommandStatusFailed); updateErr != nil {
			d.logger.Warnw("Failed to update command status", "command_id", cmd.ID, "error", updateErr)
		}
		return fmt.Errorf("failed to publish command: %w", err)
	}

	// Update status to sent
	if err := d.updateCommandStatus(ctx, cmd.ID, types.CommandStatusSent); err != nil {
		d.logger.Warnw("Failed to update command status", "error", err)
	}

	// Setup timeout
	d.setupCommandTimeout(cmd)

	d.mu.Lock()
	d.commandsIssued++
	d.mu.Unlock()

	d.logger.Infow("Command dispatched",
		"command_id", cmd.ID,
		"cluster_id", cmd.ClusterID,
		"type", cmd.Type,
		"timeout", cmd.Timeout)

	return nil
}

// HandleCommandResult handles a command execution result.
func (d *Dispatcher) HandleCommandResult(ctx context.Context, result *types.CommandResult) error {
	// Save result to database
	if err := d.store.SaveCommandResult(ctx, result); err != nil {
		return fmt.Errorf("failed to save command result: %w", err)
	}

	// Update command status
	var status types.CommandStatus
	if result.Status == "success" {
		status = types.CommandStatusCompleted
		d.mu.Lock()
		d.commandsCompleted++
		d.mu.Unlock()
	} else {
		status = types.CommandStatusFailed
		d.mu.Lock()
		d.commandsFailed++
		d.mu.Unlock()
	}

	if err := d.updateCommandStatus(ctx, result.CommandID, status); err != nil {
		d.logger.Warnw("Failed to update command status", "error", err)
	}

	// Cancel timeout timer
	d.cancelCommandTimeout(result.CommandID)

	// Remove from pending
	d.mu.Lock()
	delete(d.pendingCommands, result.CommandID)
	d.mu.Unlock()

	d.logger.Infow("Command result processed",
		"command_id", result.CommandID,
		"status", result.Status,
		"execution_time", result.ExecutionTime)

	return nil
}

// GetCommand retrieves a command by ID.
func (d *Dispatcher) GetCommand(ctx context.Context, commandID string) (*types.Command, error) {
	return d.store.GetCommand(ctx, commandID)
}

// GetCommandResult retrieves command result.
func (d *Dispatcher) GetCommandResult(ctx context.Context, commandID string) (*types.CommandResult, error) {
	return d.store.GetCommandResult(ctx, commandID)
}

// validateCommand validates command before dispatch.
func (d *Dispatcher) validateCommand(cmd *types.Command) error {
	if cmd.ClusterID == "" {
		return fmt.Errorf("cluster_id is required")
	}

	if cmd.Type == "" {
		return fmt.Errorf("command type is required")
	}

	if cmd.Tool == "" {
		return fmt.Errorf("tool is required")
	}

	if cmd.Action == "" {
		return fmt.Errorf("action is required")
	}

	// Validate tool whitelist
	if !constants.AllowedTools[cmd.Tool] {
		return fmt.Errorf("tool '%s' is not allowed", cmd.Tool)
	}

	// Additional validation for kubectl
	if cmd.Tool == "kubectl" {
		if !constants.AllowedKubectlActions[cmd.Action] {
			return fmt.Errorf("kubectl action '%s' is not allowed", cmd.Action)
		}
	}

	return nil
}

// updateCommandStatus updates command status in database.
func (d *Dispatcher) updateCommandStatus(ctx context.Context, commandID string, status types.CommandStatus) error {
	return d.store.UpdateCommandStatus(ctx, commandID, status)
}

// setupCommandTimeout sets up timeout for command.
func (d *Dispatcher) setupCommandTimeout(cmd *types.Command) {
	timer := time.AfterFunc(cmd.Timeout, func() {
		d.handleCommandTimeout(cmd.ID)
	})

	d.mu.Lock()
	d.commandTimeouts[cmd.ID] = timer
	d.mu.Unlock()
}

// cancelCommandTimeout cancels command timeout.
func (d *Dispatcher) cancelCommandTimeout(commandID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if timer, ok := d.commandTimeouts[commandID]; ok {
		timer.Stop()
		delete(d.commandTimeouts, commandID)
	}
}

// handleCommandTimeout handles command timeout.
func (d *Dispatcher) handleCommandTimeout(commandID string) {
	ctx := context.Background()

	d.logger.Warnw("Command timeout", "command_id", commandID)

	// Update status
	if err := d.updateCommandStatus(ctx, commandID, types.CommandStatusTimeout); err != nil {
		d.logger.Errorw("Failed to update timeout status",
			"command_id", commandID,
			"error", err)
	}

	// Remove from pending
	d.mu.Lock()
	delete(d.pendingCommands, commandID)
	d.commandsTimeout++
	d.mu.Unlock()
}

// GetPendingCommands returns all pending commands.
func (d *Dispatcher) GetPendingCommands() []*types.Command {
	d.mu.RLock()
	defer d.mu.RUnlock()

	commands := make([]*types.Command, 0, len(d.pendingCommands))
	for _, cmd := range d.pendingCommands {
		commands = append(commands, cmd)
	}

	return commands
}

// GetStatistics returns dispatcher statistics.
func (d *Dispatcher) GetStatistics() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return map[string]interface{}{
		"commands_issued":    d.commandsIssued,
		"commands_completed": d.commandsCompleted,
		"commands_failed":    d.commandsFailed,
		"commands_timeout":   d.commandsTimeout,
		"pending_commands":   len(d.pendingCommands),
	}
}
