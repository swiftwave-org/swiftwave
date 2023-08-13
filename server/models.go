package server

import "time"

// Domains
type Domain struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	Name          string          `json:"name"`
	SSLStatus     DomainSSLStatus `json:"ssl_status"`
	SSLPrivateKey string          `json:"ssl_private_key"`
	SSLFullChain  string          `json:"ssl_full_chain"`
	SSLIssuedAt   time.Time       `json:"ssl_issued_at"`
	SSLIssuer     string          `json:"ssl_issuer"`
}

type DomainSSLStatus string

const (
	DomainSSLStatusNone    DomainSSLStatus = "none"
	DomainSSLStatusIssued  DomainSSLStatus = "issued"
	DomainSSLStatusIssuing DomainSSLStatus = "issuing"
)

// Git Credentials
type GitCredential struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Application Sources
type ApplicationSource struct {
	ID                 uint                  `json:"id" gorm:"primaryKey"`
	Type               ApplicationSourceType `json:"type"`
	GitCredential      GitCredential         `json:"git_credential"`
	GitCredentialID    uint                  `json:"git_credential_id"`
	GitProvider        GitProvider           `json:"git_provider"`
	RepositoryUsername string                `json:"repository_username"`
	RepositoryName     string                `json:"repository_name"`
	Branch             string                `json:"branch"`
	LastCommit         string                `json:"last_commit"`
	TarballFile        string                `json:"tarball_file"`
	DockerImage        string                `json:"docker_image"`
}

type GitProvider string

const (
	GitProviderGithub GitProvider = "github"
	GitProviderGitlab GitProvider = "gitlab"
)

type ApplicationSourceType string

const (
	ApplicationSourceTypeGit     ApplicationSourceType = "git"
	ApplicationSourceTypeTarball ApplicationSourceType = "tarball"
	ApplicationSourceTypeImage   ApplicationSourceType = "image"
)

// Application
type Application struct {
	ID                   uint              `json:"id" gorm:"primaryKey"`
	ServiceName          string            `json:"service_name" gorm:"unique"`
	Source               ApplicationSource `json:"source"`
	SourceID             uint              `json:"source_id"`
	Image                string            `json:"image"`
	BuildArgs            string            `json:"build_args" validate:"required"`
	EnvironmentVariables string            `json:"environment_variables" validate:"required"`
	Volumes              string            `json:"volumes" validate:"required"`
	Dockerfile           string            `json:"dockerfile" validate:"required"`
	Replicas             uint              `json:"replicas"`
	Status               ApplicationStatus `json:"status"`
}

type ApplicationStatus string

const (
	ApplicationStatusPending                ApplicationStatus = "pending"
	ApplicationStatusBuildingImage          ApplicationStatus = "building_image"
	ApplicationStatusBuildingImageQueued    ApplicationStatus = "building_image_queued"
	ApplicationStatusBuildingImageCompleted ApplicationStatus = "building_image_completed"
	ApplicationStatusBuildingImageFailed    ApplicationStatus = "building_image_failed"
	ApplicationStatusDeployingPending       ApplicationStatus = "deploying_pending"
	ApplicationStatusDeploying              ApplicationStatus = "deploying"
	ApplicationStatusDeployingQueued        ApplicationStatus = "deploying_queued"
	ApplicationStatusDeployingFailed        ApplicationStatus = "deploying_failed"
	ApplicationStatusRunning                ApplicationStatus = "running"
	ApplicationStatusStopped                ApplicationStatus = "stopped"
	ApplicationStatusFailed                 ApplicationStatus = "failed"
	ApplicationStatusRedeployPending        ApplicationStatus = "redeploy_pending"
)

// Application build logs
type ApplicationBuildLog struct {
	ID            string      `json:"id" gorm:"primaryKey"`
	ApplicationID uint        `json:"application_id"`
	Application   Application `json:"-"`
	Logs          string      `json:"-"`
	Time          time.Time   `json:"time"`
}

// Ingress Rules
type IngressRule struct {
	ID          uint              `json:"id" gorm:"primaryKey"`
	Protocol    ProtocolType      `json:"protocol" validate:"required"`
	DomainName  string            `json:"domain_name"` // Ignored if protocol is TCP
	Port        uint              `json:"port" validate:"required"`
	ServiceName string            `json:"service_name" validate:"required"`
	ServicePort uint              `json:"service_port" validate:"required"`
	Status      IngressRuleStatus `json:"status"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type ProtocolType string

const (
	HTTPProtcol  ProtocolType = "http"
	HTTPSProtcol ProtocolType = "https"
	TCPProtcol   ProtocolType = "tcp"
)

type IngressRuleStatus string

const (
	IngressRuleStatusPending       IngressRuleStatus = "pending"
	IngressRuleStatusApplied       IngressRuleStatus = "applied"
	IngressRuleStatusFailed        IngressRuleStatus = "failed"
	IngressRuleStatusDeletePending IngressRuleStatus = "delete_pending"
)

// Redirect Rules -- only for HTTP
type RedirectRule struct {
	ID          uint               `json:"id" gorm:"primaryKey"`
	Port        uint               `json:"port" validate:"required"`
	DomainName  string             `json:"domain_name" validate:"required"`
	RedirectURL string             `json:"redirect_url" validate:"required"`
	Status      RedirectRuleStatus `json:"status"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type RedirectRuleStatus string

const (
	RedirectRuleStatusPending       RedirectRuleStatus = "pending"
	RedirectRuleStatusApplied       RedirectRuleStatus = "applied"
	RedirectRuleStatusDeletePending RedirectRuleStatus = "delete_pending"
)

// Migrate database
func (server *Server) MigrateDatabaseTables() {
	server.DB_CLIENT.AutoMigrate(&Domain{})
	server.DB_CLIENT.AutoMigrate(&GitCredential{})
	server.DB_CLIENT.AutoMigrate(&ApplicationSource{})
	server.DB_CLIENT.AutoMigrate(&Application{})
	server.DB_CLIENT.AutoMigrate(&ApplicationBuildLog{})
	server.DB_CLIENT.AutoMigrate(&IngressRule{})
	server.DB_CLIENT.AutoMigrate(&RedirectRule{})
}

// Application deploy request
type ApplicationDeployRequest struct {
	ServiceName           string                `json:"service_name" validate:"required"`
	ApplicationSourceType ApplicationSourceType `json:"source_type" validate:"required"`
	GitCredentialID       uint                  `json:"git_credential_id"`
	RepositoryURL         string                `json:"repository_url"`
	Branch                string                `json:"branch"`
	TarballFile           string                `json:"tarball_file"`
	Dockerfile            string                `json:"dockerfile" validate:"required"`
	EnvironmentVariables  map[string]string     `json:"environment_variables" validate:"required"`
	BuildArgs             map[string]string     `json:"build_args" validate:"required"`
	Volumes               map[string]string     `json:"volumes" validate:"required"`
	DockerImage           string                `json:"docker_image"`
	Replicas              uint                  `json:"replicas"`
}

// Application deploy update request
type ApplicationDeployUpdateRequest struct {
	Source               ApplicationSource `json:"source"`
	BuildArgs            map[string]string `json:"build_args" validate:"required"`
	EnvironmentVariables map[string]string `json:"environment_variables" validate:"required"`
	Dockerfile           string            `json:"dockerfile" validate:"required"`
	Replicas             uint              `json:"replicas"`
}

// Application summary
type ApplicationSummary struct {
	ID          uint              `json:"id"`
	ServiceName string            `json:"service_name"`
	Source      string            `json:"source"`
	Replicas    uint              `json:"replicas"`
	Status      ApplicationStatus `json:"status"`
}
