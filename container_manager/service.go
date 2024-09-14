package containermanger

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

const maxRetriesForVersionConflict = 5

func (m Manager) GetService(serviceName string) (Service, error) {
	serviceData, _, err := m.client.ServiceInspectWithRaw(m.ctx, serviceName, types.ServiceInspectOptions{})
	if err != nil {
		return Service{}, errors.New("error getting service")
	}
	// Create service object
	service := Service{
		Name:     serviceData.Spec.Name,
		Image:    serviceData.Spec.TaskTemplate.ContainerSpec.Image,
		Hostname: serviceData.Spec.TaskTemplate.ContainerSpec.Hostname,
		Command:  serviceData.Spec.TaskTemplate.ContainerSpec.Command,
		Env:      make(map[string]string),
		Networks: []string{},
		Replicas: 0,
	}
	// Set env
	for _, env := range serviceData.Spec.TaskTemplate.ContainerSpec.Env {
		// try to split env at first occurrence of '='
		envSplit := strings.SplitN(env, "=", 2)
		if len(envSplit) == 2 {
			service.Env[envSplit[0]] = envSplit[1]
		} else {
			service.Env[env] = ""
		}
	}
	// Set volume mounts and binds
	for _, volumeMount := range serviceData.Spec.TaskTemplate.ContainerSpec.Mounts {
		if volumeMount.Type == mount.TypeVolume {
			service.VolumeMounts = append(service.VolumeMounts, VolumeMount{
				Source:   volumeMount.Source,
				Target:   volumeMount.Target,
				ReadOnly: volumeMount.ReadOnly,
			})
		}
		if volumeMount.Type == mount.TypeBind {
			service.VolumeBinds = append(service.VolumeBinds, VolumeBind{
				Source: volumeMount.Source,
				Target: volumeMount.Target,
			})
		}
	}
	// set placement constraints
	if serviceData.Spec.TaskTemplate.Placement != nil {
		service.PlacementConstraints = serviceData.Spec.TaskTemplate.Placement.Constraints
	}
	// Set networks
	for _, network := range serviceData.Spec.TaskTemplate.Networks {
		service.Networks = append(service.Networks, network.Target)
	}
	// Set replicas
	if serviceData.Spec.Mode.Replicated != nil {
		service.Replicas = *serviceData.Spec.Mode.Replicated.Replicas
	}
	// Set deployment mode
	if serviceData.Spec.Mode.Replicated != nil {
		service.DeploymentMode = DeploymentModeReplicated
	} else {
		service.DeploymentMode = DeploymentModeGlobal
	}
	// Set Resource data
	if serviceData.Spec.TaskTemplate.Resources != nil && serviceData.Spec.TaskTemplate.Resources.Limits != nil {
		service.ResourceLimit = Resource{
			MemoryMB: int(serviceData.Spec.TaskTemplate.Resources.Limits.MemoryBytes / 1024 / 1024),
		}
	}
	// Set reserved resource
	if serviceData.Spec.TaskTemplate.Resources != nil && serviceData.Spec.TaskTemplate.Resources.Reservations != nil {
		service.ReservedResource = Resource{
			MemoryMB: int(serviceData.Spec.TaskTemplate.Resources.Reservations.MemoryBytes / 1024 / 1024),
		}
	}
	return service, nil
}

func (m Manager) CreateService(service Service, username string, password string, queryRegistry bool) error {
	authHeader, err := generateAuthHeader(username, password)
	if err != nil {
		return errors.New("failed to generate auth header")
	}
	// check if required networks exist
	for _, network := range service.Networks {
		if !m.ExistsNetwork(network) {
			// try to create network
			err := m.CreateNetwork(network)
			if err != nil {
				return errors.New("as network does not exist, while creating network" + network + " error occurred")
			}
		}
	}
	serviceSpec, err := m.serviceToServiceSpec(service)
	if err != nil {
		return err
	}
	_, err = m.client.ServiceCreate(m.ctx, serviceSpec, types.ServiceCreateOptions{
		EncodedRegistryAuth: authHeader,
		QueryRegistry:       queryRegistry,
	})
	if err != nil {
		return errors.New("error creating service")
	}
	return nil
}

