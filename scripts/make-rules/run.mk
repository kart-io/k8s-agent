# Run targets for services
# Supports running services with different environment configurations
#
# Usage Examples:
#   make dev                                    # Run with hot reload (requires air)
#   make run-auth                               # Run auth service with default config
#   make run-agent-manager                      # Run agent-manager with default config
#   make run-auth-local                         # Run auth service with local config
#   make run-auth-dev                           # Run auth service with dev config
#   make run-agent-manager-test                 # Run agent-manager with test config
#   make run SERVICE=auth ENV=local             # Generic: run auth with local config
#   make run SERVICE=orchestrator ENV=dev       # Generic: run orchestrator with dev config
#   make run-deps                               # Start MySQL, Redis, NATS in Docker
#   make run-check-deps                         # Check if dependencies are running
#   make run-stop-deps                          # Stop all Docker dependencies

##@ Run

# Development helper target
.PHONY: dev
dev: ## Run service with hot reload (requires air). Example: make dev
	$(call print_target,$@)
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		$(call print_error,"Air not installed. Run 'make tools.install.air'"); \
		exit 1; \
	fi

# Generic run target for any service with default config
# Examples:
#   make run-auth                  # Run auth service
#   make run-agent-manager         # Run agent-manager service
#   make run-orchestrator          # Run orchestrator service
#   make run-reasoning             # Run reasoning service
.PHONY: run-%
run-%: ## Run specific service. Examples: make run-auth | make run-agent-manager | make run-orchestrator
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Running $*...$(COLOR_RESET)\n"
	@if [ -f "$(CONFIG_DIR)/$*/config.yaml" ]; then \
		$(GO) run $(CMD_DIR)/$*/main.go -c $(CONFIG_DIR)/$*/config.yaml; \
	elif [ -f "$(CONFIG_DIR)/$*.yaml" ]; then \
		$(GO) run $(CMD_DIR)/$*/main.go -c $(CONFIG_DIR)/$*.yaml; \
	else \
		echo "$(COLOR_RED)Error: No config file found for $*$(COLOR_RESET)"; \
		echo "Tried: $(CONFIG_DIR)/$*/config.yaml and $(CONFIG_DIR)/$*.yaml"; \
		exit 1; \
	fi

# ==================================================================================
# Environment-specific run targets for services
# Examples:
#   make run-auth-local            # Run auth with local development config
#   make run-auth-dev              # Run auth with dev environment config
#   make run-auth-test             # Run auth with test environment config
#   make run-auth-prod             # Run auth with production config
# ==================================================================================

# Auth service environment targets
.PHONY: run-auth-local
run-auth-local: ## Run auth service with local config. Example: make run-auth-local
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Running auth with local config...$(COLOR_RESET)\n"
	@$(GO) run $(CMD_DIR)/auth/main.go -c $(CONFIG_DIR)/auth/config-local.yaml

.PHONY: run-auth-dev
run-auth-dev: ## Run auth service with dev config. Example: make run-auth-dev
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Running auth with dev config...$(COLOR_RESET)\n"
	@$(GO) run $(CMD_DIR)/auth/main.go -c $(CONFIG_DIR)/auth/config-dev.yaml

.PHONY: run-auth-test
run-auth-test: ## Run auth service with test config. Example: make run-auth-test
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Running auth with test config...$(COLOR_RESET)\n"
	@$(GO) run $(CMD_DIR)/auth/main.go -c $(CONFIG_DIR)/auth/config-test.yaml

.PHONY: run-auth-prod
run-auth-prod: ## Run auth service with prod config. Example: make run-auth-prod
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Running auth with prod config...$(COLOR_RESET)\n"
	@$(GO) run $(CMD_DIR)/auth/main.go -c $(CONFIG_DIR)/auth/config-prod.yaml

# Orchestrator service environment targets
.PHONY: run-orchestrator-local
run-orchestrator-local: ## Run orchestrator service with local config. Example: make run-orchestrator-local
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Running orchestrator with local config...$(COLOR_RESET)\n"
	@$(GO) run $(CMD_DIR)/orchestrator/main.go -c $(CONFIG_DIR)/orchestrator/config-local.yaml

