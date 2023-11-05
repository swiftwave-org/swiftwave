package core

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/hashicorp/go-set"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
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
	tx := db.Where("id = ?", id).First(&application)
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
	// TODO: split this function into smaller functions
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
		var gitCredential = &GitCredential{}
		err := gitCredential.FindById(ctx, db, *application.LatestDeployment.GitCredentialID)
		if err != nil {
			return err
		}
	}
	// For UpstreamType = Image, verify image registry credential id
	if application.LatestDeployment.UpstreamType == UpstreamTypeImage {
		if application.LatestDeployment.ImageRegistryCredentialID != nil {
			var imageRegistryCredential = &ImageRegistryCredential{}
			err := imageRegistryCredential.FindById(ctx, db, *application.LatestDeployment.ImageRegistryCredentialID)
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
	createdApplication := Application{
		ID:             uuid.NewString(),
		Name:           application.Name,
		DeploymentMode: application.DeploymentMode,
		Replicas:       uint(int(application.Replicas)),
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
	if len(createdEnvironmentVariables) > 0 {
		tx = db.Create(&createdEnvironmentVariables)
		if tx.Error != nil {
			return tx.Error
		}
	}
	// create persistent volume bindings
	createdPersistentVolumeBindings := make([]PersistentVolumeBinding, 0)
	persistedVolumeBindingsMountingPathSet := set.From[string](make([]string, 0))
	for _, persistentVolumeBinding := range application.PersistentVolumeBindings {
		// check if mounting path is already used
		if persistedVolumeBindingsMountingPathSet.Contains(persistentVolumeBinding.MountingPath) {
			return errors.New("mounting path already used")
		} else {
			persistedVolumeBindingsMountingPathSet.Insert(persistentVolumeBinding.MountingPath)
		}
		// verify persistent volume exists
		var persistentVolume = &PersistentVolume{}
		err := persistentVolume.FindById(ctx, db, persistentVolumeBinding.PersistentVolumeID)
		if err != nil {
			return err
		}
		createdPersistentVolumeBinding := PersistentVolumeBinding{
			ApplicationID:      createdApplication.ID,
			PersistentVolumeID: persistentVolumeBinding.PersistentVolumeID,
			MountingPath:       persistentVolumeBinding.MountingPath,
		}
		createdPersistentVolumeBindings = append(createdPersistentVolumeBindings, createdPersistentVolumeBinding)
	}
	if len(createdPersistentVolumeBindings) > 0 {
		tx = db.Create(&createdPersistentVolumeBindings)
		if tx.Error != nil {
			return tx.Error
		}
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
	if len(createdBuildArgs) > 0 {
		tx = db.Create(&createdBuildArgs)
		if tx.Error != nil {
			return tx.Error
		}
	}
	// update application details
	*application = createdApplication
	return nil
}

func (application *Application) Update(ctx context.Context, db gorm.DB, dockerManager containermanger.Manager) (*ApplicationUpdateResult, error) {
	var err error
	// ensure that application is not deleted
	isDeleted, err := application.IsApplicationDeleted(ctx, db)
	if err != nil {
		return nil, err
	}
	if isDeleted {
		return nil, errors.New("application is deleted")
	}
	// status
	isReloadRequired := false
	// fetch application with environment variables and persistent volume bindings
	var applicationFull = &Application{}
	tx := db.Preload("EnvironmentVariables").Preload("PersistentVolumeBindings").First(&applicationFull, application.ID)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// create array of environment variables
	var newEnvironmentVariableMap = make(map[string]string)
	for _, environmentVariable := range application.EnvironmentVariables {
		newEnvironmentVariableMap[environmentVariable.Key] = environmentVariable.Value
	}
	// update environment variables -- if required
	if applicationFull.EnvironmentVariables != nil {
		for _, environmentVariable := range applicationFull.EnvironmentVariables {
			// check if environment variable is present in new environment variables
			if _, ok := newEnvironmentVariableMap[environmentVariable.Key]; ok {
				// check if value is changed
				if environmentVariable.Value != newEnvironmentVariableMap[environmentVariable.Key] {
					// update environment variable
					environmentVariable.Value = newEnvironmentVariableMap[environmentVariable.Key]
					err = environmentVariable.Update(ctx, db)
					if err != nil {
						return nil, err
					}
					// delete from newEnvironmentVariableMap
					delete(newEnvironmentVariableMap, environmentVariable.Key)
					// reload application
					isReloadRequired = true
				}
			} else {
				// delete environment variable
				err = environmentVariable.Delete(ctx, db)
				if err != nil {
					return nil, err
				}
				// reload application
				isReloadRequired = true
			}
		}
	}
	// add new environment variables which are not present
	for key, value := range newEnvironmentVariableMap {
		environmentVariable := EnvironmentVariable{
			ApplicationID: application.ID,
			Key:           key,
			Value:         value,
		}
		err := environmentVariable.Create(ctx, db)
		if err != nil {
			return nil, err
		}
		// reload application
		isReloadRequired = true
	}
	// create array of persistent volume bindings
	var newPersistentVolumeBindingMap = make(map[string]uint)
	newPersistentVolumeBindingMountingPathSet := set.From[string](make([]string, 0))
	for _, persistentVolumeBinding := range application.PersistentVolumeBindings {
		// check if mounting path is already used
		if newPersistentVolumeBindingMountingPathSet.Contains(persistentVolumeBinding.MountingPath) {
			return nil, errors.New("duplicate mounting path found")
		} else {
			newPersistentVolumeBindingMountingPathSet.Insert(persistentVolumeBinding.MountingPath)
		}
		newPersistentVolumeBindingMap[persistentVolumeBinding.MountingPath] = persistentVolumeBinding.PersistentVolumeID
	}
	// update persistent volume bindings -- if required
	if applicationFull.PersistentVolumeBindings != nil {
		for _, persistentVolumeBinding := range applicationFull.PersistentVolumeBindings {
			// check if persistent volume binding is present in new persistent volume bindings
			if _, ok := newPersistentVolumeBindingMap[persistentVolumeBinding.MountingPath]; ok {
				// check if value is changed
				if persistentVolumeBinding.PersistentVolumeID != newPersistentVolumeBindingMap[persistentVolumeBinding.MountingPath] {
					// update persistent volume binding
					persistentVolumeBinding.PersistentVolumeID = newPersistentVolumeBindingMap[persistentVolumeBinding.MountingPath]
					err = persistentVolumeBinding.Update(ctx, db)
					if err != nil {
						return nil, err
					}
					// delete from newPersistentVolumeBindingMap
					delete(newPersistentVolumeBindingMap, persistentVolumeBinding.MountingPath)
					// reload application
					isReloadRequired = true
				}
			} else {
				// delete persistent volume binding
				err = persistentVolumeBinding.Delete(ctx, db)
				if err != nil {
					return nil, err
				}
				// reload application
				isReloadRequired = true
			}
		}
	}
	// add new persistent volume bindings which are not present
	for mountingPath, persistentVolumeID := range newPersistentVolumeBindingMap {
		persistentVolumeBinding := PersistentVolumeBinding{
			ApplicationID:      application.ID,
			PersistentVolumeID: persistentVolumeID,
			MountingPath:       mountingPath,
		}
		err := persistentVolumeBinding.Create(ctx, db)
		if err != nil {
			return nil, err
		}
		// reload application
		isReloadRequired = true
	}
	// update deployment -- if required
	currentDeploymentID, err := FindLatestDeploymentIDByApplicationId(ctx, db, application.ID)
	if err != nil {
		return nil, err
	}
	// set deployment id
	application.LatestDeployment.ID = currentDeploymentID
	// send call to update deployment
	updateDeploymentStatus, err := application.LatestDeployment.Update(ctx, db, dockerManager)
	if err != nil {
		return nil, err
	}
	return &ApplicationUpdateResult{
		ReloadRequired:  isReloadRequired,
		RebuildRequired: updateDeploymentStatus.RebuildRequired,
	}, nil
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
	// ensure there is no ingress rule associated with this application
	ingressRules, err := FindIngressRulesByApplicationID(ctx, db, application.ID)
	if err != nil {
		return err
	}
	if ingressRules != nil && len(ingressRules) > 0 {
		return errors.New("application has ingress rules associated with it")
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
