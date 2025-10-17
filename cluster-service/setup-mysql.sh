#!/bin/bash

# MySQL Setup Script for cluster-service Development
# This script sets up a local MySQL database for development

set -e

echo "🚀 Setting up MySQL for cluster-service development..."
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
CONTAINER_NAME="cluster-mysql"
MYSQL_ROOT_PASSWORD="root123"
MYSQL_DATABASE="cluster_db"
MYSQL_USER="cluster_user"
MYSQL_PASSWORD="cluster_pass"
MYSQL_PORT="3306"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed. Please install Docker first.${NC}"
    exit 1
fi

# Check if container already exists
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo -e "${YELLOW}⚠️  Container '${CONTAINER_NAME}' already exists.${NC}"
    read -p "Do you want to remove it and create a new one? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Removing existing container..."
        docker rm -f ${CONTAINER_NAME}
    else
        echo "Checking if container is running..."
        if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
            echo "Starting existing container..."
            docker start ${CONTAINER_NAME}
        fi
        echo -e "${GREEN}✅ MySQL container is ready!${NC}"
        exit 0
    fi
fi

# Start MySQL container
echo "🐳 Starting MySQL 8.0 container..."
docker run -d \
  --name ${CONTAINER_NAME} \
  -p ${MYSQL_PORT}:3306 \
  -e MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD} \
  -e MYSQL_DATABASE=${MYSQL_DATABASE} \
  -e MYSQL_USER=${MYSQL_USER} \
  -e MYSQL_PASSWORD=${MYSQL_PASSWORD} \
  mysql:8.0 \
  --character-set-server=utf8mb4 \
  --collation-server=utf8mb4_unicode_ci

echo ""
echo "⏳ Waiting for MySQL to be ready (this may take 15-30 seconds)..."

# Wait for MySQL to be ready
MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if docker exec ${CONTAINER_NAME} mysqladmin ping -h localhost -u root -p${MYSQL_ROOT_PASSWORD} --silent 2>/dev/null; then
        echo -e "${GREEN}✅ MySQL is ready!${NC}"
        break
    fi

    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo -n "."
    sleep 1
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo -e "${RED}❌ MySQL failed to start within the expected time.${NC}"
    echo "Check container logs: docker logs ${CONTAINER_NAME}"
    exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}✅ MySQL setup completed successfully!${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📊 Database Connection Details:"
echo "   Host: localhost"
echo "   Port: ${MYSQL_PORT}"
echo "   Database: ${MYSQL_DATABASE}"
echo "   User: ${MYSQL_USER}"
echo "   Password: ${MYSQL_PASSWORD}"
echo ""
echo "🔧 Quick Commands:"
echo "   Start service:     make run-local"
echo "   Stop MySQL:        docker stop ${CONTAINER_NAME}"
echo "   Start MySQL:       docker start ${CONTAINER_NAME}"
echo "   Remove MySQL:      docker rm -f ${CONTAINER_NAME}"
echo "   View logs:         docker logs ${CONTAINER_NAME}"
echo "   MySQL shell:       docker exec -it ${CONTAINER_NAME} mysql -u ${MYSQL_USER} -p${MYSQL_PASSWORD} ${MYSQL_DATABASE}"
echo ""
echo "🚀 Next steps:"
echo "   1. Run the service: make run-local"
echo "   2. Test version API: curl http://localhost:8082/version"
echo ""
