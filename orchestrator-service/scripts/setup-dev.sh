#!/bin/bash

# orchestrator-service 开发环境快速设置脚本

set -e

echo "========================================="
echo "Orchestrator Service - 开发环境设置"
echo "========================================="
echo ""

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 检查 Docker 是否运行
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}错误: Docker 未运行${NC}"
    echo "请先启动 Docker Desktop"
    exit 1
fi

echo -e "${GREEN}✓ Docker 正在运行${NC}"
echo ""

# 1. 启动 PostgreSQL
echo -e "${YELLOW}[1/4] 启动 PostgreSQL...${NC}"
if docker ps | grep -q aetherius-postgres-dev; then
    echo -e "${GREEN}✓ PostgreSQL 已在运行${NC}"
elif docker ps -a | grep -q aetherius-postgres-dev; then
    echo "启动已存在的容器..."
    docker start aetherius-postgres-dev
    sleep 3
else
    echo "创建新的 PostgreSQL 容器..."
    docker run -d \
      --name aetherius-postgres-dev \
      -e POSTGRES_USER=postgres \
      -e POSTGRES_PASSWORD=dev-postgres-password \
      -e POSTGRES_DB=aetherius_orchestrator \
      -p 5432:5432 \
      postgres:14-alpine

    echo "等待 PostgreSQL 启动..."
    sleep 10
fi

# 验证 PostgreSQL
for i in {1..30}; do
    if docker exec aetherius-postgres-dev pg_isready -U postgres > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PostgreSQL 已就绪${NC}"
        break
    fi
    echo "等待 PostgreSQL 就绪... ($i/30)"
    sleep 1
done
echo ""

# 2. 启动 Redis
echo -e "${YELLOW}[2/4] 启动 Redis...${NC}"
if docker ps | grep -q aetherius-redis-dev; then
    echo -e "${GREEN}✓ Redis 已在运行${NC}"
elif docker ps -a | grep -q aetherius-redis-dev; then
    echo "启动已存在的容器..."
    docker start aetherius-redis-dev
    sleep 2
else
    echo "创建新的 Redis 容器..."
    docker run -d \
      --name aetherius-redis-dev \
      -e REDIS_PASSWORD=dev-redis-password \
      -p 6379:6379 \
      redis:7-alpine \
      redis-server --requirepass dev-redis-password

    sleep 3
fi

# 验证 Redis
if docker exec aetherius-redis-dev redis-cli -a dev-redis-password ping > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Redis 已就绪${NC}"
else
    echo -e "${YELLOW}⚠ Redis 可能未完全就绪${NC}"
fi
echo ""

# 3. 启动 NATS
echo -e "${YELLOW}[3/4] 启动 NATS...${NC}"
if docker ps | grep -q aetherius-nats-dev; then
    echo -e "${GREEN}✓ NATS 已在运行${NC}"
elif docker ps -a | grep -q aetherius-nats-dev; then
    echo "启动已存在的容器..."
    docker start aetherius-nats-dev
    sleep 2
else
    echo "创建新的 NATS 容器..."
    docker run -d \
      --name aetherius-nats-dev \
      -p 4222:4222 \
      -p 8222:8222 \
      nats:2.10-alpine \
      -js -m 8222

    sleep 3
fi

# 验证 NATS
if curl -s http://localhost:8222/healthz > /dev/null 2>&1; then
    echo -e "${GREEN}✓ NATS 已就绪${NC}"
else
    echo -e "${YELLOW}⚠ NATS 可能未完全就绪${NC}"
fi
echo ""

# 4. 初始化数据库（如果需要）
echo -e "${YELLOW}[4/4] 检查数据库...${NC}"
# 数据库会在服务启动时自动初始化
echo -e "${GREEN}✓ 数据库将在服务启动时自动初始化${NC}"
echo ""

# 显示服务状态
echo "========================================="
echo -e "${GREEN}开发环境已就绪！${NC}"
echo "========================================="
echo ""
echo "运行中的服务:"
docker ps --filter "name=aetherius-" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo ""
echo "连接信息:"
echo "  PostgreSQL: localhost:5432"
echo "    - 用户: postgres"
echo "    - 密码: dev-postgres-password"
echo "    - 数据库: aetherius_orchestrator"
echo ""
echo "  Redis: localhost:6379"
echo "    - 密码: dev-redis-password"
echo ""
echo "  NATS: localhost:4222"
echo "    - 监控: http://localhost:8222"
echo ""
echo "现在可以运行:"
echo -e "  ${GREEN}make run${NC}  或  ${GREEN}go run ./cmd/server${NC}"
echo ""
echo "停止服务:"
echo "  docker stop aetherius-postgres-dev aetherius-redis-dev aetherius-nats-dev"
echo ""
echo "清理环境:"
echo "  docker rm -f aetherius-postgres-dev aetherius-redis-dev aetherius-nats-dev"
echo ""
