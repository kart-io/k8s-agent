#!/bin/bash

# Auth Service MySQL Initialization Script
# This script reads database configuration from configs/config.yaml and initializes the MySQL database

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CONFIG_FILE="$PROJECT_ROOT/configs/config-dev.yaml"
SQL_FILE="$SCRIPT_DIR/init-mysql.sql"

echo -e "${GREEN}=== Auth Service MySQL Initialization ===${NC}"

# Check if config file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}Error: Config file not found at $CONFIG_FILE${NC}"
    exit 1
fi

# Check if SQL file exists
if [ ! -f "$SQL_FILE" ]; then
    echo -e "${RED}Error: SQL file not found at $SQL_FILE${NC}"
    exit 1
fi

# Check if yq is installed for YAML parsing
if ! command -v yq &> /dev/null; then
    echo -e "${YELLOW}Warning: yq not found. Attempting to parse config.yaml with grep/awk...${NC}"

    # Fallback: Parse YAML with grep and awk (simple parsing, works for simple YAML)
    DB_HOST=$(grep -A 10 "^database:" "$CONFIG_FILE" | grep "host:" | awk '{print $2}' | tr -d '"')
    DB_PORT=$(grep -A 10 "^database:" "$CONFIG_FILE" | grep "port:" | awk '{print $2}')
    DB_USER=$(grep -A 10 "^database:" "$CONFIG_FILE" | grep "user:" | awk '{print $2}' | tr -d '"')
    DB_PASSWORD=$(grep -A 10 "^database:" "$CONFIG_FILE" | grep "password:" | awk '{print $2}' | tr -d '"')
    DB_NAME=$(grep -A 10 "^database:" "$CONFIG_FILE" | grep "dbname:" | awk '{print $2}' | tr -d '"')
else
    # Use yq for proper YAML parsing
    DB_HOST=$(yq eval '.database.host' "$CONFIG_FILE")
    DB_PORT=$(yq eval '.database.port' "$CONFIG_FILE")
    DB_USER=$(yq eval '.database.user' "$CONFIG_FILE")
    DB_PASSWORD=$(yq eval '.database.password' "$CONFIG_FILE")
    DB_NAME=$(yq eval '.database.dbname' "$CONFIG_FILE")
fi

# Validate required fields
if [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ] || [ -z "$DB_USER" ] || [ -z "$DB_NAME" ]; then
    echo -e "${RED}Error: Missing required database configuration in config.yaml${NC}"
    echo "Required fields: database.host, database.port, database.user, database.dbname"
    exit 1
fi

# Display connection info (without password)
echo -e "${GREEN}Database Configuration:${NC}"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  User: $DB_USER"
echo "  Database: $DB_NAME"
echo ""

# Determine connection method based on config
# If host is 'localhost' or '127.0.0.1', check if we can use docker exec or TCP
if [ "$DB_HOST" = "localhost" ] || [ "$DB_HOST" = "127.0.0.1" ]; then
    # Try to detect if MySQL is in Docker container
    if docker ps 2>/dev/null | grep -q cluster-mysql; then
        echo -e "${GREEN}✓ Detected MySQL in Docker container 'cluster-mysql'${NC}"
        echo -e "${YELLOW}Using Docker exec for connection...${NC}"

        # Use Docker exec to run MySQL commands
        if [ -n "$DB_PASSWORD" ]; then
            MYSQL_CMD="docker exec -i cluster-mysql mysql -u $DB_USER -p$DB_PASSWORD"
        else
            MYSQL_CMD="docker exec -i cluster-mysql mysql -u $DB_USER"
        fi
    else
        # No Docker container found, try TCP connection to localhost
        echo -e "${YELLOW}Connecting to MySQL via TCP protocol...${NC}"
        MYSQL_CMD="mysql -h $DB_HOST -P $DB_PORT -u $DB_USER --protocol=TCP"

        if [ -n "$DB_PASSWORD" ]; then
            MYSQL_CMD="$MYSQL_CMD -p$DB_PASSWORD"
        fi
    fi
else
    # Remote MySQL server - use standard TCP connection
    echo -e "${YELLOW}Connecting to remote MySQL at ${DB_HOST}:${DB_PORT}...${NC}"
    MYSQL_CMD="mysql -h $DB_HOST -P $DB_PORT -u $DB_USER"

    if [ -n "$DB_PASSWORD" ]; then
        MYSQL_CMD="$MYSQL_CMD -p$DB_PASSWORD"
    fi
fi

# Display connection method
echo -e "${GREEN}Connection method: ${NC}${MYSQL_CMD% -p*}"  # Hide password in output

# Execute SQL script
echo -e "${GREEN}Executing initialization script...${NC}"
if $MYSQL_CMD < "$SQL_FILE"; then
    echo -e "${GREEN}✓ Database initialized successfully!${NC}"
    echo ""
    echo -e "${GREEN}Default Admin Account:${NC}"
    echo "  Username: admin"
    echo "  Password: admin123"
    echo "  Email: admin@example.com"
    echo ""
    echo -e "${YELLOW}⚠️  Please change the default password after first login!${NC}"
else
    echo -e "${RED}✗ Failed to initialize database${NC}"
    exit 1
fi
