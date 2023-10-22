package containermanger

import (
	"context"

	"github.com/docker/docker/client"
)

// Initializes the container manager with the given context and docker client.
func (m *Manager) Init(ctx context.Context, client client.Client) error {
	m.ctx = ctx
	m.client = &client
	return nil
}
