package core

import (
	"context"
	"fmt"
	local_config2 "github.com/swiftwave-org/swiftwave/swiftwave_service/config/local_config"
	UDP_PROXY "github.com/swiftwave-org/swiftwave/udp_proxy_manager"

	"github.com/go-redis/redis/v8"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	DOCKER_CONFIG_GENERATOR "github.com/swiftwave-org/swiftwave/docker_config_generator"
	HAPROXY "github.com/swiftwave-org/swiftwave/haproxy_manager"
	"github.com/swiftwave-org/swiftwave/local_config"
	"github.com/swiftwave-org/swiftwave/pubsub"
	SSL "github.com/swiftwave-org/swiftwave/ssl_manager"
	"github.com/swiftwave-org/swiftwave/task_queue"
)

func (manager *ServiceManager) Load(config local_config2.Config) {
	// Initiating database client
	dbClient, err := createDbClient(config.PostgresqlConfig.DSN())
	if err != nil {
		panic(err.Error())
	}
	manager.DbClient = *dbClient

	// Initiating SSL Manager
	options := SSL.ManagerOptions{
		IsStaging:         config.LetsEncryptConfig.StagingEnvironment,
		Email:             config.LetsEncryptConfig.EmailID,
		AccountPrivateKey: config.LetsEncryptConfig.AccountPrivateKeyPath,
	}
	sslManager := SSL.Manager{}
	err = sslManager.Init(context.Background(), *dbClient, options)
	if err != nil {
		panic(err)
	}
	manager.SslManager = sslManager

	// Initiating HAPROXY Manager
	manager.HaproxyManager = HAPROXY.NewManager(config.HAProxyConfig.UnixSocketPath, config.HAProxyConfig.User, config.HAProxyConfig.Password)

	// Initiating UDP Proxy Manager
	udpProxyManager := UDP_PROXY.NewManager(config.UDPProxyConfig.UnixSocketPath)
	manager.UDPProxyManager = udpProxyManager

	// Initiating Docker Manager
	dockerManager, err := containermanger.NewDockerManager(config.ServiceConfig.DockerUnixSocketPath)
	if err != nil {
		panic(err)
	}
	manager.DockerManager = *dockerManager

	// Initiating Docker Config Generator
	dockerConfigGenerator := DOCKER_CONFIG_GENERATOR.Manager{}
	err = dockerConfigGenerator.Init()
	if err != nil {
		panic(err)
	}
	manager.DockerConfigGenerator = dockerConfigGenerator

	// Create PubSub client
	if config.PubSubConfig.Mode == local_config.LocalPubSub {
		pubSubClient, err := pubsub.NewClient(pubsub.Options{
			Type:         pubsub.Local,
			BufferLength: config.PubSubConfig.BufferLength,
			RedisClient:  nil,
		})
		if err != nil {
			panic(err)
		}
		manager.PubSubClient = pubSubClient
	} else if config.PubSubConfig.Mode == local_config.RemotePubSub {
		pubSubClient, err := pubsub.NewClient(pubsub.Options{
			Type:         pubsub.Remote,
			BufferLength: config.PubSubConfig.BufferLength,
			RedisClient: redis.NewClient(&redis.Options{
				Addr:     fmt.Sprintf("%s:%d", config.PubSubConfig.RedisConfig.Host, config.PubSubConfig.RedisConfig.Port),
				Password: config.PubSubConfig.RedisConfig.Password,
				DB:       config.PubSubConfig.RedisConfig.DatabaseID,
			}),
			TopicsChannelName: "topics",
			EventsChannelName: "events",
		})
		if err != nil {
			panic(err)
		}
		manager.PubSubClient = pubSubClient
	} else {
		panic("Invalid PubSub Mode in config")
	}

	// Create TaskQueue client
	if config.TaskQueueConfig.Mode == local_config.LocalTaskQueue {
		taskQueueClient, err := task_queue.NewClient(task_queue.Options{
			Type:                task_queue.Local,
			Mode:                task_queue.Both, // TODO: option to configure this
			MaxMessagesPerQueue: config.TaskQueueConfig.MaxOutstandingMessagesPerQueue,
			NoOfWorkersPerQueue: config.TaskQueueConfig.NoOfWorkersPerQueue,
		})
		if err != nil {
			panic(err)
		}
		manager.TaskQueueClient = taskQueueClient
	} else if config.TaskQueueConfig.Mode == local_config.RemoteTaskQueue {
		taskQueueClient, err := task_queue.NewClient(task_queue.Options{
			Type:                task_queue.Remote,
			Mode:                task_queue.Both, // TODO: option to configure this
			NoOfWorkersPerQueue: config.TaskQueueConfig.NoOfWorkersPerQueue,
			MaxMessagesPerQueue: config.TaskQueueConfig.MaxOutstandingMessagesPerQueue,
			AMQPUri:             config.TaskQueueConfig.AMQPConfig.URI(),
			AMQPVhost:           config.TaskQueueConfig.AMQPConfig.VHost,
			AMQPClientName:      config.TaskQueueConfig.AMQPConfig.ClientName,
		})
		if err != nil {
			panic(err)
		}
		manager.TaskQueueClient = taskQueueClient
	} else {
		panic("Invalid TaskQueue Mode in config")
	}

}
