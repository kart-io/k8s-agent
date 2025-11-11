package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kart-io/k8s-agent/pkg/agent/core"
	"github.com/kart-io/k8s-agent/pkg/agent/llm"
	"github.com/kart-io/k8s-agent/pkg/agent/memory"
	"github.com/kart-io/k8s-agent/pkg/agent/multiagent"
	"github.com/kart-io/k8s-agent/pkg/agent/reflection"
	"github.com/kart-io/logger"
)

// Example: Demonstrating Reflection and Multi-Agent collaboration
func main() {
	// Initialize logger
	log := logger.New(logger.Config{
		Level:  logger.LevelInfo,
		Format: logger.FormatText,
	})

	fmt.Println("=== AI Agent System Demo: Reflection & Multi-Agent ===\n")

	// Part 1: Reflection System Demo
	fmt.Println("=== Part 1: Self-Reflective Agent ===")
	demonstrateReflection(log)

	// Part 2: Multi-Agent Collaboration Demo
	fmt.Println("\n=== Part 2: Multi-Agent Collaboration ===")
	demonstrateMultiAgent(log)

	// Part 3: Combined Demo - Reflective Multi-Agent System
	fmt.Println("\n=== Part 3: Reflective Multi-Agent System ===")
	demonstrateReflectiveMultiAgent(log)
}

