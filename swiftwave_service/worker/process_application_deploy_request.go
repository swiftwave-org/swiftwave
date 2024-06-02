package worker

import (
	"context"
	"errors"
	haproxymanager "github.com/swiftwave-org/swiftwave/haproxy_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/manager"
	"log"
	"strings"

	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
)

func (m Manager) DeployApplication(request DeployApplicationRequest, _ context.Context, _ context.CancelFunc) error {
	// fetch the swarm server
	swarmManager, err := core.FetchSwarmManager(&m.ServiceManager.DbClient)
	if err != nil {
		return err
	}
	// create docker manager
	dockerManager, err := manager.DockerClient(context.Background(), swarmManager)
	if err != nil {
		return err
	}
	// fetch all proxy servers
	proxyServers := make([]core.Server, 0)
	if !request.IgnoreProxyUpdate {
		proxyServers, err = core.FetchProxyActiveServers(&m.ServiceManager.DbClient)
		if err != nil {
			return err
		}
	}
	// fetch all haproxy managers
	haproxyManagers, err := manager.HAProxyClients(context.Background(), proxyServers)
	if err != nil {
		return err
	}
	err = m.deployApplicationHelper(request, dockerManager, haproxyManagers)
	if err != nil {
		// mark as failed
		ctx := context.Background()
		addPersistentDeploymentLog(m.ServiceManager.DbClient, m.ServiceManager.PubSubClient, request.DeploymentId, "Deployment failed > \n"+err.Error()+"\n", false)
		deployment := &core.Deployment{}
		deployment.ID = request.DeploymentId
		err = deployment.UpdateStatus(ctx, m.ServiceManager.DbClient, core.DeploymentStatusFailed)
		if err != nil {
			log.Println("failed to update deployment status to failed", err)
		}
	}
	return nil
}

