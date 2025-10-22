# Makefile for k8s-agent Project (Restructured following onex v2 pattern)
# Root Makefile orchestrating all services

# ==================================================================================
# Project Configuration
# ==================================================================================

PROJECT_NAME := k8s-agent
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Docker configuration
DOCKER_REGISTRY ?= docker.io
DOCKER_NAMESPACE ?= aetherius
IMAGE_TAG ?= $(VERSION)

# Build configuration
GO := go
GOFLAGS := -trimpath
LDFLAGS := -s -w \
	-X 'github.com/kart-io/k8s-agent/internal/pkg/version.Version=$(VERSION)' \
	-X 'github.com/kart-io/k8s-agent/internal/pkg/version.GitCommit=$(GIT_COMMIT)' \
	-X 'github.com/kart-io/k8s-agent/internal/pkg/version.BuildTime=$(BUILD_TIME)'

# Directories
ROOT_DIR := $(shell pwd)
BIN_DIR := $(ROOT_DIR)/bin
BUILD_DIR := $(ROOT_DIR)/build
CMD_DIR := $(ROOT_DIR)/cmd
TOOLS_DIR := $(ROOT_DIR)/tools

# Services (can be specified with BINS variable)
ALL_SERVICES := agent-manager orchestrator reasoning auth gateway monitor cluster collect-agent
BINS ?= $(ALL_SERVICES)

# Platform configuration
PLATFORMS ?= linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

# Colors for output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[36m
COLOR_RED := \033[31m

# ==================================================================================
# Default Target
# ==================================================================================

.PHONY: all
all: help

# ==================================================================================
# Help
# ==================================================================================

.PHONY: help
help: ## Display this help message
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)╔══════════════════════════════════════════════════════════════╗$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)║     K8s-Agent Monorepo (onex v2 pattern)                     ║$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)╚══════════════════════════════════════════════════════════════╝$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Project: $(PROJECT_NAME)$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)Version: $(VERSION)$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)Commit:  $(GIT_COMMIT)$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Services:$(COLOR_RESET) $(ALL_SERVICES)"
	@echo ""
	@echo "$(COLOR_BOLD)Available targets:$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)Build Commands:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -v grep | grep -E '^(build|compile|deps|clean|gen)' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_YELLOW)%-30s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)Development:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -v grep | grep -E '^(run|dev|test|fmt|lint|vet)' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_YELLOW)%-30s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)Docker:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -v grep | grep -E '^(docker|image)' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_YELLOW)%-30s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)Deployment:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -v grep | grep -E '^(deploy|manifests|k8s)' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_YELLOW)%-30s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_BOLD)Examples:$(COLOR_RESET)"
	@echo "  make build                           # Build all services"
	@echo "  make build BINS=agent-manager        # Build specific service"
	@echo "  make docker BINS=\"agent-manager orchestrator\" # Build specific Docker images"
	@echo "  make test                            # Run all tests"
	@echo "  make deploy ENV=dev                  # Deploy to dev environment"
	@echo ""

# ==================================================================================
# Build Targets
# ==================================================================================

.PHONY: build
build: ## Build all services (or specific: make build BINS=agent-manager)
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)Building services...$(COLOR_RESET)"
	@mkdir -p $(BIN_DIR)
	@for service in $(BINS); do \
		echo "$(COLOR_BLUE)Building $$service...$(COLOR_RESET)"; \
		$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$$service $(CMD_DIR)/$$service || exit 1; \
		echo "$(COLOR_GREEN)✓ Built $$service$(COLOR_RESET)"; \
	done
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)Build complete!$(COLOR_RESET)"

.PHONY: build-all
build-all: ## Build all services for all platforms
	@$(MAKE) build BINS="$(ALL_SERVICES)"

.PHONY: build-%
build-%: ## Build specific service (e.g., make build-agent-manager)
	@$(MAKE) build BINS=$*

.PHONY: compile
compile: build ## Alias for build

