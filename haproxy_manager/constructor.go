package haproxymanager

// Init HaProxy Manager with a unix socket
func (s *Manager) InitUnixSocket(unixSocketPath string) {
	s.isUnix = true
	s.unixSocketPath = unixSocketPath
}

// Update auth credentials for HaProxy Manager
func (s *Manager) Auth(username string, password string) {
	s.username = username
	s.password = password
}
