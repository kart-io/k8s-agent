.PHONY: build run test clean deps

# Build the gateway service
build:
	@echo "Building gateway-service..."
	@mkdir -p bin
	@go build -o bin/gateway-service cmd/server/main.go
	@echo "Build complete: bin/gateway-service"

# Run the service
run:
	@echo "Running gateway-service..."
	@go run cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf logs/
	@echo "Clean complete"

# Initialize dependencies
deps:
	@echo "Initializing dependencies..."
	@go mod tidy
	@go mod download
	@echo "Dependencies initialized"

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

# Create logs directory
init-logs:
	@mkdir -p logs
	@echo "Logs directory created"

# Run with hot reload (requires air)
dev:
	@air

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t aetherius/gateway-service:latest .

docker-buildx: ## Build multi-platform Docker image (linux/amd64,linux/arm64)
	@echo "Building multi-platform Docker image..."
	@if command -v docker-buildx >/dev/null 2>&1 || docker buildx version >/dev/null 2>&1; then \
		docker buildx build --platform linux/amd64,linux/arm64 \
			-t aetherius/gateway-service:latest --load . 2>/dev/null || \
		docker buildx build --platform linux/amd64,linux/arm64 \
			-t aetherius/gateway-service:latest . ; \
	else \
		echo "Docker buildx not available, building for current platform only..."; \
		docker build -t aetherius/gateway-service:latest . ; \
	fi

docker-buildx-push: ## Build and push multi-platform Docker image
	@echo "Building and pushing multi-platform Docker image..."
	@if command -v docker-buildx >/dev/null 2>&1 || docker buildx version >/dev/null 2>&1; then \
		docker buildx build --platform linux/amd64,linux/arm64 \
			-t aetherius/gateway-service:latest --push . ; \
	else \
		echo "Docker buildx not available, building and pushing for current platform only..."; \
		docker build -t aetherius/gateway-service:latest . && \
		docker push aetherius/gateway-service:latest; \
	fi

docker-run: ## Run Docker container
	@docker run --rm -p 8080:8080 aetherius/gateway-service:latest
