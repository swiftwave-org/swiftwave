package containermanger

import (
	"context"

	"github.com/docker/docker/client"
)

func (m *Manager) Init(ctx context.Context, client client.Client) error {
	m.ctx = ctx
	m.client = &client
	return nil
}
