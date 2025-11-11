package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kart-io/k8s-agent/pkg/agent/core"
	"github.com/kart-io/k8s-agent/pkg/agent/llm"
	"github.com/kart-io/k8s-agent/pkg/agent/memory"
	"github.com/kart-io/k8s-agent/pkg/agent/planning"
	"github.com/kart-io/logger"
)

// Example: Using the Planning module for complex task orchestration
func main() {
	// Initialize logger
	log := logger.New(logger.Config{
		Level:  logger.LevelInfo,
		Format: logger.FormatText,
	})

	// Initialize LLM client (mock for example)
	llmClient := &mockLLMClient{}

	// Initialize memory
	mem := memory.NewConversationMemory(100)

	// Create planner
	planner := planning.NewSmartPlanner(llmClient, mem,
		planning.WithMaxDepth(3),
		planning.WithTimeout(10*time.Minute),
	)

	// Create executor with agent registry
	executor := planning.NewAgentExecutor(log,
		planning.WithMaxConcurrency(3),
		planning.WithRetryPolicy(planning.RetryPolicy{
			MaxRetries:    2,
			RetryDelay:    1 * time.Second,
			BackoffFactor: 2.0,
		}),
	)

	// Register some agents for execution
	registry := executor.registry
	registry.RegisterAgent("analysis_agent", &AnalysisAgent{})
	registry.RegisterAgent("action_agent", &ActionAgent{})
	registry.RegisterAgent("validation_agent", &ValidationAgent{})

	// Example 1: Create and execute a plan for a complex goal
	fmt.Println("=== Example 1: Planning and Executing a Complex Task ===")
	demonstratePlanning(planner, executor)

	// Example 2: Using the PlanningAgent
	fmt.Println("\n=== Example 2: Using PlanningAgent ===")
	demonstratePlanningAgent(planner, executor)

	// Example 3: Task Decomposition
	fmt.Println("\n=== Example 3: Task Decomposition ===")
	demonstrateTaskDecomposition(planner)

	// Example 4: Strategy Application
	fmt.Println("\n=== Example 4: Strategy Application ===")
	demonstrateStrategyApplication()

	// Example 5: Plan Optimization
	fmt.Println("\n=== Example 5: Plan Optimization ===")
	demonstratePlanOptimization()
}

func demonstratePlanning(planner planning.Planner, executor planning.PlanExecutor) {
	ctx := context.Background()

	// Define a complex goal
	goal := "Deploy a microservices application to Kubernetes with zero downtime"

	// Define constraints
	constraints := planning.PlanConstraints{
		MaxSteps:    15,
		MaxDuration: 30 * time.Minute,
		RequiredSteps: []string{
			"health_check",
			"rollback_plan",
		},
		Resources: map[string]interface{}{
			"k8s_cluster": "production",
			"namespace":   "microservices",
		},
	}

	// Create a plan
	plan, err := planner.CreatePlan(ctx, goal, constraints)
	if err != nil {
		log.Fatalf("Failed to create plan: %v", err)
	}

	fmt.Printf("Created plan: %s\n", plan.ID)
	fmt.Printf("Goal: %s\n", plan.Goal)
	fmt.Printf("Strategy: %s\n", plan.Strategy)
	fmt.Printf("Total steps: %d\n", len(plan.Steps))

	// Print plan steps
	fmt.Println("\nPlan steps:")
	for _, step := range plan.Steps {
		fmt.Printf("  %s: %s (Type: %s, Priority: %d)\n",
			step.ID, step.Name, step.Type, step.Priority)

		// Show dependencies
		if deps, ok := plan.Dependencies[step.ID]; ok && len(deps) > 0 {
			fmt.Printf("    Dependencies: %v\n", deps)
		}
	}

	// Validate the plan
	valid, issues, err := planner.ValidatePlan(ctx, plan)
	if err != nil {
		log.Printf("Failed to validate plan: %v", err)
	}
	fmt.Printf("\nPlan validation: %v\n", valid)
	if len(issues) > 0 {
		fmt.Println("Issues found:")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
	}

	// Execute the plan
	fmt.Println("\nExecuting plan...")
	result, err := executor.Execute(ctx, plan)
	if err != nil {
		log.Printf("Failed to execute plan: %v", err)
	}

	fmt.Printf("\nExecution result:\n")
	fmt.Printf("  Success: %v\n", result.Success)
	fmt.Printf("  Completed steps: %d\n", result.CompletedSteps)
	fmt.Printf("  Failed steps: %d\n", result.FailedSteps)
	fmt.Printf("  Skipped steps: %d\n", result.SkippedSteps)
	fmt.Printf("  Total duration: %s\n", result.TotalDuration)

	// Show step results
	fmt.Println("\nStep results:")
	for stepID, stepResult := range result.StepResults {
		fmt.Printf("  %s: Success=%v, Duration=%s\n",
			stepID, stepResult.Success, stepResult.Duration)
		if !stepResult.Success && stepResult.Error != "" {
			fmt.Printf("    Error: %s\n", stepResult.Error)
		}
	}

	// Show plan metrics
	if plan.Metrics != nil {
		fmt.Printf("\nPlan metrics:\n")
		fmt.Printf("  Success rate: %.2f%%\n", plan.Metrics.SuccessRate*100)
		fmt.Printf("  Total duration: %s\n", plan.Metrics.TotalDuration)
	}
}

