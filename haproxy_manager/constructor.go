package haproxymanager

// Manager constructors

// Init Manager with a unix socket
func (s *Manager) InitUnixSocket(unixSocketPath string) {
	s.isUnix = true
	s.unixSocketPath = unixSocketPath
}

// Init Manager with a tcp socket
func (s *Manager) InitTcpSocket(host string, port int) {
	s.isUnix = false
	s.Host = host
	s.Port = port
}

// Manager update auth credentials
func (s *Manager) Auth(username string, password string) {
	s.username = username
	s.password = password
}
