.PHONY: all build run clean test docker-build docker-run

# Build configuration
BINARY_NAME=reasoning-service
BUILD_DIR=bin
MAIN_PATH=cmd/server/main.go

# Docker configuration
DOCKER_IMAGE=reasoning-service-go
DOCKER_TAG=latest

all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the service
run:
	@echo "Running $(BINARY_NAME)..."
	go run $(MAIN_PATH) -config configs/config.yaml

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	go clean

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -d \
		-p 8082:8082 \
		-e OPENAI_API_KEY="${OPENAI_API_KEY}" \
		-e GEMINI_API_KEY="${GEMINI_API_KEY}" \
		-e DEEPSEEK_API_KEY="${DEEPSEEK_API_KEY}" \
		-v $(PWD)/configs:/app/configs \
		--name $(BINARY_NAME) \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Stop Docker container
docker-stop:
	docker stop $(BINARY_NAME)
	docker rm $(BINARY_NAME)

# View logs
logs:
	docker logs -f $(BINARY_NAME)

# Development with auto-reload (requires air)
dev:
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

help:
	@echo "Available targets:"
	@echo "  make build        - Build the binary"
	@echo "  make run          - Run the service"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make test         - Run tests"
	@echo "  make deps         - Install dependencies"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker container"
	@echo "  make docker-stop  - Stop Docker container"
	@echo "  make dev          - Run with auto-reload"
