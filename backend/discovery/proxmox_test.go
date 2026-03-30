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
