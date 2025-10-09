# Makefile for Aetherius Collect Agent

# Build variables
APP_NAME := collect-agent
VERSION ?= v1.0.0
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -w -s -X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildDate=$(BUILD_DATE)

# Docker image configuration
DOCKER_REGISTRY ?= aetherius
IMAGE_NAME ?= $(APP_NAME)
IMAGE_TAG ?= $(VERSION)
ENV ?= dev
PLATFORMS ?= linux/amd64,linux/arm64
DOCKER_IMAGE = $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
DOCKER_IMAGE_ENV = $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)-$(ENV)

# Go variables
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOVET := $(GOCMD) vet
GOFMT := gofmt
GOMOD := $(GOCMD) mod

# Directories
BUILD_DIR := build
DIST_DIR := dist
MANIFESTS_DIR := manifests

# Platforms
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

.PHONY: all
all: clean fmt vet test build

.PHONY: help
help: ## Display this help message
	@echo "Aetherius Collect Agent - Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@rm -f $(APP_NAME)
	@echo "Clean complete"

.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting code..."
	@$(GOFMT) -s -w .
	@echo "Format complete"

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	@$(GOVET) ./...
	@echo "Vet complete"

.PHONY: lint
lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	@$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

.PHONY: test-coverage
test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@$(GOTEST) -v -tags=integration ./test/integration/...
	@echo "Integration tests complete"

.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "Dependencies downloaded"

.PHONY: build
build: ## Build the binary
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) ./main.go
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

.PHONY: build-all
build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d/ -f1); \
		GOARCH=$$(echo $$platform | cut -d/ -f2); \
		output=$(DIST_DIR)/$(APP_NAME)-$$GOOS-$$GOARCH; \
		if [ $$GOOS = "windows" ]; then output=$$output.exe; fi; \
		echo "Building for $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH CGO_ENABLED=0 $(GOBUILD) \
			-ldflags="$(LDFLAGS)" \
			-o $$output ./main.go; \
	done
	@echo "Build complete for all platforms"

.PHONY: run
run: ## Run the application locally
	@echo "Running $(APP_NAME)..."
	@$(GOCMD) run ./main.go -config examples/config-development.yaml

.PHONY: docker-build
docker-build: ## Build Docker image (fast, local platform only)
	@echo "Building Docker image $(DOCKER_IMAGE) for local platform..."
	@docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

docker-buildx-setup: ## Setup Docker buildx builder for multi-platform builds
	@echo "Setting up Docker buildx builder..."
	@docker buildx create --name multiarch-builder --driver docker-container --use 2>/dev/null || \
		docker buildx use multiarch-builder 2>/dev/null || \
		(docker buildx create --name multiarch-builder --driver docker-container --use && docker buildx inspect --bootstrap)
	@echo "Docker buildx builder ready"

.PHONY: docker-build-fast
docker-build-fast: ## Fast local build with cache (recommended for development)
	@echo "Fast building Docker image $(DOCKER_IMAGE)..."
	@if docker buildx version >/dev/null 2>&1; then \
		DOCKER_BUILDKIT=1 docker buildx build \
			--build-arg BUILDKIT_INLINE_CACHE=1 \
			-t $(DOCKER_IMAGE) --load . ; \
	else \
		echo "BuildKit not available, using standard build..."; \
		docker build -t $(DOCKER_IMAGE) . ; \
	fi
	@echo "Docker image built: $(DOCKER_IMAGE)"

.PHONY: docker-buildx
docker-buildx: docker-buildx-setup ## Build multi-platform Docker image
	@echo "Building multi-platform Docker image $(DOCKER_IMAGE) for $(PLATFORMS)..."
	@docker buildx build --builder multiarch-builder --platform $(PLATFORMS) \
		-t $(DOCKER_IMAGE) --load . 2>/dev/null || \
	docker buildx build --builder multiarch-builder --platform $(PLATFORMS) \
		-t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

.PHONY: docker-buildx-push
docker-buildx-push: docker-buildx-setup ## Build and push multi-platform Docker image
	@echo "Building and pushing multi-platform Docker image $(DOCKER_IMAGE) for $(PLATFORMS)..."
	@docker buildx build --builder multiarch-builder --platform $(PLATFORMS) \
		-t $(DOCKER_IMAGE) --push .
	@echo "Docker image pushed: $(DOCKER_IMAGE)"

docker-buildx-env: docker-buildx-setup ## Build multi-platform Docker image with environment tag
	@echo "Building multi-platform Docker image $(DOCKER_IMAGE_ENV) for $(PLATFORMS)..."
	@docker buildx build --builder multiarch-builder --platform $(PLATFORMS) \
		-t $(DOCKER_IMAGE_ENV) --load . 2>/dev/null || \
	docker buildx build --builder multiarch-builder --platform $(PLATFORMS) \
		-t $(DOCKER_IMAGE_ENV) .
	@echo "Docker image built: $(DOCKER_IMAGE_ENV)"

docker-buildx-push-env: docker-buildx-setup ## Build and push multi-platform Docker image with environment tag
	@echo "Building and pushing multi-platform Docker image $(DOCKER_IMAGE_ENV) for $(PLATFORMS)..."
	@docker buildx build --builder multiarch-builder --platform $(PLATFORMS) \
		-t $(DOCKER_IMAGE_ENV) --push .
	@echo "Docker image pushed: $(DOCKER_IMAGE_ENV)"

.PHONY: docker-push
docker-push: ## Push Docker image to registry
	@echo "Pushing Docker image $(DOCKER_IMAGE)..."
	@docker push $(DOCKER_IMAGE)
	@echo "Docker image pushed"

