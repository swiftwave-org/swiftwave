package cronjob

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"time"
)

func (m Manager) RenewApplicationDomains() {
	logger.CronJobLogger.Println("Starting renew application domains [cronjob]")
	for {
		m.renewApplicationDomains()
		time.Sleep(1 * time.Hour)
	}
}

func (m Manager) renewApplicationDomains() {

}
