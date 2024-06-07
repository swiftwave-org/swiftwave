package core

import (
	"context"
	"gorm.io/gorm"
)

// This file contains the operations for the ConfigMount model.
// This functions will perform necessary validation before doing the actual database operation.

// Each function's argument format should be (ctx context.Context, db gorm.DB, ...)
// context used to pass some data to the function e.g. user id, auth info, etc.

func FindConfigMountsByApplicationId(_ context.Context, db gorm.DB, applicationId string) ([]*ConfigMount, error) {
	var configMounts []*ConfigMount
	tx := db.Where("application_id = ?", applicationId).Find(&configMounts)
	return configMounts, tx.Error
}

func DeleteConfigMountsByApplicationId(_ context.Context, db gorm.DB, applicationId string) error {
	tx := db.Delete(&ConfigMount{}, "application_id = ?", applicationId)
	return tx.Error
}

func (c *ConfigMount) Create(_ context.Context, db gorm.DB) error {
	tx := db.Create(c)
	return tx.Error
}

func (c *ConfigMount) Update(_ context.Context, db gorm.DB) error {
	// make config id empty to ensure recreation of config in swarm
	c.ConfigID = ""
	tx := db.Save(c)
	return tx.Error
}

func (c *ConfigMount) Delete(_ context.Context, db gorm.DB) error {
	tx := db.Delete(c)
	return tx.Error
}
