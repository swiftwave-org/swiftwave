package haproxymanager

// NewManager : Constructor for HAPROXY Manager
func NewManager(unixSocketPath string, username string, password string) Manager {
	m := Manager{}
	m.unixSocketPath = unixSocketPath
	m.username = username
	m.password = password
	return m
}
