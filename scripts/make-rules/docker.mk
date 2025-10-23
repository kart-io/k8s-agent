# Docker build rules
# Based on OneX project structure

##@ Docker

# Docker image tag
IMAGE_TAG ?= $(VERSION)

# Docker build options
DOCKER_BUILD_OPTS ?= --pull
DOCKER_BUILDX_PLATFORM ?= linux/amd64,linux/arm64

.PHONY: docker.build
docker.build: ## Build Docker images for all services
	$(call print_target,$@)
	@for svc in $(SERVICES); do \
		$(call print_info,"Building Docker image for $$svc..."); \
		$(DOCKER) build $(DOCKER_BUILD_OPTS) \
			--build-arg VERSION=$(VERSION) \
			--build-arg GIT_COMMIT=$(GIT_COMMIT) \
			--build-arg BUILD_TIME=$(BUILD_TIME) \
			-t $(IMAGE_PREFIX)/$$svc:$(IMAGE_TAG) \
			-t $(IMAGE_PREFIX)/$$svc:latest \
			-f $$svc/Dockerfile $$svc; \
	done
	$(call print_info,"Docker images built successfully")

.PHONY: docker.build.%
docker.build.%: ## Build Docker image for specific service (e.g., make docker.build.agent-manager)
	$(call print_target,$@)
	$(eval SERVICE := $(subst docker.build.,,$@))
	@if [ -d "$(SERVICE)" ] && [ -f "$(SERVICE)/Dockerfile" ]; then \
		$(call print_info,"Building Docker image for $(SERVICE)..."); \
		$(DOCKER) build $(DOCKER_BUILD_OPTS) \
			--build-arg VERSION=$(VERSION) \
			--build-arg GIT_COMMIT=$(GIT_COMMIT) \
			--build-arg BUILD_TIME=$(BUILD_TIME) \
			-t $(IMAGE_PREFIX)/$(SERVICE):$(IMAGE_TAG) \
			-t $(IMAGE_PREFIX)/$(SERVICE):latest \
			-f $(SERVICE)/Dockerfile $(SERVICE); \
	else \
		$(call print_error,"Service $(SERVICE) or Dockerfile not found"); \
		exit 1; \
	fi

.PHONY: docker.buildx
docker.buildx: docker.buildx.setup ## Build multi-platform Docker images
	$(call print_target,$@)
	@for svc in $(SERVICES); do \
		$(call print_info,"Building multi-platform image for $$svc..."); \
		$(DOCKER) buildx build \
			--platform $(DOCKER_BUILDX_PLATFORM) \
			--build-arg VERSION=$(VERSION) \
			--build-arg GIT_COMMIT=$(GIT_COMMIT) \
			--build-arg BUILD_TIME=$(BUILD_TIME) \
			-t $(IMAGE_PREFIX)/$$svc:$(IMAGE_TAG) \
			-t $(IMAGE_PREFIX)/$$svc:latest \
			-f $$svc/Dockerfile $$svc; \
	done

.PHONY: docker.buildx.%
docker.buildx.%: docker.buildx.setup ## Build multi-platform image for specific service
	$(call print_target,$@)
	$(eval SERVICE := $(subst docker.buildx.,,$@))
	@if [ -d "$(SERVICE)" ] && [ -f "$(SERVICE)/Dockerfile" ]; then \
		$(call print_info,"Building multi-platform image for $(SERVICE)..."); \
		$(DOCKER) buildx build \
			--platform $(DOCKER_BUILDX_PLATFORM) \
			--build-arg VERSION=$(VERSION) \
			--build-arg GIT_COMMIT=$(GIT_COMMIT) \
			--build-arg BUILD_TIME=$(BUILD_TIME) \
			-t $(IMAGE_PREFIX)/$(SERVICE):$(IMAGE_TAG) \
			-t $(IMAGE_PREFIX)/$(SERVICE):latest \
			-f $(SERVICE)/Dockerfile $(SERVICE); \
	else \
		$(call print_error,"Service $(SERVICE) or Dockerfile not found"); \
		exit 1; \
	fi

