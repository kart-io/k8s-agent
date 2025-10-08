package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"reasoning-service-go/internal/api"
	"reasoning-service-go/internal/config"
	"reasoning-service-go/pkg/llm"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "Path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Aetherius Reasoning Service (Go)\n")
	fmt.Printf("=================================\n")
	fmt.Printf("Config: %s\n", *configPath)
	fmt.Printf("Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)

	// Initialize LLM clients
	var llmClients []llm.Client
	if cfg.LLM.Enabled {
		fmt.Printf("\nInitializing LLM providers:\n")

		// Sort providers by priority
		sort.Slice(cfg.LLM.Providers, func(i, j int) bool {
			return cfg.LLM.Providers[i].Priority < cfg.LLM.Providers[j].Priority
		})

		for _, providerCfg := range cfg.LLM.Providers {
			if providerCfg.APIKey == "" {
				fmt.Printf("  [SKIP] %s: No API key\n", providerCfg.Name)
				continue
			}

			llmConfig := &llm.Config{
				Provider:    llm.Provider(providerCfg.Name),
				APIKey:      providerCfg.APIKey,
				BaseURL:     providerCfg.BaseURL,
				Model:       providerCfg.Model,
				MaxTokens:   providerCfg.MaxTokens,
				Temperature: providerCfg.Temperature,
				Timeout:     providerCfg.Timeout,
			}

			client, err := llm.NewClient(llmConfig)
			if err != nil {
				fmt.Printf("  [ERROR] %s: %v\n", providerCfg.Name, err)
				continue
			}

			llmClients = append(llmClients, client)
			fmt.Printf("  [OK] %s (model: %s, priority: %d)\n",
				providerCfg.Name, providerCfg.Model, providerCfg.Priority)
		}

		if len(llmClients) == 0 {
			fmt.Printf("  [WARNING] No LLM providers available\n")
		}
	} else {
		fmt.Printf("\nLLM support: Disabled\n")
	}

	// Create and start server
	fmt.Printf("\nStarting server...\n")
	server := api.NewServer(cfg, llmClients)

	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
