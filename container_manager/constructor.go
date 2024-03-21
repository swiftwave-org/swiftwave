package containermanger

import (
	"context"
	"github.com/docker/docker/client"
	"net"
	"net/http"
)

// New creates a new container manager
func New(ctx context.Context, conn net.Conn) (*Manager, error) {
	manager := &Manager{}
	c, err := client.NewClientWithOpts(
		client.WithAPIVersionNegotiation(),
		client.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return conn, nil
				},
			},
		}),
	)
	if err != nil {
		return nil, err
	}
	manager.ctx = ctx
	manager.client = c
	return manager, nil
}

// NewLocalClient creates a new container manager with a local client
func NewLocalClient(ctx context.Context) (*Manager, error) {
	manager := &Manager{}
	c, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	manager.ctx = ctx
	manager.client = c
	return manager, nil
}

// Close closes the manager
func (m Manager) Close() error {
	return m.client.Close()
}
