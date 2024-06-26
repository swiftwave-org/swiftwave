package graphql

import (
	"context"
	"crypto"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/dgryski/trifles/uuid"
	"strings"
	"time"

	gitmanager "github.com/swiftwave-org/swiftwave/git_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/stack_parser"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"

	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/graphql/model"
)

// This file contains object mappers
// 1. Convert Database models to GraphQL models > <type>ToGraphqlObject.go
// 2. Convert GraphQL models to Database models > <type>ToDatabaseObject.go

// Why _ToDatabaseObject() dont adding ID field?
// Because ID field is provided directly to Mutation or Query function

// gitCredentialToGraphqlObject converts GitCredential to GitCredentialGraphqlObject
func gitCredentialToGraphqlObject(record *core.GitCredential) *model.GitCredential {
	return &model.GitCredential{
		ID:           record.ID,
		Type:         model.GitType(record.Type),
		Name:         record.Name,
		Username:     record.Username,
		SSHPublicKey: record.SshPublicKey,
	}
}

// gitCredentialInputToDatabaseObject converts GitCredentialInput to GitCredentialDatabaseObject
func gitCredentialInputToDatabaseObject(record *model.GitCredentialInput, createSSHKeysIfNotProvided bool) *core.GitCredential {
	sshPrivateKey := ""
	sshPublicKey := ""

	record.SSHPrivateKey = strings.TrimSpace(record.SSHPrivateKey)

	if record.Type == model.GitTypeSSH && createSSHKeysIfNotProvided {
		if strings.Compare(record.SSHPrivateKey, "") == 0 {
			// create ssh private key
			pub, priv, err := ed25519.GenerateKey(nil)
			if err == nil {
				p, err := ssh.MarshalPrivateKey(crypto.PrivateKey(priv), "")
				if err == nil {
					privateKeyPem := pem.EncodeToMemory(p)
					sshPrivateKey = string(privateKeyPem) + "\n"
					publicKey, err := ssh.NewPublicKey(pub)
					if err == nil {
						sshPublicKey = "ssh-ed25519" + " " + base64.StdEncoding.EncodeToString(publicKey.Marshal()) + " swiftwave"
					}
				}
			}
		} else {
			// parse ssh private key
			sshPrivateKey = record.SSHPrivateKey
			if !strings.HasSuffix(sshPrivateKey, "\n") {
				sshPrivateKey += "\n"
			}
			p, err := ssh.ParsePrivateKey([]byte(sshPrivateKey))
			if err == nil {
				if p.PublicKey().Type() == ssh.KeyAlgoED25519 {
					sshPublicKey = "ssh-ed25519" + " " + base64.StdEncoding.EncodeToString(p.PublicKey().Marshal()) + " swiftwave"
				}
			}
		}
	}

	// parse ssh private key
	return &core.GitCredential{
		Name:          record.Name,
		Type:          core.GitType(record.Type),
		Username:      record.Username,
		Password:      record.Password,
		SshPrivateKey: sshPrivateKey,
		SshPublicKey:  sshPublicKey,
	}
}

// imageRegistryCredentialToGraphqlObject converts ImageRegistryCredential to ImageRegistryCredentialGraphqlObject
func imageRegistryCredentialToGraphqlObject(record *core.ImageRegistryCredential) *model.ImageRegistryCredential {
	return &model.ImageRegistryCredential{
		ID:       record.ID,
		URL:      record.Url,
		Username: record.Username,
		Password: record.Password,
	}
}

// imageRegistryCredentialInputToDatabaseObject converts ImageRegistryCredentialInput to ImageRegistryCredentialDatabaseObject
func imageRegistryCredentialInputToDatabaseObject(record *model.ImageRegistryCredentialInput) *core.ImageRegistryCredential {
	return &core.ImageRegistryCredential{
		Url:      record.URL,
		Username: record.Username,
		Password: record.Password,
	}
}

// persistentVolumeToGraphqlObject converts PersistentVolume to PersistentVolumeGraphqlObject
func persistentVolumeToGraphqlObject(record *core.PersistentVolume) *model.PersistentVolume {
	return &model.PersistentVolume{
		ID:   record.ID,
		Name: record.Name,
		Type: model.PersistentVolumeType(record.Type),
		NfsConfig: &model.NFSConfig{
			Host:    record.NFSConfig.Host,
			Path:    record.NFSConfig.Path,
			Version: record.NFSConfig.Version,
		},
		CifsConfig: &model.CIFSConfig{
			Share:    record.CIFSConfig.Share,
			Host:     record.CIFSConfig.Host,
			Username: record.CIFSConfig.Username,
			Password: record.CIFSConfig.Password,
			FileMode: record.CIFSConfig.FileMode,
			DirMode:  record.CIFSConfig.DirMode,
			UID:      record.CIFSConfig.Uid,
			Gid:      record.CIFSConfig.Gid,
		},
	}
}

