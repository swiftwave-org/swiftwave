package haproxymanager

import "net"

// New : Constructor for new instance of haproxy manager
func New(con net.Conn, username string, password string) Manager {
	return Manager{
		netConn:  con,
		username: username,
		password: password,
	}
}
