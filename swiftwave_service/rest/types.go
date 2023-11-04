package rest

import (
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/worker"
)

// Server : hold references to other components of service
type Server struct {
	EchoServer     *echo.Echo
	ServiceConfig  *core.ServiceConfig
	ServiceManager *core.ServiceManager
	WorkerManager  *worker.Manager
}
