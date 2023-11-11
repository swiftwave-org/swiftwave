package core

import (
	"context"
	DOCKER_CLIENT "github.com/docker/docker/client"
	DOCKER "github.com/swiftwave-org/swiftwave/container_manager"
	DOCKER_CONFIG_GENERATOR "github.com/swiftwave-org/swiftwave/docker_config_generator"
	HAPROXY "github.com/swiftwave-org/swiftwave/haproxy_manager"
	"github.com/swiftwave-org/swiftwave/pubsub"
	SSL "github.com/swiftwave-org/swiftwave/ssl_manager"
	"github.com/swiftwave-org/swiftwave/system_config"
	"github.com/swiftwave-org/swiftwave/task_queue"
)

func (manager *ServiceManager) Load(config system_config.Config) {
	// Initiating database client
	dbClient, err := createDbClient(config.PostgresqlConfig.DSN())
	if err != nil {
		panic(err.Error())
	}
	manager.DbClient = *dbClient

	// Initiating SSL Manager
	options := SSL.ManagerOptions{
		IsStaging:                 config.LetsEncryptConfig.StagingEnvironment,
		Email:                     config.LetsEncryptConfig.EmailID,
		AccountPrivateKeyFilePath: config.LetsEncryptConfig.PrivateKeyPath,
	}
	sslManager := SSL.Manager{}
	err = sslManager.Init(context.Background(), *dbClient, options)
	if err != nil {
		panic(err)
	}
	manager.SslManager = sslManager

	// Initiating HAPROXY Manager
	var haproxyManager = HAPROXY.Manager{}
	haproxyManager.InitUnixSocket(config.HAProxyConfig.UnixSocketPath)
	haproxyManager.Auth(config.HAProxyConfig.User, config.HAProxyConfig.Password)
	manager.HaproxyManager = haproxyManager

	// Initiating Docker Manager
	dockerClient, err := DOCKER_CLIENT.NewClientWithOpts(DOCKER_CLIENT.WithHost("unix://"+config.ServiceConfig.DockerUnixSocketPath), DOCKER_CLIENT.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	dockerManager := DOCKER.Manager{}
	err = dockerManager.Init(context.Background(), *dockerClient)
	if err != nil {
		panic(err)
	}
	manager.DockerManager = dockerManager

	// Initiating Docker Config Generator
	dockerConfigGenerator := DOCKER_CONFIG_GENERATOR.Manager{}
	err = dockerConfigGenerator.Init()
	if err != nil {
		panic(err)
	}
	manager.DockerConfigGenerator = dockerConfigGenerator

	// TODO based on configuration use remote or local redis
	pubSubClient, err := pubsub.NewClient(pubsub.Options{
		Type:         pubsub.Local,
		BufferLength: 1000,
		RedisClient:  nil,
	})
	if err != nil {
		panic(err)
	}
	manager.PubSubClient = pubSubClient

	taskQueueClient, err := task_queue.NewClient(task_queue.Options{
		Type:                task_queue.Local,
		Mode:                task_queue.Both,
		MaxMessagesPerQueue: 1000,
	})
	if err != nil {
		panic(err)
	}
	manager.TaskQueueClient = taskQueueClient

}
