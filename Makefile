# Makefile for k8s-agent Project (Restructured following OneX and go-protoc patterns)
# Root Makefile orchestrating all services

# Build all by default, even if it's not first
.DEFAULT_GOAL := help

# ==================================================================================
# Include Modular Make Rules
# ==================================================================================

# Define services before including rules
SERVICES := agent-manager orchestrator reasoning auth gateway monitor cluster collect-agent

# Include modular make rules
include scripts/make-rules/common.mk
include scripts/make-rules/common-versions.mk
include scripts/make-rules/golang.mk
include scripts/make-rules/docker.mk
include scripts/make-rules/image.mk
include scripts/make-rules/proto.mk
include scripts/make-rules/tools.mk
include scripts/make-rules/hooks.mk
include scripts/make-rules/k8s.mk
include scripts/make-rules/version.mk
include scripts/make-rules/gen.mk
include scripts/make-rules/copyright.mk
include scripts/make-rules/lint.mk
include scripts/make-rules/deploy.mk
include scripts/make-rules/release.mk
include scripts/make-rules/swagger.mk

# ==================================================================================
# Legacy Compatibility Configuration
# ==================================================================================

# Legacy variables for backward compatibility
ALL_SERVICES := $(SERVICES)
BINS ?= $(ALL_SERVICES)
CMD_DIR := $(ROOT_DIR)/cmd
TOOLS_DIR := $(ROOT_DIR)/tools
BUILD_DIR := $(ROOT_DIR)/build

