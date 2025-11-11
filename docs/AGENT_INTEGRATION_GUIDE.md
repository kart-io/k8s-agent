# Agent Integration Guide

This guide shows how to integrate the new Agent framework into orchestrator and agent-manager services.

## Orchestrator Service Integration

### Step 1: Update Service Initialization

```go
// internal/orchestrator/startup/service.go

import (
    workflowagent "github.com/kart-io/k8s-agent/internal/orchestrator/agents/workflow"
    "github.com/kart-io/k8s-agent/pkg/agent/distributed"
    "github.com/kart-io/k8s-agent/pkg/agent/observability"
)

type OrchestratorService struct {
    // Existing fields
    executor *workflow.Executor

    // New Agent fields
    workflowAgent    agentcore.Agent
    stepAgent        agentcore.Agent
    agentRegistry    *distributed.Registry
    agentClient      *distributed.Client
    agentCoordinator *distributed.Coordinator
}

func NewOrchestratorService(
    executor *workflow.Executor,
    logger core.Logger,
) *OrchestratorService {
    // Create agents
    workflowAgent := workflowagent.NewWorkflowAgent(executor, logger)
    stepAgent := workflowagent.NewStepAgent(executor, logger)

    // Wrap with observability
    workflowAgent = observability.WrapAgent(workflowAgent, "orchestrator", logger)
    stepAgent = observability.WrapAgent(stepAgent, "orchestrator", logger)

    // Create distributed infrastructure
    registry := distributed.NewRegistry(logger)
    client := distributed.NewClient(logger)
    coordinator := distributed.NewCoordinator(registry, client, logger)

    // Register this orchestrator instance
    registry.Register(&distributed.ServiceInstance{
        ID:          "orchestrator-1",
        ServiceName: "orchestrator",
        Endpoint:    "http://localhost:8081",
        Agents:      []string{"workflow-agent", "step-agent"},
    })

    return &OrchestratorService{
        executor:         executor,
        workflowAgent:    workflowAgent,
        stepAgent:        stepAgent,
        agentRegistry:    registry,
        agentClient:      client,
        agentCoordinator: coordinator,
    }
}
```

### Step 2: Add Agent API Endpoints

```go
// internal/orchestrator/api/agent_handler.go

package api

import (
    "github.com/gin-gonic/gin"
    agentcore "github.com/kart-io/k8s-agent/pkg/agent/core"
)

type AgentHandler struct {
    workflowAgent agentcore.Agent
    stepAgent     agentcore.Agent
}

// POST /api/v1/agents/:name/execute
func (h *AgentHandler) ExecuteAgent(c *gin.Context) {
    agentName := c.Param("name")

    var input agentcore.AgentInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    var agent agentcore.Agent
    switch agentName {
    case "workflow-agent":
        agent = h.workflowAgent
    case "step-agent":
        agent = h.stepAgent
    default:
        c.JSON(404, gin.H{"error": "agent not found"})
        return
    }

    output, err := agent.Execute(c.Request.Context(), &input)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, output)
}

// GET /api/v1/agents
func (h *AgentHandler) ListAgents(c *gin.Context) {
    agents := []map[string]interface{}{
        {
            "name":         "workflow-agent",
            "description":  h.workflowAgent.Description(),
            "capabilities": h.workflowAgent.Capabilities(),
        },
        {
            "name":         "step-agent",
            "description":  h.stepAgent.Description(),
            "capabilities": h.stepAgent.Capabilities(),
        },
    }

    c.JSON(200, gin.H{"agents": agents})
}

// Register routes
func RegisterAgentRoutes(r *gin.RouterGroup, handler *AgentHandler) {
    r.POST("/agents/:name/execute", handler.ExecuteAgent)
    r.GET("/agents", handler.ListAgents)
}
```

### Step 3: Use Agents in Workflow Execution