.PHONY: run-orchestrator-dev
run-orchestrator-dev: ## Run orchestrator service with dev config. Example: make run-orchestrator-dev
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Running orchestrator with dev config...$(COLOR_RESET)\n"
	@$(GO) run $(CMD_DIR)/orchestrator/main.go -c $(CONFIG_DIR)/orchestrator/config-dev.yaml

.PHONY: run-orchestrator-test
run-orchestrator-test: ## Run orchestrator service with test config. Example: make run-orchestrator-test
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Running orchestrator with test config...$(COLOR_RESET)\n"
	@$(GO) run $(CMD_DIR)/orchestrator/main.go -c $(CONFIG_DIR)/orchestrator/config-test.yaml

.PHONY: run-orchestrator-prod
run-orchestrator-prod: ## Run orchestrator service with prod config. Example: make run-orchestrator-prod
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Running orchestrator with prod config...$(COLOR_RESET)\n"
	@$(GO) run $(CMD_DIR)/orchestrator/main.go -c $(CONFIG_DIR)/orchestrator/config-prod.yaml

# Reasoning service environment targets
.PHONY: run-reasoning-local
run-reasoning-local: ## Run reasoning service with local config. Example: make run-reasoning-local
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Running reasoning with local config...$(COLOR_RESET)\n"
	@$(GO) run $(CMD_DIR)/reasoning/main.go -c $(CONFIG_DIR)/reasoning/config-local.yaml

.PHONY: run-reasoning-dev
run-reasoning-dev: ## Run reasoning service with dev config. Example: make run-reasoning-dev
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Running reasoning with dev config...$(COLOR_RESET)\n"
	@$(GO) run $(CMD_DIR)/reasoning/main.go -c $(CONFIG_DIR)/reasoning/config-dev.yaml

.PHONY: run-reasoning-test
run-reasoning-test: ## Run reasoning service with test config. Example: make run-reasoning-test
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Running reasoning with test config...$(COLOR_RESET)\n"
	@$(GO) run $(CMD_DIR)/reasoning/main.go -c $(CONFIG_DIR)/reasoning/config-test.yaml

.PHONY: run-reasoning-prod
run-reasoning-prod: ## Run reasoning service with prod config. Example: make run-reasoning-prod
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Running reasoning with prod config...$(COLOR_RESET)\n"
	@$(GO) run $(CMD_DIR)/reasoning/main.go -c $(CONFIG_DIR)/reasoning/config-prod.yaml

# Agent Manager service environment targets
.PHONY: run-agent-manager-local
run-agent-manager-local: ## Run agent-manager service with local config. Example: make run-agent-manager-local
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Running agent-manager with local config...$(COLOR_RESET)\n"
	@$(GO) run $(CMD_DIR)/agent-manager/main.go -c $(CONFIG_DIR)/agent-manager/config-local.yaml

.PHONY: run-agent-manager-dev
run-agent-manager-dev: ## Run agent-manager service with dev config. Example: make run-agent-manager-dev
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Running agent-manager with dev config...$(COLOR_RESET)\n"
	@$(GO) run $(CMD_DIR)/agent-manager/main.go -c $(CONFIG_DIR)/agent-manager/config-dev.yaml

.PHONY: run-agent-manager-test
run-agent-manager-test: ## Run agent-manager service with test config. Example: make run-agent-manager-test
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Running agent-manager with test config...$(COLOR_RESET)\n"
	@$(GO) run $(CMD_DIR)/agent-manager/main.go -c $(CONFIG_DIR)/agent-manager/config-test.yaml

.PHONY: run-agent-manager-prod
run-agent-manager-prod: ## Run agent-manager service with prod config. Example: make run-agent-manager-prod
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Running agent-manager with prod config...$(COLOR_RESET)\n"
	@$(GO) run $(CMD_DIR)/agent-manager/main.go -c $(CONFIG_DIR)/agent-manager/config-prod.yaml

# ==================================================================================
# Generic run target with environment support
# Usage: make run SERVICE=<service> ENV=<environment>
# Examples:
#   make run SERVICE=auth ENV=local          # Run auth with local config
#   make run SERVICE=agent-manager ENV=dev   # Run agent-manager with dev config
#   make run SERVICE=orchestrator ENV=test   # Run orchestrator with test config
#   make run SERVICE=reasoning ENV=prod      # Run reasoning with prod config
#   make run SERVICE=auth                    # Run auth with default config (no ENV)
# ==================================================================================

