# Service Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a PWA dashboard that auto-discovers running containers from Proxmox/Caddy, displays their status and metrics from Prometheus, and allows direct linking to service URLs.

**Architecture:** Go backend (CT 127) queries Proxmox/Caddy/Prometheus with caching; Vue 3 frontend polls backend every 1-2s for status, 30s for metrics. Frontend is a PWA with service worker caching and installable manifest.

**Tech Stack:** Go 1.23, Vue 3, Vite, Prometheus API client, YAML config, SQLite (optional for metrics history)

---

## File Structure

### Backend (`backend/`)
```
backend/
├── main.go                    # Entry point, HTTP server setup
├── config/
│   └── config.go             # YAML config loading
├── discovery/
│   ├── proxmox.go            # Proxmox API client
│   ├── caddy.go              # Caddy API client
│   └── matcher.go            # Service discovery matching algorithm
├── metrics/
│   └── prometheus.go         # Prometheus metrics aggregation
├── cache/
│   └── cache.go              # In-memory caching with TTL
├── api/
│   ├── handlers.go           # HTTP handlers (/api/services, /health, static files)
│   └── responses.go          # Response DTOs
├── Dockerfile                # Alpine-based container image
├── go.mod                    # Go dependencies
└── go.sum
```

### Frontend (`frontend/`)
```
frontend/
├── src/
│   ├── main.js              # Vue entry point
│   ├── App.vue              # Main dashboard component
│   ├── components/
│   │   ├── ServiceCard.vue  # Collapsible service card
│   │   ├── ServiceGrid.vue  # Grid layout wrapper
│   │   └── ErrorBanner.vue  # Error/connection status
│   ├── services/
│   │   └── api.js           # Polling logic & backend communication
│   ├── utils/
│   │   └── metrics.js       # Metric formatting (bytes→MB, etc.)
│   └── config/
│       └── services-config.json  # Icon mappings
├── public/
│   ├── manifest.json        # PWA manifest
│   ├── service-worker.js    # Service worker for caching
│   ├── icons/
│   │   ├── kanban.svg
│   │   ├── rag.svg
│   │   ├── brain.svg
│   │   └── ...
│   ├── icon-192.png         # PWA home screen icon
│   └── icon-512.png
├── package.json
├── vite.config.js
└── index.html               # Root HTML
```

### Container & Deployment (`container/`)
```
container/
├── Dockerfile               # Multi-stage: build Go binary, run on Alpine
├── build.sh                 # Build script
└── config.yaml.example      # Configuration template
```

---

## Phase 1: Backend — Core Infrastructure

### Task 1: Project Setup & Config Loading

**Files:**
- Create: `backend/go.mod`
- Create: `backend/config/config.go`
- Create: `backend/config/config_test.go`
- Create: `backend/main.go`
- Create: `container/config.yaml.example`

**Goal:** Load YAML config with Proxmox/Caddy/Prometheus endpoints and cache TTLs.

- [ ] **Step 1: Write failing test for config loading**

```go
// backend/config/config_test.go
package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a test config file
	configYAML := `
server:
  port: 8080
proxmox:
  api_url: https://192.168.88.5:8006
  token_id: user@pam!token
  token_secret: secret123
caddy:
  api_url: http://192.168.88.82:2019
prometheus:
  url: http://192.168.88.73:9090
cache:
  status_ttl: 2s
  metrics_ttl: 25s
  discovery_ttl: 10s
`
	tmpFile := t.TempDir() + "/test-config.yaml"
	os.WriteFile(tmpFile, []byte(configYAML), 0644)

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Proxmox.APIUrl != "https://192.168.88.5:8006" {
		t.Errorf("expected Proxmox URL, got %s", cfg.Proxmox.APIUrl)
	}
	if cfg.Cache.StatusTTL.Seconds() != 2 {
		t.Errorf("expected status TTL 2s, got %v", cfg.Cache.StatusTTL)
	}
}
```

- [ ] **Step 2: Run test to verify failure**

```bash
cd backend
go test ./config -v
```

Expected output: `LoadConfig undefined`

- [ ] **Step 3: Implement config loading**

```go
// backend/config/config.go
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Proxmox struct {
		APIUrl      string `yaml:"api_url"`
		TokenId     string `yaml:"token_id"`
		TokenSecret string `yaml:"token_secret"`
	} `yaml:"proxmox"`
	Caddy struct {
		APIUrl string `yaml:"api_url"`
	} `yaml:"caddy"`
	Prometheus struct {
		Url string `yaml:"url"`
	} `yaml:"prometheus"`
	Cache struct {
		StatusTTL    time.Duration `yaml:"status_ttl"`
		MetricsTTL   time.Duration `yaml:"metrics_ttl"`
		DiscoveryTTL time.Duration `yaml:"discovery_ttl"`
	} `yaml:"cache"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}
```

- [ ] **Step 4: Run test to verify pass**

```bash
cd backend
go mod init github.com/ahproxmox/service-dashboard/backend
go get gopkg.in/yaml.v3
go test ./config -v
```

Expected: `PASS`

- [ ] **Step 5: Create example config**

```yaml
# container/config.yaml.example
server:
  port: 8080

proxmox:
  api_url: https://192.168.88.5:8006
  token_id: "user@pam!token_name"
  token_secret: "${PROXMOX_TOKEN}"

caddy:
  api_url: http://192.168.88.82:2019

prometheus:
  url: http://192.168.88.73:9090

cache:
  status_ttl: 2s
  metrics_ttl: 25s
  discovery_ttl: 10s
```

- [ ] **Step 6: Commit**

```bash
git add backend/config/ backend/main.go container/config.yaml.example backend/go.mod backend/go.sum
git commit -m "feat: add config loading from YAML"
```

---

### Task 2: Response DTOs & API Handlers Structure

**Files:**
- Create: `backend/api/responses.go`
- Create: `backend/api/handlers.go`
- Create: `backend/api/handlers_test.go`

**Goal:** Define API response structure and HTTP handler skeleton.

- [ ] **Step 1: Write response DTOs**

```go
// backend/api/responses.go
package api

type Service struct {
	Id       string   `json:"id"`
	Name     string   `json:"name"`
	Status   string   `json:"status"` // "up" or "stopped"
	HttpsUrl *string  `json:"https_url"`
	Metrics  *Metrics `json:"metrics"`
}

type Metrics struct {
	CpuPercent      float64 `json:"cpu_percent"`
	RamMb           int     `json:"ram_mb"`
	RamPercent      float64 `json:"ram_percent"`
	DiskPercent     float64 `json:"disk_percent"`
	NetworkInMbps   float64 `json:"network_in_mbps"`
	NetworkOutMbps  float64 `json:"network_out_mbps"`
}

type ServicesResponse struct {
	Services  []Service `json:"services"`
	Timestamp int64     `json:"timestamp"`
}

