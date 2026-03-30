# Service Dashboard Design — Auto-Discovery PWA

**Date:** 2026-03-30
**Spec ID:** service-dashboard-001
**Status:** Design Phase
**Repo:** `ahproxmox/service-dashboard`

---

## 1. Overview

A Progressive Web App that auto-discovers all running Proxmox containers, displays their status with visual indicators, retrieves metrics from Prometheus, and allows users to open service URLs directly from the dashboard.

**Key Requirements:**
- Auto-discover containers from Proxmox + Caddy (zero manual config)
- Display status with green glow (up) / grey (stopped) icons
- Show CPU, RAM, disk, network metrics per container
- Clickable links to HTTPS URLs for each service
- Installable PWA with service worker caching for instant load
- Backend polls Proxmox/Caddy/Prometheus; frontend polls backend

---

## 2. Architecture

### 2.1 System Components

```
┌─────────────────────────────────────────────────────────┐
│ User Browser                                            │
│ ┌────────────────────────────────────────────────────┐ │
│ │ Vue 3 PWA Frontend                                 │ │
│ │ - ServiceCard grid (icon, status, metrics, URL)    │ │
│ │ - Polling: status every 1-2s, metrics every 30s    │ │
│ │ - Service worker (asset caching)                   │ │
│ │ - Manifest (installable)                           │ │
│ └────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
                        ↑ HTTP polling
                        ↓ JSON responses
┌─────────────────────────────────────────────────────────┐
│ CT 127 — Backend (Go)                                   │
│ ┌────────────────────────────────────────────────────┐ │
│ │ REST API (:8080)                                   │ │
│ │ - GET /api/services                                │ │
│ │ - GET /health                                      │ │
│ │ - Serves static frontend files (/public)           │ │
│ └────────────────────────────────────────────────────┘ │
│                                                         │
│ Service Discovery & Metrics Engine                     │
│ ├─ Query Proxmox API (running containers)             │
│ ├─ Query Caddy API (routes → HTTPS URLs)              │
│ ├─ Intelligent matching (IP → hostname → port)        │
│ ├─ Query Prometheus (CPU, RAM, disk, network)         │
│ └─ Cache results (2s for status, 25s for metrics)     │
└─────────────────────────────────────────────────────────┘
         ↑ API calls (every 1-30 seconds)
         ↓ Responses
┌──────────────┬──────────────┬──────────────────────┐
│ Proxmox      │ Caddy        │ Prometheus           │
│ /api2/json/  │ /admin/api/  │ :9090/api/v1/query   │
│ nodes/pve/lxc│ config       │                      │
└──────────────┴──────────────┴──────────────────────┘
```

### 2.2 Deployment

**Container:** CT 127 (new)
- **Base Image:** Debian minimal
- **Port:** 8080 (internal, proxied by Caddy)
- **Caddy Route:** `dashboard.internal.ahproxmox-claude.cc` → CT 127:8080
- **Binary Location:** `/opt/dashboard-backend`
- **Config:** `/etc/dashboard/config.yaml` (Proxmox token, Caddy API URL, Prometheus URL)

**Frontend:** Served by backend's Go HTTP server from `/public` directory
- Static files cached by service worker
- Manifest allows "Install" button in browser

---

## 3. Backend (Go)

### 3.1 Core Responsibilities

1. **Service Discovery** (cached, refreshed periodically)
   - Query Proxmox: Get all running containers (id, name, status, IP)
   - Query Caddy: Get all routes and their backend targets
   - Match containers to routes:
     - **Primary:** IP-based matching (most robust)
     - **Fallback:** Hostname matching (container hostname in route domain)
     - **Result:** Container → HTTPS URL mapping
   - Handle services with no Caddy route gracefully (mark as "not proxied")

2. **Metrics Aggregation** (Prometheus queries)
   - For each discovered service, fetch metrics:
     - `node_cpu_seconds_total` → CPU %
     - `node_memory_bytes_total` and `node_memory_bytes_available` → RAM MB and %
     - `node_filesystem_avail_bytes` → Disk %
     - `node_network_receive_bytes_total` and `node_network_transmit_bytes_total` → Network in/out Mbps
   - Aggregate by container ID
   - Cache results (25 seconds)

3. **REST API Endpoints**
   - `GET /api/services` — Returns all services with current status and metrics
   - `GET /health` — Backend health check (Proxmox/Caddy/Prometheus connectivity)
   - Serves static files from `/public` for frontend

