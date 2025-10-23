# Tools installation rules
# Based on OneX project structure

##@ Tools

# Include common-versions.mk for version constants
-include scripts/make-rules/common-versions.mk

# Tool versions (override common-versions.mk if needed)
GOLANGCI_LINT_VERSION ?= v1.55.2
BUF_VERSION ?= 1.28.1
PROTOC_GEN_GO_VERSION ?= v1.31.0
PROTOC_GEN_GO_GRPC_VERSION ?= v1.3.0
PROTOC_GEN_GRPC_GATEWAY_VERSION ?= v2.18.1
PROTOC_GEN_OPENAPIV2_VERSION ?= v2.18.1
AIR_VERSION ?= v1.49.0
MOCKGEN_VERSION ?= v0.4.0
KUBE_LINTER_VERSION ?= 0.6.8
UPLIFT_VERSION ?= latest
GO_SWAGGER_VERSION ?= v0.31.0
GIT_CHGLOG_VERSION ?= v0.15.4

# Tool binaries
GOBIN := $(shell go env GOPATH)/bin
GOLANGCI_LINT := $(GOBIN)/golangci-lint
BUF := $(GOBIN)/buf
PROTOC_GEN_GO := $(GOBIN)/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(GOBIN)/protoc-gen-go-grpc
PROTOC_GEN_GRPC_GATEWAY := $(GOBIN)/protoc-gen-grpc-gateway
PROTOC_GEN_OPENAPIV2 := $(GOBIN)/protoc-gen-openapiv2
AIR := $(GOBIN)/air
MOCKGEN := $(GOBIN)/mockgen
KUBE_LINTER := $(GOBIN)/kube-linter
UPLIFT := $(GOBIN)/uplift
SWAGGER := $(GOBIN)/swagger
GIT_CHGLOG := $(GOBIN)/git-chglog

.PHONY: tools.install
tools.install: tools.install.golangci-lint tools.install.buf tools.install.protoc-plugins tools.install.air tools.install.mockgen tools.install.kube-linter tools.install.uplift tools.install.swagger tools.install.git-chglog ## Install all development tools
	$(call print_info,"All tools installed successfully")

.PHONY: tools.install.golangci-lint
tools.install.golangci-lint: ## Install golangci-lint
	$(call print_target,$@)
	@if [ ! -f "$(GOLANGCI_LINT)" ] || [ "$$($(GOLANGCI_LINT) version 2>/dev/null | grep -o 'v[0-9.]*' || echo '')" != "$(GOLANGCI_LINT_VERSION)" ]; then \
		$(call print_info,"Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."); \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCI_LINT_VERSION); \
	else \
		$(call print_info,"golangci-lint $(GOLANGCI_LINT_VERSION) already installed"); \
	fi

.PHONY: tools.install.buf
tools.install.buf: ## Install Buf
	$(call print_target,$@)
	@if [ ! -f "$(BUF)" ] || [ "$$($(BUF) --version 2>/dev/null | grep -o '[0-9.]*' | head -1 || echo '')" != "$(BUF_VERSION)" ]; then \
		$(call print_info,"Installing Buf $(BUF_VERSION)..."); \
		go install github.com/bufbuild/buf/cmd/buf@v$(BUF_VERSION); \
	else \
		$(call print_info,"Buf $(BUF_VERSION) already installed"); \
	fi

.PHONY: tools.install.protoc-plugins
tools.install.protoc-plugins: ## Install protoc plugins
	$(call print_target,$@)
	@$(call print_info,"Installing protoc-gen-go $(PROTOC_GEN_GO_VERSION)...")
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	@$(call print_info,"Installing protoc-gen-go-grpc $(PROTOC_GEN_GO_GRPC_VERSION)...")
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)
	@$(call print_info,"Installing protoc-gen-grpc-gateway $(PROTOC_GEN_GRPC_GATEWAY_VERSION)...")
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@$(PROTOC_GEN_GRPC_GATEWAY_VERSION)
	@$(call print_info,"Installing protoc-gen-openapiv2 $(PROTOC_GEN_OPENAPIV2_VERSION)...")
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@$(PROTOC_GEN_OPENAPIV2_VERSION)
	@$(call print_info,"Protoc plugins installed successfully")

.PHONY: tools.install.air
tools.install.air: ## Install Air (hot reload)
	$(call print_target,$@)
	@if [ ! -f "$(AIR)" ]; then \
		$(call print_info,"Installing Air $(AIR_VERSION)..."); \
		go install github.com/cosmtrek/air@$(AIR_VERSION); \
	else \
		$(call print_info,"Air already installed"); \
	fi

.PHONY: tools.install.mockgen
tools.install.mockgen: ## Install mockgen (mock generator)
	$(call print_target,$@)
	@if [ ! -f "$(MOCKGEN)" ]; then \
		$(call print_info,"Installing mockgen $(MOCKGEN_VERSION)..."); \
		go install go.uber.org/mock/mockgen@$(MOCKGEN_VERSION); \
	else \
		$(call print_info,"mockgen already installed"); \
	fi

