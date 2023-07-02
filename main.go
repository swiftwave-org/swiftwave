package main

import (
	"fmt"
	"sync"
	. "keroku/m/haproxy_manager"
)



func main() {
	var wg sync.WaitGroup

	// Create a new HAProxySocket
	var haproxySocket = HAProxySocket{};
	haproxySocket.InitTcpSocket("localhost", 5555);
	haproxySocket.Auth("admin", "mypassword");


	// Wait for events
	wg.Wait()
	fmt.Println("done")
}
