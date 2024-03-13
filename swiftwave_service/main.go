package swiftwave

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/dashboard"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/service_manager"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/cronjob"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/graphql"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/rest"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/worker"
)

// StartSwiftwave will start the swiftwave service [including worker manager, pubsub, cronjob, server]
func StartSwiftwave(config *config.Config) {
	// Load the manager
	manager := &service_manager.ServiceManager{
		CancelImageBuildTopic: "cancel_image_build",
	}
	manager.Load(*config)

	// Cancel pending tasks
	// TODO: don't cancel, requeue the tasks
	err := worker.CancelPendingTasksLocalQueue(*config, *manager)
	if err != nil {
		panic(err)
	}

	// Create pubsub default topics
	err = manager.PubSubClient.CreateTopic(manager.CancelImageBuildTopic)
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
	cronjobManager := cronjob.NewManager(config, manager)
	cronjobManager.Start(true)

	// create a channel to block the main thread
	var waitForever chan struct{}

	// StartSwiftwave the swift wave server
	go StartServer(config, manager, workerManager)
	// Wait for consumers
	go workerManager.WaitForConsumers()
	// Wait for cronjob
	go cronjobManager.Wait()

	// Block the main thread
	<-waitForever
}

// StartServer starts the swiftwave graphql and rest server
func StartServer(config *config.Config, manager *service_manager.ServiceManager, workerManager *worker.Manager) {
	// Create Echo Server
	echoServer := echo.New()
	echoServer.HideBanner = true
	echoServer.Pre(middleware.RemoveTrailingSlash())
	echoServer.Use(middleware.Recover())
	echoServer.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${method} ${uri} | ${remote_ip} | ${status} ${error}\n",
	}))
	echoServer.Use(middleware.CORS())
	// JWT Middleware
	echoServer.Use(echojwt.WithConfig(echojwt.Config{
		Skipper: func(c echo.Context) bool {
			if strings.Compare(c.Request().URL.Path, "/") == 0 ||
				strings.HasPrefix(c.Request().URL.Path, "/healthcheck") ||
				strings.HasPrefix(c.Request().URL.Path, "/.well-known") ||
				strings.HasPrefix(c.Request().URL.Path, "/auth") ||
				strings.HasPrefix(c.Request().URL.Path, "/webhook") ||
				strings.HasPrefix(c.Request().URL.Path, "/dashboard") ||
				strings.HasPrefix(c.Request().URL.Path, "/playground") {
				return true
			}
			// check if a GET request at /graphql and a websocket upgrade request
			if strings.HasPrefix(c.Request().URL.Path, "/graphql") &&
				strings.Compare(c.Request().Method, http.MethodGet) == 0 &&
				strings.Compare(c.Request().URL.RawQuery, "") == 0 &&
				strings.Contains(c.Request().Header.Get("Connection"), "Upgrade") &&
				strings.Compare(c.Request().Header.Get("Upgrade"), "websocket") == 0 {
				return true
			}
			return false
		},
		SigningKey: []byte(config.SystemConfig.JWTSecretKey),
		ContextKey: "jwt_data",
	}))
	// Authorization Middleware
	// Add `authorized` & `username` key to the context
	echoServer.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, ok := c.Get("jwt_data").(*jwt.Token)
			ctx := c.Request().Context()
			if !ok {
				c.Set("authorized", false)
				c.Set("username", "")
				//nolint:staticcheck
				ctx = context.WithValue(ctx, "authorized", false)
				//nolint:staticcheck
				ctx = context.WithValue(ctx, "username", "")
			} else {
				claims := token.Claims.(jwt.MapClaims)
				username := claims["username"].(string)
				c.Set("authorized", true)
				c.Set("username", username)
				//nolint:staticcheck
				ctx = context.WithValue(ctx, "authorized", true)
				//nolint:staticcheck
				ctx = context.WithValue(ctx, "username", username)
			}
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	})
	// Create Rest Server
	restServer := rest.Server{
		EchoServer:     echoServer,
		Config:         config,
		ServiceManager: manager,
		WorkerManager:  workerManager,
	}
	// Create GraphQL Server
	graphqlServer := graphql.Server{
		EchoServer:     echoServer,
		Config:         config,
		ServiceManager: manager,
		WorkerManager:  workerManager,
	}
	// Initialize Dashboard Web App
	dashboard.RegisterHandlers(echoServer)
	// Initialize Rest Server
	restServer.Initialize()
	// Initialize GraphQL Server
	graphqlServer.Initialize()

	// StartSwiftwave the server
	address := fmt.Sprintf("%s:%d", config.LocalConfig.ServiceConfig.BindAddress, config.LocalConfig.ServiceConfig.BindPort)
	if config.LocalConfig.ServiceConfig.UseTLS {
		println("TLS Server Started on " + address)

		tlsCfg := &tls.Config{
			Certificates: fetchCertificates(config.LocalConfig.ServiceConfig.SSLCertDirectoryPath),
		}

		s := http.Server{
			Addr:      address,
			Handler:   echoServer,
			TLSConfig: tlsCfg,
		}
		echoServer.Logger.Fatal(s.ListenAndServeTLS("", ""))
	} else {
		echoServer.Logger.Fatal(echoServer.Start(address))
	}
}

// private functions
func fetchCertificates(certFolderPath string) []tls.Certificate {
	var certificates []tls.Certificate
	// fetch all folders in the cert folder
	files, err := os.ReadDir(certFolderPath)
	if err != nil {
		return certificates
	}
	for _, file := range files {
		if file.IsDir() {
			// fetch the certificate
			cert, err := tls.LoadX509KeyPair(fmt.Sprintf("%s/%s/certificate.crt", certFolderPath, file.Name()), fmt.Sprintf("%s/%s/private.key", certFolderPath, file.Name()))
			if err != nil {
				log.Println("Error loading certificate: ", err)
				continue
			}
			certificates = append(certificates, cert)
		}
	}
	return certificates
}
