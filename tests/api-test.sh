#!/bin/bash

# API Endpoint Tests for Service Dashboard
# Tests the backend API without requiring frontend

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

BACKEND_URL="${1:-http://localhost:8080}"
PASSED=0
FAILED=0

echo -e "${YELLOW}🧪 Service Dashboard API Tests${NC}"
echo "Backend URL: $BACKEND_URL"
echo ""

# Test 1: Health check
echo -n "Test 1: GET /health ... "
RESPONSE=$(curl -s -w "\n%{http_code}" $BACKEND_URL/health)
HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}PASS${NC}"
    ((PASSED++))
else
    echo -e "${RED}FAIL${NC} (HTTP $HTTP_CODE)"
    ((FAILED++))
fi

# Test 2: Get services
echo -n "Test 2: GET /api/services ... "
RESPONSE=$(curl -s -w "\n%{http_code}" $BACKEND_URL/api/services)
HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "200" ]; then
    # Check if response has required fields
    if echo "$BODY" | grep -q '"services"' && echo "$BODY" | grep -q '"timestamp"'; then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED++))
    else
        echo -e "${RED}FAIL${NC} (invalid response format)"
        ((FAILED++))
    fi
else
    echo -e "${RED}FAIL${NC} (HTTP $HTTP_CODE)"
    ((FAILED++))
fi

# Test 3: Verify services response structure
echo -n "Test 3: Verify services schema ... "
RESPONSE=$(curl -s $BACKEND_URL/api/services)

if echo "$RESPONSE" | grep -q '"id"' && \
   echo "$RESPONSE" | grep -q '"name"' && \
   echo "$RESPONSE" | grep -q '"status"' && \
   echo "$RESPONSE" | grep -q '"timestamp"'; then
    echo -e "${GREEN}PASS${NC}"
    ((PASSED++))
else
    echo -e "${RED}FAIL${NC} (missing required fields)"
    ((FAILED++))
fi

# Test 4: Verify service count
echo -n "Test 4: Check for services in response ... "
RESPONSE=$(curl -s $BACKEND_URL/api/services)
SERVICE_COUNT=$(echo "$RESPONSE" | grep -o '"id"' | wc -l)

if [ "$SERVICE_COUNT" -gt 0 ]; then
    echo -e "${GREEN}PASS${NC} ($SERVICE_COUNT services found)"
    ((PASSED++))
else
    echo -e "${RED}FAIL${NC} (no services found)"
    echo "This is expected if Proxmox/Caddy are not configured"
    # Don't count as failure for this test
fi

# Test 5: Response performance
echo -n "Test 5: Response time < 5s ... "
START=$(date +%s%N)
curl -s $BACKEND_URL/api/services > /dev/null
END=$(date +%s%N)
DURATION=$((($END - $START) / 1000000))

if [ $DURATION -lt 5000 ]; then
    echo -e "${GREEN}PASS${NC} (${DURATION}ms)"
    ((PASSED++))
else
    echo -e "${RED}FAIL${NC} (${DURATION}ms - too slow)"
    ((FAILED++))
fi

echo ""
echo "Results: ${GREEN}$PASSED passed${NC}, ${RED}$FAILED failed${NC}"

if [ $FAILED -gt 0 ]; then
    exit 1
fi

exit 0
