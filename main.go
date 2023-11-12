package main

import (
	swiftwave "github.com/swiftwave-org/swiftwave/swiftwave_service"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/cronjob"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/worker"
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

	// Create the worker manager
	workerManager := worker.NewManager(config, &manager)
	err = workerManager.StartConsumers(true)
	if err != nil {
		panic(err)
	}

	// Create the cronjob manager
	cronjobManager := cronjob.NewManager(config, &manager)
	cronjobManager.Start(true)

	// create a channel to block the main thread
	var waitForever chan struct{}

	// Start the swift wave server
	go swiftwave.StartServer(config, &manager, workerManager, true)
	// Wait for consumers
	go workerManager.WaitForConsumers()
	// Wait for cronjobs
	go cronjobManager.Wait()

	<-waitForever
}
