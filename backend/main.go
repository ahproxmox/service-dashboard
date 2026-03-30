package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ahproxmox/service-dashboard/backend/config"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/etc/service-dashboard/config.yaml"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Service Dashboard loaded config from %s\n", configPath)
	fmt.Printf("Server listening on port %d\n", cfg.Server.Port)
}
