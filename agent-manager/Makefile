.PHONY: help build run test clean docker-build docker-push deploy deps lint fmt run-dev run-config run-env

# Variables
APP_NAME=agent-manager
VERSION?=v1.0.0
GO_FILES=$(shell find . -name '*.go' -not -path "./vendor/*")
MAIN_PATH=cmd/server/main.go

# Version injection variables (using kart-io/version package)
SERVICE_NAME := agent-manager
GIT_VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "unknown")
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH := $(shell git branch --show-current 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GO_VERSION := $(shell go version | awk '{print $$3}')
PLATFORM := $(shell uname -s)/$(shell uname -m)

# Docker image configuration
DOCKER_REGISTRY ?= aetherius
IMAGE_NAME ?= $(APP_NAME)
IMAGE_TAG ?= $(VERSION)
ENV ?= dev
PLATFORMS ?= linux/amd64,linux/arm64
DOCKER_IMAGE = $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
DOCKER_IMAGE_ENV = $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)-$(ENV)

# Build settings with version injection
LDFLAGS=-ldflags "\
	-X 'github.com/kart-io/version.serviceName=$(SERVICE_NAME)' \
	-X 'github.com/kart-io/version.gitVersion=$(GIT_VERSION)' \
	-X 'github.com/kart-io/version.gitCommit=$(GIT_COMMIT)' \
	-X 'github.com/kart-io/version.gitBranch=$(GIT_BRANCH)' \
	-X 'github.com/kart-io/version.buildDate=$(BUILD_DATE)' \
	-X 'github.com/kart-io/version.goVersion=$(GO_VERSION)' \
	-X 'github.com/kart-io/version.platform=$(PLATFORM)'"
BUILD_DIR=./bin

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

version: ## Display version information
	@echo "Service Name:  $(SERVICE_NAME)"
	@echo "Git Version:   $(GIT_VERSION)"
	@echo "Git Commit:    $(GIT_COMMIT)"
	@echo "Git Branch:    $(GIT_BRANCH)"
	@echo "Build Date:    $(BUILD_DATE)"
	@echo "Go Version:    $(GO_VERSION)"
	@echo "Platform:      $(PLATFORM)"

deps: ## Install dependencies
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

build: deps ## Build the application
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run: ## Run the application
	@echo "Running $(APP_NAME)..."
	@go run $(LDFLAGS) $(MAIN_PATH) --config configs/config.yaml

run-dev: ## Run with development config
	@echo "Running $(APP_NAME) with development config..."
	@go run $(LDFLAGS) $(MAIN_PATH) --config configs/config-dev.yaml

run-config: ## Run with custom config (CONFIG=path/to/config.yaml)
	@echo "Running $(APP_NAME) with custom config..."
	@go run $(LDFLAGS) $(MAIN_PATH) --config $(CONFIG)

run-env: ## Run with environment variables
	@echo "Running $(APP_NAME) with environment variables..."
	@AGENT_MANAGER_SERVER_PORT=8080 \
	AGENT_MANAGER_DATABASE_HOST=localhost \
	AGENT_MANAGER_DATABASE_PASSWORD=password \
	AGENT_MANAGER_REDIS_ADDR=localhost:6379 \
	go run $(MAIN_PATH)

dev: ## Run in development mode with hot reload
	@echo "Running in development mode..."
	@air -c .air.toml || go run ./cmd/server -c configs/config-dev.yaml

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run ./... || echo "golangci-lint not installed, skipping..."

fmt: ## Format code
	@echo "Formatting code..."
	@gofmt -s -w $(GO_FILES)
	@go mod tidy

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# Docker commands

docker-build: ## Build Docker image
	@echo "Building Docker image: $(DOCKER_IMAGE)"
	@docker build -t $(DOCKER_IMAGE) .

docker-buildx-setup: ## Setup Docker buildx builder for multi-platform builds
	@echo "Setting up Docker buildx builder..."
	@docker buildx create --name multiarch-builder --driver docker-container --use 2>/dev/null || \
		docker buildx use multiarch-builder 2>/dev/null || \
		(docker buildx create --name multiarch-builder --driver docker-container --use && docker buildx inspect --bootstrap)
	@echo "Docker buildx builder ready"

docker-buildx: docker-buildx-setup ## Build multi-platform Docker image
	@echo "Building multi-platform Docker image: $(DOCKER_IMAGE) for $(PLATFORMS)"
	@docker buildx build --builder multiarch-builder --platform $(PLATFORMS) \
		-t $(DOCKER_IMAGE) --load . 2>/dev/null || \
	docker buildx build --builder multiarch-builder --platform $(PLATFORMS) \
		-t $(DOCKER_IMAGE) .

