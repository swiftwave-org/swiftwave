package containermanger

import (
	"errors"

	"github.com/docker/docker/api/types"
)

/*
CreateNetwork
Create a new network with the given name
It will create an overlay network with the given name and allow to attach containers to it
*/
func (m Manager) CreateNetwork(name string) error {
	_, err := m.client.NetworkCreate(m.ctx, name, types.NetworkCreate{
		Driver:     "overlay",
		Attachable: true,
	})
	if err != nil {
		return errors.New("error creating network ")
	}
	return nil
}

// RemoveNetwork Delete a network
func (m Manager) RemoveNetwork(name string) error {
	err := m.client.NetworkRemove(m.ctx, name)
	if err != nil {
		return errors.New("error removing network ")
	}
	return nil
}

// ExistsNetwork Check if a network exists
func (m Manager) ExistsNetwork(name string) bool {
	_, err := m.client.NetworkInspect(m.ctx, name, types.NetworkInspectOptions{})
	return err == nil
}

// CIDRNetwork Fetch CIDR of a network
func (m Manager) CIDRNetwork(name string) (string, error) {
	network, err := m.client.NetworkInspect(m.ctx, name, types.NetworkInspectOptions{})
	if err != nil {
		return "", errors.New("error inspecting network ")
	}
	return network.IPAM.Config[0].Subnet, nil
}

// GatewayNetwork Fetch gateway of a network
func (m Manager) GatewayNetwork(name string) (string, error) {
	network, err := m.client.NetworkInspect(m.ctx, name, types.NetworkInspectOptions{})
	if err != nil {
		return "", errors.New("error inspecting network ")
	}
	return network.IPAM.Config[0].Gateway, nil
}