.PHONY: docker-run
docker-run: ## Run Docker container locally
	@echo "Running Docker container..."
	@docker run --rm -it \
		-v $(HOME)/.kube:/home/appuser/.kube:ro \
		-e CENTRAL_ENDPOINT=nats://host.docker.internal:4222 \
		-p 8080:8080 \
		$(DOCKER_IMAGE)

docker-tag-env: ## Tag Docker image with environment suffix
	@echo "Tagging Docker image for environment: $(ENV)"
	@docker tag $(DOCKER_IMAGE) $(DOCKER_IMAGE_ENV)
	@echo "Tagged: $(DOCKER_IMAGE_ENV)"

docker-push-env: docker-tag-env ## Push Docker image with environment tag
	@echo "Pushing Docker image: $(DOCKER_IMAGE_ENV)"
	@docker push $(DOCKER_IMAGE_ENV)
	@echo "Pushed: $(DOCKER_IMAGE_ENV)"

.PHONY: k8s-deploy
k8s-deploy: ## Deploy to Kubernetes
	@echo "Deploying to Kubernetes..."
	@kubectl apply -f $(MANIFESTS_DIR)/
	@echo "Deployment complete"

.PHONY: k8s-delete
k8s-delete: ## Delete from Kubernetes
	@echo "Deleting from Kubernetes..."
	@kubectl delete -f $(MANIFESTS_DIR)/
	@echo "Deletion complete"

.PHONY: k8s-logs
k8s-logs: ## Show Kubernetes logs
	@kubectl -n aetherius-agent logs deployment/aetherius-agent -f

.PHONY: k8s-status
k8s-status: ## Show Kubernetes status
	@echo "=== Namespace ==="
	@kubectl get ns aetherius-agent
	@echo ""
	@echo "=== Pods ==="
	@kubectl -n aetherius-agent get pods
	@echo ""
	@echo "=== Deployment ==="
	@kubectl -n aetherius-agent get deployment
	@echo ""
	@echo "=== Services ==="
	@kubectl -n aetherius-agent get svc

.PHONY: k8s-health
k8s-health: ## Check agent health
	@echo "Checking agent health..."
	@kubectl -n aetherius-agent port-forward svc/aetherius-agent 8080:8080 &
	@sleep 2
	@curl -s http://localhost:8080/health/status | jq .
	@pkill -f "port-forward svc/aetherius-agent"

.PHONY: install
install: build ## Install binary to /usr/local/bin
	@echo "Installing $(APP_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(APP_NAME)
	@echo "Installation complete"

.PHONY: uninstall
uninstall: ## Uninstall binary from /usr/local/bin
	@echo "Uninstalling $(APP_NAME) from /usr/local/bin..."
	@sudo rm -f /usr/local/bin/$(APP_NAME)
	@echo "Uninstallation complete"

.PHONY: version
version: ## Show version information
	@echo "Application: $(APP_NAME)"
	@echo "Version:     $(VERSION)"
	@echo "Git Commit:  $(GIT_COMMIT)"
	@echo "Build Date:  $(BUILD_DATE)"

.PHONY: config-example
config-example: ## Generate example configuration
	@echo "Generating example configuration..."
	@cat > config.example.yaml <<EOF
# Aetherius Collect Agent Configuration\n\
\n\
# Cluster ID (leave empty for auto-detection)\n\
cluster_id: ""\n\
\n\
# NATS central endpoint\n\
central_endpoint: "nats://localhost:4222"\n\
\n\
# Connection settings\n\
reconnect_delay: 5s\n\
max_retries: 10\n\
\n\
# Heartbeat and metrics intervals\n\
heartbeat_interval: 30s\n\
metrics_interval: 60s\n\
\n\
# Buffer settings\n\
buffer_size: 1000\n\
\n\
# Feature flags\n\
enable_metrics: true\n\
enable_events: true\n\
\n\
# Logging\n\
log_level: "info"  # debug, info, warn, error, fatal\n\
EOF
	@echo "Example configuration created: config.example.yaml"

.PHONY: manifests-update
manifests-update: ## Update manifests with new image tag
	@echo "Updating manifests with image tag $(IMAGE_TAG)..."
	@sed -i.bak "s|image: .*$(APP_NAME):.*|image: $(DOCKER_IMAGE)|g" $(MANIFESTS_DIR)/04-deployment.yaml
	@rm -f $(MANIFESTS_DIR)/*.bak
	@echo "Manifests updated"

.PHONY: dev
dev: ## Run in development mode with auto-reload
	@echo "Starting development mode..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to simple run..."; \
		$(MAKE) run; \
	fi

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	@$(GOTEST) -bench=. -benchmem ./...

.PHONY: security-scan
security-scan: ## Run security scan with gosec
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

.PHONY: docs
docs: ## Generate documentation
	@echo "Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "Documentation server starting at http://localhost:6060/pkg/github.com/kart/k8s-agent/collect-agent/"; \
		godoc -http=:6060; \
	else \
		echo "godoc not installed. Install with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

.PHONY: ci
ci: clean deps fmt vet test build ## Run CI pipeline
	@echo "CI pipeline complete"

.PHONY: release
release: clean deps test build-all docker-build ## Create release
	@echo "Creating release $(VERSION)..."
	@echo "Release $(VERSION) created"
	@echo "Binaries available in $(DIST_DIR)/"
	@echo "Docker image: $(DOCKER_IMAGE)"

.DEFAULT_GOAL := help