# Golang build rules
# Based on OneX project structure

##@ Go Build

.PHONY: go.build
go.build: ## Build Go binaries
	@printf "$(BLUE)==> go.build$(NC)\n"
	@mkdir -p $(BIN_DIR)
	@for svc in $(SERVICES); do \
		printf "$(GREEN)Building $$svc...$(NC)\n"; \
		CGO_ENABLED=0 $(GO_BUILD) -ldflags "$(LDFLAGS) -X 'github.com/kart-io/version.serviceName=$$svc'" -tags "$(BUILD_TAGS)" \
			-o $(BIN_DIR)/$$svc ./cmd/$$svc/main.go; \
	done
	@printf "$(GREEN)Build completed: $(BIN_DIR)$(NC)\n"

.PHONY: go.build.%
go.build.%: ## Build specific service binary (e.g., make go.build.agent-manager)
	@printf "$(BLUE)==> $@$(NC)\n"
	$(eval SERVICE := $(subst go.build.,,$@))
	@mkdir -p $(BIN_DIR)
	@if [ -d "cmd/$(SERVICE)" ]; then \
		printf "$(GREEN)Building $(SERVICE)...$(NC)\n"; \
		CGO_ENABLED=0 $(GO_BUILD) -ldflags "$(LDFLAGS) -X 'github.com/kart-io/version.serviceName=$(SERVICE)'" -tags "$(BUILD_TAGS)" \
			-o $(BIN_DIR)/$(SERVICE) ./cmd/$(SERVICE)/main.go; \
	else \
		printf "$(RED)Service directory cmd/$(SERVICE) not found$(NC)\n"; \
		exit 1; \
	fi

.PHONY: go.build.multiarch
go.build.multiarch: ## Build for multiple architectures
	@printf "$(BLUE)==> go.build.multiarch$(NC)\n"
	@mkdir -p $(BIN_DIR)
	@for os_arch in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64; do \
		os=$${os_arch%/*}; \
		arch=$${os_arch#*/}; \
		for svc in $(SERVICES); do \
			printf "$(GREEN)Building $$svc for $$os/$$arch...$(NC)\n"; \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch $(GO_BUILD) \
				-ldflags "$(LDFLAGS) -X 'github.com/kart-io/version.serviceName=$$svc'" -tags "$(BUILD_TAGS)" \
				-o $(BIN_DIR)/$$svc-$$os-$$arch ./cmd/$$svc/main.go; \
		done \
	done

##@ Go Test

.PHONY: go.test
go.test: ## Run all tests
	$(call print_target,$@)
	@printf "$(GREEN)Running tests...$(NC)\n"
	@$(GO_TEST) $(TEST_FLAGS) -timeout $(TEST_TIMEOUT) ./...

.PHONY: go.test.%
go.test.%: ## Run tests for specific service (e.g., make go.test.agent-manager)
	$(call print_target,$@)
	$(eval SERVICE := $(subst go.test.,,$@))
	@if [ -d "cmd/$(SERVICE)" ] || [ -d "internal/$(SERVICE)" ]; then \
		$(call print_info,"Testing $(SERVICE)..."); \
		$(GO_TEST) $(TEST_FLAGS) -timeout $(TEST_TIMEOUT) ./cmd/$(SERVICE)/... ./internal/$(SERVICE)/...; \
	else \
		$(call print_error,"Service directory cmd/$(SERVICE) or internal/$(SERVICE) not found"); \
		exit 1; \
	fi

.PHONY: go.test.coverage
go.test.coverage: ## Run tests with coverage
	$(call print_target,$@)
	@mkdir -p $(COVERAGE_DIR)
	@for svc in $(SERVICES); do \
		if [ -d "internal/$$svc" ]; then \
			$(call print_info,"Testing $$svc with coverage..."); \
			$(GO_TEST) -coverprofile=$(COVERAGE_DIR)/$$svc.out -covermode=atomic ./internal/$$svc/... && \
			$(GO) tool cover -html=$(COVERAGE_DIR)/$$svc.out -o $(COVERAGE_DIR)/$$svc.html; \
		fi \
	done
	$(call print_info,"Coverage reports generated in $(COVERAGE_DIR)")

