package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
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
		} else if strings.Contains(query, "node_memory_MemTotal_bytes") {
			w.Write([]byte(`{
				"status": "success",
				"data": {
					"resultType": "instant",
					"result": [
						{"metric": {"instance": "192.168.88.78:9100"}, "value": [0, "1099511627776"]}
					]
				}
			}`))
		} else if strings.Contains(query, "node_memory_MemAvailable_bytes") {
			w.Write([]byte(`{
				"status": "success",
				"data": {
					"resultType": "instant",
					"result": [
						{"metric": {"instance": "192.168.88.78:9100"}, "value": [0, "549755813888"]}
					]
				}
			}`))
		} else {
			w.Write([]byte(`{"status": "success", "data": {"resultType": "instant", "result": []}}`))
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

	if metrics.RamPercent == 0 {
		t.Error("expected non-zero RAM percent")
	}
}
