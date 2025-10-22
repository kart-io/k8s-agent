package services

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/kart-io/logger/core"

	commandv1 "github.com/kart-io/k8s-agent/api/proto/gen/agentmanager/command/v1"
	"github.com/kart-io/k8s-agent/pkg/types"
)

// CommandServiceServer 命令服务实现
type CommandServiceServer struct {
	commandv1.UnimplementedCommandServiceServer
	logger     core.Logger
	dispatcher CommandDispatcher
}

// CommandDispatcher 定义命令调度器接口
type CommandDispatcher interface {
	DispatchCommand(ctx context.Context, cmd *types.Command) error
	GetCommand(ctx context.Context, commandID string) (*types.Command, error)
	GetCommandResult(ctx context.Context, commandID string) (*types.CommandResult, error)
}

// NewCommandServiceServer 创建命令服务
func NewCommandServiceServer(logger core.Logger, dispatcher CommandDispatcher) *CommandServiceServer {
	return &CommandServiceServer{
		logger:     logger.With("service", "command"),
		dispatcher: dispatcher,
	}
}

// ExecuteCommand 执行命令
func (s *CommandServiceServer) ExecuteCommand(ctx context.Context, req *commandv1.ExecuteCommandRequest) (*commandv1.Command, error) {
	s.logger.Infow("ExecuteCommand called",
		"cluster_id", req.ClusterId,
		"type", req.Type.String(),
		"tool", req.Tool,
	)

	if req.ClusterId == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster_id is required")
	}
	if req.Tool == "" {
		return nil, status.Error(codes.InvalidArgument, "tool is required")
	}

	// 创建命令对象
	var timeout time.Duration
	if req.Timeout != nil {
		timeout = req.Timeout.AsDuration()
	}

	cmd := &types.Command{
		ClusterID: req.ClusterId,
		Tool:      req.Tool,
		Args:      req.Args,
		Status:    types.CommandStatusPending,
		Timeout:   timeout,
	}

	// 分发命令
	if err := s.dispatcher.DispatchCommand(ctx, cmd); err != nil {
		s.logger.Errorw("Failed to dispatch command", "error", err)
		return nil, status.Error(codes.Internal, "failed to dispatch command")
	}

	s.logger.Infow("Command dispatched successfully", "command_id", cmd.ID)
	return convertTypesToProtoCommand(cmd), nil
}

// GetCommand 获取命令详情
func (s *CommandServiceServer) GetCommand(ctx context.Context, req *commandv1.GetCommandRequest) (*commandv1.Command, error) {
	s.logger.Infow("GetCommand called", "command_id", req.CommandId)

	if req.CommandId == "" {
		return nil, status.Error(codes.InvalidArgument, "command_id is required")
	}

	// 从dispatcher获取命令
	cmd, err := s.dispatcher.GetCommand(ctx, req.CommandId)
	if err != nil {
		s.logger.Errorw("Failed to get command", "command_id", req.CommandId, "error", err)
		return nil, status.Error(codes.NotFound, "command not found")
	}

	return convertTypesToProtoCommand(cmd), nil
}

// ListCommands 列出命令
func (s *CommandServiceServer) ListCommands(ctx context.Context, req *commandv1.ListCommandsRequest) (*commandv1.ListCommandsResponse, error) {
	s.logger.Infow("ListCommands called",
		"cluster_id", req.ClusterId,
		"agent_id", req.AgentId,
	)

	// TODO: 从存储层获取命令列表
	return &commandv1.ListCommandsResponse{
		Commands:      []*commandv1.Command{},
		NextPageToken: "",
		TotalCount:    0,
	}, nil
}

// CancelCommand 取消命令
func (s *CommandServiceServer) CancelCommand(ctx context.Context, req *commandv1.CancelCommandRequest) (*commandv1.Command, error) {
	s.logger.Infow("CancelCommand called", "command_id", req.CommandId)

	if req.CommandId == "" {
		return nil, status.Error(codes.InvalidArgument, "command_id is required")
	}

	// TODO: 实现命令取消逻辑
	return nil, status.Error(codes.Unimplemented, "not implemented yet")
}

// GetCommandResult 获取命令结果
func (s *CommandServiceServer) GetCommandResult(ctx context.Context, req *commandv1.GetCommandResultRequest) (*commandv1.CommandResult, error) {
	s.logger.Infow("GetCommandResult called", "command_id", req.CommandId)

	if req.CommandId == "" {
		return nil, status.Error(codes.InvalidArgument, "command_id is required")
	}

	// 从dispatcher获取命令结果
	result, err := s.dispatcher.GetCommandResult(ctx, req.CommandId)
	if err != nil {
		s.logger.Errorw("Failed to get command result", "command_id", req.CommandId, "error", err)
		return nil, status.Error(codes.NotFound, "command result not found")
	}

	return convertTypesToProtoCommandResult(result), nil
}

// StreamCommandOutput 流式获取命令输出
func (s *CommandServiceServer) StreamCommandOutput(req *commandv1.StreamCommandOutputRequest, stream commandv1.CommandService_StreamCommandOutputServer) error {
	s.logger.Infow("StreamCommandOutput called", "command_id", req.CommandId)

	if req.CommandId == "" {
		return status.Error(codes.InvalidArgument, "command_id is required")
	}

	// TODO: 实现流式输出逻辑
	return status.Error(codes.Unimplemented, "not implemented yet")
}

// Type conversion functions

// convertTypesToProtoCommand converts types.Command to commandv1.Command
func convertTypesToProtoCommand(cmd *types.Command) *commandv1.Command {
	if cmd == nil {
		return nil
	}

	protoCmd := &commandv1.Command{
		CommandId: cmd.ID,
		ClusterId: cmd.ClusterID,
		Tool:      cmd.Tool,
		Args:      cmd.Args,
		Status:    convertCommandStatusToProto(cmd.Status),
	}

	if !cmd.CreatedAt.IsZero() {
		protoCmd.CreatedAt = timestamppb.New(cmd.CreatedAt)
	}
	if cmd.Timeout > 0 {
		protoCmd.Timeout = durationpb.New(cmd.Timeout)
	}

	return protoCmd
}

// convertTypesToProtoCommandResult converts types.CommandResult to commandv1.CommandResult
func convertTypesToProtoCommandResult(result *types.CommandResult) *commandv1.CommandResult {
	if result == nil {
		return nil
	}

	protoResult := &commandv1.CommandResult{
		CommandId: result.CommandID,
		ExitCode:  int32(result.ExitCode),
		Stdout:    result.Output,
		Stderr:    result.Error,
	}

	return protoResult
}

// convertCommandStatusToProto converts types.CommandStatus to commandv1.CommandStatus
func convertCommandStatusToProto(status types.CommandStatus) commandv1.CommandStatus {
	switch status {
	case types.CommandStatusPending:
		return commandv1.CommandStatus_COMMAND_STATUS_PENDING
	case types.CommandStatusSent:
		return commandv1.CommandStatus_COMMAND_STATUS_PENDING // Map sent to pending
	case types.CommandStatusExecuting:
		return commandv1.CommandStatus_COMMAND_STATUS_RUNNING
	case types.CommandStatusCompleted:
		return commandv1.CommandStatus_COMMAND_STATUS_COMPLETED
	case types.CommandStatusFailed:
		return commandv1.CommandStatus_COMMAND_STATUS_FAILED
	case types.CommandStatusTimeout:
		return commandv1.CommandStatus_COMMAND_STATUS_TIMEOUT
	default:
		return commandv1.CommandStatus_COMMAND_STATUS_UNSPECIFIED
	}
}
