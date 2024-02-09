package graphql

import (
	"time"

	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/graphql/model"
)

// This file contains object mappers
// 1. Convert Database models to GraphQL models > <type>ToGraphqlObject.go
// 2. Convert GraphQL models to Database models > <type>ToDatabaseObject.go

// Why _ToDatabaseObject() dont adding ID field?
// Because ID field is provided directly to Mutation or Query function

// gitCredentialToGraphqlObject : converts GitCredential to GitCredentialGraphqlObject
func gitCredentialToGraphqlObject(record *core.GitCredential) *model.GitCredential {
	return &model.GitCredential{
		ID:       record.ID,
		Name:     record.Name,
		Username: record.Username,
		Password: record.Password,
	}
}

// gitCredentialInputToDatabaseObject : converts GitCredentialInput to GitCredentialDatabaseObject
func gitCredentialInputToDatabaseObject(record *model.GitCredentialInput) *core.GitCredential {
	return &core.GitCredential{
		Name:     record.Name,
		Username: record.Username,
		Password: record.Password,
	}
}

// imageRegistryCredentialToGraphqlObject : converts ImageRegistryCredential to ImageRegistryCredentialGraphqlObject
func imageRegistryCredentialToGraphqlObject(record *core.ImageRegistryCredential) *model.ImageRegistryCredential {
	return &model.ImageRegistryCredential{
		ID:       record.ID,
		URL:      record.Url,
		Username: record.Username,
		Password: record.Password,
	}
}

// imageRegistryCredentialInputToDatabaseObject : converts ImageRegistryCredentialInput to ImageRegistryCredentialDatabaseObject
func imageRegistryCredentialInputToDatabaseObject(record *model.ImageRegistryCredentialInput) *core.ImageRegistryCredential {
	return &core.ImageRegistryCredential{
		Url:      record.URL,
		Username: record.Username,
		Password: record.Password,
	}
}

// persistentVolumeToGraphqlObject : converts PersistentVolume to PersistentVolumeGraphqlObject
func persistentVolumeToGraphqlObject(record *core.PersistentVolume) *model.PersistentVolume {
	return &model.PersistentVolume{
		ID:   record.ID,
		Name: record.Name,
	}
}

// persistentVolumeInputToDatabaseObject : converts PersistentVolumeInput to PersistentVolumeDatabaseObject
func persistentVolumeInputToDatabaseObject(record *model.PersistentVolumeInput) *core.PersistentVolume {
	return &core.PersistentVolume{
		Name: record.Name,
	}
}

// persistentVolumeBindingInputToDatabaseObject : converts PersistentVolumeBindingInput to PersistentVolumeBindingDatabaseObject
func persistentVolumeBindingInputToDatabaseObject(record *model.PersistentVolumeBindingInput) *core.PersistentVolumeBinding {
	return &core.PersistentVolumeBinding{
		PersistentVolumeID: record.PersistentVolumeID,
		MountingPath:       record.MountingPath,
	}
}

// persistentVolumeBindingToGraphqlObject : converts PersistentVolumeBinding to PersistentVolumeBindingGraphqlObject
func persistentVolumeBindingToGraphqlObject(record *core.PersistentVolumeBinding) *model.PersistentVolumeBinding {
	return &model.PersistentVolumeBinding{
		ID:                 record.ID,
		PersistentVolumeID: record.PersistentVolumeID,
		MountingPath:       record.MountingPath,
		ApplicationID:      record.ApplicationID,
	}
}

// environmentVariableInputToDatabaseObject : converts EnvironmentVariableInput to EnvironmentVariableDatabaseObject
func environmentVariableInputToDatabaseObject(record *model.EnvironmentVariableInput) *core.EnvironmentVariable {
	return &core.EnvironmentVariable{
		Key:   record.Key,
		Value: record.Value,
	}
}

// environmentVariableToGraphqlObject : converts EnvironmentVariable to EnvironmentVariableGraphqlObject
func environmentVariableToGraphqlObject(record *core.EnvironmentVariable) *model.EnvironmentVariable {
	return &model.EnvironmentVariable{
		Key:   record.Key,
		Value: record.Value,
	}
}

// buildArgInputToDatabaseObject : converts BuildArgInput to BuildArgDatabaseObject
func buildArgInputToDatabaseObject(record *model.BuildArgInput) *core.BuildArg {
	return &core.BuildArg{
		Key:   record.Key,
		Value: record.Value,
	}
}

