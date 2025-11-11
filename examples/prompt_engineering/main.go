package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/kart-io/k8s-agent/pkg/agent/llm"
	"github.com/kart-io/k8s-agent/pkg/agent/prompt"
	"github.com/kart-io/logger"
)

// Example: Demonstrating Prompt Engineering capabilities
func main() {
	// Initialize logger
	log := logger.New(logger.Config{
		Level:  logger.LevelInfo,
		Format: logger.FormatText,
	})

	fmt.Println("=== AI Agent System Demo: Prompt Engineering ===\n")

	// Initialize LLM client (mock for demo)
	llmClient := &mockLLMClient{}

	// Create prompt manager
	promptManager := prompt.NewPromptManager()

	// Part 1: Basic Prompt Management
	fmt.Println("=== Part 1: Basic Prompt Management ===")
	demonstrateBasicPrompts(promptManager, llmClient, log)

	// Part 2: Prompt Strategies
	fmt.Println("\n=== Part 2: Different Prompt Strategies ===")
	demonstratePromptStrategies(promptManager, llmClient, log)

	// Part 3: Prompt Chains
	fmt.Println("\n=== Part 3: Prompt Chains ===")
	demonstratePromptChains(promptManager, llmClient, log)

	// Part 4: Prompt Optimization
	fmt.Println("\n=== Part 4: Prompt Optimization ===")
	demonstratePromptOptimization(promptManager, llmClient, log)

	// Part 5: Prompt Testing
	fmt.Println("\n=== Part 5: Prompt Testing ===")
	demonstratePromptTesting(promptManager, llmClient, log)
}

// Part 1: Basic Prompt Management
func demonstrateBasicPrompts(manager *prompt.DefaultPromptManager, llmClient llm.LLMProvider, log logger.Logger) {
	ctx := context.Background()

	// Create a simple prompt
	simplePrompt := &prompt.Prompt{
		ID:          "summarize_text",
		Name:        "Text Summarizer",
		Description: "Summarizes long text into key points",
		Type:        prompt.PromptTypeUser,
		Strategy:    prompt.StrategyZeroShot,
		Template:    "Summarize the following text in 3 bullet points:\n\n{{.text}}",
		Variables: map[string]interface{}{
			"text": "Artificial Intelligence is transforming industries...",
		},
		Version: "1.0.0",
	}

	// Register the prompt
	if err := manager.CreatePrompt(simplePrompt); err != nil {
		log.Error("Failed to create prompt", logger.Error(err))
		return
	}
	fmt.Println("✓ Created simple summarization prompt")

	// Execute the prompt
	result, err := manager.ExecutePrompt(ctx, "summarize_text", map[string]interface{}{
		"text": `Artificial Intelligence (AI) is revolutionizing various sectors including healthcare,
		finance, and transportation. Machine learning algorithms can now diagnose diseases with
		remarkable accuracy, predict market trends, and enable autonomous vehicles. The rapid
		advancement in natural language processing has led to sophisticated chatbots and translation
		services. However, ethical concerns about bias, privacy, and job displacement need careful
		consideration as we integrate AI into society.`,
	})
	if err != nil {
		log.Error("Failed to execute prompt", logger.Error(err))
		return
	}

	fmt.Println("\nPrompt Output:")
	fmt.Println(result)
}