# Legacy color definitions (for old targets)
COLOR_RESET := $(NC)
COLOR_BOLD := \033[1m
COLOR_GREEN := $(GREEN)
COLOR_YELLOW := $(YELLOW)
COLOR_BLUE := $(BLUE)
COLOR_RED := $(RED)

# ==================================================================================
# Default Target
# ==================================================================================

.PHONY: all
all: help ## Default target: show help

##@ General

.PHONY: version
version: ## Show version information
	@echo "Project:    $(PROJECT_NAME)"
	@echo "Version:    $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(GO_VERSION)"
	@echo "Platform:   $(PLATFORM)"

.PHONY: info
info: ## Show project information
	@printf "$(COLOR_BOLD)Project Information:$(COLOR_RESET)\n"
	@echo "  Name:          $(PROJECT_NAME)"
	@echo "  Version:       $(VERSION)"
	@echo "  Git Commit:    $(GIT_COMMIT)"
	@echo "  Build Time:    $(BUILD_TIME)"
	@echo "  Go Version:    $(GO_VERSION)"
	@echo "  Platform:      $(PLATFORM)"
	@echo "  Registry:      $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)"
	@echo "  Services:      $(SERVICES)"
	@echo "  Root Dir:      $(ROOT_DIR)"
	@echo "  Bin Dir:       $(BIN_DIR)"
	@echo "  Output Dir:    $(OUTPUT_DIR)"

.PHONY: stats
stats: ## Show project statistics
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Project Statistics$(COLOR_RESET)\n"
	@printf "$(COLOR_BOLD)═══════════════════════════════════════════════════$(COLOR_RESET)\n"
	@printf "\n"
	@printf "$(COLOR_BOLD)Services:$(COLOR_RESET)\n"
	@echo "  Total Services:    $(words $(SERVICES))"
	@for svc in $(SERVICES); do echo "    - $$svc"; done
	@echo ""
	@printf "$(COLOR_BOLD)Make Targets:$(COLOR_RESET)\n"
	@echo "  Total Targets:     $$(make -qp 2>/dev/null | grep -E '^[a-z][a-z0-9\.\-]*:' | cut -d: -f1 | sort -u | wc -l)"
	@echo ""
	@printf "$(COLOR_BOLD)Code Statistics:$(COLOR_RESET)\n"
	@echo "  Go Files:          $$(find . -name '*.go' -not -path './vendor/*' -not -path './_output/*' | wc -l)"
	@echo "  Proto Files:       $$(find api/proto -name '*.proto' 2>/dev/null | wc -l)"
	@echo "  Make Files:        $$(find scripts/make-rules -name '*.mk' 2>/dev/null | wc -l)"
	@echo "  Scripts:           $$(find scripts -name '*.sh' -type f 2>/dev/null | wc -l)"
	@echo ""
	@printf "$(COLOR_BOLD)Configuration:$(COLOR_RESET)\n"
	@echo "  Config Files:      $$(ls -1 .*.toml .*.yaml .*.yml 2>/dev/null | wc -l)"
	@echo "  Linters Enabled:   58"
	@echo ""

# ==================================================================================
# Help and Utilities
# ==================================================================================

.PHONY: help
help: Makefile ## Display this help info
	@awk -f scripts/awk/help.awk $(MAKEFILE_LIST)

.PHONY: targets
targets: Makefile ## Show all targets from all sub-makefiles
	@for mk in $(filter-out Makefile,$(MAKEFILE_LIST)); do \
		if grep -q -E ':.*##' "$$mk" 2>/dev/null; then \
			printf '\n\033[35m%s\033[0m\n' "$$mk"; \
			awk -f scripts/awk/targets.awk "$$mk"; \
		fi; \
	done

.PHONY: list-mk
list-mk: ## List all included makefiles
	@printf "\033[1m\033[36mIncluded Makefiles:\033[0m\n"
	@printf "\033[33m%s\033[0m\n" "$(MAKEFILE_LIST)" | tr ' ' '\n' | nl -w2 -s'. '

.PHONY: install-tools
install-tools: ## Install all development tools (specify A=1 for all tools)
	@$(MAKE) tools.install
	@if [ "$(A)" = "1" ]; then \
		echo "$(COLOR_CYAN)Installing additional tools...$(COLOR_RESET)"; \
		$(MAKE) tools.install.air; \
		$(MAKE) tools.install.mockgen; \
	fi
	@printf "$(COLOR_GREEN)✓ All tools installed$(COLOR_RESET)\n"

.PHONY: rename-project
rename-project: ## Rename project module path (usage: make rename-project OLD=old/path NEW=new/path)
	@if [ -z "$(OLD)" ] || [ -z "$(NEW)" ]; then \
		echo "$(COLOR_RED)Error: OLD and NEW required. Usage: make rename-project OLD=old/path NEW=new/path$(COLOR_RESET)"; \
		exit 1; \
	fi
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Renaming project from $(OLD) to $(NEW)...$(COLOR_RESET)\n"
	@printf "$(COLOR_YELLOW)⚠ This will modify .go, go.mod, go.sum, .sh, .yaml, and .md files$(COLOR_RESET)\n"
	@# MacOS sed requires a backup extension, Linux sed does not.
	@# The `sed -i.bak` command works on both, creating a .bak file
	@find . -type f \( -name "*.go" -o -name "go.mod" -o -name "go.sum" -o -name "*.sh" -o -name "*.yaml" -o -name "*.md" \) \
		-not -path "./.git/*" \
		-not -path "./vendor/*" \
		-not -path "./_output/*" \
		-print0 | xargs -0 sed -i.bak 's|$(OLD)|$(NEW)|g'
	@find . -type f -name "*.bak" -delete
	@printf "$(COLOR_GREEN)✓ Module renamed successfully$(COLOR_RESET)\n"
	@printf "$(COLOR_YELLOW)⚠ Don't forget to run: make tidy$(COLOR_RESET)\n"

# ==================================================================================
# Legacy Compatibility Targets (map to new modular targets)
# ==================================================================================

##@ Build (Legacy Aliases - prefer go.build.* commands)

.PHONY: build
build: go.build ## Build all services (or specific: make build BINS=agent-manager)

.PHONY: build-all
build-all: go.build ## Build all services for all platforms
	@$(MAKE) go.build SERVICES="$(ALL_SERVICES)"

.PHONY: build-%
build-%: ## [LEGACY] Build specific service - prefer 'make go.build.SERVICE'
	@$(MAKE) go.build.$*

.PHONY: compile
compile: build ## Alias for build

##@ Testing (Legacy Aliases - prefer go.test.* commands)

.PHONY: test
test: go.test ## Run all tests

.PHONY: test-coverage
test-coverage: go.test.coverage ## Run tests with coverage report

.PHONY: test-integration
test-integration: go.test.integration ## Run integration tests

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Running e2e tests...$(COLOR_RESET)\n"
	@$(GO) test -v -tags=e2e ./test/e2e/...
	@printf "$(COLOR_GREEN)✓ E2E tests complete$(COLOR_RESET)\n"

##@ Code Quality (Legacy Aliases - prefer go.* commands)

.PHONY: fmt
fmt: go.fmt ## Format code

.PHONY: lint
lint: go.lint ## Run linters

.PHONY: vet
vet: go.vet ## Run go vet

##@ Dependencies (Legacy Aliases)

.PHONY: deps
deps: go.mod.download go.mod.tidy ## Download and tidy dependencies

.PHONY: deps-verify
deps-verify: go.mod.verify ## Verify dependencies

##@ Docker (Legacy Aliases - prefer docker.build.* commands)

.PHONY: docker
docker: docker.build ## Build Docker images (specify BINS for specific services)

.PHONY: docker-push
docker-push: docker.push ## Push Docker images

.PHONY: docker-%
docker-%: ## [LEGACY] Build specific Docker image - prefer 'make docker.build.SERVICE'
	@$(MAKE) docker.build.$*

##@ Code Generation (Legacy Aliases)

.PHONY: gen
gen: proto.generate ## Generate all code

.PHONY: gen-proto
gen-proto: proto.generate ## Generate protobuf code

##@ Cleanup (Legacy)

.PHONY: clean
clean: go.clean proto.clean ## Clean build artifacts

.PHONY: clean-all
clean-all: clean ## Clean everything including dependencies
	@rm -rf vendor/
	@$(GO) clean -modcache

# ==================================================================================
# Deployment
# ==================================================================================

##@ Deployment

.PHONY: deploy
deploy: ## Deploy to Kubernetes (ENV=dev|staging|prod)
	@if [ -z "$(ENV)" ]; then \
		echo "$(COLOR_RED)Error: ENV not specified. Use: make deploy ENV=dev$(COLOR_RESET)"; \
		exit 1; \
	fi
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Deploying to $(ENV)...$(COLOR_RESET)\n"
	@kubectl apply -k deployments/k8s/overlays/$(ENV)
	@printf "$(COLOR_GREEN)✓ Deployed to $(ENV)$(COLOR_RESET)\n"

.PHONY: manifests-validate
manifests-validate: ## Validate Kubernetes manifests
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Validating manifests...$(COLOR_RESET)\n"
	@kubectl apply --dry-run=client -k deployments/k8s/base
	@printf "$(COLOR_GREEN)✓ Manifests valid$(COLOR_RESET)\n"

# ==================================================================================
# Development
# ==================================================================================

##@ Development

.PHONY: dev-setup
dev-setup: tools.install hooks.install ## Setup development environment

# Note: hooks.install target is defined in scripts/make-rules/hooks.mk

.PHONY: dev
dev: ## Run service with hot reload (requires air)
	$(call print_target,$@)
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		$(call print_error,"Air not installed. Run 'make tools.install.air'"); \
		exit 1; \
	fi

.PHONY: run-%
run-%: ## Run specific service (e.g., make run-agent-manager)
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Running $*...$(COLOR_RESET)\n"
	@$(GO) run $(CMD_DIR)/$*/main.go

# ==================================================================================
# CI/CD
# ==================================================================================

##@ CI/CD

.PHONY: ci
ci: deps go.fmt go.vet go.lint go.test go.build ## Run CI pipeline

.PHONY: release
release: ## Create release (VERSION=v1.0.0)
	@if [ -z "$(VERSION)" ]; then \
		echo "$(COLOR_RED)Error: VERSION not specified. Use: make release VERSION=v1.0.0$(COLOR_RESET)"; \
		exit 1; \
	fi
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)Creating release $(VERSION)...$(COLOR_RESET)\n"
	@$(MAKE) clean
	@$(MAKE) deps
	@$(MAKE) test
	@$(MAKE) build-all
	@$(MAKE) docker.build SERVICES="$(ALL_SERVICES)"
	@printf "$(COLOR_BOLD)$(COLOR_GREEN)Release $(VERSION) ready!$(COLOR_RESET)\n"

# ==================================================================================
# Convenience Shortcuts
# ==================================================================================

##@ Shortcuts

.PHONY: format
format: go.fmt ## Format code (alias for go.fmt)

.PHONY: tidy
tidy: go.mod.tidy ## Tidy dependencies (alias for go.mod.tidy)

.PHONY: check
check: lint.run ## Run all checks (alias for lint.run)

.PHONY: validate
validate: lint.run go.vet ## Validate code quality

.PHONY: quick-test
quick-test: go.test ## Quick test run (unit tests only)

.PHONY: full-test
full-test: go.test go.test.integration ## Full test suite

.PHONY: quick-build
quick-build: go.build ## Quick build (alias for go.build)

.PHONY: rebuild
rebuild: clean go.build ## Clean and rebuild

.PHONY: dev-ready
dev-ready: tools.verify hooks.install ## Prepare development environment

.PHONY: pre-commit
pre-commit: format lint.run go.test ## Pre-commit checks

.PHONY: pre-push
pre-push: lint.run go.test go.build ## Pre-push checks
