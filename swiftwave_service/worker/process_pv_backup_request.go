package worker

import (
	"context"
	"github.com/google/uuid"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"os"
	"path/filepath"
)

func (m Manager) PersistentVolumeBackup(request PersistentVolumeBackupRequest, ctx context.Context, cancelContext context.CancelFunc) error {
	dbWithoutTx := m.ServiceManager.DbClient
	// fetch persistent volume backup
	var persistentVolumeBackup core.PersistentVolumeBackup
	err := persistentVolumeBackup.FindById(ctx, dbWithoutTx, request.Id)
	if err != nil {
		return nil
	}
	// check status should be pending
	if persistentVolumeBackup.Status != core.BackupPending {
		return nil
	}
	// fetch persistent volume
	var persistentVolume core.PersistentVolume
	err = persistentVolume.FindById(ctx, dbWithoutTx, persistentVolumeBackup.PersistentVolumeID)
	if err != nil {
		return nil
	}
	dockerManager := m.ServiceManager.DockerManager
	// generate a random filename
	backupFileName := "backup_" + persistentVolume.Name + "_" + uuid.NewString()
	backupFilePath := filepath.Join(m.SystemConfig.ServiceConfig.DataDir, backupFileName)
	// create backup
	err = dockerManager.BackupVolume(persistentVolume.Name, backupFilePath)
	if err != nil {
		return err
	}
	// update status
	persistentVolumeBackup.Status = core.BackupSuccess
	persistentVolumeBackup.File = backupFilePath
	size, err := sizeOfFileInMB(backupFilePath)
	if err != nil {
		return err
	}
	persistentVolumeBackup.FileSizeMB = size
	err = persistentVolumeBackup.Update(ctx, dbWithoutTx)
	if err != nil {
		return err
	}
	return nil
}

func sizeOfFileInMB(path string) (float64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	size := float64(fileInfo.Size()) / (1024 * 1024)
	return size, nil
}
