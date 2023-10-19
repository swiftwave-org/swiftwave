package server

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	DOCKER "github.com/swiftwave-org/swiftwave/container_manager"
	DOCKER_CONFIG_GENERATOR "github.com/swiftwave-org/swiftwave/docker_config_generator"
	HAPROXY "github.com/swiftwave-org/swiftwave/haproxy_manager"
	SSL "github.com/swiftwave-org/swiftwave/ssl_manager"

	DOCKER_CLIENT "github.com/docker/docker/client"

	"github.com/go-redis/redis/v8"
	"github.com/vmihailenco/taskq/v3"
	"gorm.io/gorm"
)

// Server struct
type Server struct {
	SSL_MANAGER                    SSL.Manager
	HAPROXY_MANAGER                HAPROXY.Manager
	DOCKER_MANAGER                 DOCKER.Manager
	DOCKER_CONFIG_GENERATOR        DOCKER_CONFIG_GENERATOR.Manager
	DOCKER_CLIENT                  DOCKER_CLIENT.Client
	DB_CLIENT                      gorm.DB
	REDIS_CLIENT                   redis.Client
	ECHO_SERVER                    echo.Echo
	PORT                           int
	HAPROXY_SERVICE                string
	CODE_TARBALL_DIR               string
	SWARM_NETWORK                  string
	RESTRICTED_PORTS               []int
	SESSION_TOKENS                 map[string]time.Time
	SESSION_TOKEN_EXPIRY_MINUTES   int
	// Worker related
	QUEUE_FACTORY         taskq.Factory
	TASK_QUEUE            taskq.Queue
	TASK_MAP              map[string]*taskq.Task
	WORKER_CONTEXT        context.Context
	WORKER_CONTEXT_CANCEL context.CancelFunc
	// ENVIRONMENT
	ENVIRONMENT string
}
