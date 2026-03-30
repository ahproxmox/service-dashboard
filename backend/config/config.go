package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Proxmox struct {
		APIUrl      string `yaml:"api_url"`
		TokenId     string `yaml:"token_id"`
		TokenSecret string `yaml:"token_secret"`
	} `yaml:"proxmox"`
	Caddy struct {
		APIUrl string `yaml:"api_url"`
	} `yaml:"caddy"`
	Prometheus struct {
		Url string `yaml:"url"`
	} `yaml:"prometheus"`
	Cache struct {
		StatusTTL    time.Duration `yaml:"status_ttl"`
		MetricsTTL   time.Duration `yaml:"metrics_ttl"`
		DiscoveryTTL time.Duration `yaml:"discovery_ttl"`
	} `yaml:"cache"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}