```go
// internal/orchestrator/service/workflow_service.go

func (s *WorkflowService) ExecuteWorkflow(ctx context.Context, workflowID string) error {
    // Load workflow
    workflow, err := s.store.GetWorkflow(ctx, workflowID)
    if err != nil {
        return err
    }

    // Create execution
    execution := &types.WorkflowExecution{
        ID:           uuid.New().String(),
        WorkflowID:   workflowID,
        Status:       "running",
        Context:      make(map[string]interface{}),
        TriggerEvent: make(map[string]interface{}),
        CreatedAt:    time.Now(),
    }

    // Execute via WorkflowAgent
    input := &agentcore.AgentInput{
        Task:        "Execute workflow",
        Instruction: workflowID,
        Context: map[string]interface{}{
            "execution": execution,
            "steps":     workflow.Steps,
        },
        Options: agentcore.AgentOptions{
            Timeout: 300 * time.Second,
        },
    }

    output, err := s.workflowAgent.Execute(ctx, input)
    if err != nil {
        execution.Status = "failed"
        execution.Error = err.Error()
    } else {
        execution.Status = output.Status
        execution.Result = output.Result
    }

    return s.store.SaveExecution(ctx, execution)
}
```

## Agent-Manager Service Integration

### Step 1: Update Service Initialization

```go
// internal/agent-manager/startup/service.go

import (
    commandagent "github.com/kart-io/k8s-agent/internal/agent-manager/agents/command"
    "github.com/kart-io/k8s-agent/pkg/agent/distributed"
    "github.com/kart-io/k8s-agent/pkg/agent/observability"
)

type AgentManagerService struct {
    // Existing fields
    dispatcher *command.Dispatcher

    // New Agent fields
    commandAgent     agentcore.Agent
    kubectlAgent     agentcore.Agent
    agentRegistry    *distributed.Registry
    agentClient      *distributed.Client
    agentCoordinator *distributed.Coordinator
}

func NewAgentManagerService(
    dispatcher *command.Dispatcher,
    logger core.Logger,
) *AgentManagerService {
    // Create agents
    commandAgent := commandagent.NewCommandAgent(dispatcher, logger)
    kubectlAgent := commandagent.NewKubectlAgent(dispatcher, logger)

    // Wrap with observability
    commandAgent = observability.WrapAgent(commandAgent, "agent-manager", logger)
    kubectlAgent = observability.WrapAgent(kubectlAgent, "agent-manager", logger)

    // Create distributed infrastructure
    registry := distributed.NewRegistry(logger)
    client := distributed.NewClient(logger)
    coordinator := distributed.NewCoordinator(registry, client, logger)

    // Register this instance
    registry.Register(&distributed.ServiceInstance{
        ID:          "agent-manager-1",
        ServiceName: "agent-manager",
        Endpoint:    "http://localhost:8080",
        Agents:      []string{"command-agent", "kubectl-agent"},
    })

    return &AgentManagerService{
        dispatcher:       dispatcher,
        commandAgent:     commandAgent,
        kubectlAgent:     kubectlAgent,
        agentRegistry:    registry,
        agentClient:      client,
        agentCoordinator: coordinator,
    }
}
```

### Step 2: Add Agent API Endpoints

```go
// internal/agent-manager/api/agent_handler.go

package api

import (
    "github.com/gin-gonic/gin"
    agentcore "github.com/kart-io/k8s-agent/pkg/agent/core"
)

type AgentHandler struct {
    commandAgent agentcore.Agent
    kubectlAgent agentcore.Agent
}

// POST /api/v1/agents/:name/execute
func (h *AgentHandler) ExecuteAgent(c *gin.Context) {
    agentName := c.Param("name")

    var input agentcore.AgentInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    var agent agentcore.Agent
    switch agentName {
    case "command-agent":
        agent = h.commandAgent
    case "kubectl-agent":
        agent = h.kubectlAgent
    default:
        c.JSON(404, gin.H{"error": "agent not found"})
        return
    }

    output, err := agent.Execute(c.Request.Context(), &input)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, output)
}

// GET /api/v1/agents
func (h *AgentHandler) ListAgents(c *gin.Context) {
    agents := []map[string]interface{}{
        {
            "name":         "command-agent",
            "description":  h.commandAgent.Description(),
            "capabilities": h.commandAgent.Capabilities(),
        },
        {
            "name":         "kubectl-agent",
            "description":  h.kubectlAgent.Description(),
            "capabilities": h.kubectlAgent.Capabilities(),
        },
    }

    c.JSON(200, gin.H{"agents": agents})
}
```

