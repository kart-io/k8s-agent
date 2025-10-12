#!/bin/bash

# Development Start Script for Cluster Service

set -e

echo "=========================================="
echo "Starting Cluster Service (Development)"
echo "=========================================="

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if PostgreSQL is running
echo -e "${YELLOW}Checking PostgreSQL...${NC}"
if ! pg_isready -h localhost -p 5432 > /dev/null 2>&1; then
    echo -e "${RED}Error: PostgreSQL is not running on localhost:5432${NC}"
    echo "Please start PostgreSQL first:"
    echo "  sudo systemctl start postgresql"
    echo "  or"
    echo "  docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:13"
    exit 1
fi
echo -e "${GREEN}✓ PostgreSQL is running${NC}"

# Check if database exists
echo -e "${YELLOW}Checking database...${NC}"
if ! psql -h localhost -U postgres -lqt | cut -d \| -f 1 | grep -qw cluster_dev; then
    echo -e "${YELLOW}Database 'cluster_dev' not found. Creating...${NC}"
    createdb -h localhost -U postgres cluster_dev
    echo -e "${GREEN}✓ Database created${NC}"
else
    echo -e "${GREEN}✓ Database exists${NC}"
fi

# Initialize schema
echo -e "${YELLOW}Initializing database schema...${NC}"
if [ -f "scripts/init-db.sql" ]; then
    # Modify init-db.sql to use cluster_dev
    sed 's/\\c cluster_db;/\\c cluster_dev;/g' scripts/init-db.sql | psql -h localhost -U postgres -q
    echo -e "${GREEN}✓ Schema initialized${NC}"
else
    echo -e "${YELLOW}Warning: init-db.sql not found, skipping schema initialization${NC}"
fi

# Build the service
echo -e "${YELLOW}Building cluster-service...${NC}"
go build -o server ./cmd/server
echo -e "${GREEN}✓ Build successful${NC}"

# Start the service
echo -e "${YELLOW}Starting cluster-service...${NC}"
echo "=========================================="
./server -config configs/config.dev.yaml
