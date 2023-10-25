package graphql

import (
	dbmodel "github.com/swiftwave-org/swiftwave/swiftwave_manager/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_manager/graphql/model"
	"time"
)

// This file contains object mappers
// 1. Convert Database models to GraphQL models > <type>ToGraphqlObject.go
// 2. Convert GraphQL models to Database models > <type>ToDatabaseObject.go

// Why _ToDatabaseObject() dont adding ID field?
// Because ID field is provided directly to Mutation or Query function

// gitCredentialToGraphqlObject : converts GitCredential to GitCredentialGraphqlObject
func gitCredentialToGraphqlObject(record *dbmodel.GitCredential) *model.GitCredential {
	return &model.GitCredential{
		ID:       record.ID,
		Name:     record.Name,
		Username: record.Username,
		Password: record.Password,
	}
}

// gitCredentialInputToDatabaseObject : converts GitCredentialInput to GitCredentialDatabaseObject
func gitCredentialInputToDatabaseObject(record *model.GitCredentialInput) *dbmodel.GitCredential {
	return &dbmodel.GitCredential{
		Name:     record.Name,
		Username: record.Username,
		Password: record.Password,
	}
}

// imageRegistryCredentialToGraphqlObject : converts ImageRegistryCredential to ImageRegistryCredentialGraphqlObject
func imageRegistryCredentialToGraphqlObject(record *dbmodel.ImageRegistryCredential) *model.ImageRegistryCredential {
	return &model.ImageRegistryCredential{
		ID:       record.ID,
		URL:      record.Url,
		Username: record.Username,
		Password: record.Password,
	}
}

// imageRegistryCredentialInputToDatabaseObject : converts ImageRegistryCredentialInput to ImageRegistryCredentialDatabaseObject
func imageRegistryCredentialInputToDatabaseObject(record *model.ImageRegistryCredentialInput) *dbmodel.ImageRegistryCredential {
	return &dbmodel.ImageRegistryCredential{
		Url:      record.URL,
		Username: record.Username,
		Password: record.Password,
	}
}

// persistentVolumeToGraphqlObject : converts PersistentVolume to PersistentVolumeGraphqlObject
func persistentVolumeToGraphqlObject(record *dbmodel.PersistentVolume) *model.PersistentVolume {
	return &model.PersistentVolume{
		ID:   record.ID,
		Name: record.Name,
	}
}

// persistentVolumeInputToDatabaseObject : converts PersistentVolumeInput to PersistentVolumeDatabaseObject
func persistentVolumeInputToDatabaseObject(record *model.PersistentVolumeInput) *dbmodel.PersistentVolume {
	return &dbmodel.PersistentVolume{
		Name: record.Name,
	}
}

// persistentVolumeBindingInputToDatabaseObject : converts PersistentVolumeBindingInput to PersistentVolumeBindingDatabaseObject
func persistentVolumeBindingInputToDatabaseObject(record *model.PersistentVolumeBindingInput) *dbmodel.PersistentVolumeBinding {
	return &dbmodel.PersistentVolumeBinding{
		PersistentVolumeID: record.PersistentVolumeID,
		MountingPath:       record.MountingPath,
	}
}

// persistentVolumeBindingToGraphqlObject : converts PersistentVolumeBinding to PersistentVolumeBindingGraphqlObject
func persistentVolumeBindingToGraphqlObject(record *dbmodel.PersistentVolumeBinding) *model.PersistentVolumeBinding {
	return &model.PersistentVolumeBinding{
		ID:                 record.ID,
		PersistentVolumeID: record.PersistentVolumeID,
		MountingPath:       record.MountingPath,
	}
}

// environmentVariableInputToDatabaseObject : converts EnvironmentVariableInput to EnvironmentVariableDatabaseObject
func environmentVariableInputToDatabaseObject(record *model.EnvironmentVariableInput) *dbmodel.EnvironmentVariable {
	return &dbmodel.EnvironmentVariable{
		Key:   record.Key,
		Value: record.Value,
	}
}

// environmentVariableToGraphqlObject : converts EnvironmentVariable to EnvironmentVariableGraphqlObject
func environmentVariableToGraphqlObject(record *dbmodel.EnvironmentVariable) *model.EnvironmentVariable {
	return &model.EnvironmentVariable{
		Key:   record.Key,
		Value: record.Value,
	}
}

// buildArgInputToDatabaseObject : converts BuildArgInput to BuildArgDatabaseObject
func buildArgInputToDatabaseObject(record *model.BuildArgInput) *dbmodel.BuildArg {
	return &dbmodel.BuildArg{
		Key:   record.Key,
		Value: record.Value,
	}
}