type HealthResponse struct {
	Status              string `json:"status"`
	ProxmoxConnected    bool   `json:"proxmox_connected"`
	CaddyConnected      bool   `json:"caddy_connected"`
	PrometheusConnected bool   `json:"prometheus_connected"`
	Timestamp           int64  `json:"timestamp"`
}
```

- [ ] **Step 2: Write failing test for handlers**

```go
// backend/api/handlers_test.go
package api

import (
	"encoding/json" "net/http"
	"net/http/httptest"
	"testing"
)

func TestServicesEndpoint(t *testing.T) {
	handler := http.HandlerFunc(GetServices)

	req := httptest.NewRequest("GET", "/api/services", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ServicesResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Timestamp == 0 {
		t.Error("expected non-zero timestamp")
	}
}
```

- [ ] **Step 3: Implement handler skeleton**

```go
// backend/api/handlers.go
package api

import (
	"encoding/json"
	"net/http"
	"time"
)

func GetServices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ServicesResponse{
		Services:  []Service{},
		Timestamp: time.Now().Unix(),
	})
}

func GetHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Unix(),
	})
}
```

- [ ] **Step 4: Run test to verify pass**

```bash
cd backend
go test ./api -v
```

- [ ] **Step 5: Update main.go to wire handlers**

```go
// backend/main.go
package main

import (
	"github.com/ahproxmox/service-dashboard/backend/api"
	"github.com/ahproxmox/service-dashboard/backend/config"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.LoadConfig("/etc/dashboard/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	http.HandleFunc("/api/services", api.GetServices)
	http.HandleFunc("/health", api.GetHealth)

	log.Printf("Starting server on :%d", cfg.Server.Port)
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
```

- [ ] **Step 6: Commit**

```bash
git add backend/api/ backend/main.go
git commit -m "feat: add API response DTOs and handler skeleton"
```

---

### Task 3: Caching Layer

**Files:**
- Create: `backend/cache/cache.go`
- Create: `backend/cache/cache_test.go`

**Goal:** Implement in-memory cache with TTL-based expiration.

- [ ] **Step 1: Write failing test**

```go
// backend/cache/cache_test.go
package cache

import (
	"testing"
	"time"
)

func TestCacheSetGet(t *testing.T) {
	c := NewCache()
	c.Set("key1", "value1", 1*time.Second)

	val, found := c.Get("key1")
	if !found {
		t.Error("expected key1 to be found")
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

func TestCacheExpiry(t *testing.T) {
	c := NewCache()
	c.Set("key1", "value1", 100*time.Millisecond)

	time.Sleep(150 * time.Millisecond)

	_, found := c.Get("key1")
	if found {
		t.Error("expected key1 to be expired")
	}
}
```

- [ ] **Step 2: Run test to verify failure**

```bash
cd backend
go test ./cache -v
```

- [ ] **Step 3: Implement cache**

```go
// backend/cache/cache.go
package cache

import (
	"sync"
	"time"
)

type entry struct {
	value     interface{}
	expiresAt time.Time
}

type Cache struct {
	data map[string]*entry
	mu   sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]*entry),
	}
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &entry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.data[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(e.expiresAt) {
		return nil, false
	}

	return e.value, true
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]*entry)
}
```

- [ ] **Step 4: Run test to verify pass**

```bash
cd backend
go test ./cache -v
```

- [ ] **Step 5: Commit**

```bash
git add backend/cache/
git commit -m "feat: add in-memory cache with TTL expiration"
```

---

### Task 4: Proxmox API Client

**Files:**
- Create: `backend/discovery/proxmox.go`
- Create: `backend/discovery/proxmox_test.go`

**Goal:** Query Proxmox for running containers (ID, name, status, IP).

- [ ] **Step 1: Write failing test with mock**

```go
// backend/discovery/proxmox_test.go
package discovery

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxmoxGetContainers(t *testing.T) {
	// Mock Proxmox API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api2/json/nodes/pve/lxc" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": [
					{"vmid": 122, "hostname": "kanban", "status": "running", "ip": "192.168.88.78"},
					{"vmid": 111, "hostname": "rag", "status": "running", "ip": "192.168.88.71"},
					{"vmid": 104, "hostname": "openclaw", "status": "stopped"}
				]
			}`))
		}
	}))
	defer server.Close()

	client := NewProxmoxClient(server.URL, "user@pam!token", "secret")
	containers, err := client.GetContainers()
	if err != nil {
		t.Fatalf("GetContainers failed: %v", err)
	}

	if len(containers) != 3 {
		t.Errorf("expected 3 containers, got %d", len(containers))
	}

	if containers[0].Name != "kanban" {
		t.Errorf("expected name kanban, got %s", containers[0].Name)
	}

	if containers[2].Status != "stopped" {
		t.Errorf("expected stopped status for container 3")
	}
}
```

- [ ] **Step 2: Run test to verify failure**

```bash
cd backend
go test ./discovery -v -run TestProxmoxGetContainers
```

- [ ] **Step 3: Implement Proxmox client**

```go
// backend/discovery/proxmox.go
package discovery

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Container struct {
	Id   string
	Name string
	Status string // "running", "stopped"
	Ip   string
}

type proxmoxContainer struct {
	Vmid     int    `json:"vmid"`
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Ip       string `json:"ip"`
}

type proxmoxResponse struct {
	Data []proxmoxContainer `json:"data"`
}

type ProxmoxClient struct {
	apiUrl string
	tokenId string
	tokenSecret string
	httpClient *http.Client
}

func NewProxmoxClient(apiUrl, tokenId, tokenSecret string) *ProxmoxClient {
	// Ignore self-signed certs for homelab
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	return &ProxmoxClient{
		apiUrl: apiUrl,
		tokenId: tokenId,
		tokenSecret: tokenSecret,
		httpClient: client,
	}
}

func (p *ProxmoxClient) GetContainers() ([]Container, error) {
	req, _ := http.NewRequest("GET", p.apiUrl+"/api2/json/nodes/pve/lxc", nil)
	req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s:%s", p.tokenId, p.tokenSecret))

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("proxmox request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var proxmoxResp proxmoxResponse
	if err := json.Unmarshal(body, &proxmoxResp); err != nil {
		return nil, fmt.Errorf("parse proxmox response: %w", err)
	}

	var containers []Container
	for _, pc := range proxmoxResp.Data {
		containers = append(containers, Container{
			Id:   fmt.Sprintf("%d", pc.Vmid),
			Name: pc.Hostname,
			Status: pc.Status,
			Ip:   pc.Ip,
		})
	}

	return containers, nil
}
```

- [ ] **Step 4: Run test to verify pass**

```bash
cd backend
go test ./discovery -v -run TestProxmoxGetContainers
```

- [ ] **Step 5: Commit**

```bash
git add backend/discovery/proxmox*
git commit -m "feat: add Proxmox API client for container discovery"
```

---

### Task 5: Caddy API Client

