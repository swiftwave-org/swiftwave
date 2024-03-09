package service_manager

import (
	dockerclient "github.com/docker/docker/client"
	"github.com/go-redis/redis/v8"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	dockerConfigGenerator "github.com/swiftwave-org/swiftwave/docker_config_generator"
	haproxy "github.com/swiftwave-org/swiftwave/haproxy_manager"
	"github.com/swiftwave-org/swiftwave/pubsub"
	SSL "github.com/swiftwave-org/swiftwave/ssl_manager"
	"github.com/swiftwave-org/swiftwave/task_queue"
	"github.com/swiftwave-org/swiftwave/udp_proxy_manager"
	"gorm.io/gorm"
)

// ServiceManager : holds the instance of all the managers
type ServiceManager struct {
	SslManager            SSL.Manager
	HaproxyManager        haproxy.Manager
	UDPProxyManager       udp_proxy_manager.Manager
	DockerManager         containermanger.Manager
	DockerConfigGenerator dockerConfigGenerator.Manager
	DockerClient          dockerclient.Client
	DbClient              gorm.DB
	RedisClient           redis.Client
	PubSubClient          pubsub.Client
	TaskQueueClient       task_queue.Client
	CancelImageBuildTopic string
}
