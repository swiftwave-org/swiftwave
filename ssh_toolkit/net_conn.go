package ssh_toolkit

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"net"
	"time"
)

func NetConnOverSSH(
	network, address string, netTimeoutSeconds int, // for target task
	host string, port int, user string, privateKey string, // for ssh client
) (net.Conn, error) {
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
		return nil, fmt.Errorf("dial timeout after %s", timeout)
	}
}
