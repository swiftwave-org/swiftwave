package swiftwave

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/cronjob"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/graphql"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/rest"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/worker"
	"github.com/swiftwave-org/swiftwave/system_config"
	"golang.org/x/crypto/acme/autocert"
	"log"
)

// Start will start the swiftwave service [including worker manager, pubsub, cronjob, server]
func Start(config *system_config.Config) {
	// Load the manager
	manager := &core.ServiceManager{}
	manager.Load(*config)

	// Create the worker manager
	workerManager := worker.NewManager(config, manager)
	err := workerManager.StartConsumers(true)
	if err != nil {
		panic(err)
	}

	// Create the cronjob manager
	cronjobManager := cronjob.NewManager(config, manager)
	cronjobManager.Start(true)

	// create a channel to block the main thread
	var waitForever chan struct{}

	// Start the swift wave server
	go StartServer(config, manager, workerManager)
	// Wait for consumers
	go workerManager.WaitForConsumers()
	// Wait for cronjob
	go cronjobManager.Wait()

	// Block the main thread
	<-waitForever
}

// StartServer starts the swiftwave graphql and rest server
func StartServer(config *system_config.Config, manager *core.ServiceManager, workerManager *worker.Manager) {
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
	if config.ServiceConfig.AutoMigrateDatabase {
		log.Println("Migrating Database")
		// Migrate Database
		err := core.MigrateDatabase(&manager.DbClient)
		if err != nil {
			panic(err)
		} else {
			log.Println("Database Migration Complete")
		}
	}
	// Start the server
	address := fmt.Sprintf("%s:%d", config.ServiceConfig.BindAddress, config.ServiceConfig.BindPort)
	if config.ServiceConfig.AutoTLS {
		echoServer.Logger.Fatal(echoServer.StartAutoTLS(address))
	} else {
		echoServer.Logger.Fatal(echoServer.Start(address))
	}
}
