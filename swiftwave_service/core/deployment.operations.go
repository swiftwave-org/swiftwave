package core

import (
	"context"
	"github.com/google/uuid"
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

func FindCurrentLiveDeploymentByApplicationId(ctx context.Context, db gorm.DB, id string) (*Deployment, error) {
	var deployment = &Deployment{}
	tx := db.Where("application_id = ? AND status = ?", id, DeploymentStatusLive).Order("created_at desc").First(&deployment)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return deployment, nil
}

func FindLatestDeploymentIDByApplicationId(ctx context.Context, db gorm.DB, id string) (string, error) {
	var deployment = &Deployment{}
	tx := db.Select("id").Where("application_id = ?", id).Order("created_at desc").First(&deployment)
	if tx.Error != nil {
		return "", tx.Error
	}
	return deployment.ID, nil
}

func FindDeploymentsByApplicationId(ctx context.Context, db gorm.DB, id string) ([]*Deployment, error) {
	var deployments = make([]*Deployment, 0)
	tx := db.Where("application_id = ?", id).Order("created_at desc").Find(&deployments)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return deployments, nil
}

func IsDeploymentFailed(ctx context.Context, db gorm.DB, id string) (bool, error) {
	var deployment = &Deployment{}
	tx := db.Select("status").Where("id = ?", id).First(&deployment)
	if tx.Error != nil {
		return false, tx.Error
	}
	return deployment.Status == DeploymentStatusFailed, nil
}

func (deployment *Deployment) FindById(ctx context.Context, db gorm.DB, id string) error {
	tx := db.First(&deployment, "id = ?", id)
	return tx.Error
}

func (deployment *Deployment) Create(ctx context.Context, db gorm.DB) error {
	deployment.ID = uuid.NewString()
	deployment.CreatedAt = time.Now()
	deployment.Status = DeploymentStatusPending
	tx := db.Create(&deployment)
	return tx.Error
}

// Update : it will works like a total ledger
// always recreate deployment, no update [except status]
// fetch latest deployment
// the `deployment` object seem to be updated by the caller and ID should be the old one
func (deployment *Deployment) Update(ctx context.Context, db gorm.DB) (*DeploymentUpdateResult, error) {
	// fetch latest deployment
	latestDeployment := &Deployment{}
	tx := db.Preload("BuildArgs").Find(&latestDeployment, "id = ?", deployment.ID)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// state
	result := &DeploymentUpdateResult{
		RebuildRequired: false,
		DeploymentId:    latestDeployment.ID,
	}
	recreationRequired := false
	// verify all params are same
	if deployment.UpstreamType != latestDeployment.UpstreamType ||
		deployment.GitCredentialID != latestDeployment.GitCredentialID ||
		deployment.GitProvider != latestDeployment.GitProvider ||
		deployment.RepositoryOwner != latestDeployment.RepositoryOwner ||
		deployment.RepositoryName != latestDeployment.RepositoryName ||
		deployment.RepositoryBranch != latestDeployment.RepositoryBranch ||
		deployment.CommitHash != latestDeployment.CommitHash ||
		deployment.SourceCodeCompressedFileName != latestDeployment.SourceCodeCompressedFileName ||
		deployment.DockerImage != latestDeployment.DockerImage ||
		deployment.ImageRegistryCredentialID != latestDeployment.ImageRegistryCredentialID ||
		deployment.Dockerfile != latestDeployment.Dockerfile {
		recreationRequired = true
	}
	// verify build args
	if len(deployment.BuildArgs) != len(latestDeployment.BuildArgs) {
		recreationRequired = true
	} else {
		// create map of latest build args
		latestBuildArgsMap := make(map[string]string)
		for _, buildArg := range latestDeployment.BuildArgs {
			latestBuildArgsMap[buildArg.Key] = buildArg.Value
		}
		// verify all build args are same
		for _, buildArg := range deployment.BuildArgs {
			if latestBuildArgsMap[buildArg.Key] != buildArg.Value {
				recreationRequired = true
				break
			}
		}
	}

	// recreate deployment
	if recreationRequired {
		err := deployment.Create(ctx, db)
		if err != nil {
			return nil, err
		} else {
			result.RebuildRequired = true
			result.DeploymentId = deployment.ID
			return result, nil
		}
	}

	return result, nil
}

func (deployment *Deployment) UpdateStatus(ctx context.Context, db gorm.DB, status DeploymentStatus) error {
	// update status
	tx := db.Model(&deployment).Update("status", status)
	return tx.Error
}

func (deployment *Deployment) Delete(ctx context.Context, db gorm.DB) error {
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

func (deployment *Deployment) DeployableDockerImageURI() string {
	// TODO : add support for providing generated docker image uri
	if deployment.UpstreamType == UpstreamTypeImage {
		return deployment.DockerImage
	} else if deployment.UpstreamType == UpstreamTypeGit {
		return deployment.ApplicationID + ":" + deployment.ID
	} else if deployment.UpstreamType == UpstreamTypeSourceCode {
		return deployment.ApplicationID + ":" + deployment.ID
	} else {
		return ""
	}
}

func (deployment *Deployment) GitRepositoryURL() string {
	if deployment.UpstreamType != UpstreamTypeGit {
		return ""
	}
	if deployment.GitProvider == GitProviderGithub {
		return "https://github.com/" + deployment.RepositoryOwner + "/" + deployment.RepositoryName + ".git"
	}
	if deployment.GitProvider == GitProviderGitlab {
		return "https://gitlab.com/" + deployment.RepositoryOwner + "/" + deployment.RepositoryName + ".git"
	}
	return ""
}

// Extra functions for resolvers

func FindDeploymentsByImageRegistryCredentialId(ctx context.Context, db gorm.DB, id uint) ([]*Deployment, error) {
	var deployments = make([]*Deployment, 0)
	tx := db.Where("image_registry_credential_id = ?", id).Find(&deployments)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return deployments, nil
}

func FindDeploymentsByGitCredentialId(ctx context.Context, db gorm.DB, id uint) ([]*Deployment, error) {
	var deployments = make([]*Deployment, 0)
	tx := db.Where("git_credential_id = ?", id).Find(&deployments)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return deployments, nil
}

func FindDeploymentStatusByID(ctx context.Context, db gorm.DB, id string) (*DeploymentStatus, error) {
	var deployment = &Deployment{}
	tx := db.Select("status").Where("id = ?", id).First(&deployment)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &deployment.Status, nil
}