func (m Manager) UpdateService(service Service, username string, password string, queryRegistry bool) error {
	authHeader, err := generateAuthHeader(username, password)
	if err != nil {
		return errors.New("failed to generate auth header")
	}

	maxRetries := maxRetriesForVersionConflict
	for {
		serviceData, _, err := m.client.ServiceInspectWithRaw(m.ctx, service.Name, types.ServiceInspectOptions{})
		if err != nil {
			return errors.New("error getting swarm server version")
		}
		version := swarm.Version{
			Index: serviceData.Version.Index,
		}
		serviceSpec, err := m.serviceToServiceSpec(service)
		if err != nil {
			return err
		}
		_, err = m.client.ServiceUpdate(m.ctx, service.Name, version, serviceSpec, types.ServiceUpdateOptions{
			EncodedRegistryAuth: authHeader,
			QueryRegistry:       queryRegistry,
		})
		if err != nil {
			if strings.Contains(err.Error(), "update out of sequence") {
				if maxRetries == 0 {
					return fmt.Errorf("error updating service due to version out of sync [retried %d times]", maxRetriesForVersionConflict)
				}
				<-time.After(3 * time.Second)
				maxRetries--
				continue
			}
			return errors.New("error updating service")
		} else {
			return nil
		}
	}
}

func (m Manager) RestartService(serviceName string) error {
	maxRetries := 5
	for {
		serviceData, _, err := m.client.ServiceInspectWithRaw(m.ctx, serviceName, types.ServiceInspectOptions{})
		if err != nil {
			return errors.New("error getting swarm server version")
		}
		version := swarm.Version{
			Index: serviceData.Version.Index,
		}
		spec := serviceData.Spec
		spec.TaskTemplate.ForceUpdate++
		_, err = m.client.ServiceUpdate(m.ctx, serviceName, version, spec, types.ServiceUpdateOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "update out of sequence") {
				if maxRetries == 0 {
					return fmt.Errorf("error updating service due to version out of sync [retried %d times]", maxRetriesForVersionConflict)
				}
				<-time.After(3 * time.Second)
				maxRetries--
				continue
			}
			return errors.New("error updating service")
		} else {
			return nil
		}
	}
}

func (m Manager) RollbackService(serviceName string) error {
	maxRetries := 5
	for {
		serviceData, _, err := m.client.ServiceInspectWithRaw(m.ctx, serviceName, types.ServiceInspectOptions{})
		if err != nil {
			return errors.New("error getting swarm server version")
		}
		version := swarm.Version{
			Index: serviceData.Version.Index,
		}
		_, err = m.client.ServiceUpdate(m.ctx, serviceName, version, *serviceData.PreviousSpec, types.ServiceUpdateOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "update out of sequence") {
				if maxRetries == 0 {
					return fmt.Errorf("error updating service due to version out of sync [retried %d times]", maxRetriesForVersionConflict)
				}
				<-time.After(3 * time.Second)
				maxRetries--
				continue
			}
			return errors.New("error updating service")
		} else {
			return nil
		}
	}
}

func (m Manager) RemoveService(serviceName string) error {
	err := m.client.ServiceRemove(m.ctx, serviceName)
	if err != nil {
		return errors.New("error removing service")
	}
	return nil
}

