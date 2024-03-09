package graphql

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/local_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/worker"
)

type Resolver struct {
	ServiceConfig  local_config.Config
	ServiceManager core.ServiceManager
	WorkerManager  worker.Manager
}