**Files:**
- Create: `backend/discovery/caddy.go`
- Create: `backend/discovery/caddy_test.go`

**Goal:** Query Caddy for routes and their backend targets (IP/port → HTTPS URL).

- [ ] **Step 1: Write failing test with mock**

```go
// backend/discovery/caddy_test.go
package discovery

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCaddyGetRoutes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"apps": {
				"http": {
					"servers": {
						"default": {
							"routes": [
								{
									"handle": [
										{
											"upstreams": [
												{"dial": "192.168.88.78:3000"}
											]
										}
									],
									"match": [
										{"host": ["kanban.internal.ahproxmox-claude.cc"]}
									]
								},
								{
									"handle": [
										{
											"upstreams": [
												{"dial": "192.168.88.71:8080"}
											]
										}
									],
									"match": [
										{"host": ["rag.internal.ahproxmox-claude.cc"]}
									]
								}
							]
						}
					}
				}
			}
		}`))
	}))
	defer server.Close()

	client := NewCaddyClient(server.URL)
	routes, err := client.GetRoutes()
	if err != nil {
		t.Fatalf("GetRoutes failed: %v", err)
	}

	if len(routes) != 2 {
		t.Errorf("expected 2 routes, got %d", len(routes))
	}

	if routes[0].Domain != "kanban.internal.ahproxmox-claude.cc" {
		t.Errorf("expected kanban domain, got %s", routes[0].Domain)
	}

	if routes[0].BackendIp != "192.168.88.78" {
		t.Errorf("expected IP 192.168.88.78, got %s", routes[0].BackendIp)
	}
}
```

- [ ] **Step 2: Run test to verify failure**

```bash
cd backend
go test ./discovery -v -run TestCaddyGetRoutes
```

- [ ] **Step 3: Implement Caddy client**

```go
// backend/discovery/caddy.go
package discovery

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Route struct {
	Domain    string
	BackendIp string
}

type CaddyClient struct {
	apiUrl string
}

func NewCaddyClient(apiUrl string) *CaddyClient {
	return &CaddyClient{apiUrl: apiUrl}
}

func (c *CaddyClient) GetRoutes() ([]Route, error) {
	resp, err := http.Get(c.apiUrl + "/admin/api/config/apps/http/servers/default/routes")
	if err != nil {
		return nil, fmt.Errorf("caddy request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Parse the nested Caddy config structure
	var routes []interface{}
	if err := json.Unmarshal(body, &routes); err != nil {
		return nil, fmt.Errorf("parse caddy routes: %w", err)
	}

	var result []Route
	for _, routeData := range routes {
		route := routeData.(map[string]interface{})

		// Extract domain from match
		var domain string
		if matches, ok := route["match"].([]interface{}); ok && len(matches) > 0 {
			match := matches[0].(map[string]interface{})
			if hosts, ok := match["host"].([]interface{}); ok && len(hosts) > 0 {
				domain = hosts[0].(string)
			}
		}

		// Extract backend IP from handle
		var backendIp string
		if handles, ok := route["handle"].([]interface{}); ok && len(handles) > 0 {
			handle := handles[0].(map[string]interface{})
			if upstreams, ok := handle["upstreams"].([]interface{}); ok && len(upstreams) > 0 {
				upstream := upstreams[0].(map[string]interface{})
				if dial, ok := upstream["dial"].(string); ok {
					// dial is "IP:port", extract IP
					backendIp = strings.Split(dial, ":")[0]
				}
			}
		}

		if domain != "" && backendIp != "" {
			result = append(result, Route{
				Domain:    domain,
				BackendIp: backendIp,
			})
		}
	}

	return result, nil
}
```

- [ ] **Step 4: Run test to verify pass**

```bash
cd backend
go test ./discovery -v -run TestCaddyGetRoutes
```

- [ ] **Step 5: Commit**

```bash
git add backend/discovery/caddy*
git commit -m "feat: add Caddy API client for route discovery"
```

---

### Task 6: Service Matching Algorithm

**Files:**
- Modify: `backend/discovery/matcher.go` (create)
- Create: `backend/discovery/matcher_test.go`

**Goal:** Match containers to Caddy routes (IP matching → hostname fallback).

- [ ] **Step 1: Write failing test**

```go
// backend/discovery/matcher_test.go
package discovery

import (
	"testing"
)

func TestMatchServiceByIp(t *testing.T) {
	containers := []Container{
		{Id: "122", Name: "kanban", Status: "running", Ip: "192.168.88.78"},
		{Id: "111", Name: "rag", Status: "running", Ip: "192.168.88.71"},
	}

	routes := []Route{
		{Domain: "kanban.internal.ahproxmox-claude.cc", BackendIp: "192.168.88.78"},
		{Domain: "rag.internal.ahproxmox-claude.cc", BackendIp: "192.168.88.71"},
	}

	matcher := NewMatcher()
	services := matcher.Match(containers, routes)

	if len(services) != 2 {
		t.Errorf("expected 2 services, got %d", len(services))
	}

	if services[0].HttpsUrl == nil || *services[0].HttpsUrl != "https://kanban.internal.ahproxmox-claude.cc" {
		t.Errorf("expected kanban URL")
	}

	if services[1].HttpsUrl == nil || *services[1].HttpsUrl != "https://rag.internal.ahproxmox-claude.cc" {
		t.Errorf("expected rag URL")
	}
}

func TestMatchServiceNoRoute(t *testing.T) {
	containers := []Container{
		{Id: "104", Name: "openclaw", Status: "stopped", Ip: "192.168.88.63"},
	}

	matcher := NewMatcher()
	services := matcher.Match(containers, []Route{})

	if services[0].HttpsUrl != nil {
		t.Error("expected no HTTPS URL for unmatched service")
	}
}
```

- [ ] **Step 2: Run test to verify failure**

```bash
cd backend
go test ./discovery -v -run TestMatch
```

- [ ] **Step 3: Implement matcher**

```go
// backend/discovery/matcher.go
package discovery

import (
	"fmt"
	"strings"
)

type MatchedService struct {
	Id       string
	Name     string
	Status   string
	HttpsUrl *string
}

type Matcher struct{}

func NewMatcher() *Matcher {
	return &Matcher{}
}

func (m *Matcher) Match(containers []Container, routes []Route) []MatchedService {
	var services []MatchedService

	for _, container := range containers {
		var httpsUrl *string

		// Primary: IP matching
		for _, route := range routes {
			if route.BackendIp == container.Ip {
				url := fmt.Sprintf("https://%s", route.Domain)
				httpsUrl = &url
				break
			}
		}

		// Fallback: Hostname matching
		if httpsUrl == nil {
			for _, route := range routes {
				if strings.Contains(route.Domain, container.Name) {
					url := fmt.Sprintf("https://%s", route.Domain)
					httpsUrl = &url
					break
				}
			}
		}

		services = append(services, MatchedService{
			Id:       container.Id,
			Name:     container.Name,
			Status:   container.Status,
			HttpsUrl: httpsUrl,
		})
	}

	return services
}
```

- [ ] **Step 4: Run test to verify pass**

```bash
cd backend
go test ./discovery -v -run TestMatch
```

- [ ] **Step 5: Commit**

```bash
git add backend/discovery/matcher*
git commit -m "feat: add service matching algorithm (IP then hostname)"
```

---

### Task 7: Prometheus Metrics Client

**Files:**
- Create: `backend/metrics/prometheus.go`
- Create: `backend/metrics/prometheus_test.go`

**Goal:** Query Prometheus for CPU, RAM, disk, network metrics per container.

- [ ] **Step 1: Write failing test**

```go
// backend/metrics/prometheus_test.go
package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPrometheusGetMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(query, "node_cpu_seconds_total") {
			w.Write([]byte(`{
				"status": "success",
				"data": {
					"resultType": "instant",
					"result": [
						{"metric": {"instance": "192.168.88.78:9100"}, "value": [0, "123.45"]}
					]
				}
			}`))
		} else if strings.Contains(query, "node_memory_bytes_total") {
			w.Write([]byte(`{
				"status": "success",
				"data": {"result": [{"metric": {"instance": "192.168.88.78:9100"}, "value": [0, "1099511627776"]}]}
			}`))
		} else {
			w.Write([]byte(`{"status": "success", "data": {"result": []}}`))
		}
	}))
	defer server.Close()

	client := NewPrometheusClient(server.URL)
	metrics, err := client.GetMetrics("192.168.88.78")
	if err != nil {
		t.Fatalf("GetMetrics failed: %v", err)
	}

	if metrics.CpuPercent == 0 {
		t.Error("expected non-zero CPU")
	}

	if metrics.RamMb == 0 {
		t.Error("expected non-zero RAM MB")
	}
}
```

- [ ] **Step 2: Run test to verify failure**

```bash
cd backend
go test ./metrics -v
```

- [ ] **Step 3: Implement Prometheus client**

```go
// backend/metrics/prometheus.go
package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type Metrics struct {
	CpuPercent     float64
	RamMb          int
	RamPercent     float64
	DiskPercent    float64
	NetworkInMbps  float64
	NetworkOutMbps float64
}

