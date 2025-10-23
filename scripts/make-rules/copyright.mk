# ==============================================================================
# Makefile helper functions for copyright
#

.PHONY: copyright.verify
copyright.verify: ## Verify the boilerplate headers for all files
	$(call print_target,$@)
	@if command -v addlicense >/dev/null 2>&1; then \
		addlicense --check -f $(SCRIPTS_DIR)/boilerplate.txt $(ROOT_DIR) --skip-dirs=third_party,vendor,_output,.git,.github,node_modules; \
	else \
		$(call print_warning,"addlicense not installed. Run 'make tools.install.addlicense'"); \
	fi

.PHONY: copyright.add
copyright.add: ## Add boilerplate headers for all missing files
	$(call print_target,$@)
	@if command -v addlicense >/dev/null 2>&1; then \
		addlicense -v -f $(SCRIPTS_DIR)/boilerplate.txt $(ROOT_DIR) --skip-dirs=third_party,vendor,_output,.git,.github,node_modules; \
	else \
		$(call print_error,"addlicense not installed. Run 'make tools.install.addlicense'"); \
		exit 1; \
	fi

.PHONY: copyright.clean
copyright.clean: ## Remove boilerplate headers from all files
	$(call print_target,$@)
	@find $(ROOT_DIR) -type f \( -name "*.go" -o -name "*.sh" \) \
		-not -path "*/vendor/*" \
		-not -path "*/_output/*" \
		-not -path "*/third_party/*" \
		-exec sed -i '/^\/\/ Copyright.*kart\.io\./d; /^\/\/ Use of this source code/d; /^\/\/ license that can be found/d' {} \;

##@ Copyright
