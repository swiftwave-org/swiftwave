package containermanger

import (
	"errors"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
)

// Service manager

// Create a new service
func (m Manager) CreateService(service Service) error {
	_, err := m.client.ServiceCreate(m.ctx, m.serviceToServiceSpec(service), types.ServiceCreateOptions{})
	if err != nil {
		return errors.New("error creating service")
	}
	return nil
}

// Update a service
func (m Manager) UpdateService(service Service) error {
	serviceData, _, err := m.client.ServiceInspectWithRaw(m.ctx, service.Name, types.ServiceInspectOptions{})
	if err != nil {
		return errors.New("error getting swarm server version")
	}
	version := swarm.Version{
		Index: serviceData.Version.Index,
	}
	if err != nil {
		return errors.New("error getting swarm server version")
	}
	_, err = m.client.ServiceUpdate(m.ctx, service.Name, version, m.serviceToServiceSpec(service), types.ServiceUpdateOptions{})
	if err != nil {
		fmt.Println(err)
		return errors.New("error updating service")
	}
	return nil
}

// Rollback a service
func (m Manager) RollbackService(service Service) error {
	serviceData, _, err := m.client.ServiceInspectWithRaw(m.ctx, service.Name, types.ServiceInspectOptions{})
	if err != nil {
		return errors.New("error getting swarm server version")
	}
	version := swarm.Version{
		Index: serviceData.Version.Index,
	}
	if err != nil {
		return errors.New("error getting swarm server version")
	}
	_, err = m.client.ServiceUpdate(m.ctx, service.Name, version, *serviceData.PreviousSpec, types.ServiceUpdateOptions{})
	if err != nil {
		fmt.Println(err)
		return errors.New("error updating service")
	}
	return nil
}

// Remove a service
func (m Manager) RemoveService(service Service) error {
	err := m.client.ServiceRemove(m.ctx, service.Name)
	if err != nil {
		return errors.New("error removing service")
	}
	return nil
}

// Get status of a service
func (m Manager) StatusService(service Service) (ServiceStatus, error) {
	serviceData, _, err := m.client.ServiceInspectWithRaw(m.ctx, service.Name, types.ServiceInspectOptions{
		InsertDefaults: true,
	})
	if err != nil {
		return ServiceStatus{}, errors.New("error getting service status")
	}

	var updateStatus ServiceUpdateStatus;
	if serviceData.UpdateStatus != nil {
		var state ServiceUpdateState;
		switch serviceData.UpdateStatus.State {
		case swarm.UpdateStateUpdating:
			state = ServiceUpdateStateUpdating
		case swarm.UpdateStatePaused:
			state = ServiceUpdateStatePaused
		case swarm.UpdateStateCompleted:
			state = ServiceUpdateStateCompleted
		case swarm.UpdateStateRollbackStarted:
			state = ServiceUpdateStateRollbackStarted
		case swarm.UpdateStateRollbackPaused:
			state = ServiceUpdateStateRollbackPaused
		case swarm.UpdateStateRollbackCompleted:
			state = ServiceUpdateStateRollbackCompleted
		default:
			state = ServiceUpdateStateUnknown
		}
		updateStatus = ServiceUpdateStatus{
			State: state,
			Message: serviceData.UpdateStatus.Message,
		}
	}

	runningReplicas := 0
	// query containers list
	// TODO: query Task list inspite of containers list
	containers, err := m.client.ContainerList(m.ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", "com.docker.swarm.service.name="+service.Name),
		),
	})

	if err != nil {
		return ServiceStatus{}, errors.New("error getting service status")
	}

	runningReplicas = len(containers)
	return ServiceStatus{
		DesiredReplicas: int(*serviceData.Spec.Mode.Replicated.Replicas),
		RunningReplicas: runningReplicas,
		LastUpdatedAt:   serviceData.UpdatedAt.String(),
		ServiceUpdateStatus: updateStatus,
	}, nil
}

// Get service logs
// TODO: Stream logs

// Private functions
func (m Manager) serviceToServiceSpec(service Service) swarm.ServiceSpec {
	// Create swarm attachment config from network names array
	networkAttachmentConfigs := []swarm.NetworkAttachmentConfig{}
	for _, networkName := range service.Networks {
		networkAttachmentConfigs = append(networkAttachmentConfigs, swarm.NetworkAttachmentConfig{
			Target: networkName,
		})
	}

	// Create volume mounts from volume mounts array
	volumeMounts := []mount.Mount{}
	for _, volumeMount := range service.VolumeMounts {
		volumeMounts = append(volumeMounts, mount.Mount{
			Type: mount.TypeVolume,
			Source:   volumeMount.Source,
			Target:   volumeMount.Target,
			ReadOnly: volumeMount.ReadOnly,
		})
	}

	// Create `ENV_VAR=value` array from env map
	env := []string{}
	for key, value := range service.Env {
		env = append(env, key+"="+value)
	}

	// Build service spec
	serviceSpec := swarm.ServiceSpec{
		// Set name of the service
		Annotations: swarm.Annotations{
			Name: service.Name,
		},
		// Set task template
		TaskTemplate: swarm.TaskSpec{
			// Set container spec
			ContainerSpec: &swarm.ContainerSpec{
				Image:   service.Image,
				Command: service.Command,
				Env:     env,
				Mounts:  volumeMounts,
			},
			// Set network name
			Networks: networkAttachmentConfigs,
		},
		// allow replicated service
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &service.Replicas,
			},
		},
		// constant endpoint
		EndpointSpec: &swarm.EndpointSpec{
			Mode: swarm.ResolutionModeDNSRR,
		},
	}
	return serviceSpec
}
