package worker

import (
	"context"
	"github.com/google/uuid"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/uploader"
	"gorm.io/gorm"
	"log"
	"os"
	"path/filepath"
)

func (m Manager) PersistentVolumeBackup(request PersistentVolumeBackupRequest, ctx context.Context, _ context.CancelFunc) error {
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
	// fetch swarm server
	server, err := core.FetchSwarmManager(&dbWithoutTx)
	if err != nil {
		return err
	}
	dockerManager, err := manager.DockerClient(ctx, server)
	if err != nil {
		return err
	}
	// generate a random filename
	backupFileName := "backup_" + persistentVolume.Name + "_" + uuid.NewString() + ".tar.gz"
	var backupFilePath string
	if persistentVolumeBackup.Type == core.LocalBackup {
		backupFilePath = filepath.Join(m.Config.LocalConfig.ServiceConfig.PVBackupDirectoryPath, backupFileName)
	} else if persistentVolumeBackup.Type == core.S3Backup {
		// fetch tmp dir
		tmpDir := os.TempDir()
		backupFilePath = filepath.Join(tmpDir, backupFileName)
		defer func() {
			err := os.Remove(backupFilePath)
			if err != nil {
				log.Println("failed to remove backup file " + err.Error())
			}
		}()
	} else {
		return nil
	}
	// create backup
	err = dockerManager.BackupVolume(persistentVolume.Name, backupFilePath)
	if err != nil {
		markPVBackupRequestAsFailed(dbWithoutTx, persistentVolumeBackup)
		return nil
	}
	// TODO: copy it from swarm to management node
	// upload to s3
	if persistentVolumeBackup.Type == core.S3Backup {
		backupFileReader, err := os.Open(backupFilePath) // TODO
		if err != nil {
			markPVBackupRequestAsFailed(dbWithoutTx, persistentVolumeBackup)
			return nil
		}
		defer func() {
			err := backupFileReader.Close()
			if err != nil {
				log.Println("failed to close backup file reader " + err.Error())
			}
		}()
		s3Config := m.Config.SystemConfig.PersistentVolumeBackupConfig.S3BackupConfig
		err = uploader.UploadFileToS3(backupFileReader, backupFileName, s3Config.Bucket, s3Config)
		if err != nil {
			log.Println("error while uploading backup to s3 > " + err.Error())
			markPVBackupRequestAsFailed(dbWithoutTx, persistentVolumeBackup)
			return nil
		}
	}
	// update status
	persistentVolumeBackup.Status = core.BackupSuccess
	persistentVolumeBackup.File = backupFileName
	size, err := sizeOfFileInMB(backupFilePath)
	if err != nil {
		markPVBackupRequestAsFailed(dbWithoutTx, persistentVolumeBackup)
		return nil
	}
	persistentVolumeBackup.FileSizeMB = size
	err = persistentVolumeBackup.Update(ctx, dbWithoutTx)
	if err != nil {
		return err
	}
	return nil
}

func markPVBackupRequestAsFailed(db gorm.DB, persistentVolumeBackup core.PersistentVolumeBackup) {
	persistentVolumeBackup.Status = core.BackupFailed
	err := persistentVolumeBackup.Update(context.Background(), db)
	if err != nil {
		log.Println("error while updating persistent volume backup status to failed")
	}
}

func sizeOfFileInMB(path string) (float64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	size := float64(fileInfo.Size()) / (1024 * 1024)
	return size, nil
}