// persistentVolumeInputToDatabaseObject converts PersistentVolumeInput to PersistentVolumeDatabaseObject
func persistentVolumeInputToDatabaseObject(record *model.PersistentVolumeInput) *core.PersistentVolume {
	nfsConfig := core.NFSConfig{}
	if record.Type == model.PersistentVolumeTypeNfs {
		nfsConfig = core.NFSConfig{
			Host:    record.NfsConfig.Host,
			Path:    record.NfsConfig.Path,
			Version: record.NfsConfig.Version,
		}
	}
	cifsConfig := core.CIFSConfig{}
	if record.Type == model.PersistentVolumeTypeCifs {
		cifsConfig = core.CIFSConfig{
			Share:    record.CifsConfig.Share,
			Host:     record.CifsConfig.Host,
			Username: record.CifsConfig.Username,
			Password: record.CifsConfig.Password,
			FileMode: record.CifsConfig.FileMode,
			DirMode:  record.CifsConfig.DirMode,
			Uid:      record.CifsConfig.UID,
			Gid:      record.CifsConfig.Gid,
		}
	}
	return &core.PersistentVolume{
		Name:       record.Name,
		Type:       core.PersistentVolumeType(record.Type),
		NFSConfig:  nfsConfig,
		CIFSConfig: cifsConfig,
	}
}

// persistentVolumeBindingInputToDatabaseObject converts PersistentVolumeBindingInput to PersistentVolumeBindingDatabaseObject
func persistentVolumeBindingInputToDatabaseObject(record *model.PersistentVolumeBindingInput) *core.PersistentVolumeBinding {
	return &core.PersistentVolumeBinding{
		PersistentVolumeID: record.PersistentVolumeID,
		MountingPath:       strings.TrimSpace(record.MountingPath),
	}
}

// persistentVolumeBindingToGraphqlObject converts PersistentVolumeBinding to PersistentVolumeBindingGraphqlObject
func persistentVolumeBindingToGraphqlObject(record *core.PersistentVolumeBinding) *model.PersistentVolumeBinding {
	return &model.PersistentVolumeBinding{
		ID:                 record.ID,
		PersistentVolumeID: record.PersistentVolumeID,
		MountingPath:       record.MountingPath,
		ApplicationID:      record.ApplicationID,
	}
}

// persistentVolumeBackupToGraphqlObject converts PersistentVolumeBackup to PersistentVolumeBackupGraphqlObject
func persistentVolumeBackupToGraphqlObject(record *core.PersistentVolumeBackup) *model.PersistentVolumeBackup {
	return &model.PersistentVolumeBackup{
		ID:          record.ID,
		Type:        model.PersistentVolumeBackupType(record.Type),
		Status:      model.PersistentVolumeBackupStatus(record.Status),
		SizeMb:      record.FileSizeMB,
		CreatedAt:   record.CreatedAt,
		CompletedAt: record.CompletedAt,
	}
}

// persistentVolumeBackupInputToDatabaseObject converts PersistentVolumeBackupInput to PersistentVolumeBackupDatabaseObject
func persistentVolumeBackupInputToDatabaseObject(record *model.PersistentVolumeBackupInput) *core.PersistentVolumeBackup {
	return &core.PersistentVolumeBackup{
		Type:               core.BackupType(record.Type),
		Status:             core.BackupPending,
		File:               "",
		FileSizeMB:         0,
		PersistentVolumeID: record.PersistentVolumeID,
		CreatedAt:          time.Now(),
		CompletedAt:        time.Now(),
	}
}

// persistentVolumeRestoreToGraphqlObject converts PersistentVolumeRestore to PersistentVolumeRestoreGraphqlObject
func persistentVolumeRestoreToGraphqlObject(record *core.PersistentVolumeRestore) *model.PersistentVolumeRestore {
	return &model.PersistentVolumeRestore{
		ID:          record.ID,
		Type:        model.PersistentVolumeRestoreType(record.Type),
		Status:      model.PersistentVolumeRestoreStatus(record.Status),
		CreatedAt:   record.CreatedAt,
		CompletedAt: record.CompletedAt,
	}
}

// environmentVariableInputToDatabaseObject converts EnvironmentVariableInput to EnvironmentVariableDatabaseObject
func environmentVariableInputToDatabaseObject(record *model.EnvironmentVariableInput) *core.EnvironmentVariable {
	return &core.EnvironmentVariable{
		Key:   strings.TrimSpace(record.Key),
		Value: strings.TrimSpace(record.Value),
	}
}

// environmentVariableToGraphqlObject converts EnvironmentVariable to EnvironmentVariableGraphqlObject
func environmentVariableToGraphqlObject(record *core.EnvironmentVariable) *model.EnvironmentVariable {
	return &model.EnvironmentVariable{
		Key:   record.Key,
		Value: record.Value,
	}
}

// buildArgInputToDatabaseObject converts BuildArgInput to BuildArgDatabaseObject
func buildArgInputToDatabaseObject(record *model.BuildArgInput) *core.BuildArg {
	return &core.BuildArg{
		Key:   record.Key,
		Value: record.Value,
	}
}

// buildArgToGraphqlObject converts BuildArg to BuildArgGraphqlObject
func buildArgToGraphqlObject(record *core.BuildArg) *model.BuildArg {
	return &model.BuildArg{
		Key:   record.Key,
		Value: record.Value,
	}
}

// configMountInputToDatabaseObject converts ConfigMountInput to ConfigMountDatabaseObject
func configMountInputToDatabaseObject(record *model.ConfigMountInput) *core.ConfigMount {
	return &core.ConfigMount{
		Content:      record.Content,
		Gid:          record.Gid,
		Uid:          record.UID,
		MountingPath: strings.TrimSpace(record.MountingPath),
	}
}

