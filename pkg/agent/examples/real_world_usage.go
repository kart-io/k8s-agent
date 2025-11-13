package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kart-io/k8s-agent/pkg/agent/agents"
	"github.com/kart-io/k8s-agent/pkg/agent/core"
	"github.com/kart-io/k8s-agent/pkg/agent/llm"
	"github.com/kart-io/k8s-agent/pkg/agent/llm/providers"
	"github.com/kart-io/k8s-agent/pkg/agent/store"
	"github.com/kart-io/k8s-agent/pkg/agent/stream"
	"github.com/kart-io/k8s-agent/pkg/agent/tools"
	"github.com/kart-io/k8s-agent/pkg/agent/tools/practical"
)

// toolWrapper wraps functions as Tools for the executor
type toolWrapper struct {
	fn   func(context.Context, interface{}) (interface{}, error)
	args interface{}
}

func (w *toolWrapper) GetName() string              { return "wrapper" }
func (w *toolWrapper) GetDescription() string       { return "Tool wrapper" }
func (w *toolWrapper) GetInputSchema() interface{}  { return nil }
func (w *toolWrapper) GetOutputSchema() interface{} { return nil }
func (w *toolWrapper) Execute(ctx context.Context, input *tools.ToolInput) (*tools.ToolOutput, error) {
	result, err := w.fn(ctx, w.args)
	return &tools.ToolOutput{Result: result}, err
}

// RealWorldExamples demonstrates practical use cases
type RealWorldExamples struct {
	llm        llm.Client
	store      store.Store
	tools      *tools.Registry
	supervisor *agents.SupervisorAgent
}

// NewRealWorldExamples creates a new examples runner
func NewRealWorldExamples(apiKey string) (*RealWorldExamples, error) {
	// Initialize OpenAI (you can swap with Gemini or DeepSeek)
	config := &llm.Config{
		APIKey:      apiKey,
		Model:       "gpt-4",
		MaxTokens:   2000,
		Temperature: 0.7,
	}

	llmProvider, err := providers.NewOpenAI(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}

	// Create store
	langStore := store.NewInMemoryStore()

	// Initialize tools
	toolRegistry := tools.NewRegistry()

	// Register practical tools
	toolRegistry.Register(practical.NewWebScraperRuntimeTool())
	toolRegistry.Register(practical.NewAPICallerRuntimeTool())
	toolRegistry.Register(practical.NewDatabaseQueryRuntimeTool())
	toolRegistry.Register(practical.NewFileOperationsRuntimeTool("/tmp/agent"))

	return &RealWorldExamples{
		llm:   llmProvider,
		store: langStore,
		tools: toolRegistry,
	}, nil
}

// Example1_CustomerSupportAgent demonstrates a customer support automation
func (e *RealWorldExamples) Example1_CustomerSupportAgent() {
	fmt.Println("=== Customer Support Agent Example ===\n")

	ctx := context.Background()

	// Create support agent with specific tools
	supportAgent := &CustomerSupportAgent{
		llm:   e.llm,
		store: e.store,
		tools: e.tools,
	}

	// Simulate customer queries
	queries := []CustomerQuery{
		{
			ID:       "ticket-001",
			Customer: "john@example.com",
			Message:  "I can't login to my account. It says my password is incorrect but I'm sure it's right.",
			Priority: "high",
		},
		{
			ID:       "ticket-002",
			Customer: "jane@example.com",
			Message:  "How do I export my data? I need it in CSV format for my quarterly report.",
			Priority: "medium",
		},
		{
			ID:       "ticket-003",
			Customer: "bob@example.com",
			Message:  "Your API is returning 500 errors when I try to update user profiles.",
			Priority: "critical",
		},
	}

	for _, query := range queries {
		fmt.Printf("Processing ticket %s from %s\n", query.ID, query.Customer)

		response, err := supportAgent.HandleQuery(ctx, query)
		if err != nil {
			log.Printf("Error handling query: %v", err)
			continue
		}

		fmt.Printf("Response:\n%s\n\n", response.Message)
		fmt.Printf("Actions taken: %v\n", response.ActionsTaken)
		fmt.Printf("Escalated: %v\n\n", response.Escalated)
	}
}

// CustomerSupportAgent handles customer support queries
type CustomerSupportAgent struct {
	llm   llm.Client
	store store.Store
	tools *tools.Registry
}

