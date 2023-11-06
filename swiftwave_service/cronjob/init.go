package cronjob

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"sync"
)

func NewManager(config *core.ServiceConfig, manager *core.ServiceManager) CronJob {
	if config == nil {
		panic("config cannot be nil")
	}
	if manager == nil {
		panic("manager cannot be nil")
	}
	return Manager{
		ServiceConfig:  config,
		ServiceManager: manager,
		wg:             &sync.WaitGroup{},
	}
}

func (m Manager) Start(nowait bool) {
	// Start cron jobs
	m.wg.Add(1)
	go m.HaProxyPortExposer()
	if !nowait {
		m.wg.Wait()
	}
}

func (m Manager) Wait() {
	m.wg.Wait()
}