# ==================================================================================
# Dependencies
# ==================================================================================

.PHONY: deps
deps: ## Download dependencies
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Downloading dependencies...$(COLOR_RESET)"
	@$(GO) mod download
	@$(GO) mod tidy
	@echo "$(COLOR_GREEN)✓ Dependencies updated$(COLOR_RESET)"

.PHONY: deps-verify
deps-verify: ## Verify dependencies
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Verifying dependencies...$(COLOR_RESET)"
	@$(GO) mod verify
	@echo "$(COLOR_GREEN)✓ Dependencies verified$(COLOR_RESET)"

# ==================================================================================
# Code Generation
# ==================================================================================

.PHONY: gen
gen: gen-proto ## Generate all code

.PHONY: gen-proto
gen-proto: ## Generate protobuf code
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Generating protobuf code...$(COLOR_RESET)"
	@cd api/proto && $(MAKE) gen-go
	@echo "$(COLOR_GREEN)✓ Protobuf code generated$(COLOR_RESET)"

# ==================================================================================
# Testing
# ==================================================================================

.PHONY: test
test: ## Run all tests
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Running tests...$(COLOR_RESET)"
	@$(GO) test -v -race -coverprofile=coverage.out ./...
	@echo "$(COLOR_GREEN)✓ Tests complete$(COLOR_RESET)"

.PHONY: test-coverage
test-coverage: test ## Run tests with coverage report
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Generating coverage report...$(COLOR_RESET)"
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report: coverage.html$(COLOR_RESET)"

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Running integration tests...$(COLOR_RESET)"
	@$(GO) test -v -tags=integration ./test/integration/...
	@echo "$(COLOR_GREEN)✓ Integration tests complete$(COLOR_RESET)"

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Running e2e tests...$(COLOR_RESET)"
	@$(GO) test -v -tags=e2e ./test/e2e/...
	@echo "$(COLOR_GREEN)✓ E2E tests complete$(COLOR_RESET)"

# ==================================================================================
# Code Quality
# ==================================================================================

.PHONY: fmt
fmt: ## Format code
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Formatting code...$(COLOR_RESET)"
	@$(GO) fmt ./...
	@echo "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)"

.PHONY: lint
lint: ## Run linters
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Running linters...$(COLOR_RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "$(COLOR_YELLOW)Warning: golangci-lint not installed$(COLOR_RESET)"; \
		echo "Install: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin"; \
	fi
	@echo "$(COLOR_GREEN)✓ Linting complete$(COLOR_RESET)"

.PHONY: vet
vet: ## Run go vet
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Running go vet...$(COLOR_RESET)"
	@$(GO) vet ./...
	@echo "$(COLOR_GREEN)✓ Vet complete$(COLOR_RESET)"

# ==================================================================================
# Docker Targets
# ==================================================================================

.PHONY: docker
docker: ## Build Docker images (specify BINS for specific services)
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Building Docker images...$(COLOR_RESET)"
	@for service in $(BINS); do \
		echo "$(COLOR_BLUE)Building $$service image...$(COLOR_RESET)"; \
		docker build -f $(BUILD_DIR)/docker/$$service.Dockerfile \
			-t $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$$service:$(IMAGE_TAG) \
			-t $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$$service:latest \
			--build-arg VERSION=$(VERSION) \
			--build-arg GIT_COMMIT=$(GIT_COMMIT) \
			. || exit 1; \
		echo "$(COLOR_GREEN)✓ Built $$service image$(COLOR_RESET)"; \
	done
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)Docker images built!$(COLOR_RESET)"

.PHONY: docker-push
docker-push: ## Push Docker images
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Pushing Docker images...$(COLOR_RESET)"
	@for service in $(BINS); do \
		echo "$(COLOR_BLUE)Pushing $$service...$(COLOR_RESET)"; \
		docker push $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$$service:$(IMAGE_TAG); \
		docker push $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$$service:latest; \
	done
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)Docker images pushed!$(COLOR_RESET)"

