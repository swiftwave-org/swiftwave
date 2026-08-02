package ssh_toolkit

import (
	"sync"

	"golang.org/x/crypto/ssh"
)

type sshConnectionPool struct {
	clients   map[string]*sshClient // map of host to sshClient
	mutex     *sync.RWMutex
	validator *ServerOnlineStatusValidator
}

type ServerOnlineStatusValidator func(host string) bool

type sshClient struct {
	client *ssh.Client
	err    error
	// ready is closed once client and err are populated
	ready chan struct{}
	// credentials this connection was established with, to detect rotation
	port     int
	user     string
	credHash string

	stopKeepAlive chan struct{}
	closeOnce     sync.Once
}

type OperatingSystem string

const (
	DebianBased OperatingSystem = "debian"
	FedoraBased OperatingSystem = "fedora"
)