### 3.2 API Response Format

```json
{
  "services": [
    {
      "id": "122",
      "name": "kanban",
      "status": "up",
      "https_url": "https://kanban.internal.ahproxmox-claude.cc",
      "metrics": {
        "cpu_percent": 2.5,
        "ram_mb": 512,
        "ram_percent": 25,
        "disk_percent": 60,
        "network_in_mbps": 0.1,
        "network_out_mbps": 0.2
      }
    },
    {
      "id": "111",
      "name": "rag",
      "status": "up",
      "https_url": "https://rag.internal.ahproxmox-claude.cc",
      "metrics": {
        "cpu_percent": 5.0,
        "ram_mb": 1024,
        "ram_percent": 50,
        "disk_percent": 45,
        "network_in_mbps": 0.3,
        "network_out_mbps": 0.1
      }
    },
    {
      "id": "104",
      "name": "openclaw",
      "status": "stopped",
      "https_url": null,
      "metrics": null
    }
  ],
  "timestamp": 1711900000
}
```

### 3.3 Caching Strategy

| Data | Refresh Interval | Cache Duration | Rationale |
|------|------------------|-----------------|-----------|
| Container status | Every 10-30s | 2 seconds | Detect container state changes quickly |
| Metrics | Every 30s | 25 seconds | Low overhead on Prometheus, metrics don't need to be faster |
| Service discovery | Every 10-30s | 10 seconds | Handle new Caddy routes with slight delay |

**Implementation:** Backend maintains a cache with timestamps. On `/api/services` call, if cache is fresh, return immediately. If stale, query all three APIs (Proxmox, Caddy, Prometheus) in parallel, update cache, return.

### 3.4 Error Handling

| Scenario | Behavior |
|----------|----------|
| Proxmox API unreachable | Return last-known container list; log error |
| Caddy API unreachable | Return containers without HTTPS URLs; log error |
| Prometheus unreachable | Return containers with `metrics: null`; log error |
| Container not in Prometheus | Metrics omitted for that container |
| Multiple errors | Return whatever data is available; `/health` endpoint exposes issues |

---

## 4. Frontend (Vue 3 PWA)

### 4.1 Structure

```
frontend/
├── src/
│   ├── App.vue
│   ├── components/
│   │   ├── ServiceCard.vue      # Collapsible card: icon, status, metrics, URL
│   │   ├── ServiceGrid.vue      # Grid layout of all service cards
│   │   └── ErrorBanner.vue      # Connection errors / status
│   ├── services/
│   │   └── api.js               # Polling logic, backend communication
│   ├── config/
│   │   └── services-config.json # Icon mappings: service name → icon name
│   ├── utils/
│   │   └── metrics.js           # Formatting helpers (bytes → MB, etc.)
│   └── main.js
├── public/
│   ├── manifest.json            # PWA manifest
│   ├── service-worker.js        # Asset caching
│   ├── icons/                   # Icon SVG files
│   │   ├── kanban.svg
│   │   ├── rag.svg
│   │   └── ...
│   └── favicon.ico
└── package.json
```

### 4.2 ServiceCard Component

**Collapsed state (default):**
```
┌──────────────────┐
│ 🎯               │  ← Icon (from services-config.json)
│ Kanban           │  ← Service name
│ ✓ UP             │  ← Status indicator (green glow / grey)
└──────────────────┘
```

**Expanded state (on click/tap):**
```
┌──────────────────────────────┐
│ 🎯 Kanban                    │
│ Status: ✓ UP                 │
│                              │
│ https://kanban.internal...   │  ← Clickable link
│ [Open in new tab →]          │
│                              │
│ Metrics:                     │
│ CPU:  2.5%    RAM: 25%       │
│ Disk: 60%     Net: ↓0.1 ↑0.2 │
│                              │
│ Last updated: 2s ago         │
└──────────────────────────────┘
```

**Interactions:**
- Click icon/name → expand/collapse card
- Click HTTPS URL or "Open" button → open URL in new tab
- If no HTTPS URL: show "Not proxied", disable link
- If metrics unavailable: show "—" or "loading..."

### 4.3 Polling Logic

