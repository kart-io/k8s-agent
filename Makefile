# Makefile for Aetherius - Intelligent Kubernetes Operations Platform
# Root level Makefile to orchestrate all services

# Project variables
PROJECT_NAME := aetherius
VERSION ?= v1.0.0
DOCKER_REGISTRY ?= docker.io
DOCKER_NAMESPACE ?= aetherius

# Service directories
COLLECT_AGENT_DIR := collect-agent
AGENT_MANAGER_DIR := agent-manager
ORCHESTRATOR_DIR := orchestrator-service
REASONING_DIR := reasoning-service-go

# Colors for output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[36m

.PHONY: all
all: help

.PHONY: help
help: ## Display this help message
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)╔══════════════════════════════════════════════════════════════╗$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)║  Aetherius - Intelligent Kubernetes Operations Platform     ║$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)╚══════════════════════════════════════════════════════════════╝$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Available targets:$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)Build Commands:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E '^(build|deps|clean)' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_YELLOW)%-25s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)Development:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E '^(run|dev|test|fmt|lint)' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_YELLOW)%-25s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)Docker:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E '^docker-' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_YELLOW)%-25s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)Deployment:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E '^(deploy|k8s-)' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_YELLOW)%-25s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)Service-Specific:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E '^(collect-agent|agent-manager|orchestrator|reasoning)-' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_YELLOW)%-25s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_BOLD)Version: $(VERSION)$(COLOR_RESET)"

# ============================================================================
# Global Build Commands
# ============================================================================

.PHONY: deps
deps: ## Install all dependencies
	@echo "$(COLOR_BOLD)Installing dependencies for all services...$(COLOR_RESET)"
	@$(MAKE) -C $(COLLECT_AGENT_DIR) deps
	@$(MAKE) -C $(AGENT_MANAGER_DIR) deps
	@$(MAKE) -C $(ORCHESTRATOR_DIR) deps
	@cd $(REASONING_DIR) && pip install -r requirements.txt
	@echo "$(COLOR_GREEN)✓ All dependencies installed$(COLOR_RESET)"

.PHONY: build
build: ## Build all services
	@echo "$(COLOR_BOLD)Building all services...$(COLOR_RESET)"
	@$(MAKE) -C $(COLLECT_AGENT_DIR) build
	@$(MAKE) -C $(AGENT_MANAGER_DIR) build
	@$(MAKE) -C $(ORCHESTRATOR_DIR) build
	@echo "$(COLOR_GREEN)✓ All services built$(COLOR_RESET)"

.PHONY: clean
clean: ## Clean all build artifacts
	@echo "$(COLOR_BOLD)Cleaning all build artifacts...$(COLOR_RESET)"
	@$(MAKE) -C $(COLLECT_AGENT_DIR) clean
	@$(MAKE) -C $(AGENT_MANAGER_DIR) clean
	@$(MAKE) -C $(ORCHESTRATOR_DIR) clean
	@cd $(REASONING_DIR) && $(MAKE) clean
	@echo "$(COLOR_GREEN)✓ Cleanup complete$(COLOR_RESET)"

# ============================================================================
# Testing
# ============================================================================

.PHONY: test
test: ## Run tests for all services
	@echo "$(COLOR_BOLD)Running tests for all services...$(COLOR_RESET)"
	@$(MAKE) -C $(COLLECT_AGENT_DIR) test
	@$(MAKE) -C $(AGENT_MANAGER_DIR) test
	@$(MAKE) -C $(ORCHESTRATOR_DIR) test
	@cd $(REASONING_DIR) && pytest tests/ -v
	@echo "$(COLOR_GREEN)✓ All tests passed$(COLOR_RESET)"

.PHONY: test-coverage
test-coverage: ## Generate coverage reports for all services
	@echo "$(COLOR_BOLD)Generating coverage reports...$(COLOR_RESET)"
	@$(MAKE) -C $(COLLECT_AGENT_DIR) test-coverage
	@$(MAKE) -C $(AGENT_MANAGER_DIR) test-coverage
	@$(MAKE) -C $(ORCHESTRATOR_DIR) test-coverage
	@cd $(REASONING_DIR) && pytest tests/ -v --cov=internal --cov-report=html
	@echo "$(COLOR_GREEN)✓ Coverage reports generated$(COLOR_RESET)"

