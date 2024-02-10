package udp_proxy_manager

func NewManager(unixSocketPath string) Manager {
	return Manager{unixSocketPath: unixSocketPath}
}
