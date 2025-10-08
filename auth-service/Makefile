.PHONY: build run test clean init-db init-mysql

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
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@echo "Clean complete"

# Initialize dependencies
deps:
	@echo "Initializing dependencies..."
	@go mod tidy
	@go mod download
	@echo "Dependencies initialized"

# Initialize PostgreSQL database
init-db:
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
	@echo "Building Docker image..."
	@docker build -t aetherius/auth-service:latest .

docker-buildx: ## Build multi-platform Docker image (linux/amd64,linux/arm64)
	@echo "Building multi-platform Docker image..."
	@if command -v docker-buildx >/dev/null 2>&1 || docker buildx version >/dev/null 2>&1; then \
		docker buildx build --platform linux/amd64,linux/arm64 \
			-t aetherius/auth-service:latest --load . 2>/dev/null || \
		docker buildx build --platform linux/amd64,linux/arm64 \
			-t aetherius/auth-service:latest . ; \
	else \
		echo "Docker buildx not available, building for current platform only..."; \
		docker build -t aetherius/auth-service:latest . ; \
	fi

docker-buildx-push: ## Build and push multi-platform Docker image
	@echo "Building and pushing multi-platform Docker image..."
	@if command -v docker-buildx >/dev/null 2>&1 || docker buildx version >/dev/null 2>&1; then \
		docker buildx build --platform linux/amd64,linux/arm64 \
			-t aetherius/auth-service:latest --push . ; \
	else \
		echo "Docker buildx not available, building and pushing for current platform only..."; \
		docker build -t aetherius/auth-service:latest . && \
		docker push aetherius/auth-service:latest; \
	fi

docker-run: ## Run Docker container
	@docker run --rm -p 8090:8090 aetherius/auth-service:latest
