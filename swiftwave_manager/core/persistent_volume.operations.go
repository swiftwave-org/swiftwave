package core

import (
	"context"
	"errors"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	"gorm.io/gorm"
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

func (persistentVolume *PersistentVolume) FindById(ctx context.Context, db gorm.DB, id int) error {
	tx := db.First(&persistentVolume, id)
	return tx.Error
}

func (persistentVolume *PersistentVolume) Create(ctx context.Context, db gorm.DB, dockerManager containermanger.Manager) error {
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
	// Create persistentVolume in docker
	err := dockerManager.CreateVolume(persistentVolume.Name)
	if err != nil {
		transaction.Rollback()
		return err
	}
	return transaction.Commit().Error
}

func (persistentVolume *PersistentVolume) Update(ctx context.Context, db gorm.DB) error {
	return errors.New("persistentVolume update is not allowed")
}

func (persistentVolume *PersistentVolume) Delete(ctx context.Context, db gorm.DB, dockerManager containermanger.Manager) error {
	// Verify there is no existing PersistentVolumeBinding with this PersistentVolume
	var count int64
	db.Model(&PersistentVolumeBinding{}).Where("persistentVolumeID = ?", persistentVolume.ID).Count(&count)
	if count > 0 {
		return errors.New("there are some applications using this volume, delete them to delete this volume")
	}
	// TODO: verify there is no container in runtime outside scope of swiftwave using this volume
	// Start a database transaction
	transaction := db.Begin()
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
