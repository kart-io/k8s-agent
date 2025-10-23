# Kubernetes rules
# Based on OneX project structure

##@ Kubernetes

# Kubernetes configuration
K8S_DIR := $(ROOT_DIR)/deployments/k8s
K8S_MANIFESTS := $(shell find $(K8S_DIR) -name '*.yaml' -type f 2>/dev/null)

.PHONY: k8s.lint
k8s.lint: ## Lint Kubernetes manifests with kube-linter
	$(call print_target,$@)
	@if command -v kube-linter >/dev/null 2>&1; then \
		$(call print_info,"Linting Kubernetes manifests..."); \
		kube-linter lint $(K8S_DIR) --config .kube-linter.yaml; \
		$(call print_success,"Kubernetes manifests linted successfully"); \
	else \
		$(call print_error,"kube-linter not found. Run 'make tools.install.kube-linter'"); \
		exit 1; \
	fi

.PHONY: k8s.lint.fix
k8s.lint.fix: ## Auto-fix Kubernetes manifest issues where possible
	$(call print_target,$@)
	@$(call print_info,"Auto-fixing Kubernetes manifest issues...")
	@kube-linter lint $(K8S_DIR) --config .kube-linter.yaml --format json > /tmp/kube-linter-report.json || true
	@$(call print_info,"Report saved to /tmp/kube-linter-report.json")

.PHONY: k8s.validate
k8s.validate: ## Validate Kubernetes manifests with kubectl dry-run
	$(call print_target,$@)
	@$(call print_info,"Validating Kubernetes manifests...")
	@for file in $(K8S_MANIFESTS); do \
		echo "Validating $$file..."; \
		kubectl apply --dry-run=client -f $$file || exit 1; \
	done
	@$(call print_success,"All Kubernetes manifests are valid")

.PHONY: k8s.apply
k8s.apply: ## Apply Kubernetes manifests
	$(call print_target,$@)
	@$(call print_info,"Applying Kubernetes manifests...")
	@kubectl apply -f $(K8S_DIR)/namespace.yaml
	@kubectl apply -f $(K8S_DIR)/dependencies.yaml
	@kubectl apply -f $(K8S_DIR)/
	@$(call print_success,"Kubernetes manifests applied")

.PHONY: k8s.delete
k8s.delete: ## Delete Kubernetes resources
	$(call print_target,$@)
	@$(call print_warning,"This will delete all k8s-agent resources")
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		kubectl delete -f $(K8S_DIR)/ --ignore-not-found=true; \
		$(call print_success,"Kubernetes resources deleted"); \
	fi

.PHONY: k8s.status
k8s.status: ## Show status of Kubernetes resources
	$(call print_target,$@)
	@echo "Deployments:"
	@kubectl get deployments -n aetherius
	@echo ""
	@echo "Services:"
	@kubectl get services -n aetherius
	@echo ""
	@echo "Pods:"
	@kubectl get pods -n aetherius

.PHONY: k8s.logs
k8s.logs: ## Show logs for a service (SERVICE=agent-manager)
	$(call print_target,$@)
	@if [ -z "$(SERVICE)" ]; then \
		$(call print_error,"SERVICE not specified. Use: make k8s.logs SERVICE=agent-manager"); \
		exit 1; \
	fi
	@kubectl logs -n aetherius -l app=$(SERVICE) --tail=100 -f

.PHONY: k8s.describe
k8s.describe: ## Describe a service (SERVICE=agent-manager)
	$(call print_target,$@)
	@if [ -z "$(SERVICE)" ]; then \
		$(call print_error,"SERVICE not specified. Use: make k8s.describe SERVICE=agent-manager"); \
		exit 1; \
	fi
	@kubectl describe deployment -n aetherius $(SERVICE)

.PHONY: k8s.restart
k8s.restart: ## Restart a service (SERVICE=agent-manager)
	$(call print_target,$@)
	@if [ -z "$(SERVICE)" ]; then \
		$(call print_error,"SERVICE not specified. Use: make k8s.restart SERVICE=agent-manager"); \
		exit 1; \
	fi
	@kubectl rollout restart deployment/$(SERVICE) -n aetherius
	@kubectl rollout status deployment/$(SERVICE) -n aetherius

.PHONY: k8s.port-forward
k8s.port-forward: ## Port forward to a service (SERVICE=agent-manager PORT=8080)
	$(call print_target,$@)
	@if [ -z "$(SERVICE)" ] || [ -z "$(PORT)" ]; then \
		$(call print_error,"SERVICE and PORT required. Use: make k8s.port-forward SERVICE=agent-manager PORT=8080"); \
		exit 1; \
	fi
	@POD=$$(kubectl get pods -n aetherius -l app=$(SERVICE) -o jsonpath='{.items[0].metadata.name}'); \
	kubectl port-forward -n aetherius $$POD $(PORT):$(PORT)

.PHONY: k8s.shell
k8s.shell: ## Get shell access to a service pod (SERVICE=agent-manager)
	$(call print_target,$@)
	@if [ -z "$(SERVICE)" ]; then \
		$(call print_error,"SERVICE not specified. Use: make k8s.shell SERVICE=agent-manager"); \
		exit 1; \
	fi
	@POD=$$(kubectl get pods -n aetherius -l app=$(SERVICE) -o jsonpath='{.items[0].metadata.name}'); \
	kubectl exec -it -n aetherius $$POD -- /bin/sh
