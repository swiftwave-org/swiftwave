package main

import (
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_manager/core"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("WARN: error loading .env file. Ignoring")
	}
	// Load the manager
	config := core.ServiceConfig{}
	manager := core.ServiceManager{}
	echoServer := echo.New()
	swiftwave_manager.Load(&config, &manager, echoServer)
	// Start the manager
	swiftwave_manager.StartServer(&config, &manager, echoServer)
}
