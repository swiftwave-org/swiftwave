package server

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	DOCKER "github.com/swiftwave-org/swiftwave/container_manager"
	DOCKER_CONFIG_GENERATOR "github.com/swiftwave-org/swiftwave/docker_config_generator"
	HAPROXY "github.com/swiftwave-org/swiftwave/haproxy_manager"
	SSL "github.com/swiftwave-org/swiftwave/ssl_manager"

	DOCKER_CLIENT "github.com/docker/docker/client"

	"github.com/go-redis/redis/v8"
	"github.com/vmihailenco/taskq/v3"
	"gorm.io/gorm"
)

// Server struct
type Server struct {
	SSL_MANAGER                  SSL.Manager
	HAPROXY_MANAGER              HAPROXY.Manager
	DOCKER_MANAGER               DOCKER.Manager
	DOCKER_CONFIG_GENERATOR      DOCKER_CONFIG_GENERATOR.Manager
	DOCKER_CLIENT                DOCKER_CLIENT.Client
	DB_CLIENT                    gorm.DB
	REDIS_CLIENT                 redis.Client
	ECHO_SERVER                  echo.Echo
	PORT                         int
	HAPROXY_SERVICE              string
	CODE_TARBALL_DIR             string
	SWARM_NETWORK                string
	RESTRICTED_PORTS             []int
	SESSION_TOKENS               map[string]time.Time
	SESSION_TOKEN_EXPIRY_MINUTES int
	// Worker related
	QUEUE_FACTORY         taskq.Factory
	TASK_QUEUE            taskq.Queue
	TASK_MAP              map[string]*taskq.Task
	WORKER_CONTEXT        context.Context
	WORKER_CONTEXT_CANCEL context.CancelFunc
	// ENVIRONMENT
	ENVIRONMENT string
}

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

// Git Provider
type GitProvider string

const (
	GitProviderGithub GitProvider = "github"
	GitProviderGitlab GitProvider = "gitlab"
)

// Upstream Type
type UpstreamType string

const (
	UpstreamTypeGit     UpstreamType = "git"
	UpstreamTypeTarball UpstreamType = "tarball"
	UpstreamTypeImage   UpstreamType = "image"
)

type DeploymentStatus string

const (
	DeploymentStatusQueued    DeploymentStatus = "queued"
	DeploymentStatusDeploying DeploymentStatus = "deploying"
	DeploymentStatusDeployed  DeploymentStatus = "deployed"
	DeploymentStatusFailed    DeploymentStatus = "failed"
)

// Application
type Application struct {
	ID                   uint       `json:"id" gorm:"primaryKey"`
	ServiceName          string     `json:"service_name" gorm:"unique"`
	EnvironmentVariables string     `json:"environment_variables" validate:"required"` // JSON string
	Volumes              string     `json:"volumes" validate:"required"`               // JSON string
}

//? Application Status <--> Latest Deployment Status

// Deployment
type Deployment struct {
	ID string `json:"id" gorm:"primaryKey"`
	// Foreign key to Application
	Application   Application  `json:"-"`
	ApplicationID uint         `json:"application_id"`
	UpstreamType  UpstreamType `json:"upstream_type" validate:"required"`
	// Fields for UpstreamType = Git
	GitCredential      GitCredential `json:"git_credential"`
	GitCredentialID    uint          `json:"git_credential_id"`
	GitProvider        GitProvider   `json:"git_provider"`
	RepositoryUsername string        `json:"repository_username"`
	RepositoryName     string        `json:"repository_name"`
	Branch             string        `json:"branch"`
	LastCommit         string        `json:"last_commit"`
	// Fields for UpstreamType = Tarball
	TarballFile string `json:"tarball_file"`
	// Fields for UpstreamType = Image
	DockerImage string `json:"docker_image"`
	// Common Fields
	BuildArgs  string `json:"build_args" validate:"required"`
	Dockerfile string `json:"dockerfile" validate:"required"`
	// No of replicas to be deployed
	Replicas uint `json:"replicas"`
	// Deployment Status
	Status    DeploymentStatus `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
}

// Deployment Log
type DeploymentLog struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	Deployment   Deployment `json:"-"`
	DeploymentID string     `json:"deployment_id"`
	Content      string     `json:"content"`
	CreatedAt    time.Time  `json:"created_at"`
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
