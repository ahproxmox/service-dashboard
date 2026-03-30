package main

import (
	"fmt"
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
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/etc/service-dashboard/config.yaml"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize cache
	c := cache.NewCache()

	// Initialize clients
	proxmoxClient := discovery.NewProxmoxClient(cfg.Proxmox.APIUrl, cfg.Proxmox.TokenId, cfg.Proxmox.TokenSecret)
	caddyClient := discovery.NewCaddyClient(cfg.Caddy.APIUrl)
	prometheusClient := metrics.NewPrometheusClient(cfg.Prometheus.Url)
	matcher := discovery.NewMatcher()

	// Initialize handlers with components
	api.InitHandlers(c, proxmoxClient, caddyClient, prometheusClient, matcher, cfg)

	// Register routes
	http.HandleFunc("/api/services", api.GetServices)
	http.HandleFunc("/health", api.GetHealth)

	// Serve frontend static files
	fs := http.FileServer(http.Dir("frontend/public"))
	http.Handle("/", fs)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Service Dashboard starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
