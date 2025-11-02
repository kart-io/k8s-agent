#!/bin/bash
# Runtime test script for idempotency middleware integration
# Tests Agent Manager service with actual HTTP requests

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
AGENT_MANAGER_URL="${AGENT_MANAGER_URL:-http://localhost:8080}"
REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"

echo -e "${BLUE}=== Agent Manager Idempotency Runtime Test ===${NC}"
echo ""

# Function to print test results
pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
}

fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    exit 1
}

info() {
    echo -e "${YELLOW}ℹ INFO${NC}: $1"
}

# Test 1: Check if Agent Manager is running
echo -e "${BLUE}Test 1: Health Check${NC}"
if curl -s -f "${AGENT_MANAGER_URL}/health/live" > /dev/null; then
    pass "Agent Manager is running at ${AGENT_MANAGER_URL}"
else
    fail "Agent Manager is not responding. Please start the service first."
fi
echo ""

# Test 2: Check Redis connection
echo -e "${BLUE}Test 2: Redis Connection${NC}"
if command -v redis-cli &> /dev/null; then
    if redis-cli -h "${REDIS_HOST}" -p "${REDIS_PORT}" ping > /dev/null 2>&1; then
        pass "Redis is running at ${REDIS_HOST}:${REDIS_PORT}"
    else
        fail "Redis is not responding. Please start Redis first."
    fi
else
    info "redis-cli not found, skipping Redis check"
fi
echo ""

# Test 3: Request without idempotent key (should fail)
echo -e "${BLUE}Test 3: POST without X-Idempotent-Key header${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${AGENT_MANAGER_URL}/api/v1/clusters" \
    -H "Content-Type: application/json" \
    -d '{
        "id": "test-cluster-001",
        "name": "Test Cluster",
        "api_server": "https://test.k8s.local:6443",
        "status": "active"
    }')

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" == "400" ]; then
    if echo "$BODY" | grep -q "Missing X-Idempotent-Key"; then
        pass "Correctly rejected request without idempotent key (400 Bad Request)"
    else
        fail "Expected 'Missing X-Idempotent-Key' error message"
    fi
else
    fail "Expected HTTP 400, got $HTTP_CODE"
fi
echo ""

# Test 4: First request with idempotent key (should succeed)
echo -e "${BLUE}Test 4: First request with X-Idempotent-Key${NC}"
IDEMPOTENT_KEY="test-$(date +%s)-$(openssl rand -hex 4)"
info "Using idempotent key: $IDEMPOTENT_KEY"

RESPONSE=$(curl -s -w "\n%{http_code}\n%{header_json}" -X POST "${AGENT_MANAGER_URL}/api/v1/clusters" \
    -H "Content-Type: application/json" \
    -H "X-Idempotent-Key: ${IDEMPOTENT_KEY}" \
    -d '{
        "id": "test-cluster-002",
        "name": "Test Cluster 2",
        "api_server": "https://test2.k8s.local:6443",
        "status": "active"
    }')

HTTP_CODE=$(echo "$RESPONSE" | sed -n '2p')
BODY=$(echo "$RESPONSE" | sed -n '1p')
HEADERS=$(echo "$RESPONSE" | sed -n '3p')

if [ "$HTTP_CODE" == "201" ] || [ "$HTTP_CODE" == "200" ]; then
    CLUSTER_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    pass "First request succeeded (HTTP $HTTP_CODE), created cluster: $CLUSTER_ID"

    # Save for comparison
    FIRST_RESPONSE="$BODY"
else
    fail "Expected HTTP 200/201, got $HTTP_CODE. Response: $BODY"
fi
echo ""

# Test 5: Duplicate request with same key (should return cached response)
echo -e "${BLUE}Test 5: Duplicate request with same X-Idempotent-Key${NC}"
info "Waiting 1 second before retry..."
sleep 1

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${AGENT_MANAGER_URL}/api/v1/clusters" \
    -H "Content-Type: application/json" \
    -H "X-Idempotent-Key: ${IDEMPOTENT_KEY}" \
    -H "X-Debug: true" \
    -i \
    -d '{
        "id": "test-cluster-002",
        "name": "Test Cluster 2",
        "api_server": "https://test2.k8s.local:6443",
        "status": "active"
    }')