.PHONY: go.test.integration
go.test.integration: ## Run integration tests
	$(call print_target,$@)
	@for svc in $(SERVICES); do \
		if [ -d "internal/$$svc/test/integration" ]; then \
			$(call print_info,"Running integration tests for $$svc..."); \
			$(GO_TEST) -tags=integration -timeout 10m ./internal/$$svc/test/integration/...; \
		fi \
	done

##@ Go Quality

.PHONY: go.fmt
go.fmt: ## Format Go code
	$(call print_target,$@)
	@find . -name "*.go" -not -path "./vendor/*" -not -path "./_output/*" -not -path "./api/proto/gen/*" | xargs $(GO_FMT) -s -w
	$(call print_info,"Go code formatted")

.PHONY: go.vet
go.vet: ## Run go vet
	$(call print_target,$@)
	@printf "$(GREEN)Running go vet...$(NC)\n"
	@$(GO_VET) ./...

.PHONY: go.lint
go.lint: ## Run golangci-lint
	$(call print_target,$@)
	@if command -v $(GO_LINT) >/dev/null 2>&1; then \
		printf "$(GREEN)Running golangci-lint...$(NC)\n"; \
		$(GO_LINT) run -c .golangci.yml --timeout 10m ./...; \
	else \
		printf "$(YELLOW)golangci-lint not installed, skipping lint$(NC)\n"; \
	fi

.PHONY: go.lint.fix
go.lint.fix: ## Run golangci-lint with auto-fix
	$(call print_target,$@)
	@if command -v $(GO_LINT) >/dev/null 2>&1; then \
		printf "$(GREEN)Running golangci-lint with auto-fix...$(NC)\n"; \
		$(GO_LINT) run -c .golangci.yml --fix --timeout 10m ./...; \
	else \
		printf "$(RED)golangci-lint not installed$(NC)\n"; \
		exit 1; \
	fi

##@ Go Dependencies

.PHONY: go.mod.tidy
go.mod.tidy: ## Run go mod tidy (monorepo: single go.mod at root)
	@printf "$(BLUE)==> go.mod.tidy$(NC)\n"
	@printf "$(GREEN)Tidying root go.mod...$(NC)\n"
	@$(GO_MOD) tidy
	@if [ -d "common" ] && [ -f "common/go.mod" ]; then \
		printf "$(GREEN)Tidying common/go.mod...$(NC)\n"; \
		cd common && $(GO_MOD) tidy && cd $(ROOT_DIR); \
	fi
	@if [ -d "api/proto" ] && [ -f "api/proto/go.mod" ]; then \
		printf "$(GREEN)Tidying api/proto/go.mod...$(NC)\n"; \
		cd api/proto && $(GO_MOD) tidy && cd $(ROOT_DIR); \
	fi
	@printf "$(GREEN)All go.mod files tidied$(NC)\n"

.PHONY: go.mod.download
go.mod.download: ## Download Go module dependencies
	@printf "$(BLUE)==> go.mod.download$(NC)\n"
	@printf "$(GREEN)Downloading dependencies...$(NC)\n"
	@$(GO_MOD) download

.PHONY: go.mod.verify
go.mod.verify: ## Verify Go module dependencies
	@printf "$(BLUE)==> go.mod.verify$(NC)\n"
	@printf "$(GREEN)Verifying dependencies...$(NC)\n"
	@$(GO_MOD) verify

##@ Go Clean

.PHONY: go.clean
go.clean: ## Clean build artifacts
	@printf "$(BLUE)==> go.clean$(NC)\n"
	@rm -rf $(OUTPUT_DIR)
	@$(GO) clean -cache -testcache
	@printf "$(GREEN)Build artifacts cleaned$(NC)\n"