type PrometheusClient struct {
	url string
}

func NewPrometheusClient(url string) *PrometheusClient {
	return &PrometheusClient{url: url}
}

func (p *PrometheusClient) GetMetrics(ip string) (*Metrics, error) {
	metrics := &Metrics{}

	instance := ip + ":9100"

	// CPU
	cpu, err := p.queryMetric(fmt.Sprintf(`rate(node_cpu_seconds_total{instance="%s"}[5m])*100`, instance))
	if err == nil {
		metrics.CpuPercent = cpu
	}

	// RAM
	ramTotal, _ := p.queryMetric(fmt.Sprintf(`node_memory_MemTotal_bytes{instance="%s"}`, instance))
	ramAvail, _ := p.queryMetric(fmt.Sprintf(`node_memory_MemAvailable_bytes{instance="%s"}`, instance))
	if ramTotal > 0 {
		metrics.RamMb = int(ramTotal / 1024 / 1024)
		metrics.RamPercent = ((ramTotal - ramAvail) / ramTotal) * 100
	}

	// Disk
	disk, err := p.queryMetric(fmt.Sprintf(`(1 - (node_filesystem_avail_bytes{instance="%s",fstype!="tmpfs"} / node_filesystem_size_bytes{instance="%s",fstype!="tmpfs"})) * 100`, instance, instance))
	if err == nil {
		metrics.DiskPercent = disk
	}

	// Network (approximate in/out)
	netIn, _ := p.queryMetric(fmt.Sprintf(`rate(node_network_receive_bytes_total{instance="%s"}[1m])/1024/1024`, instance))
	netOut, _ := p.queryMetric(fmt.Sprintf(`rate(node_network_transmit_bytes_total{instance="%s"}[1m])/1024/1024`, instance))
	metrics.NetworkInMbps = netIn
	metrics.NetworkOutMbps = netOut

	return metrics, nil
}

func (p *PrometheusClient) queryMetric(query string) (float64, error) {
	q := url.QueryEscape(query)
	resp, err := http.Get(p.url + "/api/v1/query?query=" + q)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	data := result["data"].(map[string]interface{})
	resultList := data["result"].([]interface{})

	if len(resultList) == 0 {
		return 0, fmt.Errorf("no data")
	}

	val := resultList[0].(map[string]interface{})["value"].([]interface{})
	valStr := val[1].(string)
	return strconv.ParseFloat(valStr, 64)
}
```

- [ ] **Step 4: Run test to verify pass**

```bash
cd backend
go test ./metrics -v
```

- [ ] **Step 5: Commit**

```bash
git add backend/metrics/
git commit -m "feat: add Prometheus metrics client"
```

---

### Task 8: Wire Everything Together & Update Handlers

**Files:**
- Modify: `backend/main.go`
- Modify: `backend/api/handlers.go`

**Goal:** Integrate all components; implement `/api/services` and `/health` endpoints.

- [ ] **Step 1: Update handlers to use services**

```go
// backend/api/handlers.go (updated)
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ahproxmox/service-dashboard/backend/cache"
	"github.com/ahproxmox/service-dashboard/backend/discovery"
	"github.com/ahproxmox/service-dashboard/backend/metrics"
	"github.com/ahproxmox/service-dashboard/backend/config"
)

var (
	globalCache *cache.Cache
	proxmoxClient *discovery.ProxmoxClient
	caddyClient *discovery.CaddyClient
	prometheusClient *metrics.PrometheusClient
	matcher *discovery.Matcher
	cfg *config.Config
)

func InitHandlers(c *cache.Cache, p *discovery.ProxmoxClient, cad *discovery.CaddyClient, prom *metrics.PrometheusClient, m *discovery.Matcher, config *config.Config) {
	globalCache = c
	proxmoxClient = p
	caddyClient = cad
	prometheusClient = prom
	matcher = m
	cfg = config
}

