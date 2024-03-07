package core

import (
	DOCKER_CLIENT "github.com/docker/docker/client"
	"github.com/go-redis/redis/v8"
	DOCKER "github.com/swiftwave-org/swiftwave/container_manager"
	DOCKER_CONFIG_GENERATOR "github.com/swiftwave-org/swiftwave/docker_config_generator"
	HAPROXY "github.com/swiftwave-org/swiftwave/haproxy_manager"
	"github.com/swiftwave-org/swiftwave/pubsub"
	SSL "github.com/swiftwave-org/swiftwave/ssl_manager"
	"github.com/swiftwave-org/swiftwave/task_queue"
	UDP_PROXY "github.com/swiftwave-org/swiftwave/udp_proxy_manager"
	"gorm.io/gorm"
)

// ************************************************************************************* //
//                                Swiftwave System Configuration 		   			     //
// ************************************************************************************* //

type AMQPProtocol string

const (
	AMQP  AMQPProtocol = "amqp"
	AMQPS AMQPProtocol = "amqps"
)

type PubSubMode string

const (
	LocalPubSub  PubSubMode = "local"
	RemotePubSub PubSubMode = "remote"
)

type TaskQueueMode string

const (
	LocalTaskQueue  TaskQueueMode = "local"
	RemoteTaskQueue TaskQueueMode = "remote"
)

type ServerStatus string

const (
	ServerNeedsSetup ServerStatus = "needs_setup"
	ServerPreparing  ServerStatus = "preparing"
	ServerOnline     ServerStatus = "online"
	ServerOffline    ServerStatus = "offline"
)

type ImageRegistryConfig struct {
	Endpoint  string `json:"endpoint"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Namespace string `json:"namespace"`
}

// LetsEncryptConfig : hold information about lets encrypt configuration
type LetsEncryptConfig struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	Staging    bool   `json:"staging" gorm:"default:false"`
	EmailID    string `json:"email_id"`
	PrivateKey string `json:"private_key"`
}

// FirewallConfig : hold information about firewall configuration
type FirewallConfig struct {
	Enabled          bool   `json:"enabled" gorm:"default:false"`
	AllowPortCommand string `json:"allow_port_command"` // can contain {{port}} as placeholder
	DenyPortCommand  string `json:"deny_port_command"`  // can contain {{port}} as placeholder
}

// PersistentVolumeBackupConfig : configuration for persistent volume backup
type PersistentVolumeBackupConfig struct {
	S3BackupConfig S3BackupConfig `json:"s3_backup_config" gorm:"embedded;embeddedPrefix:s3_backup_"`
}

type S3BackupConfig struct {
	Enabled         bool   `json:"enabled"`
	Endpoint        string `json:"endpoint"`
	Region          string `json:"region"`
	Bucket          string `json:"bucket"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	ForcePathStyle  bool   `json:"force_path_style"`
}

// PubSubConfig : configuration for pub-sub system
type PubSubConfig struct {
	Mode         PubSubMode  `json:"mode" gorm:"default:'local'"`
	BufferLength uint        `json:"buffer_length" gorm:"default:2000"`
	RedisConfig  RedisConfig `json:"redis_config" gorm:"embedded;embeddedPrefix:redis_"`
}

// TaskQueueConfig : configuration for task queue system
type TaskQueueConfig struct {
	Mode                           TaskQueueMode `json:"mode" gorm:"default:'local'"`
	MaxOutstandingMessagesPerQueue uint          `json:"max_outstanding_messages_per_queue" gorm:"default:2"`
	NoOfWorkersPerQueue            uint          `json:"no_of_workers_per_queue"`
	AMQPConfig                     AMQPConfig    `json:"amqp_config" gorm:"embedded;embeddedPrefix:amqp_"`
}

// RedisConfig : configuration for Redis
type RedisConfig struct {
	Host       string `json:"host"`
	Port       uint   `json:"port"`
	Password   string `json:"password"`
	DatabaseID uint   `json:"database_id"`
}

// AMQPConfig : configuration for AMQP
type AMQPConfig struct {
	Protocol AMQPProtocol `json:"protocol"`
	Host     string       `json:"host"`
	User     string       `json:"user"`
	Password string       `json:"password"`
	VHost    string       `json:"vhost"`
}

type HAProxyConfig struct {
	Image          string `json:"image"`
	UnixSocketPath string `json:"unix_socket_path"`
	Username       string `json:"username"`
	Password       string `json:"password"`
}

type UDPProxyConfig struct {
	Image string `json:"image"`
}

// ************************************************************************************* //
//                                Application Level Table       		   			     //
// ************************************************************************************* //

