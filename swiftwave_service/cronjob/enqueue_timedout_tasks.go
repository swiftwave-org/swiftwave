package cronjob

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"time"
)

func (m Manager) EnqueueTimedoutTasks() {
	for {
		err := m.ServiceManager.TaskQueueClient.EnqueueProcessingQueueExpiredTask()
		if err != nil {
			logger.CronJobLoggerError.Println("Error while enqueuing timedout tasks \n", err)
		}
		time.Sleep(1 * time.Hour)
	}
}
