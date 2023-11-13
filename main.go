package main

import (
<<<<<<< HEAD
	"fmt"
	"github.com/fatih/color"
	"github.com/swiftwave-org/swiftwave/cmd"
	"github.com/swiftwave-org/swiftwave/system_config"
=======
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	swiftwave "github.com/swiftwave-org/swiftwave/swiftwave_service"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/cronjob"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/worker"
	"github.com/swiftwave-org/swiftwave/system_config"
	"log"
>>>>>>> 5f6e33e0fb2a7d5fd0d52314aef4a850df72ec56
	"os"
)

func main() {
<<<<<<< HEAD
	// ensure program is run as root
	if os.Geteuid() != 0 {
		color.Red("This program must be run as root. Aborting.")
		os.Exit(1)
	}
	var config *system_config.Config
	var err error
	// Check whether first argument is "install" or no arguments
	if (len(os.Args) > 1 && (os.Args[1] == "install" || os.Args[1] == "init" || os.Args[1] == "config" || os.Args[1] == "completion" || os.Args[1] == "--help")) ||
		len(os.Args) == 1 {
		config = nil
	} else {
		// Load config path from environment variable
		systemConfigPath := "/etc/swiftwave/config.yaml"
		// Load the config
		config, err = system_config.ReadFromFile(systemConfigPath)
		if err != nil {
			fmt.Println("failed to load config file > ", err)
			os.Exit(1)
		}
	}

	// Start the command line interface
	cmd.Execute(config)
=======
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

	// Create Echo Server
	echoServer := echo.New()
	echoServer.Pre(middleware.RemoveTrailingSlash())
	echoServer.Use(middleware.Recover())
	echoServer.Use(middleware.CORS())

	// Start the swift wave server
	go swiftwave.StartServer(config, &manager, echoServer, workerManager, true)
	// Wait for consumers
	go workerManager.WaitForConsumers()
	// Wait for cronjobs
	go cronjobManager.Wait()

	<-waitForever
>>>>>>> 5f6e33e0fb2a7d5fd0d52314aef4a850df72ec56
}
