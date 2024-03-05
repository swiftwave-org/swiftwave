package ssh_toolkit

import (
	"golang.org/x/crypto/ssh"
	"sync"
)

type sshConnectionPool struct {
	clients map[string]*sshClient // map of <host:port> to sshClient
	mutex   *sync.RWMutex
}

type sshClient struct {
	client *ssh.Client
	mutex  *sync.RWMutex
}
