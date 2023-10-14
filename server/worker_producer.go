package server

import (
	"context"
)

/*
This file contains wrapper [producer] functions to add tasks to queue
In the source code, we can see that the tasks are needed to add to queue
This wrapper functions are used to add tasks to queue
*/

// Wrapper to add domain SSL generation task to queue
func (s *Server) AddDomainToSSLGenerateQueue(domain string) error {
	task := s.TASK_MAP["ssl-generate"]
	return s.TASK_QUEUE.Add(task.WithArgs(context.Background(), domain))
}

// Wrapper to add domain SSL update in HaProxy task to queue
func (s *Server) AddDomainToSSLUpdateHAProxyQueue(domain string) error {
	task := s.TASK_MAP["ssl-update-haproxy"]
	return s.TASK_QUEUE.Add(task.WithArgs(context.Background(), domain))
}

// Wrapper to add service to docker image generation task to queue
func (s *Server) AddServiceToDockerImageGenerationQueue(service_name string, log_id string) error {
	task := s.TASK_MAP["docker-image-preparation"]
	var application Application
	if err := s.DB_CLIENT.Where("service_name = ?", service_name).First(&application).Error; err != nil {
		return err
	}
	return s.TASK_QUEUE.Add(task.WithArgs(context.Background(), application.ID, log_id))
}

// Wrapper to add service to deploy task to queue
func (s *Server) AddServiceToDeployQueue(service_name string) error {
	task := s.TASK_MAP["deploy-service"]
	var application Application
	if err := s.DB_CLIENT.Where("service_name = ?", service_name).First(&application).Error; err != nil {
		return err
	}
	return s.TASK_QUEUE.Add(task.WithArgs(context.Background(), application.ID))
}
