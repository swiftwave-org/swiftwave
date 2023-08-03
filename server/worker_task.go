package server

import "github.com/vmihailenco/taskq/v3"

func (s *Server) RegisterSSLGenerateTask(){
	t := taskq.RegisterTask(&taskq.TaskOptions{
		Name:    "ssl-generate",
		Handler: func(name string) error {
			return s.ProcessGenerateSSLRequestFromQueue(name)
		},
	})
	s.TASK_MAP["ssl-generate"] = t
}

func (s *Server) RegisterUpdateSSLHAProxyTask(){
	t := taskq.RegisterTask(&taskq.TaskOptions{
		Name:    "ssl-update-haproxy",
		Handler: func(name string) error {
			return s.ProcessUpdateSSLHAProxyRequestFromQueue(name)
		},
	})
	s.TASK_MAP["ssl-update-haproxy"] = t
}

// Application deployment tasks
func (s *Server) RegisterDockerImageGenerationTask(){
	t := taskq.RegisterTask(&taskq.TaskOptions{
		Name:    "docker-image-preparationAddServiceToDockerImageGenerationQueue",
		Handler: func(service_name string) error {
			return s.ProcessDockerImageGenerationRequestFromQueue(service_name)
		},
	})
	s.TASK_MAP["docker-image-preparation"] = t
}

func (s *Server) RegisterDeployServiceTask(){
	t := taskq.RegisterTask(&taskq.TaskOptions{
		Name:    "deploy-service",
		Handler: func(service_name string) error {
			return s.ProcessDeployServiceRequestFromQueue(service_name)
		},
	})
	s.TASK_MAP["deploy-service"] = t
}