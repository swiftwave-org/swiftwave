package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
)

func (m Manager) DeployApplication(request DeployApplicationRequest, ctx context.Context, cancelContext context.CancelFunc) error {
	err := m.deployApplicationHelper(request)
	if err != nil {
		// mark as failed
		ctx := context.Background()
		addDeploymentLog(m.ServiceManager.DbClient, m.ServiceManager.PubSubClient, request.DeploymentId, "Deployment failed > \n"+err.Error()+"\n", false)
		deployment := &core.Deployment{}
		deployment.ID = request.DeploymentId
		err = deployment.UpdateStatus(ctx, m.ServiceManager.DbClient, core.DeploymentStatusFailed)
		if err != nil {
			log.Println("failed to update deployment status to failed", err)
		}
	}
	return nil
}

func (m Manager) deployApplicationHelper(request DeployApplicationRequest) error {
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
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Deployment starting...\n", false)
	// fetch environment variables
	environmentVariables, err := core.FindEnvironmentVariablesByApplicationId(ctx, *db, request.AppId)
	if err != nil {
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to fetch environment variables\n", false)
		return err
	}
	var environmentVariablesMap = make(map[string]string)
	for _, environmentVariable := range environmentVariables {
		environmentVariablesMap[environmentVariable.Key] = environmentVariable.Value
	}
	// fetch persistent volumes
	persistentVolumeBindings, err := core.FindPersistentVolumeBindingsByApplicationId(ctx, *db, request.AppId)
	if err != nil {
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to fetch persistent volumes\n", false)
		return err
	}
	var volumeMounts = make([]containermanger.VolumeMount, 0)
	for _, persistentVolumeBinding := range persistentVolumeBindings {
		// fetch the volume
		var persistentVolume core.PersistentVolume
		err := persistentVolume.FindById(ctx, dbWithoutTx, persistentVolumeBinding.PersistentVolumeID)
		if err != nil {
			addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to fetch persistent volume\n", false)
			return err
		}
		volumeMounts = append(volumeMounts, containermanger.VolumeMount{
			Source:   persistentVolume.Name,
			Target:   persistentVolumeBinding.MountingPath,
			ReadOnly: false,
		})
	}
	// docker pull image
	dockerImageUri := deployment.DeployableDockerImageURI()
	// check if image exists
	isImageExists := m.ServiceManager.DockerManager.ExistsImage(dockerImageUri)
	if isImageExists {
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Image already exists\n", false)
	} else {
		registryUsername := ""
		registryPassword := ""

		if deployment.ImageRegistryCredentialID != nil && *deployment.ImageRegistryCredentialID != 0 {
			// fetch image registry credential
			var imageRegistryCredential core.ImageRegistryCredential
			err := imageRegistryCredential.FindById(ctx, dbWithoutTx, *deployment.ImageRegistryCredentialID)
			if err != nil {
				addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to fetch image registry credential\n", false)
				return err
			}
			registryUsername = imageRegistryCredential.Username
			registryPassword = imageRegistryCredential.Password
		}

		scanner, err := m.ServiceManager.DockerManager.PullImage(deployment.DeployableDockerImageURI(), registryUsername, registryPassword)
		if err != nil {
			addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to pull docker image\n", false)
			return err
		}
		// read the logs
		if scanner != nil {
			var data map[string]interface{}
			for scanner.Scan() {
				err = json.Unmarshal(scanner.Bytes(), &data)
				if err != nil {
					continue
				}
				if data["status"] != nil {
					status := data["status"].(string)
					id := ""
					if data["id"] != nil {
						id = data["id"].(string)
					}
					if strings.HasPrefix(status, "Pulling from") ||
						strings.Compare(status, "Pulling fs layer") == 0 ||
						strings.Compare(status, "Verifying Checksum") == 0 ||
						strings.Compare(status, "Download complete") == 0 ||
						strings.Compare(status, "Pull complete") == 0 ||
						strings.HasPrefix(status, "Digest:") ||
						strings.HasPrefix(status, "Status:") {
						logContent := fmt.Sprintf("%s %s\n", status, id)
						addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, logContent, false)
					}

				}
			}
		}
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Image pulled successfully\n", false)
	}

	sysctls := make(map[string]string, 0)
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
	// create service
	service := containermanger.Service{
		Name:           application.Name,
		Image:          dockerImageUri,
		Command:        command,
		Env:            environmentVariablesMap,
		Networks:       []string{m.SystemConfig.ServiceConfig.NetworkName},
		DeploymentMode: containermanger.DeploymentMode(application.DeploymentMode),
		Replicas:       uint64(application.ReplicaCount()),
		VolumeMounts:   volumeMounts,
		Capabilities:   application.Capabilities,
		Sysctls:        sysctls,
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
	_, err = m.ServiceManager.DockerManager.GetService(service.Name)
	if err != nil {
		registryUsername := ""
		registryPassword := ""

		if deployment.ImageRegistryCredentialID != nil && *deployment.ImageRegistryCredentialID != 0 {
			// fetch image registry credential
			var imageRegistryCredential core.ImageRegistryCredential
			err := imageRegistryCredential.FindById(ctx, dbWithoutTx, *deployment.ImageRegistryCredentialID)
			if err != nil {
				addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to fetch image registry credential\n", false)
				return err
			}
			registryUsername = imageRegistryCredential.Username
			registryPassword = imageRegistryCredential.Password
		}

		// create service
		err = m.ServiceManager.DockerManager.CreateService(service, registryUsername, registryPassword)
		if err != nil {
			return err
		}
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Application deployed successfully\n", false)
	} else {
		// update service
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Application already exists, updating the application\n", false)
		err = m.ServiceManager.DockerManager.UpdateService(service)
		if err != nil {
			return err
		}
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Application re-deployed successfully\n", true)
	}
	// commit the changes
	err = db.Commit().Error
	// if error occurs rollback the service
	if err != nil {
		// rollback the service
		err = m.ServiceManager.DockerManager.RollbackService(service.Name)
		if err != nil {
			// don't throw error as it will create an un-recoverable state
			log.Println("failed to rollback service > "+service.Name, err)
			addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to rollback service\n", false)
		}
	} else {
		// update replicas count in proxy (don't throw error if it fails, only log the error)
		targetPorts, err := core.FetchIngressTargetPorts(ctx, dbWithoutTx, application.ID)
		if err == nil {
			// create new haproxy transaction
			haproxyTransactionId, err := m.ServiceManager.HaproxyManager.FetchNewTransactionId()
			if err != nil {
				log.Println("failed to create new haproxy transaction", err)
				addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to create new haproxy transaction\n", false)
			} else {
				for _, targetPort := range targetPorts {
					backendName := m.ServiceManager.HaproxyManager.GenerateBackendName(application.Name, targetPort)
					isBackendExist, err := m.ServiceManager.HaproxyManager.IsBackendExist(backendName)
					if err != nil {
						log.Println("failed to check if backend exist", err)
						addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to check if backend exist\n", false)
						continue
					}
					if isBackendExist {
						// fetch current replicas
						currentReplicaCount, err := m.ServiceManager.HaproxyManager.GetReplicaCount(haproxyTransactionId, application.Name, targetPort)
						if err != nil {
							log.Println("failed to fetch current replica count", err)
							addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to fetch current replica count\n", false)
							continue
						}
						// check if replica count changed
						if currentReplicaCount != int(application.ReplicaCount()) {
							err = m.ServiceManager.HaproxyManager.UpdateBackendReplicas(haproxyTransactionId, application.Name, targetPort, int(application.ReplicaCount()))
							if err != nil {
								log.Println("failed to update replica count", err)
								addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to update replica count\n", false)
							}
						}
					}
				}
			}
			// commit the haproxy transaction
			err = m.ServiceManager.HaproxyManager.CommitTransaction(haproxyTransactionId)
			if err != nil {
				log.Println("failed to commit haproxy transaction", err)
				addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to commit haproxy transaction\n", false)
				err := m.ServiceManager.HaproxyManager.DeleteTransaction(haproxyTransactionId)
				if err != nil {
					log.Println("failed to rollback haproxy transaction", err)
					addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to rollback haproxy transaction\n", false)
				}
			}
		} else {
			log.Println("failed to update replica count", err)
			addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to update replica count\n", false)
		}

	}
	return nil
}
