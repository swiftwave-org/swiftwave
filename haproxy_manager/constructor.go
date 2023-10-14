package haproxymanager

// Init HaProxy Manager with a unix socket
func (s *Manager) InitUnixSocket(unixSocketPath string) {
	s.isUnix = true
	s.unixSocketPath = unixSocketPath
}

// Init HaProxy Manager with tcp socket info (host, port)
func (s *Manager) InitTcpSocket(host string, port int) {
	s.isUnix = false
	s.Host = host
	s.Port = port
}

// Update auth credentials for HaProxy Manager
func (s *Manager) Auth(username string, password string) {
	s.username = username
	s.password = password
}