// Part 2: Different Prompt Strategies
func demonstratePromptStrategies(manager *prompt.DefaultPromptManager, llmClient llm.Client, log logger.Logger) {
	ctx := context.Background()

	// 1. Zero-shot prompt
	zeroShot := &prompt.Prompt{
		ID:       "classify_zero",
		Name:     "Zero-shot Classifier",
		Type:     prompt.PromptTypeUser,
		Strategy: prompt.StrategyZeroShot,
		Template: "Classify the following text as positive, negative, or neutral:\nText: {{.text}}",
		Version:  "1.0.0",
	}
	manager.CreatePrompt(zeroShot)
	fmt.Println("1. Created Zero-shot prompt")

	// 2. Few-shot prompt with examples
	fewShot := &prompt.Prompt{
		ID:       "classify_few",
		Name:     "Few-shot Classifier",
		Type:     prompt.PromptTypeUser,
		Strategy: prompt.StrategyFewShot,
		Template: "Classify the sentiment of the text:\nText: {{.text}}",
		Examples: []prompt.Example{
			{
				Input:  "I love this product!",
				Output: "positive",
			},
			{
				Input:  "This is terrible.",
				Output: "negative",
			},
			{
				Input:  "It's okay, nothing special.",
				Output: "neutral",
			},
		},
		Version: "1.0.0",
	}
	manager.CreatePrompt(fewShot)
	fmt.Println("2. Created Few-shot prompt with 3 examples")

	// 3. Chain-of-thought prompt
	chainOfThought := &prompt.Prompt{
		ID:       "math_cot",
		Name:     "Math Problem Solver",
		Type:     prompt.PromptTypeChainOfThought,
		Strategy: prompt.StrategyChainOfThought,
		Template: `Solve this math problem step by step:
{{.problem}}

Let's think through this step by step:`,
		Version: "1.0.0",
	}
	manager.CreatePrompt(chainOfThought)
	fmt.Println("3. Created Chain-of-Thought prompt")

	// 4. Role-playing prompt
	rolePlay := &prompt.Prompt{
		ID:           "expert_advisor",
		Name:         "Expert Advisor",
		Type:         prompt.PromptTypeSystem,
		Strategy:     prompt.StrategyRolePlay,
		SystemPrompt: "You are an experienced {{.role}} with 20 years of expertise.",
		Template:     "As an expert {{.role}}, please advise on: {{.question}}",
		Variables: map[string]interface{}{
			"role": "software architect",
		},
		Version: "1.0.0",
	}
	manager.CreatePrompt(rolePlay)
	fmt.Println("4. Created Role-playing prompt")

	// 5. Instructional prompt with constraints
	instructional := &prompt.Prompt{
		ID:       "code_reviewer",
		Name:     "Code Reviewer",
		Type:     prompt.PromptTypeInstruction,
		Strategy: prompt.StrategyInstructional,
		Template: "Review the following code:\n```{{.language}}\n{{.code}}\n```",
		Constraints: []string{
			"Focus on security vulnerabilities",
			"Check for performance issues",
			"Suggest best practices",
			"Provide specific line numbers",
			"Rate severity: High/Medium/Low",
		},
		Version: "1.0.0",
	}
	manager.CreatePrompt(instructional)
	fmt.Println("5. Created Instructional prompt with constraints")

	// Execute different strategies
	testText := "The new feature works perfectly!"

	fmt.Println("\nTesting different strategies with:", testText)

	// Test zero-shot
	result1, _ := manager.ExecutePrompt(ctx, "classify_zero", map[string]interface{}{
		"text": testText,
	})
	fmt.Println("  Zero-shot result:", result1[:50]+"...")

	// Test few-shot
	result2, _ := manager.ExecutePrompt(ctx, "classify_few", map[string]interface{}{
		"text": testText,
	})
	fmt.Println("  Few-shot result:", result2[:50]+"...")
}

