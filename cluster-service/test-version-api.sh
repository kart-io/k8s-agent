#!/bin/bash

# Version API Test Script
# Tests all 4 version endpoints and verifies responses

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8082}"
TIMEOUT=5

# Test counter
PASSED=0
FAILED=0

# Function to print section header
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

# Function to print test result
print_result() {
    local test_name="$1"
    local status="$2"

    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✓ PASS${NC} - $test_name"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC} - $test_name"
        ((FAILED++))
    fi
}

# Function to check if service is running
check_service() {
    echo -e "${YELLOW}Checking if service is running...${NC}"

    if curl -s --connect-timeout $TIMEOUT "${BASE_URL}/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Service is running at ${BASE_URL}${NC}\n"
        return 0
    else
        echo -e "${RED}✗ Service is not running at ${BASE_URL}${NC}"
        echo -e "${YELLOW}Please start the service first:${NC}"
        echo -e "  ./bin/cluster-service -config configs/config.yaml"
        echo -e "\nOr in another terminal window"
        exit 1
    fi
}

# Function to test JSON response structure
test_json_structure() {
    local endpoint="$1"
    local response="$2"
    local test_name="$3"

    # Check if response is valid JSON
    if ! echo "$response" | jq . > /dev/null 2>&1; then
        print_result "$test_name - Valid JSON" "FAIL"
        return 1
    fi

    print_result "$test_name - Valid JSON" "PASS"
    return 0
}

# Test 1: GET /version (Complete version with wrapper)
test_complete_version() {
    print_header "Test 1: GET /version (Complete Version)"

    echo "Request: GET ${BASE_URL}/version"

    response=$(curl -s -w "\nHTTP_CODE:%{http_code}" "${BASE_URL}/version")
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_CODE:/d')

    # Check HTTP status code
    if [ "$http_code" = "200" ]; then
        print_result "HTTP Status 200" "PASS"
    else
        print_result "HTTP Status 200 (got $http_code)" "FAIL"
        echo "$body"
        return
    fi

    # Validate JSON structure
    test_json_structure "/version" "$body" "Complete Version"

    # Check response structure
    if echo "$body" | jq -e '.code == 0' > /dev/null 2>&1; then
        print_result "Response code is 0" "PASS"
    else
        print_result "Response code is 0" "FAIL"
    fi

    # Check for required fields in data
    local fields=("service_name" "git_version" "git_commit" "git_branch" "build_date" "go_version" "compiler" "platform")
    for field in "${fields[@]}"; do
        if echo "$body" | jq -e ".data.$field" > /dev/null 2>&1; then
            print_result "Field '$field' exists" "PASS"
        else
            print_result "Field '$field' exists" "FAIL"
        fi
    done

    echo -e "\n${YELLOW}Sample Response:${NC}"
    echo "$body" | jq '.'
}

# Test 2: GET /version/simple (Simplified version)
test_simple_version() {
    print_header "Test 2: GET /version/simple (Simplified Version)"

    echo "Request: GET ${BASE_URL}/version/simple"

    response=$(curl -s -w "\nHTTP_CODE:%{http_code}" "${BASE_URL}/version/simple")
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_CODE:/d')

    # Check HTTP status code
    if [ "$http_code" = "200" ]; then
        print_result "HTTP Status 200" "PASS"
    else
        print_result "HTTP Status 200 (got $http_code)" "FAIL"
        echo "$body"
        return
    fi

    # Validate JSON structure
    test_json_structure "/version/simple" "$body" "Simple Version"

    # Check for service and version fields
    if echo "$body" | jq -e '.data.service' > /dev/null 2>&1; then
        print_result "Field 'service' exists" "PASS"
    else
        print_result "Field 'service' exists" "FAIL"
    fi

    if echo "$body" | jq -e '.data.version' > /dev/null 2>&1; then
        print_result "Field 'version' exists" "PASS"
    else
        print_result "Field 'version' exists" "FAIL"
    fi

    echo -e "\n${YELLOW}Sample Response:${NC}"
    echo "$body" | jq '.'
}

# Test 3: GET /version/text (Text format)
test_text_version() {
    print_header "Test 3: GET /version/text (Text Format)"

    echo "Request: GET ${BASE_URL}/version/text"

    response=$(curl -s -w "\nHTTP_CODE:%{http_code}" "${BASE_URL}/version/text")
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_CODE:/d')

    # Check HTTP status code
    if [ "$http_code" = "200" ]; then
        print_result "HTTP Status 200" "PASS"
    else
        print_result "HTTP Status 200 (got $http_code)" "FAIL"
        echo "$body"
        return
    fi

    # Check if response contains expected fields
    local fields=("serviceName" "gitVersion" "gitCommit" "gitBranch" "buildDate" "goVersion" "compiler" "platform")
    for field in "${fields[@]}"; do
        if echo "$body" | grep -q "$field"; then
            print_result "Field '$field' in text output" "PASS"
        else
            print_result "Field '$field' in text output" "FAIL"
        fi
    done

    echo -e "\n${YELLOW}Sample Response:${NC}"
    echo "$body"
}

