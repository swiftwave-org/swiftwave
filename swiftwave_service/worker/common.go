package worker

import (
	"fmt"
	"github.com/swiftwave-org/swiftwave/pubsub"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
	"log"
	"time"
)

var deploymentLogBuffer = make(chan *core.DeploymentLog, 10000)

func addDeploymentLog(_ gorm.DB, pubSubClient pubsub.Client, deploymentId string, content string, terminate bool) {
	deploymentLog := &core.DeploymentLog{
		DeploymentID: deploymentId,
		Content:      content,
	}
	deploymentLogBuffer <- deploymentLog
	err := pubSubClient.Publish(fmt.Sprintf("deployment-log-%s", deploymentId), content)
	if err != nil {
		log.Println("failed to publish deployment log")
	}
	if terminate {
		err := pubSubClient.RemoveTopic(fmt.Sprintf("deployment-log-%s", deploymentId))
		if err != nil {
			log.Println("failed to remove topic")
		}
	}
}

func bulkInsertDeploymentLogs(dbClient gorm.DB) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var deploymentLogs []*core.DeploymentLog
			for len(deploymentLogBuffer) > 0 {
				deploymentLog := <-deploymentLogBuffer
				deploymentLogs = append(deploymentLogs, deploymentLog)
			}

			if len(deploymentLogs) > 0 {
				err := dbClient.Create(&deploymentLogs).Error
				if err != nil {
					log.Println("failed to bulk insert deployment logs")
				}
			}
		}
	}
}