// buildArgToGraphqlObject : converts BuildArg to BuildArgGraphqlObject
func buildArgToGraphqlObject(record *core.BuildArg) *model.BuildArg {
	return &model.BuildArg{
		Key:   record.Key,
		Value: record.Value,
	}
}

// applicationInputToDeploymentDatabaseObject : converts ApplicationInput to DeploymentDatabaseObject
func applicationInputToDeploymentDatabaseObject(record *model.ApplicationInput) *core.Deployment {
	var buildArgs = make([]core.BuildArg, 0)
	for _, buildArg := range record.BuildArgs {
		buildArgs = append(buildArgs, *buildArgInputToDatabaseObject(buildArg))
	}
	return &core.Deployment{
		UpstreamType:                 core.UpstreamType(record.UpstreamType),
		GitCredentialID:              record.GitCredentialID,
		GitProvider:                  core.GitProvider(DefaultGitProvider(record.GitProvider)),
		RepositoryOwner:              DefaultString(record.RepositoryOwner, ""),
		RepositoryName:               DefaultString(record.RepositoryName, ""),
		RepositoryBranch:             DefaultString(record.RepositoryBranch, ""),
		CommitHash:                   "",
		CodePath:                     DefaultString(record.CodePath, ""),
		SourceCodeCompressedFileName: DefaultString(record.SourceCodeCompressedFileName, ""),
		DockerImage:                  DefaultString(record.DockerImage, ""),
		ImageRegistryCredentialID:    record.ImageRegistryCredentialID,
		BuildArgs:                    buildArgs,
		Dockerfile:                   DefaultString(record.Dockerfile, ""),
		Logs:                         make([]core.DeploymentLog, 0),
		Status:                       core.DeploymentStatusPending,
		CreatedAt:                    time.Now(),
	}
}

// applicationInputToDatabaseObject : converts ApplicationInput to ApplicationDatabaseObject
func applicationInputToDatabaseObject(record *model.ApplicationInput) *core.Application {
	var environmentVariables = make([]core.EnvironmentVariable, 0)
	for _, environmentVariable := range record.EnvironmentVariables {
		environmentVariables = append(environmentVariables, *environmentVariableInputToDatabaseObject(environmentVariable))
	}
	var persistentVolumeBindings = make([]core.PersistentVolumeBinding, 0)
	for _, persistentVolumeBinding := range record.PersistentVolumeBindings {
		persistentVolumeBindings = append(persistentVolumeBindings, *persistentVolumeBindingInputToDatabaseObject(persistentVolumeBinding))
	}
	return &core.Application{
		Name:                     record.Name,
		EnvironmentVariables:     environmentVariables,
		PersistentVolumeBindings: persistentVolumeBindings,
		DeploymentMode:           core.DeploymentMode(record.DeploymentMode),
		Replicas:                 DefaultUint(record.Replicas, 0),
		LatestDeployment:         *applicationInputToDeploymentDatabaseObject(record),
		Deployments:              make([]core.Deployment, 0),
		IngressRules:             make([]core.IngressRule, 0),
		Capabilities:             record.Capabilities,
		Sysctls:                  record.Sysctls,
	}
}

// applicationToGraphqlObject : converts Application to ApplicationGraphqlObject
func applicationToGraphqlObject(record *core.Application) *model.Application {
	return &model.Application{
		ID:             record.ID,
		Name:           record.Name,
		DeploymentMode: model.DeploymentMode(record.DeploymentMode),
		Replicas:       record.Replicas,
		IsDeleted:      record.IsDeleted,
		WebhookToken:   record.WebhookToken,
		Capabilities:   record.Capabilities,
		Sysctls:        record.Sysctls,
	}
}

