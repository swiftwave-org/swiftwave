package manager

import (
	"context"
	haproxymanager "github.com/swiftwave-org/swiftwave/haproxy_manager"
	"github.com/swiftwave-org/swiftwave/ssh_toolkit"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"net"
)

func HAProxyClient(_ context.Context, server core.Server) (*haproxymanager.Manager, error) {
	// Fetch config
	c, err := config.Fetch()
	if err != nil {
		return nil, err
	}
	// Create client
	manager := haproxymanager.New(func() (net.Conn, error) {
		return ssh_toolkit.NetConnOverSSH("unix", c.LocalConfig.ServiceConfig.HAProxyUnixSocketPath, 50, server.IP, 22, server.User, c.SystemConfig.SshPrivateKey)
	}, c.SystemConfig.HAProxyConfig.Username, c.SystemConfig.HAProxyConfig.Password)
	return &manager, nil
}

func HAProxyClients(ctx context.Context, servers []core.Server) ([]*haproxymanager.Manager, error) {
	var managers []*haproxymanager.Manager
	isErrEncountered := false
	var errEncountered error
	for _, server := range servers {
		manager, err := HAProxyClient(ctx, server)
		if err != nil {
			isErrEncountered = true
			errEncountered = err
			break
		}
		managers = append(managers, manager)
	}
	if isErrEncountered {
		return nil, errEncountered
	}
	return managers, nil
}
