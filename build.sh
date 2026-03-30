#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Service Dashboard - Docker Build${NC}"

# Check Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker is not installed${NC}"
    exit 1
fi

# Get version from git tag or use timestamp
VERSION=${1:-$(date +%Y%m%d-%H%M%S)}

echo -e "${YELLOW}Building Service Dashboard v${VERSION}${NC}"

# Build backend
echo -e "${YELLOW}Building backend...${NC}"
docker build -t ahproxmox/service-dashboard-backend:${VERSION} \
             -t ahproxmox/service-dashboard-backend:latest \
             ./backend

if [ $? -eq 0 ]; then
    echo -e "${GREEN}Backend built successfully${NC}"
else
    echo -e "${RED}Backend build failed${NC}"
    exit 1
fi

# Build frontend
echo -e "${YELLOW}Building frontend...${NC}"
docker build -t ahproxmox/service-dashboard-frontend:${VERSION} \
             -t ahproxmox/service-dashboard-frontend:latest \
             ./frontend

if [ $? -eq 0 ]; then
    echo -e "${GREEN}Frontend built successfully${NC}"
else
    echo -e "${RED}Frontend build failed${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}All containers built successfully!${NC}"
echo ""
echo "Image tags:"
echo "  - ahproxmox/service-dashboard-backend:${VERSION}"
echo "  - ahproxmox/service-dashboard-backend:latest"
echo "  - ahproxmox/service-dashboard-frontend:${VERSION}"
echo "  - ahproxmox/service-dashboard-frontend:latest"
echo ""
echo "To run locally:"
echo "  docker-compose up -d"
echo ""
echo "To push to registry:"
echo "  docker push ahproxmox/service-dashboard-backend:${VERSION}"
echo "  docker push ahproxmox/service-dashboard-frontend:${VERSION}"
