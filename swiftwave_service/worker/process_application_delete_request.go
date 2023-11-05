package worker

import (
	"context"
	"errors"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
	"log"
)

func (m Manager) DeleteApplication(request DeleteApplicationRequest) error {
	dbWithoutTx := m.ServiceManager.DbClient
	ctx := context.Background()
	dockerManager := m.ServiceManager.DockerManager
	// find application
	var application core.Application
	err := application.FindById(ctx, m.ServiceManager.DbClient, request.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// return nil as don't want to requeue the job
			return nil
		} else {
			return err
		}
	}
	// start a db transaction
	tx := dbWithoutTx.Begin()
	// delete application
	err = application.Delete(ctx, *tx, dockerManager)
	if err != nil {
		tx.Rollback()
		return err
	}
	// delete persistent volume bindings
	err = core.DeletePersistentVolumeBindingsByApplicationId(ctx, *tx, request.Id)
	if err != nil {
		tx.Rollback()
		return err
	}
	// delete environment variables
	err = core.DeleteEnvironmentVariablesByApplicationId(ctx, *tx, request.Id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// delete deployments
	deploymentIds, err := core.DeleteDeploymentsByApplicationId(ctx, *tx, dockerManager, request.Id)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, deploymentId := range deploymentIds {
		// delete build args
		err = core.DeleteBuildArgsByDeploymentId(ctx, *tx, deploymentId)
		if err != nil {
			tx.Rollback()
			return err
		}
		// delete build logs
		err = core.DeleteBuildLogsByDeploymentId(ctx, *tx, deploymentId)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// commit the transaction
	err = tx.Commit().Error
	if err != nil {
		return err
	}

	// delete application from swarm manager
	err = dockerManager.RemoveService(application.Name)
	if err != nil {
		log.Println("error deleting application from swarm manager : " + application.Name)
	}

	return nil
}