func (m Manager) deployApplicationHelper(request DeployApplicationRequest, dockerManager *containermanger.Manager, haproxyManagers []*haproxymanager.Manager) error {
	// context
	ctx := context.Background()
	dbWithoutTx := m.ServiceManager.DbClient
	db := m.ServiceManager.DbClient.Begin()
	defer func() {
		db.Rollback()
	}()
	// pubSub client
	pubSubClient := m.ServiceManager.PubSubClient
	// fetch application
	var application core.Application
	err := application.FindById(ctx, *db, request.AppId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// return nil as don't want to requeue the job
			return nil
		} else {
			return err
		}
	}
	// fetch deployment
	deployment := &core.Deployment{}
	deployment.ID = request.DeploymentId
	err = deployment.FindById(ctx, *db, request.DeploymentId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// create new deployment
			return nil
		} else {
			return err
		}
	}
	// log message
	addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Deployment starting...\n", false)
	// fetch environment variables
	environmentVariables, err := core.FindEnvironmentVariablesByApplicationId(ctx, *db, request.AppId)
	if err != nil {
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to fetch environment variables\n", false)
		return err
	}
	var environmentVariablesMap = make(map[string]string)
	for _, environmentVariable := range environmentVariables {
		environmentVariablesMap[environmentVariable.Key] = environmentVariable.Value
	}
	// fetch persistent volumes
	persistentVolumeBindings, err := core.FindPersistentVolumeBindingsByApplicationId(ctx, *db, request.AppId)
	if err != nil {
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to fetch persistent volumes\n", false)
		return err
	}
	var volumeMounts = make([]containermanger.VolumeMount, 0)
	for _, persistentVolumeBinding := range persistentVolumeBindings {
		// fetch the volume
		var persistentVolume core.PersistentVolume
		err := persistentVolume.FindById(ctx, dbWithoutTx, persistentVolumeBinding.PersistentVolumeID)
		if err != nil {
			addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to fetch persistent volume\n", false)
			return err
		}
		volumeMounts = append(volumeMounts, containermanger.VolumeMount{
			Source:   persistentVolume.Name,
			Target:   persistentVolumeBinding.MountingPath,
			ReadOnly: false,
		})
	}
	sysctls := make(map[string]string)
	for _, sysctl := range application.Sysctls {
		sysctlPart := strings.SplitN(sysctl, "=", 2)
		if len(sysctlPart) == 2 {
			sysctls[sysctlPart[0]] = sysctlPart[1]
		}
	}
	command := make([]string, 0)
	if application.Command != "" {
		command = strings.Split(application.Command, " ")
	}
	// docker image info
	dockerImageUri := deployment.DeployableDockerImageURI(m.Config.ImageRegistryURI())
	refetchImage := false
	imageRegistryUsername := m.Config.ImageRegistryUsername()
	imageRegistryPassword := m.Config.ImageRegistryPassword()

	if deployment.UpstreamType == core.UpstreamTypeImage {
		// fetch image registry credential
		if deployment.ImageRegistryCredentialID != nil && *deployment.ImageRegistryCredentialID != 0 {
			var imageRegistryCredential core.ImageRegistryCredential
			err := imageRegistryCredential.FindById(ctx, dbWithoutTx, *deployment.ImageRegistryCredentialID)
			if err != nil {
				addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to fetch image registry credential\n", false)
				return err
			}
			imageRegistryUsername = imageRegistryCredential.Username
			imageRegistryPassword = imageRegistryCredential.Password
		} else {
			imageRegistryUsername = ""
			imageRegistryPassword = ""
		}
		refetchImage = true
	}

	if refetchImage {
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "[Notice] Image will be fetched from remote during deployment\n", false)
	}
	// prepare placement constraints
	var placementConstraints = make([]string, 0)
	disabledServerHostnames, err := core.FetchDisabledDeploymentServerHostNames(&m.ServiceManager.DbClient)
	if err != nil {
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to fetch disabled deployment servers\nPlease check database connection\n", false)
		return err
	}
	for _, hostname := range disabledServerHostnames {
		placementConstraints = append(placementConstraints, "node.hostname!="+hostname)
	}
	// create service
	service := containermanger.Service{
		Name:                 application.Name,
		Image:                dockerImageUri,
		Command:              command,
		Env:                  environmentVariablesMap,
		Networks:             []string{m.Config.SystemConfig.NetworkName},
		DeploymentMode:       containermanger.DeploymentMode(application.DeploymentMode),
		Replicas:             uint64(application.ReplicaCount()),
		VolumeMounts:         volumeMounts,
		Capabilities:         application.Capabilities,
		Sysctls:              sysctls,
		PlacementConstraints: placementConstraints,
		ResourceLimit: containermanger.Resource{
			MemoryMB: application.ResourceLimit.MemoryMB,
		},
		ReservedResource: containermanger.Resource{
			MemoryMB: application.ReservedResource.MemoryMB,
		},
	}
	// find current deployment and mark it as stalled
	currentDeployment, err := core.FindCurrentLiveDeploymentByApplicationId(ctx, *db, request.AppId)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	} else {
		// Update status to stalled
		err = currentDeployment.UpdateStatus(ctx, *db, core.DeploymentStalled)
		if err != nil {
			return err
		}
	}
	// update deployment status
	err = deployment.UpdateStatus(ctx, *db, core.DeploymentStatusLive)
	if err != nil {
		return err
	}

	// check if the service already exists
	_, err = dockerManager.GetService(service.Name)
	if err != nil {
		// create service
		err = dockerManager.CreateService(service, imageRegistryUsername, imageRegistryPassword, refetchImage)
		if err != nil {
			return err
		}
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Application deployed successfully\n", false)
	} else {
		// update service
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Application already exists, updating the application\n", false)
		err = dockerManager.UpdateService(service, imageRegistryUsername, imageRegistryPassword, refetchImage)
		if err != nil {
			return err
		}
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Application re-deployed successfully\n", true)
	}
	// commit the changes
	err = db.Commit().Error
	// if error occurs rollback the service
	if err != nil {
		// rollback the service
		err = dockerManager.RollbackService(service.Name)
		if err != nil {
			// don't throw error as it will create an un-recoverable state
			log.Println("failed to rollback service > "+service.Name, err)
			addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to rollback service\n", false)
		}
	}

	if !request.IgnoreProxyUpdate {
		// update replicas count in proxy (don't throw error if it fails, only log the error)
		ingressRulesWithTargetPortAndProtocolOnly, err := core.FetchIngressRulesWithTargetPortAndProtocolOnly(ctx, dbWithoutTx, application.ID)
		if err == nil {
			// map of server ip and transaction id
			transactionIdMap := make(map[*haproxymanager.Manager]string)
			isFailed := false

			for _, haproxyManager := range haproxyManagers {
				// create new haproxy transaction
				haproxyTransactionId, err := haproxyManager.FetchNewTransactionId()
				if err != nil {
					isFailed = true
					log.Println("failed to create new haproxy transaction", err)
					addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to create new haproxy transaction\n", false)
					break
				} else {
					transactionIdMap[haproxyManager] = haproxyTransactionId
					for _, record := range ingressRulesWithTargetPortAndProtocolOnly {
						if record.Protocol == core.UDPProtocol {
							continue
						}
						backendProtocol := ingressRuleProtocolToBackendProtocol(record.Protocol)
						backendName := haproxyManager.GenerateBackendName(backendProtocol, application.Name, int(record.TargetPort))
						isBackendExist, err := haproxyManager.IsBackendExist(haproxyTransactionId, backendName)
						if err != nil {
							isFailed = true
							log.Println("failed to check if backend exist", err)
							addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to check if backend exist\n", false)
							continue
						}
						if isBackendExist {
							// fetch current replicas
							currentReplicaCount, err := haproxyManager.GetReplicaCount(haproxyTransactionId, backendProtocol, application.Name, int(record.TargetPort))
							if err != nil {
								isFailed = true
								log.Println("failed to fetch current replica count", err)
								addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to fetch current replica count\n", false)
								continue
							}
							// check if replica count changed
							if currentReplicaCount != int(application.ReplicaCount()) {
								err = haproxyManager.UpdateBackendReplicas(haproxyTransactionId, backendProtocol, application.Name, int(record.TargetPort), int(application.ReplicaCount()))
								if err != nil {
									isFailed = true
									log.Println("failed to update replica count", err)
									addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to update replica count\n", false)
								}
							}
						}
					}
				}
			}

			for haproxyManager, haproxyTransactionId := range transactionIdMap {
				if !isFailed {
					// commit the haproxy transaction
					err = haproxyManager.CommitTransaction(haproxyTransactionId)
				}
				if isFailed || err != nil {
					log.Println("failed to commit haproxy transaction", err)
					addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to commit haproxy transaction\n", false)
					err := haproxyManager.DeleteTransaction(haproxyTransactionId)
					if err != nil {
						log.Println("failed to rollback haproxy transaction", err)
						addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to rollback haproxy transaction\n", false)
					}
				}
			}
		} else {
			log.Println("failed to update replica count", err)
			addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to update replica count\n", false)
		}
	} else {
		addPersistentDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "[Notice] Ignoring proxy update as it's not requested\n", false)
	}
	return nil
}

// private functions
func ingressRuleProtocolToBackendProtocol(protocol core.ProtocolType) haproxymanager.BackendProtocol {
	if protocol == core.HTTPProtocol || protocol == core.HTTPSProtocol {
		return haproxymanager.HTTPBackend
	}
	if protocol == core.TCPProtocol {
		return haproxymanager.TCPBackend
	}
	if protocol == core.UDPProtocol {
		logger.CronJobLoggerError.Println("ingressRuleProtocolToBackendProtocol should not be called for UDP protocol. Report this issue to the team")
	}
	return haproxymanager.HTTPBackend
}

func isHAProxyAccessRequired(ingressRule *core.IngressRule) bool {
	if ingressRule.Protocol == core.HTTPProtocol || ingressRule.Protocol == core.HTTPSProtocol || ingressRule.Protocol == core.TCPProtocol {
		return true
	}
	return false
}

func isUDProxyAccessRequired(ingressRule *core.IngressRule) bool {
	return ingressRule.Protocol == core.UDPProtocol
}
