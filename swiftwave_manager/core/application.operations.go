package core

import (
	"context"
	"errors"
	"github.com/google/uuid"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_manager/graphql/model"
	"gorm.io/gorm"
	"os"
	"path/filepath"
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
	tx := db.Where("is_deleted = ?", false).Find(&applications)
	return applications, tx.Error
}

func (application *Application) FindById(ctx context.Context, db gorm.DB, id string) error {
	tx := db.First(&application, id)
	if tx.Error != nil {
		return tx.Error
	}
	// check if it's deleted
	if application.IsDeleted {
		return errors.New("application is deleted")
	}
	return nil
}

func (application *Application) Create(ctx context.Context, db gorm.DB, dockerManager containermanger.Manager, codeTarballDir string) error {
	// verify if there is no application with same name
	isExist, err := IsExistApplicationName(ctx, db, dockerManager, application.Name)
	if err != nil {
		return err
	}
	if isExist {
		return errors.New("application name not available")
	}
	// For UpstreamType = Git, verify git record id
	if application.LatestDeployment.UpstreamType == UpstreamTypeGit {
		var gitCredential GitCredential
		err := gitCredential.FindById(ctx, db, int(application.LatestDeployment.GitCredentialID))
		if err != nil {
			return err
		}
	}
	// For UpstreamType = Image, verify image registry credential id
	if application.LatestDeployment.UpstreamType == UpstreamTypeImage {
		if application.LatestDeployment.ImageRegistryCredentialID != 0 {
			var imageRegistryCredential ImageRegistryCredential
			err := imageRegistryCredential.FindById(ctx, db, int(application.LatestDeployment.ImageRegistryCredentialID))
			if err != nil {
				return err
			}
		}
	}
	// For UpstreamType = SourceCode, verify source code compressed file exists
	if application.LatestDeployment.UpstreamType == UpstreamTypeSourceCode {
		tarballPath := filepath.Join(codeTarballDir, application.LatestDeployment.SourceCodeCompressedFileName)
		// Verify file exists
		if _, err := os.Stat(tarballPath); os.IsNotExist(err) {
			return errors.New("source code not found")
		}
	}
	// create application
	createdApplication := model.Application{
		ID:             uuid.NewString(),
		Name:           application.Name,
		DeploymentMode: string(application.DeploymentMode),
		Replicas:       int(application.Replicas),
	}
	tx := db.Create(&createdApplication)
	if tx.Error != nil {
		return tx.Error
	}
	// create environment variables
	createdEnvironmentVariables := make([]EnvironmentVariable, 0)
	for _, environmentVariable := range application.EnvironmentVariables {
		createdEnvironmentVariable := EnvironmentVariable{
			ApplicationID: createdApplication.ID,
			Key:           environmentVariable.Key,
			Value:         environmentVariable.Value,
		}
		createdEnvironmentVariables = append(createdEnvironmentVariables, createdEnvironmentVariable)
	}
	tx = db.Create(&createdEnvironmentVariables)
	if tx.Error != nil {
		return tx.Error
	}
	// create persistent volume bindings
	createdPersistentVolumeBindings := make([]PersistentVolumeBinding, 0)
	for _, persistentVolumeBinding := range application.PersistentVolumeBindings {
		createdPersistentVolumeBinding := PersistentVolumeBinding{
			ApplicationID:      createdApplication.ID,
			PersistentVolumeID: persistentVolumeBinding.PersistentVolumeID,
			MountingPath:       persistentVolumeBinding.MountingPath,
		}
		createdPersistentVolumeBindings = append(createdPersistentVolumeBindings, createdPersistentVolumeBinding)
	}
	tx = db.Create(&createdPersistentVolumeBindings)
	if tx.Error != nil {
		return tx.Error
	}
	// create deployment
	createdDeployment := Deployment{
		ApplicationID: createdApplication.ID,
		UpstreamType:  application.LatestDeployment.UpstreamType,
		// Fields for UpstreamType = Git
		GitCredentialID:  application.LatestDeployment.GitCredentialID,
		GitProvider:      application.LatestDeployment.GitProvider,
		RepositoryOwner:  application.LatestDeployment.RepositoryOwner,
		RepositoryName:   application.LatestDeployment.RepositoryName,
		RepositoryBranch: application.LatestDeployment.RepositoryBranch,
		CommitHash:       application.LatestDeployment.CommitHash,
		// Fields for UpstreamType = SourceCode
		SourceCodeCompressedFileName: application.LatestDeployment.SourceCodeCompressedFileName,
		// Fields for UpstreamType = Image
		DockerImage:               application.LatestDeployment.DockerImage,
		ImageRegistryCredentialID: application.LatestDeployment.ImageRegistryCredentialID,
		// other fields
		Dockerfile: application.LatestDeployment.Dockerfile,
	}
	err = createdDeployment.Create(ctx, db, dockerManager)
	if err != nil {
		return err
	}
	// add build args to deployment
	createdBuildArgs := make([]BuildArg, 0)
	for _, buildArg := range application.LatestDeployment.BuildArgs {
		createdBuildArg := BuildArg{
			DeploymentID: createdDeployment.ID,
			Key:          buildArg.Key,
			Value:        buildArg.Value,
		}
		createdBuildArgs = append(createdBuildArgs, createdBuildArg)
	}
	tx = db.Create(&createdBuildArgs)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
	// TODO: push to queue for deployment
}

func (application *Application) Update(ctx context.Context, db gorm.DB, dockerManager containermanger.Manager) error {
	// ensure that application is not deleted
	isDeleted, err := application.IsApplicationDeleted(ctx, db)
	if err != nil {
		return err
	}
	if isDeleted {
		return errors.New("application is deleted")
	}
	// TODO: add validation, create new deployment if change required
	// create transaction
	transaction := db.Begin()
	// update environment variables -- if required
	// update persistent volume bindings -- if required
	// update deployment -- if required
	// reload application -- if changed
	return transaction.Commit().Error
}

func (application *Application) Delete(ctx context.Context, db gorm.DB, dockerManager containermanger.Manager) error {
	// ensure that application is not deleted
	isDeleted, err := application.IsApplicationDeleted(ctx, db)
	if err != nil {
		return err
	}
	if isDeleted {
		return errors.New("application is deleted")
	}
	// do soft delete
	tx := db.Model(&application).Update("is_deleted", true)
	return tx.Error
}

func (application *Application) IsApplicationDeleted(ctx context.Context, db gorm.DB) (bool, error) {
	// verify from database
	var count int64
	tx := db.Model(&Application{}).Where("id = ? AND is_deleted = ?", application.ID, true).Count(&count)
	if tx.Error != nil {
		return false, tx.Error
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}
