package graphql

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/worker"
	"github.com/swiftwave-org/swiftwave/system_config"
)

type Resolver struct {
	ServiceConfig  system_config.Config
	ServiceManager core.ServiceManager
	WorkerManager  worker.Manager
}
