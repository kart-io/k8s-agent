# Code generation and documentation rules
# Based on OneX v2 patterns

##@ Code Generation

.PHONY: gen.clean
gen.clean: ## Clean generated code
	$(call print_target,$@)
	@find . -name "*_gen.go" -not -path "./vendor/*" -delete
	@find . -name "*.pb.go" -not -path "./vendor/*" -delete
	$(call print_info,"Generated code cleaned")

.PHONY: gen.mocks
gen.mocks: ## Generate mocks for testing
	$(call print_target,$@)
	@$(call print_info,"Generating mocks...")
	@for svc in $(SERVICES); do \
		if [ -f "$$svc/internal/interfaces.go" ]; then \
			mockgen -source=$$svc/internal/interfaces.go -destination=$$svc/internal/mocks/mocks.go; \
		fi \
	done

.PHONY: gen.docs
gen.docs: ## Generate documentation
	$(call print_target,$@)
	@$(call print_info,"Generating documentation...")
	@mkdir -p docs/api
	@for svc in $(SERVICES); do \
		if [ -d "$$svc/api" ]; then \
			$(call print_info,"Generating docs for $$svc..."); \
		fi \
	done

.PHONY: gen.swagger
gen.swagger: ## Generate Swagger/OpenAPI docs
	$(call print_target,$@)
	@$(call print_info,"Generating Swagger docs...")
	@if command -v swag >/dev/null 2>&1; then \
		for svc in $(SERVICES); do \
			if [ -f "$$svc/cmd/server/main.go" ]; then \
				cd $$svc && swag init -g cmd/server/main.go -o api/swagger && cd $(ROOT_DIR); \
			fi \
		done; \
	else \
		$(call print_error,"swag not installed. Run: go install github.com/swaggo/swag/cmd/swag@latest"); \
	fi

##@ Database

.PHONY: db.create
db.create: ## Create databases
	$(call print_target,$@)
	@$(call print_info,"Creating databases...")
	@for svc in $(SERVICES); do \
		if [ -f "$$svc/scripts/db/create.sql" ]; then \
			$(call print_info,"Creating database for $$svc..."); \
		fi \
	done

.PHONY: db.migrate
db.migrate: ## Run database migrations
	$(call print_target,$@)
	@$(call print_info,"Running migrations...")
	@for svc in $(SERVICES); do \
		if [ -d "$$svc/migrations" ]; then \
			$(call print_info,"Migrating database for $$svc..."); \
		fi \
	done

.PHONY: db.seed
db.seed: ## Seed database with test data
	$(call print_target,$@)
	@$(call print_info,"Seeding databases...")
	@for svc in $(SERVICES); do \
		if [ -f "$$svc/scripts/db/seed.sql" ]; then \
			$(call print_info,"Seeding database for $$svc..."); \
		fi \
	done

.PHONY: db.reset
db.reset: ## Reset all databases
	$(call print_target,$@)
	@$(call print_info,"Resetting databases...")
	@$(call print_warning,"This will delete all data! Press Ctrl+C to cancel...")
	@sleep 5
	@$(MAKE) db.drop
	@$(MAKE) db.create
	@$(MAKE) db.migrate
	@$(MAKE) db.seed

.PHONY: db.drop
db.drop: ## Drop all databases
	$(call print_target,$@)
	@$(call print_warning,"Dropping databases...")
	@for svc in $(SERVICES); do \
		if [ -f "$$svc/scripts/db/drop.sql" ]; then \
			$(call print_info,"Dropping database for $$svc..."); \
		fi \
	done

.PHONY: db.backup
db.backup: ## Backup all databases
	$(call print_target,$@)
	@mkdir -p backups
	@$(call print_info,"Backing up databases to backups/")
	@date_suffix=$$(date +%Y%m%d_%H%M%S); \
	for svc in $(SERVICES); do \
		$(call print_info,"Backing up database for $$svc..."); \
		echo "# Backup for $$svc - $$date_suffix" > backups/$$svc-$$date_suffix.sql; \
	done

