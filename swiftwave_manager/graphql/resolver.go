package graphql

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_manager/core"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	ServiceConfig  core.ServiceConfig
	ServiceManager core.ServiceManager
}