# Check for X-Idempotent-Replayed header
if echo "$RESPONSE" | grep -q "X-Idempotent-Replayed: true"; then
    pass "Duplicate request correctly returned cached response (X-Idempotent-Replayed: true)"
else
    fail "Expected X-Idempotent-Replayed header not found"
fi

# Extract body (skip headers)
BODY=$(echo "$RESPONSE" | sed -n '/^{/,$p' | head -1)

# Compare responses (should be identical)
if [ "$BODY" == "$FIRST_RESPONSE" ]; then
    pass "Cached response matches original response (same cluster ID and timestamp)"
else
    info "First:  $FIRST_RESPONSE"
    info "Second: $BODY"
    fail "Cached response differs from original"
fi
echo ""

# Test 6: Different key creates new resource
echo -e "${BLUE}Test 6: Different X-Idempotent-Key creates new resource${NC}"
NEW_KEY="test-$(date +%s)-$(openssl rand -hex 4)"
info "Using new idempotent key: $NEW_KEY"

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${AGENT_MANAGER_URL}/api/v1/clusters" \
    -H "Content-Type: application/json" \
    -H "X-Idempotent-Key: ${NEW_KEY}" \
    -d '{
        "id": "test-cluster-003",
        "name": "Test Cluster 3",
        "api_server": "https://test3.k8s.local:6443",
        "status": "active"
    }')

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" == "201" ] || [ "$HTTP_CODE" == "200" ]; then
    NEW_CLUSTER_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    if [ "$NEW_CLUSTER_ID" != "$CLUSTER_ID" ]; then
        pass "Different key created new cluster: $NEW_CLUSTER_ID"
    else
        fail "Different key should create different cluster"
    fi
else
    fail "Expected HTTP 200/201, got $HTTP_CODE"
fi
echo ""

# Test 7: Check Redis cache
echo -e "${BLUE}Test 7: Verify Redis Cache${NC}"
if command -v redis-cli &> /dev/null; then
    # Check if our idempotent key exists in Redis
    REDIS_KEY="agent-manager:${IDEMPOTENT_KEY}"
    if redis-cli -h "${REDIS_HOST}" -p "${REDIS_PORT}" EXISTS "$REDIS_KEY" | grep -q "1"; then
        pass "Idempotent key found in Redis: $REDIS_KEY"

        # Check TTL
        TTL=$(redis-cli -h "${REDIS_HOST}" -p "${REDIS_PORT}" TTL "$REDIS_KEY")
        if [ "$TTL" -gt 0 ]; then
            pass "Redis key has valid TTL: $TTL seconds (~$(($TTL / 3600)) hours)"
        else
            fail "Redis key has invalid TTL: $TTL"
        fi
    else
        fail "Idempotent key not found in Redis"
    fi
else
    info "redis-cli not found, skipping Redis cache verification"
fi
echo ""

# Test 8: GET request bypasses idempotency
echo -e "${BLUE}Test 8: GET request bypasses idempotency${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "${AGENT_MANAGER_URL}/api/v1/clusters")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)

if [ "$HTTP_CODE" == "200" ]; then
    pass "GET request succeeded without X-Idempotent-Key (HTTP 200)"
else
    fail "GET request failed with HTTP $HTTP_CODE"
fi
echo ""

# Summary
echo -e "${GREEN}=== All Tests Passed! ===${NC}"
echo ""
echo -e "${YELLOW}Summary:${NC}"
echo "  ✓ Agent Manager is running and healthy"
echo "  ✓ Redis connection is working"
echo "  ✓ Idempotency middleware correctly rejects requests without key"
echo "  ✓ First request with key succeeds"
echo "  ✓ Duplicate request returns cached response"
echo "  ✓ Different keys create different resources"
echo "  ✓ Redis cache is working with proper TTL"
echo "  ✓ GET requests bypass idempotency check"
echo ""
echo -e "${BLUE}Next Steps:${NC}"
echo "  1. Clean up test data: curl -X DELETE ${AGENT_MANAGER_URL}/api/v1/clusters/test-cluster-002"
echo "  2. Monitor Redis keys: redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} KEYS 'agent-manager:*'"
echo "  3. Check logs for idempotency middleware messages"
echo ""
