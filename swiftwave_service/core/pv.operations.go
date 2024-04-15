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

func IsExistPersistentVolume(_ context.Context, db gorm.DB, name string, dockerManager containermanger.Manager) (bool, error) {
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

func FindAllPersistentVolumes(_ context.Context, db gorm.DB) ([]*PersistentVolume, error) {
	var persistentVolumes []*PersistentVolume
	tx := db.Find(&persistentVolumes)
	return persistentVolumes, tx.Error
}

func (persistentVolume *PersistentVolume) FindById(_ context.Context, db gorm.DB, id uint) error {
	tx := db.Where("id = ?", id).First(&persistentVolume)
	return tx.Error
}

func (persistentVolume *PersistentVolume) FindByName(_ context.Context, db gorm.DB, name string) error {
	tx := db.Where("name = ?", name).First(&persistentVolume)
	return tx.Error
}

func (persistentVolume *PersistentVolume) Create(_ context.Context, db gorm.DB, dockerManager containermanger.Manager) error {
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
	} else if persistentVolume.Type == PersistentVolumeTypeCIFS {
		err = dockerManager.CreateCIFSVolume(persistentVolume.Name, persistentVolume.CIFSConfig.Host, persistentVolume.CIFSConfig.Share, persistentVolume.CIFSConfig.Username, persistentVolume.CIFSConfig.Password, persistentVolume.CIFSConfig.FileMode, persistentVolume.CIFSConfig.DirMode)
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

func (persistentVolume *PersistentVolume) Update(_ context.Context, _ gorm.DB, _ containermanger.Manager) error {
	return errors.New("persistentVolume update is not allowed")
}

func (persistentVolume *PersistentVolume) ValidateDeletion(_ context.Context, db gorm.DB) error {
	// Verify there is no existing PersistentVolumeBinding with this PersistentVolume
	var count int64
	db.Model(&PersistentVolumeBinding{}).Where("persistent_volume_id = ?", persistentVolume.ID).Count(&count)
	if count > 0 {
		return errors.New("there are some applications using this volume, delete them to delete this volume")
	}
	return nil
}

func (persistentVolume *PersistentVolume) Delete(_ context.Context, db gorm.DB) error {
	// Delete persistentVolume from database
	tx := db.Delete(&persistentVolume)
	return tx.Error
}

func isValidVolumeName(volumeName string) bool {
	regex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return regex.MatchString(volumeName)
}