# ============================================================================
# Code Quality
# ============================================================================

.PHONY: fmt
fmt: ## Format all code
	@echo "$(COLOR_BOLD)Formatting all code...$(COLOR_RESET)"
	@$(MAKE) -C $(COLLECT_AGENT_DIR) fmt
	@$(MAKE) -C $(AGENT_MANAGER_DIR) fmt
	@$(MAKE) -C $(ORCHESTRATOR_DIR) fmt
	@cd $(REASONING_DIR) && black internal/ pkg/ cmd/ && isort internal/ pkg/ cmd/
	@echo "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)"

.PHONY: lint
lint: ## Run linters on all code
	@echo "$(COLOR_BOLD)Running linters...$(COLOR_RESET)"
	@$(MAKE) -C $(COLLECT_AGENT_DIR) lint
	@$(MAKE) -C $(AGENT_MANAGER_DIR) lint
	@$(MAKE) -C $(ORCHESTRATOR_DIR) lint
	@cd $(REASONING_DIR) && pylint internal/ pkg/ cmd/ || true
	@echo "$(COLOR_GREEN)✓ Linting complete$(COLOR_RESET)"

.PHONY: vet
vet: ## Run go vet on all Go services
	@echo "$(COLOR_BOLD)Running go vet...$(COLOR_RESET)"
	@$(MAKE) -C $(COLLECT_AGENT_DIR) vet
	@$(MAKE) -C $(AGENT_MANAGER_DIR) vet
	@$(MAKE) -C $(ORCHESTRATOR_DIR) vet
	@echo "$(COLOR_GREEN)✓ Vet complete$(COLOR_RESET)"

# ============================================================================
# Docker Commands
# ============================================================================

.PHONY: docker-build
docker-build: ## Build all Docker images
	@echo "$(COLOR_BOLD)Building all Docker images...$(COLOR_RESET)"
	@$(MAKE) -C $(COLLECT_AGENT_DIR) docker-build VERSION=$(VERSION)
	@$(MAKE) -C $(AGENT_MANAGER_DIR) docker-build VERSION=$(VERSION)
	@$(MAKE) -C $(ORCHESTRATOR_DIR) docker-build VERSION=$(VERSION)
	@cd $(REASONING_DIR) && docker build -t $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/reasoning-service:$(VERSION) .
	@echo "$(COLOR_GREEN)✓ All Docker images built$(COLOR_RESET)"

.PHONY: docker-buildx
docker-buildx: ## Build multi-platform Docker images (linux/amd64,linux/arm64)
	@echo "$(COLOR_BOLD)Building multi-platform Docker images...$(COLOR_RESET)"
	@./scripts/docker-buildx.sh collect-agent -v $(VERSION)
	@./scripts/docker-buildx.sh agent-manager -v $(VERSION)
	@./scripts/docker-buildx.sh orchestrator-service -v $(VERSION)
	@./scripts/docker-buildx.sh gateway-service -v $(VERSION)
	@./scripts/docker-buildx.sh auth-service -v $(VERSION)
	@echo "$(COLOR_GREEN)✓ All multi-platform images built$(COLOR_RESET)"

.PHONY: docker-buildx-push
docker-buildx-push: ## Build and push multi-platform Docker images
	@echo "$(COLOR_BOLD)Building and pushing multi-platform Docker images...$(COLOR_RESET)"
	@./scripts/docker-buildx.sh collect-agent -v $(VERSION) --push
	@./scripts/docker-buildx.sh agent-manager -v $(VERSION) --push
	@./scripts/docker-buildx.sh orchestrator-service -v $(VERSION) --push
	@./scripts/docker-buildx.sh gateway-service -v $(VERSION) --push
	@./scripts/docker-buildx.sh auth-service -v $(VERSION) --push
	@echo "$(COLOR_GREEN)✓ All multi-platform images built and pushed$(COLOR_RESET)"

