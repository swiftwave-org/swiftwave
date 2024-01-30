package cronjob

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/system_config"
	"sync"
)

func NewManager(config *system_config.Config, manager *core.ServiceManager) CronJob {
	if config == nil {
		panic("config cannot be nil")
	}
	if manager == nil {
		panic("manager cannot be nil")
	}
	return Manager{
		SystemConfig:   config,
		ServiceManager: manager,
		wg:             &sync.WaitGroup{},
	}
}

func (m Manager) Start(nowait bool) {
	// Start cron jobs
	m.wg.Add(1)
	go m.HaProxyPortExposer()
	m.wg.Add(1)
	go m.CleanupUnusedImages()
	if !nowait {
		m.wg.Wait()
	}
}

func (m Manager) Wait() {
	m.wg.Wait()
}
