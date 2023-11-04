package main

import (
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	swiftwave "github.com/swiftwave-org/swiftwave/swiftwave_service"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
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

	// Create Echo Server
	echoServer := echo.New()
	// Start the swift wave server
	swiftwave.StartServer(&config, &manager, echoServer, true)
}