```javascript
// Pseudocode
const pollIntervalStatus = 1500;  // 1.5 seconds (status)
const pollIntervalMetrics = 30000; // 30 seconds (metrics)

let lastMetricsTime = 0;

setInterval(async () => {
  const services = await fetch('/api/services').then(r => r.json());

  const now = Date.now();
  if (now - lastMetricsTime >= pollIntervalMetrics) {
    // Update metrics (full refresh)
    updateAllMetrics(services);
    lastMetricsTime = now;
  } else {
    // Update status only (fast refresh)
    updateAllStatus(services);
  }
}, pollIntervalStatus);
```

**Result:** Efficient polling — status updates every 1.5s, metrics every 30s, single API call per interval.

### 4.4 PWA Features

**Service Worker (asset caching):**
- Cache all static assets (HTML, CSS, JS, icons) on first load
- Strategy: Cache-first for assets, network-first for `/api/services`
- Users return to instant-load dashboard (offline-first for UI, online-first for data)

**Manifest (`public/manifest.json`):**
```json
{
  "name": "Service Dashboard",
  "short_name": "Dashboard",
  "description": "Auto-discovery service dashboard with metrics",
  "start_url": "/",
  "scope": "/",
  "display": "standalone",
  "background_color": "#ffffff",
  "theme_color": "#0066cc",
  "icons": [
    {
      "src": "/icons/icon-192.png",
      "sizes": "192x192",
      "type": "image/png"
    },
    {
      "src": "/icons/icon-512.png",
      "sizes": "512x512",
      "type": "image/png"
    }
  ]
}
```

**Result:** Users can "Install" dashboard as standalone app, add to home screen.

### 4.5 Service Icon Mapping (`services-config.json`)

Frontend looks up icons by service name. No backend involvement.

```json
{
  "services": {
    "kanban": {
      "icon": "kanban",
      "display_name": "Kanban Board"
    },
    "rag": {
      "icon": "brain",
      "display_name": "RAG"
    },
    "openclaw": {
      "icon": "robot",
      "display_name": "OpenClaw"
    }
  }
}
```

**Icon files:** Store as SVGs in `public/icons/` (e.g., `kanban.svg`, `brain.svg`).

**Future Figma integration:** Designers can update `services-config.json` and add new SVGs without touching code.

### 4.6 Error Handling

| Scenario | UI Behavior |
|----------|-------------|
| Backend unreachable | Show error banner; display last-known services (stale) |
| Metrics unavailable | Show "—" in metric fields |
| Service URL missing | Disable link, show "Not proxied" |
| Service stopped | Grey icon, "STOPPED" status |

---

## 5. Data Flow & Timing

### 5.1 Full Request Cycle

```
Time  | Frontend          | Backend              | External
──────┼──────────────────┼──────────────────────┼─────────────
t=0   | Poll /api/services
      |                  | Check cache (fresh?) │
      |                  | Yes: return cached   │
      | ← 50ms response  |                      │
      | Update status    |                      │
      |                  |                      │
t=1.5s| Poll /api/services
      |                  | (same as above)      │
      | ← 50ms response  |                      │
      | Update status    |                      │
      |                  |                      │
t=30s | Poll /api/services
      |                  | Check cache stale    │
      |                  | ├─ Query Proxmox     │ ← 200ms
      |                  | ├─ Query Caddy       │ ← 100ms
      |                  | ├─ Query Prometheus  │ ← 500ms (parallel)
      |                  | └─ Aggregate         │ ← 50ms
      | ← 200ms response |                      │
      | Update status    |                      │
      | Update metrics   |                      │
```

**Key:** Backend queries in parallel; frontend benefits from cache hits (50ms) and full refreshes (200ms).

---

## 6. Service Discovery: Intelligent Matching

### 6.1 Algorithm

**Input:**
- From Proxmox: Container ID, name, IP (e.g., 122, "kanban", 192.168.88.78)
- From Caddy: Routes and their backend targets (e.g., "kanban.internal...", 192.168.88.78:3000)

**Algorithm:**
```
For each container:
  1. Get container IP
  2. Search Caddy routes for a route that targets this IP
     → If found: use that route's domain as HTTPS URL ✓
  3. If not found, check container hostname
     → If container hostname appears in any route domain: use that route ✓
  4. If still no match: container is "undiscovered" (no HTTPS URL)
  5. Add container to services list

Result: {id, name, status, https_url}
```

