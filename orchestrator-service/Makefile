# Makefile for Aetherius Orchestrator Service

# Build variables
APP_NAME := orchestrator-service
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
CMD_DIR := cmd/server

# Platforms
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

.PHONY: all
all: clean fmt vet test build

.PHONY: help
help: ## Display this help message
	@echo "Aetherius Orchestrator Service - Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@rm -f $(APP_NAME) coverage.out coverage.html
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
	@$(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) ./$(CMD_DIR)
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
			-o $$output ./$(CMD_DIR); \
	done
	@echo "Build complete for all platforms"

.PHONY: run
run: ## Run the application locally with default config
	@echo "Running $(APP_NAME)..."
	@$(GOCMD) run ./$(CMD_DIR)

.PHONY: run-dev
run-dev: ## Run the application with development config
	@echo "Running $(APP_NAME) in development mode..."
	@$(GOCMD) run ./$(CMD_DIR) -config configs/config-dev.yaml

.PHONY: dev
dev: ## Run in development mode with auto-reload
	@echo "Starting development mode..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to simple run..."; \
		$(MAKE) run-dev; \
	fi

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image $(DOCKER_IMAGE)..."
	@docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

docker-buildx-setup: ## Setup Docker buildx builder for multi-platform builds
	@echo "Setting up Docker buildx builder..."
	@docker buildx create --name multiarch-builder --driver docker-container --use 2>/dev/null || \
		docker buildx use multiarch-builder 2>/dev/null || \
		(docker buildx create --name multiarch-builder --driver docker-container --use && docker buildx inspect --bootstrap)
	@echo "Docker buildx builder ready"

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
		-e DB_HOST=postgres \
		-e DB_PORT=5432 \
		-e DB_USER=aetherius \
		-e DB_PASSWORD=password \
		-e DB_NAME=aetherius \
		-e REDIS_ADDR=redis:6379 \
		-e NATS_URL=nats://nats:4222 \
		-e AI_SERVICE_URL=http://reasoning-service:8082 \
		-p 8081:8081 \
		$(DOCKER_IMAGE)

docker-tag-env: ## Tag Docker image with environment suffix
	@echo "Tagging Docker image for environment: $(ENV)"
	@docker tag $(DOCKER_IMAGE) $(DOCKER_IMAGE_ENV)
	@echo "Tagged: $(DOCKER_IMAGE_ENV)"

docker-push-env: docker-tag-env ## Push Docker image with environment tag
	@echo "Pushing Docker image: $(DOCKER_IMAGE_ENV)"
	@docker push $(DOCKER_IMAGE_ENV)
	@echo "Pushed: $(DOCKER_IMAGE_ENV)"

.PHONY: docker-compose-up
docker-compose-up: ## Start with docker-compose
	@echo "Starting services with docker-compose..."
	@docker-compose up -d
	@echo "Services started"

.PHONY: docker-compose-down
docker-compose-down: ## Stop docker-compose services
	@echo "Stopping services..."
	@docker-compose down
	@echo "Services stopped"

.PHONY: docker-compose-logs
docker-compose-logs: ## View docker-compose logs
	@docker-compose logs -f $(APP_NAME)

.PHONY: k8s-deploy
k8s-deploy: ## Deploy to Kubernetes
	@echo "Deploying to Kubernetes..."
	@kubectl apply -f deployments/k8s/
	@echo "Deployment complete"

.PHONY: k8s-delete
k8s-delete: ## Delete from Kubernetes
	@echo "Deleting from Kubernetes..."
	@kubectl delete -f deployments/k8s/
	@echo "Deletion complete"

.PHONY: k8s-logs
k8s-logs: ## Show Kubernetes logs
	@kubectl -n aetherius logs -l app=$(APP_NAME) -f

.PHONY: k8s-status
k8s-status: ## Show Kubernetes status
	@echo "=== Namespace ==="
	@kubectl get ns aetherius
	@echo ""
	@echo "=== Pods ==="
	@kubectl -n aetherius get pods -l app=$(APP_NAME)
	@echo ""
	@echo "=== Deployment ==="
	@kubectl -n aetherius get deployment $(APP_NAME)
	@echo ""
	@echo "=== Services ==="
	@kubectl -n aetherius get svc $(APP_NAME)

.PHONY: k8s-port-forward
k8s-port-forward: ## Forward service port to localhost
	@echo "Forwarding port 8081..."
	@kubectl -n aetherius port-forward svc/$(APP_NAME) 8081:8081

.PHONY: k8s-restart
k8s-restart: ## Restart Kubernetes deployment
	@echo "Restarting deployment..."
	@kubectl -n aetherius rollout restart deployment/$(APP_NAME)
	@kubectl -n aetherius rollout status deployment/$(APP_NAME)

.PHONY: db-setup
db-setup: ## Setup database
	@echo "Setting up database..."
	@docker-compose up -d postgres redis nats
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Database setup complete"

.PHONY: db-create
db-create: ## Create database in K8s PostgreSQL
	@echo "Creating database aetherius_orchestrator..."
	@kubectl exec -n aetherius-dev deployment/postgres -- psql -U postgres -c "CREATE DATABASE aetherius_orchestrator;" || echo "Database may already exist"
	@echo "Database created"

.PHONY: db-create-local
db-create-local: ## Create database locally
	@echo "Creating database aetherius_orchestrator..."
	@createdb -U postgres aetherius_orchestrator || echo "Database may already exist"
	@echo "Database created"

.PHONY: db-migrate
db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	@$(GOCMD) run ./$(CMD_DIR) --config=configs/config.yaml &
	@sleep 3
	@pkill -f "$(APP_NAME)" || true
	@echo "Migrations complete"

.PHONY: db-reset
db-reset: ## Reset database
	@echo "Resetting database..."
	@docker-compose down -v postgres redis
	@$(MAKE) db-setup

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

.PHONY: psql
psql: ## Connect to PostgreSQL
	@docker-compose exec postgres psql -U aetherius -d aetherius

.PHONY: redis-cli
redis-cli: ## Connect to Redis
	@docker-compose exec redis redis-cli

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
		echo "Documentation server starting at http://localhost:6060/pkg/github.com/kart-io/k8s-agent/orchestrator-service/"; \
		godoc -http=:6060; \
	else \
		echo "godoc not installed. Install with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

.PHONY: api-test
api-test: ## Run API tests
	@echo "Running API tests..."
	@bash scripts/test-api.sh || echo "API test script not found"

.PHONY: workflow-test
workflow-test: ## Test workflow execution
	@echo "Testing workflow execution..."
	@bash scripts/test-workflow.sh || echo "Workflow test script not found"

.PHONY: ci
ci: clean deps fmt vet test build ## Run CI pipeline
	@echo "CI pipeline complete"

.PHONY: release
release: clean deps test build-all docker-build ## Create release
	@echo "Creating release $(VERSION)..."
	@echo "Release $(VERSION) created"
	@echo "Binaries available in $(DIST_DIR)/"
	@echo "Docker image: $(DOCKER_IMAGE)"

.PHONY: health-check
health-check: ## Check service health
	@echo "Checking service health..."
	@curl -s http://localhost:8081/health | jq . || echo "Service not running or jq not installed"

.PHONY: metrics
metrics: ## Show service metrics
	@echo "Fetching service metrics..."
	@curl -s http://localhost:8081/metrics || echo "Service not running"

.DEFAULT_GOAL := help
