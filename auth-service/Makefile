.PHONY: help build run run-dev run-local run-prod test clean init-db init-mysql deps fmt lint docker-build docker-run test-coverage vet dev

# Configuration
CONFIG ?= configs/config.yaml
CONFIG_DEV = configs/config-dev.yaml
CONFIG_LOCAL = configs/config-local.yaml
CONFIG_PROD = configs/config-prod.yaml

# Docker image configuration
DOCKER_REGISTRY ?= aetherius
IMAGE_NAME ?= auth-service
IMAGE_TAG ?= latest
ENV ?= dev
PLATFORMS ?= linux/amd64,linux/arm64
DOCKER_IMAGE = $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
DOCKER_IMAGE_ENV = $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)-$(ENV)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build the service
build: ## Build the auth-service binary
	@echo "Building auth-service..."
	@go build -o bin/auth-service cmd/server/main.go
	@echo "✅ Build complete: bin/auth-service"

# Run commands
run: ## Run with default config (config.yaml)
	@echo "🚀 Running auth-service with default config..."
	@go run cmd/server/main.go -c $(CONFIG)

run-dev: ## Run with development config (config-dev.yaml)
	@echo "🚀 Running auth-service with development config..."
	@go run cmd/server/main.go -c $(CONFIG_DEV)

run-local: ## Run with local config (config-local.yaml)
	@echo "🚀 Running auth-service with local config..."
	@go run cmd/server/main.go -c $(CONFIG_LOCAL)

run-prod: ## Run with production config (config-prod.yaml)
	@echo "🚀 Running auth-service with production config..."
	@go run cmd/server/main.go -c $(CONFIG_PROD)

dev: run-local ## Alias for run-local (local development)

# Run tests
test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -cover -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run go vet
vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@echo "Clean complete"

# Initialize dependencies
deps: ## Install and update dependencies
	@echo "Initializing dependencies..."
	@go mod tidy
	@go mod download
	@echo "Dependencies initialized"

# Initialize MySQL database (default)
init-db: init-mysql ## Initialize database (default: MySQL)

# Initialize PostgreSQL database (legacy)
init-postgres: ## Initialize PostgreSQL database (legacy)
	@echo "Initializing PostgreSQL database..."
	@psql -U postgres -c "CREATE DATABASE k8s_agent_auth;" || true
	@echo "PostgreSQL database initialized"

# Initialize MySQL database (reads config from configs/config.yaml)
init-mysql: ## Initialize MySQL database from config
	@echo "Initializing MySQL database from config.yaml..."
	@bash scripts/init-mysql.sh
	@echo "✅ MySQL database initialized successfully"

# Database management
db-create: ## Create MySQL database
	@echo "Creating MySQL database 'user_auth'..."
	@mysql -h localhost -u root -proot -e "CREATE DATABASE IF NOT EXISTS user_auth CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" || echo "⚠️  Failed to create database. Please check MySQL connection."
	@echo "✅ Database created"

db-drop: ## Drop MySQL database (⚠️  WARNING: This will delete all data!)
	@echo "⚠️  WARNING: This will delete all data in 'user_auth' database!"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@mysql -h localhost -u root -proot -e "DROP DATABASE IF EXISTS user_auth;" || echo "⚠️  Failed to drop database"
	@echo "✅ Database dropped"

db-reset: db-drop db-create init-mysql ## Reset database (drop, create, migrate)
	@echo "✅ Database reset complete"

# Show current configuration
show-config: ## Show current configuration file being used
	@echo "Current configuration: $(CONFIG)"
	@echo ""
	@cat $(CONFIG)

show-config-dev: ## Show development configuration
	@echo "Development configuration: $(CONFIG_DEV)"
	@echo ""
	@cat $(CONFIG_DEV)

show-config-local: ## Show local configuration
	@echo "Local configuration: $(CONFIG_LOCAL)"
	@echo ""
	@cat $(CONFIG_LOCAL)

show-config-prod: ## Show production configuration
	@echo "Production configuration: $(CONFIG_PROD)"
	@echo ""
	@cat $(CONFIG_PROD)

# Format code
fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✅ Format complete"

# Run linter
lint: ## Run golangci-lint
	@echo "Running linter..."
	@golangci-lint run ./...
	@echo "✅ Lint complete"

# Clean build artifacts
clean: ## Clean build artifacts and cache
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean -cache -testcache
	@echo "✅ Clean complete"

# Quick start (for first time setup)
quick-start: deps db-create init-mysql run-local ## Quick start: setup and run locally
	@echo "✅ Quick start complete"

# Development workflow helpers
check: fmt vet lint test ## Run all checks (fmt, vet, lint, test)
	@echo "✅ All checks passed"

watch: ## Watch for changes and restart (requires: go install github.com/cosmtrek/air@latest)
	@echo "👀 Watching for changes..."
	@air -c .air.toml 2>/dev/null || (echo "⚠️  Air not found. Install with: go install github.com/cosmtrek/air@latest" && exit 1)

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

docker-run: ## Run Docker container
	@echo "Running Docker container: $(DOCKER_IMAGE)"
	@docker run --rm -p 8090:8090 $(DOCKER_IMAGE)

docker-tag-env: ## Tag Docker image with environment suffix
	@echo "Tagging Docker image for environment: $(ENV)"
	@docker tag $(DOCKER_IMAGE) $(DOCKER_IMAGE_ENV)
	@echo "Tagged: $(DOCKER_IMAGE_ENV)"

docker-push-env: docker-tag-env ## Push Docker image with environment tag
	@echo "Pushing Docker image: $(DOCKER_IMAGE_ENV)"
	@docker push $(DOCKER_IMAGE_ENV)
	@echo "Pushed: $(DOCKER_IMAGE_ENV)"

# Service health check
health: ## Check if service is running and healthy
	@echo "Checking service health..."
	@curl -s http://localhost:8090/health | jq . || echo "⚠️  Service not responding"

# Log viewing (requires running service)
logs: ## View service logs (for Docker)
	@docker logs -f auth-service 2>/dev/null || echo "⚠️  No running container found"

# Info and documentation
info: ## Show project information and useful commands
	@echo "================================================"
	@echo "  Auth Service - Project Information"
	@echo "================================================"
	@echo ""
	@echo "📦 Service: auth-service"
	@echo "🔧 Port: 8090"
	@echo "📝 Database: MySQL (user_auth)"
	@echo "🔴 Redis: localhost:6379"
	@echo ""
	@echo "📚 Available configurations:"
	@echo "  • config.yaml       - Default/Production template"
	@echo "  • config-dev.yaml   - Remote development"
	@echo "  • config-local.yaml - Local development"
	@echo "  • config-prod.yaml  - Production (env vars)"
	@echo ""
	@echo "🚀 Quick commands:"
	@echo "  make run-local      - Run with local config"
	@echo "  make run-dev        - Run with dev config"
	@echo "  make quick-start    - Setup and run (first time)"
	@echo "  make check          - Run all checks"
	@echo "  make help           - Show all available commands"
	@echo ""
	@echo "📖 Documentation:"
	@echo "  • README.md         - Service documentation"
	@echo "  • CONFIG_GUIDE.md   - Configuration guide"
	@echo "  • QUICKSTART.md     - Quick start guide"
	@echo "================================================"

# Default target when just running 'make'
.DEFAULT_GOAL := help