**Robustness:**
- IP matching catches 95% of cases (most reliable)
- Hostname matching catches edge cases (hostnames in route domains)
- Graceful degradation: services without HTTPS URLs still show status/metrics
- Zero manual config needed; auto-discovers as services are added to Caddy

---

## 7. Configuration & Deployment

### 7.1 Container Setup (CT 127)

**Dockerfile / build process:**
- Base: `golang:1.23-alpine` (build stage)
- Build binary from repo
- Runtime: `alpine:latest` (minimal)
- Copy binary, CA certs, config template

**systemd service:**
```ini
[Unit]
Description=Service Dashboard Backend
After=network.target

[Service]
Type=simple
ExecStart=/opt/dashboard-backend
WorkingDirectory=/etc/dashboard
Environment="CONFIG=/etc/dashboard/config.yaml"
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### 7.2 Configuration (`/etc/dashboard/config.yaml`)

```yaml
server:
  port: 8080

proxmox:
  api_url: https://192.168.88.5:8006
  token_id: "user!token_name"
  token_secret: "${PROXMOX_TOKEN}"  # From Vault

caddy:
  api_url: http://192.168.88.82:2019

prometheus:
  url: http://192.168.88.73:9090

cache:
  status_ttl: 2s
  metrics_ttl: 25s
  discovery_ttl: 10s
```

### 7.3 Caddy Route

```caddy
dashboard.internal.ahproxmox-claude.cc {
  reverse_proxy 192.168.88.127:8080
}
```

---

## 8. Testing Strategy

### 8.1 Backend Tests

| Test | Scope |
|------|-------|
| **Unit:** Prometheus metric parsing | Parse CPU/RAM/disk/network correctly |
| **Unit:** Service matching logic | IP matching, hostname fallback |
| **Integration:** Proxmox API mock | Correct container discovery |
| **Integration:** Caddy API mock | Correct route parsing |
| **Integration:** Caching behavior | Cache hits, misses, expiry |
| **E2E:** Full poll cycle | Correct response format, metrics present |

### 8.2 Frontend Tests

| Test | Scope |
|------|-------|
| **Unit:** services-config parsing | Icons loaded correctly |
| **Unit:** Metric formatting | Bytes → MB, % formatting |
| **Component:** ServiceCard | Expand/collapse, link click |
| **Integration:** Polling logic | Status every 1.5s, metrics every 30s |
| **E2E:** PWA functionality | Service worker caching, installable |

---

## 9. Success Criteria

- [ ] Dashboard auto-discovers all running containers without manual config
- [ ] Service icons display correctly with status indicators (green/grey)
- [ ] Status updates within 2 seconds of container state change
- [ ] Metrics (CPU, RAM, disk, network) update every 30 seconds
- [ ] Clicking URL opens service in new tab
- [ ] Dashboard installable on mobile/desktop
- [ ] Service worker caches assets; subsequent loads are instant
- [ ] Handles missing HTTPS URLs, stopped containers, and API errors gracefully
- [ ] Repo structure supports future Figma design integration

---

## 10. Future Extensions (Out of Scope)

- Real-time WebSocket updates (currently polling)
- Container action controls (restart, stop, etc.)
- Custom dashboard themes
- Multi-user dashboard with auth
- Figma design system integration (planned, separate phase)

---

## Repo Structure

```
ahproxmox/service-dashboard/
├── backend/
│   ├── main.go
│   ├── api/
│   │   ├── handlers.go
│   │   └── responses.go
│   ├── discovery/
│   │   ├── proxmox.go
│   │   ├── caddy.go
│   │   └── matcher.go
│   ├── metrics/
│   │   └── prometheus.go
│   ├── cache/
│   │   └── cache.go
│   ├── config/
│   │   └── config.go
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── App.vue
│   │   ├── components/
│   │   ├── services/
│   │   ├── config/
│   │   └── main.js
│   ├── public/
│   │   ├── manifest.json
│   │   ├── service-worker.js
│   │   └── icons/
│   ├── package.json
│   └── vite.config.js
├── docs/
│   └── superpowers/
│       └── specs/
│           └── 2026-03-30-service-dashboard-design.md (this file)
├── container/
│   ├── Dockerfile
│   ├── build.sh
│   └── config.yaml.example
├── .github/
│   └── workflows/
│       └── build.yml
├── README.md
└── .gitignore
```

---

## Sign-Off

**Design approved by:** Angelo (2026-03-30)
**Ready for implementation:** Yes

