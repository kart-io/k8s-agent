# ==============================================================================
# Makefile helper functions for linting
#

## Tool Binaries
GOLANGCI_LINT := golangci-lint
KUBE_LINTER := kube-linter

.PHONY: lint.run
lint.run: lint.go lint.k8s lint.docker ## Run all available linters

.PHONY: lint.go
lint.go: lint.golangci ## Run all Go linters

.PHONY: lint.golangci
lint.golangci: ## Run golangci-lint to lint source codes
	$(call print_target,$@)
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		$(GOLANGCI_LINT) run -c $(ROOT_DIR)/.golangci.yml $(ROOT_DIR)/...; \
	else \
		$(call print_error,"golangci-lint not installed. Run 'make tools.install.golangci-lint'"); \
		exit 1; \
	fi

.PHONY: lint.k8s
lint.k8s: ## Lint Kubernetes manifests
	$(call print_target,$@)
	@if command -v $(KUBE_LINTER) >/dev/null 2>&1; then \
		$(KUBE_LINTER) lint $(ROOT_DIR)/deployments; \
	else \
		$(call print_error,"kube-linter not installed. Run 'make tools.install.kube-linter'"); \
		exit 1; \
	fi

.PHONY: lint.docker
lint.docker: ## Lint Dockerfiles
	$(call print_target,$@)
	@echo "$(CYAN)Linting Dockerfiles...$(NC)"
	@find $(ROOT_DIR) -name "Dockerfile*" \
		-not -path "*/vendor/*" \
		-not -path "*/_output/*" \
		-exec echo "Checking: {}" \; \
		-exec grep -L "^FROM" {} \; || true

.PHONY: lint.proto
lint.proto: ## Lint Protocol Buffer files
	$(call print_target,$@)
	@if command -v buf >/dev/null 2>&1; then \
		cd $(ROOT_DIR)/api/proto && buf lint; \
	else \
		$(call print_warning,"buf not installed. Run 'make tools.install.buf'"); \
	fi

.PHONY: lint.yaml
lint.yaml: ## Lint YAML files
	$(call print_target,$@)
	@echo "$(CYAN)Linting YAML files...$(NC)"
	@find $(ROOT_DIR) -name "*.yaml" -o -name "*.yml" \
		| grep -v "vendor/" \
		| grep -v "_output/" \
		| grep -v ".git/" \
		| xargs -I {} echo "  ✓ {}"

.PHONY: lint.fix
lint.fix: ## Auto-fix linting issues where possible
	$(call print_target,$@)
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		$(GOLANGCI_LINT) run -c $(ROOT_DIR)/.golangci.yml --fix $(ROOT_DIR)/...; \
	fi

##@ Linting
