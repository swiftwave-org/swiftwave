package cronjob

import (
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"log"
	"time"
)

func (m Manager) CleanupUnusedImages() {
	for {
		log.Println("Running cleanup unused images cron job")
		db := m.ServiceManager.DbClient
		// fetch deployments which are stalled
		var deployments []*core.Deployment
		tx := db.Where("status = ?", core.DeploymentStalled).Find(&deployments)
		if tx.Error != nil {
			log.Println("Error while fetching staller/failed deployments", tx.Error)
			return
		}
		// create a empty list of images
		var images = make([]string, 0)
		// iterate over deployments
		for _, deployment := range deployments {
			// append image id to images list
			images = append(images, deployment.DeployableDockerImageURI())
		}
		// delete images
		dockerManager := m.ServiceManager.DockerManager
		for _, image := range images {
			err := dockerManager.RemoveImage(image)
			if err != nil {
				log.Println("Error while deleting image", image, err)
			}
		}
		log.Println("Cleanup unused images cron job completed")
		// sleep for 1 hour
		time.Sleep(1 * time.Hour)
	}
	m.wg.Done()
}
