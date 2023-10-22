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
	SSLStatus     DomainSSLStatus `json:"ssl_status"`
	SSLPrivateKey string          `json:"ssl_private_key"`
	SSLFullChain  string          `json:"ssl_full_chain"`
	SSLIssuedAt   time.Time       `json:"ssl_issued_at"`
	SSLIssuer     string          `json:"ssl_issuer"`
	IngressRules  []IngressRule   `json:"ingress_rules" gorm:"foreignKey:DomainID"`
	RedirectRules []RedirectRule  `json:"redirect_rules" gorm:"foreignKey:DomainID"`
}

// IngressRule : hold information about Ingress rule for service
type IngressRule struct {
	ID            uint              `json:"id" gorm:"primaryKey"`
	DomainID      uint              `json:"domain_id"`
	ApplicationID string            `json:"application_id"`
	Protocol      ProtocolType      `json:"protocol"`
	Port          uint              `json:"port"`
	TargetPort    uint              `json:"target_port"`
	Status        IngressRuleStatus `json:"status"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// RedirectRule : hold information about Redirect rules for domain
type RedirectRule struct {
	ID          uint               `json:"id" gorm:"primaryKey"`
	DomainID    uint               `json:"domain_id"`
	Port        uint               `json:"port"`
	RedirectURL string             `json:"redirect_url"`
	Status      RedirectRuleStatus `json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// Application : hold information about application
type Application struct {
	ID                   string `json:"id" gorm:"primaryKey"`
	ServiceName          string `json:"service_name" gorm:"unique"`
	EnvironmentVariables string `json:"environment_variables"` // JSON string
	Volumes              string `json:"volumes"`               // JSON string
	// Deployment
	Deployments []Deployment `json:"deployments" gorm:"foreignKey:ApplicationID"`
	// Ingress Rules
	IngressRules []IngressRule `json:"ingress_rules" gorm:"foreignKey:ApplicationID"`
}

// Deployment : hold information about deployment of application
type Deployment struct {
	ID            string       `json:"id" gorm:"primaryKey"`
	ApplicationID uint         `json:"application_id"`
	UpstreamType  UpstreamType `json:"upstream_type"`
	// Fields for UpstreamType = Git
	GitCredentialID  uint        `json:"git_credential_id"`
	GitProvider      GitProvider `json:"git_provider"`
	RepositoryOwner  string      `json:"repository_owner"`
	RepositoryName   string      `json:"repository_name"`
	RepositoryBranch string      `json:"repository_branch"`
	CommitHash       string      `json:"commit_hash"`
	// Fields for UpstreamType = SourceCode
	SourceCodeCompressedFileName string `json:"source_code_compressed_file_name"`
	// Fields for UpstreamType = Image
	DockerImage               string `json:"docker_image"`
	ImageRegistryCredentialID uint   `json:"image_registry_credential_id"`
	// Common Fields
	BuildArgs  string `json:"build_args"` // JSON string
	Dockerfile string `json:"dockerfile"`
	// No of replicas to be deployed
	DeploymentMode DeploymentMode `json:"deployment_mode"`
	Replicas       uint           `json:"replicas"`
	// Logs
	Logs []DeploymentLog `json:"logs" gorm:"foreignKey:DeploymentID"`
	// Deployment Status
	Status DeploymentStatus `json:"status"`
	// Created At
	CreatedAt time.Time `json:"created_at"`
}

// DeploymentLog : hold logs of deployment
type DeploymentLog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	DeploymentID string    `json:"deployment_id"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at"`
}
