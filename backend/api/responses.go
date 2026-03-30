package api

type Service struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"` // "running" or "stopped"
	HttpsUrl *string `json:"httpsUrl"`
	Metrics  interface{} `json:"metrics"` // Can be nil or *metrics.Metrics
}

type Metrics struct {
	CpuPercent      float64 `json:"cpuPercent"`
	RamMb           int     `json:"ramMb"`
	RamPercent      float64 `json:"ramPercent"`
	DiskPercent     float64 `json:"diskPercent"`
	NetworkInMbps   float64 `json:"networkInMbps"`
	NetworkOutMbps  float64 `json:"networkOutMbps"`
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
