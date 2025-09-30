package command

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/agent-manager/internal/agent"
	"github.com/kart-io/k8s-agent/agent-manager/internal/nats"
	"github.com/kart-io/k8s-agent/agent-manager/internal/storage"
	"github.com/kart-io/k8s-agent/agent-manager/pkg/types"
)

// Dispatcher handles command dispatch and tracking
type Dispatcher struct {
	store    *storage.PostgresStore
	cache    *storage.RedisStore
	registry *agent.Registry
	nats     *nats.Server
	logger   *zap.Logger

	// Command tracking
	mu               sync.RWMutex
	pendingCommands  map[string]*types.Command
	commandTimeouts  map[string]*time.Timer

	// Metrics
	commandsIssued   int64
	commandsCompleted int64
	commandsFailed   int64
	commandsTimeout  int64
}

// NewDispatcher creates a new command dispatcher
func NewDispatcher(
	store *storage.PostgresStore,
	cache *storage.RedisStore,
	registry *agent.Registry,
	natsServer *nats.Server,
	logger *zap.Logger,
) *Dispatcher {
	return &Dispatcher{
		store:           store,
		cache:           cache,
		registry:        registry,
		nats:            natsServer,
		logger:          logger.With(zap.String("component", "command-dispatcher")),
		pendingCommands: make(map[string]*types.Command),
		commandTimeouts: make(map[string]*time.Timer),
	}
}

// DispatchCommand dispatches a command to an agent
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
		cmd.Timeout = 30 * time.Second
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
		d.updateCommandStatus(ctx, cmd.ID, types.CommandStatusFailed)
		return fmt.Errorf("failed to publish command: %w", err)
	}

	// Update status to sent
	if err := d.updateCommandStatus(ctx, cmd.ID, types.CommandStatusSent); err != nil {
		d.logger.Warn("Failed to update command status", zap.Error(err))
	}

	// Setup timeout
	d.setupCommandTimeout(cmd)

	d.mu.Lock()
	d.commandsIssued++
	d.mu.Unlock()

	d.logger.Info("Command dispatched",
		zap.String("command_id", cmd.ID),
		zap.String("cluster_id", cmd.ClusterID),
		zap.String("type", cmd.Type),
		zap.Duration("timeout", cmd.Timeout))

	return nil
}

// HandleCommandResult handles a command execution result
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
		d.logger.Warn("Failed to update command status", zap.Error(err))
	}

	// Cancel timeout timer
	d.cancelCommandTimeout(result.CommandID)

	// Remove from pending
	d.mu.Lock()
	delete(d.pendingCommands, result.CommandID)
	d.mu.Unlock()

	d.logger.Info("Command result processed",
		zap.String("command_id", result.CommandID),
		zap.String("status", result.Status),
		zap.Duration("execution_time", result.ExecutionTime))

	return nil
}

// GetCommand retrieves a command by ID
func (d *Dispatcher) GetCommand(ctx context.Context, commandID string) (*types.Command, error) {
	return d.store.GetCommand(ctx, commandID)
}

// GetCommandResult retrieves command result
func (d *Dispatcher) GetCommandResult(ctx context.Context, commandID string) (*types.CommandResult, error) {
	return d.store.GetCommandResult(ctx, commandID)
}

// validateCommand validates command before dispatch
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
	allowedTools := map[string]bool{
		"kubectl": true,
		"ps":      true,
		"df":      true,
		"netstat": true,
		"curl":    true,
		"ping":    true,
		"top":     true,
	}

	if !allowedTools[cmd.Tool] {
		return fmt.Errorf("tool '%s' is not allowed", cmd.Tool)
	}

	return nil
}

// updateCommandStatus updates command status in database
func (d *Dispatcher) updateCommandStatus(ctx context.Context, commandID string, status types.CommandStatus) error {
	return d.store.UpdateCommandStatus(ctx, commandID, status)
}

// setupCommandTimeout sets up timeout for command
func (d *Dispatcher) setupCommandTimeout(cmd *types.Command) {
	timer := time.AfterFunc(cmd.Timeout, func() {
		d.handleCommandTimeout(cmd.ID)
	})

	d.mu.Lock()
	d.commandTimeouts[cmd.ID] = timer
	d.mu.Unlock()
}

// cancelCommandTimeout cancels command timeout
func (d *Dispatcher) cancelCommandTimeout(commandID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if timer, ok := d.commandTimeouts[commandID]; ok {
		timer.Stop()
		delete(d.commandTimeouts, commandID)
	}
}

// handleCommandTimeout handles command timeout
func (d *Dispatcher) handleCommandTimeout(commandID string) {
	ctx := context.Background()

	d.logger.Warn("Command timeout", zap.String("command_id", commandID))

	// Update status
	if err := d.updateCommandStatus(ctx, commandID, types.CommandStatusTimeout); err != nil {
		d.logger.Error("Failed to update timeout status",
			zap.String("command_id", commandID),
			zap.Error(err))
	}

	// Remove from pending
	d.mu.Lock()
	delete(d.pendingCommands, commandID)
	d.commandsTimeout++
	d.mu.Unlock()
}

// GetPendingCommands returns all pending commands
func (d *Dispatcher) GetPendingCommands() []*types.Command {
	d.mu.RLock()
	defer d.mu.RUnlock()

	commands := make([]*types.Command, 0, len(d.pendingCommands))
	for _, cmd := range d.pendingCommands {
		commands = append(commands, cmd)
	}

	return commands
}

// GetStatistics returns dispatcher statistics
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