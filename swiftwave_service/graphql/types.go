package graphql

import (
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/worker"
	"github.com/swiftwave-org/swiftwave/system_config"
)

// Server : hold references to other components of service
type Server struct {
	EchoServer     *echo.Echo
	ServiceConfig  *system_config.Config
	ServiceManager *core.ServiceManager
	WorkerManager  *worker.Manager
}
