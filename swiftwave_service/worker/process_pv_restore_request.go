package worker

import (
	"context"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/manager"
	"gorm.io/gorm"
	"os"
	"path/filepath"
)

func (m Manager) PersistentVolumeRestore(request PersistentVolumeRestoreRequest, ctx context.Context, _ context.CancelFunc) error {
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
	// fetch swarm server
	server, err := core.FetchSwarmManager(&dbWithoutTx)
	if err != nil {
		return err
	}
	dockerManager, err := manager.DockerClient(ctx, server)
	if err != nil {
		return err
	}
	// restore backup
	localRestoreFilePath := filepath.Join(m.Config.LocalConfig.ServiceConfig.PVRestoreDirectoryPath, persistentVolumeRestore.File)
	err = dockerManager.RestoreVolume(persistentVolume.Name, localRestoreFilePath, server.IP, 22, server.User, m.Config.SystemConfig.SshPrivateKey)
	if err != nil {
		markPVRestoreRequestAsFailed(dbWithoutTx, persistentVolumeRestore)
		_ = os.RemoveAll(localRestoreFilePath)
		return nil
	}
	// update status
	persistentVolumeRestore.Status = core.RestoreSuccess
	err = persistentVolumeRestore.Update(ctx, dbWithoutTx, m.Config.LocalConfig.ServiceConfig.PVRestoreDirectoryPath)
	if err != nil {
		return err
	}
	// remove local file
	_ = os.RemoveAll(localRestoreFilePath)
	return nil
}

func markPVRestoreRequestAsFailed(dbWithoutTx gorm.DB, persistentVolumeRestore core.PersistentVolumeRestore) {
	persistentVolumeRestore.Status = core.RestoreFailed
	_ = persistentVolumeRestore.Update(context.Background(), dbWithoutTx, "")

}
