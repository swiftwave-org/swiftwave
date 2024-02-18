package worker

import (
	"context"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
	"path/filepath"
)

func (m Manager) PersistentVolumeRestore(request PersistentVolumeRestoreRequest, ctx context.Context, cancelContext context.CancelFunc) error {
	dbWithoutTx := m.ServiceManager.DbClient
	// fetch persistent volume restore
	var persistentVolumeRestore core.PersistentVolumeRestore
	err := persistentVolumeRestore.FindById(ctx, dbWithoutTx, request.Id)
	if err != nil {
		return nil
	}
	// check status should be uploaded
	if persistentVolumeRestore.Status != core.RestorePending {
		return nil
	}
	// fetch persistent volume
	var persistentVolume core.PersistentVolume
	err = persistentVolume.FindById(ctx, dbWithoutTx, persistentVolumeRestore.PersistentVolumeID)
	if err != nil {
		return nil
	}
	dockerManager := m.ServiceManager.DockerManager
	// restore backup
	filePath := filepath.Join(m.SystemConfig.ServiceConfig.DataDir, persistentVolumeRestore.File)
	err = dockerManager.RestoreVolume(persistentVolume.Name, filePath)
	if err != nil {
		markPVRestoreRequestAsFailed(dbWithoutTx, persistentVolumeRestore)
		return nil
	}
	// update status
	persistentVolumeRestore.Status = core.RestoreSuccess
	err = persistentVolumeRestore.Update(ctx, dbWithoutTx, m.SystemConfig.ServiceConfig.DataDir)
	return err
}

func markPVRestoreRequestAsFailed(dbWithoutTx gorm.DB, persistentVolumeRestore core.PersistentVolumeRestore) {
	persistentVolumeRestore.Status = core.RestoreFailed
	_ = persistentVolumeRestore.Update(context.Background(), dbWithoutTx, "")
}