// configMountToGraphqlObject converts ConfigMount to ConfigMountGraphqlObject
func configMountToGraphqlObject(record *core.ConfigMount) *model.ConfigMount {
	return &model.ConfigMount{
		Content:      record.Content,
		Gid:          record.Gid,
		UID:          record.Uid,
		MountingPath: record.MountingPath,
	}
}

// resourceLimitInputToDatabaseObject converts ResourceLimitInput to ResourceLimitDatabaseObject
func resourceLimitInputToDatabaseObject(record *model.ResourceLimitInput) *core.ApplicationResourceLimit {
	return &core.ApplicationResourceLimit{
		MemoryMB: record.MemoryMb,
	}
}

// resourceLimitToGraphqlObject converts ResourceLimit to ResourceLimitGraphqlObject
func resourceLimitToGraphqlObject(record *core.ApplicationResourceLimit) *model.ResourceLimit {
	return &model.ResourceLimit{
		MemoryMb: record.MemoryMB,
	}
}

// reservedResourceInputToDatabaseObject converts ReservedResourceInput to ReservedResourceDatabaseObject
func reservedResourceInputToDatabaseObject(record *model.ReservedResourceInput) *core.ApplicationReservedResource {
	return &core.ApplicationReservedResource{
		MemoryMB: record.MemoryMb,
	}
}

// reservedResourceToGraphqlObject converts ReservedResource to ReservedResourceGraphqlObject
func reservedResourceToGraphqlObject(record *core.ApplicationReservedResource) *model.ReservedResource {
	return &model.ReservedResource{
		MemoryMb: record.MemoryMB,
	}
}