func (m Manager) SetServiceReplicaCount(serviceName string, replicas int) error {
	maxRetries := 5
	for {
		serviceData, _, err := m.client.ServiceInspectWithRaw(m.ctx, serviceName, types.ServiceInspectOptions{})
		if err != nil {
			return errors.New("error getting swarm server version")
		}
		version := swarm.Version{
			Index: serviceData.Version.Index,
		}
		spec := serviceData.Spec
		if spec.Mode.Replicated == nil {
			return errors.New("service is not a replicated service")
		}
		replicaCount := uint64(replicas)
		spec.Mode.Replicated.Replicas = &replicaCount
		_, err = m.client.ServiceUpdate(m.ctx, serviceName, version, spec, types.ServiceUpdateOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "update out of sequence") {
				if maxRetries == 0 {
					return fmt.Errorf("error updating service due to version out of sync [retried %d times]", maxRetriesForVersionConflict)
				}
				<-time.After(3 * time.Second)
				maxRetries--
				continue
			}
			return errors.New("error updating service")
		} else {
			return nil
		}
	}
}

func (m Manager) NoOfRunningTasks(serviceName string) (int, error) {
	// query task list
	tasks, err := m.client.TaskList(m.ctx, types.TaskListOptions{
		Filters: filters.NewArgs(
			filters.Arg("service", serviceName),
		),
	})
	if err != nil {
		return 0, err
	}
	runningCount := 0
	// set placement infos > how many replicas are running in each node
	for _, task := range tasks {
		if task.Status.State == swarm.TaskStateRunning {
			runningCount++
		}
	}
	return runningCount, nil
}

// ServiceRunningServers Fetch the servers where a service is running
func (m Manager) ServiceRunningServers(serviceName string) ([]string, error) {
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
	// analyze each service
	for _, service := range services {
		if service.Spec.Name == serviceName {
			// query task list
			tasks, err := m.client.TaskList(m.ctx, types.TaskListOptions{
				Filters: filters.NewArgs(
					filters.Arg("service", service.Spec.Name),
				),
			})
			if err != nil {
				return nil, errors.New("error getting task list")
			}
			var runningServers []string
			for _, task := range tasks {
				if task.Status.State == swarm.TaskStateRunning {
					runningServers = append(runningServers, nodeMap[task.NodeID].Description.Hostname)
				}
			}
			return runningServers, nil
		}
	}
	return nil, errors.New("service not found")
}

// RandomServiceContainerID returns a random container id of a service
func (m Manager) RandomServiceContainerID(serviceName string) (string, error) {
	containers, err := m.client.ContainerList(m.ctx, container.ListOptions{
		All: false,
		Filters: filters.NewArgs(
			filters.Arg("label", "com.docker.swarm.service.name="+serviceName),
		),
	})
	if err != nil {
		return "", errors.New("Failed to list containers for service " + serviceName + " " + err.Error())
	}
	if len(containers) == 0 {
		return "", errors.New("No containers found for service " + serviceName)
	}
	return containers[0].ID, nil
}

// LogsService Get service logs
func (m Manager) LogsService(serviceName string, sinceMinutes int) (io.ReadCloser, error) {
	containerLogOptions := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	}
	if sinceMinutes > 0 {
		containerLogOptions.Since = fmt.Sprintf("%dm", sinceMinutes)
	}
	logs, err := m.client.ServiceLogs(m.ctx, serviceName, containerLogOptions)
	if err != nil {
		return nil, errors.New("error getting service logs")
	}
	return logs, nil
}

