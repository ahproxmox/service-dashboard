package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a test config file
	configYAML := `
server:
  port: 8080
proxmox:
  api_url: https://192.168.88.5:8006
  token_id: user@pam!token
  token_secret: secret123
caddy:
  api_url: http://192.168.88.82:2019
prometheus:
  url: http://192.168.88.73:9090
cache:
  status_ttl: 2s
  metrics_ttl: 25s
  discovery_ttl: 10s
`
	tmpFile := t.TempDir() + "/test-config.yaml"
	_ = os.WriteFile(tmpFile, []byte(configYAML), 0644)

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Proxmox.APIUrl != "https://192.168.88.5:8006" {
		t.Errorf("expected Proxmox URL, got %s", cfg.Proxmox.APIUrl)
	}
	if cfg.Cache.StatusTTL.Seconds() != 2 {
		t.Errorf("expected status TTL 2s, got %v", cfg.Cache.StatusTTL)
	}
}
