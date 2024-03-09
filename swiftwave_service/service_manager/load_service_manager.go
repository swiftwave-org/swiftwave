package service_manager

import (
	"context"
	"fmt"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/system_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/db"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	udpproxy "github.com/swiftwave-org/swiftwave/udp_proxy_manager"
	"os"

	"github.com/go-redis/redis/v8"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	dockerConfigGenerator "github.com/swiftwave-org/swiftwave/docker_config_generator"
	haproxy "github.com/swiftwave-org/swiftwave/haproxy_manager"
	"github.com/swiftwave-org/swiftwave/pubsub"
	ssl "github.com/swiftwave-org/swiftwave/ssl_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/task_queue"
)

func (manager *ServiceManager) Load(config config.Config) {
	// current hostname
	hostname, err := os.Hostname()
	if err != nil {
		logger.InternalLoggerError.Println("Failed to get hostname")
		logger.InternalLoggerError.Println(err)
		panic(err)
	}
	// Initiating database client
	dbClient, err := db.GetClient(config.LocalConfig, 0)
	if err != nil {
		panic(err.Error())
	}
	manager.DbClient = *dbClient

	// Initiating ssl Manager
	options := ssl.ManagerOptions{
		IsStaging:         config.SystemConfig.LetsEncryptConfig.Staging,
		Email:             config.SystemConfig.LetsEncryptConfig.EmailID,
		AccountPrivateKey: config.SystemConfig.LetsEncryptConfig.PrivateKey,
	}
	sslManager := ssl.Manager{}
	err = sslManager.Init(context.Background(), *dbClient, options)
	if err != nil {
		logger.InternalLogger.Println("Failed to initiate ssl Manager")
		logger.InternalLoggerError.Println(err)
		panic(err)
	}
	manager.SslManager = sslManager

	// Initiating haproxy Manager
	manager.HaproxyManager = haproxy.NewManager(config.SystemConfig.HAProxyConfig.UnixSocketPath, config.SystemConfig.HAProxyConfig.Username, config.SystemConfig.HAProxyConfig.Password)

	// Initiating UDP Proxy Manager
	udpProxyManager := udpproxy.NewManager(config.SystemConfig.UDPProxyConfig.UnixSocketPath)
	manager.UDPProxyManager = udpProxyManager

	// Initiating Docker Manager
	dockerManager, err := containermanger.NewDockerManager()
	if err != nil {
		logger.InternalLogger.Println("Failed to initiate Docker Manager")
		logger.InternalLoggerError.Println(err)
		panic(err)
	}
	manager.DockerManager = *dockerManager

	// Initiating Docker Config Generator
	dockerConfigGeneratorInstance := dockerConfigGenerator.Manager{}
	err = dockerConfigGeneratorInstance.Init()
	if err != nil {
		logger.InternalLogger.Println("Failed to initiate Docker Config Generator")
		logger.InternalLoggerError.Println(err)
		panic(err)
	}
	manager.DockerConfigGenerator = dockerConfigGeneratorInstance

	// Create PubSub client
	if config.SystemConfig.PubSubConfig.Mode == system_config.LocalPubSub {
		pubSubClient, err := pubsub.NewClient(pubsub.Options{
			Type:         pubsub.Local,
			BufferLength: int(config.SystemConfig.PubSubConfig.BufferLength),
			RedisClient:  nil,
		})
		if err != nil {
			logger.InternalLogger.Println("Failed to initiate PubSub Client")
			logger.InternalLoggerError.Println(err)
			panic(err)
		}
		manager.PubSubClient = pubSubClient
	} else if config.SystemConfig.PubSubConfig.Mode == system_config.RemotePubSub {
		pubSubClient, err := pubsub.NewClient(pubsub.Options{
			Type:         pubsub.Remote,
			BufferLength: int(config.SystemConfig.PubSubConfig.BufferLength),
			RedisClient: redis.NewClient(&redis.Options{
				Addr:     fmt.Sprintf("%s:%d", config.SystemConfig.PubSubConfig.RedisConfig.Host, config.SystemConfig.PubSubConfig.RedisConfig.Port),
				Password: config.SystemConfig.PubSubConfig.RedisConfig.Password,
				DB:       int(config.SystemConfig.PubSubConfig.RedisConfig.DatabaseID),
			}),
			TopicsChannelName: "topics",
			EventsChannelName: "events",
		})
		if err != nil {
			logger.InternalLogger.Println("Failed to initiate PubSub Client")
			logger.InternalLoggerError.Println(err)
			panic(err)
		}
		manager.PubSubClient = pubSubClient
	} else {
		panic("Invalid PubSub Mode in config")
	}

	// Create TaskQueue client
	if config.SystemConfig.TaskQueueConfig.Mode == system_config.LocalTaskQueue {
		taskQueueClient, err := task_queue.NewClient(task_queue.Options{
			Type:                task_queue.Local,
			Mode:                task_queue.Both, // TODO: option to configure this
			MaxMessagesPerQueue: int(config.SystemConfig.TaskQueueConfig.MaxOutstandingMessagesPerQueue),
			NoOfWorkersPerQueue: int(config.SystemConfig.TaskQueueConfig.NoOfWorkersPerQueue),
		})
		if err != nil {
			logger.InternalLogger.Println("Failed to initiate TaskQueue Client")
			logger.InternalLoggerError.Println(err)
			panic(err)
		}
		manager.TaskQueueClient = taskQueueClient
	} else if config.SystemConfig.TaskQueueConfig.Mode == system_config.RemoteTaskQueue {
		taskQueueClient, err := task_queue.NewClient(task_queue.Options{
			Type:                task_queue.Remote,
			Mode:                task_queue.Both, // TODO: option to configure this
			NoOfWorkersPerQueue: int(config.SystemConfig.TaskQueueConfig.NoOfWorkersPerQueue),
			MaxMessagesPerQueue: int(config.SystemConfig.TaskQueueConfig.MaxOutstandingMessagesPerQueue),
			AMQPUri:             config.SystemConfig.TaskQueueConfig.AMQPConfig.URI(),
			AMQPVhost:           config.SystemConfig.TaskQueueConfig.AMQPConfig.VHost,
			AMQPClientName:      hostname,
		})
		if err != nil {
			logger.InternalLogger.Println("Failed to initiate TaskQueue Client")
			logger.InternalLoggerError.Println(err)
			panic(err)
		}
		manager.TaskQueueClient = taskQueueClient
	} else {
		panic("Invalid TaskQueue Mode in config")
	}

}