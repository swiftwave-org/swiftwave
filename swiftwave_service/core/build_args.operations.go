package core

import (
	"context"
	"gorm.io/gorm"
)

// This file contains the operations for the BuildArg model.
// This functions will perform necessary validation before doing the actual database operation.

// Each function's argument format should be (ctx context.Context, db gorm.DB, ...)
// context used to pass some data to the function e.g. user id, auth info, etc.

func FindBuildArgsByDeploymentId(ctx context.Context, db gorm.DB, deploymentId string) ([]*BuildArg, error) {
	var buildArgs []*BuildArg
	tx := db.Where("deployment_id = ?", deploymentId).Find(&buildArgs)
	return buildArgs, tx.Error
}

func DeleteBuildArgsByDeploymentId(ctx context.Context, db gorm.DB, deploymentId string) error {
	tx := db.Where("deployment_id = ?", deploymentId).Delete(&BuildArg{})
	return tx.Error
}
