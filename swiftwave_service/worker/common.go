package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/swiftwave-org/swiftwave/pubsub"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
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

func bulkInsertDeploymentLogs(ctx context.Context, wg *sync.WaitGroup, dbClient gorm.DB) {
	defer wg.Done()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			// Drain whatever is left so logs buffered just before shutdown
			// still make it to the database.
			flushDeploymentLogs(dbClient)
			return
		case <-ticker.C:
			flushDeploymentLogs(dbClient)
		}
	}
}

func flushDeploymentLogs(dbClient gorm.DB) {
	var deploymentLogs []core.DeploymentLog
drain:
	for {
		select {
		case deploymentLog := <-deploymentLogBuffer:
			deploymentLogs = append(deploymentLogs, deploymentLog)
		default:
			break drain
		}
	}
	if len(deploymentLogs) == 0 {
		return
	}
	db := dbClient.Session(&gorm.Session{CreateBatchSize: 1000})
	if err := db.Create(&deploymentLogs).Error; err != nil {
		log.Println("failed to bulk insert deployment logs")
	}
}
