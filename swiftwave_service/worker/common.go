package worker

import (
	"fmt"
	"github.com/swiftwave-org/swiftwave/pubsub"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
	"log"
	"time"
)

var deploymentLogBuffer = make(chan core.DeploymentLog, 10000)

func addPersistentDeploymentLog(_ gorm.DB, pubSubClient pubsub.Client, deploymentId string, content string, terminate bool) {
	addDeploymentLog(pubSubClient, deploymentId, content, terminate, true, true)
}

func addNonPersistentDeploymentLog(_ gorm.DB, pubSubClient pubsub.Client, deploymentId string, content string, terminate bool) {
	addDeploymentLog(pubSubClient, deploymentId, content, terminate, false, true)
}

func addPersistentNonRealtimeDeploymentLog(_ gorm.DB, pubSubClient pubsub.Client, deploymentId string, content string, terminate bool) {
	addDeploymentLog(pubSubClient, deploymentId, content, terminate, true, false)
}

func addDeploymentLog(pubSubClient pubsub.Client, deploymentId string, content string, terminate bool, persistent bool, realtime bool) {
	deploymentLog := core.DeploymentLog{
		DeploymentID: deploymentId,
		Content:      content,
	}
	if persistent {
		deploymentLogBuffer <- deploymentLog
	}
	if realtime {
		err := pubSubClient.Publish(fmt.Sprintf("deployment-log-%s", deploymentId), content)
		if err != nil {
			log.Println("failed to publish deployment log")
		}
	}
	if terminate {
		err := pubSubClient.RemoveTopic(fmt.Sprintf("deployment-log-%s", deploymentId))
		if err != nil {
			log.Println("failed to remove topic")
		}
	}
}

func bulkInsertDeploymentLogs(dbClient gorm.DB) {
	for {
		var deploymentLogs []core.DeploymentLog
		for len(deploymentLogBuffer) > 0 {
			deploymentLog := <-deploymentLogBuffer
			deploymentLogs = append(deploymentLogs, deploymentLog)
		}

		if len(deploymentLogs) > 0 {
			db := dbClient.Session(&gorm.Session{CreateBatchSize: 1000})
			err := db.Create(&deploymentLogs).Error
			if err != nil {
				log.Println("failed to bulk insert deployment logs")
			}
		}
		<-time.After(2 * time.Second)
	}

}
