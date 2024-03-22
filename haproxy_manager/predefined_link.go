package haproxymanager

// AddHTTPLink Add HTTP Link [Backend Switch] to HAProxy configuration
// -- Manage ACLs with frontend [only domain_name]
// -- Manage rules with frontend and backend switch
func (s Manager) AddHTTPLink(transactionId string, backendName string, domainName string) error {
	// Check if backend switch already exists
	backendSwitchIndex, err := s.FetchBackendSwitchIndex(transactionId, HTTPMode, 80, backendName, domainName)
	if err != nil {
		return err
	}
	if backendSwitchIndex != -1 {
		return nil
	}
	return s.AddBackendSwitch(transactionId, HTTPMode, 80, backendName, domainName)
}

// DeleteHTTPLink Delete HTTP Link from HAProxy configuration
func (s Manager) DeleteHTTPLink(transactionId string, backendName string, domainName string) error {
	backendSwitchIndex, err := s.FetchBackendSwitchIndex(transactionId, HTTPMode, 80, backendName, domainName)
	if err != nil {
		return err
	}
	if backendSwitchIndex == -1 {
		return nil
	}
	return s.DeleteBackendSwitch(transactionId, HTTPMode, 80, backendName, domainName)
}

// AddHTTPSLink Add HTTPS Link [Backend Switch] to HAProxy configuration
// -- Manage ACLs with frontend [only domain_name]
// -- Manage rules with frontend and backend switch
func (s Manager) AddHTTPSLink(transactionId string, backendName string, domainName string) error {
	// Check if backend switch already exists
	backendSwitchIndex, err := s.FetchBackendSwitchIndex(transactionId, HTTPMode, 443, backendName, domainName)
	if err != nil {
		return err
	}
	if backendSwitchIndex != -1 {
		return nil
	}
	return s.AddBackendSwitch(transactionId, HTTPMode, 443, backendName, domainName)
}

// DeleteHTTPSLink Delete HTTPS Link from HAProxy configuration
func (s Manager) DeleteHTTPSLink(transactionId string, backendName string, domainName string) error {
	// Build query parameters
	backendSwitchIndex, err := s.FetchBackendSwitchIndex(transactionId, HTTPMode, 443, backendName, domainName)
	if err != nil {
		return err
	}
	if backendSwitchIndex == -1 {
		return nil
	}
	return s.DeleteBackendSwitch(transactionId, HTTPMode, 443, backendName, domainName)
}
