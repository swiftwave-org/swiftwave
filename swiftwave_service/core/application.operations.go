package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/hashicorp/go-set"
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
	tx := db.Where("id = ?", id).First(&application)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (application *Application) FindByName(ctx context.Context, db gorm.DB, name string) error {
	tx := db.Where("name = ?", name).First(&application)
	if tx.Error != nil {
		return tx.Error
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
	// State
	isGitCredentialExist := false
	isImageRegistryCredentialExist := false
	// For UpstreamType = Git, verify git record id
	if application.LatestDeployment.UpstreamType == UpstreamTypeGit {
		if application.LatestDeployment.GitCredentialID != nil {
			var gitCredential = &GitCredential{}
			err := gitCredential.FindById(ctx, db, *application.LatestDeployment.GitCredentialID)
			if err != nil {
				return err
			}
			isGitCredentialExist = true
		} else {
			isGitCredentialExist = false
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
			isImageRegistryCredentialExist = true
		} else {
			isImageRegistryCredentialExist = false
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
		Replicas:       application.Replicas,
		WebhookToken:   uuid.NewString(),
		Command:        application.Command,
		Capabilities:   application.Capabilities,
		Sysctls:        application.Sysctls,
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
	var gitCredentialID *uint = nil
	if isGitCredentialExist {
		gitCredentialID = application.LatestDeployment.GitCredentialID
	}
	var imageRegistryCredentialID *uint = nil
	if isImageRegistryCredentialExist {
		imageRegistryCredentialID = application.LatestDeployment.ImageRegistryCredentialID
	}
	// create deployment
	createdDeployment := Deployment{
		ApplicationID: createdApplication.ID,
		UpstreamType:  application.LatestDeployment.UpstreamType,
		// Fields for UpstreamType = Git
		GitCredentialID:  gitCredentialID,
		GitProvider:      application.LatestDeployment.GitProvider,
		RepositoryOwner:  application.LatestDeployment.RepositoryOwner,
		RepositoryName:   application.LatestDeployment.RepositoryName,
		RepositoryBranch: application.LatestDeployment.RepositoryBranch,
		CommitHash:       application.LatestDeployment.CommitHash,
		CodePath:         application.LatestDeployment.CodePath,
		// Fields for UpstreamType = SourceCode
		SourceCodeCompressedFileName: application.LatestDeployment.SourceCodeCompressedFileName,
		// Fields for UpstreamType = Image
		DockerImage:               application.LatestDeployment.DockerImage,
		ImageRegistryCredentialID: imageRegistryCredentialID,
		// other fields
		Dockerfile: application.LatestDeployment.Dockerfile,
	}
	err = createdDeployment.Create(ctx, db)
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
	var applicationExistingFull = &Application{}
	tx := db.Preload("EnvironmentVariables").Preload("PersistentVolumeBindings").Where("id = ?", application.ID).First(&applicationExistingFull)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// check if DeploymentMode is changed
	if applicationExistingFull.DeploymentMode != application.DeploymentMode {
		// update deployment mode
		err = db.Model(&applicationExistingFull).Update("deployment_mode", application.DeploymentMode).Error
		if err != nil {
			return nil, err
		}
		// reload application
		isReloadRequired = true
	}
	// check if Command is changed
	if applicationExistingFull.Command != application.Command {
		// update command
		err = db.Model(&applicationExistingFull).Update("command", application.Command).Error
		if err != nil {
			return nil, err
		}
		// reload application
		isReloadRequired = true
	}
	// if replicated deployment, check if Replicas is changed
	if application.DeploymentMode == DeploymentModeReplicated && applicationExistingFull.Replicas != application.Replicas {
		// update replicas
		err = db.Model(&applicationExistingFull).Update("replicas", application.Replicas).Error
		if err != nil {
			return nil, err
		}
		// reload application
		isReloadRequired = true
	}
	// create array of environment variables
	var newEnvironmentVariableMap = make(map[string]string)
	for _, environmentVariable := range application.EnvironmentVariables {
		newEnvironmentVariableMap[environmentVariable.Key] = environmentVariable.Value
	}
	// update environment variables -- if required
	if applicationExistingFull.EnvironmentVariables != nil {
		for _, environmentVariable := range applicationExistingFull.EnvironmentVariables {
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
				} else {
					// delete from newEnvironmentVariableMap
					delete(newEnvironmentVariableMap, environmentVariable.Key)
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
	if applicationExistingFull.PersistentVolumeBindings != nil {
		for _, persistentVolumeBinding := range applicationExistingFull.PersistentVolumeBindings {
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
				} else {
					// delete from newPersistentVolumeBindingMap
					delete(newPersistentVolumeBindingMap, persistentVolumeBinding.MountingPath)
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
	currentDeploymentID, err := FindCurrentLiveDeploymentIDByApplicationId(ctx, db, application.ID)
	if err != nil {
		currentDeploymentID, err = FindLatestDeploymentIDByApplicationId(ctx, db, application.ID)
	}
	if err != nil {
		return nil, err
	}
	// set deployment id
	application.LatestDeployment.ID = currentDeploymentID
	// send call to update deployment
	updateDeploymentStatus, err := application.LatestDeployment.Update(ctx, db)
	if err != nil {
		return nil, err
	}
	return &ApplicationUpdateResult{
		ReloadRequired:  isReloadRequired,
		RebuildRequired: updateDeploymentStatus.RebuildRequired,
		DeploymentId:    updateDeploymentStatus.DeploymentId,
	}, nil
}

func (application *Application) SoftDelete(ctx context.Context, db gorm.DB, dockerManager containermanger.Manager) error {
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
	if len(ingressRules) > 0 {
		return errors.New("application has ingress rules associated with it")
	}
	// do soft delete
	tx := db.Model(&application).Update("is_deleted", true)
	return tx.Error
}

func (application *Application) HardDelete(ctx context.Context, db gorm.DB, dockerManager containermanger.Manager) error {
	// ensure there is no ingress rule associated with this application
	ingressRules, err := FindIngressRulesByApplicationID(ctx, db, application.ID)
	if err != nil {
		return err
	}
	if len(ingressRules) > 0 {
		return errors.New("application has ingress rules associated with it")
	}
	// delete application
	tx := db.Delete(&application)
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

func (application *Application) RebuildApplication(ctx context.Context, db gorm.DB) (deploymentId string, error error) {
	// fetch record
	err := application.FindById(ctx, db, application.ID)
	if err != nil {
		return "", err
	}
	// create a new deployment from latest deployment
	latestDeployment, err := FindCurrentLiveDeploymentByApplicationId(ctx, db, application.ID)
	if err != nil {
		latestDeployment, err = FindLatestDeploymentByApplicationId(ctx, db, application.ID)
		if err != nil {
			return "", errors.New("failed to fetch latest deployment")
		}
	}

	// fetch build args
	buildArgs, err := FindBuildArgsByDeploymentId(ctx, db, latestDeployment.ID)
	if err != nil {
		return "", err
	}
	// add new deployment
	err = latestDeployment.Create(ctx, db)
	if err != nil {
		return "", err
	}
	// update build args
	for _, buildArg := range buildArgs {
		buildArg.ID = 0
		buildArg.DeploymentID = latestDeployment.ID
	}
	if len(buildArgs) > 0 {
		err = db.Create(&buildArgs).Error
		if err != nil {
			return "", err
		}
	}
	return latestDeployment.ID, nil
}

func (application *Application) RegenerateWebhookToken(ctx context.Context, db gorm.DB) error {
	// fetch record
	err := application.FindById(ctx, db, application.ID)
	if err != nil {
		return err
	}
	// update webhook token
	application.WebhookToken = uuid.NewString()
	tx := db.Model(&application).Update("webhook_token", application.WebhookToken)
	return tx.Error
}

func (application *Application) MarkAsSleeping(ctx context.Context, db gorm.DB) error {
	// fetch record
	err := application.FindById(ctx, db, application.ID)
	if err != nil {
		return err
	}
	if application.DeploymentMode == DeploymentModeGlobal {
		return errors.New("global deployment cannot be marked as sleeping")
	}
	// update is sleeping
	tx := db.Model(&application).Update("is_sleeping", true)
	return tx.Error
}

func (application *Application) MarkAsWake(ctx context.Context, db gorm.DB) error {
	// fetch record
	err := application.FindById(ctx, db, application.ID)
	if err != nil {
		return err
	}
	if application.DeploymentMode == DeploymentModeGlobal {
		return errors.New("global deployment cannot be marked as wake")
	}
	// update is sleeping
	tx := db.Model(&application).Update("is_sleeping", false)
	return tx.Error
}
