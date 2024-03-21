package service_manager

import (
	"github.com/go-redis/redis/v8"
	dockerConfigGenerator "github.com/swiftwave-org/swiftwave/docker_config_generator"
	"github.com/swiftwave-org/swiftwave/pubsub"
	SSL "github.com/swiftwave-org/swiftwave/ssl_manager"
	"github.com/swiftwave-org/swiftwave/task_queue"
	"gorm.io/gorm"
)

// ServiceManager : holds the instance of all the managers
type ServiceManager struct {
	SslManager            SSL.Manager
	DockerConfigGenerator dockerConfigGenerator.Manager
	DbClient              gorm.DB
	RedisClient           redis.Client
	PubSubClient          pubsub.Client
	TaskQueueClient       task_queue.Client
	CancelImageBuildTopic string
}