func demonstratePlanningAgent(planner planning.Planner, executor planning.PlanExecutor) {
	ctx := context.Background()

	// Create planning agent
	planningAgent := planning.NewPlanningAgent(planner, executor)

	// Prepare input
	input := &core.Input{
		Data: map[string]interface{}{
			"goal": "Optimize database performance for high-traffic application",
			"constraints": planning.PlanConstraints{
				MaxSteps:    10,
				MaxDuration: 15 * time.Minute,
			},
		},
		Options: map[string]interface{}{
			"execute": true, // Execute the plan immediately
		},
		Context: ctx,
	}

	// Execute planning agent
	output, err := planningAgent.Execute(ctx, input)
	if err != nil {
		log.Printf("Planning agent failed: %v", err)
		return
	}

	fmt.Printf("Planning agent output:\n")
	if result, ok := output.Data.(map[string]interface{}); ok {
		if plan, ok := result["plan"].(*planning.Plan); ok {
			fmt.Printf("  Plan ID: %s\n", plan.ID)
			fmt.Printf("  Steps: %d\n", len(plan.Steps))
		}
		if execResult, ok := result["result"].(*planning.PlanResult); ok {
			fmt.Printf("  Execution success: %v\n", execResult.Success)
			fmt.Printf("  Duration: %s\n", execResult.TotalDuration)
		}
	}
}

func demonstrateTaskDecomposition(planner planning.Planner) {
	ctx := context.Background()

	// Create task decomposition agent
	decomposer := planning.NewTaskDecompositionAgent(planner)

	// Complex task to decompose
	input := &core.Input{
		Data: map[string]interface{}{
			"task": "Implement a machine learning pipeline for real-time fraud detection",
		},
		Context: ctx,
	}

	// Decompose task
	output, err := decomposer.Execute(ctx, input)
	if err != nil {
		log.Printf("Task decomposition failed: %v", err)
		return
	}

	fmt.Println("Task decomposition result:")
	if subtasks, ok := output.Data.([]map[string]interface{}); ok {
		for _, subtask := range subtasks {
			fmt.Printf("  - %s: %s\n", subtask["name"], subtask["description"])
		}
	}
	fmt.Printf("  Total subtasks: %d\n", output.Metadata["num_subtasks"])
}

func demonstrateStrategyApplication() {
	ctx := context.Background()

	// Create strategy agent
	strategyAgent := planning.NewStrategyAgent()

	// Create a sample plan
	plan := &planning.Plan{
		ID:   "plan_123",
		Goal: "Scale application to handle 10x traffic",
		Steps: []*planning.Step{
			{
				ID:          "step_1",
				Name:        "Analyze current capacity",
				Description: "Analyze current system capacity and bottlenecks",
				Type:        planning.StepTypeAnalysis,
				Priority:    1,
			},
			{
				ID:          "step_2",
				Name:        "Scale infrastructure",
				Description: "Scale up infrastructure components",
				Type:        planning.StepTypeAction,
				Priority:    2,
			},
		},
		Dependencies: make(map[string][]string),
		Context:      make(map[string]interface{}),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Status:       planning.PlanStatusDraft,
	}

	// Apply different strategies
	strategies := []string{"decomposition", "backward_chaining", "hierarchical"}

	for _, strategy := range strategies {
		fmt.Printf("\nApplying %s strategy:\n", strategy)

		input := &core.Input{
			Data: map[string]interface{}{
				"plan":     plan,
				"strategy": strategy,
				"constraints": planning.PlanConstraints{
					MaxSteps: 20,
				},
			},
			Context: ctx,
		}

		output, err := strategyAgent.Execute(ctx, input)
		if err != nil {
			log.Printf("Strategy application failed: %v", err)
			continue
		}

		if refinedPlan, ok := output.Data.(*planning.Plan); ok {
			fmt.Printf("  Original steps: %d\n", len(plan.Steps))
			fmt.Printf("  Refined steps: %d\n", len(refinedPlan.Steps))
			fmt.Printf("  Status: %s\n", refinedPlan.Status)
		}
	}
}

