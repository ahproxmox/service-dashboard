package api

import (
	"github.com/ahproxmox/service-dashboard/backend/metrics"
)

type Service struct {
	Id       string            `json:"id"`
	Name     string            `json:"name"`
	Status   string            `json:"status"` // "running" or "stopped"
	HttpsUrl *string           `json:"httpsUrl"`
	Metrics  *metrics.Metrics  `json:"metrics"` // Can be nil or *metrics.Metrics
}

type ServicesResponse struct {
	Services  []Service `json:"services"`
	Timestamp int64     `json:"timestamp"`
}

type HealthResponse struct {
	Status              string `json:"status"`
	ProxmoxConnected    bool   `json:"proxmoxConnected"`
	CaddyConnected      bool   `json:"caddyConnected"`
	PrometheusConnected bool   `json:"prometheusConnected"`
	Timestamp           int64  `json:"timestamp"`
}
