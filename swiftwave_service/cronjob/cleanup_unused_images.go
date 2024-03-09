package cronjob

import (
	"log"
	"time"
)

func (m Manager) CleanupUnusedImages() {
	for {
		dockerManager := m.ServiceManager.DockerManager
		// Prune the images
		err := dockerManager.PruneImages()
		// In stopped state also, we are going to scale down service to 0 replicas
		// so those images will not be deleted
		if err != nil {
			log.Println("Failed to prune unused images")
			log.Println(err)
		} else {
			log.Println("Unused images pruned")
		}
		// sleep for 1 hour
		time.Sleep(1 * time.Hour)
	}
}
