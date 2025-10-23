# Git hooks rules
# Based on OneX project structure

##@ Git Hooks

.PHONY: hooks.install
hooks.install: ## Install git hooks
	$(call print_target,$@)
	@bash githooks/install.sh
	$(call print_info,"Git hooks installed")

.PHONY: hooks.uninstall
hooks.uninstall: ## Uninstall git hooks
	$(call print_target,$@)
	@rm -f .git/hooks/pre-commit
	@rm -f .git/hooks/commit-msg
	$(call print_info,"Git hooks uninstalled")

.PHONY: hooks.run-pre-commit
hooks.run-pre-commit: ## Run pre-commit hook manually
	$(call print_target,$@)
	@bash githooks/pre-commit

.PHONY: hooks.run-commit-msg
hooks.run-commit-msg: ## Run commit-msg hook manually (MSG=<message>)
	$(call print_target,$@)
	@if [ -z "$(MSG)" ]; then \
		$(call print_error,"MSG variable required. Use: make hooks.run-commit-msg MSG='feat: add feature'"); \
		exit 1; \
	fi
	@echo "$(MSG)" > /tmp/commit-msg-test
	@bash githooks/commit-msg /tmp/commit-msg-test
	@rm -f /tmp/commit-msg-test
