package server


// Register the worker tasks
func (s *Server) RegisteWorkerTasks(){
	s.RegisterSSLGenerateTask()
	s.RegisterUpdateSSLHAProxyTask()
}

// Start the worker consumers
func (s *Server) StartWorkerConsumers() error {
	return s.TASK_QUEUE.Consumer().Start(s.WORKER_CONTEXT)
}