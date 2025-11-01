package service

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/kart-io/k8s-agent/internal/orchestrator/storage"
	"github.com/kart-io/k8s-agent/internal/orchestrator/types"
	"github.com/kart-io/k8s-agent/internal/orchestrator/workflow"
	paginationv1 "github.com/kart-io/k8s-agent/pkg/api/common/pagination/v1"
	orchestratorv1 "github.com/kart-io/k8s-agent/pkg/api/orchestrator/v1"
	"github.com/kart-io/logger/core"
)

// WorkflowServiceServer implements the WorkflowService gRPC service
type WorkflowServiceServer struct {
	orchestratorv1.UnimplementedWorkflowServiceServer

	engine *workflow.Engine
	store  *storage.PostgresStore
	logger core.Logger
}

// NewWorkflowServiceServer creates a new WorkflowServiceServer
func NewWorkflowServiceServer(
	engine *workflow.Engine,
	store *storage.PostgresStore,
	logger core.Logger,
) *WorkflowServiceServer {
	return &WorkflowServiceServer{
		engine: engine,
		store:  store,
		logger: logger.With("component", "grpc-workflow-service"),
	}
}

// CreateWorkflow creates a new workflow
func (s *WorkflowServiceServer) CreateWorkflow(
	ctx context.Context,
	req *orchestratorv1.CreateWorkflowRequest,
) (*orchestratorv1.CreateWorkflowResponse, error) {
	s.logger.Infow("Creating workflow",
		"name", req.Name,
	)

	// Convert proto steps to internal steps
	steps := make([]types.WorkflowStep, 0, len(req.Steps))
	for _, protoStep := range req.Steps {
		step := types.WorkflowStep{
			ID:          protoStep.Id,
			Name:        protoStep.Name,
			Type:        convertStepType(protoStep.Type),
			Config:      protoStep.Config.AsMap(),
			OnSuccess:   protoStep.NextSteps, // Use NextSteps as OnSuccess for compatibility
			Timeout:     time.Duration(protoStep.Timeout) * time.Second,
			RetryPolicy: &types.RetryPolicy{MaxRetries: int(protoStep.RetryCount)},
		}
		// Add condition if present
		if protoStep.Condition != "" {
			step.Conditions = []types.Condition{
				{
					Field:    "condition",
					Operator: "matches",
					Value:    protoStep.Condition,
				},
			}
		}
		steps = append(steps, step)
	}

	// Convert metadata map[string]string to map[string]interface{}
	metadata := make(map[string]interface{})
	for k, v := range req.Metadata {
		metadata[k] = v
	}

	// Create workflow
	workflow := &types.Workflow{
		Name:        req.Name,
		Description: req.Description,
		Steps:       steps,
		Status:      types.WorkflowStatusDraft,
		Metadata:    metadata,
	}

	// Save workflow
	if err := s.store.SaveWorkflow(ctx, workflow); err != nil {
		s.logger.Errorw("Failed to save workflow",
			"error", err,
		)
		return nil, status.Errorf(codes.Internal, "failed to save workflow: %v", err)
	}

	s.logger.Infow("Workflow created",
		"workflow_id", workflow.ID,
		"name", workflow.Name,
	)

	// Convert to proto
	protoWorkflow, err := workflowToProto(workflow)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert workflow: %v", err)
	}

	return &orchestratorv1.CreateWorkflowResponse{
		Workflow: protoWorkflow,
	}, nil
}

// GetWorkflow retrieves a workflow by ID
func (s *WorkflowServiceServer) GetWorkflow(
	ctx context.Context,
	req *orchestratorv1.GetWorkflowRequest,
) (*orchestratorv1.GetWorkflowResponse, error) {
	s.logger.Infow("Getting workflow",
		"workflow_id", req.WorkflowId,
	)

	workflow, err := s.store.GetWorkflow(ctx, req.WorkflowId)
	if err != nil {
		s.logger.Errorw("Failed to get workflow",
			"workflow_id", req.WorkflowId,
			"error", err,
		)
		return nil, status.Errorf(codes.NotFound, "workflow not found: %v", err)
	}

	protoWorkflow, err := workflowToProto(workflow)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert workflow: %v", err)
	}

	return &orchestratorv1.GetWorkflowResponse{
		Workflow: protoWorkflow,
	}, nil
}