// applicationInputToDeploymentDatabaseObject converts ApplicationInput to DeploymentDatabaseObject
func applicationInputToDeploymentDatabaseObject(record *model.ApplicationInput) *core.Deployment {
	var buildArgs = make([]core.BuildArg, 0)
	for _, buildArg := range record.BuildArgs {
		buildArgs = append(buildArgs, *buildArgInputToDatabaseObject(buildArg))
	}
	var repoInfo gitmanager.GitRepoInfo
	if record.UpstreamType == model.UpstreamTypeGit {
		parsedRepoInfo, _ := gitmanager.ParseGitRepoInfo(*record.RepositoryURL)
		if parsedRepoInfo != nil {
			repoInfo = *parsedRepoInfo
		}
	}
	gitType := core.GitHttp
	if repoInfo.IsSshEndpoint {
		gitType = core.GitSsh
	}
	return &core.Deployment{
		UpstreamType:                 core.UpstreamType(record.UpstreamType),
		GitCredentialID:              record.GitCredentialID,
		GitType:                      gitType,
		GitProvider:                  repoInfo.Provider,
		RepositoryOwner:              repoInfo.Owner,
		RepositoryName:               repoInfo.Name,
		RepositoryBranch:             DefaultString(record.RepositoryBranch, ""),
		GitEndpoint:                  repoInfo.Endpoint,
		GitSshUser:                   repoInfo.SshUser,
		CommitHash:                   "",
		CommitMessage:                "",
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

// applicationGroupInputToDatabaseObject converts ApplicationGroupInput to ApplicationGroupDatabaseObject
func applicationGroupInputToDatabaseObject(record *model.ApplicationGroupInput) *core.ApplicationGroup {
	return &core.ApplicationGroup{
		ID:   uuid.UUIDv4(),
		Name: record.Name,
	}
}

// applicationGroupToGraphqlObject converts ApplicationGroup to ApplicationGroupGraphqlObject
func applicationGroupToGraphqlObject(record *core.ApplicationGroup) *model.ApplicationGroup {
	return &model.ApplicationGroup{
		ID:   record.ID,
		Name: record.Name,
	}
}

// applicationInputToDatabaseObject converts ApplicationInput to ApplicationDatabaseObject
func applicationInputToDatabaseObject(record *model.ApplicationInput) *core.Application {
	var environmentVariables = make([]core.EnvironmentVariable, 0)
	for _, environmentVariable := range record.EnvironmentVariables {
		environmentVariables = append(environmentVariables, *environmentVariableInputToDatabaseObject(environmentVariable))
	}
	var persistentVolumeBindings = make([]core.PersistentVolumeBinding, 0)
	for _, persistentVolumeBinding := range record.PersistentVolumeBindings {
		persistentVolumeBindings = append(persistentVolumeBindings, *persistentVolumeBindingInputToDatabaseObject(persistentVolumeBinding))
	}
	var configMounts = make([]core.ConfigMount, 0)
	for _, configMount := range record.ConfigMounts {
		configMounts = append(configMounts, *configMountInputToDatabaseObject(configMount))
	}
	return &core.Application{
		Name:                     record.Name,
		EnvironmentVariables:     environmentVariables,
		PersistentVolumeBindings: persistentVolumeBindings,
		ConfigMounts:             configMounts,
		DeploymentMode:           core.DeploymentMode(record.DeploymentMode),
		Replicas:                 DefaultUint(record.Replicas, 0),
		LatestDeployment:         *applicationInputToDeploymentDatabaseObject(record),
		Deployments:              make([]core.Deployment, 0),
		IngressRules:             make([]core.IngressRule, 0),
		Command:                  record.Command,
		Capabilities:             record.Capabilities,
		Sysctls:                  record.Sysctls,
		ReservedResource:         *reservedResourceInputToDatabaseObject(record.ReservedResource),
		ResourceLimit:            *resourceLimitInputToDatabaseObject(record.ResourceLimit),
		IsSleeping:               false,
		ApplicationGroupID:       record.ApplicationGroupID,
		PreferredServerHostnames: record.PreferredServerHostnames,
		DockerProxy:              *dockerProxyConfigToDatabaseObject(record.DockerProxyConfig),
		CustomHealthCheck:        *applicationCustomHealthCheckInputToDatabaseObject(record.CustomHealthCheck),
	}
}

// applicationToGraphqlObject converts Application to ApplicationGraphqlObject
func applicationToGraphqlObject(record *core.Application) *model.Application {
	return &model.Application{
		ID:                       record.ID,
		Name:                     record.Name,
		DeploymentMode:           model.DeploymentMode(record.DeploymentMode),
		Replicas:                 record.Replicas,
		IsDeleted:                record.IsDeleted,
		WebhookToken:             record.WebhookToken,
		Capabilities:             record.Capabilities,
		Sysctls:                  record.Sysctls,
		ResourceLimit:            resourceLimitToGraphqlObject(&record.ResourceLimit),
		ReservedResource:         reservedResourceToGraphqlObject(&record.ReservedResource),
		IsSleeping:               record.IsSleeping,
		Command:                  record.Command,
		ApplicationGroupID:       record.ApplicationGroupID,
		PreferredServerHostnames: record.PreferredServerHostnames,
		DockerProxyHost:          record.DockerProxyServiceName(),
		DockerProxyConfig:        dockerProxyConfigToGraphqlObject(&record.DockerProxy),
		CustomHealthCheck:        applicationCustomHealthCheckToGraphqlObject(&record.CustomHealthCheck),
	}
}

// deploymentToGraphqlObject converts Deployment to DeploymentGraphqlObject
func deploymentToGraphqlObject(record *core.Deployment) *model.Deployment {
	gitCredentialId := uint(0)
	if record.GitCredentialID != nil {
		gitCredentialId = *record.GitCredentialID
	}
	imageRegistryCredentialId := uint(0)
	if record.ImageRegistryCredentialID != nil {
		imageRegistryCredentialId = *record.ImageRegistryCredentialID
	}
	repositoryUrl := ""
	if record.UpstreamType == core.UpstreamTypeGit {
		repositoryUrl = record.GitRepositoryURL()
	}
	return &model.Deployment{
		ID:                           record.ID,
		ApplicationID:                record.ApplicationID,
		UpstreamType:                 model.UpstreamType(record.UpstreamType),
		GitEndpoint:                  record.GitEndpoint,
		GitCredentialID:              gitCredentialId,
		GitProvider:                  record.GitProvider,
		RepositoryOwner:              record.RepositoryOwner,
		RepositoryName:               record.RepositoryName,
		RepositoryBranch:             record.RepositoryBranch,
		RepositoryURL:                repositoryUrl,
		CommitHash:                   record.CommitHash,
		CommitMessage:                record.CommitMessage,
		CodePath:                     record.CodePath,
		SourceCodeCompressedFileName: record.SourceCodeCompressedFileName,
		DockerImage:                  record.DockerImage,
		ImageRegistryCredentialID:    imageRegistryCredentialId,
		Dockerfile:                   record.Dockerfile,
		Status:                       model.DeploymentStatus(record.Status),
		CreatedAt:                    record.CreatedAt,
	}
}

// domainInputToDatabaseObject converts DomainInput to DomainDatabaseObject
func domainInputToDatabaseObject(record *model.DomainInput) *core.Domain {
	return &core.Domain{
		Name:         record.Name,
		SSLStatus:    core.DomainSSLStatusNone,
		SslAutoRenew: false,
	}
}

// domainToGraphqlObject converts Domain to DomainGraphqlObject
func domainToGraphqlObject(record *core.Domain) *model.Domain {
	return &model.Domain{
		ID:            record.ID,
		Name:          record.Name,
		SslStatus:     model.DomainSSLStatus(record.SSLStatus),
		SslPrivateKey: record.SSLPrivateKey,
		SslFullChain:  record.SSLFullChain,
		SslIssuedAt:   record.SSLIssuedAt,
		SslIssuer:     record.SSLIssuer,
		SslAutoRenew:  record.SslAutoRenew,
	}
}

// dockerProxyConfigToGraphqlObject converts DockerProxyConfig to DockerProxyConfigGraphqlObject
func dockerProxyConfigToGraphqlObject(record *core.DockerProxyConfig) *model.DockerProxyConfig {
	return &model.DockerProxyConfig{
		Enabled:    record.Enabled,
		Permission: dockerProxyPermissionToGraphqlObject(&record.Permission),
	}
}

// dockerProxyConfigToDatabaseObject converts DockerProxyConfig to DockerProxyConfigDatabaseObject
func dockerProxyConfigToDatabaseObject(record *model.DockerProxyConfigInput) *core.DockerProxyConfig {
	return &core.DockerProxyConfig{
		Enabled:    record.Enabled,
		Permission: *dockerProxyPermissionInputToDatabaseObject(record.Permission),
	}
}

// dockerProxyPermissionToGraphqlObject converts DockerProxyPermission to DockerProxyPermissionGraphqlObject
func dockerProxyPermissionToGraphqlObject(record *core.DockerProxyPermission) *model.DockerProxyPermission {
	return &model.DockerProxyPermission{
		Ping:         model.DockerProxyPermissionType(record.Ping),
		Version:      model.DockerProxyPermissionType(record.Version),
		Info:         model.DockerProxyPermissionType(record.Info),
		Events:       model.DockerProxyPermissionType(record.Events),
		Auth:         model.DockerProxyPermissionType(record.Auth),
		Secrets:      model.DockerProxyPermissionType(record.Secrets),
		Build:        model.DockerProxyPermissionType(record.Build),
		Commit:       model.DockerProxyPermissionType(record.Commit),
		Configs:      model.DockerProxyPermissionType(record.Configs),
		Containers:   model.DockerProxyPermissionType(record.Containers),
		Distribution: model.DockerProxyPermissionType(record.Distribution),
		Exec:         model.DockerProxyPermissionType(record.Exec),
		Grpc:         model.DockerProxyPermissionType(record.Grpc),
		Images:       model.DockerProxyPermissionType(record.Images),
		Networks:     model.DockerProxyPermissionType(record.Networks),
		Nodes:        model.DockerProxyPermissionType(record.Nodes),
		Plugins:      model.DockerProxyPermissionType(record.Plugins),
		Services:     model.DockerProxyPermissionType(record.Services),
		Session:      model.DockerProxyPermissionType(record.Session),
		Swarm:        model.DockerProxyPermissionType(record.Swarm),
		System:       model.DockerProxyPermissionType(record.System),
		Tasks:        model.DockerProxyPermissionType(record.Tasks),
		Volumes:      model.DockerProxyPermissionType(record.Volumes),
	}
}

// dockerProxyPermissionInputToDatabaseObject converts DockerProxyPermissionInput to DockerProxyPermissionDatabaseObject
func dockerProxyPermissionInputToDatabaseObject(record *model.DockerProxyPermissionInput) *core.DockerProxyPermission {
	return &core.DockerProxyPermission{
		Ping:         core.DockerProxyPermissionType(record.Ping),
		Version:      core.DockerProxyPermissionType(record.Version),
		Info:         core.DockerProxyPermissionType(record.Info),
		Events:       core.DockerProxyPermissionType(record.Events),
		Auth:         core.DockerProxyPermissionType(record.Auth),
		Secrets:      core.DockerProxyPermissionType(record.Secrets),
		Build:        core.DockerProxyPermissionType(record.Build),
		Commit:       core.DockerProxyPermissionType(record.Commit),
		Configs:      core.DockerProxyPermissionType(record.Configs),
		Containers:   core.DockerProxyPermissionType(record.Containers),
		Distribution: core.DockerProxyPermissionType(record.Distribution),
		Exec:         core.DockerProxyPermissionType(record.Exec),
		Grpc:         core.DockerProxyPermissionType(record.Grpc),
		Images:       core.DockerProxyPermissionType(record.Images),
		Networks:     core.DockerProxyPermissionType(record.Networks),
		Nodes:        core.DockerProxyPermissionType(record.Nodes),
		Plugins:      core.DockerProxyPermissionType(record.Plugins),
		Services:     core.DockerProxyPermissionType(record.Services),
		Session:      core.DockerProxyPermissionType(record.Session),
		Swarm:        core.DockerProxyPermissionType(record.Swarm),
		System:       core.DockerProxyPermissionType(record.System),
		Tasks:        core.DockerProxyPermissionType(record.Tasks),
		Volumes:      core.DockerProxyPermissionType(record.Volumes),
	}
}

// applicationCustomHealthCheckToGraphqlObject converts ApplicationCustomHealthCheck to ApplicationCustomHealthCheckGraphqlObject
func applicationCustomHealthCheckToGraphqlObject(record *core.ApplicationCustomHealthCheck) *model.ApplicationCustomHealthCheck {
	return &model.ApplicationCustomHealthCheck{
		Enabled:              record.Enabled,
		TestCommand:          record.TestCommand,
		IntervalSeconds:      record.IntervalSeconds,
		TimeoutSeconds:       record.TimeoutSeconds,
		StartPeriodSeconds:   record.StartPeriodSeconds,
		StartIntervalSeconds: record.StartIntervalSeconds,
		Retries:              record.Retries,
	}
}

// applicationCustomHealthCheckInputToDatabaseObject converts ApplicationCustomHealthCheckInput to ApplicationCustomHealthCheckDatabaseObject
func applicationCustomHealthCheckInputToDatabaseObject(record *model.ApplicationCustomHealthCheckInput) *core.ApplicationCustomHealthCheck {
	return &core.ApplicationCustomHealthCheck{
		Enabled:              record.Enabled,
		TestCommand:          record.TestCommand,
		IntervalSeconds:      record.IntervalSeconds,
		TimeoutSeconds:       record.TimeoutSeconds,
		StartPeriodSeconds:   record.StartPeriodSeconds,
		StartIntervalSeconds: record.StartIntervalSeconds,
		Retries:              record.Retries,
	}
}

// ingressRuleInputToDatabaseObject converts IngressRuleInput to IngressRuleDatabaseObject
func ingressRuleInputToDatabaseObject(record *model.IngressRuleInput) *core.IngressRule {
	// unset domain id if protocol is tcp or udp
	if record.Protocol == model.ProtocolTypeTCP || record.Protocol == model.ProtocolTypeUDP {
		record.DomainID = nil
	}
	var applicationId *string
	if record.TargetType == model.IngressRuleTargetTypeApplication {
		applicationId = &record.ApplicationID
	}
	return &core.IngressRule{
		TargetType:      core.IngressRuleTargetType(record.TargetType),
		ExternalService: record.ExternalService,
		ApplicationID:   applicationId,
		DomainID:        record.DomainID,
		Protocol:        core.ProtocolType(record.Protocol),
		Port:            record.Port,
		TargetPort:      record.TargetPort,
		HttpsRedirect:   false,
		Authentication: core.IngressRuleAuthentication{
			AuthType: core.IngressRuleNoAuthentication,
		},
		Status:    core.IngressRuleStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ingressRuleValidationInputToDatabaseObject converts IngressRuleValidationInput to IngressRuleDatabaseObject
func ingressRuleValidationInputToDatabaseObject(record *model.IngressRuleValidationInput) *core.IngressRule {
	return &core.IngressRule{
		TargetType:      core.ExternalServiceIngressRule,
		ExternalService: "dummy",
		ApplicationID:   nil,
		DomainID:        record.DomainID,
		Protocol:        core.ProtocolType(record.Protocol),
		Port:            record.Port,
		TargetPort:      0,
		Status:          core.IngressRuleStatusPending,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// ingressRuleToGraphqlObject converts IngressRule to IngressRuleGraphqlObject
func ingressRuleToGraphqlObject(record *core.IngressRule) *model.IngressRule {
	return &model.IngressRule{
		ID:                           record.ID,
		TargetType:                   model.IngressRuleTargetType(record.TargetType),
		ExternalService:              record.ExternalService,
		ApplicationID:                DefaultString(record.ApplicationID, ""),
		DomainID:                     record.DomainID,
		Protocol:                     model.ProtocolType(record.Protocol),
		Port:                         record.Port,
		TargetPort:                   record.TargetPort,
		AuthenticationType:           model.IngressRuleAuthenticationType(record.Authentication.AuthType),
		BasicAuthAccessControlListID: record.Authentication.AppBasicAuthAccessControlListID,
		Status:                       model.IngressRuleStatus(record.Status),
		HTTPSRedirect:                record.HttpsRedirect,
		CreatedAt:                    record.CreatedAt,
		UpdatedAt:                    record.UpdatedAt,
	}
}

// redirectRuleInputToDatabaseObject converts RedirectRuleInput to RedirectRuleDatabaseObject
func redirectRuleInputToDatabaseObject(record *model.RedirectRuleInput) *core.RedirectRule {
	return &core.RedirectRule{
		DomainID:    record.DomainID,
		Protocol:    core.ProtocolType(record.Protocol),
		RedirectURL: record.RedirectURL,
		Status:      core.RedirectRuleStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// redirectRuleToGraphqlObject converts RedirectRule to RedirectRuleGraphqlObject
func redirectRuleToGraphqlObject(record *core.RedirectRule) *model.RedirectRule {
	return &model.RedirectRule{
		ID:          record.ID,
		DomainID:    record.DomainID,
		Protocol:    model.ProtocolType(record.Protocol),
		RedirectURL: record.RedirectURL,
		Status:      model.RedirectRuleStatus(record.Status),
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}
}

// deploymentLogToGraphqlObject converts DeploymentLog to DeploymentLogGraphqlObject
func deploymentLogToGraphqlObject(record *core.DeploymentLog) *model.DeploymentLog {
	return &model.DeploymentLog{
		Content:   record.Content,
		CreatedAt: record.CreatedAt,
	}
}

// userToGraphqlObject converts User to UserGraphqlObject
func userToGraphqlObject(record *core.User) *model.User {
	if record == nil {
		return nil
	}
	return &model.User{
		ID:          record.ID,
		Username:    record.Username,
		TotpEnabled: record.TotpEnabled,
	}
}

// stackToApplicationsInput converts Stack to ApplicationInput
func stackToApplicationsInput(applicationGroupID *string, record *stack_parser.Stack, db gorm.DB) ([]model.ApplicationInput, error) {
	applications := make([]model.ApplicationInput, 0)
	for serviceName, service := range record.Services {
		environmentVariables := make([]*model.EnvironmentVariableInput, 0)
		for key, value := range service.Environment {
			environmentVariables = append(environmentVariables, &model.EnvironmentVariableInput{
				Key:   key,
				Value: value,
			})
		}
		persistentVolumeBindings := make([]*model.PersistentVolumeBindingInput, 0)
		for _, volume := range service.Volumes {
			// fetch volume from database
			pv := core.PersistentVolume{}
			err := pv.FindByName(context.Background(), db, volume.Name)
			if err != nil {
				return nil, err
			}
			persistentVolumeBindings = append(persistentVolumeBindings, &model.PersistentVolumeBindingInput{
				PersistentVolumeID: pv.ID,
				MountingPath:       volume.MountingPoint,
			})
		}
		sysctls := make([]string, 0)
		for key, val := range service.Sysctls {
			sysctls = append(sysctls, fmt.Sprintf("%s=%s", key, val))
		}
		configs := make([]*model.ConfigMountInput, 0)
		for _, config := range service.Configs {
			configs = append(configs, &model.ConfigMountInput{
				Content:      config.Content,
				Gid:          config.Gid,
				UID:          config.Uid,
				MountingPath: config.MountingPath,
			})
		}
		image := service.Image
		replicas := service.Deploy.Replicas
		command := ""
		if service.Command != nil {
			command = service.Command.String()
		}
		// docker proxy config

		app := model.ApplicationInput{
			Name:                     serviceName,
			EnvironmentVariables:     environmentVariables,
			PersistentVolumeBindings: persistentVolumeBindings,
			ConfigMounts:             configs,
			Capabilities:             service.CapAdd,
			Sysctls:                  sysctls,
			Dockerfile:               nil,
			BuildArgs:                []*model.BuildArgInput{},
			DeploymentMode:           model.DeploymentMode(service.Deploy.Mode),
			Replicas:                 &replicas,
			ResourceLimit: &model.ResourceLimitInput{
				MemoryMb: service.Deploy.Resources.Limits.MemoryMB,
			},
			ReservedResource: &model.ReservedResourceInput{
				MemoryMb: service.Deploy.Resources.Reservations.MemoryMB,
			},
			UpstreamType:                 model.UpstreamTypeImage,
			DockerImage:                  &image,
			ImageRegistryCredentialID:    nil,
			GitCredentialID:              nil,
			RepositoryURL:                nil,
			RepositoryBranch:             nil,
			CodePath:                     nil,
			SourceCodeCompressedFileName: nil,
			ApplicationGroupID:           applicationGroupID,
			Command:                      command,
			CustomHealthCheck: &model.ApplicationCustomHealthCheckInput{
				Enabled:              service.CustomHealthCheck.Enabled,
				TestCommand:          service.CustomHealthCheck.TestCommand,
				IntervalSeconds:      service.CustomHealthCheck.IntervalSeconds,
				TimeoutSeconds:       service.CustomHealthCheck.TimeoutSeconds,
				StartPeriodSeconds:   service.CustomHealthCheck.StartPeriodSeconds,
				StartIntervalSeconds: service.CustomHealthCheck.StartIntervalSeconds,
				Retries:              service.CustomHealthCheck.Retries,
			},
			PreferredServerHostnames: service.PreferredServerHostnames,
			DockerProxyConfig: &model.DockerProxyConfigInput{
				Enabled: service.DockerProxyConfig.Enabled,
				Permission: &model.DockerProxyPermissionInput{
					Ping:         model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Ping),
					Version:      model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Version),
					Info:         model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Info),
					Events:       model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Events),
					Auth:         model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Auth),
					Secrets:      model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Secrets),
					Build:        model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Build),
					Commit:       model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Commit),
					Configs:      model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Configs),
					Containers:   model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Containers),
					Distribution: model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Distribution),
					Exec:         model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Exec),
					Grpc:         model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Grpc),
					Images:       model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Images),
					Networks:     model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Networks),
					Nodes:        model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Nodes),
					Plugins:      model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Plugins),
					Services:     model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Services),
					Session:      model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Session),
					Swarm:        model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Swarm),
					System:       model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.System),
					Tasks:        model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Tasks),
					Volumes:      model.DockerProxyPermissionType(service.DockerProxyConfig.Permission.Volumes),
				},
			},
		}
		applications = append(applications, app)
	}

	return applications, nil
}

