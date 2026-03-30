package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ahproxmox/service-dashboard/backend/cache"
	"github.com/ahproxmox/service-dashboard/backend/config"
	"github.com/ahproxmox/service-dashboard/backend/discovery"
	"github.com/ahproxmox/service-dashboard/backend/metrics"
)

func TestServicesEndpoint(t *testing.T) {
	// Mock Proxmox server
	proxmoxServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api2/json/nodes/pve/lxc" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"vmid":122,"hostname":"kanban","status":"running","ip":"192.168.88.78"}]}`))
		}
	}))
	defer proxmoxServer.Close()

	// Mock Caddy server
	caddyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"handle":[{"upstreams":[{"dial":"192.168.88.78:3000"}]}],"match":[{"host":["kanban.internal"]}]}]`))
	}))
	defer caddyServer.Close()

	// Mock Prometheus server
	promServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"instant","result":[]}}`))
	}))
	defer promServer.Close()

	// Initialize handlers with mock servers
	c := cache.NewCache()
	proxmox := discovery.NewProxmoxClient(proxmoxServer.URL, "test@pam!token", "secret")
	caddy := discovery.NewCaddyClient(caddyServer.URL)
	prom := metrics.NewPrometheusClient(promServer.URL)
	matcher := discovery.NewMatcher()
	cfg := &config.Config{}

	InitHandlers(c, proxmox, caddy, prom, matcher, cfg)

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

func TestHealthEndpoint(t *testing.T) {
	handler := http.HandlerFunc(GetHealth)

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Timestamp == 0 {
		t.Error("expected non-zero timestamp")
	}
}
