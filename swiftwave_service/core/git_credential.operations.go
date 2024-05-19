package core

import (
	"context"
	"gorm.io/gorm"
	"strings"
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

func (gitCredential *GitCredential) FindById(ctx context.Context, db gorm.DB, id uint) error {
	tx := db.Where("id = ?", id).First(&gitCredential)
	return tx.Error
}

func (gitCredential *GitCredential) Create(ctx context.Context, db gorm.DB) error {
	tx := db.Create(&gitCredential)
	return tx.Error
}

func (gitCredential *GitCredential) Update(ctx context.Context, db gorm.DB) error {
	// fetch old record
	var oldGitCredential = &GitCredential{}
	err := oldGitCredential.FindById(ctx, db, gitCredential.ID)
	if err != nil {
		return err
	}
	if gitCredential.Type == GitSsh && strings.Compare(strings.TrimSpace(gitCredential.SshPrivateKey), "") == 0 {
		gitCredential.SshPrivateKey = oldGitCredential.SshPrivateKey
		gitCredential.SshPublicKey = oldGitCredential.SshPublicKey
	} else if gitCredential.Type == GitHttp && strings.Compare(strings.TrimSpace(gitCredential.Password), "") == 0 {
		gitCredential.Password = oldGitCredential.Password
	}
	tx := db.Save(&gitCredential)
	return tx.Error
}

func (gitCredential *GitCredential) Delete(ctx context.Context, db gorm.DB) error {
	// set gitCredentialID null for all deployment using this gitCredential
	err := db.Model(&Deployment{}).Where("git_credential_id = ?", gitCredential.ID).Update("git_credential_id", nil).Error
	if err != nil {
		return err
	}
	tx := db.Delete(&gitCredential)
	return tx.Error
}
