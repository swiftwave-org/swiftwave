package swiftwave_manager

import (
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_manager/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_manager/graphql"
	"github.com/swiftwave-org/swiftwave/swiftwave_manager/rest"
	"strconv"
)

func Load(config *core.ServiceConfig, manager *core.ServiceManager, echoServer *echo.Echo) {
	// Load Config
	config.Load()
	// Load Manager
	manager.Load()

	// Create Rest Server
	restServer := rest.Server{
		EchoServer:     echoServer,
		ServiceConfig:  config,
		ServiceManager: manager,
	}
	// Create GraphQL Server
	graphqlServer := graphql.Server{
		EchoServer:     echoServer,
		ServiceConfig:  config,
		ServiceManager: manager,
	}
	// Initialize Rest Server
	restServer.Initialize()
	// Initialize GraphQL Server
	graphqlServer.Initialize()
}

func StartServer(config *core.ServiceConfig, manager *core.ServiceManager, echoServer *echo.Echo) {
	// Migrate Database by default
	core.MigrateDatabase(&manager.DbClient)
	// Start the server
	echoServer.Logger.Fatal(echoServer.Start(":" + strconv.Itoa(config.Port)))
}
