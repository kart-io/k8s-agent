# ==============================================================================
# Makefile helper functions for deployment
#

.PHONY: deploy.run
deploy.run: deploy.k8s ## Run all deployments

.PHONY: deploy.k8s
deploy.k8s: ## Deploy to Kubernetes cluster
	$(call print_target,$@)
	@if [ -z "$(ENV)" ]; then \
		$(call print_error,"ENV not specified. Usage: make deploy.k8s ENV=dev|staging|prod"); \
		exit 1; \
	fi
	@echo "$(CYAN)Deploying to $(ENV) environment...$(NC)"
	@kubectl apply -k $(ROOT_DIR)/deployments/k8s/overlays/$(ENV)
	@echo "$(GREEN)✓ Deployed to $(ENV)$(NC)"

.PHONY: deploy.docker-compose
deploy.docker-compose: ## Deploy using Docker Compose
	$(call print_target,$@)
	@cd $(ROOT_DIR)/deployments/docker-compose && docker-compose up -d
	@echo "$(GREEN)✓ Docker Compose services started$(NC)"

.PHONY: deploy.verify
deploy.verify: ## Verify deployment status
	$(call print_target,$@)
	@echo "$(CYAN)Checking deployment status...$(NC)"
	@kubectl get deployments -A | grep k8s-agent || true
	@kubectl get pods -A | grep k8s-agent || true
	@kubectl get services -A | grep k8s-agent || true

.PHONY: deploy.logs
deploy.logs: ## Show deployment logs
	$(call print_target,$@)
	@if [ -z "$(SERVICE)" ]; then \
		$(call print_error,"SERVICE not specified. Usage: make deploy.logs SERVICE=agent-manager"); \
		exit 1; \
	fi
	@kubectl logs -f deployment/$(SERVICE) --all-containers=true

.PHONY: deploy.restart
deploy.restart: ## Restart deployments
	$(call print_target,$@)
	@echo "$(CYAN)Restarting deployments...$(NC)"
	@kubectl rollout restart deployment -l app.kubernetes.io/part-of=k8s-agent
	@echo "$(GREEN)✓ Deployments restarted$(NC)"

.PHONY: deploy.stop
deploy.stop: ## Stop all deployments
	$(call print_target,$@)
	@echo "$(CYAN)Stopping deployments...$(NC)"
	@kubectl delete deployment -l app.kubernetes.io/part-of=k8s-agent || true
	@echo "$(GREEN)✓ Deployments stopped$(NC)"

.PHONY: deploy.clean
deploy.clean: ## Clean up all deployed resources
	$(call print_target,$@)
	@echo "$(YELLOW)Cleaning up deployments...$(NC)"
	@kubectl delete all -l app.kubernetes.io/part-of=k8s-agent || true
	@echo "$(GREEN)✓ Cleanup complete$(NC)"

##@ Deployment
