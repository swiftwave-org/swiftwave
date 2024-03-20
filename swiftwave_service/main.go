package swiftwave

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/fatih/color"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/console"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
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

func echoLogger(_ echo.Context, err error, stack []byte) error {
	color.Red("Recovered from panic: %s\n", err)
	fmt.Println(string(stack))
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

	// Internal Service Authentication Middleware
	// Authorization : analytics_token <analytics_id>:<analytics_token>
	// Only for /service/analytics endpoints
	echoServer.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.Compare(c.Request().URL.Path, "/service/analytics") == 0 {
				authorization := c.Request().Header.Get("Authorization")
				if strings.HasPrefix(authorization, "analytics_token ") {
					token := strings.TrimPrefix(authorization, "analytics_token ")
					tokenParts := strings.Split(token, ":")
					if len(tokenParts) == 2 {
						verified, serverHostName, err := core.ValidateAnalyticsServiceToken(c.Request().Context(), manager.DbClient, tokenParts[0], tokenParts[1])
						if err != nil {
							return c.JSON(http.StatusUnauthorized, map[string]interface{}{
								"message": "invalid service token",
							})
						}
						if verified {
							c.Set("authorized", true)
							c.Set("hostname", serverHostName)
						} else {
							return c.JSON(http.StatusUnauthorized, map[string]interface{}{
								"message": "invalid service token",
							})
						}
					}
				} else {
					c.Set("authorized", false)
					c.Set("hostname", "")
				}
			}
			return next(c)
		}
	})

	// JWT Middleware
	echoServer.Use(echojwt.WithConfig(echojwt.Config{
		Skipper: func(c echo.Context) bool {
			// check if request is already authorized
			if strings.HasPrefix(c.Request().URL.Path, "/service/analytics") &&
				c.Get("authorized") != nil && c.Get("hostname") != nil {
				if c.Get("authorized").(bool) && strings.Compare(c.Get("hostname").(string), "") != 0 {
					return true
				}
			}
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

			// on console websocket connection allow without jwt, as auth will be handled by the console server
			if strings.HasPrefix(c.Request().URL.Path, "/console/ws") &&
				strings.Compare(c.Request().Method, http.MethodGet) == 0 &&
				strings.Compare(c.Request().URL.RawQuery, "") == 0 &&
				strings.Contains(c.Request().Header.Get("Connection"), "Upgrade") &&
				strings.Compare(c.Request().Header.Get("Upgrade"), "websocket") == 0 {
				return true
			}

			// Whitelist console's HTML, JS, CSS
			if (strings.Compare(c.Request().URL.Path, "/console") == 0 ||
				strings.Compare(c.Request().URL.Path, "/console/main.js") == 0 ||
				strings.Compare(c.Request().URL.Path, "/console/xterm.js") == 0 ||
				strings.Compare(c.Request().URL.Path, "/console/xterm-addon-fit.js") == 0 ||
				strings.Compare(c.Request().URL.Path, "/console/xterm.css") == 0) &&
				strings.Compare(c.Request().Method, http.MethodGet) == 0 {
				return true
			}

			return false
		},
		SigningKey: []byte(config.SystemConfig.JWTSecretKey),
		ContextKey: "jwt_data",
	}))

	// Add `authorized` & `username` key to the context
	echoServer.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// ignore if already authorized
			if c.Get("authorized") != nil && c.Get("hostname") != nil {
				if c.Get("authorized").(bool) && strings.Compare(c.Get("hostname").(string), "") != 0 {
					return next(c)
				}
			}
			token, ok := c.Get("jwt_data").(*jwt.Token)
			ctx := c.Request().Context()
			if !ok {
				c.Set("authorized", false)
				c.Set("username", "")
				c.Set("hostname", "")
				//nolint:staticcheck
				ctx = context.WithValue(ctx, "authorized", false)
				//nolint:staticcheck
				ctx = context.WithValue(ctx, "username", "")
			} else {
				claims := token.Claims.(jwt.MapClaims)
				username := claims["username"].(string)
				c.Set("authorized", true)
				c.Set("username", username)
				c.Set("hostname", "")
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
	// Create Console Server (Server + Deployed Applications Remote Shell)
	consoleServer := console.Server{
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
	dashboard.RegisterHandlers(echoServer, false)
	// Initialize Rest Server
	restServer.Initialize()
	// Initialize Console Server
	consoleServer.Initialize()
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
