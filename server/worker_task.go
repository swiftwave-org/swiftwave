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