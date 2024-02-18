package core

import (
	"context"
	"gorm.io/gorm"
	"log"
	"os"
	"path/filepath"
	"time"
)

// This file contains the operations for the PersistentVolumeRestore model.
// This functions will perform necessary validation before doing the actual database operation.

// Each function's argument format should be (ctx context.Context, db gorm.DB, ...)
// context used to pass some data to the function e.g. user id, auth info, etc.

func (persistentVolumeRestore *PersistentVolumeRestore) FindById(ctx context.Context, db gorm.DB, id uint) error {
	tx := db.Where("id = ?", id).First(&persistentVolumeRestore)
	return tx.Error
}

func (persistentVolumeRestore *PersistentVolumeRestore) Create(ctx context.Context, db gorm.DB) error {
	persistentVolumeRestore.ID = 0
	persistentVolumeRestore.Status = RestorePending
	persistentVolumeRestore.CreatedAt = time.Now()
	persistentVolumeRestore.CompletedAt = time.Now()
	tx := db.Create(&persistentVolumeRestore)
	return tx.Error
}

func (persistentVolumeRestore *PersistentVolumeRestore) Update(ctx context.Context, db gorm.DB, dataDir string) error {
	persistentVolumeRestore.CompletedAt = time.Now()
	tx := db.Save(persistentVolumeRestore)
	err := tx.Error
	if err != nil {
		return err
	}
	// delete the file
	if persistentVolumeRestore.File != "" {
		err = os.Remove(filepath.Join(dataDir, persistentVolumeRestore.File))
		if err != nil {
			log.Println("error deleting restore file: ", err)
		}
	}
	return nil
}

func (persistentVolumeRestore *PersistentVolumeRestore) Delete(ctx context.Context, db gorm.DB, dataDir string) error {
	if persistentVolumeRestore.File != "" {
		err := os.Remove(filepath.Join(dataDir, persistentVolumeRestore.File))
		if err != nil {
			log.Println("error deleting file: ", err)
		}
	}
	tx := db.Delete(persistentVolumeRestore)
	return tx.Error
}

func FindPersistentVolumeRestoresByPersistentVolumeId(ctx context.Context, db gorm.DB, persistentVolumeId uint) ([]*PersistentVolumeRestore, error) {
	var persistentVolumeRestores []*PersistentVolumeRestore
	tx := db.Where("persistent_volume_id = ?", persistentVolumeId).Order("id desc").Find(&persistentVolumeRestores)
	return persistentVolumeRestores, tx.Error
}

func DeletePersistentVolumeRestoresByPersistentVolumeId(ctx context.Context, db gorm.DB, persistentVolumeId uint, dataDir string) error {
	transaction := db.Begin()
	var persistentVolumeRestores []*PersistentVolumeRestore
	tx := transaction.Where("persistent_volume_id = ?", persistentVolumeId).Find(&persistentVolumeRestores)
	if tx.Error != nil {
		transaction.Rollback()
		return tx.Error
	}
	for _, persistentVolumeRestore := range persistentVolumeRestores {
		err := persistentVolumeRestore.Delete(ctx, *transaction, dataDir)
		if err != nil {
			transaction.Rollback()
			return err
		}
	}
	tx = transaction.Commit()
	return tx.Error
}
