package manager

import (
	"context"
	haproxymanager "github.com/swiftwave-org/swiftwave/haproxy_manager"
	"github.com/swiftwave-org/swiftwave/ssh_toolkit"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
)

func HAProxyClient(ctx context.Context, server core.Server) (*haproxymanager.Manager, error) {
	// Fetch config
	c, err := config.Fetch()
	if err != nil {
		return nil, err
	}
	// Create Net.Conn over SSH
	conn, err := ssh_toolkit.NetConnOverSSH("unix", c.SystemConfig.HAProxyConfig.UnixSocketPath, 5, server.IP, 22, server.User, c.SystemConfig.SshPrivateKey, 20)
	if err != nil {
		return nil, err
	}
	// Create Docker client
	manager := haproxymanager.New(conn, c.SystemConfig.HAProxyConfig.Username, c.SystemConfig.HAProxyConfig.Password)
	return &manager, nil
}

func HAProxyClients(ctx context.Context, servers []core.Server) ([]*haproxymanager.Manager, error) {
	var managers []*haproxymanager.Manager
	for _, server := range servers {
		manager, err := HAProxyClient(ctx, server)
		if err != nil {
			// close all the connections
			for _, manager := range managers {
				manager.Close()
			}
			return nil, err
		}
		managers = append(managers, manager)
	}
	return managers, nil
}

func KillAllHAProxyConnections(managers []*haproxymanager.Manager) {
	for _, manager := range managers {
		manager.Close()
	}
}
