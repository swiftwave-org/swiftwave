package cronjob

import (
	"context"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/manager"
	"time"
)

func (m Manager) CleanupUnusedImages() {
	isFirstTime := true
	for {
		if isFirstTime {
			time.Sleep(20 * time.Second)
			isFirstTime = false
		} else {
			// sleep for 1 hour
			time.Sleep(1 * time.Hour)
		}
		// Fetch all servers
		servers, err := core.FetchAllServers(&m.ServiceManager.DbClient)
		if err != nil {
			logger.CronJobLoggerError.Println("Failed to fetch server list")
			logger.CronJobLoggerError.Println(err)
			continue
		}
		if len(servers) == 0 {
			logger.CronJobLogger.Println("Skipping ! No server found")
			continue
		} else {
			for _, server := range servers {
				dockerManager, err := manager.DockerClient(context.Background(), server)
				if err != nil {
					logger.CronJobLoggerError.Println("Failed to create docker client")
					logger.CronJobLoggerError.Println(err)
					continue
				}
				// Prune the images
				err = dockerManager.PruneImages()
				// In stopped state also, we are going to scale down service to 0 replicas
				// so those images will not be deleted
				if err != nil {
					logger.CronJobLoggerError.Println("Failed to prune unused images")
					logger.CronJobLoggerError.Println(err)
				} else {
					logger.CronJobLogger.Println("Unused images pruned")
				}
			}
		}
	}
}