// newServerInputToDatabaseObject converts NewServerInput to ServerDatabaseObject
func newServerInputToDatabaseObject(record *model.NewServerInput) *core.Server {
	return &core.Server{
		IP:                   record.IP,
		SSHPort:              record.SSHPort,
		HostName:             "",
		User:                 record.User,
		ScheduleDeployments:  false,
		MaintenanceMode:      false,
		DockerUnixSocketPath: "",
		SwarmMode:            core.SwarmMode(model.SwarmModeWorker),
		ProxyConfig: core.ProxyConfig{
			Enabled: false,
			Type:    core.ProxyType(model.ProxyTypeActive),
		},
		Status: core.ServerStatus(model.ServerStatusNeedsSetup),
	}
}

// serverToGraphqlObject converts Server to ServerGraphqlObject
func serverToGraphqlObject(record *core.Server) *model.Server {
	return &model.Server{
		ID:                   record.ID,
		IP:                   record.IP,
		SSHPort:              record.SSHPort,
		Hostname:             record.HostName,
		User:                 record.User,
		ScheduleDeployments:  record.ScheduleDeployments,
		MaintenanceMode:      record.MaintenanceMode,
		DockerUnixSocketPath: record.DockerUnixSocketPath,
		SwarmMode:            model.SwarmMode(record.SwarmMode),
		ProxyType:            model.ProxyType(record.ProxyConfig.Type),
		ProxyEnabled:         record.ProxyConfig.Enabled,
		Status:               model.ServerStatus(record.Status),
	}
}

