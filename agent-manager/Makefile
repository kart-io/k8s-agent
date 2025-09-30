.PHONY: help build run test clean docker-build docker-push deploy deps lint fmt

# Variables
APP_NAME=agent-manager
VERSION?=v1.0.0
DOCKER_IMAGE?=aetherius/$(APP_NAME):$(VERSION)
GO_FILES=$(shell find . -name '*.go' -not -path "./vendor/*")

# Build settings
LDFLAGS=-ldflags "-X main.version=$(VERSION)"
BUILD_DIR=./bin

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

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
	@go run ./cmd/server --config=configs/config.yaml

dev: ## Run in development mode with hot reload
	@echo "Running in development mode..."
	@air -c .air.toml || go run ./cmd/server --config=configs/config.yaml

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

# Database commands

db-setup: ## Setup database
	@echo "Setting up database..."
	@docker-compose up -d postgres redis
	@sleep 3
	@psql -h localhost -U aetherius -d aetherius -c "SELECT 1" || \
		psql -h localhost -U postgres -c "CREATE DATABASE aetherius;" || true

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