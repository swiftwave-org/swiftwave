package graphql

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/service_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/worker"
)

type Resolver struct {
	Config         config.Config
	ServiceManager service_manager.ServiceManager
	WorkerManager  worker.Manager
}