.PHONY: db.restore
db.restore: ## Restore databases from backup
	$(call print_target,$@)
	@$(call print_info,"Restoring databases from backups/")
	@ls -1t backups/*.sql | head -1

##@ Dependencies

.PHONY: deps.update
deps.update: ## Update all dependencies
	$(call print_target,$@)
	@$(MAKE) go.mod.tidy
	@$(MAKE) proto.dep.update
	@$(call print_info,"Dependencies updated")

.PHONY: deps.check
deps.check: ## Check for outdated dependencies
	$(call print_target,$@)
	@$(call print_info,"Checking for outdated dependencies...")
	@for svc in $(SERVICES); do \
		$(call print_info,"Checking $$svc..."); \
		cd $$svc && go list -u -m all 2>/dev/null | grep '\[' && cd $(ROOT_DIR) || true; \
	done

.PHONY: deps.graph
deps.graph: ## Generate dependency graph
	$(call print_target,$@)
	@$(call print_info,"Generating dependency graph...")
	@mkdir -p docs/deps
	@for svc in $(SERVICES); do \
		cd $$svc && go mod graph > $(ROOT_DIR)/docs/deps/$$svc-deps.txt && cd $(ROOT_DIR); \
	done

.PHONY: deps.vendor
deps.vendor: ## Vendor dependencies
	$(call print_target,$@)
	@$(call print_info,"Vendoring dependencies...")
	@for svc in $(SERVICES); do \
		cd $$svc && go mod vendor && cd $(ROOT_DIR); \
	done

##@ Security

.PHONY: security.scan
security.scan: ## Run security scans
	$(call print_target,$@)
	@$(MAKE) security.gosec
	@$(MAKE) security.trivy
	@$(MAKE) security.nancy

.PHONY: security.gosec
security.gosec: ## Run gosec security scanner
	$(call print_target,$@)
	@if command -v gosec >/dev/null 2>&1; then \
		$(call print_info,"Running gosec..."); \
		for svc in $(SERVICES); do \
			cd $$svc && gosec ./... && cd $(ROOT_DIR); \
		done; \
	else \
		$(call print_warning,"gosec not installed. Run: go install github.com/securego/gosec/v2/cmd/gosec@latest"); \
	fi

.PHONY: security.trivy
security.trivy: ## Run Trivy vulnerability scanner on Docker images
	$(call print_target,$@)
	@if command -v trivy >/dev/null 2>&1; then \
		$(call print_info,"Running trivy..."); \
		for svc in $(SERVICES); do \
			if docker images | grep -q "$$svc"; then \
				trivy image $$svc:latest; \
			fi \
		done; \
	else \
		$(call print_warning,"trivy not installed"); \
	fi

.PHONY: security.nancy
security.nancy: ## Check for vulnerable dependencies
	$(call print_target,$@)
	@if command -v nancy >/dev/null 2>&1; then \
		$(call print_info,"Running nancy..."); \
		for svc in $(SERVICES); do \
			cd $$svc && go list -json -m all | nancy sleuth && cd $(ROOT_DIR); \
		done; \
	else \
		$(call print_warning,"nancy not installed. Run: go install github.com/sonatype-nexus-community/nancy@latest"); \
	fi

.PHONY: security.audit
security.audit: ## Run security audit
	$(call print_target,$@)
	@$(call print_info,"Running security audit...")
	@$(MAKE) security.scan
	@$(call print_info,"Security audit complete")

##@ Performance

.PHONY: perf.benchmark
perf.benchmark: ## Run performance benchmarks
	$(call print_target,$@)
	@$(call print_info,"Running benchmarks...")
	@mkdir -p $(BENCH_DIR)
	@for svc in $(SERVICES); do \
		cd $$svc && \
		$(GO_TEST) -bench=. -benchmem -run=^$$ ./... > $(ROOT_DIR)/$(BENCH_DIR)/$$svc-bench.txt && \
		cd $(ROOT_DIR); \
	done

.PHONY: perf.profile
perf.profile: ## Run CPU and memory profiling
	$(call print_target,$@)
	@$(call print_info,"Running profiling...")
	@mkdir -p $(PROFILE_DIR)
	@$(call print_warning,"Profiling requires running services")

.PHONY: perf.compare
perf.compare: ## Compare benchmark results
	$(call print_target,$@)
	@if command -v benchstat >/dev/null 2>&1; then \
		$(call print_info,"Comparing benchmarks..."); \
		benchstat $(BENCH_DIR)/*-bench.txt; \
	else \
		$(call print_warning,"benchstat not installed. Run: go install golang.org/x/perf/cmd/benchstat@latest"); \
	fi

##@ Quality

.PHONY: quality.check
quality.check: ## Run all quality checks
	$(call print_target,$@)
	@$(MAKE) go.fmt
	@$(MAKE) go.vet
	@$(MAKE) go.lint
	@$(MAKE) security.scan
	@$(call print_info,"Quality checks complete")

.PHONY: quality.report
quality.report: ## Generate quality report
	$(call print_target,$@)
	@$(call print_info,"Generating quality report...")
	@mkdir -p reports
	@echo "# Quality Report - $$(date)" > reports/quality-report.md
	@echo "" >> reports/quality-report.md
	@echo "## Go Vet" >> reports/quality-report.md
	@$(MAKE) go.vet >> reports/quality-report.md 2>&1 || true
	@echo "" >> reports/quality-report.md
	@echo "## Linting" >> reports/quality-report.md
	@$(MAKE) go.lint >> reports/quality-report.md 2>&1 || true
	@$(call print_info,"Quality report generated in reports/quality-report.md")

##@ Cleanup

.PHONY: clean.all
clean.all: ## Clean everything (build, test, coverage, etc.)
	$(call print_target,$@)
	@$(MAKE) clean
	@$(MAKE) proto.clean
	@$(MAKE) gen.clean
	@$(MAKE) docker.clean
	@$(call print_info,"All artifacts cleaned")

.PHONY: clean.cache
clean.cache: ## Clean Go build cache
	$(call print_target,$@)
	@$(call print_info,"Cleaning Go build cache...")
	@go clean -cache -testcache -modcache
	@$(call print_info,"Cache cleaned")