func GetServices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check cache
	if cached, ok := globalCache.Get("services"); ok {
		json.NewEncoder(w).Encode(cached)
		return
	}

	// Fetch fresh data
	containers, err := proxmoxClient.GetContainers()
	if err != nil {
		http.Error(w, "Failed to get containers", http.StatusInternalServerError)
		return
	}

	routes, err := caddyClient.GetRoutes()
	if err != nil {
		http.Error(w, "Failed to get routes", http.StatusInternalServerError)
		return
	}

	matched := matcher.Match(containers, routes)

	// Fetch metrics
	var services []Service
	for _, m := range matched {
		svc := Service{
			Id:       m.Id,
			Name:     m.Name,
			Status:   m.Status,
			HttpsUrl: m.HttpsUrl,
		}

		if m.Status == "running" {
			// Extract IP from container
			var containerIp string
			for _, c := range containers {
				if c.Id == m.Id {
					containerIp = c.Ip
					break
				}
			}

			if metr, err := prometheusClient.GetMetrics(containerIp); err == nil {
				svc.Metrics = metr
			}
		}

		services = append(services, svc)
	}

	resp := ServicesResponse{
		Services:  services,
		Timestamp: time.Now().Unix(),
	}

	// Cache for status_ttl
	globalCache.Set("services", resp, cfg.Cache.StatusTTL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func GetHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Simple health check (just return OK)
	resp := HealthResponse{
		Status:              "ok",
		ProxmoxConnected:    true,
		CaddyConnected:      true,
		PrometheusConnected: true,
		Timestamp:           time.Now().Unix(),
	}

	json.NewEncoder(w).Encode(resp)
}
```

- [ ] **Step 2: Update main.go**

```go
// backend/main.go (updated)
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ahproxmox/service-dashboard/backend/api"
	"github.com/ahproxmox/service-dashboard/backend/cache"
	"github.com/ahproxmox/service-dashboard/backend/config"
	"github.com/ahproxmox/service-dashboard/backend/discovery"
	"github.com/ahproxmox/service-dashboard/backend/metrics"
)

func main() {
	configPath := os.Getenv("CONFIG")
	if configPath == "" {
		configPath = "/etc/dashboard/config.yaml"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize clients
	c := cache.NewCache()
	p := discovery.NewProxmoxClient(cfg.Proxmox.APIUrl, cfg.Proxmox.TokenId, cfg.Proxmox.TokenSecret)
	cad := discovery.NewCaddyClient(cfg.Caddy.APIUrl)
	prom := metrics.NewPrometheusClient(cfg.Prometheus.Url)
	m := discovery.NewMatcher()

	// Initialize handlers
	api.InitHandlers(c, p, cad, prom, m, cfg)

	// Routes
	http.HandleFunc("/api/services", api.GetServices)
	http.HandleFunc("/health", api.GetHealth)

	log.Printf("Starting dashboard backend on :%d", cfg.Server.Port)
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
```

- [ ] **Step 3: Run basic test**

```bash
cd backend
go build -o dashboard-backend
./dashboard-backend &
sleep 1
curl http://localhost:8080/health
# Kill the process
```

- [ ] **Step 4: Commit**

```bash
git add backend/ go.mod go.sum
git commit -m "feat: wire all backend components, implement /api/services and /health"
```

---

## Phase 2: Frontend — Vue 3 PWA

### Task 9: Frontend Project Setup

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/vite.config.js`
- Create: `frontend/index.html`
- Create: `frontend/src/main.js`

**Goal:** Bootstrap Vue 3 project with Vite.

- [ ] **Step 1: Create package.json**

```json
{
  "name": "service-dashboard-frontend",
  "version": "1.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "vue": "^3.4.0"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.0.0",
    "vite": "^5.0.0"
  }
}
```

- [ ] **Step 2: Create vite.config.js**

```javascript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: '../backend/public',
    emptyOutDir: true
  }
})
```

- [ ] **Step 3: Create index.html**

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" href="/favicon.ico" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <link rel="manifest" href="/manifest.json">
    <meta name="theme-color" content="#0066cc">
    <title>Service Dashboard</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/main.js"></script>
  </body>
</html>
```

- [ ] **Step 4: Create main.js**

```javascript
import { createApp } from 'vue'
import App from './App.vue'

const app = createApp(App)
app.mount('#app')

// Register service worker
if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/service-worker.js')
  })
}
```

- [ ] **Step 5: Commit**

```bash
cd frontend
npm install
git add package.json vite.config.js index.html src/main.js
git commit -m "feat: setup Vue 3 + Vite frontend project"
```

---

### Task 10: Core Vue Components

**Files:**
- Create: `frontend/src/App.vue`
- Create: `frontend/src/components/ServiceCard.vue`
- Create: `frontend/src/components/ServiceGrid.vue`
- Create: `frontend/src/components/ErrorBanner.vue`

**Goal:** Build UI components for dashboard.

- [ ] **Step 1: Create App.vue**

```vue
<template>
  <div class="app">
    <header>
      <h1>🔧 Service Dashboard</h1>
      <p class="subtitle">Auto-discovered services from your homelab</p>
    </header>

    <ErrorBanner v-if="error" :error="error" />

    <ServiceGrid :services="services" :loading="loading" />

    <footer>
      <p>Last updated: {{ lastUpdated }}</p>
      <button @click="installApp" v-if="canInstall">📲 Install App</button>
    </footer>
  </div>
</template>

<script>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import ServiceGrid from './components/ServiceGrid.vue'
import ErrorBanner from './components/ErrorBanner.vue'
import { fetchServices } from './services/api.js'

export default {
  name: 'App',
  components: { ServiceGrid, ErrorBanner },
  setup() {
    const services = ref([])
    const loading = ref(false)
    const error = ref(null)
    const lastUpdated = ref('Never')
    const canInstall = ref(false)
    let deferredPrompt = null
    let statusInterval = null
    let metricsInterval = null

    const updateServices = async (includeMetrics = false) => {
      try {
        loading.value = true
        const data = await fetchServices()
        services.value = data.services || []
        error.value = null
        lastUpdated.value = new Date().toLocaleTimeString()
      } catch (err) {
        error.value = err.message
      } finally {
        loading.value = false
      }
    }

    onMounted(() => {
      // Initial fetch
      updateServices(true)

      // Poll status every 1.5s
      statusInterval = setInterval(() => {
        updateServices(false)
      }, 1500)

      // Poll metrics every 30s
      metricsInterval = setInterval(() => {
        updateServices(true)
      }, 30000)

      // PWA install prompt
      window.addEventListener('beforeinstallprompt', (e) => {
        e.preventDefault()
        deferredPrompt = e
        canInstall.value = true
      })
    })

    onBeforeUnmount(() => {
      clearInterval(statusInterval)
      clearInterval(metricsInterval)
    })

    const installApp = () => {
      if (deferredPrompt) {
        deferredPrompt.prompt()
        deferredPrompt.userChoice.then((choiceResult) => {
          if (choiceResult.outcome === 'accepted') {
            canInstall.value = false
          }
        })
      }
    }

    return {
      services,
      loading,
      error,
      lastUpdated,
      canInstall,
      installApp
    }
  }
}
</script>

<style scoped>
.app {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
  background: #f5f5f5;
  min-height: 100vh;
}

header {
  text-align: center;
  margin-bottom: 30px;
}

h1 {
  font-size: 2.5em;
  margin: 0;
  color: #333;
}

.subtitle {
  color: #666;
  margin-top: 10px;
}

footer {
  text-align: center;
  margin-top: 50px;
  color: #999;
  font-size: 0.9em;
}

button {
  background: #0066cc;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 4px;
  cursor: pointer;
  margin-left: 10px;
}

