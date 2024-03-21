package console

import (
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/service_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/worker"
)

// Server : hold references to other components of service
type Server struct {
	EchoServer     *echo.Echo
	Config         *config.Config
	ServiceManager *service_manager.ServiceManager
	WorkerManager  *worker.Manager
}

// PTYDimension : struct to hold console dimension
type PTYDimension struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}
