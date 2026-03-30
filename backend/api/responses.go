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
