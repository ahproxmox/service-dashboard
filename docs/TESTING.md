# Testing Guide

This guide explains how to test the Service Dashboard locally and in CI/CD pipelines.

## Unit Tests

### Backend

```bash
cd backend
go test ./... -v
go test ./... -cover
```

Tests are located in each package directory with `_test.go` suffix:
- `config/config_test.go`
- `api/handlers_test.go`
- `discovery/proxmox_test.go`
- `discovery/caddy_test.go`
- `discovery/matcher_test.go`
- `metrics/prometheus_test.go`

### Frontend

```bash
cd frontend
npm run build
```

Frontend has no unit tests configured yet. Future: Jest + Vue Test Utils.

## Integration Tests

### API Tests

Test the backend API endpoints without requiring the frontend:

```bash
# Test against localhost
./tests/api-test.sh

# Test against remote server
./tests/api-test.sh http://dashboard.example.com:8080
```

Tests covered:
- Health check endpoint
- Services list endpoint
- Response schema validation
- Response time performance

### End-to-End Tests

Test both backend and frontend together:

```bash
# Start both services first
cd backend && go run main.go &
cd frontend && npm run dev &

# Then run e2e tests
./tests/e2e-test.sh

# Optional: specify custom URLs
./tests/e2e-test.sh http://localhost:8080 http://localhost:5173
```

Tests covered:
- Backend API availability
- Frontend HTML loads
- Service worker registration
- PWA manifest presence
- Static assets served
- API integration

## Manual Testing

### Local Development

1. **Start backend**
   ```bash
   cd backend
   go run main.go
   ```
   - Accessible at http://localhost:8080
   - Health check: http://localhost:8080/health
   - API: http://localhost:8080/api/services

2. **Start frontend**
   ```bash
   cd frontend
   npm run dev
   ```
   - Accessible at http://localhost:5173
   - Hot reload enabled
   - API proxy to http://localhost:8080

3. **Test in browser**
   - Open http://localhost:5173
   - Open DevTools (F12)
   - Check Console for errors
   - Check Network tab for API calls
   - Check Application tab for Service Worker and Cache Storage

### Docker Testing

1. **Build images**
   ```bash
   ./build.sh test-build
   ```

2. **Run with docker-compose**
   ```bash
   docker-compose up -d
   ```

3. **Test containers**
   ```bash
   curl http://localhost:5173      # Frontend
   curl http://localhost:8080/health # Backend health
   curl http://localhost:8080/api/services # Services
   ```

4. **Cleanup**
   ```bash
   docker-compose down
   ```

## CI/CD Testing

GitHub Actions automatically runs:

1. **On every push:**
   - Backend tests: `go test ./...`
   - Frontend build: `npm run build`
   - Linting: `golangci-lint`

2. **On push to main:**
   - Docker image builds

3. **On version tags (v*.*.*):**
   - Deployment workflow triggers

View results in GitHub Actions tab of repository.

## Performance Testing

### Backend Response Time

API should respond in < 1 second for typical requests:

```bash
time curl http://localhost:8080/api/services
```

### Frontend Load Time

Frontend should load and be interactive in < 2 seconds:

```bash
# Measure from browser DevTools
# Lighthouse audits available in DevTools
```

### Build Time

Expected times:
- Backend: < 10 seconds
- Frontend: 30-60 seconds
- Docker images: 2-5 minutes

## Offline Testing

### With Service Worker

1. Open http://localhost:5173 in browser
2. Open DevTools → Application → Service Workers
3. Check "Offline" checkbox
4. Reload page
5. App should still display with cached UI
6. API calls show offline message or cached data

### Browser Network Throttling

1. Open DevTools → Network tab
2. Select "Slow 3G" or "Offline"
3. Reload page
4. App should handle slow/missing network gracefully

## Browser Compatibility

Test on:
- Chrome/Chromium (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)
- Mobile browsers (iOS Safari, Chrome Android)

## Checklist

Before releasing a new version:

- [ ] All backend tests pass (`go test ./...`)
- [ ] Frontend builds without errors (`npm run build`)
- [ ] No console errors in browser
- [ ] API endpoints respond correctly
- [ ] Service worker registers
- [ ] PWA installable
- [ ] Offline mode works
- [ ] Mobile layout responsive
- [ ] Icons load correctly
- [ ] Docker images build successfully
- [ ] All GitHub Actions workflows pass

## Troubleshooting

### Backend won't start
- Check port 8080 is not in use: `lsof -i :8080`
- Check config.yaml exists and is valid YAML
- Check Proxmox/Caddy URLs are reachable

### Frontend won't build
- Delete node_modules: `rm -rf node_modules`
- Reinstall: `npm install`
- Check Node.js version: `node --version` (should be 18+)

### API returns 500 errors
- Check backend logs for specific error
- Verify Proxmox/Caddy/Prometheus are configured correctly
- Check network connectivity to those services

### Service worker not caching
- Check DevTools → Application → Cache Storage
- Verify manifest.json is present
- Check service-worker.js has no syntax errors

## Performance Profiling

### Backend

```bash
# With pprof
go run -cpuprofile=cpu.prof main.go
go tool pprof cpu.prof
```

### Frontend

Use Chrome DevTools:
1. DevTools → Performance tab
2. Record interactions
3. Analyze flame graph

## Security Testing

- [ ] No credentials in code or config files
- [ ] API validates input (if applicable)
- [ ] HTTPS enforced in production
- [ ] Service worker handles auth tokens securely
- [ ] Environment variables not exposed
