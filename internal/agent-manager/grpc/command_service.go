package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/kart-io/k8s-agent/internal/agent-manager/command"
	"github.com/kart-io/k8s-agent/internal/agent-manager/storage"
	agentv1 "github.com/kart-io/k8s-agent/pkg/api/agent/v1"
	"github.com/kart-io/k8s-agent/pkg/types"
	"github.com/kart-io/logger/core"
)

// CommandServiceServer implements the CommandService gRPC service.
type CommandServiceServer struct {
	agentv1.UnimplementedCommandServiceServer
	dispatcher *command.Dispatcher
	store      *storage.MySQLStore
	logger     core.Logger
}

// NewCommandServiceServer creates a new CommandServiceServer instance.
func NewCommandServiceServer(dispatcher *command.Dispatcher, store *storage.MySQLStore, logger core.Logger) *CommandServiceServer {
	return &CommandServiceServer{
		dispatcher: dispatcher,
		store:      store,
		logger:     logger,
	}
}

// ExecuteCommand dispatches a command for execution.
func (s *CommandServiceServer) ExecuteCommand(ctx context.Context, req *agentv1.ExecuteCommandRequest) (*agentv1.ExecuteCommandResponse, error) {
	s.logger.Infow("Executing command",
		"agent_id", req.AgentId,
		"type", req.Type,
		"content", req.Content,
	)

	// Validate request
	if req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "agent_id is required")
	}
	if req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "content is required")
	}

	// Create command
	cmd := &types.Command{
		ClusterID: req.AgentId, // Using agent_id as cluster_id for now
		Type:      convertProtoCommandType(req.Type),
		Tool:      "kubectl", // Default tool
		Action:    req.Content,
		Args:      []string{},
		Timeout:   time.Duration(req.Timeout) * time.Second,
	}

	// Add params if provided
	if len(req.Params) > 0 {
		cmd.Metadata = make(map[string]interface{})
		for k, v := range req.Params {
			cmd.Metadata[k] = v
		}
	}

	// Dispatch command
	if err := s.dispatcher.DispatchCommand(ctx, cmd); err != nil {
		s.logger.Errorw("Failed to dispatch command",
			"error", err,
			"command_id", cmd.ID,
		)
		return nil, status.Errorf(codes.Internal, "failed to dispatch command: %v", err)
	}

	// Convert to protobuf
	pbCommand := convertCommandToProto(cmd)

	s.logger.Infow("Command dispatched successfully",
		"command_id", cmd.ID,
		"agent_id", req.AgentId,
	)

	return &agentv1.ExecuteCommandResponse{
		Command: pbCommand,
	}, nil
}

// GetCommandStatus retrieves the status of a command.
func (s *CommandServiceServer) GetCommandStatus(ctx context.Context, req *agentv1.GetCommandStatusRequest) (*agentv1.GetCommandStatusResponse, error) {
	if req.CommandId == "" {
		return nil, status.Error(codes.InvalidArgument, "command_id is required")
	}

	// Get command from storage
	cmd, err := s.store.GetCommand(ctx, req.CommandId)
	if err != nil {
		s.logger.Warnw("Failed to get command",
			"error", err,
			"command_id", req.CommandId,
		)
		return nil, status.Errorf(codes.NotFound, "command not found: %v", err)
	}

	// Convert to protobuf
	pbCommand := convertCommandToProto(cmd)

	return &agentv1.GetCommandStatusResponse{
		Command: pbCommand,
	}, nil
}

