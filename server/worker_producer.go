package server

import "context"

func (s *Server) AddDomainToSSLGenerateQueue(domain string) error {
	task := s.TASK_MAP["ssl-generate"]
	return s.TASK_QUEUE.Add(task.WithArgs(context.Background(), domain))
}

func (s *Server) AddDomainToSSLUpdateHAProxyQueue(domain string) error {
	task := s.TASK_MAP["ssl-update-haproxy"]
	return s.TASK_QUEUE.Add(task.WithArgs(context.Background(), domain))
}

func (s *Server) AddServiceToDockerImageGenerationQueue(service_name string) error {
	task := s.TASK_MAP["docker-image-preparation"]
	return s.TASK_QUEUE.Add(task.WithArgs(context.Background(), service_name))
}

func (s *Server) AddServiceToDeployQueue(service_name string) error {
	task := s.TASK_MAP["deploy-service"]
	return s.TASK_QUEUE.Add(task.WithArgs(context.Background(), service_name))
}