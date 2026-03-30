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
	// Initialize handlers with minimal mocks
	c := cache.NewCache()
	proxmox := discovery.NewProxmoxClient("https://test:8006", "test@pam!token", "secret")
	caddy := discovery.NewCaddyClient("http://test:2019")
	prom := metrics.NewPrometheusClient("http://test:9090")
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
