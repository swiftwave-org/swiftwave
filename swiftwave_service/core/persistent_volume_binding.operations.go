package core

import (
	"context"

	"gorm.io/gorm"
)

// This file contains the operations for the PersistentVolumeBinding model.
// This functions will perform necessary validation before doing the actual database operation.

// Each function's argument format should be (ctx context.Context, db gorm.DB, ...)
// context used to pass some data to the function e.g. user id, auth info, etc.

func FindPersistentVolumeBindingsByApplicationId(ctx context.Context, db gorm.DB, applicationId string) ([]*PersistentVolumeBinding, error) {
	var persistentVolumeBindings []*PersistentVolumeBinding
	tx := db.Where("application_id = ?", applicationId).Find(&persistentVolumeBindings)
	return persistentVolumeBindings, tx.Error
}

func FindPersistentVolumeBindingsByPersistentVolumeId(ctx context.Context, db gorm.DB, persistentVolumeId uint) ([]*PersistentVolumeBinding, error) {
	var persistentVolumeBindings []*PersistentVolumeBinding
	tx := db.Where("persistent_volume_id = ?", persistentVolumeId).Find(&persistentVolumeBindings)
	return persistentVolumeBindings, tx.Error
}

func (p *PersistentVolumeBinding) Create(ctx context.Context, db gorm.DB) error {
	tx := db.Create(p)
	return tx.Error
}

func (p *PersistentVolumeBinding) Update(ctx context.Context, db gorm.DB) error {
	tx := db.Save(p)
	return tx.Error
}

func (p *PersistentVolumeBinding) Delete(ctx context.Context, db gorm.DB) error {
	tx := db.Delete(p)
	return tx.Error
}

func DeletePersistentVolumeBindingsByApplicationId(ctx context.Context, db gorm.DB, applicationId string) error {
	tx := db.Delete(&PersistentVolumeBinding{}, "application_id = ?", applicationId)
	return tx.Error
}
