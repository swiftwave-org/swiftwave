package core

import (
	"context"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/system_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/uploader"
	"gorm.io/gorm"
	"log"
	"os"
	"path/filepath"
	"time"
)

// This file contains the operations for the PersistentVolumeBackup model.
// This functions will perform necessary validation before doing the actual database operation.

// Each function's argument format should be (ctx context.Context, db gorm.DB, ...)
// context used to pass some data to the function e.g. user id, auth info, etc.

func (persistentVolumeBackup *PersistentVolumeBackup) FindById(ctx context.Context, db gorm.DB, id uint) error {
	tx := db.Where("id = ?", id).First(&persistentVolumeBackup)
	return tx.Error
}

func (persistentVolumeBackup *PersistentVolumeBackup) Create(ctx context.Context, db gorm.DB) error {
	persistentVolumeBackup.ID = 0
	persistentVolumeBackup.Status = BackupPending
	persistentVolumeBackup.CreatedAt = time.Now()
	persistentVolumeBackup.CompletedAt = time.Now()
	tx := db.Create(&persistentVolumeBackup)
	return tx.Error
}

func (persistentVolumeBackup *PersistentVolumeBackup) Update(ctx context.Context, db gorm.DB) error {
	persistentVolumeBackup.CompletedAt = time.Now()
	tx := db.Save(persistentVolumeBackup)
	return tx.Error
}

func (persistentVolumeBackup *PersistentVolumeBackup) Delete(ctx context.Context, db gorm.DB, dataDir string, config system_config.S3BackupConfig) error {
	if persistentVolumeBackup.File != "" {
		if persistentVolumeBackup.Type == LocalBackup {
			err := os.Remove(filepath.Join(dataDir, persistentVolumeBackup.File))
			if err != nil {
				log.Println("error deleting file: ", err)
			}
		}
		if persistentVolumeBackup.Type == S3Backup {
			err := uploader.DeleteFileFromS3(persistentVolumeBackup.File, config.Bucket, config)
			if err != nil {
				log.Println("error deleting file from s3: ", err)
			}
		}
	}
	tx := db.Delete(persistentVolumeBackup)
	return tx.Error
}

func FindPersistentVolumeBackupsByPersistentVolumeId(ctx context.Context, db gorm.DB, persistentVolumeId uint) ([]*PersistentVolumeBackup, error) {
	var persistentVolumeBackups []*PersistentVolumeBackup
	tx := db.Where("persistent_volume_id = ?", persistentVolumeId).Order("id desc").Find(&persistentVolumeBackups)
	return persistentVolumeBackups, tx.Error
}

func DeletePersistentVolumeBackupsByPersistentVolumeId(ctx context.Context, db gorm.DB, persistentVolumeId uint, dataDir string, config system_config.S3BackupConfig) error {
	transaction := db.Begin()
	var persistentVolumeBackups []*PersistentVolumeBackup
	tx := transaction.Where("persistent_volume_id = ?", persistentVolumeId).Find(&persistentVolumeBackups)
	if tx.Error != nil {
		transaction.Rollback()
		return tx.Error
	}
	for _, p := range persistentVolumeBackups {
		err := p.Delete(ctx, *transaction, dataDir, config)
		if err != nil {
			log.Println("error deleting persistentVolumeBackup: ", err)
		}
	}
	tx = db.Delete(&PersistentVolumeBackup{}, "persistent_volume_id = ?", persistentVolumeId)
	if tx.Error != nil {
		transaction.Rollback()
		return tx.Error
	}
	return transaction.Commit().Error
}
