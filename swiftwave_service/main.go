package swiftwave

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/fatih/color"
	"github.com/swiftwave-org/swiftwave/pkg/ssh_toolkit"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/dashboard"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/service_manager"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/cronjob"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/graphql"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/worker"
)

// StartSwiftwave will start the swiftwave service [including worker manager, pubsub, cronjob, server]
func StartSwiftwave(config *config.Config) {
	// Load the manager
	manager := &service_manager.ServiceManager{
		CancelImageBuildTopic: "cancel_image_build",
	}
	manager.Load(*config)

	// Set the server status validator for ssh
	ssh_toolkit.SetValidator(func(host string) bool {
		server, err := core.FetchServerByIP(&manager.DbClient, host)
		if err != nil {
			return false
		}
		return server.Status != core.ServerOffline
	})

	// Create pubsub default topics
	err := manager.PubSubClient.CreateTopic(manager.CancelImageBuildTopic)
	if err != nil {
		log.Printf("Error creating topic %s: %s", manager.CancelImageBuildTopic, err.Error())
	}

	// Create the worker manager
	workerManager := worker.NewManager(config, manager)
	err = workerManager.StartConsumers(true)
	if err != nil {
		panic(err)
	}

	// Create the cronjob manager
	cronjobManager := cronjob.NewManager(config, manager, workerManager)
	cronjobManager.Start(true)

	// create a channel to block the main thread
	waitForever := make(chan struct{})

	// StartSwiftwave the swift wave server
	go StartServer(config, manager, workerManager)
	// Wait for consumers
	go workerManager.WaitForConsumers()
	// Wait for cronjob
	go cronjobManager.Wait()

	// Block the main thread
	<-waitForever
}

func echoLogger(_ echo.Context, err error, stack []byte) error {
	color.Red("Recovered from panic: %s\n", err)
	logger.HTTPLoggerError.Println("Swiftwave server is facing error : ", err.Error(), "\n", string(stack))
	return nil
}

// StartServer starts the swiftwave graphql and rest server
func StartServer(config *config.Config, manager *service_manager.ServiceManager, workerManager *worker.Manager) {
	// Create Echo Server
	echoServer := echo.New()
	echoServer.HideBanner = true
	echoServer.Pre(middleware.RemoveTrailingSlash())
	echoServer.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		Skipper:             middleware.DefaultSkipper,
		StackSize:           4 << 10, // 4 KB
		DisableStackAll:     false,
		DisablePrintStack:   false,
		LogLevel:            0,
		LogErrorFunc:        echoLogger,
		DisableErrorHandler: false,
	}))
	echoServer.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${method} ${uri} | ${remote_ip} | ${status} ${error}\n",
	}))
	echoServer.Use(middleware.CORS())

	// Cache middleware
	// Cache JS, CSS and PNG files for 1 year, as if static content changes, the uri also changes
	// So setting cache-control header to max-age to 1 year
	// + it will also set etag header to the file name
	echoServer.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.HasSuffix(c.Request().RequestURI, ".js") || strings.HasSuffix(c.Request().RequestURI, ".css") || strings.HasSuffix(c.Request().RequestURI, ".png") || strings.HasSuffix(c.Request().RequestURI, ".ttf") {
				s := strings.Split(c.Request().RequestURI, "/")
				etag := s[len(s)-1]
				c.Response().Header().Set("Etag", etag)
				c.Response().Header().Set("Cache-Control", "max-age=31536000")
				if match := c.Request().Header.Get("If-None-Match"); match != "" {
					if strings.Contains(match, etag) {
						return c.NoContent(http.StatusNotModified)
					}
				}
			}
			return next(c)
		}
	})

	// Create GraphQL Server
	graphqlServer := graphql.Server{
		EchoServer:     echoServer,
		Config:         config,
		ServiceManager: manager,
		WorkerManager:  workerManager,
	}
	// Initialize Dashboard Web App
	dashboard.RegisterHandlers(echoServer, false)
	// Initialize GraphQL Server
	graphqlServer.Initialize()

	// Start the server
	address := fmt.Sprintf("%s:%d", config.LocalConfig.ServiceConfig.BindAddress, config.LocalConfig.ServiceConfig.BindPort)
	if config.LocalConfig.ServiceConfig.UseTLS {
		println("TLS Server Started on " + address)
		s := http.Server{
			Addr:    address,
			Handler: echoServer,
		}
		certFilePath := fmt.Sprintf("%s/certificate.crt", config.LocalConfig.ServiceConfig.SSLCertDirectoryPath)
		keyFilePath := fmt.Sprintf("%s/private.key", config.LocalConfig.ServiceConfig.SSLCertDirectoryPath)
		echoServer.Logger.Fatal(s.ListenAndServeTLS(certFilePath, keyFilePath))
	} else {
		println("Server Started on " + address)
		echoServer.Logger.Fatal(echoServer.Start(address))
	}
}
