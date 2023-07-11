package containermanger

import (
	"errors"

	"github.com/docker/docker/api/types"
)

// Create a new network
func (m Manager) CreateNetwork(name string) error {
	_, err := m.client.NetworkCreate(m.ctx, name, types.NetworkCreate{
		Driver: "overlay",
		Attachable: true,
	})
	if err != nil {
		return errors.New("error creating network " + err.Error())
	}
	return nil
}

// Delete a network
func (m Manager) DeleteNetwork(name string) error {
	err := m.client.NetworkRemove(m.ctx, name)
	if err != nil {
		return errors.New("error removing network " + err.Error())
	}
	return nil
}

// Check if a network exists
func (m Manager) ExistNetwork(name string) bool {
	_, err := m.client.NetworkInspect(m.ctx, name, types.NetworkInspectOptions{})
	return err == nil
}

// Fetch CIDR of a network
func (m Manager) CIDRNetwork(name string) (string, error) {
	network, err := m.client.NetworkInspect(m.ctx, name, types.NetworkInspectOptions{})
	if err != nil {
		return "", errors.New("error inspecting network " + err.Error())
	}
	return network.IPAM.Config[0].Subnet, nil
}

// Fetch gateway of a network
func (m Manager) GatewayNetwork(name string) (string, error) {
	network, err := m.client.NetworkInspect(m.ctx, name, types.NetworkInspectOptions{})
	if err != nil {
		return "", errors.New("error inspecting network " + err.Error())
	}
	return network.IPAM.Config[0].Gateway, nil
}