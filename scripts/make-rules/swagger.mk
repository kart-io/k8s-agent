# ==============================================================================
# Makefile helper functions for Swagger API documentation
# Generate and serve Swagger/OpenAPI documentation
#

SWAGGER_DIR := $(ROOT_DIR)/api/swagger
OPENAPI_DIR := $(ROOT_DIR)/api/openapi

##@ Swagger Documentation

.PHONY: swagger.run
swagger.run: tools.verify.swagger swagger.generate ## Generate and validate Swagger documentation
	$(call print_target,$@)

.PHONY: swagger.generate
swagger.generate: tools.verify.swagger ## Generate swagger.yaml from OpenAPI specs
	$(call print_target,$@)
	@echo "$(CYAN)Generating Swagger API documentation...$(NC)"
	@mkdir -p $(SWAGGER_DIR)
	@if [ -d "$(OPENAPI_DIR)" ] && [ -n "$$(find $(OPENAPI_DIR) -name '*.swagger.json' 2>/dev/null)" ]; then \
		echo "$(CYAN)Merging OpenAPI specs...$(NC)"; \
		swagger mixin $$(find $(OPENAPI_DIR) -name "*.swagger.json") \
			--quiet \
			--keep-spec-order \
			--format=yaml \
			--ignore-conflicts \
			-o $(SWAGGER_DIR)/swagger.yaml; \
		echo "$(GREEN)✓ Generated: $(SWAGGER_DIR)/swagger.yaml$(NC)"; \
	else \
		echo "$(YELLOW)⚠ No OpenAPI specs found in $(OPENAPI_DIR)$(NC)"; \
		echo "$(YELLOW)⚠ Creating placeholder swagger.yaml$(NC)"; \
		echo 'swagger: "2.0"' > $(SWAGGER_DIR)/swagger.yaml; \
		echo 'info:' >> $(SWAGGER_DIR)/swagger.yaml; \
		echo '  title: "k8s-agent API"' >> $(SWAGGER_DIR)/swagger.yaml; \
		echo '  version: "$(VERSION)"' >> $(SWAGGER_DIR)/swagger.yaml; \
		echo 'paths: {}' >> $(SWAGGER_DIR)/swagger.yaml; \
	fi

.PHONY: swagger.validate
swagger.validate: tools.verify.swagger ## Validate Swagger specification
	$(call print_target,$@)
	@if [ ! -f "$(SWAGGER_DIR)/swagger.yaml" ]; then \
		$(call print_error,"swagger.yaml not found. Run 'make swagger.generate' first."); \
		exit 1; \
	fi
	@echo "$(CYAN)Validating Swagger specification...$(NC)"
	@swagger validate $(SWAGGER_DIR)/swagger.yaml
	@echo "$(GREEN)✓ Swagger specification is valid$(NC)"

.PHONY: swagger.serve
swagger.serve: tools.verify.swagger ## Serve Swagger UI locally (port 65534)
	$(call print_target,$@)
	@if [ ! -f "$(SWAGGER_DIR)/swagger.yaml" ]; then \
		$(call print_warning,"swagger.yaml not found. Generating..."); \
		$(MAKE) swagger.generate; \
	fi
	@echo "$(CYAN)Starting Swagger UI server...$(NC)"
	@echo "$(YELLOW)Swagger UI will be available at: http://localhost:65534/docs$(NC)"
	@swagger serve -F=swagger --no-open --port 65534 $(SWAGGER_DIR)/swagger.yaml

.PHONY: swagger.serve.redoc
swagger.serve.redoc: tools.verify.swagger ## Serve ReDoc UI locally (port 65534)
	$(call print_target,$@)
	@if [ ! -f "$(SWAGGER_DIR)/swagger.yaml" ]; then \
		$(call print_warning,"swagger.yaml not found. Generating..."); \
		$(MAKE) swagger.generate; \
	fi
	@echo "$(CYAN)Starting ReDoc UI server...$(NC)"
	@echo "$(YELLOW)ReDoc will be available at: http://localhost:65534/docs$(NC)"
	@swagger serve -F=redoc --no-open --port 65534 $(SWAGGER_DIR)/swagger.yaml

.PHONY: swagger.clean
swagger.clean: ## Clean generated Swagger files
	$(call print_target,$@)
	@echo "$(CYAN)Cleaning Swagger documentation...$(NC)"
	@rm -f $(SWAGGER_DIR)/swagger.yaml
	@echo "$(GREEN)✓ Swagger files cleaned$(NC)"

.PHONY: swagger.install-ui
swagger.install-ui: ## Download and setup Swagger UI
	$(call print_target,$@)
	@echo "$(CYAN)Setting up Swagger UI...$(NC)"
	@mkdir -p $(SWAGGER_DIR)/ui
	@if [ ! -d "$(SWAGGER_DIR)/ui/dist" ]; then \
		echo "$(CYAN)Downloading Swagger UI...$(NC)"; \
		curl -sL https://github.com/swagger-api/swagger-ui/archive/refs/tags/v5.9.0.tar.gz | \
			tar xz -C $(SWAGGER_DIR)/ui --strip-components=2 swagger-ui-5.9.0/dist; \
		echo "$(GREEN)✓ Swagger UI installed to $(SWAGGER_DIR)/ui/dist$(NC)"; \
	else \
		echo "$(GREEN)✓ Swagger UI already installed$(NC)"; \
	fi

.PHONY: swagger.ui
swagger.ui: swagger.install-ui swagger.generate ## Setup and open Swagger UI in browser
	$(call print_target,$@)
	@if [ -f "$(SWAGGER_DIR)/swagger.yaml" ]; then \
		cp $(SWAGGER_DIR)/swagger.yaml $(SWAGGER_DIR)/ui/dist/; \
	fi
	@echo "$(YELLOW)Swagger UI is ready at: $(SWAGGER_DIR)/ui/dist/index.html$(NC)"
	@echo "$(YELLOW)Open it with: file://$(SWAGGER_DIR)/ui/dist/index.html?url=swagger.yaml$(NC)"

.PHONY: swagger.diff
swagger.diff: ## Compare current swagger.yaml with previous version
	$(call print_target,$@)
	@if [ ! -f "$(SWAGGER_DIR)/swagger.yaml" ]; then \
		$(call print_error,"swagger.yaml not found. Run 'make swagger.generate' first."); \
		exit 1; \
	fi
	@if git show HEAD:api/swagger/swagger.yaml > /tmp/swagger-old.yaml 2>/dev/null; then \
		echo "$(CYAN)Comparing with previous version...$(NC)"; \
		diff -u /tmp/swagger-old.yaml $(SWAGGER_DIR)/swagger.yaml || true; \
		rm -f /tmp/swagger-old.yaml; \
	else \
		echo "$(YELLOW)⚠ No previous version found in git$(NC)"; \
	fi
