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

// Application build logs
type ApplicationBuildLog struct {
	ID            string      `json:"id" gorm:"primaryKey"`
	ApplicationID uint        `json:"application_id"`
	Application   Application `json:"-"`
	Logs          string      `json:"-"`
	Time          time.Time   `json:"time"`
}

// Migrate database
func (server *Server) MigrateDatabaseTables() {
	server.DB_CLIENT.AutoMigrate(&Domain{})
	server.DB_CLIENT.AutoMigrate(&GitCredential{})
	server.DB_CLIENT.AutoMigrate(&ApplicationSource{})
	server.DB_CLIENT.AutoMigrate(&Application{})
	server.DB_CLIENT.AutoMigrate(&ApplicationBuildLog{})
}