.PHONY: tools.install.kube-linter
tools.install.kube-linter: ## Install kube-linter for Kubernetes manifest validation
	$(call print_target,$@)
	@if ! command -v kube-linter >/dev/null 2>&1; then \
		$(call print_info,"Installing kube-linter $(KUBE_LINTER_VERSION)..."); \
		go install golang.stackrox.io/kube-linter/cmd/kube-linter@v$(KUBE_LINTER_VERSION); \
	else \
		$(call print_info,"kube-linter already installed"); \
	fi

.PHONY: tools.install.uplift
tools.install.uplift: ## Install uplift for semantic versioning
	$(call print_target,$@)
	@if ! command -v uplift >/dev/null 2>&1; then \
		$(call print_info,"Installing uplift $(UPLIFT_VERSION)..."); \
		go install github.com/gembaadvantage/uplift@$(UPLIFT_VERSION); \
	else \
		$(call print_info,"uplift already installed"); \
	fi

.PHONY: tools.install.swagger
tools.install.swagger: ## Install go-swagger for API documentation
	$(call print_target,$@)
	@if ! command -v swagger >/dev/null 2>&1; then \
		$(call print_info,"Installing go-swagger $(GO_SWAGGER_VERSION)..."); \
		go install github.com/go-swagger/go-swagger/cmd/swagger@$(GO_SWAGGER_VERSION); \
	else \
		$(call print_info,"swagger already installed"); \
	fi

.PHONY: tools.install.git-chglog
tools.install.git-chglog: ## Install git-chglog for changelog generation
	$(call print_target,$@)
	@if ! command -v git-chglog >/dev/null 2>&1; then \
		$(call print_info,"Installing git-chglog $(GIT_CHGLOG_VERSION)..."); \
		go install github.com/git-chglog/git-chglog/cmd/git-chglog@$(GIT_CHGLOG_VERSION); \
	else \
		$(call print_info,"git-chglog already installed"); \
	fi

.PHONY: tools.verify.swagger
tools.verify.swagger: ## Verify swagger is installed
	@command -v swagger >/dev/null 2>&1 || $(MAKE) tools.install.swagger

.PHONY: tools.verify.git-chglog
tools.verify.git-chglog: ## Verify git-chglog is installed
	@command -v git-chglog >/dev/null 2>&1 || $(MAKE) tools.install.git-chglog

.PHONY: tools.verify.uplift
tools.verify.uplift: ## Verify uplift is installed
	@command -v uplift >/dev/null 2>&1 || $(MAKE) tools.install.uplift

.PHONY: tools.verify
tools.verify: ## Verify all tools are installed
	$(call print_target,$@)
	@echo "Checking installed tools..."
	@command -v golangci-lint >/dev/null 2>&1 && echo "✓ golangci-lint: $$(golangci-lint version)" || echo "✗ golangci-lint not found"
	@command -v buf >/dev/null 2>&1 && echo "✓ buf: $$(buf --version)" || echo "✗ buf not found"
	@command -v protoc-gen-go >/dev/null 2>&1 && echo "✓ protoc-gen-go: installed" || echo "✗ protoc-gen-go not found"
	@command -v protoc-gen-go-grpc >/dev/null 2>&1 && echo "✓ protoc-gen-go-grpc: installed" || echo "✗ protoc-gen-go-grpc not found"
	@command -v protoc-gen-grpc-gateway >/dev/null 2>&1 && echo "✓ protoc-gen-grpc-gateway: installed" || echo "✗ protoc-gen-grpc-gateway not found"
	@command -v protoc-gen-openapiv2 >/dev/null 2>&1 && echo "✓ protoc-gen-openapiv2: installed" || echo "✗ protoc-gen-openapiv2 not found"
	@command -v air >/dev/null 2>&1 && echo "✓ air: $$(air -v)" || echo "✗ air not found"
	@command -v mockgen >/dev/null 2>&1 && echo "✓ mockgen: installed" || echo "✗ mockgen not found"
	@command -v kube-linter >/dev/null 2>&1 && echo "✓ kube-linter: $$(kube-linter version)" || echo "✗ kube-linter not found"
	@command -v uplift >/dev/null 2>&1 && echo "✓ uplift: $$(uplift version)" || echo "✗ uplift not found"
	@command -v swagger >/dev/null 2>&1 && echo "✓ swagger: $$(swagger version)" || echo "✗ swagger not found"
	@command -v git-chglog >/dev/null 2>&1 && echo "✓ git-chglog: $$(git-chglog --version)" || echo "✗ git-chglog not found"

.PHONY: tools.clean
tools.clean: ## Remove all installed tools
	$(call print_target,$@)
	$(call print_warning,"This will remove all development tools")
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		rm -f $(GOLANGCI_LINT) $(BUF) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC) \
			$(PROTOC_GEN_GRPC_GATEWAY) $(PROTOC_GEN_OPENAPIV2) $(AIR) $(MOCKGEN) \
			$(KUBE_LINTER) $(UPLIFT) $(SWAGGER) $(GIT_CHGLOG); \
		$(call print_success,"All tools removed"); \
	fi

