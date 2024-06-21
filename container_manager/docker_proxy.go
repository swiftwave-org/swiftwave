package containermanger

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"strings"
	"time"
)

const dockerProxyImage = "ghcr.io/swiftwave-org/docker-socket-proxy:latest"

func (m Manager) RemoveDockerProxy(serviceName string) {
	_ = m.RemoveService(serviceName)
}

func (m Manager) CreateDockerProxy(serviceName string, placementConstraints []string, config DockerProxyConfig, networkName string) error {
	var replicaCount uint64 = 1

	environmentVariables := []string{
		fmt.Sprintf("PING_READ=%d", boolToInt(config.Permission.Ping == DockerProxyReadPermission || config.Permission.Ping == DockerProxyReadWritePermission)),
		fmt.Sprintf("PING_WRITE=%d", boolToInt(config.Permission.Ping == DockerProxyReadWritePermission)),
		fmt.Sprintf("VERSION_READ=%d", boolToInt(config.Permission.Version == DockerProxyReadPermission || config.Permission.Version == DockerProxyReadWritePermission)),
		fmt.Sprintf("VERSION_WRITE=%d", boolToInt(config.Permission.Version == DockerProxyReadWritePermission)),
		fmt.Sprintf("INFO_READ=%d", boolToInt(config.Permission.Info == DockerProxyReadPermission || config.Permission.Info == DockerProxyReadWritePermission)),
		fmt.Sprintf("INFO_WRITE=%d", boolToInt(config.Permission.Info == DockerProxyReadWritePermission)),
		fmt.Sprintf("EVENTS_READ=%d", boolToInt(config.Permission.Events == DockerProxyReadPermission || config.Permission.Events == DockerProxyReadWritePermission)),
		fmt.Sprintf("EVENTS_WRITE=%d", boolToInt(config.Permission.Events == DockerProxyReadWritePermission)),
		fmt.Sprintf("AUTH_READ=%d", boolToInt(config.Permission.Auth == DockerProxyReadPermission || config.Permission.Auth == DockerProxyReadWritePermission)),
		fmt.Sprintf("AUTH_WRITE=%d", boolToInt(config.Permission.Auth == DockerProxyReadWritePermission)),
		fmt.Sprintf("SECRETS_READ=%d", boolToInt(config.Permission.Secrets == DockerProxyReadPermission || config.Permission.Secrets == DockerProxyReadWritePermission)),
		fmt.Sprintf("SECRETS_WRITE=%d", boolToInt(config.Permission.Secrets == DockerProxyReadWritePermission)),
		fmt.Sprintf("BUILD_READ=%d", boolToInt(config.Permission.Build == DockerProxyReadPermission || config.Permission.Build == DockerProxyReadWritePermission)),
		fmt.Sprintf("BUILD_WRITE=%d", boolToInt(config.Permission.Build == DockerProxyReadWritePermission)),
		fmt.Sprintf("COMMIT_READ=%d", boolToInt(config.Permission.Commit == DockerProxyReadPermission || config.Permission.Commit == DockerProxyReadWritePermission)),
		fmt.Sprintf("COMMIT_WRITE=%d", boolToInt(config.Permission.Commit == DockerProxyReadWritePermission)),
		fmt.Sprintf("CONFIGS_READ=%d", boolToInt(config.Permission.Configs == DockerProxyReadPermission || config.Permission.Configs == DockerProxyReadWritePermission)),
		fmt.Sprintf("CONFIGS_WRITE=%d", boolToInt(config.Permission.Configs == DockerProxyReadWritePermission)),
		fmt.Sprintf("CONTAINERS_READ=%d", boolToInt(config.Permission.Containers == DockerProxyReadPermission || config.Permission.Containers == DockerProxyReadWritePermission)),
		fmt.Sprintf("CONTAINERS_WRITE=%d", boolToInt(config.Permission.Containers == DockerProxyReadWritePermission)),
		fmt.Sprintf("DISTRIBUTION_READ=%d", boolToInt(config.Permission.Distribution == DockerProxyReadPermission || config.Permission.Distribution == DockerProxyReadWritePermission)),
		fmt.Sprintf("DISTRIBUTION_WRITE=%d", boolToInt(config.Permission.Distribution == DockerProxyReadWritePermission)),
		fmt.Sprintf("EXEC_READ=%d", boolToInt(config.Permission.Exec == DockerProxyReadPermission || config.Permission.Exec == DockerProxyReadWritePermission)),
		fmt.Sprintf("EXEC_WRITE=%d", boolToInt(config.Permission.Exec == DockerProxyReadWritePermission)),
		fmt.Sprintf("GRPC_READ=%d", boolToInt(config.Permission.Exec == DockerProxyReadPermission || config.Permission.Exec == DockerProxyReadWritePermission)),
		fmt.Sprintf("GRPC_WRITE=%d", boolToInt(config.Permission.Exec == DockerProxyReadWritePermission)),
		fmt.Sprintf("IMAGES_READ=%d", boolToInt(config.Permission.Exec == DockerProxyReadPermission || config.Permission.Exec == DockerProxyReadWritePermission)),
		fmt.Sprintf("IMAGES_WRITE=%d", boolToInt(config.Permission.Exec == DockerProxyReadWritePermission)),
		fmt.Sprintf("NETWORKS_READ=%d", boolToInt(config.Permission.Networks == DockerProxyReadPermission || config.Permission.Networks == DockerProxyReadWritePermission)),
		fmt.Sprintf("NETWORKS_WRITE=%d", boolToInt(config.Permission.Networks == DockerProxyReadWritePermission)),
		fmt.Sprintf("NODES_READ=%d", boolToInt(config.Permission.Nodes == DockerProxyReadPermission || config.Permission.Nodes == DockerProxyReadWritePermission)),
		fmt.Sprintf("NODES_WRITE=%d", boolToInt(config.Permission.Nodes == DockerProxyReadWritePermission)),
		fmt.Sprintf("PLUGINS_READ=%d", boolToInt(config.Permission.Plugins == DockerProxyReadPermission || config.Permission.Plugins == DockerProxyReadWritePermission)),
		fmt.Sprintf("PLUGINS_WRITE=%d", boolToInt(config.Permission.Plugins == DockerProxyReadWritePermission)),
		fmt.Sprintf("SERVICES_READ=%d", boolToInt(config.Permission.Services == DockerProxyReadPermission || config.Permission.Services == DockerProxyReadWritePermission)),
		fmt.Sprintf("SERVICES_WRITE=%d", boolToInt(config.Permission.Services == DockerProxyReadWritePermission)),
		fmt.Sprintf("SESSION_READ=%d", boolToInt(config.Permission.Session == DockerProxyReadPermission || config.Permission.Session == DockerProxyReadWritePermission)),
		fmt.Sprintf("SESSION_WRITE=%d", boolToInt(config.Permission.Session == DockerProxyReadWritePermission)),
		fmt.Sprintf("SWARM_READ=%d", boolToInt(config.Permission.Swarm == DockerProxyReadPermission || config.Permission.Swarm == DockerProxyReadWritePermission)),
		fmt.Sprintf("SWARM_WRITE=%d", boolToInt(config.Permission.Swarm == DockerProxyReadWritePermission)),
		fmt.Sprintf("SYSTEM_READ=%d", boolToInt(config.Permission.System == DockerProxyReadPermission || config.Permission.System == DockerProxyReadWritePermission)),
		fmt.Sprintf("SYSTEM_WRITE=%d", boolToInt(config.Permission.System == DockerProxyReadWritePermission)),
		fmt.Sprintf("TASKS_READ=%d", boolToInt(config.Permission.Tasks == DockerProxyReadPermission || config.Permission.Tasks == DockerProxyReadWritePermission)),
		fmt.Sprintf("TASKS_WRITE=%d", boolToInt(config.Permission.Tasks == DockerProxyReadWritePermission)),
		fmt.Sprintf("VOLUMES_READ=%d", boolToInt(config.Permission.Volumes == DockerProxyReadPermission || config.Permission.Volumes == DockerProxyReadWritePermission)),
		fmt.Sprintf("VOLUMES_WRITE=%d", boolToInt(config.Permission.Volumes == DockerProxyReadWritePermission)),
	}

	// Currently, there will be only changes to the environment variables and placement constraints,so those wil be checked for changes
	isUpdate := false
	existingService, _, err := m.client.ServiceInspectWithRaw(m.ctx, serviceName, types.ServiceInspectOptions{})
	if err == nil {
		if existingService.Spec.TaskTemplate.ContainerSpec != nil {
			envVars := existingService.Spec.TaskTemplate.ContainerSpec.Env
			// if env vars are the same, do not update
			if !isSameList(envVars, environmentVariables) {
				isUpdate = true
			}
		}
		if existingService.Spec.TaskTemplate.Placement != nil {
			constraints := existingService.Spec.TaskTemplate.Placement.Constraints
			// if constraints are the same, do not update
			if !isSameList(constraints, placementConstraints) {
				isUpdate = true
			}
		}

		if !isUpdate {
			return nil
		}
	}

	serviceSpec := swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: serviceName,
		},
		// Set task template
		TaskTemplate: swarm.TaskSpec{
			// Set container spec
			ContainerSpec: &swarm.ContainerSpec{
				Image:   dockerProxyImage,
				Command: []string{},
				Env:     environmentVariables,
				Mounts: []mount.Mount{
					{
						Type:     mount.TypeBind,
						Source:   "/var/run/docker.sock",
						Target:   "/var/run/docker.sock",
						ReadOnly: true,
					},
				},
				Configs: nil,
				Privileges: &swarm.Privileges{
					NoNewPrivileges: false,
					AppArmor: &swarm.AppArmorOpts{
						Mode: swarm.AppArmorModeDefault,
					},
					Seccomp: &swarm.SeccompOpts{
						Mode: swarm.SeccompModeDefault,
					},
				},
				CapabilityAdd: []string{
					"CAP_DAC_OVERRIDE",
				},
				Sysctls: map[string]string{},
			},
			Placement: &swarm.Placement{
				Constraints: placementConstraints,
			},
			Resources: &swarm.ResourceRequirements{
				Reservations: &swarm.Resources{
					MemoryBytes: 0,
				},
				Limits: &swarm.Limit{
					MemoryBytes: 0,
				},
			},
			// Set network name
			Networks: []swarm.NetworkAttachmentConfig{
				{
					Target: networkName,
				},
			},
		},
		// allow replicated service
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &replicaCount,
			},
		},
		// constant endpoint
		EndpointSpec: &swarm.EndpointSpec{
			Mode: swarm.ResolutionModeDNSRR,
		},
	}

	if isUpdate {
		maxRetries := maxRetriesForVersionConflict
		for {
			serviceData, _, err := m.client.ServiceInspectWithRaw(m.ctx, serviceName, types.ServiceInspectOptions{})
			if err != nil {
				return errors.New("error getting swarm server version")
			}
			version := swarm.Version{
				Index: serviceData.Version.Index,
			}
			_, err = m.client.ServiceUpdate(m.ctx, serviceName, version, serviceSpec, types.ServiceUpdateOptions{
				QueryRegistry: true,
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
	} else {
		_, err := m.client.ServiceCreate(m.ctx, serviceSpec, types.ServiceCreateOptions{
			QueryRegistry: true,
		})
		return err
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func isSameList(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		isFound := false
		for j := range b {
			if strings.Compare(a[i], b[j]) == 0 {
				isFound = true
				break
			}
		}
		if !isFound {
			return false
		}
	}
	return true
}
