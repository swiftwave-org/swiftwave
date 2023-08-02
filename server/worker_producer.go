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