.PHONY: docker-push
docker-push: ## Push all Docker images to registry
	@echo "$(COLOR_BOLD)Pushing all Docker images...$(COLOR_RESET)"
	@$(MAKE) -C $(COLLECT_AGENT_DIR) docker-push VERSION=$(VERSION)
	@$(MAKE) -C $(AGENT_MANAGER_DIR) docker-push VERSION=$(VERSION)
	@$(MAKE) -C $(ORCHESTRATOR_DIR) docker-push VERSION=$(VERSION)
	@docker push $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/reasoning-service:$(VERSION)
	@echo "$(COLOR_GREEN)✓ All Docker images pushed$(COLOR_RESET)"

.PHONY: docker-compose-up
docker-compose-up: ## Start all services with docker-compose
	@echo "$(COLOR_BOLD)Starting all services with docker-compose...$(COLOR_RESET)"
	@cd deployments/docker-compose && docker-compose up -d
	@echo "$(COLOR_GREEN)✓ All services started$(COLOR_RESET)"
	@echo ""
	@echo "Services available at:"
	@echo "  - Agent Manager:       http://localhost:8080"
	@echo "  - Orchestrator:        http://localhost:8081"
	@echo "  - Reasoning Service:   http://localhost:8082"
	@echo "  - Grafana:             http://localhost:3000 (admin/admin)"
	@echo "  - Prometheus:          http://localhost:9090"
	@echo "  - Neo4j:               http://localhost:7474"

.PHONY: docker-compose-down
docker-compose-down: ## Stop all docker-compose services
	@echo "$(COLOR_BOLD)Stopping all services...$(COLOR_RESET)"
	@cd deployments/docker-compose && docker-compose down
	@echo "$(COLOR_GREEN)✓ All services stopped$(COLOR_RESET)"

.PHONY: docker-compose-logs
docker-compose-logs: ## View docker-compose logs
	@cd deployments/docker-compose && docker-compose logs -f

.PHONY: docker-compose-ps
docker-compose-ps: ## Show docker-compose service status
	@cd deployments/docker-compose && docker-compose ps

# ============================================================================
# Kubernetes Deployment
# ============================================================================

.PHONY: k8s-deploy
k8s-deploy: ## Deploy all services to Kubernetes
	@echo "$(COLOR_BOLD)Deploying to Kubernetes...$(COLOR_RESET)"
	@kubectl apply -f deployments/k8s/namespace.yaml
	@kubectl apply -f deployments/k8s/dependencies.yaml
	@echo "Waiting for dependencies to be ready..."
	@kubectl -n aetherius wait --for=condition=ready pod -l app=postgres --timeout=300s
	@kubectl apply -f deployments/k8s/agent-manager.yaml
	@kubectl apply -f deployments/k8s/orchestrator-service.yaml
	@kubectl apply -f deployments/k8s/reasoning-service.yaml
	@echo "$(COLOR_GREEN)✓ Deployment complete$(COLOR_RESET)"

.PHONY: k8s-delete
k8s-delete: ## Delete all services from Kubernetes
	@echo "$(COLOR_BOLD)Deleting from Kubernetes...$(COLOR_RESET)"
	@kubectl delete -f deployments/k8s/ || true
	@echo "$(COLOR_GREEN)✓ Deletion complete$(COLOR_RESET)"

.PHONY: k8s-status
k8s-status: ## Show Kubernetes deployment status
	@echo "$(COLOR_BOLD)Kubernetes Status:$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Namespaces:$(COLOR_RESET)"
	@kubectl get ns aetherius
	@echo ""
	@echo "$(COLOR_BOLD)Pods:$(COLOR_RESET)"
	@kubectl -n aetherius get pods
	@echo ""
	@echo "$(COLOR_BOLD)Services:$(COLOR_RESET)"
	@kubectl -n aetherius get svc
	@echo ""
	@echo "$(COLOR_BOLD)Deployments:$(COLOR_RESET)"
	@kubectl -n aetherius get deployments

.PHONY: k8s-logs
k8s-logs: ## Show Kubernetes logs (use SERVICE=<service-name>)
	@if [ -z "$(SERVICE)" ]; then \
		echo "Please specify SERVICE=<service-name>"; \
		echo "Available services: agent-manager, orchestrator-service, reasoning-service"; \
	else \
		kubectl -n aetherius logs -l app=$(SERVICE) -f; \
	fi

