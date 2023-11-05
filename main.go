package main

import (
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	swiftwave "github.com/swiftwave-org/swiftwave/swiftwave_service"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/worker"
	"log"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("WARN: error loading .env file. Ignoring")
	}
	// Load the manager
	config := core.ServiceConfig{}
	manager := core.ServiceManager{}
	config.Load()
	manager.Load()

	// Create the worker manager
	workerManager := worker.NewManager(&config, &manager)
	err = workerManager.StartConsumers(true)
	if err != nil {
		panic(err)
	}

	// create a channel to block the main thread
	var waitForever chan struct{}

	// Create Echo Server
	echoServer := echo.New()
	echoServer.Pre(middleware.RemoveTrailingSlash())
	echoServer.Use(middleware.Logger())
	echoServer.Use(middleware.Recover())
	echoServer.Use(middleware.CORS())

	// Start the swift wave server
	go swiftwave.StartServer(&config, &manager, echoServer, workerManager, true)
	// Wait for consumers
	go workerManager.WaitForConsumers()

	<-waitForever
}
