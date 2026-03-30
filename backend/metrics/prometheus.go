package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
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
	url        string
	httpClient *http.Client
}

func NewPrometheusClient(url string) *PrometheusClient {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	return &PrometheusClient{
		url:        url,
		httpClient: client,
	}
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
	resp, err := p.httpClient.Get(p.url + "/api/v1/query?query=" + q)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("prometheus error: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read prometheus response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("parse prometheus response: %w", err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid prometheus response structure")
	}

	resultList, ok := data["result"].([]interface{})
	if !ok || len(resultList) == 0 {
		return 0, fmt.Errorf("no data in prometheus response")
	}

	val, ok := resultList[0].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid result structure")
	}

	value, ok := val["value"].([]interface{})
	if !ok || len(value) < 2 {
		return 0, fmt.Errorf("invalid value format")
	}

	valStr, ok := value[1].(string)
	if !ok {
		return 0, fmt.Errorf("value is not a string")
	}

	return strconv.ParseFloat(valStr, 64)
}
