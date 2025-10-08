package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"reasoning-service-go/internal/analyzer"
	"reasoning-service-go/internal/config"
	"reasoning-service-go/internal/recommender"
	"reasoning-service-go/pkg/llm"
	"reasoning-service-go/pkg/types"
)

// Server represents the HTTP API server
type Server struct {
	config      *config.Config
	analyzer    *analyzer.RootCauseAnalyzer
	recommender *recommender.Engine
	llmClients  []llm.Client
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, llmClients []llm.Client) *Server {
	return &Server{
		config:      cfg,
		analyzer:    analyzer.NewRootCauseAnalyzer(cfg, llmClients),
		recommender: recommender.NewEngine(cfg, llmClients),
		llmClients:  llmClients,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", s.handleHealth)

	// Analysis endpoints
	mux.HandleFunc("/api/v1/analyze/root-cause", s.handleRootCauseAnalysis)

	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	fmt.Printf("Starting Reasoning Service on %s\n", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      s.loggingMiddleware(mux),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server.ListenAndServe()
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		fmt.Printf("%s %s %s %v\n", r.Method, r.URL.Path, r.RemoteAddr, duration)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := types.HealthResponse{
		Status:  "healthy",
		Service: "reasoning-service-go",
		Components: map[string]bool{
			"analyzer":    true,
			"recommender": true,
			"llm":         len(s.llmClients) > 0,
		},
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (s *Server) handleRootCauseAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req types.AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.Options.MaxRecommendations == 0 {
		req.Options.MaxRecommendations = s.config.Analysis.MaxRecommendations
	}
	if req.Options.MinConfidence == 0 {
		req.Options.MinConfidence = s.config.Analysis.MinConfidence
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(r.Context(), s.config.GetRequestTimeout())
	defer cancel()

	// Analyze root cause
	result, err := s.analyzer.Analyze(ctx, &req)
	if err != nil {
		result = &types.AnalysisResult{
			RequestID: req.RequestID,
			Status:    "failed",
			Error:     err.Error(),
		}
	}

	// Generate recommendations
	if result.Result != nil && result.Result.RootCause != nil {
		s.recommender.GenerateRecommendations(ctx, result, &req.Context)
	}

	result.ProcessingTime = time.Since(start).Seconds()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