.PHONY: k8s-restart
k8s-restart: ## Restart all Kubernetes deployments
	@echo "$(COLOR_BOLD)Restarting all deployments...$(COLOR_RESET)"
	@kubectl -n aetherius rollout restart deployment
	@echo "$(COLOR_GREEN)✓ Restart initiated$(COLOR_RESET)"

.PHONY: k8s-kustomize
k8s-kustomize: ## Deploy using Kustomize (Traefik, Prometheus, Grafana)
	@echo "$(COLOR_BOLD)Deploying with Kustomize...$(COLOR_RESET)"
	@echo "For Kustomize deployments, use: cd deployments/kustomize && make help"
	@cd deployments/kustomize && $(MAKE) deploy-all

# ============================================================================
# Service-Specific Commands
# ============================================================================

.PHONY: collect-agent-build
collect-agent-build: ## Build collect-agent
	@$(MAKE) -C $(COLLECT_AGENT_DIR) build

.PHONY: collect-agent-run
collect-agent-run: ## Run collect-agent
	@$(MAKE) -C $(COLLECT_AGENT_DIR) run

.PHONY: collect-agent-test
collect-agent-test: ## Test collect-agent
	@$(MAKE) -C $(COLLECT_AGENT_DIR) test

.PHONY: agent-manager-build
agent-manager-build: ## Build agent-manager
	@$(MAKE) -C $(AGENT_MANAGER_DIR) build

.PHONY: agent-manager-run
agent-manager-run: ## Run agent-manager
	@$(MAKE) -C $(AGENT_MANAGER_DIR) run

.PHONY: agent-manager-test
agent-manager-test: ## Test agent-manager
	@$(MAKE) -C $(AGENT_MANAGER_DIR) test

.PHONY: orchestrator-build
orchestrator-build: ## Build orchestrator-service
	@$(MAKE) -C $(ORCHESTRATOR_DIR) build

.PHONY: orchestrator-run
orchestrator-run: ## Run orchestrator-service
	@$(MAKE) -C $(ORCHESTRATOR_DIR) run

.PHONY: orchestrator-test
orchestrator-test: ## Test orchestrator-service
	@$(MAKE) -C $(ORCHESTRATOR_DIR) test

.PHONY: reasoning-run
reasoning-run: ## Run reasoning-service
	@cd $(REASONING_DIR) && $(MAKE) run

.PHONY: reasoning-dev
reasoning-dev: ## Run reasoning-service in dev mode
	@cd $(REASONING_DIR) && $(MAKE) dev

.PHONY: reasoning-test
reasoning-test: ## Test reasoning-service
	@cd $(REASONING_DIR) && $(MAKE) test

# ============================================================================
# Development Helpers
# ============================================================================

.PHONY: dev-setup
dev-setup: ## Setup development environment
	@echo "$(COLOR_BOLD)Setting up development environment...$(COLOR_RESET)"
	@echo "Installing Go dependencies..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "Installing Python dependencies..."
	@pip install black isort pylint pytest pytest-cov
	@echo "$(COLOR_GREEN)✓ Development environment ready$(COLOR_RESET)"

.PHONY: dev-start
dev-start: ## Start all services in development mode
	@echo "$(COLOR_BOLD)Starting development environment...$(COLOR_RESET)"
	@$(MAKE) docker-compose-up

.PHONY: dev-stop
dev-stop: ## Stop all development services
	@$(MAKE) docker-compose-down

.PHONY: dev-restart
dev-restart: dev-stop dev-start ## Restart development environment

# ============================================================================
# Database Management
# ============================================================================

.PHONY: db-setup
db-setup: ## Setup databases
	@echo "$(COLOR_BOLD)Setting up databases...$(COLOR_RESET)"
	@cd deployments/docker-compose && docker-compose up -d postgres redis neo4j
	@echo "Waiting for databases to be ready..."
	@sleep 10
	@echo "$(COLOR_GREEN)✓ Databases ready$(COLOR_RESET)"

.PHONY: db-reset
db-reset: ## Reset all databases
	@echo "$(COLOR_BOLD)Resetting databases...$(COLOR_RESET)"
	@cd deployments/docker-compose && docker-compose down -v postgres redis neo4j
	@$(MAKE) db-setup

.PHONY: psql
psql: ## Connect to PostgreSQL
	@docker-compose -f deployments/docker-compose/docker-compose.yaml exec postgres psql -U aetherius -d aetherius

