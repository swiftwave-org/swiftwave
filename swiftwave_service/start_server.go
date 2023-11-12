package swiftwave

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/graphql"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/rest"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/worker"
	"github.com/swiftwave-org/swiftwave/system_config"
	"golang.org/x/crypto/acme/autocert"
)

func StartServer(config *system_config.Config, manager *core.ServiceManager, workerManager *worker.Manager, migrateDatabase bool) {
	// Create Echo Server
	echoServer := echo.New()
	echoServer.Pre(middleware.RemoveTrailingSlash())
	echoServer.Use(middleware.Recover())
	echoServer.Use(middleware.Logger())
	echoServer.Use(middleware.CORS())
	// enable host whitelist if not all domains are allowed
	if !config.ServiceConfig.IsAllDomainsAllowed() {
		echoServer.AutoTLSManager.HostPolicy = autocert.HostWhitelist(config.ServiceConfig.WhiteListedDomains...)
	}
	// Configure Auto TLS
	if config.ServiceConfig.AutoTLS {
		echoServer.AutoTLSManager.HostPolicy = autocert.HostWhitelist(config.ServiceConfig.NetworkName)
	}
	// Create Rest Server
	restServer := rest.Server{
		EchoServer:     echoServer,
		SystemConfig:   config,
		ServiceManager: manager,
		WorkerManager:  workerManager,
	}
	// Create GraphQL Server
	graphqlServer := graphql.Server{
		EchoServer:     echoServer,
		SystemConfig:   config,
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
	if config.ServiceConfig.AutoTLS {
		echoServer.Logger.Fatal(echoServer.StartAutoTLS(address))
	} else {
		echoServer.Logger.Fatal(echoServer.Start(address))
	}
}
