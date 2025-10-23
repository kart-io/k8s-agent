# ==============================================================================
# Makefile helper functions for docker image management
# Advanced image building with multi-platform support
#

DOCKER := docker
DOCKER_SUPPORTED_API_VERSION ?= 1.40

EXTRA_ARGS ?=
_DOCKER_BUILD_EXTRA_ARGS :=

# Track code version with Docker Label
DOCKER_LABELS ?= git-describe="$(VERSION)" git-commit="$(GIT_COMMIT)" build-date="$(BUILD_TIME)"

ifdef HTTP_PROXY
_DOCKER_BUILD_EXTRA_ARGS += --build-arg HTTP_PROXY=${HTTP_PROXY}
endif

ifdef HTTPS_PROXY
_DOCKER_BUILD_EXTRA_ARGS += --build-arg HTTPS_PROXY=${HTTPS_PROXY}
endif

ifneq ($(EXTRA_ARGS), )
_DOCKER_BUILD_EXTRA_ARGS += $(EXTRA_ARGS)
endif

# Image names (use SERVICES from root Makefile)
IMAGES ?= $(SERVICES)
ifeq ($(IMAGES),)
  $(error Could not determine IMAGES, set SERVICES or run in source dir)
endif

##@ Image Management

.PHONY: image.verify
image.verify: ## Verify docker version
	$(call print_target,$@)
	@$(eval API_VERSION := $(shell $(DOCKER) version --format '{{.Client.APIVersion}}' 2>/dev/null || echo "0"))
	@if [ "$(API_VERSION)" = "0" ]; then \
		$(call print_error,"Docker not found or not running"); \
		exit 1; \
	fi
	@$(call print_success,"Docker API version: $(API_VERSION)")

.PHONY: image.daemon.verify
image.daemon.verify: ## Verify docker daemon experimental features
	$(call print_target,$@)
	@$(eval EXPERIMENTAL := $(shell $(DOCKER) version --format '{{.Server.Experimental}}' 2>/dev/null || echo "false"))
	@if [ "$(EXPERIMENTAL)" != "true" ]; then \
		$(call print_warning,"Experimental features not enabled (optional for multi-arch builds)"); \
	else \
		$(call print_success,"Docker experimental features enabled"); \
	fi

.PHONY: image.build
image.build: image.verify $(addprefix image.build., $(IMAGES)) ## Build all docker images

.PHONY: image.build.%
image.build.%: ## Build specific docker image (e.g., make image.build.agent-manager)
	$(call print_target,$@)
	@$(eval IMAGE := $*)
	@$(eval IMAGE_TAG := $(VERSION))
	@$(eval DOCKERFILE := $(if $(wildcard $(IMAGE)/Dockerfile),$(IMAGE)/Dockerfile,Dockerfile))
	@echo "$(CYAN)Building docker image $(IMAGE):$(IMAGE_TAG)$(NC)"
	@$(DOCKER) build \
		$(_DOCKER_BUILD_EXTRA_ARGS) \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--label $(DOCKER_LABELS) \
		-t $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE):$(IMAGE_TAG) \
		-t $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE):latest \
		-f $(DOCKERFILE) \
		.
	@echo "$(GREEN)✓ Built image: $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE):$(IMAGE_TAG)$(NC)"

.PHONY: image.build.multiarch
image.build.multiarch: image.verify $(addprefix image.build.multiarch., $(IMAGES)) ## Build all images with multi-arch support

.PHONY: image.build.multiarch.%
image.build.multiarch.%: ## Build specific image with multi-arch (e.g., make image.build.multiarch.agent-manager)
	$(call print_target,$@)
	@$(eval IMAGE := $*)
	@$(eval IMAGE_TAG := $(VERSION))
	@$(eval DOCKERFILE := $(if $(wildcard $(IMAGE)/Dockerfile),$(IMAGE)/Dockerfile,Dockerfile))
	@echo "$(CYAN)Building multi-arch docker image $(IMAGE):$(IMAGE_TAG)$(NC)"
	@$(DOCKER) buildx build \
		$(_DOCKER_BUILD_EXTRA_ARGS) \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--label $(DOCKER_LABELS) \
		-t $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE):$(IMAGE_TAG) \
		-t $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE):latest \
		-f $(DOCKERFILE) \
		. \
		--load
	@echo "$(GREEN)✓ Built multi-arch image: $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE):$(IMAGE_TAG)$(NC)"

.PHONY: image.push
image.push: image.verify $(addprefix image.push., $(IMAGES)) ## Push all docker images to registry

.PHONY: image.push.%
image.push.%: image.build.% ## Push specific docker image (e.g., make image.push.agent-manager)
	$(call print_target,$@)
	@$(eval IMAGE := $*)
	@$(eval IMAGE_TAG := $(VERSION))
	@echo "$(CYAN)Pushing image $(IMAGE):$(IMAGE_TAG) to $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)$(NC)"
	@$(DOCKER) push $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE):$(IMAGE_TAG)
	@$(DOCKER) push $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE):latest
	@echo "$(GREEN)✓ Pushed image: $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE):$(IMAGE_TAG)$(NC)"

.PHONY: image.push.multiarch
image.push.multiarch: image.verify $(addprefix image.push.multiarch., $(IMAGES)) ## Push all multi-arch images

.PHONY: image.push.multiarch.%
image.push.multiarch.%: ## Push specific multi-arch image (e.g., make image.push.multiarch.agent-manager)
	$(call print_target,$@)
	@$(eval IMAGE := $*)
	@$(eval IMAGE_TAG := $(VERSION))
	@$(eval DOCKERFILE := $(if $(wildcard $(IMAGE)/Dockerfile),$(IMAGE)/Dockerfile,Dockerfile))
	@echo "$(CYAN)Building and pushing multi-arch image $(IMAGE):$(IMAGE_TAG)$(NC)"
	@$(DOCKER) buildx build \
		$(_DOCKER_BUILD_EXTRA_ARGS) \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--label $(DOCKER_LABELS) \
		-t $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE):$(IMAGE_TAG) \
		-t $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE):latest \
		-f $(DOCKERFILE) \
		. \
		--push
	@echo "$(GREEN)✓ Pushed multi-arch image: $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE):$(IMAGE_TAG)$(NC)"

.PHONY: image.list
image.list: ## List all built images
	$(call print_target,$@)
	@echo "$(CYAN)Docker images for $(DOCKER_NAMESPACE):$(NC)"
	@$(DOCKER) images | grep $(DOCKER_NAMESPACE) || echo "No images found"

.PHONY: image.clean
image.clean: ## Clean all project docker images
	$(call print_target,$@)
	@echo "$(CYAN)Cleaning docker images for $(DOCKER_NAMESPACE)...$(NC)"
	@for image in $(IMAGES); do \
		$(DOCKER) rmi -f $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$$image:$(VERSION) 2>/dev/null || true; \
		$(DOCKER) rmi -f $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$$image:latest 2>/dev/null || true; \
	done
	@echo "$(GREEN)✓ Cleaned docker images$(NC)"

.PHONY: image.prune
image.prune: ## Prune unused docker images
	$(call print_target,$@)
	@echo "$(CYAN)Pruning unused docker images...$(NC)"
	@$(DOCKER) image prune -f
	@echo "$(GREEN)✓ Pruned unused images$(NC)"
