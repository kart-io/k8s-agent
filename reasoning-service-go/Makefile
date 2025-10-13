.PHONY: all build run clean test docker-build docker-run

# Build configuration
BINARY_NAME=reasoning-service
BUILD_DIR=bin
MAIN_PATH=cmd/server/main.go

# Docker image configuration
DOCKER_REGISTRY ?= aetherius
IMAGE_NAME ?= reasoning-service-go
IMAGE_TAG ?= latest
ENV ?= dev
PLATFORMS ?= linux/amd64,linux/arm64
DOCKER_IMAGE = $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
DOCKER_IMAGE_ENV = $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)-$(ENV)

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
	go run $(MAIN_PATH)

# Run the service with dev config
run-dev:
	@echo "Running $(BINARY_NAME) with dev config..."
	go run $(MAIN_PATH) -c configs/config-dev.yaml

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
	@echo "Building Docker image: $(DOCKER_IMAGE)"
	docker build -t $(DOCKER_IMAGE) .

docker-buildx-setup: ## Setup Docker buildx builder for multi-platform builds
	@echo "Setting up Docker buildx builder..."
	@docker buildx create --name multiarch-builder --driver docker-container --use 2>/dev/null || \
		docker buildx use multiarch-builder 2>/dev/null || \
		(docker buildx create --name multiarch-builder --driver docker-container --use && docker buildx inspect --bootstrap)
	@echo "Docker buildx builder ready"

# Run Docker container
docker-run:
	@echo "Running Docker container: $(DOCKER_IMAGE)"
	docker run -d \
		-p 8082:8082 \
		-e OPENAI_API_KEY="${OPENAI_API_KEY}" \
		-e GEMINI_API_KEY="${GEMINI_API_KEY}" \
		-e DEEPSEEK_API_KEY="${DEEPSEEK_API_KEY}" \
		-v $(PWD)/configs:/app/configs \
		--name $(BINARY_NAME) \
		$(DOCKER_IMAGE)

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

docker-tag-env: ## Tag Docker image with environment suffix
	@echo "Tagging Docker image for environment: $(ENV)"
	@docker tag $(DOCKER_IMAGE) $(DOCKER_IMAGE_ENV)
	@echo "Tagged: $(DOCKER_IMAGE_ENV)"

docker-push-env: docker-tag-env ## Push Docker image with environment tag
	@echo "Pushing Docker image: $(DOCKER_IMAGE_ENV)"
	@docker push $(DOCKER_IMAGE_ENV)
	@echo "Pushed: $(DOCKER_IMAGE_ENV)"

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

help:
	@echo "Available targets:"
	@echo "  make build                  - Build the binary"
	@echo "  make run                    - Run the service (default config)"
	@echo "  make run-dev                - Run the service (dev config)"
	@echo "  make clean                  - Clean build artifacts"
	@echo "  make test                   - Run tests"
	@echo "  make deps                   - Install dependencies"
	@echo "  make docker-build           - Build Docker image"
	@echo "  make docker-buildx          - Build multi-platform Docker image"
	@echo "  make docker-buildx-push     - Build and push multi-platform image"
	@echo "  make docker-buildx-env      - Build multi-platform image with env tag"
	@echo "  make docker-buildx-push-env - Build and push multi-platform image with env tag"
	@echo "  make docker-run             - Run Docker container"
	@echo "  make docker-stop            - Stop Docker container"
	@echo "  make docker-tag-env         - Tag image with environment"
	@echo "  make docker-push-env        - Push image with environment tag"
	@echo "  make dev                    - Run with auto-reload"
