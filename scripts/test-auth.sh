#!/bin/bash

# Auth Service Testing Script
# Usage: ./scripts/test-auth.sh

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
AUTH_URL="${AUTH_URL:-http://localhost:8080}"
USERNAME="${USERNAME:-admin}"
PASSWORD="${PASSWORD:-admin123}"

echo -e "${GREEN}=== Auth Service Test Script ===${NC}"
echo -e "Auth URL: ${AUTH_URL}"
echo -e "Username: ${USERNAME}"
echo ""

# Step 1: Login
echo -e "${YELLOW}Step 1: Login with credentials${NC}"
echo -e "POST ${AUTH_URL}/api/v1/auth/login"
echo -e "Credentials: ${USERNAME}:${PASSWORD}"
echo ""

LOGIN_RESPONSE=$(curl -s -X POST "${AUTH_URL}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"${USERNAME}\",
    \"password\": \"${PASSWORD}\"
  }")

# Check if login was successful
if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Failed to connect to auth service${NC}"
    exit 1
fi

# Parse response
ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token // .access_token // .data.access_token // empty' 2>/dev/null)
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.refresh_token // .refresh_token // empty' 2>/dev/null)
ERROR_MSG=$(echo "$LOGIN_RESPONSE" | jq -r '.message // .error // empty' 2>/dev/null)

if [ -z "$ACCESS_TOKEN" ]; then
    echo -e "${RED}Error: Failed to get access token${NC}"
    echo "Response: $LOGIN_RESPONSE"
    if [ ! -z "$ERROR_MSG" ]; then
        echo -e "${RED}Error message: $ERROR_MSG${NC}"
    fi
    exit 1
fi

echo -e "${GREEN}✓ Login successful${NC}"
echo "Access Token (first 50 chars): ${ACCESS_TOKEN:0:50}..."
echo ""

