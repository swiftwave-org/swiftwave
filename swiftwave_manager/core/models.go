package core

import (
	"time"
)

// GitCredential : credential for git client
type GitCredential struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name"`
	Username    string       `json:"username"`
	Password    string       `json:"password"`
	Deployments []Deployment `json:"deployments" gorm:"foreignKey:GitCredentialID"`
}

// ImageRegistryCredential : credential for docker image registry
type ImageRegistryCredential struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Url         string       `json:"url"`
	Username    string       `json:"username"`
	Password    string       `json:"password"`
	Deployments []Deployment `json:"deployments" gorm:"foreignKey:ImageRegistryCredentialID"`
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
	IngressRules  []IngressRule   `json:"ingressRules" gorm:"foreignKey:DomainID"`
	RedirectRules []RedirectRule  `json:"redirectRules" gorm:"foreignKey:DomainID"`
}

// IngressRule : hold information about Ingress rule for service
type IngressRule struct {
	ID            uint              `json:"id" gorm:"primaryKey"`
	DomainID      uint              `json:"domainID"`
	ApplicationID string            `json:"applicationID"`
	Protocol      ProtocolType      `json:"protocol"`
	Port          uint              `json:"port"`
	TargetPort    uint              `json:"targetPort"`
	Status        IngressRuleStatus `json:"status"`
	CreatedAt     time.Time         `json:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt"`
}

// RedirectRule : hold information about Redirect rules for domain
type RedirectRule struct {
	ID          uint               `json:"id" gorm:"primaryKey"`
	DomainID    uint               `json:"domainID"`
	Port        uint               `json:"port"`
	RedirectURL string             `json:"redirectURL"`
	Status      RedirectRuleStatus `json:"status"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
}

// PersistentVolume : hold information about persistent volume
type PersistentVolume struct {
	ID                       uint                      `json:"id" gorm:"primaryKey"`
	Name                     string                    `json:"name" gorm:"unique"`
	PersistentVolumeBindings []PersistentVolumeBinding `json:"persistentVolumeBindings" gorm:"foreignKey:PersistentVolumeID"`
}

// PersistentVolumeBinding : hold information about persistent volume binding
type PersistentVolumeBinding struct {
	ID                 uint   `json:"id" gorm:"primaryKey"`
	ApplicationID      uint   `json:"applicationID"`
	PersistentVolumeID uint   `json:"persistentVolumeID"`
	MountingPath       string `json:"mountingPath"`
}

// EnvironmentVariable : hold information about environment variable
type EnvironmentVariable struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	ApplicationID uint   `json:"applicationID"`
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
	EnvironmentVariables []EnvironmentVariable `json:"environmentVariables" gorm:"foreignKey:ApplicationID"`
	// Persistent Volumes
	// On change of persistent volumes, deployment will be triggered by force update
	PersistentVolumeBindings []PersistentVolumeBinding `json:"persistentVolumeBindings" gorm:"foreignKey:ApplicationID"`
	// No of replicas to be deployed
	DeploymentMode DeploymentMode `json:"deploymentMode"`
	Replicas       uint           `json:"replicas"`
	// Deployments
	Deployments []Deployment `json:"deployments" gorm:"foreignKey:ApplicationID"`
	// Latest Deployment
	LatestDeployment Deployment `json:"-"`
	// Ingress Rules
	IngressRules []IngressRule `json:"ingressRules" gorm:"foreignKey:ApplicationID"`
}

// Deployment : hold information about deployment of application
type Deployment struct {
	ID            string       `json:"id" gorm:"primaryKey"`
	ApplicationID uint         `json:"applicationID"`
	UpstreamType  UpstreamType `json:"upstreamType"`
	// Fields for UpstreamType = Git
	GitCredentialID  uint        `json:"gitCredentialID"`
	GitProvider      GitProvider `json:"gitProvider"`
	RepositoryOwner  string      `json:"repositoryOwner"`
	RepositoryName   string      `json:"repositoryName"`
	RepositoryBranch string      `json:"repositoryBranch"`
	CommitHash       string      `json:"commitHash"`
	// Fields for UpstreamType = SourceCode
	SourceCodeCompressedFileName string `json:"sourceCodeCompressedFileName"`
	// Fields for UpstreamType = Image
	DockerImage               string `json:"dockerImage"`
	ImageRegistryCredentialID uint   `json:"imageRegistryCredentialID"`
	// Common Fields
	BuildArgs  []BuildArg `json:"buildArgs" gorm:"foreignKey:DeploymentID"`
	Dockerfile string     `json:"dockerfile"`
	// Logs
	Logs []DeploymentLog `json:"logs" gorm:"foreignKey:DeploymentID"`
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