### Step 3: Use Agents in Command Handling

```go
// internal/agent-manager/api/command_handler.go

func (h *CommandHandler) ExecuteCommand(c *gin.Context) {
    var req CommandRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Create command
    cmd := &types.Command{
        ClusterID: req.ClusterID,
        Type:      req.Type,
        Tool:      req.Tool,
        Action:    req.Action,
        Args:      req.Args,
    }

    // Execute via CommandAgent (or KubectlAgent for kubectl)
    var agent agentcore.Agent
    if cmd.Tool == "kubectl" {
        agent = h.kubectlAgent
    } else {
        agent = h.commandAgent
    }

    input := &agentcore.AgentInput{
        Task: "Execute command",
        Context: map[string]interface{}{
            "command": cmd,
        },
        Options: agentcore.AgentOptions{
            Timeout: 60 * time.Second,
        },
    }

    output, err := agent.Execute(c.Request.Context(), input)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, output)
}
```

## Cross-Service Integration

### Example: Orchestrator calls Agent-Manager

```go
// In orchestrator service

func (s *OrchestratorService) ExecuteKubectlCommand(
    ctx context.Context,
    clusterID, action string,
    args []string,
) (*agentcore.AgentOutput, error) {
    // Prepare input
    input := &agentcore.AgentInput{
        Task: "Execute kubectl command",
        Context: map[string]interface{}{
            "cluster_id": clusterID,
            "action":     action,
            "args":       args,
        },
        Options: agentcore.AgentOptions{
            Timeout: 60 * time.Second,
        },
    }

    // Execute via distributed coordinator
    return s.agentCoordinator.ExecuteAgent(
        ctx,
        "agent-manager",  // target service
        "kubectl-agent",  // target agent
        input,
    )
}

// Usage in workflow step
func (s *OrchestratorService) ExecuteCommandStep(
    ctx context.Context,
    step types.WorkflowStep,
) error {
    clusterID := step.Config["cluster_id"].(string)
    action := step.Config["action"].(string)
    args := step.Config["args"].([]string)

    output, err := s.ExecuteKubectlCommand(ctx, clusterID, action, args)
    if err != nil {
        return err
    }

    s.logger.Info("Command completed",
        "status", output.Status,
        "result", output.Result)

    return nil
}
```

## Using Common Tool Agents

### Example: HTTP Agent

```go
import "github.com/kart-io/k8s-agent/pkg/agent/tools"

func (s *Service) CallExternalAPI(ctx context.Context, url string) error {
    httpAgent := tools.NewHTTPAgent(s.logger)

    output, err := httpAgent.Get(ctx, url, map[string]string{
        "Authorization": "Bearer " + s.token,
    })

    if err != nil {
        return err
    }

    result := output.Result.(map[string]interface{})
    statusCode := result["status_code"].(int)
    body := result["body"]

    s.logger.Info("API call completed", "status", statusCode, "body", body)
    return nil
}
```

### Example: Database Agent

```go
import "github.com/kart-io/k8s-agent/pkg/agent/tools"

func (s *Service) QueryAgents(ctx context.Context) error {
    dbAgent := tools.NewDatabaseAgent(s.db, s.logger)

    output, err := dbAgent.Query(ctx,
        "SELECT * FROM agents WHERE status = ?",
        "online",
    )

    if err != nil {
        return err
    }

    result := output.Result.(map[string]interface{})
    rows := result["rows"].([]map[string]interface{})

    s.logger.Info("Query completed", "count", len(rows))
    return nil
}
```

### Example: Cache Agent

