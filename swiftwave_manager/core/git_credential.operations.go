package core

import (
	"context"
	"gorm.io/gorm"
)

// This file contains the operations for the GitCredential model.
// This functions will perform necessary validation before doing the actual database operation.

// Each function's argument format should be (ctx context.Context, db gorm.DB, ...)
// context used to pass some data to the function e.g. user id, auth info, etc.

func FindAllGitCredentials(ctx context.Context, db gorm.DB) ([]*GitCredential, error) {
	var gitCredentials []*GitCredential
	tx := db.Find(&gitCredentials)
	return gitCredentials, tx.Error
}

func (gitCredential *GitCredential) FindById(ctx context.Context, db gorm.DB, id int) error {
	tx := db.First(&gitCredential, id)
	return tx.Error
}

func (gitCredential *GitCredential) Create(ctx context.Context, db gorm.DB) error {
	tx := db.Create(&gitCredential)
	return tx.Error
}

func (gitCredential *GitCredential) Update(ctx context.Context, db gorm.DB) error {
	tx := db.Save(&gitCredential)
	return tx.Error
}

func (gitCredential *GitCredential) Delete(ctx context.Context, db gorm.DB) error {
	tx := db.Delete(&gitCredential)
	return tx.Error
}