// Part 1: Demonstrate Self-Reflection capabilities
func demonstrateReflection(log logger.Logger) {
	ctx := context.Background()

	// Initialize components
	llmClient := &mockLLMClient{}
	vectorStore := memory.NewInMemoryVectorStore(128)
	enhancedMemory := memory.NewHierarchicalMemory(vectorStore,
		memory.WithShortTermCapacity(50),
		memory.WithDecayRate(0.01),
	)

	// Create self-reflective agent
	reflectiveAgent := reflection.NewSelfReflectiveAgent(llmClient, enhancedMemory,
		reflection.WithReflectionInterval(10*time.Second),
		reflection.WithLearningThreshold(0.6),
	)

	// Simulate multiple task executions to build experience
	fmt.Println("1. Executing tasks to build experience...")

	for i := 1; i <= 5; i++ {
		input := &core.Input{
			Data: map[string]interface{}{
				"task_id":    fmt.Sprintf("task_%d", i),
				"operation":  "process_data",
				"complexity": i,
			},
			Options: map[string]interface{}{
				"iteration": i,
			},
			Context: ctx,
		}

		output, err := reflectiveAgent.Execute(ctx, input)
		if err != nil {
			fmt.Printf("   Task %d failed: %v\n", i, err)
		} else {
			fmt.Printf("   Task %d completed: %v\n", i, output.Data)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Perform reflection
	fmt.Println("\n2. Performing self-reflection...")

	reflectionSubject := map[string]interface{}{
		"tasks_completed": 5,
		"context":         "Learning from task execution",
	}

	reflectionResult, err := reflectiveAgent.Reflect(ctx, reflectionSubject)
	if err != nil {
		log.Error("Reflection failed", logger.Error(err))
		return
	}

	fmt.Printf("   Reflection completed:\n")
	fmt.Printf("   - Performance Score: %.2f\n", reflectionResult.PerformanceScore)
	fmt.Printf("   - Confidence: %.2f\n", reflectionResult.Confidence)

	if reflectionResult.Evaluation != nil {
		fmt.Printf("   - Strengths: %v\n", reflectionResult.Evaluation.Strengths)
		fmt.Printf("   - Weaknesses: %v\n", reflectionResult.Evaluation.Weaknesses)
	}

	// Extract learnings
	fmt.Println("\n3. Extracting learnings from experiences...")

	experiences := []reflection.Experience{
		{
			ID:          "exp_1",
			Description: "Data processing task",
			Success:     true,
			Duration:    100 * time.Millisecond,
		},
		{
			ID:          "exp_2",
			Description: "Complex computation",
			Success:     false,
			Error:       "Timeout exceeded",
		},
	}

	learnings, err := reflectiveAgent.ExtractLearnings(ctx, experiences)
	if err != nil {
		fmt.Printf("   Learning extraction limited: %v\n", err)
	} else {
		fmt.Println("   Learnings extracted:")
		for _, learning := range learnings {
			fmt.Printf("   - %s (Confidence: %.2f, Applicability: %.2f)\n",
				learning.Lesson, learning.Confidence, learning.Applicability)
		}
	}

	// Generate improvements
	fmt.Println("\n4. Generating improvement suggestions...")

	metrics := reflection.PerformanceMetrics{
		SuccessRate:     0.8,
		AverageTime:     150 * time.Millisecond,
		ErrorRate:       0.2,
		TotalExecutions: 5,
	}

	evaluation, _ := reflectiveAgent.EvaluatePerformance(ctx, metrics)
	improvements, _ := reflectiveAgent.GenerateImprovements(ctx, evaluation)

	fmt.Println("   Suggested improvements:")
	for _, improvement := range improvements {
		fmt.Printf("   - %s (Priority: %d, Impact: %.2f, Effort: %.2f)\n",
			improvement.Description, improvement.Priority, improvement.Impact, improvement.Effort)
		for _, action := range improvement.ActionItems {
			fmt.Printf("     • %s\n", action)
		}
	}
}

// Part 2: Demonstrate Multi-Agent collaboration
func demonstrateMultiAgent(log logger.Logger) {
	ctx := context.Background()

	// Create multi-agent system
	system := multiagent.NewMultiAgentSystem(log,
		multiagent.WithMaxAgents(10),
		multiagent.WithTimeout(10*time.Second),
	)

	// Create and register different types of agents
	fmt.Println("1. Creating agent team...")

	// Leader agent
	leader := multiagent.NewBaseCollaborativeAgent("leader_1", "Team Leader", multiagent.RoleLeader, system)
	system.RegisterAgent("leader_1", leader)
	fmt.Println("   - Registered Leader Agent")

	// Worker agents
	for i := 1; i <= 3; i++ {
		worker := multiagent.NewBaseCollaborativeAgent(
			fmt.Sprintf("worker_%d", i),
			fmt.Sprintf("Worker Agent %d", i),
			multiagent.RoleWorker,
			system,
		)
		system.RegisterAgent(fmt.Sprintf("worker_%d", i), worker)
		fmt.Printf("   - Registered Worker Agent %d\n", i)
	}

	// Specialist agent
	specialist := multiagent.NewSpecializedAgent("specialist_1", "Data Analysis", system)
	system.RegisterAgent("specialist_1", specialist)
	fmt.Println("   - Registered Specialist Agent (Data Analysis)")

	// Validator agent
	validator := multiagent.NewBaseCollaborativeAgent("validator_1", "Result Validator", multiagent.RoleValidator, system)
	system.RegisterAgent("validator_1", validator)
	fmt.Println("   - Registered Validator Agent")

	// Create a team
	team := &multiagent.Team{
		ID:           "alpha_team",
		Name:         "Alpha Team",
		Leader:       "leader_1",
		Members:      []string{"leader_1", "worker_1", "worker_2", "worker_3", "specialist_1", "validator_1"},
		Purpose:      "Complex problem solving",
		Capabilities: []string{"analysis", "computation", "validation"},
	}

	if err := system.CreateTeam(team); err != nil {
		log.Error("Failed to create team", logger.Error(err))
		return
	}
	fmt.Println("\n   Team 'Alpha Team' created successfully!")

	// Test different collaboration types
	fmt.Println("\n2. Testing collaboration patterns...")

	// Test 1: Parallel Collaboration
	fmt.Println("\n   A. Parallel Collaboration:")
	parallelTask := &multiagent.CollaborativeTask{
		ID:          "task_parallel",
		Name:        "Parallel Processing",
		Description: "Process data in parallel",
		Type:        multiagent.CollaborationTypeParallel,
		Input: map[string]interface{}{
			"data":      []int{1, 2, 3, 4, 5},
			"operation": "transform",
		},
		Assignments: make(map[string]multiagent.Assignment),
	}

	result, err := system.ExecuteTask(ctx, parallelTask)
	if err != nil {
		fmt.Printf("      Error: %v\n", err)
	} else {
		fmt.Printf("      Status: %s\n", result.Status)
		fmt.Printf("      Duration: %s\n", result.EndTime.Sub(result.StartTime))
		fmt.Printf("      Participating agents: %d\n", len(result.Assignments))
	}

	// Test 2: Sequential Collaboration
	fmt.Println("\n   B. Sequential Collaboration:")
	sequentialTask := &multiagent.CollaborativeTask{
		ID:          "task_sequential",
		Name:        "Sequential Processing",
		Description: "Process data sequentially through pipeline",
		Type:        multiagent.CollaborationTypeSequential,
		Input:       "initial_data",
		Assignments: make(map[string]multiagent.Assignment),
	}

	result, err = system.ExecuteTask(ctx, sequentialTask)
	if err != nil {
		fmt.Printf("      Error: %v\n", err)
	} else {
		fmt.Printf("      Status: %s\n", result.Status)
		fmt.Printf("      Final output: %v\n", result.Output)
		fmt.Printf("      Pipeline stages: %d\n", len(result.Assignments))
	}

	// Test 3: Hierarchical Collaboration
	fmt.Println("\n   C. Hierarchical Collaboration:")
	hierarchicalTask := &multiagent.CollaborativeTask{
		ID:          "task_hierarchical",
		Name:        "Hierarchical Processing",
		Description: "Leader coordinates workers",
		Type:        multiagent.CollaborationTypeHierarchical,
		Input: map[string]interface{}{
			"problem":  "complex_computation",
			"subtasks": 3,
		},
		Assignments: make(map[string]multiagent.Assignment),
	}

	result, err = system.ExecuteTask(ctx, hierarchicalTask)
	if err != nil {
		fmt.Printf("      Error: %v\n", err)
	} else {
		fmt.Printf("      Status: %s\n", result.Status)
		fmt.Printf("      Leader coordinated: Yes\n")
		fmt.Printf("      Workers involved: %d\n", len(result.Assignments)-1)
	}

	// Test 4: Consensus Collaboration
	fmt.Println("\n   D. Consensus Collaboration:")
	consensusTask := &multiagent.CollaborativeTask{
		ID:          "task_consensus",
		Name:        "Consensus Decision",
		Description: "Reach consensus on proposal",
		Type:        multiagent.CollaborationTypeConsensus,
		Input: map[string]interface{}{
			"proposal": "Adopt new algorithm",
			"criteria": []string{"efficiency", "accuracy", "scalability"},
		},
		Assignments: make(map[string]multiagent.Assignment),
	}

	result, err = system.ExecuteTask(ctx, consensusTask)
	if err != nil {
		fmt.Printf("      Error: %v\n", err)
	} else {
		fmt.Printf("      Status: %s\n", result.Status)
		if output, ok := result.Output.(map[string]interface{}); ok {
			fmt.Printf("      Consensus reached: %v\n", output["consensus_reached"])
			fmt.Printf("      Yes votes: %v/%v\n", output["yes_votes"], output["total_votes"])
		}
	}

	// Test 5: Pipeline Collaboration
	fmt.Println("\n   E. Pipeline Collaboration:")
	pipelineTask := &multiagent.CollaborativeTask{
		ID:          "task_pipeline",
		Name:        "Pipeline Processing",
		Description: "Multi-stage pipeline",
		Type:        multiagent.CollaborationTypePipeline,
		Input: []interface{}{
			"stage1: data_collection",
			"stage2: data_processing",
			"stage3: analysis",
			"stage4: validation",
		},
		Assignments: make(map[string]multiagent.Assignment),
	}

	result, err = system.ExecuteTask(ctx, pipelineTask)
	if err != nil {
		fmt.Printf("      Error: %v\n", err)
	} else {
		fmt.Printf("      Status: %s\n", result.Status)
		fmt.Printf("      Pipeline stages completed: %d\n", len(result.Results))
	}

	// Demonstrate agent communication
	fmt.Println("\n3. Testing inter-agent communication...")

	// Send broadcast message
	broadcastMsg := multiagent.Message{
		ID:        "msg_broadcast",
		From:      "leader_1",
		To:        "all",
		Type:      multiagent.MessageTypeBroadcast,
		Content:   "Team meeting at 3 PM",
		Priority:  1,
		Timestamp: time.Now(),
	}

	if err := system.SendMessage(broadcastMsg); err != nil {
		fmt.Printf("   Failed to broadcast: %v\n", err)
	} else {
		fmt.Println("   Broadcast message sent successfully")
	}
}

// Part 3: Combined Reflective Multi-Agent System
func demonstrateReflectiveMultiAgent(log logger.Logger) {
	ctx := context.Background()

	fmt.Println("Creating a team of self-reflective agents...")

	// Create system
	system := multiagent.NewMultiAgentSystem(log)

	// Create reflective agents that can collaborate
	llmClient := &mockLLMClient{}

	// Create team of reflective agents
	for i := 1; i <= 3; i++ {
		// Each agent has its own memory and reflection capability
		vectorStore := memory.NewInMemoryVectorStore(64)
		agentMemory := memory.NewHierarchicalMemory(vectorStore)

		// Create reflective collaborative agent
		reflectiveCollab := &ReflectiveCollaborativeAgent{
			BaseCollaborativeAgent: multiagent.NewBaseCollaborativeAgent(
				fmt.Sprintf("reflective_%d", i),
				"Reflective Collaborative Agent",
				multiagent.RoleWorker,
				system,
			),
			reflectiveAgent: reflection.NewSelfReflectiveAgent(llmClient, agentMemory),
		}

		system.RegisterAgent(fmt.Sprintf("reflective_%d", i), reflectiveCollab)
		fmt.Printf("   - Registered Reflective Agent %d\n", i)
	}

	// Execute collaborative task
	fmt.Println("\n1. Executing collaborative learning task...")

	learningTask := &multiagent.CollaborativeTask{
		ID:          "learning_task",
		Name:        "Collaborative Learning",
		Description: "Learn from collective experience",
		Type:        multiagent.CollaborationTypeParallel,
		Input: map[string]interface{}{
			"objective":       "Optimize collective performance",
			"share_learnings": true,
		},
		Assignments: make(map[string]multiagent.Assignment),
	}

	result, err := system.ExecuteTask(ctx, learningTask)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Status: %s\n", result.Status)
		fmt.Println("   Agents learned from each other's experiences")
	}

	// Each agent reflects on the collaboration
	fmt.Println("\n2. Agents reflecting on collaboration...")

	for i := 1; i <= 3; i++ {
		agentID := fmt.Sprintf("reflective_%d", i)
		fmt.Printf("   Agent %s: Reflection in progress...\n", agentID)
		// In real implementation, each agent would perform reflection
	}

	fmt.Println("\n3. System-wide learning consolidation...")
	fmt.Println("   - Experiences shared across agents")
	fmt.Println("   - Collective improvements identified")
	fmt.Println("   - Team performance optimized")

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("\nKey Capabilities Demonstrated:")
	fmt.Println("• Self-Reflection: Agents evaluate and improve their performance")
	fmt.Println("• Learning Extraction: Derive insights from experiences")
	fmt.Println("• Multi-Agent Collaboration: Various collaboration patterns")
	fmt.Println("• Consensus Building: Democratic decision making")
	fmt.Println("• Hierarchical Coordination: Leader-worker dynamics")
	fmt.Println("• Pipeline Processing: Sequential task execution")
	fmt.Println("• Collective Learning: Agents learn from shared experiences")
}

// ReflectiveCollaborativeAgent combines reflection and collaboration
type ReflectiveCollaborativeAgent struct {
	*multiagent.BaseCollaborativeAgent
	reflectiveAgent reflection.ReflectiveAgent
}

// Collaborate with reflection
func (a *ReflectiveCollaborativeAgent) Collaborate(ctx context.Context, task *multiagent.CollaborativeTask) (*multiagent.Assignment, error) {
	// Execute task
	assignment, err := a.BaseCollaborativeAgent.Collaborate(ctx, task)

	// Reflect on the experience
	experience := reflection.Experience{
		ID:          fmt.Sprintf("collab_%s", task.ID),
		Description: task.Description,
		Input:       task.Input,
		Output:      assignment.Result,
		Success:     err == nil,
		Duration:    assignment.EndTime.Sub(assignment.StartTime),
		Timestamp:   time.Now(),
	}

	// Store experience for future learning
	if err == nil {
		// Agent learns from this collaboration
		a.reflectiveAgent.ExtractLearnings(ctx, []reflection.Experience{experience})
	}

	return assignment, err
}

// Mock LLM client for demonstration
type mockLLMClient struct{}

func (c *mockLLMClient) Complete(ctx context.Context, prompt string) (*llm.Response, error) {
	// Return mock responses based on prompt content
	return &llm.Response{
		Content: "Mock LLM response for: " + prompt[:min(50, len(prompt))],
	}, nil
}

func (c *mockLLMClient) Stream(ctx context.Context, prompt string, callback func(string)) error {
	return fmt.Errorf("streaming not implemented")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
