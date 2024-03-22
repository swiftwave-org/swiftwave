package haproxymanager

// AddTCPLink Add TCP Frontend to HAProxy configuration
// -- Manage ACLs with frontend [port{required} and domain_name{optional}]
// -- Manage rules with frontend and backend switch
func (s Manager) AddTCPLink(transactionId string, backendName string, port int, domainName string, listenerMode ListenerMode, restrictedPorts []int) error {
	// Add Frontend
	err := s.AddFrontend(transactionId, listenerMode, port, restrictedPorts)
	if err != nil {
		return err
	}
	// Add Backend Switch
	err = s.AddBackendSwitch(transactionId, listenerMode, port, backendName, domainName)
	return err
}

// DeleteTCPLink Delete TCP Frontend from HAProxy configuration
func (s Manager) DeleteTCPLink(transactionId string, backendName string, port int, domainName string, listenerMode ListenerMode) error {
	// Delete Backend Switch
	err := s.DeleteBackendSwitch(transactionId, listenerMode, port, backendName, domainName)
	if err != nil {
		return err
	}
	// Delete Frontend
	err = s.DeleteFrontend(transactionId, listenerMode, port)
	return err
}