// Private functions
func (m Manager) serviceToServiceSpec(service Service) (swarm.ServiceSpec, error) {
	// Create swarm attachment config from network names array
	var networkAttachmentConfigs []swarm.NetworkAttachmentConfig
	for _, networkName := range service.Networks {
		networkAttachmentConfigs = append(networkAttachmentConfigs, swarm.NetworkAttachmentConfig{
			Target: networkName,
		})
	}

	// Create volume mounts from volume mounts array
	var volumeMounts []mount.Mount
	for _, volumeMount := range service.VolumeMounts {
		volumeMounts = append(volumeMounts, mount.Mount{
			Type:     mount.TypeVolume,
			Source:   volumeMount.Source,
			Target:   volumeMount.Target,
			ReadOnly: volumeMount.ReadOnly,
		})
	}

	for _, volumeBind := range service.VolumeBinds {
		volumeMounts = append(volumeMounts, mount.Mount{
			Type:     mount.TypeBind,
			Source:   volumeBind.Source,
			Target:   volumeBind.Target,
			ReadOnly: false,
		})
	}

	// Create `ENV_VAR=value` array from env map
	var env []string
	for key, value := range service.Env {
		env = append(env, key+"="+value)
	}

	var serviceMode swarm.ServiceMode

	if service.DeploymentMode == DeploymentModeReplicated {
		// allow replicated service
		serviceMode = swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &service.Replicas,
			},
		}
	} else if service.DeploymentMode == DeploymentModeGlobal {
		// allow global service
		serviceMode = swarm.ServiceMode{
			Global: &swarm.GlobalService{},
		}
	} else {
		print(service.DeploymentMode)
		panic("invalid deployment mode > ")
	}

	// config references
	var configs = make([]*swarm.ConfigReference, 0)
	for _, config := range service.ConfigMounts {
		configId, err := m.FetchDockerConfigId(config.ConfigID)
		if err != nil {
			return swarm.ServiceSpec{}, err
		}
		configs = append(configs, &swarm.ConfigReference{
			ConfigName: config.ConfigID,
			ConfigID:   configId,
			File: &swarm.ConfigReferenceFileTarget{
				Name: config.MountingPath,
				UID:  strconv.Itoa(int(config.Uid)),
				GID:  strconv.Itoa(int(config.Gid)),
				Mode: os.FileMode(uint32(config.FileMode)),
			},
		})
	}

	// memory bytes
	var reservedMemoryBytes int64 = 0
	if service.ReservedResource.MemoryMB >= 6 {
		reservedMemoryBytes = int64(service.ReservedResource.MemoryMB * 1024 * 1024)
	}
	var limitMemoryBytes int64 = 0
	if service.ResourceLimit.MemoryMB >= 6 {
		limitMemoryBytes = int64(service.ResourceLimit.MemoryMB * 1024 * 1024)
	}

	// healthcheck
	var healthCheck *container.HealthConfig
	if service.CustomHealthCheck.Enabled {
		healthCheck = &container.HealthConfig{
			Test:          []string{"CMD-SHELL", service.CustomHealthCheck.TestCommand},
			Interval:      time.Duration(service.CustomHealthCheck.IntervalSeconds) * time.Second,
			Timeout:       time.Duration(service.CustomHealthCheck.TimeoutSeconds) * time.Second,
			StartPeriod:   time.Duration(service.CustomHealthCheck.StartPeriodSeconds) * time.Second,
			StartInterval: time.Duration(service.CustomHealthCheck.StartIntervalSeconds) * time.Second,
			Retries:       int(service.CustomHealthCheck.Retries),
		}
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
				Image:    service.Image,
				Command:  service.Command,
				Hostname: service.Hostname,
				Env:      env,
				Mounts:   volumeMounts,
				Configs:  configs,
				Privileges: &swarm.Privileges{
					NoNewPrivileges: true,
					AppArmor: &swarm.AppArmorOpts{
						Mode: swarm.AppArmorModeDefault,
					},
					Seccomp: &swarm.SeccompOpts{
						Mode: swarm.SeccompModeDefault,
					},
				},
				CapabilityAdd: service.Capabilities,
				Sysctls:       service.Sysctls,
				Healthcheck:   healthCheck,
			},
			Placement: &swarm.Placement{
				Constraints: service.PlacementConstraints,
			},
			Resources: &swarm.ResourceRequirements{
				Reservations: &swarm.Resources{
					MemoryBytes: reservedMemoryBytes,
				},
				Limits: &swarm.Limit{
					MemoryBytes: limitMemoryBytes,
				},
			},
			// Set network name
			Networks: networkAttachmentConfigs,
		},
		// allow replicated service
		Mode: serviceMode,
		// constant endpoint
		EndpointSpec: &swarm.EndpointSpec{
			Mode: swarm.ResolutionModeDNSRR,
		},
	}
	return serviceSpec, nil
}