// buildArgToGraphqlObject : converts BuildArg to BuildArgGraphqlObject
func buildArgToGraphqlObject(record *dbmodel.BuildArg) *model.BuildArg {
	return &model.BuildArg{
		Key:   record.Key,
		Value: record.Value,
	}
}

// applicationInputToDeploymentDatabaseObject : converts ApplicationInput to DeploymentDatabaseObject
func applicationInputToDeploymentDatabaseObject(record *model.ApplicationInput) *dbmodel.Deployment {
	var buildArgs = make([]dbmodel.BuildArg, 0)
	for _, buildArg := range record.BuildArgs {
		buildArgs = append(buildArgs, *buildArgInputToDatabaseObject(buildArg))
	}
	return &dbmodel.Deployment{
		UpstreamType:                 dbmodel.UpstreamType(record.UpstreamType), // TODO: Check this
		GitCredentialID:              DefaultUint(record.GitCredentialID, 0),
		GitProvider:                  dbmodel.GitProvider(DefaultGitProvider(record.GitProvider)),
		RepositoryOwner:              DefaultString(record.RepositoryOwner, ""),
		RepositoryName:               DefaultString(record.RepositoryName, ""),
		RepositoryBranch:             DefaultString(record.RepositoryBranch, ""),
		CommitHash:                   "",
		SourceCodeCompressedFileName: DefaultString(record.SourceCodeCompressedFileName, ""),
		DockerImage:                  DefaultString(record.DockerImage, ""),
		ImageRegistryCredentialID:    DefaultUint(record.ImageRegistryCredentialID, 0),
		BuildArgs:                    buildArgs,
		Dockerfile:                   DefaultString(record.Dockerfile, ""),
		Logs:                         make([]dbmodel.DeploymentLog, 0),
		Status:                       dbmodel.DeploymentStatusPending,
		CreatedAt:                    time.Now(),
	}
}

// applicationInputToDatabaseObject : converts ApplicationInput to ApplicationDatabaseObject
func applicationInputToDatabaseObject(record *model.ApplicationInput) *dbmodel.Application {
	var environmentVariables = make([]dbmodel.EnvironmentVariable, 0)
	for _, environmentVariable := range record.EnvironmentVariables {
		environmentVariables = append(environmentVariables, *environmentVariableInputToDatabaseObject(environmentVariable))
	}
	var persistentVolumeBindings = make([]dbmodel.PersistentVolumeBinding, 0)
	for _, persistentVolumeBinding := range record.PersistentVolumeBindings {
		persistentVolumeBindings = append(persistentVolumeBindings, *persistentVolumeBindingInputToDatabaseObject(persistentVolumeBinding))
	}
	return &dbmodel.Application{
		Name:                     record.Name,
		EnvironmentVariables:     environmentVariables,
		PersistentVolumeBindings: persistentVolumeBindings,
		DeploymentMode:           dbmodel.DeploymentMode(record.DeploymentMode),
		Replicas:                 DefaultUint(record.Replicas, 0),
		LatestDeployment:         *applicationInputToDeploymentDatabaseObject(record),
		Deployments:              make([]dbmodel.Deployment, 0),
		IngressRules:             make([]dbmodel.IngressRule, 0),
	}
}

// applicationToGraphqlObject : converts Application to ApplicationGraphqlObject
func applicationToGraphqlObject(record *dbmodel.Application) *model.Application {
	return &model.Application{
		ID:             record.ID,
		Name:           record.Name,
		DeploymentMode: model.DeploymentMode(record.DeploymentMode),
		Replicas:       record.Replicas,
	}
}

// deploymentToGraphqlObject : converts Deployment to DeploymentGraphqlObject
func deploymentToGraphqlObject(record *dbmodel.Deployment) *model.Deployment {
	return &model.Deployment{
		ID:                           record.ID,
		ApplicationID:                record.ApplicationID,
		UpstreamType:                 model.UpstreamType(record.UpstreamType),
		GitCredentialID:              record.GitCredentialID,
		GitProvider:                  model.GitProvider(record.GitProvider),
		RepositoryOwner:              record.RepositoryOwner,
		RepositoryName:               record.RepositoryName,
		RepositoryBranch:             record.RepositoryBranch,
		CommitHash:                   record.CommitHash,
		SourceCodeCompressedFileName: record.SourceCodeCompressedFileName,
		DockerImage:                  record.DockerImage,
		ImageRegistryCredentialID:    record.ImageRegistryCredentialID,
		Dockerfile:                   record.Dockerfile,
		Status:                       model.DeploymentStatus(record.Status),
		CreatedAt:                    record.CreatedAt,
	}
}