```go
import "github.com/kart-io/k8s-agent/pkg/agent/tools"

func (s *Service) CacheWorkflowResult(ctx context.Context, id string, result interface{}) error {
    cacheAgent := tools.NewCacheAgent(s.redis, s.logger)

    // Cache for 1 hour
    output, err := cacheAgent.Set(ctx, "workflow:"+id, result, 3600)

    if err != nil {
        return err
    }

    s.logger.Info("Cached workflow result", "id", id)
    return nil
}
```

## Service Registration

### Orchestrator Service

```go
// cmd/orchestrator/app/app.go

func (a *OrchestratorApp) registerWithAgentRegistry() error {
    instance := &distributed.ServiceInstance{
        ID:          a.instanceID,
        ServiceName: "orchestrator",
        Endpoint:    fmt.Sprintf("http://%s:%d", a.opts.Server.Host, a.opts.Server.Port),
        Agents: []string{
            "workflow-agent",
            "step-agent",
        },
        Metadata: map[string]interface{}{
            "version": version.Get().String(),
            "region":  a.opts.Region,
        },
    }

    return a.agentRegistry.Register(instance)
}
```

### Agent-Manager Service

```go
// cmd/agent-manager/app/app.go

func (a *AgentManagerApp) registerWithAgentRegistry() error {
    instance := &distributed.ServiceInstance{
        ID:          a.instanceID,
        ServiceName: "agent-manager",
        Endpoint:    fmt.Sprintf("http://%s:%d", a.opts.Server.Host, a.opts.Server.Port),
        Agents: []string{
            "command-agent",
            "kubectl-agent",
        },
        Metadata: map[string]interface{}{
            "version":  version.Get().String(),
            "region":   a.opts.Region,
            "clusters": a.registry.ClusterCount(),
        },
    }

    return a.agentRegistry.Register(instance)
}
```

## Health Check Integration

```go
// Add agent health to service health endpoint

func (h *HealthHandler) GetHealth(c *gin.Context) {
    health := map[string]interface{}{
        "status": "healthy",
        "agents": map[string]interface{}{
            "workflow-agent": map[string]interface{}{
                "status":      "healthy",
                "description": h.workflowAgent.Description(),
            },
            "step-agent": map[string]interface{}{
                "status":      "healthy",
                "description": h.stepAgent.Description(),
            },
        },
        "registry": h.agentRegistry.GetStatistics(),
    }

    c.JSON(200, health)
}
```

## Monitoring Integration

```go
// Expose Prometheus metrics endpoint

import (
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/kart-io/k8s-agent/pkg/agent/observability"
)

func RegisterMetricsEndpoint(r *gin.Engine) {
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

// Metrics are automatically collected by observability.WrapAgent()
```

## Testing

### Unit Test Example

```go
// internal/orchestrator/agents/workflow/workflow_agent_test.go

func TestWorkflowAgent_Execute(t *testing.T) {
    executor := &mockExecutor{}
    logger := logger.NewNop()
    agent := workflowagent.NewWorkflowAgent(executor, logger)

    execution := &types.WorkflowExecution{
        ID:         "test-1",
        WorkflowID: "workflow-1",
        Context:    make(map[string]interface{}),
    }

    steps := []types.WorkflowStep{
        {ID: "step-1", Type: "command"},
        {ID: "step-2", Type: "ai"},
    }

    input := &agentcore.AgentInput{
        Context: map[string]interface{}{
            "execution": execution,
            "steps":     steps,
        },
    }

    output, err := agent.Execute(context.Background(), input)

    assert.NoError(t, err)
    assert.Equal(t, "success", output.Status)
    assert.Equal(t, 2, len(output.ReasoningSteps))
}
```

## Conclusion

The Agent framework is now fully integrated into both orchestrator and agent-manager services, providing:

✅ Unified agent interface across services
✅ Distributed agent execution with failover
✅ Automatic observability (metrics, logs, traces)
✅ Common tool agents for reuse
✅ Cross-service communication via coordinator
✅ Health checks and monitoring

Services can now leverage agents for intelligent, observable, and distributed operations.
