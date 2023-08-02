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
	RepositoryURL   string                `json:"repository_url"`
	RepositoryName  string                `json:"repository_name"`
	Branch          string                `json:"branch"`
	LastCommit      string                `json:"last_commit"`
	TarballPath     string                `json:"tarball_path"`
}

type ApplicationSourceType string

const (
	ApplicationSourceTypeGit     ApplicationSourceType = "git"
	ApplicationSourceTypeTarball ApplicationSourceType = "tarball"
)

// Application
type Application struct {
	ID           uint              `json:"id" gorm:"primaryKey"`
	Source       ApplicationSource `json:"source"`
	Image        string            `json:"image"`
	BuildArgs    string            `json:"build_args"`
	EnvVariables string            `json:"env_variables"`
	DockerConfig string            `json:"docker_config"`
	VolumeMounts string            `json:"volume_mounts"`
}