// Part 3: Prompt Chains
func demonstratePromptChains(manager *prompt.DefaultPromptManager, llmClient llm.Client, log logger.Logger) {
	ctx := context.Background()

	// Create prompts for the chain
	// Step 1: Extract key information
	extractPrompt := &prompt.Prompt{
		ID:       "extract_info",
		Name:     "Information Extractor",
		Type:     prompt.PromptTypeUser,
		Template: "Extract key facts from this text:\n{{.text}}",
		Version:  "1.0.0",
	}
	manager.CreatePrompt(extractPrompt)

	// Step 2: Analyze sentiment
	analyzePrompt := &prompt.Prompt{
		ID:       "analyze_sentiment",
		Name:     "Sentiment Analyzer",
		Type:     prompt.PromptTypeUser,
		Template: "Analyze the sentiment of these facts:\n{{.facts}}",
		Version:  "1.0.0",
	}
	manager.CreatePrompt(analyzePrompt)

	// Step 3: Generate summary
	summaryPrompt := &prompt.Prompt{
		ID:       "generate_summary",
		Name:     "Summary Generator",
		Type:     prompt.PromptTypeUser,
		Template: "Create a summary combining facts and sentiment:\nFacts: {{.facts}}\nSentiment: {{.sentiment}}",
		Version:  "1.0.0",
	}
	manager.CreatePrompt(summaryPrompt)

	// Create a sequential chain
	chain := &prompt.PromptChain{
		ID:          "analysis_pipeline",
		Name:        "Text Analysis Pipeline",
		Description: "Extract, analyze, and summarize text",
		Strategy:    prompt.ChainStrategySequential,
		Steps: []prompt.PromptStep{
			{
				ID:        "step1",
				PromptID:  "extract_info",
				Order:     1,
				OutputKey: "facts",
			},
			{
				ID:       "step2",
				PromptID: "analyze_sentiment",
				Order:    2,
				InputMapping: map[string]string{
					"facts": "facts", // Use output from step1
				},
				OutputKey: "sentiment",
			},
			{
				ID:       "step3",
				PromptID: "generate_summary",
				Order:    3,
				InputMapping: map[string]string{
					"facts":     "facts",     // From step1
					"sentiment": "sentiment", // From step2
				},
				OutputKey: "summary",
			},
		},
	}

	// Register the chain
	if err := manager.CreateChain(chain); err != nil {
		log.Error("Failed to create chain", logger.Error(err))
		return
	}
	fmt.Println("✓ Created analysis pipeline with 3 steps")

	// Execute the chain
	input := map[string]interface{}{
		"text": `The company reported record profits this quarter, exceeding analyst expectations
		by 15%. However, concerns about supply chain disruptions continue to worry investors.
		The CEO remains optimistic about future growth prospects.`,
	}

	fmt.Println("\nExecuting prompt chain...")
	results, err := manager.ExecuteChain(ctx, "analysis_pipeline", input)
	if err != nil {
		log.Error("Failed to execute chain", logger.Error(err))
		return
	}

	fmt.Println("Chain Results:")
	for key, value := range results {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

// Part 4: Prompt Optimization
func demonstratePromptOptimization(manager *prompt.DefaultPromptManager, llmClient llm.LLMProvider, log logger.Logger) {
	// Create a prompt to optimize
	initialPrompt := &prompt.Prompt{
		ID:       "task_solver",
		Name:     "Task Solver",
		Type:     prompt.PromptTypeUser,
		Strategy: prompt.StrategyZeroShot,
		Template: "Solve this: {{.task}}",
		Version:  "1.0.0",
	}
	manager.CreatePrompt(initialPrompt)

	// Simulate feedback from usage
	feedback := []prompt.Feedback{
		{
			PromptID: "task_solver",
			Input:    "Calculate 15% of 200",
			Output:   "30",
			Expected: "30",
			Score:    1.0,
			Comments: "Correct calculation",
		},
		{
			PromptID: "task_solver",
			Input:    "What is the capital of France?",
			Output:   "Paris",
			Expected: "Paris",
			Score:    1.0,
			Comments: "Correct answer",
		},
		{
			PromptID: "task_solver",
			Input:    "Explain quantum computing",
			Output:   "It's complicated",
			Expected: "Detailed explanation",
			Score:    0.3,
			Comments: "Too brief, needs more detail",
		},
		{
			PromptID: "task_solver",
			Input:    "Write a haiku about spring",
			Output:   "Spring is nice",
			Expected: "5-7-5 syllable haiku",
			Score:    0.2,
			Comments: "Wrong format, not a haiku",
		},
	}

	fmt.Println("Initial prompt performance:")
	totalScore := 0.0
	for i, f := range feedback {
		fmt.Printf("  Test %d: Score %.2f - %s\n", i+1, f.Score, f.Comments)
		totalScore += f.Score
	}
	fmt.Printf("  Average score: %.2f\n", totalScore/float64(len(feedback)))

	// Optimize the prompt based on feedback
	optimizedPrompt, err := manager.OptimizePrompt("task_solver", feedback)
	if err != nil {
		log.Error("Failed to optimize prompt", logger.Error(err))
		return
	}

	fmt.Println("\n✓ Prompt optimized based on feedback")
	fmt.Println("Optimizations applied:")

	// Check what changed
	original, _ := manager.GetPrompt("task_solver")
	if optimizedPrompt.Strategy != original.Strategy {
		fmt.Printf("  - Strategy changed: %s → %s\n", original.Strategy, optimizedPrompt.Strategy)
	}
	if len(optimizedPrompt.Examples) > len(original.Examples) {
		fmt.Printf("  - Examples added: %d → %d\n", len(original.Examples), len(optimizedPrompt.Examples))
	}
	if len(optimizedPrompt.Constraints) > len(original.Constraints) {
		fmt.Printf("  - Constraints added: %d → %d\n", len(original.Constraints), len(optimizedPrompt.Constraints))
	}
	if optimizedPrompt.Template != original.Template {
		fmt.Println("  - Template improved for clarity")
	}
	if optimizedPrompt.SystemPrompt != original.SystemPrompt && optimizedPrompt.SystemPrompt != "" {
		fmt.Println("  - System prompt added for expertise")
	}

	fmt.Printf("  - Version updated: %s → %s\n", original.Version, optimizedPrompt.Version)
}

// Part 5: Prompt Testing
func demonstratePromptTesting(manager *prompt.DefaultPromptManager, llmClient llm.Client, log logger.Logger) {
	// Create test cases
	testCases := []prompt.TestCase{
		{
			ID: "test_1",
			Input: map[string]interface{}{
				"number1": 10,
				"number2": 5,
			},
			Expected:    "15",
			Description: "Simple addition",
		},
		{
			ID: "test_2",
			Input: map[string]interface{}{
				"number1": 20,
				"number2": 4,
			},
			Expected:    "24",
			Description: "Another addition",
		},
		{
			ID: "test_3",
			Input: map[string]interface{}{
				"number1": 100,
				"number2": 25,
			},
			Expected:    "125",
			Description: "Larger numbers",
		},
	}

	// Create a math prompt to test
	mathPrompt := &prompt.Prompt{
		ID:       "math_adder",
		Name:     "Math Adder",
		Type:     prompt.PromptTypeUser,
		Template: "Calculate: {{.number1}} + {{.number2}} = ",
		Version:  "1.0.0",
	}
	manager.CreatePrompt(mathPrompt)

	fmt.Println("Running prompt tests...")
	fmt.Println("Test cases: 3 addition problems")

	// Run tests
	testResult, err := manager.TestPrompt("math_adder", testCases)
	if err != nil {
		log.Error("Failed to test prompt", logger.Error(err))
		return
	}

	// Display results
	fmt.Printf("\n✓ Test Results:\n")
	fmt.Printf("  Total cases: %d\n", testResult.TotalCases)
	fmt.Printf("  Passed: %d\n", testResult.PassedCases)
	fmt.Printf("  Failed: %d\n", testResult.FailedCases)
	fmt.Printf("  Success rate: %.1f%%\n", testResult.SuccessRate*100)
	fmt.Printf("  Duration: %s\n", testResult.Duration)

	fmt.Println("\nDetailed Results:")
	for _, detail := range testResult.Details {
		status := "✓ PASS"
		if !detail.Passed {
			status = "✗ FAIL"
		}
		fmt.Printf("  %s - Test %s: Expected=%s, Actual=%s, Score=%.2f\n",
			status, detail.TestCaseID, detail.Expected, detail.Actual, detail.Score)
	}

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("\nKey Prompt Engineering Capabilities Demonstrated:")
	fmt.Println("• Template-based prompt generation with variables")
	fmt.Println("• Multiple prompting strategies (Zero-shot, Few-shot, CoT, etc.)")
	fmt.Println("• Prompt chains for complex workflows")
	fmt.Println("• Optimization based on feedback")
	fmt.Println("• Automated testing and evaluation")
	fmt.Println("• Version management and constraints")
	fmt.Println("• System prompts and role-playing")
}

// Mock LLM client for demonstration
type mockLLMClient struct{}

func (c *mockLLMClient) Complete(ctx context.Context, prompt string) (*llm.Response, error) {
	// Return mock responses based on prompt content
	response := "Mock response for: " + prompt

	// Simulate different responses for different prompts
	if strings.Contains(prompt, "Summarize") {
		response = "• Key point 1\n• Key point 2\n• Key point 3"
	} else if strings.Contains(prompt, "positive, negative, or neutral") {
		response = "positive"
	} else if strings.Contains(prompt, "Calculate") && strings.Contains(prompt, "+") {
		// Simple addition parser for demo
		parts := strings.Split(prompt, "+")
		if len(parts) == 2 {
			// Extract numbers (simplified)
			response = "15" // Default for demo
		}
	}

	return &llm.Response{
		Content: response,
		TokensUsed: llm.TokenUsage{
			Prompt:     len(prompt),
			Completion: len(response),
			Total:      len(prompt) + len(response),
		},
	}, nil
}

func (c *mockLLMClient) Stream(ctx context.Context, prompt string, callback func(string)) error {
	return fmt.Errorf("streaming not implemented")
}
