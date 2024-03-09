package containermanger

import (
	"context"

	"github.com/docker/docker/client"
)

func NewDockerManager() (*Manager, error) {
	unixSocketPath := "/var/run/docker.sock" // TODO: Read from config
	manager := &Manager{}
	c, err := client.NewClientWithOpts(client.WithHost("unix://"+unixSocketPath), client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	err = manager.init(context.Background(), *c)
	if err != nil {
		return nil, err
	}
	return manager, nil
}

// init: Initializes the container manager with the given context and docker client.
func (m *Manager) init(ctx context.Context, client client.Client) error {
	m.ctx = ctx
	m.client = &client
	return nil
}
