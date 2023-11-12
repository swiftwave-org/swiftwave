package swiftwave

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/graphql"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/rest"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/worker"
	"github.com/swiftwave-org/swiftwave/system_config"
)

func StartServer(config *system_config.Config, manager *core.ServiceManager, echoServer *echo.Echo, workerManager *worker.Manager, migrateDatabase bool) {
	// Create Rest Server
	restServer := rest.Server{
		EchoServer:     echoServer,
		ServiceConfig:  config,
		ServiceManager: manager,
		WorkerManager:  workerManager,
	}
	// Create GraphQL Server
	graphqlServer := graphql.Server{
		EchoServer:     echoServer,
		ServiceConfig:  config,
		ServiceManager: manager,
		WorkerManager:  workerManager,
	}
	// Initialize Rest Server
	restServer.Initialize()
	// Initialize GraphQL Server
	graphqlServer.Initialize()
	if migrateDatabase {
		// Migrate Database
		core.MigrateDatabase(&manager.DbClient)
	}
	// Start the server
	address := fmt.Sprintf("%s:%d", config.ServiceConfig.BindAddress, config.ServiceConfig.BindPort)
	echoServer.Logger.Fatal(echoServer.Start(address))
}
