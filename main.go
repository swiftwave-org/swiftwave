package main

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/system_config"
	"log"
	"os"
)

func main() {
	// Load config path from environment variable
	systemConfigPath := os.Getenv("SWIFTWAVE_CONFIG_PATH")
	if systemConfigPath == "" {
		systemConfigPath = "~/swiftwave/config.yaml"
		log.Println("SWIFTWAVE_CONFIG_PATH environment variable not set, using default path > ", systemConfigPath)
	}
	// Load the config
	config, err := system_config.ReadFromFile(systemConfigPath)
	if err != nil {
		panic(err)
	}
	// Load the manager
	manager := core.ServiceManager{}
	manager.Load(*config)
}