type CustomerQuery struct {
	ID       string
	Customer string
	Message  string
	Priority string
}

type SupportResponse struct {
	Message      string
	ActionsTaken []string
	Escalated    bool
	Resolution   string
}

func (a *CustomerSupportAgent) HandleQuery(ctx context.Context, query CustomerQuery) (*SupportResponse, error) {
	// Store query in memory
	a.store.Put(ctx, []string{"tickets"}, query.ID, query)

	// Analyze query intent
	intent, err := a.analyzeIntent(ctx, query.Message)
	if err != nil {
		return nil, err
	}

	// Route to appropriate handler
	var response *SupportResponse
	switch intent {
	case "password_reset":
		response = a.handlePasswordReset(ctx, query)
	case "data_export":
		response = a.handleDataExport(ctx, query)
	case "api_error":
		response = a.handleAPIError(ctx, query)
	default:
		response = a.handleGeneral(ctx, query)
	}

	// Store response
	a.store.Put(ctx, []string{"responses"}, query.ID, response)

	return response, nil
}

func (a *CustomerSupportAgent) analyzeIntent(ctx context.Context, message string) (string, error) {
	prompt := fmt.Sprintf(`Analyze this customer support message and classify the intent.

Message: %s

Possible intents:
- password_reset: User has login/password issues
- data_export: User wants to export or download data
- api_error: Technical API issues or errors
- billing: Payment or subscription issues
- feature_request: Requesting new features
- general: General questions or other

Return only the intent label.`, message)

	req := &llm.CompletionRequest{
		Messages: []llm.Message{
			{Role: "system", Content: "You are a support intent classifier."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.3, // Lower temperature for classification
		MaxTokens:   50,
	}

	resp, err := a.llm.Complete(ctx, req)
	if err != nil {
		return "general", err
	}

	return resp.Content, nil
}

func (a *CustomerSupportAgent) handlePasswordReset(ctx context.Context, query CustomerQuery) *SupportResponse {
	return &SupportResponse{
		Message: fmt.Sprintf(`Hello %s,

I understand you're having trouble logging in. I've initiated a password reset for your account.

You should receive an email at %s within the next few minutes with instructions to reset your password.

If you don't see the email:
1. Check your spam/junk folder
2. Make sure you're checking the correct email account
3. Try adding noreply@ourservice.com to your contacts

If you still don't receive it within 15 minutes, please let us know and we'll investigate further.

Best regards,
Support Team`, query.Customer, query.Customer),
		ActionsTaken: []string{
			"Initiated password reset",
			"Sent reset email",
			"Logged support interaction",
		},
		Escalated:  false,
		Resolution: "password_reset_sent",
	}
}

func (a *CustomerSupportAgent) handleDataExport(ctx context.Context, query CustomerQuery) *SupportResponse {
	// Simulate data export using file operations tool
	fileOps := a.tools.Get("file_operations")
	if fileOps != nil {
		// Create sample CSV data
		exportData := `Date,Amount,Description,Category
2024-01-15,125.50,Software License,Technology
2024-01-20,75.00,Cloud Storage,Infrastructure
2024-02-01,250.00,API Usage,Services`

		fileOps.Execute(ctx, map[string]interface{}{
			"operation": "write",
			"path":      fmt.Sprintf("/tmp/agent/exports/%s_data.csv", query.Customer),
			"content":   exportData,
		})
	}

	return &SupportResponse{
		Message: fmt.Sprintf(`Hello %s,

I've prepared your data export in CSV format as requested.

You can download your data from the following link:
https://download.ourservice.com/exports/%s_data.csv

This link will be active for the next 7 days. The CSV file includes all your data from the current quarter formatted for easy import into spreadsheet applications.

If you need the data in a different format or have any issues accessing the file, please let me know.

Best regards,
Support Team`, query.Customer, query.ID),
		ActionsTaken: []string{
			"Generated CSV export",
			"Created download link",
			"Set 7-day expiration",
		},
		Escalated:  false,
		Resolution: "data_exported",
	}
}

func (a *CustomerSupportAgent) handleAPIError(ctx context.Context, query CustomerQuery) *SupportResponse {
	// This would typically check monitoring systems
	return &SupportResponse{
		Message: fmt.Sprintf(`Hello %s,

Thank you for reporting the API errors you're experiencing with user profile updates.

I've escalated this issue to our engineering team with high priority. Here's what we're doing:

1. Our engineers are investigating the 500 errors on the user profile update endpoint
2. We've increased monitoring on this endpoint
3. A temporary workaround: Try using smaller batch sizes (max 10 profiles per request)

We'll update you within the next 2 hours with our findings and expected resolution time.

Your ticket reference is: %s

Best regards,
Support Team`, query.Customer, query.ID),
		ActionsTaken: []string{
			"Created engineering ticket",
			"Increased API monitoring",
			"Provided workaround",
			"Set follow-up reminder",
		},
		Escalated:  true,
		Resolution: "escalated_to_engineering",
	}
}

func (a *CustomerSupportAgent) handleGeneral(ctx context.Context, query CustomerQuery) *SupportResponse {
	// Use LLM for general responses
	prompt := fmt.Sprintf(`Generate a helpful customer support response for this query:

Customer: %s
Message: %s

Be professional, helpful, and concise.`, query.Customer, query.Message)

	req := &llm.CompletionRequest{
		Messages: []llm.Message{
			{Role: "system", Content: "You are a helpful customer support agent."},
			{Role: "user", Content: prompt},
		},
	}

	resp, _ := a.llm.Complete(ctx, req)

	return &SupportResponse{
		Message:      resp.Content,
		ActionsTaken: []string{"Provided general assistance"},
		Escalated:    false,
		Resolution:   "general_response",
	}
}

// Example2_DataAnalysisPipeline demonstrates automated data analysis
func (e *RealWorldExamples) Example2_DataAnalysisPipeline() {
	fmt.Println("=== Data Analysis Pipeline Example ===\n")

	ctx := context.Background()

	// Create analysis pipeline
	pipeline := &DataAnalysisPipeline{
		llm:      e.llm,
		store:    e.store,
		tools:    e.tools,
		executor: tools.NewToolExecutor(),
	}

	// Sample sales data
	salesData := SalesData{
		Period: "Q1 2024",
		Regions: []RegionData{
			{Region: "North America", Sales: 1250000, Growth: 0.15},
			{Region: "Europe", Sales: 980000, Growth: 0.08},
			{Region: "Asia Pacific", Sales: 1450000, Growth: 0.22},
			{Region: "Latin America", Sales: 450000, Growth: -0.05},
		},
		Products: []ProductData{
			{Name: "Enterprise", Revenue: 2800000, Units: 140},
			{Name: "Professional", Revenue: 980000, Units: 280},
			{Name: "Starter", Revenue: 350000, Units: 1750},
		},
	}

	// Run analysis
	report, err := pipeline.AnalyzeSales(ctx, salesData)
	if err != nil {
		log.Fatal("Analysis failed:", err)
	}

	// Display results
	fmt.Printf("Analysis Report for %s\n", report.Period)
	fmt.Printf("=====================================\n\n")

	fmt.Printf("Key Insights:\n")
	for _, insight := range report.Insights {
		fmt.Printf("• %s\n", insight)
	}

	fmt.Printf("\nRecommendations:\n")
	for _, rec := range report.Recommendations {
		fmt.Printf("• %s\n", rec)
	}

	fmt.Printf("\nPredicted Next Quarter: $%.2f M\n", report.Forecast/1000000)
}

// DataAnalysisPipeline performs automated data analysis
type DataAnalysisPipeline struct {
	llm      llm.Client
	store    store.Store
	tools    *tools.Registry
	executor *tools.ToolExecutor
}

type SalesData struct {
	Period   string
	Regions  []RegionData
	Products []ProductData
}

type RegionData struct {
	Region string
	Sales  float64
	Growth float64
}

type ProductData struct {
	Name    string
	Revenue float64
	Units   int
}

type AnalysisReport struct {
	Period          string
	TotalRevenue    float64
	TopRegion       string
	TopProduct      string
	Insights        []string
	Recommendations []string
	Forecast        float64
	GeneratedAt     time.Time
}

func (p *DataAnalysisPipeline) AnalyzeSales(ctx context.Context, data SalesData) (*AnalysisReport, error) {
	// Store raw data
	p.store.Put(ctx, []string{"sales_data"}, data.Period, data)

	// Parallel analysis tasks
	tasks := []*tools.ToolCall{
		{
			ID:    "region_analysis",
			Tool:  &toolWrapper{fn: p.analyzeRegions, args: data.Regions},
			Input: &tools.ToolInput{},
		},
		{
			ID:    "product_analysis",
			Tool:  &toolWrapper{fn: p.analyzeProducts, args: data.Products},
			Input: &tools.ToolInput{},
		},
		{
			ID:    "forecast",
			Tool:  &toolWrapper{fn: p.generateForecast, args: data},
			Input: &tools.ToolInput{},
		},
	}

	// Execute analyses in parallel
	results, err := p.executor.ExecuteParallel(ctx, tasks)
	if err != nil {
		return nil, err
	}

	// Aggregate results
	report := &AnalysisReport{
		Period:       data.Period,
		GeneratedAt:  time.Now(),
		TotalRevenue: p.calculateTotalRevenue(data),
	}

	// Process parallel results
	for _, result := range results {
		if result.Error != nil {
			log.Printf("Task %s failed: %v", result.CallID, result.Error)
			continue
		}

		switch result.CallID {
		case "region_analysis":
			if result.Output != nil && result.Output.Result != nil {
				if analysis, ok := result.Output.Result.(RegionAnalysis); ok {
					report.TopRegion = analysis.TopRegion
					report.Insights = append(report.Insights, analysis.Insights...)
				}
			}
		case "product_analysis":
			if result.Output != nil && result.Output.Result != nil {
				if analysis, ok := result.Output.Result.(ProductAnalysis); ok {
					report.TopProduct = analysis.TopProduct
					report.Recommendations = append(report.Recommendations, analysis.Recommendations...)
				}
			}
		case "forecast":
			if result.Output != nil && result.Output.Result != nil {
				if forecast, ok := result.Output.Result.(float64); ok {
					report.Forecast = forecast
				}
			}
		}
	}

	// Generate AI insights
	aiInsights, err := p.generateAIInsights(ctx, data, report)
	if err == nil {
		report.Insights = append(report.Insights, aiInsights...)
	}

	// Store report
	p.store.Put(ctx, []string{"analysis_reports"}, data.Period, report)

	return report, nil
}

type RegionAnalysis struct {
	TopRegion string
	Insights  []string
}

func (p *DataAnalysisPipeline) analyzeRegions(ctx context.Context, args interface{}) (interface{}, error) {
	regions := args.([]RegionData)

	var topRegion string
	var maxSales float64

	for _, r := range regions {
		if r.Sales > maxSales {
			maxSales = r.Sales
			topRegion = r.Region
		}
	}

	insights := []string{
		fmt.Sprintf("%s is the top performing region with $%.2fM in sales", topRegion, maxSales/1000000),
	}

	// Find growth patterns
	for _, r := range regions {
		if r.Growth > 0.15 {
			insights = append(insights,
				fmt.Sprintf("%s shows strong growth at %.1f%%", r.Region, r.Growth*100))
		} else if r.Growth < 0 {
			insights = append(insights,
				fmt.Sprintf("%s needs attention with %.1f%% decline", r.Region, r.Growth*100))
		}
	}

	return RegionAnalysis{
		TopRegion: topRegion,
		Insights:  insights,
	}, nil
}

type ProductAnalysis struct {
	TopProduct      string
	Recommendations []string
}

func (p *DataAnalysisPipeline) analyzeProducts(ctx context.Context, args interface{}) (interface{}, error) {
	products := args.([]ProductData)

	var topProduct string
	var maxRevenue float64

	for _, prod := range products {
		if prod.Revenue > maxRevenue {
			maxRevenue = prod.Revenue
			topProduct = prod.Name
		}
	}

	recommendations := []string{
		fmt.Sprintf("Focus marketing efforts on %s tier (highest revenue)", topProduct),
	}

	// Analyze unit economics
	for _, prod := range products {
		avgPrice := prod.Revenue / float64(prod.Units)
		if avgPrice < 500 {
			recommendations = append(recommendations,
				fmt.Sprintf("Consider price optimization for %s (current avg: $%.2f)",
					prod.Name, avgPrice))
		}
	}

	return ProductAnalysis{
		TopProduct:      topProduct,
		Recommendations: recommendations,
	}, nil
}

func (p *DataAnalysisPipeline) generateForecast(ctx context.Context, args interface{}) (interface{}, error) {
	data := args.(SalesData)

	// Simple forecast based on growth trends
	totalRevenue := p.calculateTotalRevenue(data)
	avgGrowth := p.calculateAverageGrowth(data.Regions)

	forecast := totalRevenue * (1 + avgGrowth)

	return forecast, nil
}

func (p *DataAnalysisPipeline) calculateTotalRevenue(data SalesData) float64 {
	var total float64
	for _, r := range data.Regions {
		total += r.Sales
	}
	return total
}

func (p *DataAnalysisPipeline) calculateAverageGrowth(regions []RegionData) float64 {
	var totalGrowth float64
	for _, r := range regions {
		totalGrowth += r.Growth
	}
	return totalGrowth / float64(len(regions))
}

func (p *DataAnalysisPipeline) generateAIInsights(ctx context.Context, data SalesData, report *AnalysisReport) ([]string, error) {
	// Use LLM to generate additional insights
	dataJSON, _ := json.Marshal(data)

	prompt := fmt.Sprintf(`Analyze this sales data and provide 3 strategic insights:

%s

Current findings:
- Top Region: %s
- Top Product: %s
- Total Revenue: $%.2fM

Provide exactly 3 strategic insights in bullet points.`,
		string(dataJSON), report.TopRegion, report.TopProduct, report.TotalRevenue/1000000)

	req := &llm.CompletionRequest{
		Messages: []llm.Message{
			{Role: "system", Content: "You are a business analyst providing strategic insights."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.5,
		MaxTokens:   300,
	}

	resp, err := p.llm.Complete(ctx, req)
	if err != nil {
		return nil, err
	}

	// Parse insights (simplified - would normally parse more carefully)
	insights := []string{resp.Content}

	return insights, nil
}

// Example3_CodeReviewAutomation demonstrates automated code review
func (e *RealWorldExamples) Example3_CodeReviewAutomation() {
	fmt.Println("=== Automated Code Review Example ===\n")

	ctx := context.Background()

	reviewer := &CodeReviewer{
		llm:   e.llm,
		store: e.store,
		tools: e.tools,
	}

	// Sample code for review
	code := `
package main

import "fmt"

func calculateDiscount(price float64, customerType string) float64 {
    discount := 0.0

    if customerType == "gold" {
        discount = price * 0.20
    } else if customerType == "silver" {
        discount = price * 0.10
    } else if customerType == "bronze" {
        discount = price * 0.05
    }

    return price - discount
}

func main() {
    price := 100.0
    finalPrice := calculateDiscount(price, "gold")
    fmt.Println("Final price:", finalPrice)
}
`

	review, err := reviewer.ReviewCode(ctx, code, "go")
	if err != nil {
		log.Fatal("Review failed:", err)
	}

	// Display review results
	fmt.Printf("Code Review Results\n")
	fmt.Printf("==================\n\n")
	fmt.Printf("Overall Score: %d/10\n\n", review.Score)

	if len(review.Issues) > 0 {
		fmt.Printf("Issues Found:\n")
		for _, issue := range review.Issues {
			fmt.Printf("• [%s] Line %d: %s\n", issue.Severity, issue.Line, issue.Description)
			if issue.Suggestion != "" {
				fmt.Printf("  Suggestion: %s\n", issue.Suggestion)
			}
		}
	}

	if len(review.Improvements) > 0 {
		fmt.Printf("\nSuggested Improvements:\n")
		for _, imp := range review.Improvements {
			fmt.Printf("• %s\n", imp)
		}
	}

	fmt.Printf("\nSecurity Analysis: %s\n", review.SecurityStatus)
}

// CodeReviewer performs automated code reviews
type CodeReviewer struct {
	llm   llm.Client
	store store.Store
	tools *tools.Registry
}

type CodeReview struct {
	Score          int
	Issues         []CodeIssue
	Improvements   []string
	SecurityStatus string
	BestPractices  map[string]bool
}

type CodeIssue struct {
	Line        int
	Severity    string // "error", "warning", "info"
	Description string
	Suggestion  string
}

func (r *CodeReviewer) ReviewCode(ctx context.Context, code, language string) (*CodeReview, error) {
	// Multi-aspect review using LLM
	review := &CodeReview{
		BestPractices: make(map[string]bool),
	}

	// Parallel review tasks
	reviewTasks := []func() error{
		func() error {
			issues, err := r.findIssues(ctx, code, language)
			review.Issues = issues
			return err
		},
		func() error {
			improvements, err := r.suggestImprovements(ctx, code, language)
			review.Improvements = improvements
			return err
		},
		func() error {
			security, err := r.checkSecurity(ctx, code, language)
			review.SecurityStatus = security
			return err
		},
	}

	// Execute reviews
	for _, task := range reviewTasks {
		if err := task(); err != nil {
			log.Printf("Review task failed: %v", err)
		}
	}

	// Calculate overall score
	review.Score = r.calculateScore(review)

	// Store review
	r.store.Put(ctx, []string{"code_reviews"}, fmt.Sprintf("review_%d", time.Now().Unix()), review)

	return review, nil
}

func (r *CodeReviewer) findIssues(ctx context.Context, code, language string) ([]CodeIssue, error) {
	prompt := fmt.Sprintf(`Review this %s code and identify issues:

%s

For each issue found, provide:
1. Line number
2. Severity (error/warning/info)
3. Description
4. Suggestion for fix

Format as JSON array.`, language, code)

	req := &llm.CompletionRequest{
		Messages: []llm.Message{
			{Role: "system", Content: "You are an expert code reviewer."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.3,
		MaxTokens:   500,
	}

	resp, err := r.llm.Complete(ctx, req)
	if err != nil {
		return nil, err
	}

	// Parse response (simplified)
	// In production, would parse JSON response properly
	issues := []CodeIssue{
		{
			Line:        8,
			Severity:    "warning",
			Description: "String comparison for customer type is not type-safe",
			Suggestion:  "Use constants or enum for customer types",
		},
		{
			Line:        6,
			Severity:    "info",
			Description: "Magic numbers should be defined as constants",
			Suggestion:  "Define discount rates as named constants",
		},
	}

	return issues, nil
}

func (r *CodeReviewer) suggestImprovements(ctx context.Context, code, language string) ([]string, error) {
	prompt := fmt.Sprintf(`Suggest improvements for this %s code:

%s

Provide 3-5 specific improvements focusing on:
- Code structure
- Performance
- Maintainability
- Best practices`, language, code)

	req := &llm.CompletionRequest{
		Messages: []llm.Message{
			{Role: "system", Content: "You are a senior software engineer providing constructive feedback."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.5,
		MaxTokens:   400,
	}

	resp, err := r.llm.Complete(ctx, req)
	if err != nil {
		return nil, err
	}

	// Parse improvements (simplified)
	improvements := []string{
		"Consider using a map or strategy pattern for discount rates",
		"Add input validation for price and customerType",
		"Include unit tests for different customer types",
		"Add documentation comments for exported functions",
	}

	return improvements, nil
}

func (r *CodeReviewer) checkSecurity(ctx context.Context, code, language string) (string, error) {
	prompt := fmt.Sprintf(`Analyze this %s code for security issues:

%s

Check for:
- Input validation
- Potential vulnerabilities
- Security best practices

Provide a brief security assessment.`, language, code)

	req := &llm.CompletionRequest{
		Messages: []llm.Message{
			{Role: "system", Content: "You are a security expert reviewing code."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.2,
		MaxTokens:   200,
	}

	resp, err := r.llm.Complete(ctx, req)
	if err != nil {
		return "Unable to assess", err
	}

	return resp.Content, nil
}

func (r *CodeReviewer) calculateScore(review *CodeReview) int {
	score := 10

	// Deduct points for issues
	for _, issue := range review.Issues {
		switch issue.Severity {
		case "error":
			score -= 2
		case "warning":
			score--
		}
	}

	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}

	return score
}

// RunAllExamples executes all example scenarios
func RunAllExamples(apiKey string) {
	examples, err := NewRealWorldExamples(apiKey)
	if err != nil {
		log.Fatal("Failed to initialize examples:", err)
	}

	// Run examples
	examples.Example1_CustomerSupportAgent()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	examples.Example2_DataAnalysisPipeline()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	examples.Example3_CodeReviewAutomation()

	// Cleanup
	examples.store.Close()
}
