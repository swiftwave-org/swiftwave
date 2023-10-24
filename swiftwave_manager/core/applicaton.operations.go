package core

import (
	"context"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	"gorm.io/gorm"
)

// This file contains the operations for the Application model.
// This functions will perform necessary validation before doing the actual database operation.

// Each function's argument format should be (ctx context.Context, db gorm.DB, ...)
// context used to pass some data to the function e.g. user id, auth info, etc.

func IsExistApplicationName(ctx context.Context, db gorm.DB, dockerManager containermanger.Manager, name string) (bool, error) {
	// verify from database
	var count int64
	tx := db.Model(&Application{}).Where("name = ?", name).Count(&count)
	if tx.Error != nil {
		return false, tx.Error
	}
	if count > 0 {
		return true, nil
	}
	// verify from docker client
	_, err := dockerManager.GetService(name)
	if err == nil {
		return true, nil
	}
	return false, nil
}

func FindAllApplications(ctx context.Context, db gorm.DB) ([]*Application, error) {
	var applications []*Application
	tx := db.Find(&applications)
	return applications, tx.Error
}

func (application *Application) FindById(ctx context.Context, db gorm.DB, id string) error {
	tx := db.First(&application, id)
	return tx.Error
}

func (application *Application) Create(ctx context.Context, db gorm.DB, dockerManager containermanger.Manager) error {
	// TODO: add validation, create new deployment
	tx := db.Create(&application)
	return tx.Error
}

func (application *Application) Update(ctx context.Context, db gorm.DB, dockerManager containermanger.Manager) error {
	// TODO: add validation, create new deployment if change required
	tx := db.Save(&application)
	return tx.Error
}

func (application *Application) Delete(ctx context.Context, db gorm.DB, dockerManager containermanger.Manager) error {
	// TODO: add validation, delete all deployments and application
	tx := db.Delete(&application)
	return tx.Error
}
