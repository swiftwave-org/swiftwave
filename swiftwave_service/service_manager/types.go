package service_manager

import (
	"github.com/go-redis/redis/v8"
	dockerConfigGenerator "github.com/swiftwave-org/swiftwave/docker_config_generator"
	"github.com/swiftwave-org/swiftwave/pubsub"
	SSL "github.com/swiftwave-org/swiftwave/ssl_manager"
	"github.com/swiftwave-org/swiftwave/task_queue"
	"gorm.io/gorm"
)

type ServiceManager struct {
	SslManager            SSL.Manager
	DockerConfigGenerator dockerConfigGenerator.Manager
	DbClient              gorm.DB
	PubSubRedisClient     redis.Client
	PubSubClient          pubsub.Client
	TaskQueueClient       task_queue.Client
	TaskQueueRedisClient  redis.Client
	CancelImageBuildTopic string
}
