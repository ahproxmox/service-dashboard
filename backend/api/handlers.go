package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ahproxmox/service-dashboard/backend/cache"
	"github.com/ahproxmox/service-dashboard/backend/config"
	"github.com/ahproxmox/service-dashboard/backend/discovery"
	"github.com/ahproxmox/service-dashboard/backend/metrics"
)

// Global references (set by InitHandlers)
var (
	globalCache      *cache.Cache
	proxmoxClient    *discovery.ProxmoxClient
	caddyClient      *discovery.CaddyClient
	prometheusClient *metrics.PrometheusClient
	matcher          *discovery.Matcher
	cfg              *config.Config
)

// InitHandlers initializes global handlers with required components
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

	// Check cache first
	if cached, ok := globalCache.Get("services"); ok {
		if resp, ok := cached.(ServicesResponse); ok {
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
	}

	// Fetch fresh data
	containers, err := proxmoxClient.GetContainers()
	if err != nil {
		log.Printf("failed to get containers: %v", err)
		http.Error(w, fmt.Sprintf("failed to get containers: %v", err), http.StatusInternalServerError)
		return
	}

	routes, err := caddyClient.GetRoutes()
	if err != nil {
		log.Printf("caddy routes unavailable (non-fatal): %v", err)
		routes = nil
	} else {
		log.Printf("found %d caddy routes", len(routes))
		for _, r := range routes {
			log.Printf("  route: %s -> %s", r.Domain, r.BackendIp)
		}
	}

	// Match services
	matched := matcher.Match(containers, routes)

	// Build response with metrics
	var services []Service
	for _, m := range matched {
		svc := Service{
			Id:       m.Id,
			Name:     m.Name,
			Status:   m.Status,
			HttpsUrl: m.HttpsUrl,
		}

		// Fetch metrics only for running containers
		if m.Status == "running" && m.Id != "" {
			// Find container IP for metrics query
			var containerIp string
			for _, c := range containers {
				if c.Id == m.Id {
					containerIp = c.Ip
					break
				}
			}

			if containerIp != "" {
				if metr, err := prometheusClient.GetMetrics(containerIp); err == nil {
					svc.Metrics = metr
				}
			}
		}

		services = append(services, svc)
	}

	// Build response
	resp := ServicesResponse{
		Services:  services,
		Timestamp: time.Now().Unix(),
	}

	// Cache for StatusTTL duration
	globalCache.Set("services", resp, cfg.Cache.StatusTTL)

	_ = json.NewEncoder(w).Encode(resp)
}

func GetHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check connectivity (quick probe)
	_, proxmoxErr := proxmoxClient.GetContainers()
	_, caddyErr := caddyClient.GetRoutes()
	_, prometheusErr := prometheusClient.GetMetrics("127.0.0.1")

	resp := HealthResponse{
		Status:              "ok",
		ProxmoxConnected:    proxmoxErr == nil,
		CaddyConnected:      caddyErr == nil,
		PrometheusConnected: prometheusErr == nil,
		Timestamp:           time.Now().Unix(),
	}

	_ = json.NewEncoder(w).Encode(resp)
}