// deploymentToGraphqlObject : converts Deployment to DeploymentGraphqlObject
func deploymentToGraphqlObject(record *core.Deployment) *model.Deployment {
	gitCredentialId := uint(0)
	if record.GitCredentialID != nil {
		gitCredentialId = *record.GitCredentialID
	}
	imageRegistryCredentialId := uint(0)
	if record.ImageRegistryCredentialID != nil {
		imageRegistryCredentialId = *record.ImageRegistryCredentialID
	}
	return &model.Deployment{
		ID:                           record.ID,
		ApplicationID:                record.ApplicationID,
		UpstreamType:                 model.UpstreamType(record.UpstreamType),
		GitCredentialID:              gitCredentialId,
		GitProvider:                  model.GitProvider(record.GitProvider),
		RepositoryOwner:              record.RepositoryOwner,
		RepositoryName:               record.RepositoryName,
		RepositoryBranch:             record.RepositoryBranch,
		CommitHash:                   record.CommitHash,
		CodePath:                     record.CodePath,
		SourceCodeCompressedFileName: record.SourceCodeCompressedFileName,
		DockerImage:                  record.DockerImage,
		ImageRegistryCredentialID:    imageRegistryCredentialId,
		Dockerfile:                   record.Dockerfile,
		Status:                       model.DeploymentStatus(record.Status),
		CreatedAt:                    record.CreatedAt,
	}
}

// domainInputToDatabaseObject : converts DomainInput to DomainDatabaseObject
func domainInputToDatabaseObject(record *model.DomainInput) *core.Domain {
	return &core.Domain{
		Name:         record.Name,
		SSLStatus:    core.DomainSSLStatusNone,
		SSLAutoRenew: false,
	}
}

// domainToGraphqlObject : converts Domain to DomainGraphqlObject
func domainToGraphqlObject(record *core.Domain) *model.Domain {
	return &model.Domain{
		ID:            record.ID,
		Name:          record.Name,
		SslStatus:     model.DomainSSLStatus(record.SSLStatus),
		SslPrivateKey: record.SSLPrivateKey,
		SslFullChain:  record.SSLFullChain,
		SslIssuedAt:   record.SSLIssuedAt,
		SslIssuer:     record.SSLIssuer,
		SslAutoRenew:  record.SSLAutoRenew,
	}
}

// ingressRuleInputToDatabaseObject : converts IngressRuleInput to IngressRuleDatabaseObject
func ingressRuleInputToDatabaseObject(record *model.IngressRuleInput) *core.IngressRule {
	return &core.IngressRule{
		ApplicationID: record.ApplicationID,
		DomainID:      record.DomainID,
		Protocol:      core.ProtocolType(record.Protocol),
		Port:          record.Port,
		TargetPort:    record.TargetPort,
		Status:        core.IngressRuleStatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// ingressRuleToGraphqlObject : converts IngressRule to IngressRuleGraphqlObject
func ingressRuleToGraphqlObject(record *core.IngressRule) *model.IngressRule {
	return &model.IngressRule{
		ID:            record.ID,
		ApplicationID: record.ApplicationID,
		DomainID:      record.DomainID,
		Protocol:      model.ProtocolType(record.Protocol),
		Port:          record.Port,
		TargetPort:    record.TargetPort,
		Status:        model.IngressRuleStatus(record.Status),
		CreatedAt:     record.CreatedAt,
		UpdatedAt:     record.UpdatedAt,
	}
}

// redirectRuleInputToDatabaseObject : converts RedirectRuleInput to RedirectRuleDatabaseObject
func redirectRuleInputToDatabaseObject(record *model.RedirectRuleInput) *core.RedirectRule {
	return &core.RedirectRule{
		DomainID:    record.DomainID,
		Port:        record.Port,
		RedirectURL: record.RedirectURL,
		Status:      core.RedirectRuleStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// redirectRuleToGraphqlObject : converts RedirectRule to RedirectRuleGraphqlObject
func redirectRuleToGraphqlObject(record *core.RedirectRule) *model.RedirectRule {
	return &model.RedirectRule{
		ID:          record.ID,
		DomainID:    record.DomainID,
		Port:        record.Port,
		RedirectURL: record.RedirectURL,
		Status:      model.RedirectRuleStatus(record.Status),
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}
}

// deploymentLogToGraphqlObject : converts DeploymentLog to DeploymentLogGraphqlObject
func deploymentLogToGraphqlObject(record *core.DeploymentLog) *model.DeploymentLog {
	return &model.DeploymentLog{
		Content:   record.Content,
		CreatedAt: record.CreatedAt,
	}
}

// userToGraphqlObject : converts User to UserGraphqlObject
func userToGraphqlObject(record *core.User) *model.User {
	if record == nil {
		return nil
	}
	return &model.User{
		ID:       record.ID,
		Username: record.Username,
	}
}
