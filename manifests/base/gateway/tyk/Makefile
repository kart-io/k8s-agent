.PHONY: help start stop restart status logs clean test reload validate setup

# 默认目标
.DEFAULT_GOAL := help

# 颜色输出
GREEN  := \033[0;32m
YELLOW := \033[1;33m
RED    := \033[0;31m
NC     := \033[0m # No Color

help: ## 显示帮助信息
	@echo ""
	@echo "$(GREEN)Tyk Gateway 管理命令$(NC)"
	@echo ""
	@echo "可用命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'
	@echo ""

setup: ## 初始化配置文件
	@echo "$(YELLOW)初始化配置...$(NC)"
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "$(GREEN)✓ 已创建 .env 文件$(NC)"; \
	else \
		echo "$(YELLOW)⚠ .env 文件已存在$(NC)"; \
	fi

start: ## 启动所有服务
	@echo "$(YELLOW)启动 Tyk Gateway 服务...$(NC)"
	@docker-compose up -d
	@echo "$(GREEN)✓ 服务启动成功$(NC)"
	@make status

stop: ## 停止所有服务
	@echo "$(YELLOW)停止服务...$(NC)"
	@docker-compose down
	@echo "$(GREEN)✓ 服务已停止$(NC)"

restart: ## 重启所有服务
	@echo "$(YELLOW)重启服务...$(NC)"
	@docker-compose restart
	@echo "$(GREEN)✓ 服务已重启$(NC)"

status: ## 查看服务状态
	@echo "$(YELLOW)服务状态:$(NC)"
	@docker-compose ps

logs: ## 查看所有服务日志
	@docker-compose logs -f

logs-gateway: ## 查看 Gateway 日志
	@docker-compose logs -f tyk-gateway

logs-dashboard: ## 查看 Dashboard 日志
	@docker-compose logs -f tyk-dashboard

logs-pump: ## 查看 Pump 日志
	@docker-compose logs -f tyk-pump

clean: ## 停止并清理所有数据
	@echo "$(RED)警告: 这将删除所有数据!$(NC)"
	@read -p "确认继续? [y/N] " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		docker-compose down -v; \
		echo "$(GREEN)✓ 清理完成$(NC)"; \
	else \
		echo "$(YELLOW)操作已取消$(NC)"; \
	fi

reload: ## 重载 API 定义
	@echo "$(YELLOW)重载 API 定义...$(NC)"
	@curl -s -H "x-tyk-authorization: 352d20ee67be67f6340b4c0605b044b7" \
		http://localhost:8080/tyk/reload/group > /dev/null && \
		echo "$(GREEN)✓ API 定义已重载$(NC)" || \
		echo "$(RED)✗ 重载失败$(NC)"

test: ## 运行测试
	@echo "$(YELLOW)测试 Tyk Gateway...$(NC)"
	@echo ""
	@echo "测试 Gateway 健康检查..."
	@curl -sf http://localhost:8080/hello > /dev/null && \
		echo "$(GREEN)✓ Gateway 运行正常$(NC)" || \
		echo "$(RED)✗ Gateway 未响应$(NC)"
	@echo ""
	@echo "测试 Dashboard 健康检查..."
	@curl -sf http://localhost:3000/hello > /dev/null && \
		echo "$(GREEN)✓ Dashboard 运行正常$(NC)" || \
		echo "$(YELLOW)⚠ Dashboard 未响应$(NC)"
	@echo ""
	@echo "测试 Redis 连接..."
	@docker-compose exec -T redis redis-cli ping > /dev/null 2>&1 && \
		echo "$(GREEN)✓ Redis 运行正常$(NC)" || \
		echo "$(RED)✗ Redis 未响应$(NC)"

validate: ## 验证配置文件
	@echo "$(YELLOW)验证配置文件...$(NC)"
	@echo ""
	@echo "检查 tyk.conf..."
	@cat tyk.conf | jq . > /dev/null 2>&1 && \
		echo "$(GREEN)✓ tyk.conf 格式正确$(NC)" || \
		echo "$(RED)✗ tyk.conf 格式错误$(NC)"
	@echo ""
	@echo "检查 pump.conf..."
	@cat pump.conf | jq . > /dev/null 2>&1 && \
		echo "$(GREEN)✓ pump.conf 格式正确$(NC)" || \
		echo "$(RED)✗ pump.conf 格式错误$(NC)"
	@echo ""
	@echo "检查 API 定义..."
	@for file in apps/*.json; do \
		cat $$file | jq . > /dev/null 2>&1 && \
			echo "$(GREEN)✓ $$file 格式正确$(NC)" || \
			echo "$(RED)✗ $$file 格式错误$(NC)"; \
	done
	@echo ""
	@echo "检查策略定义..."
	@cat policies/policies.json | jq . > /dev/null 2>&1 && \
		echo "$(GREEN)✓ policies.json 格式正确$(NC)" || \
		echo "$(RED)✗ policies.json 格式错误$(NC)"

info: ## 显示服务信息
	@echo ""
	@echo "$(GREEN)========================================$(NC)"
	@echo "$(GREEN)  Tyk Gateway 服务信息$(NC)"
	@echo "$(GREEN)========================================$(NC)"
	@echo ""
	@echo "$(YELLOW)访问地址:$(NC)"
	@echo "  Tyk Gateway:   http://localhost:8080"
	@echo "  Tyk Dashboard: http://localhost:3000"
	@echo "  Prometheus:    http://localhost:9090"
	@echo ""
	@echo "$(YELLOW)API 端点:$(NC)"
	@echo "  健康检查: http://localhost:8080/hello"
	@echo "  认证服务: http://localhost:8080/api/v1/auth/"
	@echo "  Agent 管理: http://localhost:8080/api/v1/agents"
	@echo ""
	@echo "$(YELLOW)常用命令:$(NC)"
	@echo "  查看日志: make logs"
	@echo "  重载 API: make reload"
	@echo "  运行测试: make test"
	@echo ""

shell-gateway: ## 进入 Gateway 容器
	@docker-compose exec tyk-gateway sh

shell-redis: ## 进入 Redis 容器
	@docker-compose exec redis redis-cli

backup: ## 备份配置文件
	@echo "$(YELLOW)备份配置文件...$(NC)"
	@mkdir -p backups
	@tar czf backups/tyk-backup-$$(date +%Y%m%d-%H%M%S).tar.gz \
		tyk.conf pump.conf apps/ policies/ .env 2>/dev/null || true
	@echo "$(GREEN)✓ 备份完成$(NC)"
	@ls -lh backups/ | tail -1

restore: ## 恢复配置文件(需指定备份文件)
	@if [ -z "$(file)" ]; then \
		echo "$(RED)错误: 请指定备份文件$(NC)"; \
		echo "用法: make restore file=backups/tyk-backup-XXXXXX.tar.gz"; \
		exit 1; \
	fi
	@echo "$(YELLOW)恢复配置文件...$(NC)"
	@tar xzf $(file)
	@echo "$(GREEN)✓ 恢复完成$(NC)"

ps: ## 查看容器进程
	@docker-compose ps

top: ## 查看容器资源使用
	@docker stats $$(docker-compose ps -q)
