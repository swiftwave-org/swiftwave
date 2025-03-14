package manager

import (
	"context"
	"github.com/swiftwave-org/swiftwave/pkg/udp_proxy_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
)

func UDPProxyClient(_ context.Context, server core.Server) (*udp_proxy_manager.Manager, error) {
	// TODO fix
	//// Fetch config
	//c, err := config.Fetch()
	//if err != nil {
	//	return nil, err
	//}
	//// Create Net.Conn over SSH
	//manager := udp_proxy_manager.New(func() (net.Conn, error) {
	//	return ssh_toolkit.NetConnOverSSH("unix", c.LocalConfig.ServiceConfig.UDPProxyUnixSocketPath, 20, server.IP, server.SSHPort, server.User, c.SystemConfig.SshPrivateKey)
	//})
	//return &manager, nil
	return nil, nil
}

func UDPProxyClients(ctx context.Context, servers []core.Server) ([]*udp_proxy_manager.Manager, error) {
	var managers []*udp_proxy_manager.Manager
	for _, server := range servers {
		manager, err := UDPProxyClient(ctx, server)
		if err != nil {
			return nil, err
		}
		managers = append(managers, manager)
	}
	return managers, nil
}
