package discovery

import (
	"testing"
)

func TestMatchServiceByIp(t *testing.T) {
	containers := []Container{
		{Id: "122", Name: "kanban", Status: "running", Ip: "192.168.88.78"},
		{Id: "111", Name: "rag", Status: "running", Ip: "192.168.88.71"},
	}

	routes := []Route{
		{Domain: "kanban.internal.ahproxmox-claude.cc", BackendIp: "192.168.88.78"},
		{Domain: "rag.internal.ahproxmox-claude.cc", BackendIp: "192.168.88.71"},
	}

	matcher := NewMatcher()
	services := matcher.Match(containers, routes)

	if len(services) != 2 {
		t.Errorf("expected 2 services, got %d", len(services))
	}

	if services[0].HttpsUrl == nil || *services[0].HttpsUrl != "https://kanban.internal.ahproxmox-claude.cc" {
		t.Errorf("expected kanban URL, got %v", services[0].HttpsUrl)
	}

	if services[1].HttpsUrl == nil || *services[1].HttpsUrl != "https://rag.internal.ahproxmox-claude.cc" {
		t.Errorf("expected rag URL, got %v", services[1].HttpsUrl)
	}
}

func TestMatchServiceNoRoute(t *testing.T) {
	containers := []Container{
		{Id: "104", Name: "openclaw", Status: "stopped", Ip: "192.168.88.63"},
	}

	matcher := NewMatcher()
	services := matcher.Match(containers, []Route{})

	if len(services) != 1 {
		t.Errorf("expected 1 service, got %d", len(services))
	}

	if services[0].HttpsUrl != nil {
		t.Error("expected no HTTPS URL for unmatched service")
	}

	if services[0].Status != "stopped" {
		t.Errorf("expected stopped status, got %s", services[0].Status)
	}
}
