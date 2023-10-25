package core

import (
	"context"
	"github.com/google/uuid"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	"gorm.io/gorm"
	"time"
)

// This file contains the operations for the Deployment model.
// This functions will perform necessary validation before doing the actual database operation.

// Each function's argument format should be (ctx context.Context, db gorm.DB, ...)
// context used to pass some data to the function e.g. user id, auth info, etc.

func FindLatestDeploymentByApplicationId(ctx context.Context, db gorm.DB, id string) (*Deployment, error) {
	var deployment = &Deployment{}
	tx := db.Where("application_id = ?", id).Order("created_at desc").First(&deployment)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return deployment, nil
}

func (deployment *Deployment) Create(ctx context.Context, db gorm.DB, dockerManager containermanger.Manager) error {
	deployment.ID = uuid.NewString()
	deployment.CreatedAt = time.Now()
	deployment.Status = DeploymentStatusPending
	tx := db.Create(&deployment)
	return tx.Error
}

func (deployment *Deployment) Delete(ctx context.Context, db gorm.DB, dockerManager containermanger.Manager) error {
	// delete all build args
	tx := db.Where("deployment_id = ?", deployment.ID).Delete(&BuildArg{})
	if tx.Error != nil {
		return tx.Error
	}
	// delete all logs
	tx = db.Where("deployment_id = ?", deployment.ID).Delete(&DeploymentLog{})
	if tx.Error != nil {
		return tx.Error
	}
	// delete deployment
	tx = db.Delete(&deployment)
	return tx.Error
}
