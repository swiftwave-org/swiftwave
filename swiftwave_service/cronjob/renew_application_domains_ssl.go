package cronjob

import (
	"context"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"time"
)

func (m Manager) RenewApplicationDomainsSSL() {
	logger.CronJobLogger.Println("Starting renew application domains [cronjob]")
	for {
		m.renewApplicationDomainsSSL()
		time.Sleep(24 * time.Hour)
	}
}

func (m Manager) renewApplicationDomainsSSL() {
	// fetch domains having 15 days to expire
	domains, err := core.FetchDomainsThoseWillExpire(context.TODO(), m.ServiceManager.DbClient, 15)
	if err != nil {
		logger.CronJobLogger.Println("Error while fetching domains those will expire \n", err)
		return
	}
	for _, domain := range domains {
		// enqueue domain for renewal
		err := m.WorkerManager.EnqueueSSLGenerateRequest(domain.ID)
		if err != nil {
			logger.CronJobLoggerError.Println("Error while enqueueing domain ", domain.Name, " for renewal \n", err)
		} else {
			logger.CronJobLogger.Println("Domain ", domain.Name, " is enqueued for renewal")
		}
	}
}
