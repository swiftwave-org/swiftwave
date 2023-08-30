package server

import (
	"context"
)

func (s *Server) AddDomainToSSLGenerateQueue(domain string) error {
	task := s.TASK_MAP["ssl-generate"]
	return s.TASK_QUEUE.Add(task.WithArgs(context.Background(), domain))
}

func (s *Server) AddDomainToSSLUpdateHAProxyQueue(domain string) error {
	task := s.TASK_MAP["ssl-update-haproxy"]
	return s.TASK_QUEUE.Add(task.WithArgs(context.Background(), domain))
}

func (s *Server) AddServiceToDockerImageGenerationQueue(service_name string, log_id string) error {
	task := s.TASK_MAP["docker-image-preparation"]
	var application Application
	if err := s.DB_CLIENT.Where("service_name = ?", service_name).First(&application).Error; err != nil {
		return err
	}
	return s.TASK_QUEUE.Add(task.WithArgs(context.Background(), application.ID, log_id))
}

func (s *Server) AddServiceToDeployQueue(service_name string) error {
	task := s.TASK_MAP["deploy-service"]
	var application Application
	if err := s.DB_CLIENT.Where("service_name = ?", service_name).First(&application).Error; err != nil {
		return err
	}
	return s.TASK_QUEUE.Add(task.WithArgs(context.Background(), application.ID))
}
