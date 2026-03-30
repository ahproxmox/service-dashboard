# Service Dashboard

A Progressive Web App (PWA) for monitoring and managing services running on a Proxmox home lab. Auto-discovers containers, displays real-time status with visual indicators, and shows Prometheus metrics.

## Features

- **Auto-Discovery:** Automatically discovers running containers from Proxmox API
- **Status Indicators:** Green glow for running, grey for stopped services
- **Real-time Metrics:** CPU, RAM, disk, and network metrics from Prometheus
- **PWA Support:** Installable as an app, works offline with cached data
- **Responsive Design:** Works on desktop, tablet, and mobile devices
- **Service Icons:** Visual representation with configurable icons and colors

## Architecture

### Backend (Go)
- RESTful API for service discovery and metrics
- Integrates with Proxmox API for container status
- Queries Caddy for HTTPS route mapping
- Retrieves metrics from Prometheus
- In-memory caching with configurable TTLs

### Frontend (Vue 3 + Vite)
- Modern SPA with component-based architecture
- Axios HTTP client with error handling
- Service worker for offline support
- Responsive grid layout with hover effects
- Auto-refresh every 30 seconds

### Deployment
- Docker containerization for both backend and frontend
- Multi-stage builds for minimal image sizes
- docker-compose for local development
- GitHub Actions CI/CD pipeline

## Quick Start

### Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/ahproxmox/service-dashboard.git
   cd service-dashboard
   ```

2. **Backend setup**
   ```bash
   cd backend
   go mod download
   # Edit config.yaml with your Proxmox/Caddy/Prometheus URLs
   go run main.go
   ```

3. **Frontend setup**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

4. **Access the app**
   - Frontend: http://localhost:5173
   - Backend API: http://localhost:8080
   - Health check: http://localhost:8080/health

### Docker Deployment

```bash
# Build containers
./build.sh v1.0.0

# Run with docker-compose
docker-compose up -d

# Access
- Frontend: http://localhost:5173
- Backend: http://localhost:8080
```

## Configuration

### Backend
Create `backend/config.yaml`:
```yaml
server:
  port: 8080

proxmox:
  api_url: https://pve:8006
  token_id: user@pam!token-id
  token_secret: token-secret

caddy:
  api_url: http://caddy:2019

prometheus:
  url: http://prometheus:9090

cache:
  status_ttl: 30s
  metrics_ttl: 30s
  discovery_ttl: 5m
```

### Frontend
Service configuration in `frontend/public/config/services-config.json`:
- Service display names
- Icon paths
- Brand colors
- Metric thresholds

## API Endpoints

- `GET /api/services` - List all services with status and metrics
- `GET /health` - Health check for all dependencies

## Technology Stack

### Backend
- Go 1.21
- Standard library HTTP
- In-memory caching
- Goroutines for concurrent requests

### Frontend
- Vue 3
- Vite 4.5
- Axios
- Service Worker
- PWA manifest

### Infrastructure
- Docker & Docker Compose
- GitHub Actions
- Nginx (frontend)
- Alpine Linux (minimal images)

## Project Structure

```
service-dashboard/
├── backend/                 # Go backend server
│   ├── main.go
│   ├── config/config.go
│   ├── api/handlers.go
│   ├── discovery/           # Container & route discovery
│   ├── metrics/             # Prometheus client
│   ├── cache/              # In-memory cache
│   └── Dockerfile
├── frontend/               # Vue 3 SPA
│   ├── src/
│   │   ├── App.vue
│   │   ├── components/
│   │   ├── utils/
│   │   └── api.js
│   ├── public/
│   │   ├── icons/         # Service SVG icons
│   │   ├── config/
│   │   └── manifest.json
│   ├── Dockerfile
│   ├── vite.config.js
│   └── package.json
├── .github/workflows/      # CI/CD pipelines
├── docker-compose.yml
├── build.sh
└── README.md
```

## CI/CD Pipeline

GitHub Actions automatically:
1. Runs Go tests on every push
2. Builds and tests frontend on every push
3. Builds Docker images on main branch
4. Lints code with golangci-lint
5. Deploys on version tags (optional)

## Development

### Running Tests
```bash
# Backend
cd backend && go test ./... -v

# Frontend
cd frontend && npm run build
```

### Building Locally
```bash
# Using build script
./build.sh v1.0.0

# Or using docker-compose
docker-compose up --build
```

## Deployment

See [DEPLOYMENT.md](docs/DEPLOYMENT.md) for detailed deployment instructions to CT 127.

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see LICENSE file for details.

## Status

- [x] Backend: Container discovery, route mapping, metrics collection
- [x] Frontend: Vue 3 SPA with responsive design
- [x] PWA: Installable with offline support
- [x] Docker: Multi-stage builds, optimized images
- [x] CI/CD: GitHub Actions pipeline
- [ ] Deployment: CT 127 setup (Task 18)

## Support

For issues, questions, or suggestions, open an issue on GitHub.

---

Built with dedication for home lab monitoring