# Generic target for running any service with environment config
# Usage: make run SERVICE=auth ENV=local|dev|test|prod
.PHONY: run
run: ## Run service with environment config (SERVICE=auth ENV=local|dev|test|prod)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(COLOR_RED)Error: SERVICE not specified. Use: make run SERVICE=auth ENV=local$(COLOR_RESET)"; \
		exit 1; \
	fi; \
	if [ -z "$(ENV)" ]; then \
		CONFIG_FILE="$(CONFIG_DIR)/$(SERVICE)/config.yaml"; \
		printf "$(COLOR_BOLD)$(COLOR_BLUE)Running $(SERVICE) with default config...$(COLOR_RESET)\n"; \
	else \
		CONFIG_FILE="$(CONFIG_DIR)/$(SERVICE)/config-$(ENV).yaml"; \
		printf "$(COLOR_BOLD)$(COLOR_BLUE)Running $(SERVICE) with $(ENV) config...$(COLOR_RESET)\n"; \
	fi; \
	if [ ! -f "$$CONFIG_FILE" ]; then \
		echo "$(COLOR_RED)Error: Config file not found: $$CONFIG_FILE$(COLOR_RESET)"; \
		exit 1; \
	fi; \
	$(GO) run $(CMD_DIR)/$(SERVICE)/main.go -c $$CONFIG_FILE

# ==================================================================================
# Helper targets for common tasks
# Examples:
#   make run-all                   # Instructions for running all services
#   make run-check-deps            # Check MySQL, Redis, NATS status
#   make run-deps                  # Start MySQL, Redis, NATS in Docker
#   make run-stop-deps             # Stop all Docker services
# ==================================================================================

# Run all core services (for local development)
# Note: This target provides instructions. For actual parallel execution,
# run each service in separate terminals or use a tool like tmux/screen
.PHONY: run-all
run-all: ## Run all core services in background (requires tmux or screen)
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Starting all core services...$(COLOR_RESET)"
	@echo "This target requires implementation based on your preference (tmux/screen/docker-compose)"
	@echo "For now, please run each service in separate terminals:"
	@echo "  Terminal 1: make run-agent-manager-local"
	@echo "  Terminal 2: make run-orchestrator-local"
	@echo "  Terminal 3: make run-reasoning-local"
	@echo "  Terminal 4: make run-auth-local"

# Check if all required services are available
.PHONY: run-check-deps
run-check-deps: ## Check if required services (MySQL, Redis, NATS) are running
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Checking required services...$(COLOR_RESET)"
	@echo -n "MySQL: "
	@if nc -z localhost 3306 2>/dev/null; then \
		echo "$(GREEN)✓ Running$(NC)"; \
	else \
		echo "$(RED)✗ Not running$(NC) - Run: cd deployments/docker-compose && docker-compose up -d mysql"; \
	fi
	@echo -n "Redis: "
	@if nc -z localhost 6379 2>/dev/null; then \
		echo "$(GREEN)✓ Running$(NC)"; \
	else \
		echo "$(RED)✗ Not running$(NC) - Run: cd deployments/docker-compose && docker-compose up -d redis"; \
	fi
	@echo -n "NATS: "
	@if nc -z localhost 4222 2>/dev/null; then \
		echo "$(GREEN)✓ Running$(NC)"; \
	else \
		echo "$(RED)✗ Not running$(NC) - Run: cd deployments/docker-compose && docker-compose up -d nats"; \
	fi

# Start required Docker services
.PHONY: run-deps
run-deps: ## Start required Docker services (MySQL, Redis, NATS)
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Starting required services...$(COLOR_RESET)"
	@cd deployments/docker-compose && docker-compose up -d mysql redis nats
	@echo "$(GREEN)Services started. Waiting for them to be ready...$(NC)"
	@sleep 5
	@$(MAKE) run-check-deps

# Stop all Docker services
.PHONY: run-stop-deps
run-stop-deps: ## Stop all Docker services
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Stopping Docker services...$(COLOR_RESET)"
	@cd deployments/docker-compose && docker-compose down
	@echo "$(GREEN)Services stopped$(NC)"