docker-buildx-push: docker-buildx-setup ## Build and push multi-platform Docker image
	@echo "Building and pushing multi-platform Docker image: $(DOCKER_IMAGE) for $(PLATFORMS)"
	@docker buildx build --builder multiarch-builder --platform $(PLATFORMS) \
		-t $(DOCKER_IMAGE) --push .

docker-buildx-env: docker-buildx-setup ## Build multi-platform Docker image with environment tag
	@echo "Building multi-platform Docker image: $(DOCKER_IMAGE_ENV) for $(PLATFORMS)"
	@docker buildx build --builder multiarch-builder --platform $(PLATFORMS) \
		-t $(DOCKER_IMAGE_ENV) --load . 2>/dev/null || \
	docker buildx build --builder multiarch-builder --platform $(PLATFORMS) \
		-t $(DOCKER_IMAGE_ENV) .

docker-buildx-push-env: docker-buildx-setup ## Build and push multi-platform Docker image with environment tag
	@echo "Building and pushing multi-platform Docker image: $(DOCKER_IMAGE_ENV) for $(PLATFORMS)"
	@docker buildx build --builder multiarch-builder --platform $(PLATFORMS) \
		-t $(DOCKER_IMAGE_ENV) --push .

docker-push: docker-build ## Push Docker image
	@echo "Pushing Docker image: $(DOCKER_IMAGE)"
	@docker push $(DOCKER_IMAGE)

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run --rm -p 8080:8080 \
		-e DB_HOST=postgres \
		-e REDIS_ADDR=redis:6379 \
		-e NATS_URL=nats://nats:4222 \
		$(DOCKER_IMAGE)

docker-tag-env: ## Tag Docker image with environment suffix
	@echo "Tagging Docker image for environment: $(ENV)"
	@docker tag $(DOCKER_IMAGE) $(DOCKER_IMAGE_ENV)
	@echo "Tagged: $(DOCKER_IMAGE_ENV)"

docker-push-env: docker-tag-env ## Push Docker image with environment tag
	@echo "Pushing Docker image: $(DOCKER_IMAGE_ENV)"
	@docker push $(DOCKER_IMAGE_ENV)
	@echo "Pushed: $(DOCKER_IMAGE_ENV)"

# Database commands

db-setup: ## Setup database
	@echo "Setting up database..."
	@docker-compose up -d postgres redis
	@sleep 3
	@psql -h localhost -U aetherius -d aetherius -c "SELECT 1" || \
		psql -h localhost -U postgres -c "CREATE DATABASE aetherius;" || true

db-create: ## Create database in K8s PostgreSQL
	@echo "Creating database aetherius..."
	@kubectl exec -n aetherius-dev deployment/postgres -- psql -U postgres -c "CREATE DATABASE aetherius;" || echo "Database may already exist"
	@echo "Database created"

db-create-local: ## Create database locally
	@echo "Creating database aetherius..."
	@createdb -U postgres aetherius || echo "Database may already exist"
	@echo "Database created"

db-migrate: ## Run database migrations
	@echo "Running migrations..."
	@go run ./cmd/server --config=configs/config.yaml &
	@sleep 3
	@pkill -f "agent-manager"

db-reset: ## Reset database
	@echo "Resetting database..."
	@docker-compose down -v
	@$(MAKE) db-setup

# Deployment commands

deploy-local: docker-build ## Deploy locally with docker-compose
	@echo "Deploying locally..."
	@docker-compose up -d

deploy-k8s: ## Deploy to Kubernetes
	@echo "Deploying to Kubernetes..."
	@kubectl apply -f deployments/k8s/

deploy-stop: ## Stop local deployment
	@echo "Stopping deployment..."
	@docker-compose down

# Development helpers

logs: ## View application logs
	@docker-compose logs -f agent-manager

shell: ## Open shell in container
	@docker-compose exec agent-manager sh

psql: ## Connect to PostgreSQL
	@docker-compose exec postgres psql -U aetherius -d aetherius

redis-cli: ## Connect to Redis
	@docker-compose exec redis redis-cli

# CI/CD

ci: deps lint test ## Run CI checks

install: build ## Install binary
	@echo "Installing $(APP_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/

uninstall: ## Uninstall binary
	@echo "Uninstalling $(APP_NAME)..."
	@sudo rm -f /usr/local/bin/$(APP_NAME)