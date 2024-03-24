package service_manager

import (
	"context"
	"errors"
	"fmt"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/system_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/db"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"gorm.io/gorm"
	"os"

	"github.com/go-redis/redis/v8"
	dockerConfigGenerator "github.com/swiftwave-org/swiftwave/docker_config_generator"
	"github.com/swiftwave-org/swiftwave/pubsub"
	ssl "github.com/swiftwave-org/swiftwave/ssl_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/task_queue"
)

func (manager *ServiceManager) Load(config config.Config) {
	// Initiating database client
	dbClient, err := db.GetClient(config.LocalConfig, 0)
	if err != nil {
		panic(err.Error())
	}
	manager.DbClient = *dbClient
	// Initiating ssl manager
	options := ssl.ManagerOptions{
		IsStaging:         config.SystemConfig.LetsEncryptConfig.Staging,
		Email:             config.SystemConfig.LetsEncryptConfig.EmailID,
		AccountPrivateKey: config.SystemConfig.LetsEncryptConfig.PrivateKey,
	}
	sslManager := ssl.Manager{}
	err = sslManager.Init(context.Background(), *dbClient, options)
	if err != nil {
		logger.InternalLogger.Println("Failed to initiate ssl manager")
		logger.InternalLoggerError.Println(err)
		panic(err)
	}
	manager.SslManager = sslManager

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
	taskQueueClient, err := FetchTaskQueueClient(&config, dbClient)
	if err != nil {
		logger.InternalLoggerError.Println("Failed to initiate TaskQueue Client\n", err)
		panic(err)
	}
	manager.TaskQueueClient = taskQueueClient
}

func FetchTaskQueueClient(c *config.Config, db *gorm.DB) (task_queue.Client, error) {
	if c.SystemConfig.TaskQueueConfig.Mode == system_config.LocalTaskQueue {
		taskQueueClient, err := task_queue.NewClient(task_queue.Options{
			Type:                task_queue.Local,
			MaxMessagesPerQueue: int(c.SystemConfig.TaskQueueConfig.MaxOutstandingMessagesPerQueue),
			NoOfWorkersPerQueue: int(c.SystemConfig.TaskQueueConfig.NoOfWorkersPerQueue),
			DbClient:            db,
		})
		return taskQueueClient, err
	} else if c.SystemConfig.TaskQueueConfig.Mode == system_config.RemoteTaskQueue {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, errors.New("failed to get hostname")
		}
		var redisClient *redis.Client
		if c.SystemConfig.TaskQueueConfig.RemoteTaskQueueType == system_config.RedisQueue {
			redisClient = redis.NewClient(&redis.Options{
				Addr:     fmt.Sprintf("%s:%d", c.SystemConfig.TaskQueueConfig.RedisConfig.Host, c.SystemConfig.TaskQueueConfig.RedisConfig.Port),
				Password: c.SystemConfig.TaskQueueConfig.RedisConfig.Password,
				DB:       int(c.SystemConfig.TaskQueueConfig.RedisConfig.DatabaseID),
			})
		}
		taskQueueClient, err := task_queue.NewClient(task_queue.Options{
			Type:                task_queue.Remote,
			RemoteQueueType:     task_queue.RemoteQueueType(c.SystemConfig.TaskQueueConfig.RemoteTaskQueueType),
			NoOfWorkersPerQueue: int(c.SystemConfig.TaskQueueConfig.NoOfWorkersPerQueue),
			MaxMessagesPerQueue: int(c.SystemConfig.TaskQueueConfig.MaxOutstandingMessagesPerQueue),
			AMQPUri:             c.SystemConfig.TaskQueueConfig.AMQPConfig.URI(),
			AMQPVhost:           c.SystemConfig.TaskQueueConfig.AMQPConfig.VHost,
			AMQPClientName:      hostname,
			RedisClient:         redisClient,
		})
		if err != nil {
			return nil, err
		}
		return taskQueueClient, nil
	} else {
		return nil, errors.New("invalid TaskQueue Mode in config")
	}
}