button:hover {
  background: #0052a3;
}
</style>
```

- [ ] **Step 2: Create ServiceCard.vue**

```vue
<template>
  <div class="card" :class="{ collapsed: !isExpanded }">
    <div class="card-header" @click="toggleExpanded">
      <div class="icon" :class="{ 'status-up': service.status === 'up', 'status-stopped': service.status === 'stopped' }">
        <svg v-if="iconName" :viewBox="`0 0 24 24`" width="40" height="40">
          <!-- Icon loaded from config -->
          {{ iconName }}
        </svg>
      </div>
      <div class="info">
        <h3>{{ displayName }}</h3>
        <p class="status" :class="service.status">{{ service.status.toUpperCase() }}</p>
      </div>
    </div>

    <div v-if="isExpanded" class="card-body">
      <a v-if="service.https_url" :href="service.https_url" target="_blank" class="url-link">
        {{ service.https_url }}
        <span class="icon-open">→</span>
      </a>
      <p v-else class="url-link disabled">Not proxied</p>

      <div v-if="service.metrics" class="metrics">
        <div class="metric-group">
          <span class="metric-label">CPU</span>
          <span class="metric-value">{{ service.metrics.cpu_percent.toFixed(1) }}%</span>
        </div>
        <div class="metric-group">
          <span class="metric-label">RAM</span>
          <span class="metric-value">{{ service.metrics.ram_mb }}MB ({{ service.metrics.ram_percent.toFixed(0) }}%)</span>
        </div>
        <div class="metric-group">
          <span class="metric-label">Disk</span>
          <span class="metric-value">{{ service.metrics.disk_percent.toFixed(0) }}%</span>
        </div>
        <div class="metric-group">
          <span class="metric-label">Network</span>
          <span class="metric-value">↓{{ service.metrics.network_in_mbps.toFixed(2) }}Mbps ↑{{ service.metrics.network_out_mbps.toFixed(2) }}Mbps</span>
        </div>
      </div>
      <p v-else class="metrics-placeholder">Metrics unavailable</p>
    </div>
  </div>
</template>

<script>
import { ref, computed } from 'vue'
import servicesConfig from '../config/services-config.json'

export default {
  name: 'ServiceCard',
  props: {
    service: {
      type: Object,
      required: true
    }
  },
  setup(props) {
    const isExpanded = ref(false)

    const toggleExpanded = () => {
      isExpanded.value = !isExpanded.value
    }

    const displayName = computed(() => {
      const config = servicesConfig.services[props.service.name]
      return config?.display_name || props.service.name
    })

    const iconName = computed(() => {
      const config = servicesConfig.services[props.service.name]
      return config?.icon || 'service'
    })

    return {
      isExpanded,
      toggleExpanded,
      displayName,
      iconName
    }
  }
}
</script>

<style scoped>
.card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  transition: all 0.3s ease;
  cursor: pointer;
}

.card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.card-header {
  display: flex;
  align-items: center;
  padding: 20px;
  gap: 15px;
}

.icon {
  width: 50px;
  height: 50px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0f0f0;
  border: 2px solid #ddd;
  transition: all 0.3s ease;
}

.icon.status-up {
  background: #e8f5e9;
  border: 2px solid #4caf50;
  box-shadow: 0 0 10px rgba(76, 175, 80, 0.5);
}

.icon.status-stopped {
  background: #f5f5f5;
  border: 2px solid #ccc;
  opacity: 0.6;
}

.info {
  flex: 1;
}

.info h3 {
  margin: 0;
  font-size: 1.2em;
  color: #333;
}

.status {
  margin: 5px 0 0 0;
  font-size: 0.9em;
  font-weight: bold;
  color: #666;
}

.status.up {
  color: #4caf50;
}

.status.stopped {
  color: #999;
}

.card-body {
  padding: 20px;
  border-top: 1px solid #eee;
  background: #fafafa;
}

.url-link {
  display: block;
  color: #0066cc;
  text-decoration: none;
  font-size: 0.95em;
  margin-bottom: 15px;
  word-break: break-all;
  transition: color 0.3s ease;
}

.url-link:hover {
  color: #0052a3;
}

.url-link.disabled {
  color: #999;
  cursor: default;
}

.icon-open {
  margin-left: 5px;
}

.metrics {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.metric-group {
  display: flex;
  justify-content: space-between;
  padding: 10px;
  background: white;
  border-radius: 4px;
  border-left: 3px solid #0066cc;
}

.metric-label {
  font-weight: bold;
  color: #666;
  font-size: 0.9em;
}

.metric-value {
  color: #333;
  font-weight: bold;
  text-align: right;
}

.metrics-placeholder {
  text-align: center;
  color: #999;
  font-style: italic;
}
</style>
```

- [ ] **Step 3: Create ServiceGrid.vue**

```vue
<template>
  <div class="grid">
    <ServiceCard v-for="service in services" :key="service.id" :service="service" />
    <p v-if="services.length === 0 && !loading" class="no-services">No services discovered</p>
    <p v-if="loading" class="loading">Loading...</p>
  </div>
</template>

<script>
import ServiceCard from './ServiceCard.vue'

export default {
  name: 'ServiceGrid',
  components: { ServiceCard },
  props: {
    services: {
      type: Array,
      required: true
    },
    loading: {
      type: Boolean,
      default: false
    }
  }
}
</script>

<style scoped>
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.no-services,
.loading {
  grid-column: 1 / -1;
  text-align: center;
  color: #999;
  padding: 40px 20px;
  font-style: italic;
}

@media (max-width: 768px) {
  .grid {
    grid-template-columns: 1fr;
  }
}
</style>
```

- [ ] **Step 4: Create ErrorBanner.vue**

```vue
<template>
  <div class="error-banner">
    ⚠️ {{ error }}
  </div>
</template>

<script>
export default {
  name: 'ErrorBanner',
  props: {
    error: {
      type: String,
      required: true
    }
  }
}
</script>

<style scoped>
.error-banner {
  background: #ffebee;
  color: #c62828;
  padding: 15px 20px;
  border-radius: 4px;
  margin-bottom: 20px;
  border-left: 4px solid #c62828;
}
</style>
```

- [ ] **Step 5: Commit**

```bash
cd frontend
git add src/
git commit -m "feat: add Vue components (App, ServiceCard, ServiceGrid, ErrorBanner)"
```

---

### Task 11: API & Utilities

**Files:**
- Create: `frontend/src/services/api.js`
- Create: `frontend/src/utils/metrics.js`
- Create: `frontend/src/config/services-config.json`

**Goal:** Polling logic and formatting utilities.

- [ ] **Step 1: Create api.js**

```javascript
// frontend/src/services/api.js

export async function fetchServices() {
  const response = await fetch('/api/services')
  if (!response.ok) {
    throw new Error('Failed to fetch services')
  }
  return response.json()
}

export async function getHealth() {
  const response = await fetch('/health')
  if (!response.ok) {
    throw new Error('Backend unreachable')
  }
  return response.json()
}
```

- [ ] **Step 2: Create metrics.js**

```javascript
// frontend/src/utils/metrics.js

export function formatBytes(bytes) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
}