.PHONY: docker-%
docker-%: ## Build specific Docker image (e.g., make docker-agent-manager)
	@$(MAKE) docker BINS=$*

# ==================================================================================
# Deployment
# ==================================================================================

.PHONY: deploy
deploy: ## Deploy to Kubernetes (ENV=dev|staging|prod)
	@if [ -z "$(ENV)" ]; then \
		echo "$(COLOR_RED)Error: ENV not specified. Use: make deploy ENV=dev$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Deploying to $(ENV)...$(COLOR_RESET)"
	@kubectl apply -k manifests/overlays/$(ENV)
	@echo "$(COLOR_GREEN)✓ Deployed to $(ENV)$(COLOR_RESET)"

.PHONY: manifests-validate
manifests-validate: ## Validate Kubernetes manifests
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Validating manifests...$(COLOR_RESET)"
	@kubectl apply --dry-run=client -k manifests/base
	@echo "$(COLOR_GREEN)✓ Manifests valid$(COLOR_RESET)"

# ==================================================================================
# Development
# ==================================================================================

.PHONY: dev-setup
dev-setup: ## Setup development environment
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Setting up development environment...$(COLOR_RESET)"
	@$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@$(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@echo "$(COLOR_GREEN)✓ Development tools installed$(COLOR_RESET)"

.PHONY: run-%
run-%: ## Run specific service (e.g., make run-agent-manager)
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Running $*...$(COLOR_RESET)"
	@$(GO) run $(CMD_DIR)/$*/main.go

# ==================================================================================
# Migration Tools
# ==================================================================================

.PHONY: migrate
migrate: ## Run migration script (specify SERVICE=agent-manager)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(COLOR_RED)Error: SERVICE not specified. Use: make migrate SERVICE=agent-manager$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Migrating $(SERVICE)...$(COLOR_RESET)"
	@$(TOOLS_DIR)/migration/migrate-service.sh $(SERVICE)

# ==================================================================================
# Cleanup
# ==================================================================================

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Cleaning...$(COLOR_RESET)"
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out coverage.html
	@echo "$(COLOR_GREEN)✓ Clean complete$(COLOR_RESET)"

.PHONY: clean-all
clean-all: clean ## Clean everything including dependencies
	@rm -rf vendor/
	@$(GO) clean -modcache

# ==================================================================================
# CI/CD
# ==================================================================================

.PHONY: ci
ci: deps fmt vet lint test build ## Run CI pipeline

.PHONY: release
release: ## Create release (VERSION=v1.0.0)
	@if [ -z "$(VERSION)" ]; then \
		echo "$(COLOR_RED)Error: VERSION not specified. Use: make release VERSION=v1.0.0$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Creating release $(VERSION)...$(COLOR_RESET)"
	@$(MAKE) clean
	@$(MAKE) deps
	@$(MAKE) test
	@$(MAKE) build-all
	@$(MAKE) docker BINS="$(ALL_SERVICES)"
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)Release $(VERSION) ready!$(COLOR_RESET)"

# ==================================================================================
# Utility
# ==================================================================================

.PHONY: version
version: ## Show version information
	@echo "Project:    $(PROJECT_NAME)"
	@echo "Version:    $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

.PHONY: info
info: ## Show project information
	@echo "$(COLOR_BOLD)Project Information:$(COLOR_RESET)"
	@echo "  Name:          $(PROJECT_NAME)"
	@echo "  Version:       $(VERSION)"
	@echo "  Git Commit:    $(GIT_COMMIT)"
	@echo "  Build Time:    $(BUILD_TIME)"
	@echo "  Go Version:    $(shell $(GO) version)"
	@echo "  Registry:      $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)"
	@echo "  Services:      $(ALL_SERVICES)"
	@echo "  Root Dir:      $(ROOT_DIR)"
	@echo "  Bin Dir:       $(BIN_DIR)"
	@echo "  Build Dir:     $(BUILD_DIR)"
