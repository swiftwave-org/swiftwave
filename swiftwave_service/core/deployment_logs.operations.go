package core

import (
	"context"
	"gorm.io/gorm"
)

// This file contains the operations for the DeploymentLog model.
// This functions will perform necessary validation before doing the actual database operation.

// Each function's argument format should be (ctx context.Context, db gorm.DB, ...)
// context used to pass some data to the function e.g. user id, auth info, etc.

func FindAllDeploymentLogsByDeploymentId(ctx context.Context, db gorm.DB, id string) ([]DeploymentLog, error) {
	// fetch all deployment logs
	var deploymentLogs = make([]DeploymentLog, 0)
	err := db.Where("deployment_id = ?", id).Order("created_at desc").Find(&deploymentLogs).Error
	if err != nil {
		return nil, err
	}
	return deploymentLogs, nil
}

func DeleteBuildLogsByDeploymentId(ctx context.Context, db gorm.DB, deploymentId string) error {
	tx := db.Where("deployment_id = ?", deploymentId).Delete(&DeploymentLog{})
	return tx.Error
}
