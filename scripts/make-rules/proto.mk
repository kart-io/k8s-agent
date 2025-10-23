# Proto build rules
# Proto 文件和生成代码在同一位置 (pkg/api)

##@ Proto

PROTO_DIR := pkg/api
PROTO_GEN_DIR := pkg/api

.PHONY: proto.generate
proto.generate: ## Generate code from proto files using Buf
	$(call print_target,$@)
	@export PATH="$$PATH:$(shell go env GOPATH)/bin" && \
	buf generate
	$(call print_info,"Proto code generated successfully in $(PROTO_DIR)")

.PHONY: proto.lint
proto.lint: ## Lint proto files using Buf
	$(call print_target,$@)
	@buf lint
	$(call print_info,"Proto linting completed")

.PHONY: proto.breaking
proto.breaking: ## Check for breaking changes in proto files
	$(call print_target,$@)
	@buf breaking --against '.git#branch=master'
	$(call print_info,"Breaking changes check completed")

.PHONY: proto.format
proto.format: ## Format proto files using Buf
	$(call print_target,$@)
	@buf format -w
	$(call print_info,"Proto files formatted")

.PHONY: proto.dep.update
proto.dep.update: ## Update proto dependencies (googleapis, etc.)
	$(call print_target,$@)
	@buf dep update
	$(call print_info,"Proto dependencies updated")

.PHONY: proto.clean
proto.clean: ## Clean generated proto code (keeps .proto files)
	$(call print_target,$@)
	@find $(PROTO_DIR) -name '*.pb.go' -delete
	@find $(PROTO_DIR) -name '*.pb.gw.go' -delete
	@rm -f $(PROTO_DIR)/docs/swagger/*.json
	$(call print_info,"Generated proto code cleaned (proto files preserved)")

.PHONY: proto.clean.all
proto.clean.all: ## Clean all proto files (including .proto sources)
	$(call print_target,$@)
	@rm -rf $(PROTO_DIR)
	$(call print_info,"All proto files cleaned")

.PHONY: proto.push
proto.push: ## Push proto to Buf Schema Registry
	$(call print_target,$@)
	@buf push
	$(call print_info,"Proto pushed to registry")

.PHONY: proto.build
proto.build: ## Build proto image with Buf
	$(call print_target,$@)
	@buf build -o image.bin
	$(call print_info,"Proto image built")

.PHONY: proto.mod.update
proto.mod.update: ## Update buf.yaml module configuration
	$(call print_target,$@)
	@buf mod update
	$(call print_info,"Buf module configuration updated")

.PHONY: proto.info
proto.info: ## Show proto configuration info
	$(call print_target,$@)
	@echo "$(COLOR_BOLD)Proto Configuration:$(COLOR_RESET)"
	@echo "  Buf Config:       $(ROOT_DIR)/buf.yaml"
	@echo "  Gen Config:       $(ROOT_DIR)/buf.gen.yaml"
	@echo "  Proto Location:   $(PROTO_DIR) (sources + generated)"
	@echo "  Swagger Docs:     $(PROTO_DIR)/docs/swagger"
	@echo ""
	@echo "$(COLOR_BOLD)Proto Files:$(COLOR_RESET)"
	@find $(PROTO_DIR) -name "*.proto" -type f | wc -l | xargs echo "  Total .proto files:"
	@echo ""
	@echo "$(COLOR_BOLD)Generated Files:$(COLOR_RESET)"
	@find $(PROTO_DIR) -name "*.pb.go" -type f | wc -l | xargs echo "  Total .pb.go files:"
	@echo ""
	@echo "$(COLOR_BOLD)Available Commands:$(COLOR_RESET)"
	@echo "  make proto.generate    - Generate code from proto files"
	@echo "  make proto.lint        - Lint proto files"
	@echo "  make proto.breaking    - Check for breaking changes"
	@echo "  make proto.format      - Format proto files"
	@echo "  make proto.clean       - Clean generated code only"

.PHONY: proto.list
proto.list: ## List all proto files
	$(call print_target,$@)
	@echo "$(COLOR_BOLD)Proto Files:$(COLOR_RESET)"
	@find $(PROTO_DIR) -name "*.proto" -type f | sort