.PHONY: docker.push
docker.push: ## Push Docker images
	$(call print_target,$@)
	@for svc in $(SERVICES); do \
		$(call print_info,"Pushing $$svc:$(IMAGE_TAG)..."); \
		$(DOCKER) push $(IMAGE_PREFIX)/$$svc:$(IMAGE_TAG); \
		$(call print_info,"Pushing $$svc:latest..."); \
		$(DOCKER) push $(IMAGE_PREFIX)/$$svc:latest; \
	done

.PHONY: docker.push.%
docker.push.%: ## Push specific service image
	$(call print_target,$@)
	$(eval SERVICE := $(subst docker.push.,,$@))
	$(call print_info,"Pushing $(SERVICE):$(IMAGE_TAG)...")
	@$(DOCKER) push $(IMAGE_PREFIX)/$(SERVICE):$(IMAGE_TAG)
	$(call print_info,"Pushing $(SERVICE):latest...")
	@$(DOCKER) push $(IMAGE_PREFIX)/$(SERVICE):latest

.PHONY: docker.buildx.push
docker.buildx.push: docker.buildx.setup ## Build and push multi-platform images
	$(call print_target,$@)
	@for svc in $(SERVICES); do \
		$(call print_info,"Building and pushing multi-platform image for $$svc..."); \
		$(DOCKER) buildx build --push \
			--platform $(DOCKER_BUILDX_PLATFORM) \
			--build-arg VERSION=$(VERSION) \
			--build-arg GIT_COMMIT=$(GIT_COMMIT) \
			--build-arg BUILD_TIME=$(BUILD_TIME) \
			-t $(IMAGE_PREFIX)/$$svc:$(IMAGE_TAG) \
			-t $(IMAGE_PREFIX)/$$svc:latest \
			-f $$svc/Dockerfile $$svc; \
	done

.PHONY: docker.buildx.push.%
docker.buildx.push.%: docker.buildx.setup ## Build and push multi-platform image for specific service
	$(call print_target,$@)
	$(eval SERVICE := $(subst docker.buildx.push.,,$@))
	@if [ -d "$(SERVICE)" ] && [ -f "$(SERVICE)/Dockerfile" ]; then \
		$(call print_info,"Building and pushing multi-platform image for $(SERVICE)..."); \
		$(DOCKER) buildx build --push \
			--platform $(DOCKER_BUILDX_PLATFORM) \
			--build-arg VERSION=$(VERSION) \
			--build-arg GIT_COMMIT=$(GIT_COMMIT) \
			--build-arg BUILD_TIME=$(BUILD_TIME) \
			-t $(IMAGE_PREFIX)/$(SERVICE):$(IMAGE_TAG) \
			-t $(IMAGE_PREFIX)/$(SERVICE):latest \
			-f $(SERVICE)/Dockerfile $(SERVICE); \
	else \
		$(call print_error,"Service $(SERVICE) or Dockerfile not found"); \
		exit 1; \
	fi

.PHONY: docker.buildx.setup
docker.buildx.setup: ## Setup Docker buildx builder
	@if ! $(DOCKER) buildx ls | grep -q aetherius-builder; then \
		$(call print_info,"Creating buildx builder..."); \
		$(DOCKER) buildx create --name aetherius-builder --use --bootstrap; \
	else \
		$(call print_info,"Using existing buildx builder"); \
	fi

.PHONY: docker.clean
docker.clean: ## Remove local Docker images
	$(call print_target,$@)
	@for svc in $(SERVICES); do \
		$(call print_info,"Removing images for $$svc..."); \
		$(DOCKER) rmi -f $(IMAGE_PREFIX)/$$svc:$(IMAGE_TAG) 2>/dev/null || true; \
		$(DOCKER) rmi -f $(IMAGE_PREFIX)/$$svc:latest 2>/dev/null || true; \
	done

.PHONY: docker.prune
docker.prune: ## Prune Docker system
	$(call print_target,$@)
	$(call print_warning,"This will remove all unused containers, networks, images and build cache")
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		$(DOCKER) system prune -af --volumes; \
	fi
