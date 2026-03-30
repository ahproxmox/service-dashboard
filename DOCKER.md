# Docker Configuration & Build Guide

## Overview

This project includes complete Docker support with:
- Multi-stage builds for both backend (Go) and frontend (Vue.js)
- Docker Compose for local development
- Optimized build script with version tagging
- Health checks and proper networking

## Files Created

### Backend
- `backend/Dockerfile` - Multi-stage Go build (golang:1.21 → alpine:3.18)
- `backend/.dockerignore` - Excludes unnecessary files from context

### Frontend
- `frontend/Dockerfile` - Multi-stage Node build (node:18 → nginx:alpine)
- `frontend/.dockerignore` - Excludes unnecessary files from context
- `frontend/nginx.conf` - Production nginx configuration with SPA routing

### Root Directory
- `docker-compose.yml` - Local development orchestration
- `build.sh` - Automated build script

## Quick Start

### Build Images
```bash
# Build with timestamp (auto-versioned)
./build.sh

# Build with custom version
./build.sh v1.0.0
```

### Run Locally
```bash
# Start both services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## Build Details

### Backend (Go)
- **Base Image**: golang:1.21-alpine (build), alpine:3.18 (runtime)
- **Build Process**: 
  - Download Go dependencies
  - Build static binary with `-installsuffix cgo`
  - Strip CGO for maximum portability
- **Port**: 8080
- **Config**: Mounted at `/etc/service-dashboard/config.yaml`
- **Health Check**: GET /health endpoint

### Frontend (Vue.js)
- **Base Image**: node:18-alpine (build), nginx:alpine (runtime)
- **Build Process**:
  - Install npm dependencies
  - Run Vite build
  - Serve with nginx
- **Port**: 80
- **SPA Routing**: All non-asset routes redirect to index.html
- **Caching**: 
  - Assets: 30 days (immutable)
  - HTML: No cache (must revalidate)
  - Service worker: No cache
- **Compression**: gzip enabled

## Image Tags

Built images follow this naming convention:
```
ahproxmox/service-dashboard-backend:[VERSION]
ahproxmox/service-dashboard-backend:latest
ahproxmox/service-dashboard-frontend:[VERSION]
ahproxmox/service-dashboard-frontend:latest
```

## Configuration

### Backend Environment Variables
- `CONFIG_PATH`: Path to config.yaml (default: /etc/service-dashboard/config.yaml)
- `LOG_LEVEL`: Logging level (optional)

### Config File (config.yaml)
See `container/config.yaml.example` for template. Includes:
- Server port
- Proxmox API credentials
- Caddy API endpoint
- Prometheus URL
- Cache TTLs

## Docker Compose Details

### Services
- **backend**: Go HTTP server on port 8080
- **frontend**: nginx on port 5173 (mapped from 80)

### Networking
- Both services share `service-dashboard` bridge network
- Backend accessible internally as `http://backend:8080`
- Frontend accessible at `http://localhost:5173`

### Health Checks
- Both services have health checks enabled
- 5 second startup grace period
- 30 second check interval
- 10 second timeout
- 3 retry attempts

## Production Deployment

### Image Size
- Backend: ~15-20 MB (static binary + ca-certificates)
- Frontend: ~10-15 MB (nginx + compiled assets)

### Registry Push
```bash
# Login to registry
docker login

# Push images
docker push ahproxmox/service-dashboard-backend:v1.0.0
docker push ahproxmox/service-dashboard-frontend:v1.0.0
```

### Running on Proxmox
Images can be deployed to LXC containers or used with Docker on any container orchestration platform.

## Troubleshooting

### Build Fails
- Ensure Docker daemon is running
- Check disk space for intermediate layers
- Verify Go modules are accessible (internet connectivity)

### Frontend Not Loading
- Verify nginx.conf is properly mounted
- Check browser console for API endpoint errors
- Ensure backend is healthy

### Health Check Failures
- Backend: Verify service is listening on :8080
- Frontend: Check nginx is running and responsive

## Notes

- Multi-stage builds optimize final image sizes
- Static Go binary ensures portability
- Nginx serves as both web server and reverse proxy for SPA
- Config mounted as read-only volume
- Images designed for stateless deployment