func demonstratePlanOptimization() {
	ctx := context.Background()

	// Create optimization agent
	optimizer := planning.NewOptimizationAgent(nil)

	// Create a plan with redundant and unoptimized steps
	plan := &planning.Plan{
		ID:   "plan_456",
		Goal: "Migrate database from MySQL to PostgreSQL",
		Steps: []*planning.Step{
			{ID: "1", Name: "Backup data", Type: planning.StepTypeAction, Priority: 1},
			{ID: "2", Name: "Backup data", Type: planning.StepTypeAction, Priority: 2}, // Duplicate
			{ID: "3", Name: "Create schema", Type: planning.StepTypeAction, Priority: 3},
			{ID: "4", Name: "Migrate data", Type: planning.StepTypeAction, Priority: 4},
			{ID: "5", Name: "Validate data", Type: planning.StepTypeValidation, Priority: 5},
			{ID: "6", Name: "Test queries", Type: planning.StepTypeValidation, Priority: 6},
			{ID: "7", Name: "Update connections", Type: planning.StepTypeAction, Priority: 7},
		},
		Dependencies: map[string][]string{
			"4": {"3"},      // Migrate depends on schema
			"5": {"4"},      // Validate depends on migration
			"6": {"4"},      // Test also depends on migration (can parallel with validate)
			"7": {"5", "6"}, // Update depends on validation
		},
		Context: make(map[string]interface{}),
		Status:  planning.PlanStatusDraft,
	}

	// Optimize the plan
	input := &core.Input{
		Data: map[string]interface{}{
			"plan": plan,
		},
		Context: ctx,
	}

	output, err := optimizer.Execute(ctx, input)
	if err != nil {
		log.Printf("Optimization failed: %v", err)
		return
	}

	fmt.Println("Plan optimization result:")
	fmt.Printf("  Original steps: %d\n", output.Metadata["original_steps"])
	fmt.Printf("  Optimized steps: %d\n", output.Metadata["optimized_steps"])
	fmt.Printf("  Reduction: %.1f%%\n", output.Metadata["reduction_pct"])
	fmt.Printf("  Parallel steps: %d\n", output.Metadata["parallel_steps"])

	if optimizedPlan, ok := output.Data.(*planning.Plan); ok {
		fmt.Println("\nOptimized plan steps:")
		for _, step := range optimizedPlan.Steps {
			parallel := ""
			if p, ok := step.Parameters["parallel"].(bool); ok && p {
				parallel = " [PARALLEL]"
			}
			fmt.Printf("  %s: %s%s\n", step.ID, step.Name, parallel)
		}
	}
}

// Mock implementations for the example

type mockLLMClient struct{}

func (c *mockLLMClient) Complete(ctx context.Context, prompt string) (*llm.Response, error) {
	// Return mock response based on prompt content
	response := &llm.Response{
		Content: `Strategy: Progressive deployment with rollback capability

Steps:
1. Prepare deployment artifacts and configurations
2. Run comprehensive test suite
3. Create rollback snapshot
4. Deploy to staging environment
5. Run health checks on staging
6. Deploy to production (blue-green)
7. Monitor metrics and logs
8. Validate deployment success`,
	}
	return response, nil
}

func (c *mockLLMClient) Stream(ctx context.Context, prompt string, callback func(string)) error {
	return fmt.Errorf("streaming not implemented")
}

// Mock agents for execution

type AnalysisAgent struct {
	*core.BaseAgent
}

func (a *AnalysisAgent) Execute(ctx context.Context, input *core.Input) (*core.Output, error) {
	time.Sleep(100 * time.Millisecond) // Simulate work
	return &core.Output{
		Data: map[string]interface{}{
			"analysis": "Analysis completed successfully",
		},
	}, nil
}

type ActionAgent struct {
	*core.BaseAgent
}

func (a *ActionAgent) Execute(ctx context.Context, input *core.Input) (*core.Output, error) {
	time.Sleep(200 * time.Millisecond) // Simulate work
	return &core.Output{
		Data: map[string]interface{}{
			"action": "Action executed successfully",
		},
	}, nil
}

type ValidationAgent struct {
	*core.BaseAgent
}

func (a *ValidationAgent) Execute(ctx context.Context, input *core.Input) (*core.Output, error) {
	time.Sleep(50 * time.Millisecond) // Simulate work
	return &core.Output{
		Data: map[string]interface{}{
			"validation": "Validation passed",
			"valid":      true,
		},
	}, nil
}
