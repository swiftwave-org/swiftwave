package core

import (
	"context"
	"gorm.io/gorm"
)

// This file contains the operations for the EnvironmentVariable model.
// This functions will perform necessary validation before doing the actual database operation.

// Each function's argument format should be (ctx context.Context, db gorm.DB, ...)
// context used to pass some data to the function e.g. user id, auth info, etc.

func FindEnvironmentVariablesByApplicationId(ctx context.Context, db gorm.DB, applicationId string) ([]*EnvironmentVariable, error) {
	var environmentVariables []*EnvironmentVariable
	tx := db.Where("application_id = ?", applicationId).Find(&environmentVariables)
	return environmentVariables, tx.Error
}
