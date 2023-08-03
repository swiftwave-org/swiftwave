package server

import (
	"log"
	"time"

	"gorm.io/gorm"
)


func (s *Server) InitCronJobs(){
	go s.MovePendingApplicationsToImageGenerationQueueCronjob()
}

func (s *Server) MovePendingApplicationsToImageGenerationQueueCronjob(){
	for {
		// Get all pending applications
		var applications []Application
		tx := s.DB_CLIENT.Where("status = ?", ApplicationStatusPending).Find(&applications)
		if tx.Error != nil {
			log.Println(tx.Error)
			time.Sleep(5 * time.Second)
			continue
		}
		// Move them to image generation queue
		for _, application := range applications {
			log.Println("Moving application to image generation queue: ", application.ServiceName)
			err := s.DB_CLIENT.Transaction(func(tx *gorm.DB) error {
				// Update status
				application.Status = ApplicationStatusBuildingImageQueued
				tx2 := tx.Save(&application)
				if tx2.Error != nil {
					log.Println(tx2.Error)
					return tx2.Error
				}
				// Enqueue
				err := s.AddServiceToDockerImageGenerationQueue(application.ServiceName)
				if err != nil {
					log.Println(err)
					return err
				}
				return nil
			})
			if err != nil {
				log.Println("Error while moving pending applications to image generation queue: ", err)
			}
		}
		time.Sleep(10 * time.Second)
	}
}