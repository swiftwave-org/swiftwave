package core

import (
	"github.com/lib/pq"
	"time"
)

// User : hold information about user
type User struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	Username     string `json:"username" gorm:"unique"`
	PasswordHash string `json:"passwordHash"`
}

// GitCredential : credential for git client
type GitCredential struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name"`
	Username    string       `json:"username"`
	Password    string       `json:"password"`
	Deployments []Deployment `json:"deployments" gorm:"foreignKey:GitCredentialID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" `
}

// ImageRegistryCredential : credential for docker image registry
type ImageRegistryCredential struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Url         string       `json:"url"`
	Username    string       `json:"username"`
	Password    string       `json:"password"`
	Deployments []Deployment `json:"deployments" gorm:"foreignKey:ImageRegistryCredentialID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// Domain : hold information about domain
type Domain struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	Name          string          `json:"name" gorm:"unique"`
	SSLStatus     DomainSSLStatus `json:"sslStatus"`
	SSLPrivateKey string          `json:"sslPrivateKey"`
	SSLFullChain  string          `json:"sslFullChain"`
	SSLIssuedAt   time.Time       `json:"sslIssuedAt"`
	SSLIssuer     string          `json:"sslIssuer"`
	SSLAutoRenew  bool            `json:"sslAutoRenew" gorm:"default:false"`
	IngressRules  []IngressRule   `json:"ingressRules" gorm:"foreignKey:DomainID"`
	RedirectRules []RedirectRule  `json:"redirectRules" gorm:"foreignKey:DomainID"`
}

// IngressRule : hold information about Ingress rule for service
type IngressRule struct {
	ID            uint              `json:"id" gorm:"primaryKey"`
	DomainID      *uint             `json:"domainID,omitempty" gorm:"default:null"`
	ApplicationID string            `json:"applicationID"`
	Protocol      ProtocolType      `json:"protocol"`
	Port          uint              `json:"port"`       // external port
	TargetPort    uint              `json:"targetPort"` // port of the application
	Status        IngressRuleStatus `json:"status"`
	CreatedAt     time.Time         `json:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt"`
}

// RedirectRule : hold information about Redirect rules for domain
type RedirectRule struct {
	ID          uint               `json:"id" gorm:"primaryKey"`
	DomainID    uint               `json:"domainID"`
	Protocol    ProtocolType       `json:"protocol"`
	RedirectURL string             `json:"redirectURL"`
	Status      RedirectRuleStatus `json:"status"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
}

// PersistentVolume : hold information about persistent volume
type PersistentVolume struct {
	ID                       uint                      `json:"id" gorm:"primaryKey"`
	Name                     string                    `json:"name" gorm:"unique"`
	Type                     PersistentVolumeType      `json:"type" gorm:"default:'local'"`
	NFSConfig                NFSConfig                 `json:"nfsConfig" gorm:"embedded;embeddedPrefix:nfs_config_"`
	PersistentVolumeBindings []PersistentVolumeBinding `json:"persistentVolumeBindings" gorm:"foreignKey:PersistentVolumeID"`
	PersistentVolumeBackups  []PersistentVolumeBackup  `json:"persistentVolumeBackups" gorm:"foreignKey:PersistentVolumeID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	PersistentVolumeRestores []PersistentVolumeRestore `json:"persistentVolumeRestores" gorm:"foreignKey:PersistentVolumeID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// PersistentVolumeBinding : hold information about persistent volume binding
type PersistentVolumeBinding struct {
	ID                 uint   `json:"id" gorm:"primaryKey"`
	ApplicationID      string `json:"applicationID"`
	PersistentVolumeID uint   `json:"persistentVolumeID"`
	MountingPath       string `json:"mountingPath"`
}

// PersistentVolumeBackup : hold information about persistent volume backup
type PersistentVolumeBackup struct {
	ID                 uint         `json:"id" gorm:"primaryKey"`
	Type               BackupType   `json:"type"`
	Status             BackupStatus `json:"status"`
	File               string       `json:"file"`
	FileSizeMB         float64      `json:"fileSizeMB"`
	PersistentVolumeID uint         `json:"persistentVolumeID"`
	CreatedAt          time.Time    `json:"createdAt"`
	CompletedAt        time.Time    `json:"completedAt"`
}

// PersistentVolumeRestore : hold information about persistent volume restore
type PersistentVolumeRestore struct {
	ID                 uint          `json:"id" gorm:"primaryKey"`
	Type               RestoreType   `json:"type"`
	File               string        `json:"file"`
	Status             RestoreStatus `json:"status"`
	PersistentVolumeID uint          `json:"persistentVolumeID"`
	CreatedAt          time.Time     `json:"uploadedAt"`
	CompletedAt        time.Time     `json:"completedAt"`
}

// EnvironmentVariable : hold information about environment variable
type EnvironmentVariable struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	ApplicationID string `json:"applicationID"`
	Key           string `json:"key"`
	Value         string `json:"value"`
}

