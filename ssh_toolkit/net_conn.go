package ssh_toolkit

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"net"
	"time"
)

func isLocalHost(host string) bool {
	return host == "127.0.0.1" || host == "localhost" || host == "::1" || host == ""
}

func NetConnOverSSH(
	network, address string, netTimeoutSeconds int, // for target task
	host string, port int, user string, privateKey string, // for ssh client
) (net.Conn, error) {
	// If connecting to the local server, bypass SSH entirely and connect directly
	if isLocalHost(host) {
		return net.DialTimeout(network, address, time.Duration(netTimeoutSeconds)*time.Second)
	}
	// fetch ssh client
	sshRecord, err := getSSHClientWithOptions(host, port, user, privateKey, true)
	if err != nil {
		if isErrorWhenSSHClientNeedToBeRecreated(err) {
			DeleteSSHClient(host)
		}
		return nil, err
	}
	// create net connection
	conn, err := dialWithTimeout(sshRecord, network, address, time.Duration(netTimeoutSeconds)*time.Second)
	if err != nil && isErrorWhenSSHClientNeedToBeRecreated(err) {
		DeleteSSHClient(host)
	}
	return conn, err
}

// private functions
func dialWithTimeout(client *ssh.Client, network, address string, timeout time.Duration) (net.Conn, error) {
	type dialResult struct {
		conn net.Conn
		err  error
	}
	resultCh := make(chan dialResult, 1)
	go func() {
		conn, err := client.Dial(network, address)
		resultCh <- dialResult{conn, err}
	}()
	select {
	case result := <-resultCh:
		return result.conn, result.err
	case <-time.After(timeout):
		// the connection may still turn up later, nobody would be left to close it
		go func() {
			if result := <-resultCh; result.conn != nil {
				_ = result.conn.Close()
			}
		}()
		return nil, fmt.Errorf("dial timeout after %s", timeout)
	}
}