# Test 4: GET /version/json (Raw JSON format)
test_json_version() {
    print_header "Test 4: GET /version/json (Raw JSON Format)"

    echo "Request: GET ${BASE_URL}/version/json"

    response=$(curl -s -w "\nHTTP_CODE:%{http_code}" "${BASE_URL}/version/json")
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_CODE:/d')

    # Check HTTP status code
    if [ "$http_code" = "200" ]; then
        print_result "HTTP Status 200" "PASS"
    else
        print_result "HTTP Status 200 (got $http_code)" "FAIL"
        echo "$body"
        return
    fi

    # Validate JSON structure (raw JSON, no wrapper)
    if ! echo "$body" | jq . > /dev/null 2>&1; then
        print_result "Valid JSON" "FAIL"
        return
    fi
    print_result "Valid JSON" "PASS"

    # Check that it's raw JSON (no 'code' or 'message' wrapper)
    if echo "$body" | jq -e '.code' > /dev/null 2>&1; then
        print_result "Raw JSON (no wrapper)" "FAIL"
    else
        print_result "Raw JSON (no wrapper)" "PASS"
    fi

    # Check for required fields directly in root
    local fields=("serviceName" "gitVersion" "gitCommit" "gitBranch" "buildDate" "goVersion" "compiler" "platform")
    for field in "${fields[@]}"; do
        if echo "$body" | jq -e ".$field" > /dev/null 2>&1; then
            print_result "Field '$field' exists" "PASS"
        else
            print_result "Field '$field' exists" "FAIL"
        fi
    done

    echo -e "\n${YELLOW}Sample Response:${NC}"
    echo "$body" | jq '.'
}

# Test 5: Version consistency check
test_version_consistency() {
    print_header "Test 5: Version Consistency Check"

    # Get versions from all endpoints
    version1=$(curl -s "${BASE_URL}/version" | jq -r '.data.git_version')
    version2=$(curl -s "${BASE_URL}/version/simple" | jq -r '.data.version')
    version3=$(curl -s "${BASE_URL}/version/text" | grep "gitVersion:" | awk '{print $2}')
    version4=$(curl -s "${BASE_URL}/version/json" | jq -r '.gitVersion')

    echo "Version from /version:        $version1"
    echo "Version from /version/simple: $version2"
    echo "Version from /version/text:   $version3"
    echo "Version from /version/json:   $version4"

    # Check consistency
    if [ "$version1" = "$version2" ] && [ "$version2" = "$version3" ] && [ "$version3" = "$version4" ]; then
        print_result "Version consistency across all endpoints" "PASS"
    else
        print_result "Version consistency across all endpoints" "FAIL"
    fi
}

# Test 6: Response time check
test_response_time() {
    print_header "Test 6: Response Time Check"

    local endpoints=("/version" "/version/simple" "/version/text" "/version/json")

    for endpoint in "${endpoints[@]}"; do
        local url="${BASE_URL}${endpoint}"
        local time_total=$(curl -s -o /dev/null -w "%{time_total}" "$url")
        local time_ms=$(echo "$time_total * 1000" | bc | cut -d. -f1)

        echo "Endpoint: $endpoint - ${time_ms}ms"

        # Check if response time is under 100ms
        if [ "$time_ms" -lt 100 ]; then
            print_result "$endpoint response time < 100ms" "PASS"
        else
            print_result "$endpoint response time < 100ms (${time_ms}ms)" "FAIL"
        fi
    done
}

# Main execution
main() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════╗"
    echo "║                                                  ║"
    echo "║       Version API Test Suite                    ║"
    echo "║       cluster-service                           ║"
    echo "║                                                  ║"
    echo "╚══════════════════════════════════════════════════╝"
    echo -e "${NC}"

    echo -e "${YELLOW}Configuration:${NC}"
    echo -e "  Base URL: ${BASE_URL}"
    echo -e "  Timeout: ${TIMEOUT}s"

    # Check if service is running
    check_service

    # Run all tests
    test_complete_version
    test_simple_version
    test_text_version
    test_json_version
    test_version_consistency
    test_response_time

    # Print summary
    print_header "Test Summary"

    local total=$((PASSED + FAILED))
    local pass_rate=0

    if [ "$total" -gt 0 ]; then
        pass_rate=$(echo "scale=2; $PASSED * 100 / $total" | bc)
    fi

    echo -e "${GREEN}Passed: $PASSED${NC}"
    echo -e "${RED}Failed: $FAILED${NC}"
    echo -e "Total: $total"
    echo -e "Pass Rate: ${pass_rate}%"

    if [ "$FAILED" -eq 0 ]; then
        echo -e "\n${GREEN}╔══════════════════════════════════════════╗${NC}"
        echo -e "${GREEN}║  ✓ All tests passed successfully!       ║${NC}"
        echo -e "${GREEN}╚══════════════════════════════════════════╝${NC}\n"
        exit 0
    else
        echo -e "\n${RED}╔══════════════════════════════════════════╗${NC}"
        echo -e "${RED}║  ✗ Some tests failed                     ║${NC}"
        echo -e "${RED}╚══════════════════════════════════════════╝${NC}\n"
        exit 1
    fi
}

# Run main function
main
