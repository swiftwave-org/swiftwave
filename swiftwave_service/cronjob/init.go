package cronjob

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/service_manager"
	"sync"
)

func NewManager(config *config.Config, manager *service_manager.ServiceManager) CronJob {
	if config == nil {
		panic("config cannot be nil")
	}
	if manager == nil {
		panic("manager cannot be nil")
	}
	return Manager{
		Config:         config,
		ServiceManager: manager,
		wg:             &sync.WaitGroup{},
	}
}

func (m Manager) Start(nowait bool) {
	// Start cron jobs
	m.wg.Add(1)
	go m.HaProxyPortExposer()
	m.wg.Add(1)
	go m.UDPProxyPortExposer()
	m.wg.Add(1)
	go m.CleanupUnusedImages()
	m.wg.Add(1)
	go m.SyncProxy()
	m.wg.Add(1)
	go m.SyncBackupProxyServer()
	if !nowait {
		m.wg.Wait()
	}
}

func (m Manager) Wait() {
	m.wg.Wait()
}