.PHONY: redis-cli
redis-cli: ## Connect to Redis
	@docker-compose -f deployments/docker-compose/docker-compose.yaml exec redis redis-cli

# ============================================================================
# Monitoring & Health
# ============================================================================

.PHONY: health-check
health-check: ## Check health of all services
	@echo "$(COLOR_BOLD)Checking service health...$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Agent Manager:$(COLOR_RESET)"
	@curl -s http://localhost:8080/health | jq . || echo "Service not available"
	@echo ""
	@echo "$(COLOR_BOLD)Orchestrator Service:$(COLOR_RESET)"
	@curl -s http://localhost:8081/health | jq . || echo "Service not available"
	@echo ""
	@echo "$(COLOR_BOLD)Reasoning Service:$(COLOR_RESET)"
	@curl -s http://localhost:8082/health | jq . || echo "Service not available"

.PHONY: logs
logs: ## Show logs for all services
	@$(MAKE) docker-compose-logs

.PHONY: monitoring-start
monitoring-start: ## Start monitoring stack (Prometheus + Grafana)
	@echo "$(COLOR_BOLD)Starting monitoring stack...$(COLOR_RESET)"
	@cd deployments/docker-compose && docker-compose up -d prometheus grafana
	@echo "$(COLOR_GREEN)✓ Monitoring stack started$(COLOR_RESET)"
	@echo "  - Grafana:    http://localhost:3000 (admin/admin)"
	@echo "  - Prometheus: http://localhost:9090"

# ============================================================================
# CI/CD
# ============================================================================

.PHONY: ci
ci: clean deps fmt vet lint test build ## Run full CI pipeline
	@echo "$(COLOR_GREEN)✓ CI pipeline complete$(COLOR_RESET)"

.PHONY: release
release: ## Create a release (VERSION=x.x.x)
	@echo "$(COLOR_BOLD)Creating release $(VERSION)...$(COLOR_RESET)"
	@$(MAKE) clean
	@$(MAKE) deps
	@$(MAKE) test
	@$(MAKE) build
	@$(MAKE) docker-build
	@echo "$(COLOR_GREEN)✓ Release $(VERSION) created$(COLOR_RESET)"
	@echo ""
	@echo "To push to registry: make docker-push VERSION=$(VERSION)"

# ============================================================================
# Documentation
# ============================================================================

.PHONY: docs-serve
docs-serve: ## Serve documentation locally
	@echo "Starting documentation server..."
	@if command -v mkdocs >/dev/null 2>&1; then \
		mkdocs serve; \
	else \
		echo "mkdocs not installed. Install with: pip install mkdocs"; \
	fi

.PHONY: docs-build
docs-build: ## Build documentation
	@echo "Building documentation..."
	@if command -v mkdocs >/dev/null 2>&1; then \
		mkdocs build; \
	else \
		echo "mkdocs not installed. Install with: pip install mkdocs"; \
	fi

# ============================================================================
# Utility
# ============================================================================

.PHONY: version
version: ## Show version information
	@echo "Project:  $(PROJECT_NAME)"
	@echo "Version:  $(VERSION)"
	@echo "Registry: $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)"

.PHONY: info
info: ## Show project information
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)╔══════════════════════════════════════════════════════════════╗$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)║  Aetherius - Intelligent Kubernetes Operations Platform     ║$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)╚══════════════════════════════════════════════════════════════╝$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Version:$(COLOR_RESET) $(VERSION)"
	@echo ""
	@echo "$(COLOR_BOLD)Services:$(COLOR_RESET)"
	@echo "  • Collect Agent       - Edge data collection layer"
	@echo "  • Agent Manager       - Central control plane"
	@echo "  • Orchestrator        - Workflow orchestration layer"
	@echo "  • Reasoning Service   - AI intelligence layer"
	@echo ""
	@echo "$(COLOR_BOLD)Quick Start:$(COLOR_RESET)"
	@echo "  1. make dev-setup      # Setup development environment"
	@echo "  2. make dev-start      # Start all services"
	@echo "  3. make health-check   # Verify services are running"
	@echo ""
	@echo "$(COLOR_BOLD)Documentation:$(COLOR_RESET) See README.md and docs/"

.DEFAULT_GOAL := help