# Step 2: Decode JWT to get JTI
echo -e "${YELLOW}Step 2: Decode JWT Token${NC}"
# Extract payload (second part of JWT)
JWT_PAYLOAD=$(echo "$ACCESS_TOKEN" | cut -d. -f2)
# Add padding if needed
while [ $((${#JWT_PAYLOAD} % 4)) -ne 0 ]; do
    JWT_PAYLOAD="${JWT_PAYLOAD}="
done
# Decode
DECODED_JWT=$(echo "$JWT_PAYLOAD" | base64 -d 2>/dev/null)

if [ ! -z "$DECODED_JWT" ]; then
    JTI=$(echo "$DECODED_JWT" | jq -r '.jti // empty' 2>/dev/null)
    USER_ID=$(echo "$DECODED_JWT" | jq -r '.user_id // .sub // empty' 2>/dev/null)
    USERNAME_JWT=$(echo "$DECODED_JWT" | jq -r '.username // empty' 2>/dev/null)
    EXP=$(echo "$DECODED_JWT" | jq -r '.exp // empty' 2>/dev/null)

    echo "JWT Claims:"
    echo "  JTI: $JTI"
    echo "  User ID: $USER_ID"
    echo "  Username: $USERNAME_JWT"
    if [ ! -z "$EXP" ]; then
        # Convert timestamp to readable date (works on macOS and Linux)
        if command -v gdate &> /dev/null; then
            # macOS with GNU date installed
            EXPIRY_DATE=$(gdate -d "@$EXP" "+%Y-%m-%d %H:%M:%S")
        elif date --version >/dev/null 2>&1; then
            # Linux with GNU date
            EXPIRY_DATE=$(date -d "@$EXP" "+%Y-%m-%d %H:%M:%S")
        else
            # macOS with BSD date
            EXPIRY_DATE=$(date -r "$EXP" "+%Y-%m-%d %H:%M:%S" 2>/dev/null || echo "Timestamp: $EXP")
        fi
        echo "  Expires: $EXPIRY_DATE"
    fi
fi
echo ""

# Step 3: Test /me endpoint
echo -e "${YELLOW}Step 3: Test /api/v1/auth/me endpoint${NC}"
echo -e "GET ${AUTH_URL}/api/v1/auth/me"
echo ""

ME_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X GET "${AUTH_URL}/api/v1/auth/me" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Accept: application/json")

# Extract HTTP status code
HTTP_STATUS=$(echo "$ME_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
RESPONSE_BODY=$(echo "$ME_RESPONSE" | sed '/HTTP_STATUS:/d')

echo "HTTP Status: $HTTP_STATUS"
echo "Response:"
echo "$RESPONSE_BODY" | jq '.' 2>/dev/null || echo "$RESPONSE_BODY"

if [ "$HTTP_STATUS" = "200" ]; then
    echo -e "${GREEN}✓ /me endpoint successful${NC}"
else
    echo -e "${RED}✗ /me endpoint failed${NC}"

    # Check for specific error messages
    if echo "$RESPONSE_BODY" | grep -q "session has been terminated"; then
        echo -e "${YELLOW}Diagnosis: Session was terminated by administrator${NC}"
        echo "Possible reasons:"
        echo "  1. Admin forced logout this session"
        echo "  2. Session expired in Redis"
        echo "  3. Redis was restarted and lost session data"
        echo "  4. Session was never created during login"
    elif echo "$RESPONSE_BODY" | grep -q "token is expired"; then
        echo -e "${YELLOW}Diagnosis: JWT token has expired${NC}"
    elif echo "$RESPONSE_BODY" | grep -q "invalid token"; then
        echo -e "${YELLOW}Diagnosis: JWT token is invalid${NC}"
    fi
fi
echo ""

# Step 4: Check Redis session (if JTI exists and docker-compose is available)
if [ ! -z "$JTI" ] && command -v docker-compose &> /dev/null; then
    echo -e "${YELLOW}Step 4: Check Redis for session data${NC}"
    echo "Checking Redis keys for JTI: $JTI"
    echo ""

    # Check if docker-compose file exists
    COMPOSE_FILE="deployments/docker-compose/docker-compose.yaml"
    if [ -f "$COMPOSE_FILE" ]; then
        # Check session key
        echo "Checking session:$JTI ..."
        SESSION_DATA=$(docker-compose -f "$COMPOSE_FILE" exec -T redis redis-cli GET "session:$JTI" 2>/dev/null)
        if [ ! -z "$SESSION_DATA" ]; then
            echo -e "${GREEN}✓ Session exists in Redis${NC}"
            echo "Session data: $SESSION_DATA"
        else
            echo -e "${RED}✗ Session NOT found in Redis${NC}"
        fi

        # Check if revoked
        echo ""
        echo "Checking revoked:$JTI ..."
        REVOKED_DATA=$(docker-compose -f "$COMPOSE_FILE" exec -T redis redis-cli GET "revoked:$JTI" 2>/dev/null)
        if [ ! -z "$REVOKED_DATA" ]; then
            echo -e "${RED}✗ Session is REVOKED${NC}"
            echo "Revoked data: $REVOKED_DATA"
        else
            echo -e "${GREEN}✓ Session is not revoked${NC}"
        fi

        # Check user sessions
        if [ ! -z "$USER_ID" ]; then
            echo ""
            echo "Checking user:sessions:$USER_ID ..."
            USER_SESSIONS=$(docker-compose -f "$COMPOSE_FILE" exec -T redis redis-cli ZRANGE "user:sessions:$USER_ID" 0 -1 2>/dev/null)
            if [ ! -z "$USER_SESSIONS" ]; then
                echo "User's active sessions:"
                echo "$USER_SESSIONS"
            else
                echo -e "${YELLOW}No active sessions found for user${NC}"
            fi
        fi
    else
        echo -e "${YELLOW}Docker-compose file not found at $COMPOSE_FILE${NC}"
        echo "Cannot check Redis directly"
    fi
fi

echo ""
echo -e "${GREEN}=== Test Complete ===${NC}"

# Step 5: Additional diagnostic endpoints (optional)
echo ""
echo -e "${YELLOW}Step 5: Additional diagnostics${NC}"

# Test refresh token endpoint
if [ ! -z "$REFRESH_TOKEN" ]; then
    echo "Testing token refresh..."
    REFRESH_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${AUTH_URL}/api/v1/auth/refresh" \
      -H "Content-Type: application/json" \
      -d "{\"refresh_token\": \"${REFRESH_TOKEN}\"}")

    REFRESH_STATUS=$(echo "$REFRESH_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
    if [ "$REFRESH_STATUS" = "200" ]; then
        echo -e "${GREEN}✓ Token refresh works${NC}"
    else
        echo -e "${RED}✗ Token refresh failed (Status: $REFRESH_STATUS)${NC}"
    fi
fi

# Test logout endpoint
echo ""
echo "Testing logout..."
LOGOUT_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${AUTH_URL}/api/v1/auth/logout" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}")

LOGOUT_STATUS=$(echo "$LOGOUT_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
if [ "$LOGOUT_STATUS" = "200" ]; then
    echo -e "${GREEN}✓ Logout successful${NC}"

    # Try /me again after logout (should fail)
    echo "Verifying token is invalidated..."
    ME_AFTER_LOGOUT=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X GET "${AUTH_URL}/api/v1/auth/me" \
      -H "Authorization: Bearer ${ACCESS_TOKEN}")

    AFTER_LOGOUT_STATUS=$(echo "$ME_AFTER_LOGOUT" | grep "HTTP_STATUS:" | cut -d: -f2)
    if [ "$AFTER_LOGOUT_STATUS" = "401" ]; then
        echo -e "${GREEN}✓ Token properly invalidated after logout${NC}"
    else
        echo -e "${YELLOW}⚠ Token still valid after logout (Status: $AFTER_LOGOUT_STATUS)${NC}"
    fi
else
    echo -e "${RED}✗ Logout failed (Status: $LOGOUT_STATUS)${NC}"
fi

echo ""
echo "=== Summary ==="
echo "Auth Service URL: $AUTH_URL"
echo "Test User: $USERNAME"
if [ "$HTTP_STATUS" = "200" ]; then
    echo -e "Result: ${GREEN}PASSED${NC} - Authentication working correctly"
else
    echo -e "Result: ${RED}FAILED${NC} - Check the diagnostics above"
    echo ""
    echo "Troubleshooting tips:"
    echo "1. Check if auth service is running: docker-compose ps auth"
    echo "2. Check auth service logs: docker-compose logs -f auth"
    echo "3. Check Redis is running: docker-compose ps redis"
    echo "4. Check database is running: docker-compose ps mysql"
    echo "5. Verify credentials in database"
fi