// ListWorkflows lists workflows
func (s *WorkflowServiceServer) ListWorkflows(
	ctx context.Context,
	req *orchestratorv1.ListWorkflowsRequest,
) (*orchestratorv1.ListWorkflowsResponse, error) {
	s.logger.Infow("Listing workflows")

	// Get workflows from storage
	workflows, err := s.store.ListWorkflows(ctx)
	if err != nil {
		s.logger.Errorw("Failed to list workflows",
			"error", err,
		)
		return nil, status.Errorf(codes.Internal, "failed to list workflows: %v", err)
	}

	// Filter by status if specified
	if req.Status != orchestratorv1.Workflow_STATUS_UNSPECIFIED {
		filtered := make([]*types.Workflow, 0)
		for _, wf := range workflows {
			if workflowStatusToProto(wf.Status) == req.Status {
				filtered = append(filtered, wf)
			}
		}
		workflows = filtered
	}

	// Convert to proto
	protoWorkflows := make([]*orchestratorv1.Workflow, 0, len(workflows))
	for _, wf := range workflows {
		protoWf, err := workflowToProto(wf)
		if err != nil {
			s.logger.Warnw("Failed to convert workflow",
				"workflow_id", wf.ID,
				"error", err,
			)
			continue
		}
		protoWorkflows = append(protoWorkflows, protoWf)
	}

	return &orchestratorv1.ListWorkflowsResponse{
		Workflows: protoWorkflows,
		Pagination: &paginationv1.PaginationMetadata{
			Total:      int64(len(protoWorkflows)),
			Page:       1,
			PageSize:   int32(len(protoWorkflows)),
			TotalPages: 1,
		},
	}, nil
}

// ExecuteWorkflow executes a workflow
func (s *WorkflowServiceServer) ExecuteWorkflow(
	ctx context.Context,
	req *orchestratorv1.ExecuteWorkflowRequest,
) (*orchestratorv1.ExecuteWorkflowResponse, error) {
	s.logger.Infow("Executing workflow",
		"workflow_id", req.WorkflowId,
		"event_id", req.EventId,
	)

	// Convert context
	triggerEvent := req.Context.AsMap()
	if req.EventId != "" {
		triggerEvent["event_id"] = req.EventId
	}

	// Start workflow execution
	execution, err := s.engine.StartWorkflow(ctx, req.WorkflowId, triggerEvent)
	if err != nil {
		s.logger.Errorw("Failed to execute workflow",
			"workflow_id", req.WorkflowId,
			"error", err,
		)
		return nil, status.Errorf(codes.Internal, "failed to execute workflow: %v", err)
	}

	s.logger.Infow("Workflow execution started",
		"execution_id", execution.ID,
		"workflow_id", req.WorkflowId,
	)

	// Convert to proto
	protoExecution, err := executionToProto(execution)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert execution: %v", err)
	}

	return &orchestratorv1.ExecuteWorkflowResponse{
		Execution: protoExecution,
	}, nil
}

// GetExecutionStatus retrieves execution status
func (s *WorkflowServiceServer) GetExecutionStatus(
	ctx context.Context,
	req *orchestratorv1.GetExecutionStatusRequest,
) (*orchestratorv1.GetExecutionStatusResponse, error) {
	s.logger.Infow("Getting execution status",
		"execution_id", req.ExecutionId,
	)

	execution, err := s.store.GetWorkflowExecution(ctx, req.ExecutionId)
	if err != nil {
		s.logger.Errorw("Failed to get execution",
			"execution_id", req.ExecutionId,
			"error", err,
		)
		return nil, status.Errorf(codes.NotFound, "execution not found: %v", err)
	}

	protoExecution, err := executionToProto(execution)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert execution: %v", err)
	}

	return &orchestratorv1.GetExecutionStatusResponse{
		Execution: protoExecution,
	}, nil
}

// Helper functions for type conversion

func convertStepType(protoType orchestratorv1.WorkflowStep_Type) types.StepType {
	switch protoType {
	case orchestratorv1.WorkflowStep_COMMAND:
		return types.StepTypeCommand
	case orchestratorv1.WorkflowStep_AI_ANALYSIS:
		return types.StepTypeAIAnalysis
	case orchestratorv1.WorkflowStep_DECISION:
		return types.StepTypeDecision
	case orchestratorv1.WorkflowStep_REMEDIATION:
		return types.StepTypeRemediation
	case orchestratorv1.WorkflowStep_NOTIFICATION:
		return types.StepTypeNotification
	case orchestratorv1.WorkflowStep_WAIT:
		return types.StepTypeWait
	default:
		return types.StepTypeCommand
	}
}

func stepTypeToProto(stepType types.StepType) orchestratorv1.WorkflowStep_Type {
	switch stepType {
	case types.StepTypeCommand:
		return orchestratorv1.WorkflowStep_COMMAND
	case types.StepTypeAIAnalysis:
		return orchestratorv1.WorkflowStep_AI_ANALYSIS
	case types.StepTypeDecision:
		return orchestratorv1.WorkflowStep_DECISION
	case types.StepTypeRemediation:
		return orchestratorv1.WorkflowStep_REMEDIATION
	case types.StepTypeNotification:
		return orchestratorv1.WorkflowStep_NOTIFICATION
	case types.StepTypeWait:
		return orchestratorv1.WorkflowStep_WAIT
	default:
		return orchestratorv1.WorkflowStep_COMMAND
	}
}

