package haproxymanager

// HAProxySocket constructors

// Init HAProxySocket with a unix socket
func (s *HAProxySocket) InitUnixSocket(unixSocketPath string){
	s.isUnix = true;
	s.unixSocketPath = unixSocketPath;
}

// Init HAProxySocket with a tcp socket
func (s *HAProxySocket) InitTcpSocket(host string, port int){
	s.isUnix = false;
	s.Host = host;
	s.Port = port;
}

// HAProxySocket update auth credentials
func (s *HAProxySocket) Auth(username string, password string){
	s.username = username;
	s.password = password;
}