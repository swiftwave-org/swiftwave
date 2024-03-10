package udp_proxy_manager

import "net"

// New : Constructor for new instance of udp proxy manager
func New(conn net.Conn) Manager {
	return Manager{
		netConn: conn,
	}
}