// CancelCommand cancels a pending or running command.
func (s *CommandServiceServer) CancelCommand(ctx context.Context, req *agentv1.CancelCommandRequest) (*agentv1.CancelCommandResponse, error) {
	if req.CommandId == "" {
		return nil, status.Error(codes.InvalidArgument, "command_id is required")
	}

	// Get command from storage
	cmd, err := s.store.GetCommand(ctx, req.CommandId)
	if err != nil {
		s.logger.Warnw("Failed to get command for cancellation",
			"error", err,
			"command_id", req.CommandId,
		)
		return nil, status.Errorf(codes.NotFound, "command not found: %v", err)
	}

	// Check if command is in a cancellable state
	if cmd.Status != types.CommandStatusPending && cmd.Status != types.CommandStatusSent {
		return nil, status.Errorf(codes.FailedPrecondition, "command cannot be cancelled in current status: %s", cmd.Status)
	}

	// Update command status to cancelled (simplified - should use dispatcher method)
	cmd.Status = types.CommandStatusFailed
	cmd.UpdatedAt = time.Now()
	if err := s.store.SaveCommand(ctx, cmd); err != nil {
		s.logger.Errorw("Failed to update command status",
			"error", err,
			"command_id", req.CommandId,
		)
		return nil, status.Errorf(codes.Internal, "failed to cancel command: %v", err)
	}

	s.logger.Infow("Command cancelled",
		"command_id", req.CommandId,
	)

	return &agentv1.CancelCommandResponse{
		Success: true,
	}, nil
}

// Helper functions

// convertCommandToProto converts types.Command to protobuf Command.
func convertCommandToProto(cmd *types.Command) *agentv1.Command {
	if cmd == nil {
		return nil
	}

	// Convert metadata to params
	params := make(map[string]string)
	for k, v := range cmd.Metadata {
		if str, ok := v.(string); ok {
			params[k] = str
		}
	}

	return &agentv1.Command{
		Id:          cmd.ID,
		AgentId:     cmd.ClusterID,
		Type:        convertCommandTypeToProto(cmd.Type),
		Content:     cmd.Action,
		Params:      params,
		Status:      convertCommandStatusToProto(cmd.Status),
		Result:      "", // Result would come from CommandResult table
		Error:       "", // Error would come from CommandResult table
		CreatedAt:   timestamppb.New(cmd.CreatedAt),
		StartedAt:   nil, // Would need to track this separately
		CompletedAt: nil, // Would need to track this separately
		Timeout:     int32(cmd.Timeout.Seconds()),
	}
}

// convertCommandTypeToProto converts command type string to protobuf Type.
func convertCommandTypeToProto(cmdType string) agentv1.Command_Type {
	switch cmdType {
	case "kubectl":
		return agentv1.Command_KUBECTL
	case "diagnostic":
		return agentv1.Command_DIAGNOSTIC
	case "remediation":
		return agentv1.Command_REMEDIATION
	case "custom":
		return agentv1.Command_CUSTOM
	default:
		return agentv1.Command_TYPE_UNSPECIFIED
	}
}

// convertProtoCommandType converts protobuf Type to command type string.
func convertProtoCommandType(cmdType agentv1.Command_Type) string {
	switch cmdType {
	case agentv1.Command_KUBECTL:
		return "kubectl"
	case agentv1.Command_DIAGNOSTIC:
		return "diagnostic"
	case agentv1.Command_REMEDIATION:
		return "remediation"
	case agentv1.Command_CUSTOM:
		return "custom"
	default:
		return "unknown"
	}
}

// convertCommandStatusToProto converts types.CommandStatus to protobuf Status.
func convertCommandStatusToProto(cmdStatus types.CommandStatus) agentv1.Command_Status {
	switch cmdStatus {
	case types.CommandStatusPending:
		return agentv1.Command_PENDING
	case types.CommandStatusSent:
		return agentv1.Command_RUNNING
	case types.CommandStatusExecuting:
		return agentv1.Command_RUNNING
	case types.CommandStatusCompleted:
		return agentv1.Command_SUCCESS
	case types.CommandStatusFailed:
		return agentv1.Command_FAILED
	case types.CommandStatusTimeout:
		return agentv1.Command_TIMEOUT
	default:
		return agentv1.Command_STATUS_UNSPECIFIED
	}
}