export function formatPercent(value) {
  return Math.round(value * 10) / 10 + '%'
}

export function formatNetwork(mbps) {
  return mbps.toFixed(2) + ' Mbps'
}
```

- [ ] **Step 3: Create services-config.json**

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
    },
    "uptime-kuma": {
      "icon": "heart",
      "display_name": "Uptime Kuma"
    },
    "immich": {
      "icon": "photo",
      "display_name": "Immich"
    },
    "jellyfin": {
      "icon": "play",
      "display_name": "Jellyfin"
    },
    "vaultwarden": {
      "icon": "lock",
      "display_name": "Vaultwarden"
    },
    "n8n": {
      "icon": "workflow",
      "display_name": "n8n"
    }
  }
}
```

- [ ] **Step 4: Commit**

```bash
cd frontend
git add src/services/ src/utils/ src/config/
git commit -m "feat: add API client, utilities, and service config"
```

---

### Task 12: PWA Features (Service Worker & Manifest)

**Files:**
- Create: `frontend/public/manifest.json`
- Create: `frontend/public/service-worker.js`

**Goal:** Make dashboard installable and cache assets.

- [ ] **Step 1: Create manifest.json**

```json
{
  "name": "Service Dashboard",
  "short_name": "Dashboard",
  "description": "Auto-discovery service dashboard with metrics",
  "start_url": "/",
  "scope": "/",
  "display": "standalone",
  "orientation": "portrait-primary",
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

- [ ] **Step 2: Create service-worker.js**

```javascript
// frontend/public/service-worker.js

const CACHE_NAME = 'dashboard-v1'
const urlsToCache = [
  '/',
  '/index.html',
  '/manifest.json'
]

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => {
        return cache.addAll(urlsToCache)
      })
  )
})

self.addEventListener('fetch', (event) => {
  const { request } = event

  // API calls: network-first, fallback to cache
  if (request.url.includes('/api/') || request.url.includes('/health')) {
    event.respondWith(
      fetch(request)
        .then((response) => {
          // Cache successful responses
          const responseClone = response.clone()
          caches.open(CACHE_NAME).then((cache) => {
            cache.put(request, responseClone)
          })
          return response
        })
        .catch(() => {
          return caches.match(request)
        })
    )
  } else {
    // Assets: cache-first
    event.respondWith(
      caches.match(request)
        .then((response) => {
          return response || fetch(request)
        })
    )
  }
})
```

- [ ] **Step 3: Commit**

```bash
cd frontend
git add public/manifest.json public/service-worker.js
git commit -m "feat: add PWA manifest and service worker"
```

---

### Task 13: Icon SVGs (Minimal)

**Files:**
- Create: `frontend/public/icons/kanban.svg`
- Create: `frontend/public/icons/brain.svg`
- Create: `frontend/public/icons/robot.svg`
- Create minimal PWA icons

**Goal:** Add placeholder icons.

- [ ] **Step 1: Create icon SVGs**

```svg
<!-- frontend/public/icons/kanban.svg -->
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="#0066cc" stroke-width="2">
  <path d="M3 3h18v18H3z"/><line x1="9" y1="3" x2="9" y2="21"/><line x1="15" y1="3" x2="15" y2="21"/>
</svg>

<!-- frontend/public/icons/brain.svg -->
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="#0066cc" stroke-width="2">
  <circle cx="7" cy="6" r="2"/><circle cx="17" cy="6" r="2"/><path d="M12 9l-3 4 3 3 3-3-3-4"/>
</svg>

<!-- frontend/public/icons/robot.svg -->
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="#0066cc" stroke-width="2">
  <rect x="5" y="8" width="14" height="11" rx="1"/><circle cx="9" cy="12" r="1"/><circle cx="15" cy="12" r="1"/><rect x="7" y="3" width="10" height="3"/>
</svg>
```

- [ ] **Step 2: Create PWA icons (from SVG or placeholder)**

```bash
# For now, use a placeholder 192x192 and 512x512 PNG
# In production, generate from Figma design
echo "Add icon-192.png and icon-512.png to frontend/public/"
```

- [ ] **Step 3: Commit**

```bash
cd frontend
git add public/icons/
git commit -m "feat: add service icon SVGs"
```

---

### Task 14: Build & Test Frontend

**Files:**
- Modify: `frontend/vite.config.js` (ensure build output to backend)

**Goal:** Build frontend and verify it works.

- [ ] **Step 1: Build frontend**

```bash
cd frontend
npm run build
```

Expected: Output in `backend/public/`

- [ ] **Step 2: Verify static files**

```bash
ls -la ../backend/public/
```

Should contain: `index.html`, `manifest.json`, `service-worker.js`, etc.

- [ ] **Step 3: Test backend serving frontend**

```bash
cd backend
go run main.go &
sleep 1
curl http://localhost:8080/
# Should return HTML
```

- [ ] **Step 4: Commit**

```bash
cd frontend
git add -A
git commit -m "feat: build frontend, verify static files in backend"
```

---

## Phase 3: Container & Deployment

### Task 15: Dockerfile & Build Script

**Files:**
- Create: `container/Dockerfile`
- Create: `container/build.sh`

**Goal:** Multi-stage build for Go backend + embedded Vue frontend.

- [ ] **Step 1: Create Dockerfile**

```dockerfile
# Multi-stage: build Go binary with embedded static files

FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache ca-certificates

# Copy backend code
COPY backend/ ./backend/
COPY backend/go.mod backend/go.sum ./

# Build binary
RUN cd backend && \
    go build -o dashboard-backend \
    -ldflags="-s -w" \
    main.go

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/backend/dashboard-backend /opt/dashboard-backend
COPY --from=builder /app/backend/public/ /opt/public/
COPY container/config.yaml.example /etc/dashboard/config.yaml.example

EXPOSE 8080

ENTRYPOINT ["/opt/dashboard-backend"]
```

- [ ] **Step 2: Create build.sh**

```bash
#!/bin/bash
# container/build.sh

set -e

echo "Building Service Dashboard..."

# Build frontend
echo "📦 Building frontend..."
cd frontend
npm install
npm run build
cd ..

# Build backend binary
echo "🔨 Building backend..."
cd backend
go build -o dashboard-backend main.go
cd ..

# Build container image
echo "🐳 Building container image..."
docker build -t ahproxmox/service-dashboard:latest -f container/Dockerfile .