// serverLogToGraphqlObject converts ServerLog to ServerLogGraphqlObject
func serverLogToGraphqlObject(record *core.ServerLog) *model.ServerLog {
	return &model.ServerLog{
		ID:        record.ID,
		Title:     record.Title,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}

// serverResourceStatToGraphqlObject converts ServerResourceStat to ServerResourceStatGraphqlObject
func serverResourceStatToGraphqlObject(record *core.ServerResourceStat) *model.ServerResourceAnalytics {
	return &model.ServerResourceAnalytics{
		CPUUsagePercent: int(record.CpuUsagePercent),
		MemoryTotalGb:   float64(record.MemStat.TotalGB),
		MemoryUsedGb:    float64(record.MemStat.UsedGB),
		MemoryCachedGb:  float64(record.MemStat.CachedGB),
		NetworkSentKb:   record.NetStat.SentKB,
		NetworkRecvKb:   record.NetStat.RecvKB,
		NetworkSentKbps: record.NetStat.SentKBPS,
		NetworkRecvKbps: record.NetStat.RecvKBPS,
		Timestamp:       record.RecordedAt,
	}
}

// serverDiskStatToGraphqlObject converts ServerDiskStat to ServerDiskStatGraphqlObject
func serverDiskStatToGraphqlObject(record core.ServerDiskStat, timestamp time.Time) *model.ServerDiskUsage {
	return &model.ServerDiskUsage{
		Path:       record.Path,
		MountPoint: record.MountPoint,
		TotalGb:    float64(record.TotalGB),
		UsedGb:     float64(record.UsedGB),
		Timestamp:  timestamp,
	}
}

// severDisksStatToGraphqlObject converts ServerDiskStat to ServerDiskStatGraphqlObject
func severDisksStatToGraphqlObject(records core.ServerDiskStats, timestamp time.Time) model.ServerDisksUsage {
	disks := make([]*model.ServerDiskUsage, 0)
	for _, disk := range records {
		disks = append(disks, serverDiskStatToGraphqlObject(disk, timestamp))
	}
	return model.ServerDisksUsage{
		Disks:     disks,
		Timestamp: timestamp,
	}
}

// applicationServiceResourceStatToGraphqlObject converts ApplicationServiceResourceStat to ApplicationServiceResourceStatGraphqlObject
func applicationServiceResourceStatToGraphqlObject(record *core.ApplicationServiceResourceStat) *model.ApplicationResourceAnalytics {
	return &model.ApplicationResourceAnalytics{
		CPUUsagePercent:      int(record.CpuUsagePercent),
		ServiceCPUTime:       record.ServiceCpuTime,
		SystemCPUTime:        record.SystemCpuTime,
		MemoryUsedMb:         record.UsedMemoryMB,
		NetworkRecvKb:        record.NetStat.RecvKB,
		NetworkSentKb:        record.NetStat.SentKB,
		NetworkRecvKbps:      record.NetStat.RecvKBPS,
		NetworkSentKbps:      record.NetStat.SentKBPS,
		ReportingServerCount: int(record.ReportingServerCount),
		Timestamp:            record.RecordedAt,
	}
}

// appBasicAuthAccessControlListToGraphqlObject converts AppBasicAuthAccessControlList to AppBasicAuthAccessControlListGraphqlObject
func appBasicAuthAccessControlListToGraphqlObject(record *core.AppBasicAuthAccessControlList) *model.AppBasicAuthAccessControlList {
	return &model.AppBasicAuthAccessControlList{
		ID:            record.ID,
		Name:          record.Name,
		GeneratedName: record.GeneratedName,
	}
}

// appBasicAuthAccessControlUserToGraphqlObject converts AppBasicAuthAccessControlUser to AppBasicAuthAccessControlUserGraphqlObject
func appBasicAuthAccessControlUserToGraphqlObject(record *core.AppBasicAuthAccessControlUser) *model.AppBasicAuthAccessControlUser {
	return &model.AppBasicAuthAccessControlUser{
		ID:       record.ID,
		Username: record.Username,
	}
}

// appBasicAuthAccessControlListInputToDatabaseObject converts AppBasicAuthAccessControlListInput to AppBasicAuthAccessControlListDatabaseObject
func appBasicAuthAccessControlListInputToDatabaseObject(record *model.AppBasicAuthAccessControlListInput) *core.AppBasicAuthAccessControlList {
	return &core.AppBasicAuthAccessControlList{
		Name: record.Name,
	}
}

// appBasicAuthAccessControlUserInputToDatabaseObject converts AppBasicAuthAccessControlUserInput to AppBasicAuthAccessControlUserDatabaseObject
func appBasicAuthAccessControlUserInputToDatabaseObject(record *model.AppBasicAuthAccessControlUserInput) *core.AppBasicAuthAccessControlUser {
	return &core.AppBasicAuthAccessControlUser{
		Username:                        record.Username,
		PlainTextPassword:               record.Password,
		AppBasicAuthAccessControlListID: record.AppBasicAuthAccessControlListID,
	}
}