// ServiceManager : holds the instance of all the managers
type ServiceManager struct {
	SslManager            SSL.Manager
	HaproxyManager        HAPROXY.Manager
	UDPProxyManager       UDP_PROXY.Manager
	DockerManager         DOCKER.Manager
	DockerConfigGenerator DOCKER_CONFIG_GENERATOR.Manager
	DockerClient          DOCKER_CLIENT.Client
	DbClient              gorm.DB
	RedisClient           redis.Client
	PubSubClient          pubsub.Client
	TaskQueueClient       task_queue.Client
	CancelImageBuildTopic string
}

// UpstreamType : type of source for the codebase of the application
type UpstreamType string

const (
	UpstreamTypeGit        UpstreamType = "git"
	UpstreamTypeSourceCode UpstreamType = "sourceCode"
	UpstreamTypeImage      UpstreamType = "image"
)

// GitProvider : type of git provider
type GitProvider string

const (
	GitProviderNone   GitProvider = "none"
	GitProviderGithub GitProvider = "github"
	GitProviderGitlab GitProvider = "gitlab"
)

// DomainSSLStatus : status of the ssl certificate for a domain
type DomainSSLStatus string

const (
	DomainSSLStatusNone    DomainSSLStatus = "none"
	DomainSSLStatusPending DomainSSLStatus = "pending"
	DomainSSLStatusFailed  DomainSSLStatus = "failed"
	DomainSSLStatusIssued  DomainSSLStatus = "issued"
)

// DeploymentStatus : status of the deployment
type DeploymentStatus string

const (
	DeploymentStatusPending       DeploymentStatus = "pending"
	DeploymentStatusDeployPending DeploymentStatus = "deployPending"
	DeploymentStatusLive          DeploymentStatus = "live"
	DeploymentStatusStopped       DeploymentStatus = "stopped"
	DeploymentStatusFailed        DeploymentStatus = "failed"
	DeploymentStalled             DeploymentStatus = "stalled"
)

// ProtocolType : type of protocol for ingress rule
type ProtocolType string

const (
	HTTPProtocol  ProtocolType = "http"
	HTTPSProtocol ProtocolType = "https"
	TCPProtocol   ProtocolType = "tcp"
	UDPProtocol   ProtocolType = "udp"
)

// IngressRuleStatus : status of the ingress rule
type IngressRuleStatus string

const (
	IngressRuleStatusPending  IngressRuleStatus = "pending"
	IngressRuleStatusApplied  IngressRuleStatus = "applied"
	IngressRuleStatusFailed   IngressRuleStatus = "failed"
	IngressRuleStatusDeleting IngressRuleStatus = "deleting"
)

// RedirectRuleStatus : status of the redirect rule
type RedirectRuleStatus string

const (
	RedirectRuleStatusPending  RedirectRuleStatus = "pending"
	RedirectRuleStatusApplied  RedirectRuleStatus = "applied"
	RedirectRuleStatusFailed   RedirectRuleStatus = "failed"
	RedirectRuleStatusDeleting RedirectRuleStatus = "deleting"
)

// DeploymentMode : mode of deployment of application (replicated or global)
type DeploymentMode string

const (
	DeploymentModeReplicated DeploymentMode = "replicated"
	DeploymentModeGlobal     DeploymentMode = "global"
)

// ApplicationUpdateResult : result of application update
type ApplicationUpdateResult struct {
	RebuildRequired bool
	ReloadRequired  bool
	DeploymentId    string
}

// DeploymentUpdateResult : result of deployment update
type DeploymentUpdateResult struct {
	RebuildRequired bool
	DeploymentId    string
}

// BackupType : type of backup
type BackupType string

const (
	LocalBackup BackupType = "local"
	S3Backup    BackupType = "s3"
)

// BackupStatus : status of backup
type BackupStatus string

const (
	BackupPending BackupStatus = "pending"
	BackupFailed  BackupStatus = "failed"
	BackupSuccess BackupStatus = "success"
)

// RestoreType : type of restore
type RestoreType string

const (
	LocalRestore RestoreType = "local"
)

// RestoreStatus : status of restore
type RestoreStatus string

const (
	RestorePending RestoreStatus = "pending"
	RestoreFailed  RestoreStatus = "failed"
	RestoreSuccess RestoreStatus = "success"
)

// PersistentVolumeType : type of persistent volume
type PersistentVolumeType string

const (
	PersistentVolumeTypeLocal PersistentVolumeType = "local"
	PersistentVolumeTypeNFS   PersistentVolumeType = "nfs"
)

// NFSConfig : configuration for NFS Storage
type NFSConfig struct {
	Host    string `json:"host,omitempty"`
	Path    string `json:"path,omitempty"`
	Version int    `json:"version,omitempty"`
}
