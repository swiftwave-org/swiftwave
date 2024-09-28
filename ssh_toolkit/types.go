package ssh_toolkit

import (
	"sync"

	"golang.org/x/crypto/ssh"
)

type sshConnectionPool struct {
	clients   map[string]*sshClient // map of <host:port> to sshClient
	mutex     *sync.RWMutex
	validator *ServerOnlineStatusValidator
}

type ServerOnlineStatusValidator func(host string) bool

type sshClient struct {
	client *ssh.Client
	mutex  *sync.RWMutex
}

type OperatingSystem string

const (
	DebianBased OperatingSystem = "debian"
	FedoraBased OperatingSystem = "fedora"
)
