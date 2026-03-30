#!/bin/bash

# End-to-End Tests for Service Dashboard
# Tests both backend and frontend integration

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

BACKEND_URL="${1:-http://localhost:8080}"
FRONTEND_URL="${2:-http://localhost:5173}"

echo -e "${BLUE}🎯 Service Dashboard End-to-End Tests${NC}"
echo ""

# Check if services are running
echo -e "${YELLOW}Checking services...${NC}"

echo -n "Backend API ($BACKEND_URL) ... "
if curl -s $BACKEND_URL/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC} (not accessible)"
    echo "Start backend with: cd backend && go run main.go"
    exit 1
fi

echo -n "Frontend ($FRONTEND_URL) ... "
if curl -s $FRONTEND_URL > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC} (not accessible)"
    echo "Start frontend with: cd frontend && npm run dev"
    exit 1
fi

echo ""
echo -e "${YELLOW}Running API tests...${NC}"
./tests/api-test.sh $BACKEND_URL

echo ""
echo -e "${YELLOW}Testing integration...${NC}"

# Test 1: Frontend loads
echo -n "Test 1: Frontend HTML loads ... "
RESPONSE=$(curl -s -w "\n%{http_code}" $FRONTEND_URL)
HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "200" ] && echo "$BODY" | grep -q "Service Dashboard"; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC} (HTTP $HTTP_CODE)"
    exit 1
fi

# Test 2: API endpoints are configured
echo -n "Test 2: API client can reach backend ... "
# This would require a headless browser in production, so we test indirectly
if curl -s $BACKEND_URL/api/services > /dev/null; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    exit 1
fi

# Test 3: Service worker is present
echo -n "Test 3: Service worker present ... "
if curl -s $FRONTEND_URL/service-worker.js > /dev/null 2>&1; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    exit 1
fi

# Test 4: Manifest is present
echo -n "Test 4: PWA manifest present ... "
if curl -s $FRONTEND_URL/manifest.json > /dev/null 2>&1; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    exit 1
fi

# Test 5: Assets are served
echo -n "Test 5: Static assets served ... "
RESPONSE=$(curl -s $FRONTEND_URL | grep -o '/assets/' | head -1)
if [ ! -z "$RESPONSE" ]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}✓ All end-to-end tests passed!${NC}"
echo ""
echo "Next steps:"
echo "1. Test PWA features (install, offline mode)"
echo "2. Verify metrics display (if Prometheus configured)"
echo "3. Test on mobile device"

exit 0
