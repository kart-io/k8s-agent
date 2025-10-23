# Version management rules
# Based on OneX project structure with Uplift integration

##@ Version Management

VERSION_FILE := $(ROOT_DIR)/VERSION
CHANGELOG_FILE := $(ROOT_DIR)/CHANGELOG.md

.PHONY: version.show
version.show: ## Show current version information
	$(call print_target,$@)
	@echo "Current Version: $$(cat $(VERSION_FILE))"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Git Branch: $(GIT_BRANCH)"
	@echo "Build Time: $(BUILD_TIME)"

.PHONY: version.bump
version.bump: ## Bump version using uplift (TYPE=patch|minor|major)
	$(call print_target,$@)
	@if ! command -v uplift >/dev/null 2>&1; then \
		$(call print_error,"uplift not found. Run 'make tools.install.uplift'"); \
		exit 1; \
	fi
	@if [ -z "$(TYPE)" ]; then \
		$(call print_error,"TYPE not specified. Use: make version.bump TYPE=patch|minor|major"); \
		exit 1; \
	fi
	@$(call print_info,"Bumping $(TYPE) version with uplift...")
	@uplift bump $(TYPE)
	@$(call print_success,"Version bumped to $$(cat $(VERSION_FILE))")

.PHONY: version.bump.patch
version.bump.patch: ## Bump patch version (x.y.Z)
	$(call print_target,$@)
	@$(MAKE) version.bump TYPE=patch

.PHONY: version.bump.minor
version.bump.minor: ## Bump minor version (x.Y.z)
	$(call print_target,$@)
	@$(MAKE) version.bump TYPE=minor

.PHONY: version.bump.major
version.bump.major: ## Bump major version (X.y.z)
	$(call print_target,$@)
	@$(MAKE) version.bump TYPE=major

.PHONY: version.set
version.set: ## Set specific version (VER=x.y.z)
	$(call print_target,$@)
	@if [ -z "$(VER)" ]; then \
		$(call print_error,"VER not specified. Use: make version.set VER=1.2.3"); \
		exit 1; \
	fi
	@$(call print_info,"Setting version to $(VER)...")
	@echo "$(VER)" > $(VERSION_FILE)
	@$(call print_success,"Version set to $(VER)")

.PHONY: version.tag
version.tag: ## Create git tag for current version
	$(call print_target,$@)
	@VERSION=$$(cat $(VERSION_FILE)); \
	$(call print_info,"Creating git tag v$$VERSION..."); \
	git tag -a "v$$VERSION" -m "Release v$$VERSION"; \
	$(call print_success,"Git tag v$$VERSION created")

.PHONY: version.tag.push
version.tag.push: ## Push git tags to remote
	$(call print_target,$@)
	@$(call print_info,"Pushing git tags...")
	@git push --tags
	@$(call print_success,"Git tags pushed")

.PHONY: version.changelog
version.changelog: ## Generate changelog using uplift
	$(call print_target,$@)
	@if ! command -v uplift >/dev/null 2>&1; then \
		$(call print_error,"uplift not found. Run 'make tools.install.uplift'"); \
		exit 1; \
	fi
	@$(call print_info,"Generating changelog with uplift...")
	@uplift changelog
	@$(call print_success,"Changelog updated")

.PHONY: version.release
version.release: ## Create a new release (TYPE=patch|minor|major)
	$(call print_target,$@)
	@if [ -z "$(TYPE)" ]; then \
		$(call print_error,"TYPE not specified. Use: make version.release TYPE=patch|minor|major"); \
		exit 1; \
	fi
	@$(call print_info,"Creating $(TYPE) release...")
	@$(MAKE) go.test
	@$(MAKE) go.lint
	@$(MAKE) version.bump TYPE=$(TYPE)
	@$(MAKE) version.changelog
	@VERSION=$$(cat $(VERSION_FILE)); \
	$(call print_info,"Committing version $$VERSION..."); \
	git add $(VERSION_FILE) $(CHANGELOG_FILE); \
	git commit -m "chore(release): bump version to $$VERSION"; \
	git tag -a "v$$VERSION" -m "Release v$$VERSION"; \
	$(call print_success,"Release v$$VERSION created")
	@$(call print_info,"Push changes with: git push && git push --tags")

.PHONY: version.validate
version.validate: ## Validate version format
	$(call print_target,$@)
	@VERSION=$$(cat $(VERSION_FILE)); \
	if echo "$$VERSION" | grep -Eq '^v?[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$$'; then \
		$(call print_success,"Version $$VERSION is valid"); \
	else \
		$(call print_error,"Version $$VERSION is invalid. Must follow semver format"); \
		exit 1; \
	fi

.PHONY: version.info
version.info: ## Show detailed version information
	$(call print_target,$@)
	@echo "═══════════════════════════════════════"
	@echo "  Version Information"
	@echo "═══════════════════════════════════════"
	@echo "Version:        $$(cat $(VERSION_FILE))"
	@echo "Git Commit:     $(GIT_COMMIT)"
	@echo "Git Branch:     $(GIT_BRANCH)"
	@echo "Git Tag:        $$(git describe --tags --exact-match 2>/dev/null || echo 'none')"
	@echo "Build Time:     $(BUILD_TIME)"
	@echo "Go Version:     $(GO_VERSION)"
	@echo "Platform:       $(PLATFORM)"
	@echo "Project:        $(PROJECT_NAME)"
	@echo "Registry:       $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)"
	@echo "═══════════════════════════════════════"
