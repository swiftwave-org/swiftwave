package worker

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/swiftwave-org/swiftwave/ssh_toolkit"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/uploader"
	"gorm.io/gorm"
	"io"
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
	// fetch tmp dir
	tmpDir := os.TempDir()
	backupFilePath = filepath.Join(tmpDir, backupFileName)
	defer func() {
		// delete original backup from current node
		_ = ssh_toolkit.ExecCommandOverSSH(fmt.Sprintf("rm -f %s", backupFilePath), nil, nil, 20, server.IP, 22, server.User, m.Config.SystemConfig.SshPrivateKey, 30)
	}()
	// create backup
	err = dockerManager.BackupVolume(persistentVolume.Name, backupFilePath)
	if err != nil {
		logger.CronJobLoggerError.Println("error while creating backup > " + err.Error())
		markPVBackupRequestAsFailed(dbWithoutTx, persistentVolumeBackup)
		return nil
	}
	// copy to management node
	err = ssh_toolkit.CopyFileFromRemoteServer(backupFilePath, backupFilePath, server.IP, 22, server.User, m.Config.SystemConfig.SshPrivateKey)
	if err != nil {
		logger.CronJobLoggerError.Println("error while copying backup file to management node > " + err.Error())
		markPVBackupRequestAsFailed(dbWithoutTx, persistentVolumeBackup)
		return nil
	}
	if persistentVolumeBackup.Type == core.LocalBackup {
		localBackupPath := filepath.Join(m.Config.LocalConfig.ServiceConfig.PVBackupDirectoryPath, backupFileName)
		// open backup file
		backupFileReader, err := os.Open(backupFilePath)
		if err != nil {
			logger.CronJobLoggerError.Println("error while opening backup file > " + err.Error())
			markPVBackupRequestAsFailed(dbWithoutTx, persistentVolumeBackup)
			return nil
		}
		defer func() {
			_ = backupFileReader.Close()
		}()
		// create backup file
		backupFile, err := os.Create(localBackupPath)
		if err != nil {
			logger.CronJobLoggerError.Println("error while creating backup file > " + err.Error())
			markPVBackupRequestAsFailed(dbWithoutTx, persistentVolumeBackup)
			return nil
		}
		defer func() {
			_ = backupFile.Close()
		}()
		// copy backup file
		_, err = io.Copy(backupFile, backupFileReader)
		if err != nil {
			logger.CronJobLoggerError.Println("error while copying backup file > " + err.Error())
			markPVBackupRequestAsFailed(dbWithoutTx, persistentVolumeBackup)
			return nil
		}
	} else if persistentVolumeBackup.Type == core.S3Backup {
		// upload to s3
		backupFileReader, err := os.Open(backupFilePath)
		if err != nil {
			logger.CronJobLoggerError.Println("error while opening backup file > " + err.Error())
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
		logger.CronJobLoggerError.Println("error while getting backup file size > " + err.Error())
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
