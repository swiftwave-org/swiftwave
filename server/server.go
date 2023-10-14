package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	DOCKER "swiftwave/m/container_manager"
	DOCKER_CONFIG_GENERATOR "swiftwave/m/docker_config_generator"
	HAPROXY "swiftwave/m/haproxy_manager"
	SSL "swiftwave/m/ssl_manager"
	"time"

	DOCKER_CLIENT "github.com/docker/docker/client"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/redisq"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Initialize all the components of the server
func (server *Server) Init() {
	server_port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal("PORT environment variable is not set")
		panic(err)
	}
	server.PORT = server_port
	server.CODE_TARBALL_DIR = os.Getenv("CODE_TARBALL_DIR")
	server.SWARM_NETWORK = os.Getenv("SWARM_NETWORK")
	server.HAPROXY_SERVICE = os.Getenv("HAPROXY_SERVICE_NAME")
	server.ENVIRONMENT = os.Getenv("ENVIRONMENT")
	if server.ENVIRONMENT == "" {
		server.ENVIRONMENT = "production"
	}
	restricted_ports_str := os.Getenv("RESTRICTED_PORTS")
	restricted_ports_str_split := strings.Split(restricted_ports_str, ",")
	restricted_ports := []int{}
	for _, port := range restricted_ports_str_split {
		port_int, err := strconv.Atoi(string(port))
		if err != nil {
			panic(err)
		}
		restricted_ports = append(restricted_ports, port_int)
	}
	server.RESTRICTED_PORTS = restricted_ports
	token_expiry_minutes, err := strconv.Atoi(os.Getenv("SESSION_TOKEN_EXPIRY_MINUTES"))
	if err != nil {
		panic(err)
	}
	server.WEBSOCKET_TOKENS = make(map[string]time.Time)
	server.WEBSOCKET_TOKEN_EXPIRY_MINUTES = token_expiry_minutes // have different expiry for websocket tokens
	server.SESSION_TOKENS = make(map[string]time.Time)
	server.SESSION_TOKEN_EXPIRY_MINUTES = token_expiry_minutes
	// Initiating database client
	db_type := os.Getenv("DATABASE_TYPE")
	var db_dialect gorm.Dialector
	if db_type == "postgres" {
		db_dialect = postgres.Open(os.Getenv("POSTGRESQL_URI"))
	} else if db_type == "sqlite" {
		db_dialect = sqlite.Open(os.Getenv("SQLITE_DATABASE"))
	} else {
		panic("Unknown database type")
	}
	db_client, err := gorm.Open(db_dialect, &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Initiating Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0, // use default DB
	})
	server.REDIS_CLIENT = *rdb

	// Initiating SSL Manager
	options := SSL.ManagerOptions{
		IsStaging:                 !server.isProductionEnvironment(),
		Email:                     os.Getenv("ACCOUNT_EMAIL_ID"),
		AccountPrivateKeyFilePath: os.Getenv("ACCOUNT_PRIVATE_KEY_FILE_PATH"),
	}
	ssl_manager := SSL.Manager{}
	err = ssl_manager.Init(context.Background(), *db_client, options)
	if err != nil {
		panic(err)
	}

	// Initiating HAPROXY Manager
	var haproxy_manager = HAPROXY.Manager{}
	haproxy_port, err := strconv.Atoi(os.Getenv("HAPROXY_MANAGER_PORT"))
	if err != nil {
		log.Fatal("HAPROXY_MANAGER_PORT environment variable is not set")
		panic(err)
	}
	haproxy_manager.InitTcpSocket(os.Getenv("HAPROXY_MANAGER_HOST"), haproxy_port)
	haproxy_manager.Auth(os.Getenv("HAPROXY_MANAGER_USERNAME"), os.Getenv("HAPROXY_MANAGER_PASSWORD"))

	// Initiating Docker Manager
	docker_client, err := DOCKER_CLIENT.NewClientWithOpts(DOCKER_CLIENT.WithHost(os.Getenv("DOCKER_HOST")))
	if err != nil {
		panic(err)
	}
	docker_manager := DOCKER.Manager{}
	err = docker_manager.Init(context.Background(), *docker_client)
	if err != nil {
		panic(err)
	}

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
	server.WEBSOCKET_UPGRADER = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// TODO: allow to set origin via config
			return true
		},
	}

	// Set middlewares
	server.ECHO_SERVER.Pre(middleware.RemoveTrailingSlash())
	server.ECHO_SERVER.Use(middleware.CORS())
	server.ECHO_SERVER.Pre(server.authMiddleware)

	// Migrating database
	server.MigrateDatabaseTables()

	// Initiating Routes for ACME Challenge
	server.SSL_MANAGER.InitHttpHandlers(&server.ECHO_SERVER)

	// Worker related
	server.WORKER_CONTEXT, server.WORKER_CONTEXT_CANCEL = context.WithCancel(context.Background())
	server.QUEUE_FACTORY = redisq.NewFactory()
	// Registering main queue to push tasks
	server.TASK_QUEUE = server.QUEUE_FACTORY.RegisterQueue(&taskq.QueueOptions{
		Name:  "main-queue",
		Redis: &server.REDIS_CLIENT,
	})
	// Map of task name to task
	server.TASK_MAP = make(map[string]*taskq.Task)
	// Registering worker tasks
	server.RegisteWorkerTasks()
}

// Start server
func (server *Server) Start() {
	// Initiating REST API
	server.InitAuthRestAPI()
	server.InitDomainRestAPI()
	server.InitApplicationRestAPI()
	server.InitTestRestAPI()
	server.InitGitRestAPI()
	server.InitIngressRestAPI()
	server.InitRedirectRestAPI()
	server.InitPersistentVolumeAPI()

	// Create default git user
	server.CreateDefaultGitUser()

	// Start worker consumers
	err := server.StartWorkerConsumers()
	if err != nil {
		panic(err)
	}

	// Cron related
	server.InitCronJobs()

	// Starting server
	server.ECHO_SERVER.Logger.Fatal(server.ECHO_SERVER.Start(":" + strconv.Itoa(server.PORT)))
}
