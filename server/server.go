package server

import (
	"context"
	DOCKER "keroku/m/container_manager"
	DOCKER_CONFIG_GENERATOR "keroku/m/docker_config_generator"
	HAPROXY "keroku/m/haproxy_manager"
	SSL "keroku/m/ssl_manager"
	"strconv"

	DOCKER_CLIENT "github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Server struct
type Server struct {
	SSL_MANAGER             SSL.Manager
	HAPROXY_MANAGER         HAPROXY.Manager
	DOCKER_MANAGER          DOCKER.Manager
	DOCKER_CONFIG_GENERATOR DOCKER_CONFIG_GENERATOR.Manager
	DOCKER_CLIENT           DOCKER_CLIENT.Client
	DB_CLIENT               gorm.DB
	ECHO_SERVER             echo.Echo
	PORT                    int
}

// Init function
func (server *Server) Init(port int) {
	server.PORT = port
	// Initiating database client
	db_client, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Initiating SSL Manager
	options := SSL.ManagerOptions{
		IsStaging:                 false,
		Email:                     "tanmoysrt@gmail.com",
		AccountPrivateKeyFilePath: "/home/ubuntu/client_program/data/account_private_key.key",
		DomainPrivateKeyStorePath: "/home/ubuntu/client_program/data/domain/private_key",
		DomainFullChainStorePath:  "/home/ubuntu/client_program/data/domain/full_chain",
	}
	ssl_manager := SSL.Manager{}
	ssl_manager.Init(context.Background(), *db_client, options)

	// Initiating HAPROXY Manager
	var haproxy_manager = HAPROXY.Manager{}
	haproxy_manager.InitTcpSocket("localhost", 5555)
	haproxy_manager.Auth("admin", "mypassword")

	// Initiating Docker Manager
	docker_client, err := DOCKER_CLIENT.NewClientWithOpts(DOCKER_CLIENT.WithHost("tcp://127.0.0.1:2375"))
	if err != nil {
		panic(err)
	}
	docker_manager := DOCKER.Manager{}
	docker_manager.Init(context.Background(), *docker_client)

	// Initiating Docker Image Manager
	docker_config_generator := DOCKER_CONFIG_GENERATOR.Manager{}
	err = docker_config_generator.Init()
	if err != nil {
		panic(err)
	}

	// Assigning to server
	server.SSL_MANAGER = ssl_manager
	server.HAPROXY_MANAGER = haproxy_manager
	server.DOCKER_MANAGER = docker_manager
	server.DOCKER_CONFIG_GENERATOR = docker_config_generator
	server.DOCKER_CLIENT = *docker_client
	server.DB_CLIENT = *db_client
	server.ECHO_SERVER = *echo.New()

	// Migrating database
	server.MigrateDomainDB()

	// Initiating REST API
	server.InitDomainRestAPI()
	server.InitTestRestAPI()

	// Initiating Routes for ACME Challenge
	server.SSL_MANAGER.InitHttpHandlers(&server.ECHO_SERVER)
}

// Start server
func (server *Server) Start() {
	server.ECHO_SERVER.Logger.Fatal(server.ECHO_SERVER.Start(":"+strconv.Itoa(server.PORT)))
}
