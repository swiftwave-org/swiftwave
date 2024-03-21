package worker

import (
	"context"
	"errors"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/manager"
	"gorm.io/gorm"
)

func (m Manager) PersistentVolumeDeletion(request PersistentVolumeDeletionRequest, ctx context.Context, _ context.CancelFunc) error {
	// Fetch volume
	var volume core.PersistentVolume
	err := volume.FindById(ctx, m.ServiceManager.DbClient, request.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	// Fetch all servers
	servers, err := core.FetchAllServers(&m.ServiceManager.DbClient)
	if err != nil {
		return err
	}
	// Delete volume from all servers
	isDeleted := false
	for _, server := range servers {
		dockerManager, err := manager.DockerClient(ctx, server)
		if err != nil {
			return err
		}
		err = dockerManager.RemoveVolume(volume.Name)
		if err != nil {
			logger.WorkerLoggerError.Println("Error deleting volume", volume.Name, " from server", server.ID, err.Error())
		} else {
			isDeleted = true
		}
	}
	if !isDeleted {
		return errors.New("error deleting volume")
	}
	// Delete volume from database
	err = volume.Delete(ctx, m.ServiceManager.DbClient)
	if err != nil {
		logger.WorkerLoggerError.Println("Error deleting volume from database", err.Error())
	}
	return nil
}
