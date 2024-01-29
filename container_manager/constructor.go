package containermanger

import (
	"context"

	"github.com/docker/docker/client"
)

func NewDockerManager(unixSocketPath string) (*Manager, error) {
	manager := &Manager{}
	client, err := client.NewClientWithOpts(client.WithHost("unix://"+unixSocketPath), client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	err = manager.init(context.Background(), *client)
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
