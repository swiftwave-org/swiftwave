package containermanger

import "github.com/docker/docker/client"

func (m *Manager) Client() *client.Client {
	return m.client
}
