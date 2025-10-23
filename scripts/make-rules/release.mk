# ==============================================================================
# Makefile helper functions for release management
# Uses uplift for semantic versioning and changelog generation
#

##@ Release Management

.PHONY: release.run
release.run: tools.verify.uplift release.check ## Generate CHANGELOG, commit, and tag with uplift
	$(call print_target,$@)
	@echo "$(CYAN)Running uplift release...$(NC)"
	@uplift release --fetch-all
	@echo "$(GREEN)✓ Release created successfully$(NC)"

.PHONY: release.check
release.check: ## Check if git repository is clean before release
	$(call print_target,$@)
	@if ! git diff --quiet; then \
		$(call print_error,"Git repository is not clean. Please commit or stash changes."); \
		echo "$(YELLOW)Modified files:$(NC)"; \
		git status --short; \
		exit 1; \
	fi
	@echo "$(GREEN)✓ Git repository is clean$(NC)"

.PHONY: release.tag
release.tag: ## Create a new git tag (VERSION=vX.Y.Z required)
	$(call print_target,$@)
	@if [ -z "$(VERSION)" ]; then \
		$(call print_error,"VERSION not specified. Usage: make release.tag VERSION=v1.0.0"); \
		exit 1; \
	fi
	@echo "$(CYAN)Creating tag $(VERSION)...$(NC)"
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "$(GREEN)✓ Tag $(VERSION) created$(NC)"
	@echo "$(YELLOW)To push tag: git push origin $(VERSION)$(NC)"

.PHONY: release.push
release.push: ## Push current tag to origin (VERSION=vX.Y.Z required)
	$(call print_target,$@)
	@if [ -z "$(VERSION)" ]; then \
		$(call print_error,"VERSION not specified. Usage: make release.push VERSION=v1.0.0"); \
		exit 1; \
	fi
	@echo "$(CYAN)Pushing tag $(VERSION) to origin...$(NC)"
	@git push origin $(VERSION)
	@echo "$(GREEN)✓ Tag $(VERSION) pushed to origin$(NC)"

.PHONY: release.delete
release.delete: ## Delete a git tag locally and remotely (VERSION=vX.Y.Z required)
	$(call print_target,$@)
	@if [ -z "$(VERSION)" ]; then \
		$(call print_error,"VERSION not specified. Usage: make release.delete VERSION=v1.0.0"); \
		exit 1; \
	fi
	@echo "$(CYAN)Deleting tag $(VERSION)...$(NC)"
	@git tag -d $(VERSION) || true
	@git push origin :refs/tags/$(VERSION) || true
	@echo "$(GREEN)✓ Tag $(VERSION) deleted$(NC)"

.PHONY: release.changelog
release.changelog: tools.verify.git-chglog ## Generate CHANGELOG.md
	$(call print_target,$@)
	@echo "$(CYAN)Generating CHANGELOG.md...$(NC)"
	@if [ ! -d .chglog ]; then \
		$(call print_error,".chglog directory not found. Run 'git-chglog --init' first."); \
		exit 1; \
	fi
	@git-chglog -o CHANGELOG.md
	@echo "$(GREEN)✓ CHANGELOG.md generated$(NC)"

.PHONY: release.changelog.tag
release.changelog.tag: tools.verify.git-chglog ## Generate changelog for specific tag (VERSION=vX.Y.Z)
	$(call print_target,$@)
	@if [ -z "$(VERSION)" ]; then \
		$(call print_error,"VERSION not specified. Usage: make release.changelog.tag VERSION=v1.0.0"); \
		exit 1; \
	fi
	@echo "$(CYAN)Generating changelog for $(VERSION)...$(NC)"
	@git-chglog $(VERSION)

.PHONY: release.notes
release.notes: ## Generate release notes for current version
	$(call print_target,$@)
	@echo "$(CYAN)Generating release notes...$(NC)"
	@echo "# Release $(VERSION)"
	@echo ""
	@echo "## Changes"
	@git log --oneline --no-merges $(shell git describe --tags --abbrev=0 2>/dev/null)..HEAD
	@echo ""
	@echo "## Docker Images"
	@for svc in $(SERVICES); do \
		echo "- \`$(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$$svc:$(VERSION)\`"; \
	done

.PHONY: release.verify
release.verify: ## Verify release artifacts
	$(call print_target,$@)
	@echo "$(CYAN)Verifying release artifacts...$(NC)"
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Services: $(SERVICES)"
	@echo ""
	@if [ -f CHANGELOG.md ]; then \
		echo "$(GREEN)✓ CHANGELOG.md exists$(NC)"; \
	else \
		echo "$(RED)✗ CHANGELOG.md missing$(NC)"; \
	fi
	@if git tag | grep -q "^$(VERSION)$$"; then \
		echo "$(GREEN)✓ Tag $(VERSION) exists$(NC)"; \
	else \
		echo "$(YELLOW)⚠ Tag $(VERSION) not found$(NC)"; \
	fi

.PHONY: release.build
release.build: ## Build release artifacts (clean, test, build all)
	$(call print_target,$@)
	@echo "$(CYAN)Building release artifacts...$(NC)"
	@$(MAKE) clean
	@$(MAKE) go.mod.tidy
	@$(MAKE) go.test
	@$(MAKE) go.build SERVICES="$(SERVICES)"
	@echo "$(GREEN)✓ Release artifacts built$(NC)"

.PHONY: release.docker
release.docker: ## Build and push all docker images for release
	$(call print_target,$@)
	@echo "$(CYAN)Building and pushing release docker images...$(NC)"
	@$(MAKE) docker.build SERVICES="$(SERVICES)"
	@$(MAKE) docker.push SERVICES="$(SERVICES)"
	@echo "$(GREEN)✓ Release docker images pushed$(NC)"

.PHONY: release.docker.multiarch
release.docker.multiarch: ## Build and push multi-arch docker images for release
	$(call print_target,$@)
	@echo "$(CYAN)Building and pushing multi-arch release images...$(NC)"
	@$(MAKE) image.push.multiarch
	@echo "$(GREEN)✓ Multi-arch release images pushed$(NC)"

.PHONY: release.all
release.all: release.check release.build release.docker release.changelog ## Complete release workflow
	$(call print_target,$@)
	@echo "$(GREEN)✓ Release workflow completed$(NC)"
	@echo "$(YELLOW)Next steps:$(NC)"
	@echo "  1. Review CHANGELOG.md"
	@echo "  2. Commit changes: git add CHANGELOG.md && git commit -m 'docs: update CHANGELOG for $(VERSION)'"
	@echo "  3. Create tag: make release.tag VERSION=$(VERSION)"
	@echo "  4. Push tag: make release.push VERSION=$(VERSION)"
