package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/swiftwave-org/swiftwave/container_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
	"log"
	"strings"
)

func (m Manager) DeployApplication(request DeployApplicationRequest) error {
	err := m.deployApplicationHelper(request)
	if err != nil {
		// mark as failed
		ctx := context.Background()
		addDeploymentLog(m.ServiceManager.DbClient, m.ServiceManager.PubSubClient, request.DeploymentId, "Deployment failed \n"+err.Error(), false)
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
	addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Deployment starting...", false)
	// fetch environment variables
	environmentVariables, err := core.FindEnvironmentVariablesByApplicationId(ctx, *db, request.AppId)
	if err != nil {
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to fetch environment variables", false)
		return err
	}
	var environmentVariablesMap = make(map[string]string)
	for _, environmentVariable := range environmentVariables {
		environmentVariablesMap[environmentVariable.Key] = environmentVariable.Value
	}
	// fetch persistent volumes
	persistentVolumeBindings, err := core.FindPersistentVolumeBindingsByApplicationId(ctx, *db, request.AppId)
	var volumeMounts = make([]containermanger.VolumeMount, 0)
	for _, persistentVolumeBinding := range persistentVolumeBindings {
		// fetch the volume
		var persistentVolume core.PersistentVolume
		err := persistentVolume.FindById(ctx, dbWithoutTx, persistentVolumeBinding.PersistentVolumeID)
		if err != nil {
			addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to fetch persistent volume", false)
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
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Image already exists", false)
	} else {
		scanner, err := m.ServiceManager.DockerManager.PullImage(deployment.DeployableDockerImageURI()) // TODO: add support for providing auth credentials
		if err != nil {
			addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to pull docker image", false)
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
						logContent := fmt.Sprintf("%s %s", status, id)
						addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, logContent, false)
					}

				}
			}
		}
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Image pulled successfully", false)
	}
	// create service
	service := containermanger.Service{
		Name:           application.Name,
		Image:          dockerImageUri,
		Command:        []string{},
		Env:            environmentVariablesMap,
		Networks:       []string{m.SystemConfig.ServiceConfig.NetworkName},
		DeploymentMode: containermanger.DeploymentMode(application.DeploymentMode),
		Replicas:       uint64(application.Replicas),
		VolumeMounts:   volumeMounts,
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
		// create service
		err = m.ServiceManager.DockerManager.CreateService(service)
		if err != nil {
			return err
		}
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Application deployed successfully", false)
	} else {
		// update service
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Application already exists, updating the application", false)
		err = m.ServiceManager.DockerManager.UpdateService(service)
		if err != nil {
			return err
		}
		addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Application re-deployed successfully", true)
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
			addDeploymentLog(dbWithoutTx, pubSubClient, deployment.ID, "Failed to rollback service", false)
		}
	}
	return err
}
