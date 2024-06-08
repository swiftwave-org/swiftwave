package worker

import (
	"context"
	"errors"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/manager"
	"gorm.io/gorm"
	"log"
)

func (m Manager) DeleteApplication(request DeleteApplicationRequest, ctx context.Context, _ context.CancelFunc) error {
	dbWithoutTx := m.ServiceManager.DbClient
	// fetch the swarm server
	swarmManager, err := core.FetchSwarmManager(&dbWithoutTx)
	if err != nil {
		return err
	}
	// create docker manager
	dockerManager, err := manager.DockerClient(context.Background(), swarmManager)
	if err != nil {
		return err
	}
	// find application
	var application core.Application
	err = application.FindById(ctx, m.ServiceManager.DbClient, request.Id)
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
	err = application.HardDelete(ctx, *tx, *dockerManager)
	if err != nil {
		tx.Rollback()
		return err
	}

	// commit the transaction
	err = tx.Commit().Error
	if err != nil {
		return err
	}

	// delete application from swarm manager
	err = dockerManager.RemoveService(application.Name)
	if err != nil {
		log.Println("[WARN] error deleting application from swarm manager : " + application.Name)
	}
	// prune config mounts
	dockerManager.PruneConfig(application.ID)
	return nil
}
