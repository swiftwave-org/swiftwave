package containermanger

import (
	"errors"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
)

// Get Service
func (m Manager) GetService(serviceName string) (Service, error) {
	serviceData, _, err := m.client.ServiceInspectWithRaw(m.ctx, serviceName, types.ServiceInspectOptions{})
	if err != nil {
		return Service{}, errors.New("error getting service")
	}
	// Create service object
	service := Service{
		Name:     serviceData.Spec.Name,
		Image:    serviceData.Spec.TaskTemplate.ContainerSpec.Image,
		Command:  serviceData.Spec.TaskTemplate.ContainerSpec.Command,
		Env:      make(map[string]string),
		Networks: []string{},
		Replicas: 0,
	}
	// Set env
	for _, env := range serviceData.Spec.TaskTemplate.ContainerSpec.Env {
		service.Env[env] = ""
	}
	// Set volume mounts
	for _, volumeMount := range serviceData.Spec.TaskTemplate.ContainerSpec.Mounts {
		service.VolumeMounts = append(service.VolumeMounts, VolumeMount{
			Source:   volumeMount.Source,
			Target:   volumeMount.Target,
			ReadOnly: volumeMount.ReadOnly,
		})
	}
	// Set networks
	for _, network := range serviceData.Spec.TaskTemplate.Networks {
		service.Networks = append(service.Networks, network.Target)
	}
	// Set replicas
	if serviceData.Spec.Mode.Replicated != nil {
		service.Replicas = *serviceData.Spec.Mode.Replicated.Replicas
	}
	return service, nil
}

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
		return errors.New("error updating service")
	}
	return nil
}

// Remove a service
func (m Manager) RemoveService(servicename string) error {
	err := m.client.ServiceRemove(m.ctx, servicename)
	if err != nil {
		return errors.New("error removing service")
	}
	return nil
}

// Fetch Realtime Info of a services in bulk
func (m Manager) RealtimeInfoRunningServices() (map[string]ServiceRealtimeInfo, error) {
	// fetch all nodes and store in map > nodeID:nodeDetails
	nodes, err := m.client.NodeList(m.ctx, types.NodeListOptions{})
	if err != nil {
		return nil, errors.New("error getting node list")
	}
	nodeMap := make(map[string]swarm.Node)
	for _, node := range nodes {
		nodeMap[node.ID] = node
	}
	// fetch all services and store in map > serviceName:serviceDetails
	services, err := m.client.ServiceList(m.ctx, types.ServiceListOptions{})
	if err != nil {
		return nil, errors.New("error getting service list")
	}
	// create map of service name to service realtime info
	serviceRealtimeInfoMap := make(map[string]ServiceRealtimeInfo)
	// analyze each service
	for _, service := range services {
		runningCount := 0

		// inspect service to get desired count
		serviceData, _, err := m.client.ServiceInspectWithRaw(m.ctx, service.ID, types.ServiceInspectOptions{})
		if err != nil {
			continue
		}
		// create service realtime info
		serviceRealtimeInfo := ServiceRealtimeInfo{}
		serviceRealtimeInfo.Name = serviceData.Spec.Name
		serviceRealtimeInfo.PlacementInfos = []ServiceTaskPlacementInfo{}
		// set desired count
		if serviceData.Spec.Mode.Replicated != nil {
			serviceRealtimeInfo.DesiredReplicas = int(*serviceData.Spec.Mode.Replicated.Replicas)
			serviceRealtimeInfo.ReplicatedService = true
		} else {
			serviceRealtimeInfo.DesiredReplicas = -1
			serviceRealtimeInfo.ReplicatedService = false
		}

		// query task list
		tasks, err := m.client.TaskList(m.ctx, types.TaskListOptions{
			Filters: filters.NewArgs(
				filters.Arg("desired-state", "running"),
				filters.Arg("name", serviceData.Spec.Name),
			),
		})
		// set running count
		if err != nil {
			continue
		}
		runningCount = len(tasks)
		serviceRealtimeInfo.RunningReplicas = runningCount
		servicePlacementCountMap := make(map[string]int) // nodeID:count
		// set placement infos > how many replicas are running in each node
		for _, task := range tasks {
			servicePlacementCountMap[task.NodeID]++
		}
		for nodeID, count := range servicePlacementCountMap {
			node := nodeMap[nodeID]
			serviceRealtimeInfo.PlacementInfos = append(serviceRealtimeInfo.PlacementInfos, ServiceTaskPlacementInfo{
				NodeID:          nodeID,
				NodeName:        node.Description.Hostname,
				IsManagerNode:   node.Spec.Role != swarm.NodeRoleManager,
				RunningReplicas: count,
			})
			runningCount += count
		}
		// set service realtime info in map
		serviceRealtimeInfo.RunningReplicas = runningCount
		serviceRealtimeInfoMap[serviceRealtimeInfo.Name] = serviceRealtimeInfo
	}
	return serviceRealtimeInfoMap, nil
}

// Get service logs
func (m Manager) LogsService(serviceName string) (io.ReadCloser, error) {
	logs, err := m.client.ServiceLogs(m.ctx, serviceName, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return nil, errors.New("error getting service logs")
	}
	return logs, nil
}


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
			Type:     mount.TypeVolume,
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
