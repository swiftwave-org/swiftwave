package core

import (
	"context"
	"errors"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	"gorm.io/gorm"
	"regexp"
)

// This file contains the operations for the PersistentVolume model.
// This functions will perform necessary validation before doing the actual database operation.

// Each function's argument format should be (ctx context.Context, db gorm.DB, ...)
// context used to pass some data to the function e.g. user id, auth info, etc.

func IsExistPersistentVolume(ctx context.Context, db gorm.DB, name string, dockerManager containermanger.Manager) (bool, error) {
	// verify from database
	var count int64
	tx := db.Model(&PersistentVolume{}).Where("name = ?", name).Count(&count)
	if tx.Error != nil {
		return false, tx.Error
	}
	if count > 0 {
		return true, nil
	}
	// verify from docker client
	isExists := dockerManager.ExistsVolume(name)
	if isExists {
		return true, nil
	}
	return false, nil
}

func FindAllPersistentVolumes(ctx context.Context, db gorm.DB) ([]*PersistentVolume, error) {
	var persistentVolumes []*PersistentVolume
	tx := db.Find(&persistentVolumes)
	return persistentVolumes, tx.Error
}

func (persistentVolume *PersistentVolume) FindById(ctx context.Context, db gorm.DB, id uint) error {
	tx := db.Where("id = ?", id).First(&persistentVolume)
	return tx.Error
}

func (persistentVolume *PersistentVolume) FindByName(ctx context.Context, db gorm.DB, name string) error {
	tx := db.Where("name = ?", name).First(&persistentVolume)
	return tx.Error
}

func (persistentVolume *PersistentVolume) Create(ctx context.Context, db gorm.DB, dockerManager containermanger.Manager) error {
	// verify name is valid
	if !isValidVolumeName(persistentVolume.Name) {
		return errors.New("name can only contain alphabets, numbers and underscore")
	}
	// verify there is no existing persistentVolume with same name
	// verify from database
	var count int64
	db.Model(&PersistentVolume{}).Where("name = ?", persistentVolume.Name).Count(&count)
	if count > 0 {
		return errors.New("persistentVolume with same name already exists")
	}
	// verify from docker client
	isExists := dockerManager.ExistsVolume(persistentVolume.Name)
	if isExists {
		return errors.New("persistentVolume with same name already exists")
	}
	// Start a database transaction
	transaction := db.Begin()
	// Create persistentVolume in database
	tx := transaction.Create(&persistentVolume)
	if tx.Error != nil {
		transaction.Rollback()
		return tx.Error
	}
	var err error
	// Create persistentVolume in docker
	if persistentVolume.Type == PersistentVolumeTypeLocal {
		err = dockerManager.CreateLocalVolume(persistentVolume.Name)
	} else if persistentVolume.Type == PersistentVolumeTypeNFS {
		err = dockerManager.CreateNFSVolume(persistentVolume.Name, persistentVolume.NFSConfig.Host, persistentVolume.NFSConfig.Path, persistentVolume.NFSConfig.Version)
	} else {
		transaction.Rollback()
		return errors.New("invalid persistentVolume type")
	}
	if err != nil {
		transaction.Rollback()
		return err
	}
	return transaction.Commit().Error
}

func (persistentVolume *PersistentVolume) Update(ctx context.Context, db gorm.DB, dockerManager containermanger.Manager) error {
	return errors.New("persistentVolume update is not allowed")
}

func (persistentVolume *PersistentVolume) Delete(ctx context.Context, db gorm.DB, dockerManager containermanger.Manager) error {
	// Verify there is no existing PersistentVolumeBinding with this PersistentVolume
	var count int64
	db.Model(&PersistentVolumeBinding{}).Where("persistent_volume_id = ?", persistentVolume.ID).Count(&count)
	if count > 0 {
		return errors.New("there are some applications using this volume, delete them to delete this volume")
	}
	// Start a database transaction
	transaction := db.Begin()
	// check if there is any backup of this persistentVolume
	var backupCount int64
	transaction.Model(&PersistentVolumeBackup{}).Where("persistent_volume_id = ?", persistentVolume.ID).Count(&backupCount)
	if backupCount > 0 {
		transaction.Rollback()
		return errors.New("there are some backups of this volume, delete them first to delete this volume")
	}
	var restoreCount int64
	transaction.Model(&PersistentVolumeRestore{}).Where("persistent_volume_id = ?", persistentVolume.ID).Count(&restoreCount)
	if restoreCount > 0 {
		transaction.Rollback()
		return errors.New("there are some restore histories of this volume, delete them first to delete this volume")
	}
	// Delete persistentVolume from database
	tx := transaction.Delete(&persistentVolume)
	if tx.Error != nil {
		transaction.Rollback()
		return tx.Error
	}
	// Delete persistentVolume from docker
	err := dockerManager.RemoveVolume(persistentVolume.Name)
	if err != nil {
		transaction.Rollback()
		return err
	}
	return transaction.Commit().Error
}

func isValidVolumeName(volumeName string) bool {
	regex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return regex.MatchString(volumeName)
}