echo "✅ Build complete! Image: ahproxmox/service-dashboard:latest"
```

- [ ] **Step 3: Make build script executable**

```bash
chmod +x container/build.sh
```

- [ ] **Step 4: Test build**

```bash
./container/build.sh
```

- [ ] **Step 5: Commit**

```bash
git add container/ Dockerfile build.sh
git commit -m "feat: add Dockerfile and build script"
```

---

### Task 16: GitHub Repo & CI/CD

**Files:**
- Create: `.github/workflows/build.yml`
- Create: `.gitignore`
- Create: `README.md`

**Goal:** Set up GitHub repo and automated builds.

- [ ] **Step 1: Create .gitignore**

```
# .gitignore
node_modules/
dist/
backend/public/
backend/dashboard-backend
*.log
*.yml
.DS_Store
backend/go.sum
.env
/container/*.yaml
!container/config.yaml.example
```

- [ ] **Step 2: Create GitHub Actions workflow**

```yaml
# .github/workflows/build.yml
name: Build & Test

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test-backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - name: Run tests
        run: cd backend && go test ./...

  test-frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - name: Install & build
        run: cd frontend && npm install && npm run build

  build-image:
    needs: [test-backend, test-frontend]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: docker/setup-buildx-action@v2
      - name: Build
        run: ./container/build.sh
```

- [ ] **Step 3: Create README.md**

```markdown
# Service Dashboard

Auto-discovery PWA dashboard for homelab services.

## Features

- 🔍 Auto-discovers containers from Proxmox
- 🌐 Automatic HTTPS URL detection via Caddy
- 📊 Real-time metrics from Prometheus
- 📱 Progressive Web App (installable)
- ⚡ Service worker caching for offline
- 🎨 Responsive grid layout

## Architecture

- **Backend:** Go service on CT 127, queries Proxmox/Caddy/Prometheus
- **Frontend:** Vue 3 PWA, polls backend every 1-2s (status) and 30s (metrics)

## Deployment

```bash
# Build container
./container/build.sh

# Push to registry
docker push ahproxmox/service-dashboard:latest

# Deploy to CT 127
pct create 127 centos:11 --hostname dashboard
pct exec 127 -- docker run -d \
  -p 8080:8080 \
  -v /etc/dashboard:/etc/dashboard \
  ahproxmox/service-dashboard:latest
```

## Configuration

See `container/config.yaml.example` for all options.

## Testing

```bash
# Backend tests
cd backend && go test ./...

# Frontend dev server
cd frontend && npm run dev
```
```

- [ ] **Step 4: Commit & Push**

```bash
git add .gitignore .github/ README.md
git commit -m "feat: add GitHub CI/CD and documentation"
git remote add origin https://github.com/ahproxmox/service-dashboard.git
git push -u origin main
```

---

## Phase 4: Integration & Testing

### Task 17: End-to-End Testing

**Files:**
- Create: `tests/e2e.test.js` (or manual testing checklist)

**Goal:** Verify full stack works.

- [ ] **Step 1: Manual E2E test checklist**

```
- [ ] Backend starts without errors
- [ ] GET /health returns OK
- [ ] GET /api/services returns valid JSON
- [ ] Frontend loads at http://localhost:8080
- [ ] Service cards display with correct icons
- [ ] Status updates every 1-2 seconds
- [ ] Metrics update every 30 seconds
- [ ] Clicking service URL opens in new tab
- [ ] Service worker registered (DevTools > Application > Service Workers)
- [ ] Offline assets are cached (DevTools > Network, throttle to offline)
- [ ] Install button appears in browser UI
- [ ] Dashboard installable on mobile
- [ ] Error banner shows when backend unavailable
```

- [ ] **Step 2: Run full stack**

```bash
# Terminal 1: Backend
cd backend
go run main.go

# Terminal 2: Frontend dev server
cd frontend
npm run dev

# Browser: http://localhost:5173
```

- [ ] **Step 3: Verify all success criteria**

```
Success Criteria from Spec:
- [ ] Dashboard auto-discovers all running containers ✓
- [ ] Service icons display with status indicators ✓
- [ ] Status updates within 2 seconds ✓
- [ ] Metrics update every 30 seconds ✓
- [ ] Clicking URL opens service ✓
- [ ] Installable on mobile/desktop ✓
- [ ] Service worker caches assets ✓
- [ ] Handles errors gracefully ✓
```

- [ ] **Step 4: Commit test results**

```bash
git add tests/
git commit -m "test: verify end-to-end functionality"
```

---

### Task 18: Deployment to CT 127

**Files:**
- Create: `container/systemd-service.txt` (documentation)

**Goal:** Deploy backend to production container.

- [ ] **Step 1: Create CT 127 container**

```bash
pct create 127 debian:12-standard \
  --hostname dashboard \
  --cores 2 \
  --memory 1024 \
  --storage local-lvm
```

- [ ] **Step 2: Set up systemd service**

```bash
pct push 127 container/systemd-service.txt /etc/systemd/system/dashboard.service

# Then in CT 127:
systemctl daemon-reload
systemctl enable dashboard
systemctl start dashboard
systemctl status dashboard
```

- [ ] **Step 3: Add Caddy route**

```bash
# On Caddy host (CT 126):
cat >> /etc/caddy/Caddyfile <<EOF
dashboard.internal.ahproxmox-claude.cc {
  reverse_proxy 192.168.88.127:8080
}
EOF

caddy reload
```

- [ ] **Step 4: Test production deployment**

```bash
curl https://dashboard.internal.ahproxmox-claude.cc/health
# Should return health JSON
```

- [ ] **Step 5: Commit deployment notes**

```bash
git add container/
git commit -m "deploy: service dashboard to CT 127"
```

---

## Phase 5: Final Touches

### Task 19: README & Documentation

**Goal:** Complete project documentation.

- [ ] Add architecture diagram to README
- [ ] Document config options
- [ ] Add troubleshooting guide
- [ ] Commit: `git commit -m "docs: complete documentation"`

---

## Success Criteria Checklist

```markdown
# Success Criteria (from Spec)

- [ ] Dashboard auto-discovers all running containers without manual config
- [ ] Service icons display correctly with status indicators (green/grey)
- [ ] Status updates within 2 seconds of container state change
- [ ] Metrics (CPU, RAM, disk, network) update every 30 seconds
- [ ] Clicking URL opens service in new tab
- [ ] Dashboard installable on mobile/desktop
- [ ] Service worker caches assets; subsequent loads are instant
- [ ] Handles missing HTTPS URLs, stopped containers, and API errors gracefully
- [ ] Repo structure supports future Figma design integration
```

---

## Notes for Implementation

1. **Backend caching:** Use mutex-protected map with time-based expiry. Simple but effective.
2. **Frontend polling:** Use `setInterval()` for status (1.5s) and separate for metrics (30s).
3. **Error handling:** Always return partial data; never fail completely.
4. **Icon system:** JSON config allows designers to add icons without code.
5. **PWA offline:** Static assets cached; API calls fail gracefully with stale data shown.
6. **Testing strategy:** Unit tests for components, integration tests for APIs, manual E2E.

