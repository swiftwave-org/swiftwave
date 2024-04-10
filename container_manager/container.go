package containermanger

import (
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"strings"
)

// RemoveServiceContainers removes all containers for a service in a node
func (m Manager) RemoveServiceContainers(serviceName string) error {
	containers, err := m.client.ContainerList(m.ctx, container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", "com.docker.swarm.service.name="+serviceName),
		),
	})
	if err != nil {
		return errors.New("Failed to list containers for service " + serviceName + " " + err.Error())
	}
	for _, c := range containers {
		err = m.client.ContainerRemove(m.ctx, c.ID, container.RemoveOptions{
			Force: true,
		})
		if err != nil {
			return errors.New("Failed to remove container " + c.ID + " " + err.Error())
		}
	}
	return nil
}

// RunCommandInServiceContainers runs a command in all containers for a service
func (m Manager) RunCommandInServiceContainers(serviceName string, command []string) error {
	containers, err := m.client.ContainerList(m.ctx, container.ListOptions{
		All: false,
		Filters: filters.NewArgs(
			filters.Arg("label", "com.docker.swarm.service.name="+serviceName),
		),
	})
	if err != nil {
		return errors.New("Failed to list containers for service " + serviceName + " " + err.Error())
	}
	errorText := ""
	for _, c := range containers {
		res, err := m.client.ContainerExecCreate(m.ctx, c.ID, types.ExecConfig{
			Cmd: command,
		})
		if err != nil {
			errorText += "Failed to create exec for container " + c.ID + " " + err.Error() + "\n"
			continue
		}
		err = m.client.ContainerExecStart(m.ctx, res.ID, types.ExecStartCheck{
			Detach: false,
		})
		if err != nil {
			errorText += "Failed to start exec for container " + c.ID + " " + err.Error() + "\n"
			continue
		}
	}
	if strings.Compare(errorText, "") != 0 {
		return errors.New(errorText)
	}
	return nil
}

// IsContainerRunning checks if a container is running
func (m Manager) IsContainerRunning(containerName string) (bool, error) {
	containers, err := m.client.ContainerList(m.ctx, container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("name", containerName),
		),
	})
	if err != nil {
		return false, errors.New("Failed to list containers " + err.Error())
	}
	if len(containers) == 0 {
		return false, nil
	}
	return containers[0].State == "running", nil
}

// PruneContainers prunes all containers
func (m Manager) PruneContainers() error {
	_, err := m.client.ContainersPrune(m.ctx, filters.NewArgs())
	return err
}
