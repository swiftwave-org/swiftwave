package graphql

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/worker"
)

type Resolver struct {
	ServiceConfig  core.ServiceConfig
	ServiceManager core.ServiceManager
	WorkerManager  worker.Manager
}
