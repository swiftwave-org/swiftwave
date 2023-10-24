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

func (deployment *Deployment) Create(ctx context.Context, db gorm.DB, dockerManager containermanger.Manager) error {
	deployment.ID = uuid.NewString()
	deployment.CreatedAt = time.Now()
	deployment.Status = DeploymentStatusPending
	tx := db.Create(&deployment)
	return tx.Error
}

func (deployment *Deployment) Delete(ctx context.Context, db gorm.DB, dockerManager containermanger.Manager) error {
	tx := db.Delete(&deployment)
	// TODO: delete build args, logs and other related data
	return tx.Error
}
