package server

import (
	"context"
	"log"
	"os"
	"strconv"
	DOCKER "swiftwave/m/container_manager"
	DOCKER_CONFIG_GENERATOR "swiftwave/m/docker_config_generator"
	HAPROXY "swiftwave/m/haproxy_manager"
	SSL "swiftwave/m/ssl_manager"

	DOCKER_CLIENT "github.com/docker/docker/client"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/redisq"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/postgres"
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
	REDIS_CLIENT            redis.Client
	ECHO_SERVER             echo.Echo
	PORT                    int
	HAPROXY_SERVICE         string
	CODE_TARBALL_DIR        string
	SWARM_NETWORK           string
	RESTRICTED_PORTS        []int
	// Worker related
	QUEUE_FACTORY         taskq.Factory
	TASK_QUEUE            taskq.Queue
	TASK_MAP              map[string]*taskq.Task
	WORKER_CONTEXT        context.Context
	WORKER_CONTEXT_CANCEL context.CancelFunc
}

// Init function
func (server *Server) Init() {
	server_port , err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal("PORT environment variable is not set")
		panic(err)
	}
	server.PORT = server_port
	server.CODE_TARBALL_DIR = os.Getenv("CODE_TARBALL_DIR")
	server.SWARM_NETWORK = os.Getenv("SWARM_NETWORK")
	server.HAPROXY_SERVICE = os.Getenv("HAPROXY_SERVICE_NAME")
	restricted_ports_str := os.Getenv("RESTRICTED_PORTS")
	restricted_ports := []int{}
	for _, port := range restricted_ports_str {
		port_int, err := strconv.Atoi(string(port))
		if err != nil {
			panic(err)
		}
		restricted_ports = append(restricted_ports, port_int)
	}
	server.RESTRICTED_PORTS = restricted_ports
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
		DB:       0,  // use default DB
	})
	server.REDIS_CLIENT = *rdb

	// Initiating SSL Manager
	options := SSL.ManagerOptions{
		IsStaging:                 false,
		Email:                     os.Getenv("ACCOUNT_EMAIL_ID"),
		AccountPrivateKeyFilePath: os.Getenv("ACCOUNT_PRIVATE_KEY_FILE_PATH"),
	}
	ssl_manager := SSL.Manager{}
	ssl_manager.Init(context.Background(), *db_client, options)

	// Initiating HAPROXY Manager
	var haproxy_manager = HAPROXY.Manager{}
	haproxy_port , err := strconv.Atoi(os.Getenv("HAPROXY_MANAGER_PORT"))
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
	server.MigrateDatabaseTables()

	// Initiating Routes for ACME Challenge
	server.SSL_MANAGER.InitHttpHandlers(&server.ECHO_SERVER)

	// Worker related
	server.WORKER_CONTEXT, server.WORKER_CONTEXT_CANCEL = context.WithCancel(context.Background())
	server.QUEUE_FACTORY = redisq.NewFactory()
	server.TASK_QUEUE = server.QUEUE_FACTORY.RegisterQueue(&taskq.QueueOptions{
		Name:  "main-queue",
		Redis: &server.REDIS_CLIENT,
	})
	server.TASK_MAP = make(map[string]*taskq.Task)
	// Registering worker tasks
	server.RegisteWorkerTasks()
}

// Start server
func (server *Server) Start() {
	// Initiating REST API
	server.InitDomainRestAPI()
	server.InitApplicationRestAPI()
	server.InitTestRestAPI()
	server.InitGitRestAPI()
	server.InitIngressRestAPI()
	server.InitRedirectRestAPI()

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
