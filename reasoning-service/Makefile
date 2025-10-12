.PHONY: help install run dev test clean docker-build docker-run

# Docker image configuration
DOCKER_REGISTRY ?= aetherius
IMAGE_NAME ?= reasoning-service
IMAGE_TAG ?= latest
ENV ?= dev
PLATFORMS ?= linux/amd64,linux/arm64
DOCKER_IMAGE = $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
DOCKER_IMAGE_ENV = $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)-$(ENV)

help:
	@echo "Aetherius Reasoning Service - Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  install       - Install Python dependencies"
	@echo "  run           - Run the service"
	@echo "  dev           - Run in development mode with auto-reload"
	@echo "  test          - Run tests"
	@echo "  lint          - Run linting"
	@echo "  format        - Format code"
	@echo "  clean         - Clean temporary files"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run in Docker container"

install:
	pip install -r requirements.txt

run:
	python cmd/server/main.py --config=configs/config.yaml

dev:
	uvicorn internal.api.server:app --reload --host 0.0.0.0 --port 8082

test:
	pytest tests/ -v --cov=internal --cov-report=html

lint:
	pylint internal/ pkg/ cmd/

format:
	black internal/ pkg/ cmd/
	isort internal/ pkg/ cmd/

clean:
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete
	find . -type d -name "*.egg-info" -exec rm -rf {} + 2>/dev/null || true
	rm -rf .pytest_cache htmlcov .coverage

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-buildx-setup: ## Setup Docker buildx builder for multi-platform builds
	@echo "Setting up Docker buildx builder..."
	@docker buildx create --name multiarch-builder --driver docker-container --use 2>/dev/null || \
		docker buildx use multiarch-builder 2>/dev/null || \
		(docker buildx create --name multiarch-builder --driver docker-container --use && docker buildx inspect --bootstrap)
	@echo "Docker buildx builder ready"

docker-run:
	docker run -d \
		--name reasoning-service \
		-p 8082:8082 \
		-e NEO4J_URI=bolt://neo4j:7687 \
		-e NEO4J_USER=neo4j \
		-e NEO4J_PASSWORD=password \
		$(DOCKER_IMAGE)

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