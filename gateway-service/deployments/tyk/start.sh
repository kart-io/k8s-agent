#!/bin/bash

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Tyk API Gateway 启动脚本${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo -e "${RED}错误: Docker 未安装${NC}"
    exit 1
fi

# 检查 Docker Compose 是否安装
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}错误: Docker Compose 未安装${NC}"
    exit 1
fi

# 切换到脚本所在目录
cd "$(dirname "$0")"

# 检查配置文件
if [ ! -f "tyk.conf" ]; then
    echo -e "${RED}错误: tyk.conf 配置文件不存在${NC}"
    exit 1
fi

# 创建 .env 文件
if [ ! -f ".env" ]; then
    echo -e "${YELLOW}警告: .env 文件不存在，从 .env.example 复制${NC}"
    cp .env.example .env
fi

# 启动服务
echo -e "${GREEN}启动 Tyk Gateway 服务...${NC}"
docker-compose up -d

# 等待服务启动
echo -e "${YELLOW}等待服务启动...${NC}"
sleep 10

# 检查服务状态
echo ""
echo -e "${GREEN}检查服务状态:${NC}"
docker-compose ps

# 健康检查
echo ""
echo -e "${GREEN}执行健康检查...${NC}"

# 检查 Redis
if docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Redis 运行正常${NC}"
else
    echo -e "${RED}✗ Redis 未响应${NC}"
fi

# 检查 Tyk Gateway
if curl -sf http://localhost:8080/hello > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Tyk Gateway 运行正常${NC}"
else
    echo -e "${RED}✗ Tyk Gateway 未响应${NC}"
fi

# 检查 Tyk Dashboard
if curl -sf http://localhost:3000/hello > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Tyk Dashboard 运行正常${NC}"
else
    echo -e "${YELLOW}⚠ Tyk Dashboard 未响应 (可能仍在启动)${NC}"
fi

# 显示访问信息
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  服务启动完成!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "Tyk Gateway: ${GREEN}http://localhost:8080${NC}"
echo -e "Tyk Dashboard: ${GREEN}http://localhost:3000${NC}"
echo -e "Redis: ${GREEN}localhost:6379${NC}"
echo -e "PostgreSQL: ${GREEN}localhost:5432${NC}"
echo ""
echo -e "${YELLOW}查看日志: docker-compose logs -f${NC}"
echo -e "${YELLOW}停止服务: docker-compose down${NC}"
echo ""
