.PHONY: build run clean test docker-build version help

SERVICE_NAME=cluster-service
BINARY_NAME=bin/$(SERVICE_NAME)

# Version injection variables
VERSION_PKG = github.com/kart-io/version
GIT_VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH := $(shell git branch --show-current 2>/dev/null || echo "unknown")
GIT_TREE_STATE := $(shell test -n "`git status --porcelain 2>/dev/null`" && echo "dirty" || echo "clean")
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# Build flags with version injection
LDFLAGS := -X '$(VERSION_PKG).serviceName=$(SERVICE_NAME)' \
           -X '$(VERSION_PKG).gitVersion=$(GIT_VERSION)' \
           -X '$(VERSION_PKG).gitCommit=$(GIT_COMMIT)' \
           -X '$(VERSION_PKG).gitBranch=$(GIT_BRANCH)' \
           -X '$(VERSION_PKG).gitTreeState=$(GIT_TREE_STATE)' \
           -X '$(VERSION_PKG).buildDate=$(BUILD_DATE)'

# Docker image configuration
DOCKER_REGISTRY ?= aetherius
IMAGE_NAME ?= $(SERVICE_NAME)
IMAGE_TAG ?= latest
ENV ?= dev
PLATFORMS ?= linux/amd64,linux/arm64
DOCKER_IMAGE = $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
DOCKER_IMAGE_ENV = $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)-$(ENV)

build: ## Build the service with version injection
	@echo "Building $(SERVICE_NAME) with version injection..."
	@echo "  Version: $(GIT_VERSION)"
	@echo "  Commit: $(GIT_COMMIT)"
	@echo "  Branch: $(GIT_BRANCH)"
	@echo "  Build Date: $(BUILD_DATE)"
	@echo "  Tree State: $(GIT_TREE_STATE)"
	@mkdir -p bin
	@go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) cmd/server/main.go
	@echo "Build complete: $(BINARY_NAME)"

version: ## Display version information that will be injected
	@echo "Service Name: $(SERVICE_NAME)"
	@echo "Git Version: $(GIT_VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Git Branch: $(GIT_BRANCH)"
	@echo "Git Tree State: $(GIT_TREE_STATE)"
	@echo "Build Date: $(BUILD_DATE)"

run:
	@echo "Running $(SERVICE_NAME)..."
	@go run cmd/server/main.go

run-dev:
	@echo "Running $(SERVICE_NAME) with dev config..."
	@go run cmd/server/main.go -config configs/config.dev.yaml

run-local: ## Run the service with local MySQL config
	@echo "Running $(SERVICE_NAME) with local config..."
	@go run cmd/server/main.go -config configs/config-local.yaml

clean:
	@echo "Cleaning..."
	@rm -rf bin

test:
	@echo "Running tests..."
	@go test -v ./...

help: ## Display this help message
	@echo "Available commands:"
	@echo "  make build              - Build the service with version injection"
	@echo "  make run                - Run the service directly"
	@echo "  make run-dev            - Run the service with dev config"
	@echo "  make run-local          - Run the service with local MySQL config"
	@echo "  make version            - Display version information"
	@echo "  make test               - Run tests"
	@echo "  make clean              - Clean build artifacts"
	@echo "  make docker-build       - Build Docker image"
	@echo "  make docker-buildx      - Build multi-platform Docker image"
	@echo "  make docker-push-env    - Build and push Docker image with environment tag"
	@echo "  make help               - Display this help message"

docker-build: ## Build Docker image with version injection
	@echo "Building Docker image with version injection..."
	@docker build \
		--build-arg SERVICE_NAME=$(SERVICE_NAME) \
		--build-arg VERSION=$(GIT_VERSION) \
		--build-arg COMMIT=$(GIT_COMMIT) \
		--build-arg BRANCH=$(GIT_BRANCH) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(DOCKER_IMAGE) .

docker-buildx-setup: ## Setup Docker buildx builder for multi-platform builds
	@echo "Setting up Docker buildx builder..."
	@docker buildx create --name multiarch-builder --driver docker-container --use 2>/dev/null || \
		docker buildx use multiarch-builder 2>/dev/null || \
		(docker buildx create --name multiarch-builder --driver docker-container --use && docker buildx inspect --bootstrap)
	@echo "Docker buildx builder ready"

docker-tag-env: ## Tag Docker image with environment suffix
	@echo "Tagging Docker image for environment: $(ENV)"
	@docker tag $(DOCKER_IMAGE) $(DOCKER_IMAGE_ENV)
	@echo "Tagged: $(DOCKER_IMAGE_ENV)"

docker-push-env: docker-tag-env ## Push Docker image with environment tag
	@echo "Pushing Docker image: $(DOCKER_IMAGE_ENV)"
	@docker push $(DOCKER_IMAGE_ENV)
	@echo "Pushed: $(DOCKER_IMAGE_ENV)"

docker-buildx: docker-buildx-setup ## Build multi-platform Docker image with version injection
	@echo "Building multi-platform Docker image: $(DOCKER_IMAGE) for $(PLATFORMS)"
	@docker buildx build --builder multiarch-builder --platform $(PLATFORMS) \
		--build-arg SERVICE_NAME=$(SERVICE_NAME) \
		--build-arg VERSION=$(GIT_VERSION) \
		--build-arg COMMIT=$(GIT_COMMIT) \
		--build-arg BRANCH=$(GIT_BRANCH) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(DOCKER_IMAGE) --load . 2>/dev/null || \
	docker buildx build --builder multiarch-builder --platform $(PLATFORMS) \
		--build-arg SERVICE_NAME=$(SERVICE_NAME) \
		--build-arg VERSION=$(GIT_VERSION) \
		--build-arg COMMIT=$(GIT_COMMIT) \
		--build-arg BRANCH=$(GIT_BRANCH) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(DOCKER_IMAGE) .

docker-buildx-push: docker-buildx-setup ## Build and push multi-platform Docker image with version injection
	@echo "Building and pushing multi-platform Docker image: $(DOCKER_IMAGE) for $(PLATFORMS)"
	@docker buildx build --builder multiarch-builder --platform $(PLATFORMS) \
		--build-arg SERVICE_NAME=$(SERVICE_NAME) \
		--build-arg VERSION=$(GIT_VERSION) \
		--build-arg COMMIT=$(GIT_COMMIT) \
		--build-arg BRANCH=$(GIT_BRANCH) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
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

.DEFAULT_GOAL := build
