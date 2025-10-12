.PHONY: help build run test clean init-db init-mysql deps fmt lint docker-build docker-run test-coverage vet dev

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
build:
	@echo "Building auth-service..."
	@go build -o bin/auth-service cmd/server/main.go
	@echo "Build complete: bin/auth-service"

# Run the service
run:
	@echo "Running auth-service..."
	@go run cmd/server/main.go

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
init-db: init-mysql

# Initialize PostgreSQL database (legacy)
init-postgres:
	@echo "Initializing PostgreSQL database..."
	@psql -U postgres -c "CREATE DATABASE k8s_agent_auth;" || true
	@echo "PostgreSQL database initialized"

# Initialize MySQL database
init-mysql:
	@echo "Initializing MySQL database..."
	@mysql -u root -p < scripts/init-mysql.sql
	@echo "MySQL database initialized"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete"

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./...
	@echo "Lint complete"

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