// BuildArg : hold information about build args
type BuildArg struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	DeploymentID string `json:"deploymentID"`
	Key          string `json:"key"`
	Value        string `json:"value"`
}

// Application : hold information about application
type Application struct {
	ID   string `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"unique"`
	// Environment Variables
	// On change of environment variables, deployment will be triggered by force update
	EnvironmentVariables []EnvironmentVariable `json:"environmentVariables" gorm:"foreignKey:ApplicationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// Persistent Volumes
	// On change of persistent volumes, deployment will be triggered by force update
	PersistentVolumeBindings []PersistentVolumeBinding `json:"persistentVolumeBindings" gorm:"foreignKey:ApplicationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// No of replicas to be deployed
	DeploymentMode DeploymentMode `json:"deploymentMode"`
	Replicas       uint           `json:"replicas"`
	// Deployments
	Deployments []Deployment `json:"deployments" gorm:"foreignKey:ApplicationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// Latest Deployment
	LatestDeployment Deployment `json:"-"`
	// Ingress Rules
	IngressRules []IngressRule `json:"ingressRules" gorm:"foreignKey:ApplicationID"`
	// Capabilities
	Capabilities pq.StringArray `json:"capabilities" gorm:"type:text[]"`
	// Sysctls
	Sysctls pq.StringArray `json:"sysctls" gorm:"type:text[]"`
	// Is deleted - soft delete - will be deleted from database in background
	IsDeleted bool `json:"isDeleted" gorm:"default:false"`
	// Webhook token
	WebhookToken string `json:"webhookToken"`
	// Sleeping
	IsSleeping bool `json:"isSleeping" gorm:"default:false"`
}

// Deployment : hold information about deployment of application
type Deployment struct {
	ID            string       `json:"id" gorm:"primaryKey"`
	ApplicationID string       `json:"applicationID"`
	UpstreamType  UpstreamType `json:"upstreamType"`
	// Fields for UpstreamType = Git
	GitCredentialID  *uint       `json:"gitCredentialID"`
	GitProvider      GitProvider `json:"gitProvider"`
	RepositoryOwner  string      `json:"repositoryOwner"`
	RepositoryName   string      `json:"repositoryName"`
	RepositoryBranch string      `json:"repositoryBranch"`
	CodePath         string      `json:"codePath"`
	CommitHash       string      `json:"commitHash"`
	// Fields for UpstreamType = SourceCode
	SourceCodeCompressedFileName string `json:"sourceCodeCompressedFileName"`
	// Fields for UpstreamType = Image
	DockerImage               string `json:"dockerImage"`
	ImageRegistryCredentialID *uint  `json:"imageRegistryCredentialID"`
	// Common Fields
	BuildArgs  []BuildArg `json:"buildArgs" gorm:"foreignKey:DeploymentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Dockerfile string     `json:"dockerfile"`
	// Logs
	Logs []DeploymentLog `json:"logs" gorm:"foreignKey:DeploymentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// Deployment Status
	Status DeploymentStatus `json:"status"`
	// Created At
	CreatedAt time.Time `json:"createdAt"`
}

// DeploymentLog : hold logs of deployment
type DeploymentLog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	DeploymentID string    `json:"deploymentID"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"createdAt"`
}
