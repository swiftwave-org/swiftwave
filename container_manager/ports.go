package containermanger

import (
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
)

// Fetch published host ports of a service
func (m Manager) FetchPublishedHostPorts(service_name string) ([]int, error) {
	serviceData, _, err := m.client.ServiceInspectWithRaw(m.ctx, service_name, types.ServiceInspectOptions{})
	if err != nil {
		return nil, errors.New("error getting service details > " + service_name)
	}
	ports := []int{}
	for _, port := range serviceData.Endpoint.Ports {
		ports = append(ports, int(port.PublishedPort))
	}
	return ports, nil
}

// FetchPublishedPortRules Fetch published port rules of a service
func (m Manager) FetchPublishedPortRules(service_name string) ([]swarm.PortConfig, error) {
	serviceData, _, err := m.client.ServiceInspectWithRaw(m.ctx, service_name, types.ServiceInspectOptions{})
	if err != nil {
		return nil, errors.New("error getting service details > " + service_name)
	}
	return serviceData.Endpoint.Ports, nil
}

// update published host ports of a service
func (m Manager) UpdatePublishedHostPorts(service_name string, ports []swarm.PortConfig) error {
	serviceData, _, err := m.client.ServiceInspectWithRaw(m.ctx, service_name, types.ServiceInspectOptions{})
	if err != nil {
		return errors.New("error getting swarm server version")
	}
	serviceData.Endpoint.Ports = ports
	serviceData.Spec.EndpointSpec.Mode = swarm.ResolutionModeVIP
	serviceData.Spec.EndpointSpec.Ports = ports
	_, err = m.client.ServiceUpdate(m.ctx, service_name, serviceData.Version, serviceData.Spec, types.ServiceUpdateOptions{})
	if err != nil {
		return errors.New("error updating service")
	}
	return nil
}
