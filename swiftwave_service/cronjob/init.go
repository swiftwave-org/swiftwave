package cronjob

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
)

func NewManager(config *core.ServiceConfig, manager *core.ServiceManager) *Manager {
	if config == nil {
		panic("config cannot be nil")
	}
	if manager == nil {
		panic("manager cannot be nil")
	}
	return &Manager{
		ServiceConfig:  config,
		ServiceManager: manager,
	}
}