func workflowStatusToProto(status types.WorkflowStatus) orchestratorv1.Workflow_Status {
	switch status {
	case types.WorkflowStatusDraft:
		return orchestratorv1.Workflow_DRAFT
	case types.WorkflowStatusActive:
		return orchestratorv1.Workflow_ACTIVE
	case types.WorkflowStatusInactive:
		return orchestratorv1.Workflow_INACTIVE
	default:
		return orchestratorv1.Workflow_STATUS_UNSPECIFIED
	}
}

func executionStatusToProto(status types.ExecutionStatus) orchestratorv1.WorkflowExecution_Status {
	switch status {
	case types.ExecutionStatusPending:
		return orchestratorv1.WorkflowExecution_PENDING
	case types.ExecutionStatusRunning:
		return orchestratorv1.WorkflowExecution_RUNNING
	case types.ExecutionStatusCompleted:
		return orchestratorv1.WorkflowExecution_SUCCESS
	case types.ExecutionStatusFailed:
		return orchestratorv1.WorkflowExecution_FAILED
	case types.ExecutionStatusCancelled:
		return orchestratorv1.WorkflowExecution_CANCELLED
	case types.ExecutionStatusTimeout:
		return orchestratorv1.WorkflowExecution_TIMEOUT
	default:
		return orchestratorv1.WorkflowExecution_STATUS_UNSPECIFIED
	}
}

func workflowToProto(wf *types.Workflow) (*orchestratorv1.Workflow, error) {
	// Convert steps
	protoSteps := make([]*orchestratorv1.WorkflowStep, 0, len(wf.Steps))
	for _, step := range wf.Steps {
		configStruct, err := structpb.NewStruct(step.Config)
		if err != nil {
			return nil, err
		}

		// Extract condition string from Conditions slice if present
		conditionStr := ""
		if len(step.Conditions) > 0 {
			if val, ok := step.Conditions[0].Value.(string); ok {
				conditionStr = val
			}
		}

		// Get retry count from RetryPolicy
		retryCount := int32(0)
		if step.RetryPolicy != nil {
			retryCount = int32(step.RetryPolicy.MaxRetries)
		}

		protoStep := &orchestratorv1.WorkflowStep{
			Id:         step.ID,
			Name:       step.Name,
			Type:       stepTypeToProto(step.Type),
			Config:     configStruct,
			NextSteps:  step.OnSuccess, // Map OnSuccess to NextSteps
			Condition:  conditionStr,
			Timeout:    int32(step.Timeout.Seconds()),
			RetryCount: retryCount,
		}
		protoSteps = append(protoSteps, protoStep)
	}

	// Convert metadata map[string]interface{} to map[string]string
	metadata := make(map[string]string)
	for k, v := range wf.Metadata {
		if str, ok := v.(string); ok {
			metadata[k] = str
		}
	}

	return &orchestratorv1.Workflow{
		Id:          wf.ID,
		Name:        wf.Name,
		Description: wf.Description,
		Version:     "1.0", // Default version as the field doesn't exist in types.Workflow
		Steps:       protoSteps,
		Status:      workflowStatusToProto(wf.Status),
		CreatedAt:   timestamppb.New(wf.CreatedAt),
		UpdatedAt:   timestamppb.New(wf.UpdatedAt),
		Metadata:    metadata,
	}, nil
}

func executionToProto(exec *types.WorkflowExecution) (*orchestratorv1.WorkflowExecution, error) {
	contextStruct, err := structpb.NewStruct(exec.Context)
	if err != nil {
		return nil, err
	}

	// Convert Result map to string
	resultStr := ""
	if exec.Result != nil {
		if str, ok := exec.Result["result"].(string); ok {
			resultStr = str
		}
	}

	protoExec := &orchestratorv1.WorkflowExecution{
		Id:          exec.ID,
		WorkflowId:  exec.WorkflowID,
		EventId:     getEventID(exec.TriggerEvent),
		Status:      executionStatusToProto(exec.Status),
		CurrentStep: exec.CurrentStepID, // Map CurrentStepID to CurrentStep
		Context:     contextStruct,
		Result:      resultStr,
		Error:       exec.Error,
		StartedAt:   timestamppb.New(exec.StartedAt),
	}

	if exec.CompletedAt != nil {
		protoExec.CompletedAt = timestamppb.New(*exec.CompletedAt)
	}

	return protoExec, nil
}

func getEventID(triggerEvent map[string]interface{}) string {
	if triggerEvent == nil {
		return ""
	}
	if eventID, ok := triggerEvent["event_id"].(string); ok {
		return eventID
	}
	return ""
}
