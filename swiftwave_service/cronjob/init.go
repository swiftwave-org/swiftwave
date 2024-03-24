package cronjob

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/service_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/worker"
	"sync"
)

func NewManager(config *config.Config, manager *service_manager.ServiceManager, workerManager *worker.Manager) CronJob {
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
		WorkerManager:  workerManager,
	}
}

func (m Manager) Start(nowait bool) {
	// Start cron jobs
	m.wg.Add(1)
	go m.CleanupUnusedImages()
	m.wg.Add(1)
	go m.SyncProxy()
	m.wg.Add(1)
	go m.SyncBackupProxyServer()
	m.wg.Add(1)
	go m.MonitorServerStatus()
	m.wg.Add(1)
	go m.RenewApplicationDomainsSSL()
	if m.Config.LocalConfig.ServiceConfig.UseTLS && m.Config.LocalConfig.ServiceConfig.AutoRenewManagementNodeCert {
		m.wg.Add(1)
		go m.RenewManagementNodeSSL()
	} else {
		logger.CronJobLogger.Println("[IGNORE JOB] Management node SSL auto renew is disabled")
	}
	m.wg.Add(1)
	go m.EnqueueTimedoutTasks()
	if !nowait {
		m.wg.Wait()
	}
}

func (m Manager) Wait() {
	m.wg.Wait()
}
