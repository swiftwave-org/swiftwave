package manager

import (
	"context"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	"github.com/swiftwave-org/swiftwave/ssh_toolkit"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
)

func DockerClient(ctx context.Context, server core.Server) (*containermanger.Manager, error) {
	// Fetch config
	c, err := config.Fetch()
	if err != nil {
		return nil, err
	}
	// Create Net.Conn over SSH
	conn, err := ssh_toolkit.NetConnOverSSH("unix", server.DockerUnixSocketPath, 5, server.IP, 22, server.User, c.SystemConfig.SshPrivateKey)
	if err != nil {
		return nil, err
	}
	// Create Docker client
	manager, err := containermanger.New(ctx, conn)
	if err != nil {
		return nil, err
	}
	return manager, nil
}
