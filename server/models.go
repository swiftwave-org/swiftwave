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
	ID              uint                  `json:"id" gorm:"primaryKey"`
	Type            ApplicationSourceType `json:"type"`
	GitCredential   GitCredential         `json:"git_credential"`
	GitCredentialID uint                  `json:"git_credential_id"`
	GitProvider     string                `json:"git_provider"`
	RepositoryUsername string             `json:"repository_username"`
	RepositoryName string                 `json:"repository_name"`
	Branch          string                `json:"branch"`
	LastCommit      string                `json:"last_commit"`
	TarballPath     string                `json:"tarball_path"`
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
)

// Application
type Application struct {
	ID           uint              `json:"id" gorm:"primaryKey"`
	Source       ApplicationSource `json:"source"`
	SourceID     uint              `json:"source_id"`
	Image        string            `json:"image"`
	BuildArgs    string            `json:"build_args"`
	EnvVariables string            `json:"env_variables"`
	DockerConfig string            `json:"docker_config"`
	VolumeMounts string            `json:"volume_mounts"`
}


// Migrate database
func (server *Server) MigrateDatabaseTables() {
	server.DB_CLIENT.AutoMigrate(&Domain{})
	server.DB_CLIENT.AutoMigrate(&GitCredential{})
	server.DB_CLIENT.AutoMigrate(&ApplicationSource{})
	server.DB_CLIENT.AutoMigrate(&Application{})
}