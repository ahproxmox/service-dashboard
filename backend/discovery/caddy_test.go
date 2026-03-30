package discovery

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCaddyGetRoutes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"handle": [
					{
						"upstreams": [
							{"dial": "192.168.88.78:3000"}
						]
					}
				],
				"match": [
					{"host": ["kanban.internal.ahproxmox-claude.cc"]}
				]
			},
			{
				"handle": [
					{
						"upstreams": [
							{"dial": "192.168.88.71:8080"}
						]
					}
				],
				"match": [
					{"host": ["rag.internal.ahproxmox-claude.cc"]}
				]
			}
		]`))
	}))
	defer server.Close()

	client := NewCaddyClient(server.URL)
	routes, err := client.GetRoutes()
	if err != nil {
		t.Fatalf("GetRoutes failed: %v", err)
	}

	if len(routes) != 2 {
		t.Errorf("expected 2 routes, got %d", len(routes))
	}

	if routes[0].Domain != "kanban.internal.ahproxmox-claude.cc" {
		t.Errorf("expected kanban domain, got %s", routes[0].Domain)
	}

	if routes[0].BackendIp != "192.168.88.78" {
		t.Errorf("expected IP 192.168.88.78, got %s", routes[0].BackendIp)
	}

	if routes[1].Domain != "rag.internal.ahproxmox-claude.cc" {
		t.Errorf("expected rag domain, got %s", routes[1].Domain)
	}

	if routes[1].BackendIp != "192.168.88.71" {
		t.Errorf("expected IP 192.168.88.71, got %s", routes[1].BackendIp)
	}
}

func TestCaddyGetRoutesStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewCaddyClient(server.URL)
	_, err := client.GetRoutes()
	if err == nil {
		t.Error("expected error for non-200 status")
	}
}

func TestCaddyGetRoutesMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	client := NewCaddyClient(server.URL)
	_, err := client.GetRoutes()
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestCaddyGetRoutesSkipsInvalidRoutes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{"invalid": "structure"},
			{
				"handle": [{"upstreams": [{"dial": "192.168.88.78:3000"}]}],
				"match": [{"host": ["kanban.internal.ahproxmox-claude.cc"]}]
			},
			{
				"handle": [{}],
				"match": [{"host": ["missing.internal.ahproxmox-claude.cc"]}]
			}
		]`))
	}))
	defer server.Close()

	client := NewCaddyClient(server.URL)
	routes, err := client.GetRoutes()
	if err != nil {
		t.Fatalf("GetRoutes failed: %v", err)
	}

	if len(routes) != 1 {
		t.Errorf("expected 1 valid route (skipped 2 invalid), got %d", len(routes))
	}

	if routes[0].Domain != "kanban.internal.ahproxmox-claude.cc" {
		t.Errorf("expected kanban domain, got %s", routes[0].Domain)
	}
}
